package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeRelease serves a goreleaser-style manifest plus one binary asset.
// Returns (manifestURL, binaryURL) so tests can plug them into the package's
// constants via the script-style `t.Cleanup` swap below.
type fakeRelease struct {
	srv      *httptest.Server
	manifest string
	binary   []byte
	asset    string
}

func newFakeRelease(t *testing.T, asset string, body []byte, declared string) *fakeRelease {
	t.Helper()
	r := &fakeRelease{binary: body, asset: asset}
	r.manifest = fmt.Sprintf("%s  %s\n", declared, asset)
	r.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/" + asset:
			_, _ = w.Write(body)
		case "/codehamr_checksums.txt":
			_, _ = w.Write([]byte(r.manifest))
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(r.srv.Close)
	return r
}

// withReleaseURLs swaps the `checksumsURL` and `releaseBase` package vars
// for the duration of one test. The test relies on this not running in
// parallel, which Go's default sequential ordering already guarantees.
func withReleaseURLs(t *testing.T, base string) {
	t.Helper()
	origCS := checksumsURL
	origBase := releaseBase
	checksumsURL = base + "/codehamr_checksums.txt"
	releaseBase = base + "/"
	t.Cleanup(func() {
		checksumsURL = origCS
		releaseBase = origBase
	})
}

// hashOf streams sha256 over body and returns the hex digest, mirroring
// the format goreleaser writes into the manifest.
func hashOf(body []byte) string {
	h := sha256.New()
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// platformAsset is the asset name Apply expects for the current runtime —
// pulled from the package's own helper so the test follows the same pattern
// production does.
func platformAsset(t *testing.T) string {
	t.Helper()
	asset, ok := assetName(runtime.GOOS, runtime.GOARCH)
	if !ok {
		t.Skipf("Apply test skipped: unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return asset
}

// TestApplyRejectsChecksumMismatch is the regression case: a binary whose
// hash doesn't match the published manifest must NOT replace the local
// executable. Without this guard a corrupted CDN response or an attacker
// who swapped the binary asset (but not the checksums) would silently
// install whatever bytes arrived.
func TestApplyRejectsChecksumMismatch(t *testing.T) {
	asset := platformAsset(t)
	good := []byte("genuine binary v1\n")
	tampered := []byte("malicious binary v1\n") // different bytes → different hash

	r := newFakeRelease(t, asset, tampered, hashOf(good))
	withReleaseURLs(t, r.srv.URL)

	tmpDir := t.TempDir()
	exec := filepath.Join(tmpDir, "codehamr")
	if err := os.WriteFile(exec, []byte("original\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	beforeHash := hashOf([]byte("original\n"))

	err := Apply(context.Background(), exec)
	if err == nil {
		t.Fatal("Apply must reject a binary that doesn't match the published checksum")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("error must explain the mismatch, got: %v", err)
	}
	got, _ := os.ReadFile(exec)
	if hashOf(got) != beforeHash {
		t.Fatalf("local exec was replaced despite checksum mismatch")
	}
}

// TestApplyAcceptsMatchingChecksum: positive case — a download whose hash
// equals the manifest entry promotes the binary into place.
func TestApplyAcceptsMatchingChecksum(t *testing.T) {
	asset := platformAsset(t)
	body := []byte("legit binary content\n")
	r := newFakeRelease(t, asset, body, hashOf(body))
	withReleaseURLs(t, r.srv.URL)

	tmpDir := t.TempDir()
	exec := filepath.Join(tmpDir, "codehamr")
	if err := os.WriteFile(exec, []byte("old\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Apply(context.Background(), exec); err != nil {
		t.Fatalf("Apply on matching checksum should succeed: %v", err)
	}
	got, err := os.ReadFile(exec)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Fatalf("exec not replaced with downloaded body: got %q", got)
	}
	st, _ := os.Stat(exec)
	if st.Mode()&0o100 == 0 {
		t.Fatalf("exec should be executable, got mode %v", st.Mode())
	}
	// Temp file must be cleaned up.
	matches, _ := filepath.Glob(filepath.Join(tmpDir, ".codehamr-update-*"))
	if len(matches) != 0 {
		t.Fatalf("temp file leaked after successful Apply: %+v", matches)
	}
}

// TestApplyRejectsMissingManifestEntry: if the published manifest exists
// but doesn't list our asset (e.g. a bad release), Apply must abort rather
// than skip the verification step and install an unverified binary.
func TestApplyRejectsMissingManifestEntry(t *testing.T) {
	asset := platformAsset(t)
	body := []byte("would-be binary\n")
	// declare a hash for a DIFFERENT asset name so fetchHash returns ""
	other := "codehamr-not-our-asset"
	r := newFakeRelease(t, asset, body, hashOf(body))
	r.manifest = fmt.Sprintf("%s  %s\n", hashOf(body), other) // no entry for `asset`
	withReleaseURLs(t, r.srv.URL)

	tmpDir := t.TempDir()
	exec := filepath.Join(tmpDir, "codehamr")
	if err := os.WriteFile(exec, []byte("o\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := Apply(context.Background(), exec)
	if err == nil {
		t.Fatal("Apply must abort when no manifest entry exists for the asset")
	}
	if !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("error should mention the missing checksum, got: %v", err)
	}
}

// TestApplyCleansTempOnFailure: a failed download (server returns 500)
// must not leave a half-written temp file in the install directory.
func TestApplyCleansTempOnFailure(t *testing.T) {
	asset := platformAsset(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasSuffix(req.URL.Path, "checksums.txt"):
			_, _ = w.Write([]byte(hashOf([]byte{}) + "  " + asset + "\n"))
		default:
			http.Error(w, "boom", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)
	withReleaseURLs(t, srv.URL)

	tmpDir := t.TempDir()
	exec := filepath.Join(tmpDir, "codehamr")
	if err := os.WriteFile(exec, []byte("o\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Apply(context.Background(), exec); err == nil {
		t.Fatal("Apply must error on download failure")
	}
	matches, _ := filepath.Glob(filepath.Join(tmpDir, ".codehamr-update-*"))
	if len(matches) != 0 {
		t.Fatalf("temp file leaked after failed Apply: %+v", matches)
	}
}

// TestCheckRejectsCorruptManifest is a sanity test for fetchHash: a
// manifest that isn't in the expected `<hash>  <name>` form must not
// crash; just yields an empty hash and Check returns false.
func TestFetchHashHandlesCorruptManifest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not a real manifest\nrandom text\n"))
	}))
	t.Cleanup(srv.Close)
	origCS := checksumsURL
	checksumsURL = srv.URL + "/codehamr_checksums.txt"
	t.Cleanup(func() { checksumsURL = origCS })

	got, err := fetchHash(context.Background(), "codehamr-linux-amd64")
	if err != nil {
		t.Fatalf("corrupt manifest should not error, got: %v", err)
	}
	if got != "" {
		t.Fatalf("missing entry should yield empty hash, got %q", got)
	}
}

// TestCheckHonoursEnvDisableFlag: the air-gap escape hatch
// CODEHAMR_NO_UPDATE_CHECK=1 must short-circuit Check before any HTTP work.
// Without it, CI runs and offline launches would each pay the full 2-second
// fetch deadline on every start.
func TestCheckHonoursEnvDisableFlag(t *testing.T) {
	t.Setenv("CODEHAMR_NO_UPDATE_CHECK", "1")
	if Check(context.Background(), "/nonexistent/binary") {
		t.Fatal("Check must return false when the env disable flag is set")
	}
}

// TestCheckUnsupportedPlatform: when assetName returns ok=false, Check
// must return false without hitting the network at all — no manifest fetch
// would help on a platform that has no published asset.
func TestCheckUnsupportedPlatform(t *testing.T) {
	asset, ok := assetName("plan9", "riscv")
	if ok || asset != "" {
		t.Fatalf("plan9/riscv should not have an asset, got %q ok=%v", asset, ok)
	}
}

// TestCheckReportsUpToDate: the running binary's hash matches the
// published manifest entry → Check returns false (no update needed). This
// is the happy steady-state path that keeps the spinner from firing on
// every launch once a release is rolled out.
func TestCheckReportsUpToDate(t *testing.T) {
	asset := platformAsset(t)
	tmpDir := t.TempDir()
	exec := filepath.Join(tmpDir, "codehamr")
	body := []byte("running binary content\n")
	if err := os.WriteFile(exec, body, 0o755); err != nil {
		t.Fatal(err)
	}
	r := newFakeRelease(t, asset, body, hashOf(body))
	withReleaseURLs(t, r.srv.URL)
	t.Setenv("CODEHAMR_NO_UPDATE_CHECK", "")

	if Check(context.Background(), exec) {
		t.Fatal("Check should return false when local hash matches published")
	}
}

// TestCheckReportsStale: published hash differs from local → Check
// returns true (update available). Drives the maybeSelfUpdate trigger
// in main.go.
func TestCheckReportsStale(t *testing.T) {
	asset := platformAsset(t)
	tmpDir := t.TempDir()
	exec := filepath.Join(tmpDir, "codehamr")
	if err := os.WriteFile(exec, []byte("local v1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := newFakeRelease(t, asset, []byte("remote v2\n"), hashOf([]byte("remote v2\n")))
	withReleaseURLs(t, r.srv.URL)
	t.Setenv("CODEHAMR_NO_UPDATE_CHECK", "")

	if !Check(context.Background(), exec) {
		t.Fatal("Check should return true when local hash differs from published")
	}
}

// TestApplyRespectsContextCancel: a cancelled ctx aborts the download and
// the local exec stays untouched.
func TestApplyRespectsContextCancel(t *testing.T) {
	asset := platformAsset(t)
	body := []byte("matters not\n")
	r := newFakeRelease(t, asset, body, hashOf(body))
	withReleaseURLs(t, r.srv.URL)

	tmpDir := t.TempDir()
	exec := filepath.Join(tmpDir, "codehamr")
	original := []byte("origcontents\n")
	if err := os.WriteFile(exec, original, 0o755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	if err := Apply(ctx, exec); err == nil {
		t.Fatal("cancelled ctx must propagate as an Apply error")
	}
	got, _ := os.ReadFile(exec)
	if string(got) != string(original) {
		t.Fatalf("exec was replaced after cancelled Apply, got %q", got)
	}
}

