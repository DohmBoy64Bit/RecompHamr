package skills

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed *.md
var fsys embed.FS

var customDir string

func SetCustomDir(dir string) {
	customDir = dir
}

func Names() []string {
	set := map[string]bool{}

	entries, _ := fsys.ReadDir(".")
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		set[strings.TrimSuffix(e.Name(), ".md")] = true
	}

	if customDir != "" {
		if de, err := os.ReadDir(customDir); err == nil {
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

	if customDir != "" {
		path := filepath.Join(customDir, n+".md")
		if b, err := os.ReadFile(path); err == nil {
			return string(b), nil
		}
	}

	b, err := fsys.ReadFile(filepath.ToSlash(n + ".md"))
	if err != nil {
		return "", fmt.Errorf("skill %q not found", name)
	}
	return string(b), nil
}

func ListMarkdown(active []string) string {
	activeSet := map[string]bool{}
	for _, a := range active {
		activeSet[a] = true
	}
	embedded := map[string]bool{}
	entries, _ := fsys.ReadDir(".")
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			embedded[strings.TrimSuffix(e.Name(), ".md")] = true
		}
	}
	var b strings.Builder
	b.WriteString("Built-in RE skills:\n")
	for _, n := range Names() {
		mark := " "
		if activeSet[n] {
			mark = "*"
		}
		label := n
		if !embedded[n] {
			label = n + " (custom)"
		}
		fmt.Fprintf(&b, "%s %s\n", mark, label)
	}
	b.WriteString("\nLoad one with /skill <name>.\n")
	return b.String()
}
