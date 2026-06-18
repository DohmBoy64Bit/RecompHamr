// Command recomphamr is the lightweight, fast coding agent for the terminal.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DohmBoy64Bit/recomphamr/internal/config"
	"github.com/DohmBoy64Bit/recomphamr/internal/llm"
	"github.com/DohmBoy64Bit/recomphamr/internal/mcp"
	"github.com/DohmBoy64Bit/recomphamr/internal/skills"
	"github.com/DohmBoy64Bit/recomphamr/internal/tui"
	"github.com/DohmBoy64Bit/recomphamr/internal/update"
)

// updateBudget caps the pre-launch auto-update (checksum fetch + download +
// rename): enough for a ~10MB binary on a slow link, short enough that an
// offline user isn't stalled before the TUI appears.
const updateBudget = 20 * time.Second

// version is injected via -ldflags at build time; "dev" when running `go run`.
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("recomphamr", version)
			return
		case "-h", "--help", "help":
			printHelp()
			return
		}
	}

	// Wipe last session's superseded binary. Apply renames the running exe
	// to <path>.old before promoting the new one; on Windows that old file
	// stays locked until the prior process exits, so deleting it at launch
	// (not at Apply time) always wins.
	if exe, err := os.Executable(); err == nil {
		update.CleanupOld(exe)
	}

	// Pre-launch auto-update; all failures are non-fatal and fall through to
	// the old binary.
	maybeSelfUpdate()

	cwd := mustCwd()
	// created is ignored: any first-run notice printed here is wiped milliseconds
	// later by the unconditional screen+scrollback clear below, before the TUI
	// draws, so there's nothing to announce.
	cfg, _, err := config.Bootstrap(cwd)
	if err != nil {
		log.Fatalf("recomphamr: %v", err)
	}
	applyEnvOverrides(cfg)

	// Allow project-local custom skills in .rehamr/skills/.
	skills.SetCustomDir(filepath.Join(cfg.Dir, "skills"))

	// Opt-in debug log (`logging: true`): truncates .rehamr/log.txt and
	// records every chat exchange. See tui.OpenDebugLog / dbgWrite.
	if cfg.Logging {
		tui.OpenDebugLog(cfg.Dir)
		defer tui.CloseDebugLog()
	}

	p := cfg.ActiveProfile()
	client := llm.New(cfg.ActiveURL(), p.LLM, p.Key)

	mcpmgr := mcp.NewManager()
	for _, s := range mcp.BuiltinServers() {
		mcpmgr.Register(s)
	}
	if userServers, err := mcp.LoadMCPConfig(cfg.Dir); err != nil {
		log.Printf("recomphamr: mcp.json: %v", err)
	} else {
		for _, s := range userServers {
			mcpmgr.ApplyUserConfig(s)
		}
	}
	mcpmgr.ApplyEnvOverrides()

	abs, _ := filepath.Abs(cwd)
	m := tui.New(cfg, client, abs, version, mcpmgr)

	if os.Getenv("RECOMPHAMR_MCP_AUTOSTART") == "1" {
		go mcpmgr.ConnectAll(context.Background())
	}

	// Hard clear before the TUI takes over: \x1b[2J viewport, \x1b[3J
	// scrollback, \x1b[H cursor home, a clean canvas free of prior shell
	// history. Inline-mode safe: the session's own scrollback still
	// accumulates via tea.Println.
	os.Stdout.WriteString("\x1b[2J\x1b[3J\x1b[H")

	// Inline mode (no AltScreen, no mouse capture): only the prompt + status
	// bar render live at the bottom; everything else goes to native
	// scrollback via tea.Println, leaving scrolling/selection/copy to the
	// terminal.
	//
	// WithReportFocus types raw focus-in/out sequences (\x1b[I / \x1b[O) as
	// tea.FocusMsg / tea.BlurMsg so Update can swallow them; otherwise
	// xterm.js hosts (VS Code) leak those bytes as runes into the textarea
	// on every window switch, inflating prompt height with invisible chars.
	if _, err := tea.NewProgram(m, tea.WithReportFocus()).Run(); err != nil {
		log.Fatalf("recomphamr: %v", err)
	}
}

func printHelp() {
	fmt.Println(strings.TrimSpace(`
recomphamr, a lightweight coding agent for RE · decomp · recomp · evidence-backed reconstruction.

Usage:
  recomphamr             start interactive TUI
  recomphamr --version   print version

Slash commands (inside TUI):`))
	tui.PrintHelp(os.Stdout)
	fmt.Println(strings.TrimSpace(`
Keys (inside TUI):
  ctrl+l   clear the screen (keeps conversation)
  ctrl+c   cancel running op · press again to quit
  ctrl+d   quit (on empty input)

Config:
  .rehamr/config.yaml   per-project settings (models, profiles, active)
  .rehamr/mcp.json      MCP server configs (commands, URLs, tools)

Env:
  RECOMPHAMR_URL               override the active profile's url at runtime
  RECOMPHAMR_IDLE_TIMEOUT      stream idle timeout, e.g. 90m or 1h (default 1h)
  RECOMPHAMR_NO_UPDATE_CHECK   set to 1 to skip self-update on launch
  RECOMPHAMR_MCP_AUTOSTART     set to 1 to enable MCP auto-connect on startup
  RECOMPHAMR_MCP_<NAME>_COMMAND  override MCP server stdio command (e.g. GHIDRA, N64, PCRECOMP)
  RECOMPHAMR_MCP_<NAME>_URL      override MCP server HTTP endpoint (streamable-http transport)
  RECOMPHAMR_MCP_<NAME>_TOOLS    comma-separated tool list or * for all`))
}

// isLocalBuild reports whether the binary came from a working tree rather
// than an official release. `go run` leaves version "dev"; `make install` on
// a dirty tree adds a "-dirty" suffix. Goreleaser tags read as non-local and
// still self-update.
func isLocalBuild(version string) bool {
	return version == "dev" || strings.HasSuffix(version, "-dirty")
}

// maybeSelfUpdate runs the pre-launch auto-update. No-op for local builds,
// an already-current hash, an unsupported platform (see update.assetName),
// or any network/filesystem refusal. On success it swaps the binary and
// re-execs via reExec (which only returns on failure). Any failure past
// "update available" prints one stderr line and proceeds with the old binary.
func maybeSelfUpdate() {
	// Skip local builds: hashing a `go run` temp binary against the
	// published checksum would otherwise swap in the last release and hide
	// unreleased work behind an "update applied" banner.
	if isLocalBuild(version) {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), updateBudget)
	defer cancel()
	if !update.Check(ctx, exe) {
		return
	}
	fmt.Fprintln(os.Stderr, "◉ applying recomphamr update...")
	if err := update.Apply(ctx, exe); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ update failed: %v\n", err)
		if os.IsPermission(err) {
			fmt.Fprintln(os.Stderr, "  tip: rerun with sudo, or reinstall with PREFIX=$HOME/.local")
		}
		return
	}
	// Re-launch the new binary. reExec is platform-split: unix execve (same
	// PID) vs. Windows spawn-and-wait. RECOMPHAMR_NO_UPDATE_CHECK=1 stops the
	// replacement run from re-checking its own freshly-written hash. On
	// reExec failure we fall through to the old in-memory binary.
	//
	// os.Setenv overwrites in place so os.Environ() carries exactly one entry;
	// append(os.Environ(), …) would leave a pre-existing user-set value first,
	// and Unix execve resolves os.Getenv to the FIRST match, silently defeating
	// the guard if someone exported RECOMPHAMR_NO_UPDATE_CHECK to a non-"1" value.
	os.Setenv("RECOMPHAMR_NO_UPDATE_CHECK", "1")
	if err := reExec(exe, os.Args, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ re-exec failed: %v (continuing with previous version)\n", err)
	}
}

// mustCwd returns the working directory or exits 1, called only where
// there's nothing sensible to recover to.
func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("recomphamr: %v", err)
	}
	return cwd
}

// applyEnvOverrides folds runtime env vars into cfg. RECOMPHAMR_URL overrides
// the active profile's URL (devcontainers / CI), held on a non-serialised
// field so it never round-trips into config.yaml on Save.
func applyEnvOverrides(cfg *config.Config) {
	if envURL := os.Getenv("RECOMPHAMR_URL"); envURL != "" {
		cfg.URLOverride = envURL
	}
}




