package tools

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

var gitLookPath = exec.LookPath
var repomixReadFile = os.ReadFile
var walkRepomix = filepath.WalkDir

// RepomixrName is the stable model-facing repository packer name.
const RepomixrName = "repomixr"

const (
	maxRepomixFileBytes  = 512 << 10
	maxRepomixFiles      = 10_000
	maxRepomixTotalBytes = 64 << 20
)

type repomixArgs struct {
	repoURL, branch                        string
	removeComments, removeEmpty, showLines bool
	compress                               bool
}

func repomixArgsFrom(args map[string]any) repomixArgs {
	repoURL, _ := args["repo_url"].(string)
	branch, _ := args["branch"].(string)
	removeComments, _ := args["remove_comments"].(bool)
	removeEmpty, _ := args["remove_empty_lines"].(bool)
	showLines, _ := args["show_line_numbers"].(bool)
	compress, _ := args["compress"].(bool)
	return repomixArgs{repoURL, branch, removeComments, removeEmpty, showLines, compress}
}

// RepomixrSchema returns the stable OpenAI-compatible function definition.
func RepomixrSchema() map[string]any {
	return map[string]any{"type": "function", "function": map[string]any{
		"name":        RepomixrName,
		"description": "Clone a public GitHub repository and pack bounded UTF-8 source files into a cached XML file for analysis. The cache is private to the current project user; use read_file on the returned path.",
		"parameters": map[string]any{"type": "object", "properties": map[string]any{
			"repo_url":           map[string]any{"type": "string", "description": "Public HTTPS GitHub repository URL, exactly https://github.com/owner/repository."},
			"branch":             map[string]any{"type": "string", "description": "Branch or tag to clone; defaults to main."},
			"remove_comments":    map[string]any{"type": "boolean", "description": "Remove simple // and full-line # comments; defaults to false."},
			"remove_empty_lines": map[string]any{"type": "boolean", "description": "Remove blank lines before packing; defaults to false."},
			"show_line_numbers":  map[string]any{"type": "boolean", "description": "Prefix retained lines with their packed sequence number; defaults to false."},
			"compress":           map[string]any{"type": "boolean", "description": "Collapse whitespace in each packed file; defaults to false."},
		}, "required": []string{"repo_url"}},
	}}
}

func parsePublicGitHubRepo(raw string) (owner, repo, canonical string, err error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Hostname(), "github.com") || u.Port() != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", "", "", errors.New("expected a credential-free https://github.com/owner/repository URL")
	}
	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	if len(parts) != 2 {
		return "", "", "", errors.New("expected exactly one owner and repository path")
	}
	owner = parts[0]
	repo = strings.TrimSuffix(parts[1], ".git")
	if !safeGitHubComponent(owner) || !safeGitHubComponent(repo) {
		return "", "", "", errors.New("owner and repository may contain only letters, digits, '.', '_', and '-'")
	}
	return owner, repo, "https://github.com/" + owner + "/" + repo, nil
}

func safeGitHubComponent(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func (s Set) repomix(ctx context.Context, args repomixArgs) string {
	if s.privateRoot == "" || s.restrict == nil {
		return "(repomixr: output directory not configured)"
	}
	owner, repo, canonical, err := parsePublicGitHubRepo(args.repoURL)
	if err != nil {
		return fmt.Sprintf("(repomixr: invalid repository URL: %v)", err)
	}
	branch := strings.TrimSpace(args.branch)
	if branch == "" {
		branch = "main"
	}
	if strings.HasPrefix(branch, "-") || strings.ContainsAny(branch, "\x00\r\n") {
		return "(repomixr: invalid branch or tag)"
	}
	cacheRoot := filepath.Join(s.privateRoot, "repos")
	repoRoot := filepath.Join(cacheRoot, owner+"-"+repo)
	cloneRoot := filepath.Join(repoRoot, "repo")
	if err := preparePrivateDir(cacheRoot, s.restrict); err != nil {
		return "(repomixr: cache: " + err.Error() + ")"
	}
	if err := refuseExistingSpecial(repoRoot); err != nil {
		return "(repomixr: cache: " + err.Error() + ")"
	}
	if err := cacheRemoveAll(repoRoot); err != nil {
		return "(repomixr: clean cache: " + err.Error() + ")"
	}
	if err := preparePrivateDir(repoRoot, s.restrict); err != nil {
		return "(repomixr: cache: " + err.Error() + ")"
	}
	gitPath, err := gitLookPath("git")
	if err != nil {
		return "(repomixr: git unavailable)"
	}
	out, err := s.runGit(ctx, gitPath, []string{"clone", "--depth", "1", "--single-branch", "--branch", branch, "--", canonical, cloneRoot}, repoRoot)
	if err != nil {
		if ctx.Err() != nil {
			return "(repomixr: cancelled)"
		}
		return fmt.Sprintf("(repomixr: git clone failed: %s)", boundedDiagnostic(out))
	}
	files, total, err := collectRepomixFiles(cloneRoot)
	if err != nil {
		return "(repomixr: scan: " + err.Error() + ")"
	}
	packed, err := buildRepomixXML(cloneRoot, canonical, branch, files, args, filepath.Join(s.privateRoot, "repomix-instruction.md"))
	if err != nil {
		return "(repomixr: pack: " + err.Error() + ")"
	}
	outPath := filepath.Join(repoRoot, "packed.xml")
	if err := atomicPrivateWrite(outPath, packed, s.restrict); err != nil {
		return "(repomixr: write: " + err.Error() + ")"
	}
	return fmt.Sprintf("Packed %d files (%s source) from %s/%s (branch: %s)\n  → %s · %s\n\nUse read_file to inspect the packed output.", len(files), humanSize(total), owner, repo, branch, outPath, humanSize(int64(len(packed))))
}

func collectRepomixFiles(root string) ([]string, int64, error) {
	return collectRepomixFilesWithLimits(root, maxRepomixFiles, maxRepomixTotalBytes)
}

func collectRepomixFilesWithLimits(root string, maxFiles int, maxTotal int64) ([]string, int64, error) {
	var files []string
	var total int64
	err := walkRepomix(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if path != root && filepath.Base(path) == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Size() > maxRepomixFileBytes || skippedRepomixExtension(filepath.Ext(path)) {
			return nil
		}
		data, err := repomixReadFile(path)
		if err != nil || !utf8.Valid(data) {
			return nil
		}
		if len(files) >= maxFiles || total+int64(len(data)) > maxTotal {
			return errors.New("repository exceeds the 10,000-file or 64 MiB source limit")
		}
		files = append(files, path)
		total += int64(len(data))
		return nil
	})
	sort.Strings(files)
	return files, total, err
}

func skippedRepomixExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".exe", ".dll", ".so", ".dylib", ".o", ".a", ".lib", ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg", ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".mp3", ".mp4", ".wav", ".avi", ".mov", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".pyc", ".class", ".jar", ".war":
		return true
	default:
		return false
	}
}

func buildRepomixXML(root, repoURL, branch string, files []string, args repomixArgs, instructionPath string) ([]byte, error) {
	var out bytes.Buffer
	out.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<repository url=\"")
	_ = xml.EscapeText(&out, []byte(repoURL))
	out.WriteString("\" branch=\"")
	_ = xml.EscapeText(&out, []byte(branch))
	fmt.Fprintf(&out, "\" files=\"%d\">\n  <directory_structure>\n%s  </directory_structure>\n", len(files), directoryTree(root, files))
	for _, path := range files {
		data, err := repomixReadFile(path)
		if err != nil {
			return nil, err
		}
		lines := transformRepomixLines(string(data), args)
		rel, _ := filepath.Rel(root, path)
		out.WriteString("  <file path=\"")
		_ = xml.EscapeText(&out, []byte(filepath.ToSlash(rel)))
		fmt.Fprintf(&out, "\" lines=\"%d\">\n    <![CDATA[", len(lines))
		writeCDATA(&out, strings.Join(lines, "\n"))
		out.WriteString("\n]]>\n  </file>\n")
	}
	if instruction, err := readOptionalRegular(instructionPath, 64<<10); err == nil && len(instruction) > 0 {
		out.WriteString("  <instruction>\n    <![CDATA[")
		writeCDATA(&out, string(instruction))
		out.WriteString("\n]]>\n  </instruction>\n")
	}
	out.WriteString("</repository>\n")
	return out.Bytes(), nil
}

func transformRepomixLines(content string, args repomixArgs) []string {
	lines := strings.Split(content, "\n")
	if args.removeEmpty {
		kept := lines[:0]
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				kept = append(kept, line)
			}
		}
		lines = kept
	}
	if args.removeComments {
		joined := stripSimpleComments(strings.Join(lines, "\n"))
		lines = strings.Split(joined, "\n")
	}
	if args.showLines {
		for i := range lines {
			lines[i] = fmt.Sprintf("%6d | %s", i+1, lines[i])
		}
	}
	if args.compress {
		return []string{strings.Join(strings.Fields(strings.Join(lines, "\n")), " ")}
	}
	return lines
}

func stripSimpleComments(content string) string {
	var kept []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if index := strings.Index(line, "//"); index >= 0 && (index == 0 || line[index-1] != ':') {
			line = line[:index]
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func writeCDATA(out *bytes.Buffer, content string) {
	out.WriteString(strings.ReplaceAll(content, "]]>", "]]]]><![CDATA[>"))
}

func directoryTree(root string, files []string) string {
	var out bytes.Buffer
	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		out.WriteString("    ")
		_ = xml.EscapeText(&out, []byte(filepath.ToSlash(rel)))
		out.WriteByte('\n')
	}
	return out.String()
}

func boundedDiagnostic(data []byte) string {
	const limit = 4096
	text := strings.TrimSpace(string(data))
	if len(text) > limit {
		text = text[:limit] + "…"
	}
	if text == "" {
		return "git exited without a diagnostic"
	}
	return text
}
