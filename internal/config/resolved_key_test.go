package config

import "testing"

func TestResolvedKeyExpandsWholeEnvReference(t *testing.T) {
	t.Setenv("RECOMPHAMR_TEST_KEY", "secret-value")
	p := &Profile{Key: "${RECOMPHAMR_TEST_KEY}"}
	if got := p.ResolvedKey(); got != "secret-value" {
		t.Fatalf("ResolvedKey() = %q, want %q", got, "secret-value")
	}
}

func TestResolvedKeyPreservesLiteralDollarCharacters(t *testing.T) {
	p := &Profile{Key: "pa$$word"}
	if got := p.ResolvedKey(); got != "pa$$word" {
		t.Fatalf("ResolvedKey() = %q, want literal key", got)
	}
}

func TestResolvedKeyUnsetEnvReferenceIsEmpty(t *testing.T) {
	p := &Profile{Key: "${RECOMPHAMR_TEST_KEY_UNSET}"}
	if got := p.ResolvedKey(); got != "" {
		t.Fatalf("ResolvedKey() = %q, want empty string for unset variable", got)
	}
}
