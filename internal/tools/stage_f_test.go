package tools

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func testRestrict(path string, directory bool) error {
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	return os.Chmod(path, mode)
}

func configuredTestSet(t *testing.T) Set {
	t.Helper()
	root := filepath.Join(t.TempDir(), ".rehamr")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	set := NewSet(root, testRestrict)
	set.now = func() time.Time { return time.Date(2026, 7, 16, 12, 30, 0, 0, time.UTC) }
	return set
}

func TestStageFToolSchemasAndStatus(t *testing.T) {
	for _, item := range []struct {
		name   string
		schema map[string]any
		key    string
	}{
		{RepomixrName, RepomixrSchema(), "repo_url"},
		{RecompReferenceName, RecompReferenceSchema(), "url"},
	} {
		fn := item.schema["function"].(map[string]any)
		if item.schema["type"] != "function" || fn["name"] != item.name || fn["description"] == "" {
			t.Fatalf("schema = %#v", item.schema)
		}
		params := fn["parameters"].(map[string]any)
		if _, ok := params["properties"].(map[string]any)[item.key]; !ok {
			t.Fatalf("missing %s", item.key)
		}
	}
	if got := InlineStatus(chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r\nsecret"}}); got != "▶ repomixr: https://github.com/o/r" {
		t.Fatal(got)
	}
	if got := InlineStatus(chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com/doc"}}); got != "▶ recomp_reference: https://example.com/doc" {
		t.Fatal(got)
	}
}

func TestSetDispatchAndConfigurationFailures(t *testing.T) {
	empty := NewSet("", nil)
	for _, call := range []chmctx.ToolCall{
		{ID: "1", Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}},
		{ID: "2", Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com"}},
	} {
		msg := empty.Execute(context.Background(), call)
		if msg.Role != chmctx.RoleTool || msg.ToolCallID != call.ID || msg.ToolName != call.Name || !strings.Contains(msg.Content, "not configured") {
			t.Fatalf("msg=%#v", msg)
		}
	}
	if got := empty.Execute(context.Background(), chmctx.ToolCall{Name: ReadFileName, Arguments: map[string]any{"path": ""}}).Content; !strings.Contains(got, "empty path") {
		t.Fatal(got)
	}
	if got := empty.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"_parse_error": "cut"}}).Content; !strings.Contains(got, "not valid JSON") {
		t.Fatal(got)
	}
	if got := Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName}).Content; !strings.Contains(got, "not configured") {
		t.Fatal(got)
	}
}

func TestParsePublicGitHubRepo(t *testing.T) {
	owner, repo, canonical, err := parsePublicGitHubRepo("https://github.com/Owner/repo.git")
	if err != nil || owner != "Owner" || repo != "repo" || canonical != "https://github.com/Owner/repo" {
		t.Fatalf("%q %q %q %v", owner, repo, canonical, err)
	}
	invalid := []string{"http://github.com/o/r", "https://user@github.com/o/r", "https://github.com:443/o/r", "https://gitlab.com/o/r", "https://github.com/o", "https://github.com/o/r/x", "https://github.com/o/r?q=1", "https://github.com/o/r#x", "https://github.com/../r", "https://github.com/o/r%2Fx", "%"}
	for _, raw := range invalid {
		if _, _, _, err := parsePublicGitHubRepo(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	for _, value := range []string{"", ".", "..", "bad/name", "bad space"} {
		if safeGitHubComponent(value) {
			t.Errorf("safe %q", value)
		}
	}
}

func TestRepomixSuccessTransformsAndSafeXML(t *testing.T) {
	set := configuredTestSet(t)
	set.runGit = func(ctx context.Context, executable string, args []string, dir string) ([]byte, error) {
		if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) <= 0 {
			t.Fatal("clone deadline missing")
		}
		if executable == "" || dir == "" || args[0] != "clone" {
			t.Fatalf("git %q %#v %q", executable, args, dir)
		}
		clone := args[len(args)-1]
		if err := os.MkdirAll(filepath.Join(clone, "src"), 0o700); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Join(clone, ".git"), 0o700); err != nil {
			return nil, err
		}
		files := map[string][]byte{
			"README.md":                     []byte("# comment\nhello ]]> world\n\n"),
			filepath.Join("src", "main.go"): []byte("package main // trim\n"),
			"bad.exe":                       {0, 1},
			"invalid.txt":                   {0xff},
			filepath.Join(".git", "config"): []byte("secret"),
		}
		for name, data := range files {
			if err := os.WriteFile(filepath.Join(clone, name), data, 0o600); err != nil {
				return nil, err
			}
		}
		return []byte("ok"), nil
	}
	if err := os.WriteFile(filepath.Join(set.privateRoot, "repomix-instruction.md"), []byte("focus ]]> now"), 0o600); err != nil {
		t.Fatal(err)
	}
	call := chmctx.ToolCall{ID: "r", Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/acme/demo.git", "branch": "dev&one", "remove_comments": true, "remove_empty_lines": true, "show_line_numbers": true}}
	msg := set.Execute(context.Background(), call)
	if !strings.Contains(msg.Content, "Packed 2 files") || msg.ToolCallID != "r" {
		t.Fatal(msg.Content)
	}
	path := filepath.Join(set.privateRoot, "repos", "acme-demo", "packed.xml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"branch=\"dev&amp;one\"", "README.md", "src/main.go", "]]]]><![CDATA[>", "<instruction>", "focus"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %s", want, text)
		}
	}
	for _, bad := range []string{".git/config", "bad.exe", "invalid.txt", "# comment", "// trim"} {
		if strings.Contains(text, bad) {
			t.Errorf("unexpected %q", bad)
		}
	}
	if info, _ := os.Stat(path); runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%o", info.Mode().Perm())
	}
}

func TestRepomixValidationCloneAndScanFailures(t *testing.T) {
	set := configuredTestSet(t)
	invalid := []map[string]any{
		{"repo_url": "not-url"},
		{"repo_url": "https://github.com/o/r", "branch": "--bad"},
		{"repo_url": "https://github.com/o/r", "branch": "bad\nbranch"},
	}
	for _, args := range invalid {
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: args}).Content
		if !strings.Contains(got, "invalid") {
			t.Errorf("got %q", got)
		}
	}
	set.runGit = func(context.Context, string, []string, string) ([]byte, error) {
		return []byte(strings.Repeat("x", 5000)), errors.New("fail")
	}
	if got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content; !strings.Contains(got, "clone failed") || len(got) > 4500 {
		t.Fatal(got)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if got := set.Execute(cancelled, chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content; !strings.Contains(got, "cancelled") {
		t.Fatal(got)
	}
	set.runGit = func(_ context.Context, _ string, args []string, _ string) ([]byte, error) {
		clone := args[len(args)-1]
		if err := os.MkdirAll(clone, 0o700); err != nil {
			return nil, err
		}
		return nil, os.WriteFile(filepath.Join(clone, "huge.txt"), bytes.Repeat([]byte{'x'}, maxRepomixFileBytes+1), 0o600)
	}
	if got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content; !strings.Contains(got, "Packed 0 files") {
		t.Fatal(got)
	}
}

func TestRepomixHelpers(t *testing.T) {
	if !skippedRepomixExtension(".PNG") || skippedRepomixExtension(".go") {
		t.Fatal("extension filter")
	}
	if got := stripSimpleComments("http://x\n# x\na // b"); got != "http://x\na " {
		t.Fatal(got)
	}
	args := repomixArgs{removeEmpty: true, removeComments: true, showLines: true, compress: true}
	if got := transformRepomixLines("# x\na   b\n\n", args); len(got) != 1 || got[0] != "1 | a b" {
		t.Fatalf("%#v", got)
	}
	if humanSize(100) != "100 B" || humanSize(1024) != "1.0 KB" || humanSize(1<<20) != "1.0 MB" {
		t.Fatal("sizes")
	}
	if boundedDiagnostic(nil) != "git exited without a diagnostic" || !strings.HasSuffix(boundedDiagnostic(bytes.Repeat([]byte{'x'}, 5000)), "…") {
		t.Fatal("diagnostic")
	}
}

func TestReferenceURLAndNetworkPolicy(t *testing.T) {
	for _, raw := range []string{"https://example.com/a", "http://example.com"} {
		if _, err := validateReferenceURL(raw); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{"file:///x", "https://user@example.com/x", "https://localhost/x", "http://127.0.0.1/x", "http://[::1]/x", "%"} {
		if _, err := validateReferenceURL(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	for _, ip := range []string{"127.0.0.1", "10.0.0.1", "100.64.0.1", "169.254.1.1", "192.0.2.1", "198.18.0.1", "203.0.113.1", "240.0.0.1", "224.0.0.1", "0.0.0.0", "::1", "100::1", "2001:db8::1", "fc00::1", "fe80::1", "::ffff:10.0.0.1"} {
		if publicIP(net.ParseIP(ip)) {
			t.Errorf("public %s", ip)
		}
	}
	if !publicIP(net.ParseIP("8.8.8.8")) {
		t.Fatal("public refused")
	}
	if !publicIP(net.ParseIP("2606:4700:4700::1111")) {
		t.Fatal("public IPv6 refused")
	}
	if publicIP(nil) {
		t.Fatal("nil IP accepted")
	}
	client := newPublicHTTPClient()
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	if _, err := client.Transport.RoundTrip(req); err == nil {
		t.Fatal("private dial accepted")
	}
	if err := client.CheckRedirect(req, make([]*http.Request, 5)); err == nil {
		t.Fatal("redirect limit")
	}
}

func TestReferenceFetchHTMLCacheAndPlain(t *testing.T) {
	set := configuredTestSet(t)
	requests := 0
	set.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		body := "<html><header>hide</header><h1>Title</h1><p>Hello <b>world</b></p><script>bad</script></html>"
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/html"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})}
	call := chmctx.ToolCall{ID: "x", Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com/docs?q=a"}}
	first := set.Execute(context.Background(), call)
	if !strings.Contains(first.Content, "Fetched https://example.com/docs?redacted") || strings.Contains(first.Content, "q=a") || requests != 1 {
		t.Fatalf("%s requests=%d", first.Content, requests)
	}
	files, err := filepath.Glob(filepath.Join(set.privateRoot, "reference", "*.txt"))
	if err != nil || len(files) != 1 {
		t.Fatalf("%v %#v", err, files)
	}
	if err := os.Chtimes(files[0], set.now(), set.now()); err != nil {
		t.Fatal(err)
	}
	second := set.Execute(context.Background(), call)
	if !strings.Contains(second.Content, "Cached") || requests != 1 {
		t.Fatalf("%s requests=%d", second.Content, requests)
	}
	data, _ := os.ReadFile(files[0])
	text := string(data)
	if !strings.Contains(text, "Title") || !strings.Contains(text, "Hello world") || strings.Contains(text, "hide") || strings.Contains(text, "bad") {
		t.Fatal(text)
	}
	old := set.now().Add(-25 * time.Hour)
	if err := os.Chtimes(files[0], old, old); err != nil {
		t.Fatal(err)
	}
	set.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/plain"}}, Body: io.NopCloser(strings.NewReader("plain")), Request: req}, nil
	})}
	if got := set.Execute(context.Background(), call).Content; !strings.Contains(got, "Fetched") {
		t.Fatal(got)
	}
}

func TestReferenceFailuresAndHelpers(t *testing.T) {
	set := configuredTestSet(t)
	if got := set.Execute(context.Background(), chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "http://127.0.0.1"}}).Content; !strings.Contains(got, "invalid URL") {
		t.Fatal(got)
	}
	cases := []struct {
		name string
		trip roundTripFunc
		want string
	}{
		{"network", func(*http.Request) (*http.Response, error) { return nil, errors.New("dial failed") }, "fetch failed"},
		{"status", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 404, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("secret")), Request: req}, nil
		}, "HTTP 404"},
		{"large", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(io.LimitReader(strings.NewReader(strings.Repeat("x", maxReferenceBytes+2)), maxReferenceBytes+2)), Request: req}, nil
		}, "exceeds"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			local := configuredTestSet(t)
			local.httpClient = &http.Client{Transport: tc.trip}
			got := local.Execute(context.Background(), chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com/x"}}).Content
			if !strings.Contains(got, tc.want) {
				t.Fatal(got)
			}
		})
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	set.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) { return nil, req.Context().Err() })}
	if got := set.Execute(cancelled, chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com"}}).Content; !strings.Contains(got, "cancelled") {
		t.Fatal(got)
	}
	if got := sanitizeCacheName("a/b?c"); got != "a_b_c" {
		t.Fatal(got)
	}
	if got := extractHTMLText([]byte("<p>A<br>B</p>")); !strings.Contains(got, "A") || !strings.Contains(got, "B") {
		t.Fatalf("%q", got)
	}
	if _, err := readBoundedBody(errorReader{}, 10); err == nil {
		t.Fatal("read error hidden")
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read") }

type failingCacheTemp struct {
	name                        string
	writeErr, syncErr, closeErr error
}

func (f failingCacheTemp) Name() string { return f.name }
func (f failingCacheTemp) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return len(p), nil
}
func (f failingCacheTemp) Sync() error  { return f.syncErr }
func (f failingCacheTemp) Close() error { return f.closeErr }

func restoreCacheHooks(t *testing.T) {
	t.Helper()
	lstat, mkdir, create, rename, read, remove := cacheLstat, cacheMkdirAll, cacheCreateTemp, cacheRename, cacheReadFile, cacheRemoveAll
	gitLookup := gitLookPath
	repomixRead, repomixWalk := repomixReadFile, walkRepomix
	t.Cleanup(func() {
		cacheLstat, cacheMkdirAll, cacheCreateTemp, cacheRename, cacheReadFile, cacheRemoveAll = lstat, mkdir, create, rename, read, remove
		gitLookPath = gitLookup
		repomixReadFile, walkRepomix = repomixRead, repomixWalk
	})
}

func TestCacheSafetyHelpers(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "d")
	if err := preparePrivateDir(dir, testRestrict); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "x")
	if err := atomicPrivateWrite(file, []byte("one"), testRestrict); err != nil {
		t.Fatal(err)
	}
	if data, err := readOptionalRegular(file, 10); err != nil || string(data) != "one" {
		t.Fatalf("%q %v", data, err)
	}
	if _, err := readOptionalRegular(file, 1); err == nil {
		t.Fatal("oversize accepted")
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(dir, link); err == nil {
		if err := refuseExistingSpecial(link); err == nil {
			t.Fatal("link accepted")
		}
	}
	nondir := filepath.Join(root, "plain")
	if err := os.WriteFile(nondir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := refuseExistingSpecial(nondir); err == nil {
		t.Fatal("file dir accepted")
	}
	directoryTarget := filepath.Join(root, "directory-target")
	if err := os.Mkdir(directoryTarget, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := atomicPrivateWrite(directoryTarget, []byte("x"), testRestrict); err == nil {
		t.Fatal("nonregular target accepted")
	}
}

func TestCacheInjectedFailures(t *testing.T) {
	boom := errors.New("boom")
	t.Run("lstat", func(t *testing.T) {
		restoreCacheHooks(t)
		cacheLstat = func(string) (os.FileInfo, error) { return nil, boom }
		if err := refuseExistingSpecial("x"); !errors.Is(err, boom) {
			t.Fatal(err)
		}
		if err := atomicPrivateWrite("x", nil, testRestrict); !errors.Is(err, boom) {
			t.Fatal(err)
		}
	})
	t.Run("mkdir", func(t *testing.T) {
		restoreCacheHooks(t)
		cacheMkdirAll = func(string, os.FileMode) error { return boom }
		if err := preparePrivateDir(filepath.Join(t.TempDir(), "x"), testRestrict); !errors.Is(err, boom) {
			t.Fatal(err)
		}
	})
	t.Run("restrict-dir", func(t *testing.T) {
		if err := preparePrivateDir(filepath.Join(t.TempDir(), "x"), func(string, bool) error { return boom }); !errors.Is(err, boom) {
			t.Fatal(err)
		}
	})
	t.Run("create", func(t *testing.T) {
		restoreCacheHooks(t)
		cacheCreateTemp = func(string, string) (cacheTemporaryFile, error) { return nil, boom }
		if err := atomicPrivateWrite(filepath.Join(t.TempDir(), "x"), nil, testRestrict); !errors.Is(err, boom) {
			t.Fatal(err)
		}
	})
	for _, kind := range []string{"write", "sync", "close"} {
		t.Run(kind, func(t *testing.T) {
			restoreCacheHooks(t)
			dir := t.TempDir()
			tmp := filepath.Join(dir, "tmp")
			cacheCreateTemp = func(string, string) (cacheTemporaryFile, error) {
				failing := failingCacheTemp{name: tmp}
				switch kind {
				case "write":
					failing.writeErr = boom
				case "sync":
					failing.syncErr = boom
				case "close":
					failing.closeErr = boom
				}
				return failing, nil
			}
			if err := atomicPrivateWrite(filepath.Join(dir, "out"), []byte("x"), testRestrict); !errors.Is(err, boom) {
				t.Fatal(err)
			}
		})
	}
	t.Run("restrict-temp", func(t *testing.T) {
		dir := t.TempDir()
		if err := atomicPrivateWrite(filepath.Join(dir, "x"), nil, func(string, bool) error { return boom }); !errors.Is(err, boom) {
			t.Fatal(err)
		}
	})
	t.Run("rename", func(t *testing.T) {
		restoreCacheHooks(t)
		cacheRename = func(string, string) error { return boom }
		if err := atomicPrivateWrite(filepath.Join(t.TempDir(), "x"), nil, testRestrict); !errors.Is(err, boom) {
			t.Fatal(err)
		}
	})
	t.Run("read", func(t *testing.T) {
		restoreCacheHooks(t)
		path := filepath.Join(t.TempDir(), "x")
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		cacheReadFile = func(string) ([]byte, error) { return nil, boom }
		if _, err := readOptionalRegular(path, 10); !errors.Is(err, boom) {
			t.Fatal(err)
		}
	})
}

func TestPublicHTTPClientInjectedDialPaths(t *testing.T) {
	boom := errors.New("boom")
	client := newPublicHTTPClientWith(func(context.Context, string, string) ([]net.IP, error) { return nil, boom }, func(context.Context, string, string) (net.Conn, error) { return nil, boom })
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if _, err := client.Transport.RoundTrip(req); err == nil {
		t.Fatal("resolver error hidden")
	}
	transport := client.Transport.(*http.Transport)
	if _, err := transport.DialContext(context.Background(), "tcp", "missing-port"); err == nil {
		t.Fatal("bad address accepted")
	}
	client = newPublicHTTPClientWith(func(context.Context, string, string) ([]net.IP, error) { return nil, nil }, func(context.Context, string, string) (net.Conn, error) { return nil, boom })
	if _, err := client.Transport.RoundTrip(req); err == nil || !strings.Contains(err.Error(), "no addresses") {
		t.Fatal(err)
	}
	dialed := ""
	client = newPublicHTTPClientWith(func(context.Context, string, string) ([]net.IP, error) { return []net.IP{net.ParseIP("8.8.8.8")}, nil }, func(_ context.Context, _ string, address string) (net.Conn, error) {
		dialed = address
		return nil, boom
	})
	if _, err := client.Transport.RoundTrip(req); err == nil || dialed != "8.8.8.8:443" {
		t.Fatalf("%q %v", dialed, err)
	}
	redirect, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	if err := client.CheckRedirect(redirect, nil); err == nil {
		t.Fatal("unsafe redirect")
	}
}

func TestRunGitCommand(t *testing.T) {
	git, err := gitLookPath("git")
	if err != nil {
		t.Skip("git unavailable")
	}
	out, err := runGitCommand(context.Background(), git, []string{"--version"}, t.TempDir())
	if err != nil || !strings.Contains(string(out), "git version") {
		t.Fatalf("%q %v", out, err)
	}
}

type stageFInfo struct {
	name string
	mode os.FileMode
	size int64
}

func (f stageFInfo) Name() string      { return f.name }
func (f stageFInfo) Size() int64       { return f.size }
func (f stageFInfo) Mode() os.FileMode { return f.mode }
func (stageFInfo) ModTime() time.Time  { return time.Time{} }
func (f stageFInfo) IsDir() bool       { return f.mode.IsDir() }
func (stageFInfo) Sys() any            { return nil }

type stageFEntry struct {
	info    stageFInfo
	infoErr error
}

func (e stageFEntry) Name() string               { return e.info.name }
func (e stageFEntry) IsDir() bool                { return e.info.IsDir() }
func (e stageFEntry) Type() fs.FileMode          { return e.info.mode.Type() }
func (e stageFEntry) Info() (fs.FileInfo, error) { return e.info, e.infoErr }

func TestRepomixAndReferenceInjectedFailures(t *testing.T) {
	boom := errors.New("boom")
	t.Run("prepare-existing-file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "x")
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := preparePrivateDir(path, testRestrict); err == nil {
			t.Fatal("special accepted")
		}
	})
	t.Run("symlink-info", func(t *testing.T) {
		restoreCacheHooks(t)
		cacheLstat = func(string) (os.FileInfo, error) { return stageFInfo{mode: os.ModeSymlink}, nil }
		if err := refuseExistingSpecial("x"); err == nil {
			t.Fatal("link accepted")
		}
	})
	t.Run("repomix-cache", func(t *testing.T) {
		set := configuredTestSet(t)
		set.restrict = func(string, bool) error { return boom }
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "cache") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-special", func(t *testing.T) {
		set := configuredTestSet(t)
		root := filepath.Join(set.privateRoot, "repos")
		if err := os.Mkdir(root, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "o-r"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "expected directory") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-remove", func(t *testing.T) {
		restoreCacheHooks(t)
		set := configuredTestSet(t)
		cacheRemoveAll = func(string) error { return boom }
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "clean cache") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-restrict-repo", func(t *testing.T) {
		set := configuredTestSet(t)
		calls := 0
		set.restrict = func(path string, dir bool) error {
			calls++
			if calls == 2 {
				return boom
			}
			return testRestrict(path, dir)
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "cache") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-git-missing", func(t *testing.T) {
		restoreCacheHooks(t)
		set := configuredTestSet(t)
		gitLookPath = func(string) (string, error) { return "", boom }
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "git unavailable") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-scan", func(t *testing.T) {
		set := configuredTestSet(t)
		set.runGit = func(context.Context, string, []string, string) ([]byte, error) { return nil, nil }
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "scan") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-timeout", func(t *testing.T) {
		set := configuredTestSet(t)
		set.repomixTimeout = time.Millisecond
		set.runGit = func(ctx context.Context, _ string, _ []string, _ string) ([]byte, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "timeout after 1ms") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-pack", func(t *testing.T) {
		restoreCacheHooks(t)
		set := configuredTestSet(t)
		set.runGit = func(_ context.Context, _ string, args []string, _ string) ([]byte, error) {
			clone := args[len(args)-1]
			if err := os.MkdirAll(clone, 0o700); err != nil {
				return nil, err
			}
			return nil, os.WriteFile(filepath.Join(clone, "x"), []byte("x"), 0o600)
		}
		reads := 0
		repomixReadFile = func(path string) ([]byte, error) {
			reads++
			if reads > 1 {
				return nil, boom
			}
			return os.ReadFile(path)
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "pack") {
			t.Fatal(got)
		}
	})
	t.Run("repomix-write", func(t *testing.T) {
		set := configuredTestSet(t)
		set.runGit = func(_ context.Context, _ string, args []string, _ string) ([]byte, error) {
			return nil, os.MkdirAll(args[len(args)-1], 0o700)
		}
		calls := 0
		set.restrict = func(path string, dir bool) error {
			calls++
			if calls == 3 {
				return boom
			}
			return testRestrict(path, dir)
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RepomixrName, Arguments: map[string]any{"repo_url": "https://github.com/o/r"}}).Content
		if !strings.Contains(got, "write") {
			t.Fatal(got)
		}
	})
	t.Run("reference-cache", func(t *testing.T) {
		set := configuredTestSet(t)
		set.restrict = func(string, bool) error { return boom }
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com"}}).Content
		if !strings.Contains(got, "cache") {
			t.Fatal(got)
		}
	})
	t.Run("reference-long-special", func(t *testing.T) {
		set := configuredTestSet(t)
		long := "https://example.com/" + strings.Repeat("a", 150)
		u, _ := validateReferenceURL(long)
		name := strings.Trim(sanitizeCacheName(u.Hostname()+"-"+strings.Trim(u.EscapedPath(), "/")), "-_.")
		name = name[:96]
		digest := sha256.Sum256([]byte(u.String()))
		name += fmt.Sprintf("-%x", digest[:6])
		root := filepath.Join(set.privateRoot, "reference")
		if err := os.Mkdir(root, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(root, name+".txt"), 0o700); err != nil {
			t.Fatal(err)
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": long}}).Content
		if !strings.Contains(got, "unsafe cache") {
			t.Fatal(got)
		}
	})
	t.Run("reference-secure-existing", func(t *testing.T) {
		set := configuredTestSet(t)
		set.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("x")), Request: req}, nil
		})}
		call := chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com/x"}}
		if got := set.Execute(context.Background(), call).Content; !strings.Contains(got, "Fetched") {
			t.Fatal(got)
		}
		set.now = func() time.Time { return time.Now().Add(time.Hour) }
		calls := 0
		set.restrict = func(path string, directory bool) error {
			calls++
			if calls == 2 {
				return boom
			}
			return testRestrict(path, directory)
		}
		got := set.Execute(context.Background(), call).Content
		if !strings.Contains(got, "cache") {
			t.Fatal(got)
		}
	})
	t.Run("reference-lstat", func(t *testing.T) {
		restoreCacheHooks(t)
		set := configuredTestSet(t)
		orig := cacheLstat
		cacheLstat = func(path string) (os.FileInfo, error) {
			if strings.HasSuffix(path, ".txt") {
				return nil, boom
			}
			return orig(path)
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com/x"}}).Content
		if !strings.Contains(got, "cache") {
			t.Fatal(got)
		}
	})
	t.Run("reference-write", func(t *testing.T) {
		set := configuredTestSet(t)
		set.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("x")), Request: req}, nil
		})}
		calls := 0
		set.restrict = func(path string, dir bool) error {
			calls++
			if calls == 2 {
				return boom
			}
			return testRestrict(path, dir)
		}
		got := set.Execute(context.Background(), chmctx.ToolCall{Name: RecompReferenceName, Arguments: map[string]any{"url": "https://example.com/x"}}).Content
		if !strings.Contains(got, "write") {
			t.Fatal(got)
		}
	})
}

func TestCollectRepomixLimitsAndSpecialEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := collectRepomixFilesWithLimits(dir, 0, 10); err == nil {
		t.Fatal("file limit")
	}
	if _, _, err := collectRepomixFilesWithLimits(dir, 10, 0); err == nil {
		t.Fatal("byte limit")
	}
	restoreCacheHooks(t)
	walkRepomix = func(root string, fn fs.WalkDirFunc) error {
		if err := fn(root, nil, boomError{}); err == nil {
			return errors.New("walk error hidden")
		}
		_ = fn(filepath.Join(root, "dir"), stageFEntry{info: stageFInfo{name: "dir", mode: os.ModeDir | os.ModeSymlink}}, nil)
		_ = fn(filepath.Join(root, "link"), stageFEntry{info: stageFInfo{name: "link", mode: os.ModeSymlink}}, nil)
		return nil
	}
	files, _, err := collectRepomixFiles(dir)
	if err != nil || len(files) != 0 {
		t.Fatalf("%#v %v", files, err)
	}
}

func TestDirectoryTreeEscapesAndReferenceRedaction(t *testing.T) {
	root := t.TempDir()
	if got := directoryTree(root, []string{filepath.Join(root, "a&b")}); !strings.Contains(got, "a&amp;b") {
		t.Fatal(got)
	}
	u, _ := url.Parse("https://example.com/x?q=secret#fragment")
	got := redactedReferenceURL(u)
	if got != "https://example.com/x?redacted#fragment" || strings.Contains(got, "secret") {
		t.Fatal(got)
	}
}

type boomError struct{}

func (boomError) Error() string { return "boom" }
