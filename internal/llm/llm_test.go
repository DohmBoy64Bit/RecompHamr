package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

func TestProbeReadsContextWindow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Context-Window", "131072")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	res, err := New(srv.URL, "test-model", "").Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.ContextWindow != 131072 {
		t.Fatalf("ContextWindow = %d", res.ContextWindow)
	}
}

func TestProbeMapsUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	_, err := New(srv.URL, "test-model", "bad").Probe(context.Background())
	if !errors.Is(err, provider.ErrUnauthorized) {
		t.Fatalf("err = %v", err)
	}
}
