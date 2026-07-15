package agent

import (
	"errors"
	"testing"

	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

func TestDiagnosticMessages(t *testing.T) {
	if got := StreamErrorMessage(nil, "p", "u"); got != "" {
		t.Fatalf("nil stream error = %q", got)
	}
	if got := StreamErrorMessage(provider.ErrUnauthorized, "p", "u"); got != "⚠ key rejected · check models.p.key in .rehamr/config.yaml" {
		t.Fatalf("unauthorized stream error = %q", got)
	}
	unreachable := provider.ErrUnreachable{Err: errors.New("down")}
	if got := StreamErrorMessage(unreachable, "p", "u"); got != "⚠ unreachable: u · /models to switch profile" || !IsUnreachable(unreachable) {
		t.Fatalf("unreachable stream error = %q", got)
	}
	if got := StreamErrorMessage(errors.New("raw"), "p", "u"); got != "⚠ raw" || IsUnreachable(errors.New("raw")) {
		t.Fatalf("raw stream error = %q", got)
	}
	if got := ProbeErrorMessage(provider.ErrUnauthorized); got != "key rejected" {
		t.Fatalf("unauthorized probe error = %q", got)
	}
	if got := ProbeErrorMessage(unreachable); got != "unreachable (down)" {
		t.Fatalf("unreachable probe error = %q", got)
	}
	if got := ProbeErrorMessage(errors.New("raw")); got != "raw" {
		t.Fatalf("raw probe error = %q", got)
	}
}
