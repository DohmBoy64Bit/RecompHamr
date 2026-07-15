package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    int
	}{
		{"complete", "mode: set\na.go:1.1,2.1 2 1\n", 0},
		{"partial", "mode: set\na.go:1.1,2.1 2 0\n", 1},
		{"empty", "mode: set\n", 1},
		{"malformed row", "mode: set\nbad\n", 1},
		{"bad statements", "mode: set\na.go:1.1,2.1 x 1\n", 1},
		{"bad count", "mode: set\na.go:1.1,2.1 1 x\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "coverage.out")
			if err := os.WriteFile(path, []byte(tt.profile), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := check(path); got != tt.want {
				t.Fatalf("check() = %d, want %d", got, tt.want)
			}
		})
	}

	if got := check(filepath.Join(t.TempDir(), "missing")); got != 1 {
		t.Fatalf("missing check() = %d, want 1", got)
	}

	longPath := filepath.Join(t.TempDir(), "long.out")
	if err := os.WriteFile(longPath, []byte("mode: set\n"+strings.Repeat("x", 70_000)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := check(longPath); got != 1 {
		t.Fatalf("scanner-error check() = %d, want 1", got)
	}
}

func TestMainFunction(t *testing.T) {
	oldArgs, oldExit := os.Args, exitProcess
	t.Cleanup(func() { os.Args, exitProcess = oldArgs, oldExit })

	exits := []int{}
	exitProcess = func(code int) { exits = append(exits, code) }
	os.Args = []string{"coveragecheck"}
	main()

	path := filepath.Join(t.TempDir(), "coverage.out")
	if err := os.WriteFile(path, []byte("mode: set\na.go:1.1,2.1 1 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.Args = []string{"coveragecheck", path}
	main()
	if len(exits) != 2 || exits[0] != 2 || exits[1] != 0 {
		t.Fatalf("exit codes = %v, want [2 0]", exits)
	}
}
