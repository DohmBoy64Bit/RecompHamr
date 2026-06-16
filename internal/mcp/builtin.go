package mcp

import (
	"os"
	"strings"
)

var ghidraDefaultTools = []string{
	"decompile_function",
	"get_xrefs_to",
	"get_function_callers",
	"get_function_callees",
	"search_functions",
	"search_strings",
	"rename_function_by_address",
	"rename_or_label",
	"read_memory",
	"get_function_by_address",
	"list_functions",
	"list_strings",
	"disassemble_function",
	"analyze_function_complete",
	"get_function_signature",
	"get_struct_layout",
	"list_imports",
	"list_exports",
	"search_instructions",
	"analyze_data_region",
}

func parseToolsEnv(val string) []string {
	if val == "" || val == "*" {
		return nil // allow all
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func BuiltinServers() []ServerConfig {
	ghidraCmd := os.Getenv("RECOMPHAMR_MCP_GHIDRA_COMMAND")
	if ghidraCmd == "" {
		ghidraCmd = "ghidra-mcp"
	}

	n64Cmd := os.Getenv("RECOMPHAMR_MCP_N64_COMMAND")
	if n64Cmd == "" {
		n64Cmd = "n64-debug-mcp"
	}

	ghidraTools := ghidraDefaultTools
	if env := os.Getenv("RECOMPHAMR_MCP_GHIDRA_TOOLS"); env == "*" {
		ghidraTools = nil // allow all
	} else if env != "" {
		ghidraTools = parseToolsEnv(env)
	}

	return []ServerConfig{
		{
			Name:         "ghidra",
			Command:      ghidraCmd,
			Args:         []string{},
			AllowedTools: ghidraTools,
			RequireSkill: true,
		},
		{
			Name:         "n64-debug-mcp",
			Command:      n64Cmd,
			Args:         []string{},
			RequireSkill: true,
		},
	}
}
