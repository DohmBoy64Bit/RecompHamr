package session

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/DohmBoy64Bit/RecompHamr/internal/config"
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/provider"
)

var (
	bootstrapRuntimeConfig = config.Bootstrap
	newRuntimeClient       = llm.New
	reachableRuntime       = provider.Reachable
)

// ProfileSnapshot is the non-secret presentation view of one configured
// endpoint. URL is the stored profile URL used in model lists; ActiveURL on
// Snapshot is the effective process endpoint after any runtime override.
type ProfileSnapshot struct {
	Name        string
	Model       string
	URL         string
	ContextSize int
	Active      bool
	Keyed       bool
}

// Snapshot is an immutable, non-secret view of session configuration.
type Snapshot struct {
	Active      string
	ActiveURL   string
	ActiveModel string
	ContextSize int
	ActiveKeyed bool
	Profiles    []ProfileSnapshot
}

// Profile returns the named immutable profile facts, if present.
func (s Snapshot) Profile(name string) (ProfileSnapshot, bool) {
	for _, profile := range s.Profiles {
		if profile.Name == name {
			return profile, true
		}
	}
	return ProfileSnapshot{}, false
}

// Runtime owns live configuration, concrete model-client replacement, backend
// work capture, and prompt-history persistence for one application session.
type Runtime struct {
	mu      sync.RWMutex
	cfg     *config.Config
	client  *llm.Client
	history History
}

// NewRuntime constructs a session around an already bootstrapped configuration
// and creates the concrete client for its effective active profile.
func NewRuntime(cfg *config.Config) *Runtime {
	profile := cfg.ActiveProfile()
	return &Runtime{
		cfg:     cfg,
		client:  newRuntimeClient(cfg.ActiveURL(), profile.LLM, profile.ResolvedKey()),
		history: NewHistory(cfg.Dir),
	}
}

// Snapshot returns the current presentation-safe configuration facts by value.
func (r *Runtime) Snapshot() Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return snapshotOf(r.cfg)
}

func snapshotOf(cfg *config.Config) Snapshot {
	active := cfg.ActiveProfile()
	snapshot := Snapshot{
		Active:      cfg.Active,
		ActiveURL:   cfg.ActiveURL(),
		ActiveModel: active.LLM,
		ContextSize: active.ContextSize,
		ActiveKeyed: active.ResolvedKey() != "",
		Profiles:    make([]ProfileSnapshot, 0, len(cfg.Models)),
	}
	for _, name := range cfg.ModelNames() {
		profile := cfg.Models[name]
		snapshot.Profiles = append(snapshot.Profiles, ProfileSnapshot{
			Name:        name,
			Model:       profile.LLM,
			URL:         profile.URL,
			ContextSize: profile.ContextSize,
			Active:      name == cfg.Active,
			Keyed:       profile.ResolvedKey() != "",
		})
	}
	return snapshot
}

func endpointIdentity(cfg *config.Config) (string, string, string) {
	profile := cfg.ActiveProfile()
	return cfg.ActiveURL(), profile.LLM, profile.ResolvedKey()
}

func (r *Runtime) rebuildClient() {
	url, model, key := endpointIdentity(r.cfg)
	r.client = newRuntimeClient(url, model, key)
}

// Reload re-bootstraps config.yaml, preserves the process-only URL override,
// and replaces the concrete client only when the resolved endpoint identity
// changed. It returns the new immutable facts and whether replacement occurred.
func (r *Runtime) Reload() (Snapshot, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	projectRoot := filepath.Dir(r.cfg.Dir)
	fresh, _, err := bootstrapRuntimeConfig(projectRoot)
	if err != nil {
		return Snapshot{}, false, err
	}
	fresh.URLOverride = r.cfg.URLOverride
	oldURL, oldModel, oldKey := endpointIdentity(r.cfg)
	newURL, newModel, newKey := endpointIdentity(fresh)
	r.cfg = fresh
	replaced := oldURL != newURL || oldModel != newModel || oldKey != newKey
	if replaced {
		r.rebuildClient()
	}
	return snapshotOf(r.cfg), replaced, nil
}

// Activate persists a named profile and installs a fresh concrete client for
// it. Save failures retain the previous active profile and client.
func (r *Runtime) Activate(name string) (Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.cfg.SetActive(name); err != nil {
		return snapshotOf(r.cfg), err
	}
	r.rebuildClient()
	return snapshotOf(r.cfg), nil
}

// Chat implements the provider-neutral agent client boundary using the
// concrete client current when the request starts.
func (r *Runtime) Chat(ctx context.Context, messages []chmctx.Message, tools []llm.Tool) <-chan llm.Event {
	r.mu.RLock()
	client := r.client
	r.mu.RUnlock()
	return client.Chat(ctx, messages, tools)
}

// ProbeWork is an opaque authenticated probe captured for one profile/client.
type ProbeWork struct {
	profile string
	client  *llm.Client
}

// Probe captures the current concrete client for a profile-tagged asynchronous
// activation check. The returned work remains stable after later switches.
func (r *Runtime) Probe(profile string) ProbeWork {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return ProbeWork{profile: profile, client: r.client}
}

// ProbeResult contains non-secret authenticated probe facts.
type ProbeResult struct {
	Profile       string
	ContextWindow int
	Err           error
}

// Run executes the captured probe under the caller's bounded context.
func (w ProbeWork) Run(ctx context.Context) ProbeResult {
	result, err := w.client.Probe(ctx)
	return ProbeResult{Profile: w.profile, ContextWindow: result.ContextWindow, Err: err}
}

// ReachabilityWork is an opaque endpoint captured for a connectivity check.
type ReachabilityWork struct {
	url string
}

// Reachability captures the effective active URL for asynchronous checking.
func (r *Runtime) Reachability() ReachabilityWork {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return ReachabilityWork{url: r.cfg.ActiveURL()}
}

// ReachabilityResult contains the checked URL and its transport error.
type ReachabilityResult struct {
	URL string
	Err error
}

// Run checks the captured endpoint using the retained provider contract.
func (w ReachabilityWork) Run(ctx context.Context) ReachabilityResult {
	return ReachabilityResult{URL: w.url, Err: reachableRuntime(ctx, w.url)}
}

// LoadHistory returns persisted prompt recall oldest-first.
func (r *Runtime) LoadHistory() []string { return r.history.Load() }

// AppendHistory persists one prompt using the retained bounded history policy.
func (r *Runtime) AppendHistory(value string) error { return r.history.Append(value) }

// ClearHistory removes persisted prompt recall.
func (r *Runtime) ClearHistory() error { return r.history.Clear() }
