//go:build windows

package tools

import "os/exec"

// CommandContext still terminates the PowerShell process itself on cancellation.
// A future process-tree hardening change belongs in this file and must be proven
// with a Windows integration test before the baseline gate can claim child-tree
// cancellation. The baseline does not claim that stronger behavior.
func configureProcessTree(_ *exec.Cmd) {}

func powerShellCandidates() []string { return []string{"pwsh", "powershell.exe", "powershell"} }
