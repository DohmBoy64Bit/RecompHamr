package tools

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const RepomixrName = "repomixr"

var RepomixrDir string

func RepomixrSchema() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        RepomixrName,
			"description": "Clone a GitHub repository and pack all source files into a single XML file for LLM consumption. Use when you need to reference code from another repository. The packed file is saved to the project's repo cache directory.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"repo_url": map[string]any{
						"type":        "string",
						"description": "GitHub repository URL (e.g. https://github.com/user/repo)",
					},
					"branch": map[string]any{
						"type":        "string",
						"description": "Branch or tag to checkout (default: main)",
					},
					"remove_comments": map[string]any{
						"type":        "boolean",
						"description": "Remove comments from source files (default: false)",
					},
					"remove_empty_lines": map[string]any{
						"type":        "boolean",
						"description": "Remove empty lines from output (default: false)",
					},
					"show_line_numbers": map[string]any{
						"type":        "boolean",
						"description": "Prefix each line with its line number (default: false)",
					},
					"compress": map[string]any{
						"type":        "boolean",
						"description": "Collapse multiple whitespace to single space (default: false)",
					},
				},
				"required": []string{"repo_url"},
			},
		},
	}
}

func Repomixr(ctx context.Context, repoURL, branch string, removeComments, removeEmpty, showLines, compress bool) string {
	if RepomixrDir == "" {
		return "(repomixr: output directory not configured)"
	}

	owner, name := parseRepoURL(repoURL)
	if owner == "" || name == "" {
		return fmt.Sprintf("(repomixr: could not parse repo URL %q — expected https://github.com/owner/repo)", repoURL)
	}

	if branch == "" {
		branch = "main"
	}

	repoDir := filepath.Join(RepomixrDir, owner+"-"+name)
	cloneDir := filepath.Join(repoDir, "repo")
	outPath := filepath.Join(repoDir, "packed.xml")

	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		return fmt.Sprintf("(repomixr: mkdir %s: %v)", repoDir, err)
	}

	cloneArgs := []string{"clone", "--depth", "1", "--branch", branch, repoURL, cloneDir}
	cmd := exec.CommandContext(ctx, "git", cloneArgs...)
	var cloneErr bytes.Buffer
	cmd.Stderr = &cloneErr
	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("(repomixr: git clone failed: %s)", cloneErr.String())
	}

	files, err := walkTextFiles(cloneDir)
	if err != nil {
		return fmt.Sprintf("(repomixr: walk: %v)", err)
	}
	sort.Strings(files)

	var out bytes.Buffer
	out.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(&out, "<repository url=\"%s\" branch=\"%s\" files=\"%d\">\n", repoURL, branch, len(files))

	out.WriteString("  <directory_structure>\n")
	out.WriteString(dirTree(cloneDir, files))
	out.WriteString("  </directory_structure>\n")

	for _, f := range files {
		rel, _ := filepath.Rel(cloneDir, f)
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")

		if removeEmpty {
			filtered := lines[:0]
			for _, l := range lines {
				if strings.TrimSpace(l) != "" {
					filtered = append(filtered, l)
				}
			}
			lines = filtered
		}

		if showLines {
			for i := range lines {
				lines[i] = fmt.Sprintf("%6d | %s", i+1, lines[i])
			}
		}

		joined := strings.Join(lines, "\n")

		if removeComments {
			joined = stripComments(joined)
		}

		if compress {
			joined = collapseWhitespace(joined)
		}

		fmt.Fprintf(&out, "  <file path=\"%s\" lines=\"%d\">\n", rel, len(lines))
		out.WriteString("    <![CDATA[")
		if len(joined) > 0 && joined[0] == '\n' {
			out.WriteByte('\n')
		}
		out.WriteString(joined)
		if len(joined) > 0 && joined[len(joined)-1] != '\n' {
			out.WriteByte('\n')
		}
		out.WriteString("]]>\n")
		out.WriteString("  </file>\n")
	}

	instPath := filepath.Join(RepomixrDir, "..", "repomix-instruction.md")
	if inst, err := os.ReadFile(instPath); err == nil && len(inst) > 0 {
		out.WriteString("  <instruction>\n")
		out.WriteString("    <![CDATA[")
		out.Write(inst)
		if len(inst) > 0 && inst[len(inst)-1] != '\n' {
			out.WriteByte('\n')
		}
		out.WriteString("]]>\n")
		out.WriteString("  </instruction>\n")
	}

	out.WriteString("</repository>\n")

	if err := os.WriteFile(outPath, out.Bytes(), 0o644); err != nil {
		return fmt.Sprintf("(repomixr: write: %v)", err)
	}

	info, _ := os.Stat(outPath)
	size := ""
	if info != nil {
		size = fmt.Sprintf(" · %s", humanSize(info.Size()))
	}
	return fmt.Sprintf("Packed %d files from %s/%s (branch: %s)\n  → %s%s\n\nUse read_file to inspect the packed output.", len(files), owner, name, branch, outPath, size)
}

func parseRepoURL(url string) (owner, name string) {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	url = strings.TrimPrefix(url, "github.com/")
	parts := strings.SplitN(url, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func walkTextFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".exe", ".dll", ".so", ".dylib", ".o", ".a", ".lib",
			".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg",
			".zip", ".tar", ".gz", ".bz2", ".xz", ".7z",
			".mp3", ".mp4", ".wav", ".avi", ".mov",
			".pdf", ".doc", ".docx", ".xls", ".xlsx",
			".pyc", ".class", ".jar", ".war":
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > 512*1024 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if !utf8.Valid(data) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func stripComments(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if idx := strings.Index(line, "//"); idx >= 0 {
			if idx == 0 || line[idx-1] != ':' {
				line = line[:idx]
			}
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return strings.TrimRight(out.String(), "\n")
}

func collapseWhitespace(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	ws := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !ws {
				out.WriteByte(' ')
				ws = true
			}
		} else if r == '\n' || r == '\r' {
			if !ws {
				out.WriteByte(' ')
				ws = true
			}
		} else {
			out.WriteRune(r)
			ws = false
		}
	}
	return out.String()
}

func humanSize(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func dirTree(root string, files []string) string {
	type node struct {
		name     string
		children map[string]*node
		isFile   bool
	}
	tree := &node{children: make(map[string]*node)}

	for _, f := range files {
		rel, _ := filepath.Rel(root, f)
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		cur := tree
		for i, p := range parts {
			isFile := i == len(parts)-1
			if cur.children[p] == nil {
				cur.children[p] = &node{name: p, children: make(map[string]*node)}
			}
			cur = cur.children[p]
			cur.isFile = isFile
		}
	}

	var b strings.Builder
	var walk func(n *node, depth int)
	walk = func(n *node, depth int) {
		names := make([]string, 0, len(n.children))
		for name := range n.children {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			child := n.children[name]
			indent := strings.Repeat("  ", depth+1)
			if child.isFile && len(child.children) == 0 {
				fmt.Fprintf(&b, "%s%s\n", indent, child.name)
			} else {
				fmt.Fprintf(&b, "%s%s/\n", indent, child.name)
				walk(child, depth+1)
			}
		}
	}
	walk(tree, 0)
	return b.String()
}
