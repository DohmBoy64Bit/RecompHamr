package tools

import (
	"fmt"
	"os"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

// ReadFile returns path's contents, truncated to the shared tool-output budget
// (Truncate). The model gets exact bytes, not a shell-mangled approximation.
// Per the PowerShell/write/edit convention, filesystem errors come back in the
// output string, never as a Go error: the model reacts to them like a failed tool call.
func ReadFile(path string) string {
	if path == "" {
		return "(empty path)"
	}
	// Refuse non-regular files up front: open(2) on a FIFO blocks forever
	// waiting for a writer (leaking the tool goroutine past Ctrl+C, which
	// cancels the turn but can't unblock the read), and an endless device file
	// (/dev/zero) grows ReadFile's buffer without bound. Stat never blocks.
	if info, err := os.Stat(path); err == nil && !info.Mode().IsRegular() && !info.IsDir() {
		return fmt.Sprintf("(read error: %s is not a regular file)", path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("(read error: %v)", err)
	}
	return chmctx.Truncate(string(raw))
}

// ReadFileSchema is the OpenAI tool definition for read_file. The description
// nudges the model toward direct file reads instead of routing simple inspection
// through a command shell.
func ReadFileSchema() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        ReadFileName,
			"description": "Read a file directly and return exact content bytes with no shell quoting. Prefer this over PowerShell `Get-Content` when inspecting a whole file. Output over 6k tokens is truncated to first+last 2k; for targeted slices of large files use PowerShell commands such as `Select-String` or indexed `Get-Content` ranges.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Absolute or relative file path.",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}
