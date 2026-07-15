package provider

import (
	"net/http"
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
