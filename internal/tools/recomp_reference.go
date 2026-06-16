package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const RecompRefName = "recomp_reference"

var RecompRefDir string

func RecompRefSchema() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        RecompRefName,
			"description": "Fetch and cache a web page locally for offline reading. Use when recomp-foundations skill points to a reference URL. Saves a readable text version to the project cache.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "URL of the reference page to fetch and cache",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}

func RecompRef(refURL string) string {
	if RecompRefDir == "" {
		return "(recomp_reference: output directory not configured)"
	}

	parsed, err := url.Parse(refURL)
	if err != nil || parsed.Host == "" {
		return fmt.Sprintf("(recomp_reference: invalid URL %q)", refURL)
	}

	name := sanitizeName(parsed.Host + "-" + strings.Trim(parsed.Path, "/"))
	if len(name) > 60 {
		name = name[:60]
	}
	outPath := filepath.Join(RecompRefDir, name+".txt")

	if err := os.MkdirAll(RecompRefDir, 0o755); err != nil {
		return fmt.Sprintf("(recomp_reference: mkdir: %v)", err)
	}

	if existing, err := os.Stat(outPath); err == nil && time.Since(existing.ModTime()) < 24*time.Hour {
		return fmt.Sprintf("Cached (from %s)\n  → %s\n\nUse read_file to inspect.", existing.ModTime().Format("2006-01-02 15:04"), outPath)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(refURL)
	if err != nil {
		return fmt.Sprintf("(recomp_reference: fetch failed: %v)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("(recomp_reference: HTTP %d from %s)", resp.StatusCode, parsed.Host)
	}

	contentType := resp.Header.Get("Content-Type")
	var text string

	if strings.Contains(contentType, "text/html") {
		text = extractText(resp.Body)
	} else {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
		if err != nil {
			return fmt.Sprintf("(recomp_reference: read: %v)", err)
		}
		text = string(body)
	}

	text = "# Source: " + refURL + "\n# Fetched: " + time.Now().Format("2006-01-02 15:04") + "\n\n" + text

	if err := os.WriteFile(outPath, []byte(text), 0o644); err != nil {
		return fmt.Sprintf("(recomp_reference: write: %v)", err)
	}

	info, _ := os.Stat(outPath)
	size := ""
	if info != nil {
		size = fmt.Sprintf(" · %s", humanSize(info.Size()))
	}
	return fmt.Sprintf("Fetched %s\n  → %s%s\n\nUse read_file to inspect. Cached for 24 hours.", refURL, outPath, size)
}

func sanitizeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

func extractText(r io.Reader) string {
	doc, err := html.Parse(r)
	if err != nil {
		return "(could not parse HTML)"
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "nav", "footer", "header", "noscript":
				return
			case "br", "p", "div", "li", "tr", "h1", "h2", "h3", "h4", "h5", "h6":
				b.WriteByte('\n')
			}
		}
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				b.WriteString(text)
				b.WriteByte(' ')
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return strings.TrimSpace(b.String())
}
