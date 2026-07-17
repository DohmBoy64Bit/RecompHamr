package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

func TestActivateSkillSchemaAndDispatch(t *testing.T) {
	names := []string{"alpha"}
	schema := ActivateSkillSchema(names)
	names[0] = "mutated"
	fn := schema["function"].(map[string]any)
	params := fn["parameters"].(map[string]any)
	properties := params["properties"].(map[string]any)
	if fn["name"] != ActivateSkillName || properties["name"].(map[string]any)["enum"].([]string)[0] != "alpha" {
		t.Fatalf("schema = %#v", schema)
	}
	call := chmctx.ToolCall{ID: "1", Name: ActivateSkillName, Arguments: map[string]any{"name": "alpha"}}
	set := NewSet("", nil)
	if got := set.Execute(context.Background(), call).Content; got != "(activate_skill: unavailable)" {
		t.Fatalf("unavailable = %q", got)
	}
	set = set.WithSkillActivator(func(name string) (string, error) { return "activated " + name, nil })
	if got := set.Execute(context.Background(), call).Content; got != "activated alpha" {
		t.Fatalf("activated = %q", got)
	}
	set = set.WithSkillActivator(func(string) (string, error) { return "", errors.New(strings.Repeat("x", 5000)) })
	if got := set.Execute(context.Background(), call).Content; !strings.HasPrefix(got, "(activate_skill: ") || !strings.HasSuffix(got, "…)") {
		t.Fatalf("error = %q", got)
	}
}

func TestReadSkillResourceSchemaAndDispatch(t *testing.T) {
	names := []string{"alpha"}
	schema := ReadSkillResourceSchema(names)
	names[0] = "mutated"
	fn := schema["function"].(map[string]any)
	params := fn["parameters"].(map[string]any)
	properties := params["properties"].(map[string]any)
	if fn["name"] != ReadSkillResourceName || properties["name"].(map[string]any)["enum"].([]string)[0] != "alpha" {
		t.Fatalf("schema = %#v", schema)
	}
	call := chmctx.ToolCall{ID: "1", Name: ReadSkillResourceName, Arguments: map[string]any{"name": "alpha", "path": "references/guide.md"}}
	set := NewSet("", nil)
	if got := set.Execute(context.Background(), call).Content; got != "(read_skill_resource: unavailable)" {
		t.Fatalf("unavailable = %q", got)
	}
	set = set.WithSkillResourceReader(func(name, path string) ([]byte, error) { return []byte(name + ":" + path), nil })
	if got := set.Execute(context.Background(), call).Content; got != "alpha:references/guide.md" {
		t.Fatalf("resource = %q", got)
	}
	set = set.WithSkillResourceReader(func(string, string) ([]byte, error) { return nil, errors.New(strings.Repeat("x", 5000)) })
	if got := set.Execute(context.Background(), call).Content; !strings.HasPrefix(got, "(read_skill_resource: ") || !strings.HasSuffix(got, "…)") {
		t.Fatalf("error = %q", got)
	}
}
