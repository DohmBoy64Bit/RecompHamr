package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
)

func restoreRuntimeHooks(t *testing.T) {
	originalBootstrap, originalClient, originalReachable := bootstrapRuntimeConfig, newRuntimeClient, reachableRuntime
	t.Cleanup(func() {
		bootstrapRuntimeConfig, newRuntimeClient, reachableRuntime = originalBootstrap, originalClient, originalReachable
	})
}

func runtimeConfig(t *testing.T, url string) *config.Config {
	t.Helper()
	root := t.TempDir()
	cfg, _, err := config.Bootstrap(root)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ActiveProfile().URL = url
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestRuntimeSnapshotActivationAndReload(t *testing.T) {
	restoreRuntimeHooks(t)
	var clients []string
	newRuntimeClient = func(url, model, key string) *llm.Client {
		clients = append(clients, strings.Join([]string{url, model, key}, "|"))
		return llm.New(url, model, key)
	}
	cfg := runtimeConfig(t, "http://one")
	cfg.Models["other"] = &config.Profile{LLM: "other-model", URL: "http://two", Key: "secret", ContextSize: 4096}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	runtime := NewRuntime(cfg)
	snapshot := runtime.Snapshot()
	if snapshot.Active != "local" || snapshot.ActiveURL != "http://one" || snapshot.ActiveModel == "" || snapshot.ActiveKeyed || len(snapshot.Profiles) != 2 {
		t.Fatalf("initial snapshot = %#v", snapshot)
	}
	other, ok := snapshot.Profile("other")
	if !ok || !other.Keyed || other.URL != "http://two" || other.Active {
		t.Fatalf("other profile = %#v %v", other, ok)
	}
	if _, ok := snapshot.Profile("missing"); ok {
		t.Fatal("missing profile found")
	}
	snapshot, err := runtime.Activate("other")
	if err != nil || snapshot.Active != "other" || !snapshot.ActiveKeyed || len(clients) != 2 {
		t.Fatalf("activate = %#v %v clients=%#v", snapshot, err, clients)
	}
	if _, err := runtime.Activate("missing"); err == nil || runtime.Snapshot().Active != "other" || len(clients) != 2 {
		t.Fatalf("unknown activate = %v clients=%#v", err, clients)
	}

	snapshot, replaced, err := runtime.Reload()
	if err != nil || replaced || snapshot.Active != "other" {
		t.Fatalf("unchanged reload = %#v %v %v", snapshot, replaced, err)
	}
	path := filepath.Join(cfg.Dir, "config.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(raw), "http://two", "http://changed", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, replaced, err = runtime.Reload()
	if err != nil || !replaced || snapshot.ActiveURL != "http://changed" || len(clients) != 3 {
		t.Fatalf("changed reload = %#v %v %v clients=%#v", snapshot, replaced, err, clients)
	}

	boom := errors.New("boom")
	bootstrapRuntimeConfig = func(string) (*config.Config, bool, error) { return nil, false, boom }
	if _, _, err := runtime.Reload(); !errors.Is(err, boom) {
		t.Fatalf("reload error = %v", err)
	}
}

func TestRuntimeOverrideChatAndCapturedWork(t *testing.T) {
	restoreRuntimeHooks(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Context-Window", "8192")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: {\"choices\":[],\"usage\":{\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()
	cfg := runtimeConfig(t, "http://stored")
	cfg.URLOverride = server.URL
	runtime := NewRuntime(cfg)
	if runtime.Snapshot().ActiveURL != server.URL {
		t.Fatalf("override URL = %q", runtime.Snapshot().ActiveURL)
	}
	events := runtime.Chat(context.Background(), []chmctx.Message{{Role: chmctx.RoleUser, Content: "hi"}}, nil)
	for event := range events {
		if event.Kind == llm.EventError {
			t.Fatal(event.Err)
		}
	}
	probe := runtime.Probe(runtime.Snapshot().Active)
	probeResult := probe.Run(context.Background())
	if probeResult.Err != nil || probeResult.Profile != "local" || probeResult.ContextWindow != 8192 {
		t.Fatalf("probe = %#v", probeResult)
	}
	reach := runtime.Reachability()
	boom := errors.New("offline")
	reachableRuntime = func(_ context.Context, url string) error {
		if url != server.URL {
			t.Fatalf("reach URL = %q", url)
		}
		return boom
	}
	reachResult := reach.Run(context.Background())
	if reachResult.URL != server.URL || !errors.Is(reachResult.Err, boom) {
		t.Fatalf("reach = %#v", reachResult)
	}
}

func TestRuntimeHistoryDelegationAndActivateRollback(t *testing.T) {
	restoreRuntimeHooks(t)
	cfg := config.Default()
	cfg.Models["other"] = &config.Profile{LLM: "other", URL: "http://other", ContextSize: 1}
	runtime := NewRuntime(cfg)
	if _, err := runtime.Activate("other"); err == nil || runtime.Snapshot().Active != "local" {
		t.Fatalf("save rollback = %v %#v", err, runtime.Snapshot())
	}
	dir := t.TempDir()
	cfg.Dir = dir
	runtime = NewRuntime(cfg)
	if err := runtime.AppendHistory("remember"); err != nil {
		t.Fatal(err)
	}
	if got := runtime.LoadHistory(); len(got) != 1 || got[0] != "remember" {
		t.Fatalf("history = %#v", got)
	}
	if err := runtime.ClearHistory(); err != nil || len(runtime.LoadHistory()) != 0 {
		t.Fatalf("clear history = %v %#v", err, runtime.LoadHistory())
	}
}
