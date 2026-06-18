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

var pcrecompDefaultTools = []string{
	"pe.analyze",
	"pe.extract_imports",
	"disasm32.run",
	"disasm32.callgraph",
	"lift32.run",
	"classify.run",
	"ghidra.decompile_all",
	"ghidra.function_stats",
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

// legacyEnvSuffix maps server names whose env var suffix doesn't match the
// mechanical derivation (replace hyphens with underscores, uppercase). The
// old code used explicit os.Getenv calls with these shorter names; this map
// preserves backward compatibility so existing env var setups continue to work.
var legacyEnvSuffix = map[string]string{
	"n64-debug-mcp": "N64",
	"mcp-pine":      "PINE",
}

// envKey returns the RECOMPHAMR_MCP_<KEY>_* suffix for a server name.
func envKey(name string) string {
	if suffix, ok := legacyEnvSuffix[name]; ok {
		return suffix
	}
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}

// applyEnvOverrides mutates a server config from RECOMPHAMR_MCP_<NAME>_*
// environment variables. Empty env vars leave the config unchanged.
func applyEnvOverrides(name string, cfg *ServerConfig) {
	key := envKey(name)
	if v := os.Getenv("RECOMPHAMR_MCP_" + key + "_COMMAND"); v != "" {
		cfg.Command = v
		cfg.URL = ""
	}
	if v := os.Getenv("RECOMPHAMR_MCP_" + key + "_URL"); v != "" {
		cfg.URL = v
		cfg.Command = ""
		cfg.Args = nil
	}
	if v := os.Getenv("RECOMPHAMR_MCP_" + key + "_TOOLS"); v != "" {
		cfg.AllowedTools = parseToolsEnv(v)
	}
}

func BuiltinServers() []ServerConfig {
	servers := []ServerConfig{
		{
			Name:         "ghidra",
			Command:      "ghidra-mcp",
			Args:         []string{},
			AllowedTools: ghidraDefaultTools,
			RequireSkill: true,
		},
		{
			Name:         "n64-debug-mcp",
			Command:      "n64-debug-mcp",
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "pcrecomp",
			Command:      "pcrecomp-mcp",
			Args:         []string{},
			AllowedTools: pcrecompDefaultTools,
			RequireSkill: true,
		},
		{
			Name:         "mcp-pine",
			Command:      "mcp-pine",
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "objdiff",
			Command:      "objdiff-mcp",
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "pcsx2",
			Command:      "pcsx2-mcp",
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "bizhawk",
			Command:      "mcp-bizhawk",
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "sega2asm",
			Command:      "sega2asm-mcp",
			Args:         []string{},
			RequireSkill: true,
		},
	}
	for i := range servers {
		applyEnvOverrides(servers[i].Name, &servers[i])
	}
	return servers
}
