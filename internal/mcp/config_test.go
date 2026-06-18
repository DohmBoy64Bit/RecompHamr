package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStripJSONCommentsRemovesLineComments(t *testing.T) {
	input := []byte("{\"a\": 1 // line\n, \"b\": 2}")
	out := stripJSONComments(input)
	got := string(out)
	want := "{\"a\": 1 , \"b\": 2}"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestStripJSONCommentsRemovesBlockComments(t *testing.T) {
	input := []byte("{\"a\": 1 /* block */, \"b\": 2}")
	out := stripJSONComments(input)
	got := string(out)
	want := "{\"a\": 1 , \"b\": 2}"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLoadMCPConfigWithComments(t *testing.T) {
	dir := t.TempDir()
	json := []byte(`{
  // override ghidra
  "ghidra": {
    "command": "python",
    "args": ["test.py"]
  },
  /* http transport */
  "pcrecomp": {
    "url": "http://localhost:9000",
    "tools": "*"
  }
}`)
	if err := os.WriteFile(filepath.Join(dir, "mcp.json"), json, 0o644); err != nil {
		t.Fatal(err)
	}
	cfgs, err := LoadMCPConfig(dir)
	if err != nil {
		t.Fatalf("LoadMCPConfig with comments: %v", err)
	}
	if len(cfgs) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(cfgs))
	}
	g, ok := cfgs["ghidra"]
	if !ok {
		t.Fatal("missing ghidra entry")
	}
	if g.Command == nil || *g.Command != "python" {
		t.Fatalf("ghidra command wrong: %v", g.Command)
	}
	p, ok := cfgs["pcrecomp"]
	if !ok {
		t.Fatal("missing pcrecomp entry")
	}
	if p.URL == nil || *p.URL != "http://localhost:9000" {
		t.Fatalf("pcrecomp url wrong: %v", p.URL)
	}
	if p.Tools == nil || !p.Tools.AllowAll {
		t.Fatal("pcrecomp tools should be allow-all")
	}
}

func TestLoadMCPConfigMissingFileIsNotError(t *testing.T) {
	dir := t.TempDir()
	cfgs, err := LoadMCPConfig(dir)
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if len(cfgs) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(cfgs))
	}
}
