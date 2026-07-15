// Package config owns the .rehamr/ directory: config.yaml plus the
// embedded default system prompt. The prompt lives only in the binary,
// never on disk, so it's untamperable and every release ships it consistent.
package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type temporaryFile interface {
	Name() string
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

var (
	lstatPath      = os.Lstat
	mkdirAllPath   = os.MkdirAll
	readConfigFile = os.ReadFile
	marshalYAML    = yaml.Marshal
	createTempFile = func(dir, pattern string) (temporaryFile, error) { return os.CreateTemp(dir, pattern) }
	renameConfig   = os.Rename
	restrictPath   = RestrictPrivatePath
)

//go:embed PROMPT_SYS.md
var DefaultSystemPrompt string

const DirName = ".rehamr"

// defaultContextSize is the seeded LM Studio profile's packing budget and the
// floor Bootstrap coerces a bogus/missing context_size to. It matches the
// accepted Gemma runtime configuration; users who load the model with a
// different context window should update their persisted profile to match.
const defaultContextSize = 16177

// managedProfiles are seeded on first run with the same local profile shape as
// the pinned upstream baseline, minus the hosted-service profile. After first
// run config.yaml is the user's: deletions and renames stick, and Bootstrap
// never re-adds anything.
var managedProfiles = map[string]Profile{
	"local": {
		LLM:         "google/gemma-4-12b-qat",
		URL:         "http://localhost:1234",
		Key:         "",
		ContextSize: defaultContextSize,
	},
}

// Profile is one named model endpoint in config.yaml; `/models` switches
// between them. ContextSize is persisted for every baseline profile.
type Profile struct {
	LLM         string `yaml:"llm"`
	URL         string `yaml:"url"`
	Key         string `yaml:"key"`
	ContextSize int    `yaml:"context_size,omitempty"`
}

// Config is the on-disk schema at .rehamr/config.yaml. Strict decoding:
// unknown top-level keys fail Bootstrap so typos and stale schemas surface
// immediately rather than being silently ignored.
type Config struct {
	Active string              `yaml:"active"`
	Models map[string]*Profile `yaml:"models"`
	// Logging writes a fresh log.txt each start and appends every exchange.
	// Debug instrumentation; removable with this field, debuglog.go, and the
	// dbgWrite call sites.
	Logging bool `yaml:"logging,omitempty"`
	// runtime-only (not serialized)
	Dir string `yaml:"-"`
	// URLOverride, if set, wins over ActiveProfile().URL everywhere we dial
	// out. Kept off the Profile map so the runtime RECOMPHAMR_URL override never
	// round-trips into Save().
	URLOverride string `yaml:"-"`
}

func Default() *Config {
	models := make(map[string]*Profile, len(managedProfiles))
	for name, p := range managedProfiles {
		cp := p
		models[name] = &cp
	}
	return &Config{
		Active: "local",
		Models: models,
	}
}

// Bootstrap returns the config for the current project, creating .rehamr/
// and config.yaml on first use. config.yaml is never overwritten; the prompt
// is embedded, never written to disk.
//
// The directory check uses Lstat (not Stat) and refuses a pre-existing
// .rehamr that isn't a real directory: a symlink there would let a co-tenant
// redirect config.yaml to an attacker path.
func Bootstrap(projectRoot string) (*Config, bool, error) {
	dir := filepath.Join(projectRoot, DirName)
	created := false
	info, err := lstatPath(dir)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, false, fmt.Errorf("%s: refuses to follow symlink, remove or replace with a real directory", dir)
		}
		if !info.IsDir() {
			return nil, false, fmt.Errorf("%s: exists but is not a directory", dir)
		}
	case errors.Is(err, os.ErrNotExist):
		// 0o700: only the project owner should read the config directory.
		if err := mkdirAllPath(dir, 0o700); err != nil {
			return nil, false, err
		}
		created = true
	default:
		return nil, false, err
	}
	if err := restrictPath(dir, true); err != nil {
		return nil, false, fmt.Errorf("secure %s: %w", dir, err)
	}

	cfgPath := filepath.Join(dir, "config.yaml")
	// Same symlink defence as the directory check: a symlinked config.yaml
	// could redirect the read (which config we honour) or the write (clobbering
	// an arbitrary user-writable file with the seed). Refuse with a clear error.
	if li, err := lstatPath(cfgPath); err == nil && li.Mode()&os.ModeSymlink != 0 {
		return nil, false, fmt.Errorf("%s: refuses to follow symlink, remove or replace with a real file", cfgPath)
	}
	if _, err := lstatPath(cfgPath); err == nil {
		if err := restrictPath(cfgPath, false); err != nil {
			return nil, false, fmt.Errorf("secure %s: %w", cfgPath, err)
		}
	}
	var cfg *Config
	if b, err := readConfigFile(cfgPath); err == nil {
		cfg = &Config{} // do NOT merge Default here; strict means strict
		dec := yaml.NewDecoder(bytes.NewReader(b))
		dec.KnownFields(true)
		if err := dec.Decode(cfg); err != nil {
			return nil, false, fmt.Errorf("config.yaml: %w", err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		cfg = Default()
		if err := writeYAML(cfgPath, cfg); err != nil {
			return nil, false, err
		}
	} else {
		return nil, false, err
	}
	cfg.Dir = dir

	// YAML `models: { name: ~ }` decodes to a nil *Profile that would panic on
	// the ContextSize deref below. Reject up front for a readable error.
	for name, p := range cfg.Models {
		if p == nil {
			return nil, false, fmt.Errorf("config.yaml: profile %q is empty; remove it or fill in the required fields", name)
		}
	}
	// Coerce missing/zero/negative context_size to the safe default. An endpoint
	// may report a larger live value at runtime, but config remains the fallback.
	for _, p := range cfg.Models {
		if p.ContextSize <= 0 {
			p.ContextSize = defaultContextSize
		}
	}
	// Coerce a dangling Active to the first profile in sorted order
	// (deterministic). With no profiles at all, fail loud, since runtime would
	// otherwise nil-deref on the first dial-out.
	if _, ok := cfg.Models[cfg.Active]; !ok {
		names := cfg.ModelNames()
		if len(names) == 0 {
			return nil, false, errors.New("config.yaml: no profiles configured; add one under `models:` or delete .rehamr/config.yaml to reseed defaults")
		}
		cfg.Active = names[0]
	}

	return cfg, created, nil
}

// Save rewrites config.yaml.
func (c *Config) Save() error {
	if c.Dir == "" {
		return errors.New("config: Dir not set")
	}
	return writeYAML(filepath.Join(c.Dir, "config.yaml"), c)
}

func writeYAML(path string, v any) error {
	b, err := marshalYAML(v)
	if err != nil {
		return err
	}
	// Re-prepended every Save since yaml.Marshal drops free-form comments. Keep
	// the durable hint focused on the one value that can silently affect context
	// packing across otherwise OpenAI-compatible servers.
	header := []byte(`# recomphamr configuration
#
# context_size is what recomphamr packs to. Set it to the context window your
# active server actually accepts, not merely the model's theoretical maximum.

`)
	// Write to a sibling temp then rename over config.yaml. Rename is atomic
	// within the directory, so a crash, signal, or full disk mid-write can never
	// leave a truncated config.yaml, which Bootstrap's strict decode would fatal
	// on, bricking the next launch until the file is hand-deleted. Mirrors
	// the same promote-by-rename pattern used for crash-safe configuration writes. os.CreateTemp makes the temp 0o600 and
	// rename installs that fresh inode in place.
	tmp, err := createTempFile(filepath.Dir(path), ".config-*.yaml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op after a successful rename; cleans up early returns
	if _, err := tmp.Write(append(header, b...)); err != nil {
		tmp.Close()
		return err
	}
	// Sync before the rename: rename is metadata-only, so a power loss right
	// after Save could otherwise journal the rename ahead of the data and
	// leave the truncated config.yaml the crash-safety above promises away.
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := renameConfig(tmpPath, path); err != nil {
		return err
	}
	return restrictPath(path, false)
}

// RestrictPrivatePath applies the platform-native owner-only protection used
// for configuration, history, and debug logs. On POSIX it sets 0700 for a
// directory or 0600 for a file. On Windows it installs a protected DACL that
// grants full control only to the current process user.
func RestrictPrivatePath(path string, directory bool) error {
	return restrictPrivatePath(path, directory)
}

// ActiveProfile returns the selected profile. Bootstrap guarantees c.Active
// names a real one, so this is a straight map lookup.
func (c *Config) ActiveProfile() *Profile {
	return c.Models[c.Active]
}

// ResolvedKey returns the profile key, expanding it only when the entire value
// is a ${VAR} environment-variable reference. Literal keys are returned
// unchanged so characters such as '$' are never silently rewritten.
func (p *Profile) ResolvedKey() string {
	key := p.Key
	if name, ok := strings.CutPrefix(key, "${"); ok {
		if name, ok = strings.CutSuffix(name, "}"); ok && isEnvName(name) {
			return os.Getenv(name)
		}
	}
	return key
}

// isEnvName reports whether s is a valid environment-variable name
// ([A-Za-z_][A-Za-z0-9_]*).
func isEnvName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (i > 0 && r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

// ActiveURL is the endpoint every dial-out uses: the runtime override if set,
// else the active profile's URL. Use this over ActiveProfile().URL so
// RECOMPHAMR_URL doesn't leak back into Save.
func (c *Config) ActiveURL() string {
	if c.URLOverride != "" {
		return c.URLOverride
	}
	return c.ActiveProfile().URL
}

// ModelNames returns the profile names sorted, so the popover cycles
// deterministically regardless of map iteration order.
func (c *Config) ModelNames() []string {
	return slices.Sorted(maps.Keys(c.Models))
}

// SetActive switches the active profile and persists. Fails on an unknown name,
// no silent coercion. On Save failure it reverts in-memory Active so the live
// model and config.yaml stay in lockstep; otherwise the switch would stick this
// session but vanish on the next Bootstrap.
func (c *Config) SetActive(name string) error {
	if _, ok := c.Models[name]; !ok {
		return fmt.Errorf("unknown model: %s", name)
	}
	prev := c.Active
	c.Active = name
	if err := c.Save(); err != nil {
		c.Active = prev
		return err
	}
	return nil
}
