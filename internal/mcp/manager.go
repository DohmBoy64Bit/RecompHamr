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
	Command      string
	Args         []string
	AllowedTools []string // nil = allow all, slice = whitelist
	RequireSkill bool     // true = tools only injected when matching skill is active
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
	client       *Client
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
	if entry.state == StateConnected {
		m.mu.Unlock()
		return nil
	}
	entry.state = StateConnecting
	entry.err = ""
	entry.client = NewClient("recomphamr", "0.2.0")
	m.mu.Unlock()

	err := entry.client.Connect(ctx, entry.config.Command, entry.config.Args...)

	m.mu.Lock()
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
		if entry.client.serverInfo.Name != "" {
			version = entry.client.serverInfo.Version
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
	allowed := map[string]bool{}
	for _, s := range activeSkills {
		if server, ok := SkillServers[s]; ok {
			allowed[server] = true
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
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
