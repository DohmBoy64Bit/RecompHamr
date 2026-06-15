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
	b.WriteString("\nEndpoint check\n--------------\n")
	b.WriteString(endpointCheck(p.URL))
	if _, err := os.Stat(filepath.Join(projectDir, ".rehamr", "PROJECT.md")); err == nil {
		b.WriteString(".rehamr workspace: present\n")
	} else {
		b.WriteString(".rehamr workspace: not initialized; run /init-re\n")
	}
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

