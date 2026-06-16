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

func BuiltinServers() []ServerConfig {
	ghidraCmd := os.Getenv("RECOMPHAMR_MCP_GHIDRA_COMMAND")
	if ghidraCmd == "" {
		ghidraCmd = "ghidra-mcp"
	}

	n64Cmd := os.Getenv("RECOMPHAMR_MCP_N64_COMMAND")
	if n64Cmd == "" {
		n64Cmd = "n64-debug-mcp"
	}

	pcrecompCmd := os.Getenv("RECOMPHAMR_MCP_PCRECOMP_COMMAND")
	if pcrecompCmd == "" {
		pcrecompCmd = "pcrecomp-mcp"
	}

	pineCmd := os.Getenv("RECOMPHAMR_MCP_PINE_COMMAND")
	if pineCmd == "" {
		pineCmd = "mcp-pine"
	}

	objdiffCmd := os.Getenv("RECOMPHAMR_MCP_OBJDIFF_COMMAND")
	if objdiffCmd == "" {
		objdiffCmd = "objdiff-mcp"
	}

	pcsx2Cmd := os.Getenv("RECOMPHAMR_MCP_PCSX2_COMMAND")
	if pcsx2Cmd == "" {
		pcsx2Cmd = "pcsx2-mcp"
	}

	bizhawkCmd := os.Getenv("RECOMPHAMR_MCP_BIZHAWK_COMMAND")
	if bizhawkCmd == "" {
		bizhawkCmd = "mcp-bizhawk"
	}

	sega2asmCmd := os.Getenv("RECOMPHAMR_MCP_SEGA2ASM_COMMAND")
	if sega2asmCmd == "" {
		sega2asmCmd = "sega2asm-mcp"
	}

	ghidraTools := ghidraDefaultTools
	if env := os.Getenv("RECOMPHAMR_MCP_GHIDRA_TOOLS"); env == "*" {
		ghidraTools = nil
	} else if env != "" {
		ghidraTools = parseToolsEnv(env)
	}

	pcrecompTools := pcrecompDefaultTools
	if env := os.Getenv("RECOMPHAMR_MCP_PCRECOMP_TOOLS"); env == "*" {
		pcrecompTools = nil
	} else if env != "" {
		pcrecompTools = parseToolsEnv(env)
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
		{
			Name:         "pcrecomp",
			Command:      pcrecompCmd,
			Args:         []string{},
			AllowedTools: pcrecompTools,
			RequireSkill: true,
		},
		{
			Name:         "mcp-pine",
			Command:      pineCmd,
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "objdiff",
			Command:      objdiffCmd,
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "pcsx2",
			Command:      pcsx2Cmd,
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "bizhawk",
			Command:      bizhawkCmd,
			Args:         []string{},
			RequireSkill: true,
		},
		{
			Name:         "sega2asm",
			Command:      sega2asmCmd,
			Args:         []string{},
			RequireSkill: true,
		},
	}
}
