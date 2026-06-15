package skills

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed *.md
var fsys embed.FS

func Names() []string {
	entries, _ := fsys.ReadDir(".")
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".md"))
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
	b, err := fsys.ReadFile(filepath.ToSlash(n + ".md"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ListMarkdown(active []string) string {
	activeSet := map[string]bool{}
	for _, a := range active {
		activeSet[a] = true
	}
	var b strings.Builder
	b.WriteString("Built-in skills for v0.1.0:\n")
	for _, n := range Names() {
		mark := " "
		if activeSet[n] {
			mark = "*"
		}
		fmt.Fprintf(&b, "%s %s\n", mark, n)
	}
	b.WriteString("\nLoad one with /skill <name>. Platform-specific packs intentionally omitted in this first release.\n")
	return b.String()
}

