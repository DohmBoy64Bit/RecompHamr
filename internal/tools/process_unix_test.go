//go:build unix

package tools

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestConfigureProcessTreeCancellation(t *testing.T) {
	if err := killProcessGroup(1 << 30); err == nil {
		t.Fatal("killing a nonexistent process group unexpectedly succeeded")
	}

	cmd := exec.CommandContext(context.Background(), "pwsh", "-NoProfile", "-Command", "Start-Sleep -Seconds 30")
	configureProcessTree(cmd)
	if cmd.SysProcAttr == nil || !cmd.SysProcAttr.Setpgid || cmd.Cancel == nil {
		t.Fatal("process-group cancellation was not configured")
	}
	if err := cmd.Cancel(); err != nil {
		t.Fatalf("pre-start cancel = %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Cancel(); err != nil {
		t.Fatalf("running cancel = %v", err)
	}
	_ = cmd.Wait()

	original := killProcessGroup
	t.Cleanup(func() { killProcessGroup = original })
	boom := errors.New("boom")
	killProcessGroup = func(pid int) error {
		if pid != 42 {
			t.Fatalf("pid = %d", pid)
		}
		return boom
	}
	fake := exec.CommandContext(context.Background(), "pwsh")
	configureProcessTree(fake)
	fake.Process = &os.Process{Pid: 42}
	if err := fake.Cancel(); !errors.Is(err, boom) {
		t.Fatalf("kill error = %v", err)
	}
}
