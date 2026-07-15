//go:build unix

package tools

import (
	"os/exec"
	"syscall"
)

var killProcessGroup = func(pid int) error { return syscall.Kill(-pid, syscall.SIGKILL) }

func configureProcessTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return killProcessGroup(cmd.Process.Pid)
	}
}
