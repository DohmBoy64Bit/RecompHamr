package doctor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/DohmBoy64Bit/recomphamr/internal/config"
)

func Run(projectDir string, cfg config.Config, cfgPath string) string {
	var b strings.Builder
	b.WriteString("RecompHAMR doctor\n")
	b.WriteString("=================\n")
	fmt.Fprintf(&b, "OS/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&b, "Go runtime: %s\n", runtime.Version())
	fmt.Fprintf(&b, "Project: %s\n", projectDir)
	fmt.Fprintf(&b, "Config: %s\n", cfgPath)
	fmt.Fprintf(&b, "Active profile: %s\n", cfg.Active)
	p := cfg.ActiveProfile()
	fmt.Fprintf(&b, "Model: %s\nEndpoint: %s\nContext: %d\n", p.LLM, p.URL, p.ContextSize)

	b.WriteString("\nProfiles\n--------\n")
	for _, n := range cfg.ModelNames() {
		pr := cfg.Models[n]
		mark := "  "
		if n == cfg.Active {
			mark = " *"
		}
		fmt.Fprintf(&b, "%s %-16s %-24s %s\n", mark, n, pr.LLM, pr.URL)
	}

	b.WriteString("\nMemory/GPU hints\n----------------\n")
	b.WriteString(memoryInfo())
	b.WriteString(commandHint("rocm-smi", "--showproductname"))
	b.WriteString(commandHint("rocminfo"))
	b.WriteString(commandHint("vulkaninfo", "--summary"))
	b.WriteString(commandHint("nvidia-smi"))
	b.WriteString(commandHint("lspci"))

	b.WriteString("\nToolchain hints\n---------------\n")
	for _, tool := range []string{"git", "go", "python", "python3", "cmake", "ninja", "make", "ghidraRun", "java"} {
		b.WriteString(which(tool))
	}

	b.WriteString("\nMCP servers\n-----------\n")
	b.WriteString(which("ghidra-mcp"))
	b.WriteString(which("n64-debug-mcp"))
	b.WriteString(which("pcrecomp-mcp"))
	b.WriteString(which("mcp-pine"))
	b.WriteString(which("objdiff-mcp"))
	b.WriteString(which("pcsx2-mcp"))
	b.WriteString(which("mcp-bizhawk"))
	b.WriteString(which("sega2asm-mcp"))
	for _, ev := range []string{
		"RECOMPHAMR_MCP_GHIDRA_COMMAND", "RECOMPHAMR_MCP_N64_COMMAND",
		"RECOMPHAMR_MCP_PCRECOMP_COMMAND", "RECOMPHAMR_MCP_PINE_COMMAND",
		"RECOMPHAMR_MCP_OBJDIFF_COMMAND", "RECOMPHAMR_MCP_PCSX2_COMMAND",
		"RECOMPHAMR_MCP_BIZHAWK_COMMAND",
		"RECOMPHAMR_MCP_SEGA2ASM_COMMAND",
		"RECOMPHAMR_MCP_GHIDRA_TOOLS", "RECOMPHAMR_MCP_PCRECOMP_TOOLS",
		"RECOMPHAMR_MCP_AUTOSTART",
	} {
		if v := os.Getenv(ev); v != "" {
			fmt.Fprintf(&b, "  %s=%s\n", ev, v)
		} else {
			fmt.Fprintf(&b, "  %s (unset)\n", ev)
		}
	}

	b.WriteString("\nEndpoint check\n--------------\n")
	b.WriteString(endpointCheck(p.URL))

	b.WriteString("\nWorkspace (.rehamr/)\n-------------------\n")
	rehamr := filepath.Join(projectDir, ".rehamr")
	workspaceEntries := []struct{ path, label string }{
		{"PROJECT.md", "PROJECT.md"},
		{"REPHAMR_STATE.md", "REPHAMR_STATE.md (persistent memory)"},
		{"EVIDENCE.md", "EVIDENCE.md"},
		{"BLOCKERS.md", "BLOCKERS.md"},
		{"CHANGELOG.md", "CHANGELOG.md"},
		{"repomix-instruction.md", "repomix-instruction.md"},
		{"mcp.json", "mcp.json (MCP server config)"},
	}
	for _, e := range workspaceEntries {
		full := filepath.Join(rehamr, e.path)
		if info, err := os.Stat(full); err == nil {
			fmt.Fprintf(&b, "  %-50s present (%s)\n", e.label, humanSize(info.Size()))
		} else {
			fmt.Fprintf(&b, "  %-50s missing\n", e.label)
		}
	}

	skillsDir := filepath.Join(rehamr, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		count := 0
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				count++
			}
		}
		fmt.Fprintf(&b, "  %-50s %d custom skill(s)\n", "skills/", count)
	} else {
		fmt.Fprintf(&b, "  %-50s none\n", "skills/")
	}

	reposDir := filepath.Join(rehamr, "repos")
	if entries, err := os.ReadDir(reposDir); err == nil {
		fmt.Fprintf(&b, "  %-50s %d cached repo(s)\n", "repos/", len(entries))
	} else {
		fmt.Fprintf(&b, "  %-50s none\n", "repos/")
	}

	b.WriteString("\nPCRECOMP-Next\n-------------\n")
	if pcrecompPath := os.Getenv("RECOMPHAMR_PCRECOMP_PATH"); pcrecompPath != "" {
		if info, err := os.Stat(pcrecompPath); err == nil && info.IsDir() {
			fmt.Fprintf(&b, "  path: %s  (present)\n", pcrecompPath)
		} else {
			fmt.Fprintf(&b, "  path: %s  (not found)\n", pcrecompPath)
		}
	} else {
		b.WriteString("  RECOMPHAMR_PCRECOMP_PATH (unset)\n")
	}
	b.WriteString(which("python"))

	return b.String()
}

func memoryInfo() string {
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/meminfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			var out strings.Builder
			for _, key := range []string{"MemTotal", "MemAvailable", "SwapTotal"} {
				for _, line := range lines {
					if strings.HasPrefix(line, key+":") {
						out.WriteString(line + "\n")
					}
				}
			}
			return out.String()
		}
	}
	if runtime.GOOS == "windows" {
		return commandHint("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_ComputerSystem | Select-Object TotalPhysicalMemory")
	}
	return "memory: not detected on this OS\n"
}

func which(tool string) string {
	path, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Sprintf("%-12s missing\n", tool)
	}
	return fmt.Sprintf("%-12s %s\n", tool, path)
}

func commandHint(name string, args ...string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Sprintf("%s: missing\n", name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if len(text) > 1200 {
		text = text[:1200] + "\n...truncated..."
	}
	if err != nil {
		if text == "" {
			return fmt.Sprintf("%s: installed, command failed: %v\n", name, err)
		}
		return fmt.Sprintf("%s: installed, command failed: %v\n%s\n", name, err, text)
	}
	if text == "" {
		return fmt.Sprintf("%s: installed\n", name)
	}
	return fmt.Sprintf("%s:\n%s\n", name, text)
}

func endpointCheck(base string) string {
	url := strings.TrimRight(base, "/")
	if strings.HasSuffix(url, "/v1") {
		url += "/models"
	} else if !strings.HasSuffix(url, "/models") {
		url += "/v1/models"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "endpoint: invalid URL: " + err.Error() + "\n"
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("endpoint: unreachable at %s (%v)\n", url, err)
	}
	defer resp.Body.Close()
	return fmt.Sprintf("endpoint: %s returned %s\n", url, resp.Status)
}

func humanSize(n int64) string {
	const kb = 1024
	const mb = 1024 * kb
	switch {
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

