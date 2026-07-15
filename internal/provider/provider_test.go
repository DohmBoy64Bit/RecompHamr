package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContextWindowFromHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("X-Context-Window", "65536")
	if got := ContextWindowFromHeaders(h); got != 65536 {
		t.Fatalf("got %d", got)
	}
	h.Set("X-Context-Window", "not-a-number")
	if got := ContextWindowFromHeaders(h); got != 0 {
		t.Fatalf("invalid header got %d", got)
	}
}

func TestProviderHelpers(t *testing.T) {
	if got := AuthHeader("token"); got != "Bearer token" {
		t.Fatalf("AuthHeader = %q", got)
	}
	base := errors.New("dial")
	err := ErrUnreachable{Err: base}
	if err.Error() != "provider unreachable: dial" || !errors.Is(err, base) {
		t.Fatalf("ErrUnreachable contract failed: %v", err)
	}
}

func TestReachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("still reachable"))
	}))
	defer srv.Close()
	if err := Reachable(context.Background(), srv.URL); err != nil {
		t.Fatal(err)
	}
	if err := Reachable(context.Background(), "://bad"); err == nil {
		t.Fatal("malformed URL should fail")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Reachable(ctx, srv.URL); err == nil {
		t.Fatal("cancelled request should fail")
	}
}

func TestContextWindowHeaderBoundaries(t *testing.T) {
	tests := []struct {
		value string
		want  int
	}{
		{"", 0}, {"1023", 0}, {"1024", 1024}, {"8388608", 8388608}, {"8388609", 0},
	}
	for _, tt := range tests {
		h := http.Header{}
		if tt.value != "" {
			h.Set("X-Context-Window", tt.value)
		}
		if got := ContextWindowFromHeaders(h); got != tt.want {
			t.Errorf("value %q = %d, want %d", tt.value, got, tt.want)
		}
	}
}
