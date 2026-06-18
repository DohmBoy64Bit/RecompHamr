package skills

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

//go:embed *.md
var fsys embed.FS

var (
	mu        sync.RWMutex
	customDir string
)

func SetCustomDir(dir string) {
	mu.Lock()
	defer mu.Unlock()
	customDir = dir
}

func Names() []string {
	mu.RLock()
	dir := customDir
	mu.RUnlock()

	set := map[string]bool{}

	entries, _ := fsys.ReadDir(".")
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		set[strings.TrimSuffix(e.Name(), ".md")] = true
	}

	if dir != "" {
		if de, err := os.ReadDir(dir); err == nil {
			for _, e := range de {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				set[strings.TrimSuffix(e.Name(), ".md")] = true
			}
		}
	}

	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func Resolve(name string) (string, error) {
	name = strings.TrimSpace(strings.TrimSuffix(name, ".md"))
	for _, n := range Names() {
		if strings.EqualFold(name, n) {
			return n, nil
		}
	}
	return "", fmt.Errorf("unknown skill %q", name)
}

func Get(name string) (string, error) {
	n, err := Resolve(name)
	if err != nil {
		return "", err
	}

	mu.RLock()
	dir := customDir
	mu.RUnlock()

	if dir != "" {
		path := filepath.Join(dir, n+".md")
		if b, err := os.ReadFile(path); err == nil {
			return string(b), nil
		}
	}

	b, err := fsys.ReadFile(filepath.ToSlash(n + ".md"))
	if err != nil {
		return "", fmt.Errorf("skill %q: %w", n, err)
	}
	return string(b), nil
}

func ListMarkdown(active []string) string {
	activeSet := map[string]bool{}
	for _, a := range active {
		activeSet[a] = true
	}
	var b strings.Builder
	b.WriteString("Built-in RE skills:\n")
	for _, n := range Names() {
		mark := " "
		if activeSet[n] {
			mark = "*"
		}
		label := n
		if !IsEmbedded(n) {
			label = n + " (custom)"
		}
		fmt.Fprintf(&b, "%s %s\n", mark, label)
	}
	b.WriteString("\nLoad one with /skill <name>.\n")
	return b.String()
}

func IsEmbedded(name string) bool {
	entries, _ := fsys.ReadDir(".")
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			if strings.TrimSuffix(e.Name(), ".md") == name {
				return true
			}
		}
	}
	return false
}
