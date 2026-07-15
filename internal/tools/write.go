package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes content to path, creating parent dirs. Errors return as
// part of the output string (PowerShell tool convention), never as a Go error, so
// the model receives a write failure as an ordinary tool result.
func WriteFile(path, content string) string {
	if path == "" {
		return "(empty path)"
	}
	// Refuse an existing non-regular target: open(2) with O_WRONLY on a FIFO
	// with no reader blocks forever, leaking the tool goroutine past Ctrl+C
	// (which cancels the turn but can't unblock the open). Stat never blocks;
	// directories fall through to os.WriteFile's immediate EISDIR.
	if info, err := statPath(path); err == nil && !info.Mode().IsRegular() && !info.IsDir() {
		return fmt.Sprintf("(write error: %s is not a regular file)", path)
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Sprintf("(mkdir error: %v)", err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Sprintf("(write error: %v)", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), path)
}

// WriteFileSchema is the OpenAI tool definition for write_file. The description
// steers the model away from shell-quoting failure modes for
// small-to-medium writes. Large files should be built in smaller verified chunks
// because streamed tool-call arguments may truncate server-side.
func WriteFileSchema() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        WriteFileName,
			"description": "Write exact content bytes to a file. Creates parent directories and overwrites existing files without shell quoting. Use this for small-to-medium multi-line content, including text with quotes, dollar signs, or backticks. Large tool-call arguments can be truncated by some model servers, so build large files in smaller verified chunks; PowerShell here-strings with `Set-Content`/`Add-Content` are an alternative when appropriate.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Absolute or relative file path. Relative paths resolve against the working directory.",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Exact bytes to write to the file.",
					},
				},
				"required": []string{"path", "content"},
			},
		},
	}
}
