package tui

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/recomphamr/internal/classifier"
	"github.com/DohmBoy64Bit/recomphamr/internal/cloud"
	"github.com/DohmBoy64Bit/recomphamr/internal/config"
	"github.com/DohmBoy64Bit/recomphamr/internal/doctor"
	"github.com/DohmBoy64Bit/recomphamr/internal/llm"
	"github.com/DohmBoy64Bit/recomphamr/internal/project"
	"github.com/DohmBoy64Bit/recomphamr/internal/skills"
)

// argOption is one popover entry, used at command-level (one row per command)
// and argument-level (one row per accepted value for the active command).
type argOption struct {
	value       string // what gets inserted / committed
	description string // right-aligned help text
	current     bool   // rendered bold; default-selected when the popover opens
}

// command is one row in the popover, --help, and the dispatch table.
// args, if non-nil, supplies the argument-level popover entries.
type command struct {
	name        string
	description string
	handler     func(Model, []string) (tea.Model, tea.Cmd)
	args        func(Model) []argOption
}

// commands lists every slash command, in popover/--help order. Keep it short.
var commands = []command{
	{
		name:        "/clear",
		description: "reset the conversation",
		handler:     (Model).cmdClear,
	},
	{
		name:        "/models",
		description: "list · <name> set (Tab cycles in the popover)",
		handler:     (Model).cmdModel,
		args: func(m Model) []argOption {
			out := make([]argOption, 0, len(m.cfg.Models))
			for _, n := range m.cfg.ModelNames() {
				p := m.cfg.Models[n]
				out = append(out, argOption{
					value:       n,
					description: p.LLM + " @ " + p.URL,
					current:     n == m.cfg.Active,
				})
			}
			return out
		},
	},
	{
		name:        "/skills",
		description: "list built-in and custom RE skills",
		handler:     (Model).cmdSkills,
	},
	{
		name:        "/skill",
		description: "load a skill by name (Tab for list)",
		handler:     (Model).cmdSkill,
		args: func(m Model) []argOption {
			names := skills.Names()
			out := make([]argOption, 0, len(names))
			for _, n := range names {
				active := false
				for _, a := range m.activeSkills {
					if strings.EqualFold(a, n) {
						active = true
						break
					}
				}
				out = append(out, argOption{
					value:   n,
					current: active,
				})
			}
			return out
		},
	},
	{
		name:        "/skill-audit",
		description: "classify a skill and suggest a template",
		handler:     (Model).cmdSkillAudit,
		args: func(m Model) []argOption {
			names := skills.Names()
			out := make([]argOption, 0, len(names))
			for _, n := range names {
				out = append(out, argOption{value: n})
			}
			return out
		},
	},
	{
		name:        "/skill-new",
		description: "fetch a URL and classify as a new skill",
		handler:     (Model).cmdSkillNew,
	},
	{
		name:        "/init-re",
		description: "create .rehamr/ evidence workspace",
		handler:     (Model).cmdInitRE,
	},
	{
		name:        "/status-re",
		description: "summarize RE project state",
		handler:     (Model).cmdStatusRE,
	},
	{
		name:        "/doctor",
		description: "run environment diagnostics",
		handler:     (Model).cmdDoctor,
	},
	{
		name:        "/mcp",
		description: "show or manage MCP servers and tools",
		handler:     (Model).cmdMcp,
	},
	{
		name:        "/help",
		description: "show this help",
		handler:     (Model).cmdHelp,
	},
}

// commandByName returns the registered command for a slash name, or nil.
// Centralises the linear scan shared by completion, dispatch, and runSlash.
func commandByName(name string) *command {
	for i := range commands {
		if commands[i].name == name {
			return &commands[i]
		}
	}
	return nil
}

// runSlash dispatches a slash-prefixed submission; unknown commands produce a
// quiet hint, not an error. config.yaml is re-read before every slash so
// hand-edits take effect without a restart (see reloadConfigFromDisk).
func (m Model) runSlash(text string) (tea.Model, tea.Cmd) {
	if err := m.reloadConfigFromDisk(); err != nil {
		m.appendLine(styleWarn.Render("⚠ " + err.Error()))
	}
	fields := strings.Fields(text)
	if c := commandByName(fields[0]); c != nil {
		return c.handler(m, fields[1:])
	}
	m.appendLine(styleWarn.Render("unknown command - type / to see options"))
	return m, nil
}

// reloadConfigFromDisk re-runs config.Bootstrap and replaces m.cfg so hand-edits
// to config.yaml between slash commands take effect immediately. URLOverride
// (from RECOMPHAMR_URL) is carried across the swap so the env var keeps applying.
//
// Returns the Bootstrap error verbatim; callers decide whether to surface it
// (runSlash warns on submit; the popover-refresh path ignores it so a broken
// file doesn't spam a warning on every keystroke).
//
// Rebuilds the llm.Client when the active profile's resolved (URL, model, key)
// triple changed: covers both within-profile edits and a moved active.
func (m *Model) reloadConfigFromDisk() error {
	projectRoot := filepath.Dir(m.cfg.Dir)
	fresh, _, err := config.Bootstrap(projectRoot)
	if err != nil {
		return err
	}
	fresh.URLOverride = m.cfg.URLOverride

	prevURL := m.cfg.ActiveURL()
	prevProfile := m.cfg.ActiveProfile()
	prevLLM, prevKey := prevProfile.LLM, prevProfile.Key

	m.cfg = fresh

	newProfile := m.cfg.ActiveProfile()
	if prevURL != m.cfg.ActiveURL() || prevLLM != newProfile.LLM || prevKey != newProfile.Key {
		m.rebuildClient()
	}
	return nil
}

// PrintHelp writes the canonical human-readable command list. Used by --help.
func PrintHelp(out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, c := range commands {
		fmt.Fprintf(w, "  %s\t%s\n", c.name, c.description)
	}
	w.Flush()
}

// --- handlers ---------------------------------------------------------------

// cmdModel: `/models` lists, `/models <name>` sets. Cycling is Tab/Shift+Tab
// in the popover, no separate "next" command.
func (m Model) cmdModel(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		m.printModelList()
		return m, nil
	}
	if err := m.cfg.SetActive(args[0]); err != nil {
		m.appendLine(styleError.Render("⚠ " + err.Error()))
		return m, nil
	}
	m.rebuildClient()
	return m, m.confirmActive(args[0])
}

// printModelList writes the "▸ active, name, llm @ url" rollup to scroll.
func (m *Model) printModelList() {
	m.appendLine(styleDim.Render("models (▸ active, /models <name> to switch):"))
	for _, n := range m.cfg.ModelNames() {
		mark := "  "
		if n == m.cfg.Active {
			mark = "▸ "
		}
		p := m.cfg.Models[n]
		m.appendLine(fmt.Sprintf("%s%s  %s",
			mark, n, styleDim.Render(p.LLM+" @ "+p.URL)))
	}
}

// confirmActive emits the activation line for the active profile and returns
// its reachability cmd. Keyed profiles (cloud) probe: the success line is
// delayed until the response arrives so it can carry the live ctx window from
// X-Context-Window. Keyless profiles (local Ollama) ping and print
// synchronously. Shared by /models.
func (m *Model) confirmActive(profile string) tea.Cmd {
	p := m.cfg.ActiveProfile()
	if p.Key != "" {
		m.appendLine(styleDim.Render(fmt.Sprintf("▶ probing %s · %s @ %s", profile, p.LLM, p.URL)))
		return probeBackend(m.cli, profile, false)
	}
	m.appendLine(styleOK.Render(fmt.Sprintf("✓ active: %s · %s @ %s", profile, p.LLM, p.URL)))
	return pingBackend(m.cli.BaseURL)
}

// rebuildClient swaps in a fresh llm.Client for the now-active profile.
// Replacing the pointer (not mutating fields) drops the prior Client's sticky
// state (noReasoningEffort, keep-alive pool tied to the old URL): new
// endpoint, fresh slate.
func (m *Model) rebuildClient() {
	p := m.cfg.ActiveProfile()
	m.cli = llm.New(m.cfg.ActiveURL(), p.LLM, p.Key)
	// Drop the prior profile's cached BudgetStatus. m.budget has no profile
	// association, so without this reset the footer keeps rendering the old
	// "88% pass" segment after switching to a local profile that emits no
	// X-Budget-* headers (nothing would overwrite it). A fresh BudgetStatus{}
	// hides the segment until the new backend reports its own.
	m.budget = cloud.BudgetStatus{}
}

func (m Model) cmdClear(_ []string) (tea.Model, tea.Cmd) {
	m.history = nil
	m.scroll.Reset()
	m.sessionTokens = 0
	m.streamingEstimate = 0
	// Reset the repeated-failure streak so the next turn starts clean.
	m.failKey, m.failStreak = "", 0
	// Wipe prompt recall too: in-memory ring and on-disk .rehamr/history,
	// or leftover history would contradict the "fresh start" promise.
	m.promptHistory = nil
	m.histIdx = -1
	_ = clearPromptHistory(m.cfg.Dir)
	// Full wipe (unlike Ctrl+L, which redraws but keeps scrollback).
	// tea.ClearScreen emits \x1b[2J, which only wipes the viewport; the
	// saved-lines buffer needs eraseScrollback (DECSED 3) too, or old replies
	// stay scrollable above the reset line. tea.Sequence keeps the print from
	// racing past the clear (tea.Batch runs both concurrently and the print
	// could land first, then get wiped). scroll keeps the line for resize
	// replay; outbox is cleared because the Sequence owns the print now.
	line := styleOK.Render("✓ conversation reset")
	m.scroll.WriteString(line + "\n")
	m.outbox = nil
	return m, tea.Sequence(tea.ClearScreen, eraseScrollback, tea.Println(line))
}

// ---------------------------------------------------------------------------
// RE slash command handlers
// ---------------------------------------------------------------------------

func (m Model) cmdSkills(_ []string) (tea.Model, tea.Cmd) {
	m.appendLine(skills.ListMarkdown(m.activeSkills))
	return m, nil
}

func (m Model) cmdSkill(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		m.appendLine(styleError.Render("usage: /skill <name>"))
		return m, nil
	}
	name, err := skills.Resolve(args[0])
	if err != nil {
		m.appendLine(styleError.Render("unknown skill: " + args[0]))
		return m, nil
	}
	for _, s := range m.activeSkills {
		if strings.EqualFold(s, name) {
			m.appendLine(fmt.Sprintf("skill %s is already active", name))
			return m, nil
		}
	}
	m.activeSkills = append(m.activeSkills, name)
	m.system = m.rebuildSystem()
	m.appendLine(styleOK.Render("loaded skill: " + name))
	return m, nil
}

func (m Model) cmdInitRE(_ []string) (tea.Model, tea.Cmd) {
	err := project.InitRE(m.ProjectDir)
	if err != nil {
		m.appendLine(styleError.Render("init-re error: " + err.Error()))
		return m, nil
	}
	m.appendLine(styleOK.Render(".rehamr/ evidence workspace initialized"))
	m.appendLine(styleDim.Render("  Use /status-re to check project state"))
	return m, nil
}

func (m Model) cmdStatusRE(_ []string) (tea.Model, tea.Cmd) {
	rehamrDir := filepath.Join(m.ProjectDir, config.DirName)
	status, err := project.StatusRE(rehamrDir)
	if err != nil {
		m.appendLine(styleError.Render("status-re: " + err.Error()))
		return m, nil
	}
	m.appendLine(status)
	return m, nil
}

func (m Model) cmdSkillAudit(args []string) (tea.Model, tea.Cmd) {
	if len(args) < 1 {
		m.appendLine(styleWarn.Render("usage: /skill-audit <skill-name>"))
		return m, nil
	}
	name, err := skills.Resolve(args[0])
	if err != nil {
		m.appendLine(styleError.Render("unknown skill: " + args[0]))
		return m, nil
	}
	body, err := skills.Get(name)
	if err != nil {
		m.appendLine(styleError.Render("failed to read skill: " + err.Error()))
		return m, nil
	}
	r := classifier.Classify(name, body)

	m.appendLine(styleOK.Render(fmt.Sprintf("skill-audit: %s", name)))
	m.appendLine(styleDim.Render(""))
	m.appendLine(fmt.Sprintf("  Class:        %s", r.Class.TemplateName()))
	m.appendLine(fmt.Sprintf("  Confidence:   %.0f%%", r.Confidence*100))
	m.appendLine(styleDim.Render("  ── Feature Scores ──"))
	m.appendLine(fmt.Sprintf("  Full Workflow: %3d", r.Scores[classifier.FullWorkflow]))
	m.appendLine(fmt.Sprintf("  Micro-Skill:   %3d", r.Scores[classifier.MicroSkill]))
	m.appendLine(fmt.Sprintf("  Tool Bridge:   %3d", r.Scores[classifier.ToolBridge]))
	m.appendLine(styleDim.Render("  ── Reasoning ──"))
	for _, re := range r.Reasoning {
		m.appendLine(styleDim.Render(fmt.Sprintf("  · %s", re)))
	}
	if len(r.Alternatives) > 0 {
		m.appendLine(styleDim.Render("  ── Alternatives ──"))
		for _, a := range r.Alternatives {
			m.appendLine(styleDim.Render(fmt.Sprintf("  · %s (score: %d)", a.TemplateName(), r.Scores[a])))
		}
	}
	if r.Class == classifier.NoneClass {
		m.appendLine("")
		m.appendLine(styleWarn.Render("No matching template. Edit recomphamr_skill_audit_and_template.md to add a new template class, then re-run /skill-audit."))
	}
	return m, nil
}

func (m Model) cmdSkillNew(args []string) (tea.Model, tea.Cmd) {
	usage := func() { m.appendLine(styleWarn.Render("usage: /skill-new <url>")) }

	if len(args) < 1 || args[0] == "" {
		usage()
		return m, nil
	}
	rawURL := args[0]
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		m.appendLine(styleError.Render(fmt.Sprintf("invalid URL: %s", rawURL)))
		return m, nil
	}

	// Fetch with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		m.appendLine(styleError.Render("request failed: " + err.Error()))
		return m, nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.appendLine(styleError.Render("fetch failed: " + err.Error()))
		return m, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		m.appendLine(styleError.Render(fmt.Sprintf("HTTP %d", resp.StatusCode)))
		return m, nil
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 100_000))
	if err != nil {
		m.appendLine(styleError.Render("read failed: " + err.Error()))
		return m, nil
	}
	content := string(raw)
	if len(content) < 60 {
		m.appendLine(styleWarn.Render("fetched content too short to classify"))
		return m, nil
	}

	// Derive skill name.
	name := skillNameFromURL(parsed)
	m.appendLine(styleDim.Render(fmt.Sprintf("Fetched %d bytes → %s", len(content), name)))

	// Classify.
	r := classifier.Classify(name, content)

	// Format output.
	m.appendLine(styleOK.Render(fmt.Sprintf("skill-new: %s", name)))
	m.appendLine(styleDim.Render(fmt.Sprintf("  URL: %s", rawURL)))
	m.appendLine(styleDim.Render(""))
	m.appendLine(fmt.Sprintf("  Class:        %s", r.Class.TemplateName()))
	m.appendLine(fmt.Sprintf("  Confidence:   %.0f%%", r.Confidence*100))
	m.appendLine(styleDim.Render("  ── Feature Scores ──"))
	m.appendLine(fmt.Sprintf("  Full Workflow: %3d", r.Scores[classifier.FullWorkflow]))
	m.appendLine(fmt.Sprintf("  Micro-Skill:   %3d", r.Scores[classifier.MicroSkill]))
	m.appendLine(fmt.Sprintf("  Tool Bridge:   %3d", r.Scores[classifier.ToolBridge]))
	m.appendLine(styleDim.Render("  ── Reasoning ──"))
	for _, re := range r.Reasoning {
		m.appendLine(styleDim.Render(fmt.Sprintf("  · %s", re)))
	}
	if len(r.Alternatives) > 0 {
		m.appendLine(styleDim.Render("  ── Alternatives ──"))
		for _, a := range r.Alternatives {
			m.appendLine(styleDim.Render(fmt.Sprintf("  · %s (score: %d)", a.TemplateName(), r.Scores[a])))
		}
	}
	if r.Class == classifier.NoneClass {
		m.appendLine("")
		m.appendLine(styleWarn.Render("No matching template. Reply to suggest a new template class, or choose one manually."))
		return m, nil
	}

	// Save fetched content to .rehamr/fetched/<name>.md so the AI can read it
	// back when creating the final skill file.
	fetchedDir := filepath.Join(m.ProjectDir, ".rehamr", "fetched")
	if err := os.MkdirAll(fetchedDir, 0o755); err != nil {
		m.appendLine(styleWarn.Render("cache dir: " + err.Error()))
	} else {
		fetchedPath := filepath.Join(fetchedDir, name+".md")
		if err := os.WriteFile(fetchedPath, raw, 0o644); err != nil {
			m.appendLine(styleWarn.Render("cache save: " + err.Error()))
		} else {
			m.appendLine(styleDim.Render(fmt.Sprintf("  Cached: %s", fetchedPath)))
		}
	}
	m.appendLine("")
	m.appendLine(styleDim.Render("Ask the user to confirm the classification. If confirmed, read the"))
	m.appendLine(styleDim.Render("correct template section from recomphamr_skill_audit_and_template.md,"))
	m.appendLine(styleDim.Render(fmt.Sprintf("wrap the fetched content from .rehamr/fetched/%s.md, and", name)))
	m.appendLine(styleDim.Render(fmt.Sprintf("write internal/skills/%s.md — then run go build ./... .", name)))

	return m, nil
}

// skillNameFromURL derives a kebab-case skill name from the last path segment.
func skillNameFromURL(u *url.URL) string {
	// Use the last non-empty path segment.
	seg := strings.TrimRight(u.Path, "/")
	if idx := strings.LastIndexByte(seg, '/'); idx >= 0 {
		seg = seg[idx+1:]
	}
	if seg == "" {
		seg = u.Host
	}
	// Strip extension.
	if ext := filepath.Ext(seg); ext != "" {
		seg = strings.TrimSuffix(seg, ext)
	}
	// Lowercase, replace non-alnum with hyphens, collapse runs.
	seg = strings.ToLower(seg)
	var b strings.Builder
	lastHyphen := false
	for _, r := range seg {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			if r == '-' {
				if !lastHyphen && b.Len() > 0 {
					b.WriteByte('-')
				}
				lastHyphen = true
				continue
			}
			b.WriteRune(r)
			lastHyphen = false
		} else if r == ' ' || r == '_' || r == '.' {
			if !lastHyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			lastHyphen = true
		}
	}
	name := strings.Trim(b.String(), "-")
	if len(name) < 2 {
		name = "new-skill"
	}
	return name
}

func (m Model) cmdDoctor(_ []string) (tea.Model, tea.Cmd) {
	m.appendLine("Running diagnostics...")
	result := doctor.Run(m.ProjectDir, *m.cfg, "")
	m.appendLine(result)
	return m, nil
}

func (m Model) cmdMcp(args []string) (tea.Model, tea.Cmd) {
	if m.mcpManager == nil {
		m.appendLine(styleWarn.Render("MCP not available."))
		return m, nil
	}
	if len(args) == 0 {
		m.appendLine(m.mcpManager.FormatStatus())
		m.appendLine("")
		m.appendLine(styleDim.Render("/mcp connect|disconnect|tools|enable|disable <server> [tool]"))
		m.appendLine(styleDim.Render("Servers are configured in .rehamr/mcp.json or via RECOMPHAMR_MCP_* env vars."))
		return m, nil
	}
	switch args[0] {
	case "connect":
		if len(args) < 2 {
			m.appendLine(styleWarn.Render("usage: /mcp connect <name>"))
			return m, nil
		}
		m.appendLine(fmt.Sprintf("mcp: connecting to %s...", args[1]))
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return mcpConnectMsg{name: args[1], err: m.mcpManager.Connect(ctx, args[1])}
		}
	case "disconnect":
		if len(args) < 2 {
			m.appendLine(styleWarn.Render("usage: /mcp disconnect <name>"))
			return m, nil
		}
		m.mcpManager.Disconnect(args[1])
		m.appendLine(fmt.Sprintf("mcp %s: disconnected", args[1]))
	case "tools":
		if len(args) < 2 {
			m.appendLine(styleWarn.Render("usage: /mcp tools <server>"))
			return m, nil
		}
		m.appendLine(m.mcpManager.FormatTools(args[1]))
	case "enable":
		if len(args) < 2 {
			m.appendLine(styleWarn.Render("usage: /mcp enable <server> <tool | *>"))
			return m, nil
		}
		if len(args) >= 3 && args[2] == "*" {
			if err := m.mcpManager.SetAllToolsEnabled(args[1], true); err != nil {
				m.appendLine(styleError.Render(err.Error()))
			} else {
				m.appendLine(styleOK.Render(fmt.Sprintf("mcp %s: all tools enabled", args[1])))
			}
		} else if len(args) >= 3 {
			if err := m.mcpManager.SetToolEnabled(args[1], args[2], true); err != nil {
				m.appendLine(styleError.Render(err.Error()))
			} else {
				m.appendLine(styleOK.Render(fmt.Sprintf("mcp %s: enabled %s", args[1], args[2])))
			}
		} else {
			m.appendLine(styleWarn.Render("usage: /mcp enable <server> <tool | *>"))
		}
	case "disable":
		if len(args) < 2 {
			m.appendLine(styleWarn.Render("usage: /mcp disable <server> <tool | *>"))
			return m, nil
		}
		if len(args) >= 3 && args[2] == "*" {
			if err := m.mcpManager.SetAllToolsEnabled(args[1], false); err != nil {
				m.appendLine(styleError.Render(err.Error()))
			} else {
				m.appendLine(styleOK.Render(fmt.Sprintf("mcp %s: all tools disabled", args[1])))
			}
		} else if len(args) >= 3 {
			if err := m.mcpManager.SetToolEnabled(args[1], args[2], false); err != nil {
				m.appendLine(styleError.Render(err.Error()))
			} else {
				m.appendLine(styleOK.Render(fmt.Sprintf("mcp %s: disabled %s", args[1], args[2])))
			}
		} else {
			m.appendLine(styleWarn.Render("usage: /mcp disable <server> <tool | *>"))
		}
	default:
			m.appendLine(styleWarn.Render("usage: /mcp [connect|disconnect|tools|enable|disable] <server> [tool]"))
	}
	return m, nil
}

func (m Model) cmdHelp(_ []string) (tea.Model, tea.Cmd) {
	// Avoid import cycle: list commands statically instead of iterating commands slice.
	m.appendLine("Commands:")
	type helpCmd struct{ name, desc string }
	for _, c := range []helpCmd{
		{"/clear", "reset the conversation"},
		{"/models", "list · <name> set"},
		{"/skills", "list built-in RE skills"},
		{"/skill", "load a skill by name"},
		{"/skill-audit", "classify a skill and suggest a template"},
		{"/skill-new", "fetch URL and classify as new skill"},
		{"/init-re", "create .rehamr/ evidence workspace"},
		{"/status-re", "summarize RE project state"},
		{"/doctor", "run environment diagnostics"},
		{"/mcp", "manage MCP servers"},
		{"/help", "show this help"},
	} {
		m.appendLine(fmt.Sprintf("  %-14s %s", c.name, c.desc))
	}
	m.appendLine("")
	m.appendLine("Keys:")
	m.appendLine("  ctrl+l   clear the screen (keeps conversation)")
	m.appendLine("  ctrl+c   cancel running op · press again to quit")
	m.appendLine("  ctrl+d   quit (on empty input)")
	m.appendLine("  tab      autocomplete slash commands and arguments")
	m.appendLine("  up/down  walk prompt history")
	return m, nil
}



