package tools

import (
	"os"
	"strings"
	"testing"
	"time"
)

type specialFileInfo struct{}

func (specialFileInfo) Name() string       { return "pipe" }
func (specialFileInfo) Size() int64        { return 0 }
func (specialFileInfo) Mode() os.FileMode  { return os.ModeNamedPipe }
func (specialFileInfo) ModTime() time.Time { return time.Time{} }
func (specialFileInfo) IsDir() bool        { return false }
func (specialFileInfo) Sys() any           { return nil }

func TestFileToolsRefuseSpecialFiles(t *testing.T) {
	original := statPath
	statPath = func(string) (os.FileInfo, error) { return specialFileInfo{}, nil }
	t.Cleanup(func() { statPath = original })

	for name, got := range map[string]string{
		"read":  ReadFile("pipe"),
		"write": WriteFile("pipe", "data"),
		"edit":  EditFile("pipe", "old", "new"),
	} {
		if !strings.Contains(got, "not a regular file") {
			t.Fatalf("%s special-file result = %q", name, got)
		}
	}
}
