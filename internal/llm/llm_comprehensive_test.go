package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

func TestIdleTimeoutFromEnv(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  time.Duration
	}{
		{"", streamIdleTimeout},
		{"45m", 45 * time.Minute},
		{"12", 12 * time.Second},
		{"0", streamIdleTimeout},
		{"-1s", streamIdleTimeout},
		{"bad", streamIdleTimeout},
		{"999999999999999999999999", streamIdleTimeout},
	} {
		t.Setenv("RECOMPHAMR_IDLE_TIMEOUT", tt.value)
		if got := idleTimeoutFromEnv(); got != tt.want {
			t.Errorf("%q = %s, want %s", tt.value, got, tt.want)
		}
	}
}

func TestWireAndParsingHelpers(t *testing.T) {
	msgs := []chmctx.Message{{
		Role: chmctx.RoleAssistant, Content: "answer", ToolName: "tool", ToolCallID: "result",
		ToolCalls: []chmctx.ToolCall{{ID: "c1", Name: "read_file", Arguments: map[string]any{"path": "a.go"}}},
	}}
	wire := toWire(msgs)
	if len(wire) != 1 || wire[0].ToolCalls[0].Function.Arguments != `{"path":"a.go"}` {
		t.Fatalf("wire = %+v", wire)
	}
	good := toolSlot{id: "c", name: "write_file"}
	good.args.WriteString(`{"path":"x","content":"y"}`)
	if call := good.resolve(); call.ID != "c" || call.Arguments["path"] != "x" {
		t.Fatalf("resolved = %+v", call)
	}
	bad := toolSlot{}
	bad.args.WriteString("{")
	if _, ok := bad.resolve().Arguments["_parse_error"]; !ok {
		t.Fatal("malformed arguments need parse sentinel")
	}
	empty := toolSlot{}
	if len(empty.resolve().Arguments) != 0 {
		t.Fatal("empty args should resolve to empty map")
	}
	if firstLine(" first\r\nsecond ") != "first" || firstLine(" one ") != "one" {
		t.Fatal("firstLine failed")
	}
}

func TestErrorClassification(t *testing.T) {
	if got := errorMessageFromBody([]byte(`{"error":{"provider_hint":"hint","message":"message"}}`)); got != "hint" {
		t.Fatalf("hint = %q", got)
	}
	if got := errorMessageFromBody([]byte(`{"error":{"message":"message"}}`)); got != "message" {
		t.Fatalf("message = %q", got)
	}
	if got := errorMessageFromBody([]byte("raw\nrest")); got != "raw" {
		t.Fatalf("raw = %q", got)
	}
	hs := &httpStatusError{status: 429, msg: "busy"}
	if hs.Error() != "429: busy" || !retryable(hs) {
		t.Fatalf("status error = %v", hs)
	}
	for _, code := range []int{404, 408, 500} {
		if !retryable(&httpStatusError{status: code}) {
			t.Errorf("%d should retry", code)
		}
	}
	for _, code := range []int{400, 401, 403} {
		if retryable(&httpStatusError{status: code}) {
			t.Errorf("%d should not retry", code)
		}
	}
	if !retryable(provider.ErrUnreachable{Err: errors.New("dial")}) || retryable(errors.New("plain")) {
		t.Fatal("typed retry classification failed")
	}
}

func TestDispatchDeltaAndReadSSE(t *testing.T) {
	out := make(chan Event, 16)
	var content strings.Builder
	slots := map[int]*toolSlot{}
	order := []int{}
	d := streamDelta{Reasoning: "think", Content: "hello", ToolCalls: []toolCall{{
		Index: 2, ID: "c2", Function: toolCallFunc{Name: "write_file", Arguments: `{"path":"x"}`},
	}}}
	if !dispatchDelta(context.Background(), d, &content, slots, &order, out) {
		t.Fatal("dispatch failed")
	}
	if content.String() != "hello" || len(order) != 1 || slots[2].name != "write_file" {
		t.Fatalf("dispatch state content=%q order=%v slots=%v", content.String(), order, slots)
	}
	for i, want := range []EventKind{EventReasoning, EventContent, EventToolArgs} {
		if got := (<-out).Kind; got != want {
			t.Fatalf("event %d = %v, want %v", i, got, want)
		}
	}

	sse := strings.Join([]string{
		"data: not-json", "", ": keepalive", "",
		`data: {"choices":[{"delta":{"reasoning":"r","content":"A","tool_calls":[{"index":0,"id":"c1","function":{"name":"read_file","arguments":"{\"path\":"}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"x\"}"}}]}}],"usage":{"completion_tokens":7,"prompt_tokens":11}}`,
		"data: [DONE]", "",
	}, "\n")
	events := make(chan Event, 16)
	frames := 0
	final, tokens, prompt, err := readSSE(context.Background(), strings.NewReader(sse), events, func() { frames++ })
	if err != nil || final.Content != "A" || len(final.ToolCalls) != 1 || tokens != 7 || prompt != 11 || frames == 0 {
		t.Fatalf("final=%+v tokens=%d prompt=%d frames=%d err=%v", final, tokens, prompt, frames, err)
	}
	if final.ToolCalls[0].Arguments["path"] != "x" {
		t.Fatalf("call = %+v", final.ToolCalls[0])
	}

	errorSSE := `data: {"error":{"message":"provider died"}}` + "\n"
	if _, _, _, err := readSSE(context.Background(), strings.NewReader(errorSSE), make(chan Event, 1), func() {}); err == nil || !strings.Contains(err.Error(), "provider died") {
		t.Fatalf("stream error = %v", err)
	}
	errorRaw := `data: {"error":{}}` + "\n"
	if _, _, _, err := readSSE(context.Background(), strings.NewReader(errorRaw), make(chan Event, 1), func() {}); err == nil {
		t.Fatal("empty stream error should surface raw frame")
	}
	long := "data: " + strings.Repeat("x", 5<<20)
	if _, _, _, err := readSSE(context.Background(), strings.NewReader(long), make(chan Event, 1), func() {}); err == nil {
		t.Fatal("scanner overflow should fail")
	}
}

func TestSendEventCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if sendEvent(ctx, make(chan Event), Event{}) {
		t.Fatal("cancelled send should fail")
	}
	var b strings.Builder
	if dispatchDelta(ctx, streamDelta{Content: "x"}, &b, map[int]*toolSlot{}, new([]int), make(chan Event)) {
		t.Fatal("cancelled delta send should fail")
	}
	if dispatchDelta(ctx, streamDelta{Reasoning: "x"}, &b, map[int]*toolSlot{}, new([]int), make(chan Event)) {
		t.Fatal("cancelled reasoning send should fail")
	}
	if dispatchDelta(ctx, streamDelta{ToolCalls: []toolCall{{Function: toolCallFunc{Arguments: "x"}}}}, &b, map[int]*toolSlot{}, new([]int), make(chan Event)) {
		t.Fatal("cancelled tool-args send should fail")
	}
	contentSSE := "data: {\"choices\":[{\"delta\":{\"content\":\"x\"}}]}\n"
	if _, _, _, err := readSSE(ctx, strings.NewReader(contentSSE), make(chan Event), func() {}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled content read = %v", err)
	}
	toolSSE := "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"c\",\"function\":{\"name\":\"x\",\"arguments\":\"\"}}]}}]}\n"
	if _, _, _, err := readSSE(ctx, strings.NewReader(toolSSE), make(chan Event), func() {}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled tool read = %v", err)
	}
}

type errorRoundTripper struct{ err error }

func (r errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) { return nil, r.err }

func TestHTTPFailureShapesAndReasoningFallback(t *testing.T) {
	t.Run("bad request URL", func(t *testing.T) {
		c := New("://bad", "m", "")
		if _, _, err := c.doPost(context.Background(), chatRequest{}); err == nil {
			t.Fatal("bad URL should fail")
		}
	})
	t.Run("marshal", func(t *testing.T) {
		c := New("http://example.test", "m", "")
		_, _, err := c.doPost(context.Background(), chatRequest{Tools: []Tool{{Function: FunctionDef{Parameters: map[string]any{"bad": make(chan int)}}}}})
		if err == nil {
			t.Fatal("unsupported JSON value should fail")
		}
	})
	t.Run("transport", func(t *testing.T) {
		c := New("http://127.0.0.1:1", "m", "")
		_, _, err := c.doPost(context.Background(), chatRequest{})
		var unreachable provider.ErrUnreachable
		if !errors.As(err, &unreachable) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("unauthorized", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) }))
		defer srv.Close()
		c := New(srv.URL, "m", "secret")
		_, _, err := c.doPost(context.Background(), chatRequest{})
		if !errors.Is(err, provider.ErrUnauthorized) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("status and auth header", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer secret" || r.Header.Get("Accept") != "text/event-stream" {
				t.Errorf("headers = %v", r.Header)
			}
			w.WriteHeader(418)
			_, _ = io.WriteString(w, `{"error":{"message":"teapot"}}`)
		}))
		defer srv.Close()
		c := New(srv.URL, "m", "secret")
		_, body, err := c.doPost(context.Background(), chatRequest{})
		if string(body) == "" || err == nil || err.Error() != "418: teapot" {
			t.Fatalf("body=%q err=%v", body, err)
		}
	})
	t.Run("reasoning fallback sticks", func(t *testing.T) {
		requests := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			body, _ := io.ReadAll(r.Body)
			if requests == 1 {
				w.WriteHeader(400)
				_, _ = io.WriteString(w, `{"error":{"message":"does not support thinking"}}`)
				return
			}
			if strings.Contains(string(body), "reasoning_effort") {
				t.Error("fallback request retained reasoning_effort")
			}
			w.WriteHeader(200)
		}))
		defer srv.Close()
		c := New(srv.URL, "m", "")
		resp, err := c.postChat(context.Background(), chatRequest{ReasoningEffort: "high"})
		if err != nil || !c.noReasoningEffort.Load() {
			t.Fatalf("resp=%v err=%v sticky=%v", resp, err, c.noReasoningEffort.Load())
		}
		_ = resp.Body.Close()
		resp, err = c.postChat(context.Background(), chatRequest{ReasoningEffort: "high"})
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
	})
}

func TestChatSuccessRetryAndFailure(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Context-Window", "65536")
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}],\"usage\":{\"completion_tokens\":2,\"prompt_tokens\":3}}\n\ndata: [DONE]\n\n")
		}))
		defer srv.Close()
		c := New(srv.URL, "m", "")
		var events []Event
		for e := range c.Chat(context.Background(), []chmctx.Message{{Role: chmctx.RoleUser, Content: "hi"}}, nil) {
			events = append(events, e)
		}
		if len(events) != 2 || events[0].Kind != EventContent || events[1].Kind != EventDone || events[1].ContextWindow != 65536 {
			t.Fatalf("events = %+v", events)
		}
	})
	t.Run("retry then success", func(t *testing.T) {
		requests := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if requests == 1 {
				w.WriteHeader(500)
				return
			}
			_, _ = io.WriteString(w, "data: [DONE]\n\n")
		}))
		defer srv.Close()
		c := New(srv.URL, "m", "")
		c.RetryBackoff = []time.Duration{0}
		var kinds []EventKind
		for e := range c.Chat(context.Background(), nil, nil) {
			kinds = append(kinds, e.Kind)
		}
		if requests != 2 || len(kinds) != 2 || kinds[0] != EventRetry || kinds[1] != EventDone {
			t.Fatalf("requests=%d kinds=%v", requests, kinds)
		}
	})
	t.Run("permanent error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400) }))
		defer srv.Close()
		c := New(srv.URL, "m", "")
		var events []Event
		for e := range c.Chat(context.Background(), nil, nil) {
			events = append(events, e)
		}
		if len(events) != 1 || events[0].Kind != EventError {
			t.Fatalf("events=%+v", events)
		}
	})
	t.Run("cancel before retry notification", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c := New("http://example.test", "m", "")
		c.HTTP.Transport = errorRoundTripper{err: errors.New("dial")}
		c.RetryBackoff = []time.Duration{0}
		out := make(chan Event)
		c.run(ctx, nil, nil, out)
		if _, ok := <-out; ok {
			t.Fatal("cancelled retry notification should close without an event")
		}
	})
	t.Run("nonpositive idle timeout falls back", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "data: [DONE]\n\n")
		}))
		defer srv.Close()
		c := New(srv.URL, "m", "")
		c.IdleTimeout = 0
		for range c.Chat(context.Background(), nil, nil) {
		}
	})
	t.Run("cancel during retry", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
		defer srv.Close()
		c := New(srv.URL, "m", "")
		c.RetryBackoff = []time.Duration{time.Hour}
		ctx, cancel := context.WithCancel(context.Background())
		ch := c.Chat(ctx, nil, nil)
		if e := <-ch; e.Kind != EventRetry {
			t.Fatalf("event=%+v", e)
		}
		cancel()
		if _, ok := <-ch; ok {
			t.Fatal("channel should close after cancellation")
		}
	})
	t.Run("idle stall", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			<-r.Context().Done()
		}))
		defer srv.Close()
		c := New(srv.URL, "m", "")
		c.IdleTimeout = 10 * time.Millisecond
		var last Event
		for e := range c.Chat(context.Background(), nil, nil) {
			last = e
		}
		if last.Kind != EventError || !strings.Contains(last.Err.Error(), "stopped sending") {
			t.Fatalf("event=%+v", last)
		}
	})
}
