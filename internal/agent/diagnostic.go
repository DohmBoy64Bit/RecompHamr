package agent

import (
	"errors"

	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

// StreamErrorMessage maps a model-stream failure to the accepted concise
// frontend diagnostic without exposing provider error types to presentation.
func StreamErrorMessage(err error, profile, baseURL string) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, provider.ErrUnauthorized) {
		return "⚠ key rejected · check models." + profile + ".key in .rehamr/config.yaml"
	}
	if IsUnreachable(err) {
		return "⚠ unreachable: " + baseURL + " · /models to switch profile"
	}
	return "⚠ " + err.Error()
}

// ProbeErrorMessage maps an activation probe failure to the accepted inline
// detail while retaining an unreachable error's underlying cause.
func ProbeErrorMessage(err error) string {
	if errors.Is(err, provider.ErrUnauthorized) {
		return "key rejected"
	}
	if unreachable, ok := errors.AsType[provider.ErrUnreachable](err); ok {
		return "unreachable (" + unreachable.Err.Error() + ")"
	}
	return err.Error()
}

// IsUnreachable reports whether err represents provider reachability failure.
func IsUnreachable(err error) bool {
	_, ok := errors.AsType[provider.ErrUnreachable](err)
	return ok
}
