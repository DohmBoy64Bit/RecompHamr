//go:build unix

package tools

import (
	"os/exec"
	"testing"
)

func TestConfigureProcessTreeCancellation(t *testing.T) {
	cmd := exec.Command("pwsh", "-NoProfile", "-Command", "Start-Sleep -Seconds 30")
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
}
