package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UserServerConfig is one entry from .rehamr/mcp.json (Option B schema).
// Pointer fields indicate whether the JSON explicitly set them, so overrides
// merge cleanly with built-in defaults instead of wiping them.
type UserServerConfig struct {
	Name         string
	Command      *string
	Args         *[]string
	URL          *string
	Tools        *toolList
	ToolsSet     bool
	RequireSkill *bool
}

// toolList represents the "tools" whitelist. It accepts either the string "*"
// or an array of tool names. "*" (or an array containing "*") means allow all.
type toolList struct {
	AllowAll bool
	List     []string
}

func (t *toolList) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "*" {
			t.AllowAll = true
		}
		return nil
	}
	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return fmt.Errorf("tools must be \"*\" or a list of strings: %w", err)
	}
	for _, x := range list {
		if x == "*" {
			t.AllowAll = true
		} else {
			t.List = append(t.List, x)
		}
	}
	return nil
}

// LoadMCPConfig reads .rehamr/mcp.json and returns user-defined or overriding
// server configs. A missing file is not an error; it simply means no overrides.
func LoadMCPConfig(dir string) (map[string]UserServerConfig, error) {
	path := filepath.Join(dir, "mcp.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(stripJSONComments(data), &raw); err != nil {
		return nil, fmt.Errorf("mcp.json: %w", err)
	}

	out := make(map[string]UserServerConfig, len(raw))
	for name, msg := range raw {
		var j struct {
			Command      *string   `json:"command,omitempty"`
			Args         *[]string `json:"args,omitempty"`
			URL          *string   `json:"url,omitempty"`
			Tools        *toolList `json:"tools,omitempty"`
			RequireSkill *bool     `json:"requireSkill,omitempty"`
		}
		if err := json.Unmarshal(msg, &j); err != nil {
			return nil, fmt.Errorf("mcp.json server %q: %w", name, err)
		}

		cfg := UserServerConfig{
			Name:         name,
			Command:      j.Command,
			Args:         j.Args,
			URL:          j.URL,
			Tools:        j.Tools,
			ToolsSet:     j.Tools != nil,
			RequireSkill: j.RequireSkill,
		}
		out[name] = cfg
	}
	return out, nil
}

// stripJSONComments removes // and /* */ comments from JSON data so we can
// use .rehamr/mcp.json with human-readable documentation.
func stripJSONComments(data []byte) []byte {
	var out []byte
	text := string(data)
	inStr := false
	for i := 0; i < len(text); {
		if !inStr && text[i] == '"' {
			inStr = true
			out = append(out, text[i])
			i++
			continue
		}
		if inStr {
			if text[i] == '\\' && i+1 < len(text) {
				out = append(out, text[i], text[i+1])
				i += 2
				continue
			}
			if text[i] == '"' {
				inStr = false
			}
			out = append(out, text[i])
			i++
			continue
		}
		if text[i] == '/' && i+1 < len(text) && text[i+1] == '/' {
			end := strings.IndexByte(text[i:], '\n')
			if end < 0 {
				break
			}
			i += end + 1
			continue
		}
		if text[i] == '/' && i+1 < len(text) && text[i+1] == '*' {
			end := strings.Index(text[i:], "*/")
			if end < 0 {
				break
			}
			i += end + 2
			continue
		}
		out = append(out, text[i])
		i++
	}
	return out
}
