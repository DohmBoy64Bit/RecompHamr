package mcp

import "os"

func BuiltinServers() []ServerConfig {
	ghidraCmd := os.Getenv("RECOMPHAMR_MCP_GHIDRA_COMMAND")
	if ghidraCmd == "" {
		ghidraCmd = "ghidra-mcp"
	}

	n64Cmd := os.Getenv("RECOMPHAMR_MCP_N64_COMMAND")
	if n64Cmd == "" {
		n64Cmd = "n64-debug-mcp"
	}

	return []ServerConfig{
		{
			Name:    "ghidra",
			Command: ghidraCmd,
			Args:    []string{},
		},
		{
			Name:    "n64-debug-mcp",
			Command: n64Cmd,
			Args:    []string{},
		},
	}
}
