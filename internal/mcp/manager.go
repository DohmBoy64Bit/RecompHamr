package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var SkillServers = map[string]string{
	"ghidra-mcp":      "ghidra",
	"n64-debug-mcp":   "n64-debug-mcp",
	"pcrecomp":        "pcrecomp",
	"mcp-pine":        "mcp-pine",
	"objdiff":         "objdiff",
	"pcsx2":           "pcsx2",
	"bizhawk":          "bizhawk",
	"sega2asm":         "sega2asm",
}

type ServerState int

const (
	StateDisconnected ServerState = iota
	StateConnecting
	StateConnected
	StateError
)

func (s ServerState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

type ServerConfig struct {
	Name         string
	Command      string   // command to launch (stdio transport)
	Args         []string // arguments for Command
	URL          string   // HTTP endpoint URL (streamable-http transport)
	AllowedTools []string // nil = allow all, slice = whitelist
	RequireSkill bool     // true = tools only injected when matching skill is active
}

// mcpClient is the common interface for both stdio and HTTP MCP clients.
type mcpClient interface {
	Connected() bool
	Name() string
	Version() string
	ServerName() string
	Tools() []ToolDef
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error)
	Disconnect()
}

type ServerStatus struct {
	Name    string
	State   ServerState
	Tools   int
	Err     string
	Version string
}

type Manager struct {
	mu      sync.Mutex
	servers map[string]*serverEntry
}

type serverEntry struct {
	config       ServerConfig
	client       mcpClient
	state        ServerState
	err          string
	allowedTools map[string]bool // nil = allow all
}

func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]*serverEntry),
	}
}

func (m *Manager) Register(config ServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.servers[config.Name]; !ok {
		entry := &serverEntry{
			config: config,
			state:  StateDisconnected,
		}
		if len(config.AllowedTools) > 0 {
			entry.allowedTools = make(map[string]bool, len(config.AllowedTools))
			for _, t := range config.AllowedTools {
				entry.allowedTools[t] = true
			}
		}
		m.servers[config.Name] = entry
	}
}

func (m *Manager) Connect(ctx context.Context, name string) error {
	m.mu.Lock()
	entry, ok := m.servers[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("mcp: unknown server %q", name)
	}
	if entry.state == StateConnected || entry.state == StateConnecting {
		m.mu.Unlock()
		return nil
	}
	entry.state = StateConnecting
	entry.err = ""
	config := entry.config // copy under lock
	m.mu.Unlock()

	var err error
	if config.URL != "" {
		client := NewHTTPClient("recomphamr", "0.2.0")
		err = client.Connect(ctx, config.URL)
		m.mu.Lock()
		if err == nil {
			entry.client = client
		}
	} else {
		client := NewClient("recomphamr", "0.2.0")
		err = client.Connect(ctx, config.Command, config.Args...)
		m.mu.Lock()
		if err == nil {
			entry.client = client
		}
	}

	if err != nil {
		entry.state = StateError
		entry.err = err.Error()
	} else {
		entry.state = StateConnected
		entry.err = ""
	}
	m.mu.Unlock()

	return err
}

func (m *Manager) Disconnect(name string) {
	m.mu.Lock()
	entry, ok := m.servers[name]
	if !ok {
		m.mu.Unlock()
		return
	}
	if entry.client != nil {
		entry.client.Disconnect()
	}
	entry.state = StateDisconnected
	entry.err = ""
	m.mu.Unlock()
}

func (m *Manager) ConnectAll(ctx context.Context) []ServerStatus {
	m.mu.Lock()
	names := make([]string, 0, len(m.servers))
	for _, e := range m.servers {
		names = append(names, e.config.Name)
	}
	m.mu.Unlock()

	sort.Strings(names)
	results := make([]ServerStatus, 0, len(names))

	for _, name := range names {
		if err := m.Connect(ctx, name); err != nil {
			m.mu.Lock()
			if e, ok := m.servers[name]; ok {
				e.err = err.Error()
			}
			m.mu.Unlock()
		}
		results = append(results, m.Status(name))
	}

	return results
}

func (m *Manager) Status(name string) ServerStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.servers[name]
	if !ok {
		return ServerStatus{Name: name, State: StateDisconnected}
	}
	toolCount := 0
	version := ""
	if entry.client != nil && entry.state == StateConnected {
		toolCount = m.toolCount(entry)
		if entry.config.URL == "" {
			if c, ok := entry.client.(*Client); ok && c.serverInfo.Name != "" {
				version = c.serverInfo.Version
			}
		} else {
			if c, ok := entry.client.(*HTTPClient); ok && c.serverInfo.Name != "" {
				version = c.serverInfo.Version
			}
		}
	}
	return ServerStatus{
		Name:    entry.config.Name,
		State:   entry.state,
		Tools:   toolCount,
		Err:     entry.err,
		Version: version,
	}
}

func (m *Manager) AllStatus() []ServerStatus {
	m.mu.Lock()
	names := make([]string, 0, len(m.servers))
	for n := range m.servers {
		names = append(names, n)
	}
	sort.Strings(names)
	m.mu.Unlock()

	result := make([]ServerStatus, 0, len(names))
	for _, n := range names {
		result = append(result, m.Status(n))
	}
	return result
}

func (m *Manager) AllTools() []ToolDef {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []ToolDef
	for _, entry := range m.servers {
		if entry.state != StateConnected || entry.client == nil {
			continue
		}
		if entry.config.RequireSkill {
			continue
		}
		prefix := entry.config.Name + "."
		for _, t := range entry.client.Tools() {
			if !entry.toolAllowed(t.Name) {
				continue
			}
			all = append(all, ToolDef{
				Name:        prefix + t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			})
		}
	}
	return all
}

func (m *Manager) ToolsForSkills(activeSkills []string) []ToolDef {
	m.mu.Lock()
	defer m.mu.Unlock()

	allowed := map[string]bool{}
	for _, s := range activeSkills {
		if server, ok := SkillServers[s]; ok {
			allowed[server] = true
		}
		if _, ok := m.servers[s]; ok {
			allowed[s] = true
		}
	}
	var out []ToolDef
	for _, entry := range m.servers {
		if entry.state != StateConnected || entry.client == nil {
			continue
		}
		if entry.config.RequireSkill {
			if !allowed[entry.config.Name] {
				continue
			}
		}
		prefix := entry.config.Name + "."
		for _, t := range entry.client.Tools() {
			if !entry.toolAllowed(t.Name) {
				continue
			}
			out = append(out, ToolDef{
				Name:        prefix + t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			})
		}
	}
	return out
}

func (e *serverEntry) toolAllowed(name string) bool {
	if e.allowedTools == nil {
		return true
	}
	return e.allowedTools[name]
}

// syncAllowedTools rebuilds the runtime allowedTools map from config.AllowedTools.
// Must be called under the manager lock after any mutation to config.AllowedTools.
func (e *serverEntry) syncAllowedTools() {
	if e.config.AllowedTools == nil {
		e.allowedTools = nil // allow all
	} else {
		e.allowedTools = make(map[string]bool, len(e.config.AllowedTools))
		for _, t := range e.config.AllowedTools {
			e.allowedTools[t] = true
		}
	}
}

func (m *Manager) toolCount(entry *serverEntry) int {
	if entry.client == nil {
		return 0
	}
	n := 0
	for _, t := range entry.client.Tools() {
		if entry.toolAllowed(t.Name) {
			n++
		}
	}
	return n
}

// ApplyUserConfig merges a user-supplied config from .rehamr/mcp.json into an
// existing server entry, or registers a new server if the name is unknown.
// Only explicitly set fields are merged; omitted fields keep their current
// values (built-in defaults for known servers, zero values for new ones).
func (m *Manager) ApplyUserConfig(cfg UserServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.servers[cfg.Name]
	if !ok {
		newCfg := ServerConfig{Name: cfg.Name}
		if cfg.Command != nil {
			newCfg.Command = *cfg.Command
			if cfg.Args != nil {
				newCfg.Args = *cfg.Args
			}
		}
		if cfg.URL != nil {
			newCfg.URL = *cfg.URL
		}
		if cfg.RequireSkill != nil {
			newCfg.RequireSkill = *cfg.RequireSkill
		}
		if cfg.ToolsSet {
			if cfg.Tools != nil && cfg.Tools.AllowAll {
				newCfg.AllowedTools = nil
			} else if cfg.Tools != nil {
				newCfg.AllowedTools = cfg.Tools.List
			}
		}
		entry = &serverEntry{config: newCfg}
		entry.syncAllowedTools()
		m.servers[cfg.Name] = entry
		return
	}

	if cfg.Command != nil {
		entry.config.Command = *cfg.Command
		entry.config.Args = nil
		if cfg.Args != nil {
			entry.config.Args = *cfg.Args
		}
		entry.config.URL = "" // stdio and HTTP are mutually exclusive
	}
	if cfg.URL != nil {
		entry.config.URL = *cfg.URL
		entry.config.Command = ""
		entry.config.Args = nil
	}
	if cfg.RequireSkill != nil {
		entry.config.RequireSkill = *cfg.RequireSkill
	}
	if cfg.ToolsSet {
		if cfg.Tools != nil && cfg.Tools.AllowAll {
			entry.config.AllowedTools = nil
		} else if cfg.Tools != nil {
			entry.config.AllowedTools = cfg.Tools.List
		}
		entry.syncAllowedTools()
	}
}

// ApplyEnvOverrides re-applies RECOMPHAMR_MCP_<NAME>_* environment variables
// to every registered server. This lets env vars override both built-in
// defaults and .rehamr/mcp.json settings.
func (m *Manager) ApplyEnvOverrides() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.servers {
		applyEnvOverrides(entry.config.Name, &entry.config)
		entry.syncAllowedTools()
	}
}

func (m *Manager) ConnectedNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var names []string
	for _, entry := range m.servers {
		if entry.state == StateConnected {
			names = append(names, entry.config.Name)
		}
	}
	sort.Strings(names)
	return names
}

func (m *Manager) CallTool(ctx context.Context, fullName string, args map[string]interface{}) (*CallToolResult, error) {
	parts := strings.SplitN(fullName, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("mcp: invalid tool name %q (expected server.tool)", fullName)
	}
	serverName, toolName := parts[0], parts[1]

	m.mu.Lock()
	entry, ok := m.servers[serverName]
	if !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("mcp: unknown server %q", serverName)
	}
	if entry.state != StateConnected || entry.client == nil {
		m.mu.Unlock()
		return nil, fmt.Errorf("mcp: server %q is not connected", serverName)
	}
	client := entry.client
	m.mu.Unlock()

	return client.CallTool(ctx, toolName, args)
}

func (m *Manager) SetToolEnabled(serverName, toolName string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.servers[serverName]
	if !ok {
		return fmt.Errorf("mcp: unknown server %q", serverName)
	}
	if enabled {
		if entry.allowedTools == nil {
			entry.allowedTools = make(map[string]bool)
		}
		entry.allowedTools[toolName] = true
	} else {
		if entry.allowedTools == nil {
			entry.allowedTools = make(map[string]bool)
			if entry.client != nil {
				for _, t := range entry.client.Tools() {
					entry.allowedTools[t.Name] = true
				}
			}
		}
		delete(entry.allowedTools, toolName)
	}
	return nil
}

func (m *Manager) SetAllToolsEnabled(serverName string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.servers[serverName]
	if !ok {
		return fmt.Errorf("mcp: unknown server %q", serverName)
	}
	if enabled {
		entry.allowedTools = nil
	} else {
		entry.allowedTools = map[string]bool{}
	}
	return nil
}

func (m *Manager) FormatTools(serverName string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "MCP tools for %s:\n", serverName)

	m.mu.Lock()
	entry, ok := m.servers[serverName]
	if !ok {
		m.mu.Unlock()
		return fmt.Sprintf("unknown server %q", serverName)
	}
	client := entry.client
	allowed := entry.allowedTools
	m.mu.Unlock()

	if client == nil {
		return fmt.Sprintf("server %q is not connected", serverName)
	}

	for _, t := range client.Tools() {
		mark := "  "
		if allowed == nil || allowed[t.Name] {
			mark = " *"
		}
		fmt.Fprintf(&b, "%s %s - %s\n", mark, t.Name, firstSentence(t.Description))
	}
	return b.String()
}

func firstSentence(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' && (i+1 >= len(s) || s[i+1] == ' ') {
			return s[:i+1]
		}
	}
	return s
}

func (m *Manager) FormatStatus() string {
	var b strings.Builder
	b.WriteString("MCP Servers:\n")
	for _, s := range m.AllStatus() {
		icon := "  "
		switch s.State {
		case StateConnected:
			icon = " *"
		case StateConnecting:
			icon = " ~"
		case StateError:
			icon = " !"
		}
		extra := ""
		if s.State == StateConnected {
			extra = fmt.Sprintf(" (%d tools)", s.Tools)
		} else if s.State == StateError && s.Err != "" {
			extra = " - " + s.Err
		}
		fmt.Fprintf(&b, "%s %s %s%s\n", icon, s.Name, s.State, extra)
	}
	return b.String()
}
