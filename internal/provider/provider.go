// Package provider contains generic OpenAI-compatible endpoint helpers.
// It has no bundled hosted service, product-specific account system, or quota protocol.
package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
)

// Reachable probes the standard OpenAI-compatible models endpoint. Any HTTP
// response counts as reachable; only transport failures are returned.
func Reachable(ctx context.Context, baseURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

// ErrUnauthorized maps an HTTP 401 from an endpoint.
var ErrUnauthorized = errors.New("provider rejected credentials")

// ErrUnreachable wraps transport failures so callers can distinguish them
// from an HTTP error response.
type ErrUnreachable struct{ Err error }

func (e ErrUnreachable) Error() string { return "provider unreachable: " + e.Err.Error() }
func (e ErrUnreachable) Unwrap() error { return e.Err }

// AuthHeader builds the conventional bearer-token Authorization value.
func AuthHeader(token string) string { return "Bearer " + token }

const (
	headerContextWindow = "X-Context-Window"
	contextWindowMin    = 1024
	contextWindowMax    = 8 * 1024 * 1024
)

// ContextWindowFromHeaders accepts an optional endpoint-provided context-window
// hint. Missing or implausible values return 0 so config remains authoritative.
func ContextWindowFromHeaders(h http.Header) int {
	raw := h.Get(headerContextWindow)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < contextWindowMin || n > contextWindowMax {
		return 0
	}
	return n
}
