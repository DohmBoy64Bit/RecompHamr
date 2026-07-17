package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestRuntimeProgressiveDisclosureActivationAndReset(t *testing.T) {
	root := t.TempDir()
	dir := writeSkill(t, root, "alpha", "Use this skill for <alpha> work.", "Do alpha & verify.")
	writeSkill(t, root, "beta", "Use this skill for beta work.", "Do beta.")
	catalog := Discover([]Root{{Path: root, Scope: ScopeUser}})
	runtime := NewRuntime(catalog)
	if len(runtime.Entries()) != 2 || len(runtime.Diagnostics()) != 0 {
		t.Fatalf("runtime catalog = %#v %#v", runtime.Entries(), runtime.Diagnostics())
	}
	text := runtime.SystemText()
	if !strings.Contains(text, "&lt;alpha&gt;") || strings.Contains(text, "Do alpha") || !strings.Contains(text, "activate_skill") {
		t.Fatalf("catalog disclosure = %q", text)
	}
	activation, fresh, err := runtime.Activate("alpha")
	if err != nil || !fresh || activation.Directory != dir || activation.Instructions != "Do alpha & verify." {
		t.Fatalf("activation = %#v %v %v", activation, fresh, err)
	}
	active := runtime.ActiveNames()
	active[0] = "mutated"
	if got := runtime.ActiveNames(); len(got) != 1 || got[0] != "alpha" {
		t.Fatalf("active names = %#v", got)
	}
	activation.Resources = append(activation.Resources, "mutation")
	activation, fresh, err = runtime.Activate("alpha")
	if err != nil || fresh || len(activation.Resources) != 0 {
		t.Fatalf("dedupe = %#v %v %v", activation, fresh, err)
	}
	text = runtime.SystemText()
	if !strings.Contains(text, `<skill_content name="alpha">`) || !strings.Contains(text, "Do alpha & verify.") || !strings.Contains(text, "Skill directory: "+dir) {
		t.Fatalf("activated disclosure = %q", text)
	}
	if _, _, err := runtime.Activate("missing"); err == nil {
		t.Fatal("unknown skill activated")
	}
	runtime.Reset()
	if len(runtime.ActiveNames()) != 0 {
		t.Fatal("reset retained active names")
	}
	if text := runtime.SystemText(); strings.Contains(text, "Do alpha") || !strings.Contains(text, "available_skills") {
		t.Fatalf("reset disclosure = %q", text)
	}
}

func TestRuntimeResourcesEmptyCatalogAndConcurrentDeduplication(t *testing.T) {
	empty := NewRuntime(Discover(nil))
	if empty.SystemText() != "" {
		t.Fatalf("empty system = %q", empty.SystemText())
	}

	root := t.TempDir()
	dir := writeSkill(t, root, "alpha", "Use alpha.", "Alpha body.")
	for _, resource := range []string{"assets/icon.txt", "scripts/run.ps1"} {
		path := filepath.Join(dir, filepath.FromSlash(resource))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("resource"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runtime := NewRuntime(Discover([]Root{{Path: root, Scope: ScopeUser}}))
	var wait sync.WaitGroup
	results := make(chan bool, 2)
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, fresh, err := runtime.Activate("alpha")
			if err != nil {
				t.Errorf("activate: %v", err)
			}
			results <- fresh
		}()
	}
	wait.Wait()
	close(results)
	freshCount := 0
	for fresh := range results {
		if fresh {
			freshCount++
		}
	}
	if freshCount != 1 {
		t.Fatalf("fresh activations = %d", freshCount)
	}
	text := runtime.SystemText()
	if !strings.Contains(text, "<skill_resources>") || !strings.Contains(text, "assets/icon.txt") || !strings.Contains(text, "scripts/run.ps1") {
		t.Fatalf("resource disclosure = %q", text)
	}
	data, err := runtime.ReadResource("alpha", "assets/icon.txt")
	if err != nil || string(data) != "resource" {
		t.Fatalf("resource = %q, %v", data, err)
	}
	if _, err := runtime.ReadResource("alpha", "../SKILL.md"); err == nil {
		t.Fatal("unlisted traversal resource was read")
	}
	if _, err := runtime.ReadResource("beta", "assets/icon.txt"); err == nil {
		t.Fatal("inactive skill resource was read")
	}
}

func TestRuntimeCatalogPromptIsBounded(t *testing.T) {
	root := t.TempDir()
	for i := range 20 {
		name := fmt.Sprintf("skill-%02d", i)
		writeSkill(t, root, name, strings.Repeat("description ", 80), "body")
	}
	runtime := NewRuntime(Discover([]Root{{Path: root, Scope: ScopeUser, Trusted: true}}))
	text := runtime.SystemText()
	if len(text) > maxCatalogPromptBytes+256 || !strings.Contains(text, "catalog_truncated") || strings.Contains(text, "skill-19") {
		t.Fatalf("bounded catalog bytes=%d tail=%q", len(text), text[len(text)-128:])
	}
}

func TestRuntimeReplaceCatalogRetainsOnlyRevalidatedActivations(t *testing.T) {
	first := t.TempDir()
	writeSkill(t, first, "alpha", "Use alpha.", "first body")
	runtime := NewRuntime(Discover([]Root{{Path: first, Scope: ScopeUser, Trusted: true}}))
	if _, _, err := runtime.Activate("alpha"); err != nil {
		t.Fatal(err)
	}
	second := t.TempDir()
	writeSkill(t, second, "alpha", "Use alpha.", "second body")
	writeSkill(t, second, "beta", "Use beta.", "beta body")
	runtime.ReplaceCatalog(Discover([]Root{{Path: second, Scope: ScopeUser, Trusted: true}}))
	if got := runtime.ActiveNames(); len(got) != 1 || got[0] != "alpha" || !strings.Contains(runtime.SystemText(), "second body") {
		t.Fatalf("retained activation = %#v %q", got, runtime.SystemText())
	}
	runtime.ReplaceCatalog(Discover(nil))
	if len(runtime.Entries()) != 0 || len(runtime.ActiveNames()) != 0 {
		t.Fatal("removed catalog retained entries or activations")
	}
}
