package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func restoreMainHooks(t *testing.T) {
	origStart, origHelp, origExit, origArgs := startApplication, printApplicationHelp, exitProcess, os.Args
	t.Cleanup(func() {
		startApplication, printApplicationHelp, exitProcess, os.Args = origStart, origHelp, origExit, origArgs
	})
}

func TestRunHelpVersionAndStart(t *testing.T) {
	for _, arg := range []string{"-v", "--version", "version"} {
		var out bytes.Buffer
		if err := run([]string{arg}, &out); err != nil || !strings.Contains(out.String(), "recomphamr") {
			t.Fatalf("%s: %q %v", arg, out.String(), err)
		}
	}

	restoreMainHooks(t)
	printApplicationHelp = func(w io.Writer) error {
		_, err := io.WriteString(w, "help")
		return err
	}
	for _, arg := range []string{"-h", "--help", "help"} {
		var out bytes.Buffer
		if err := run([]string{arg}, &out); err != nil || out.String() != "help" {
			t.Fatalf("%s: %q %v", arg, out.String(), err)
		}
	}

	started := 0
	startApplication = func(w io.Writer, gotVersion string) error {
		started++
		if w != io.Discard || gotVersion != version {
			t.Fatalf("start args = %v %q", w, gotVersion)
		}
		return nil
	}
	if err := run(nil, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"unknown-is-start"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if started != 2 {
		t.Fatalf("starts = %d", started)
	}
}

func TestRunPropagatesErrors(t *testing.T) {
	restoreMainHooks(t)
	boom := errors.New("boom")
	startApplication = func(io.Writer, string) error { return boom }
	if err := run(nil, io.Discard); !errors.Is(err, boom) {
		t.Fatalf("start error = %v", err)
	}
	printApplicationHelp = func(io.Writer) error { return boom }
	if err := run([]string{"--help"}, io.Discard); !errors.Is(err, boom) {
		t.Fatalf("help error = %v", err)
	}
	if err := run([]string{"--version"}, failingWriter{boom}); !errors.Is(err, boom) {
		t.Fatalf("version error = %v", err)
	}
}

func TestMainVersionPath(t *testing.T) {
	restoreMainHooks(t)
	os.Args = []string{"recomphamr", "--version"}
	main()
}

func TestHandleRunResultExitContract(t *testing.T) {
	restoreMainHooks(t)
	code := 0
	exitProcess = func(got int) { code = got }
	handleRunResult(nil)
	if code != 0 {
		t.Fatalf("success exit = %d", code)
	}
	handleRunResult(errors.New("startup"))
	if code != 1 {
		t.Fatalf("failure exit = %d", code)
	}
}
