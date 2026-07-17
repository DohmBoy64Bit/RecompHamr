package skills

import (
	"fmt"
	"html"
	"strings"
	"sync"
)

const maxCatalogPromptBytes = 8 << 10

// Runtime owns one session's immutable catalog and activated instruction set.
// Activation is idempotent and safe when invoked by asynchronous tool work.
type Runtime struct {
	catalog Catalog
	mu      sync.RWMutex
	active  map[string]Activation
	order   []string
}

// NewRuntime constructs a session skill owner from a completed catalog.
func NewRuntime(catalog Catalog) *Runtime {
	return &Runtime{catalog: catalog, active: make(map[string]Activation)}
}

// Entries returns the tier-one catalog metadata.
func (r *Runtime) Entries() []Entry { return r.catalog.Entries() }

// Diagnostics returns display-safe discovery diagnostics.
func (r *Runtime) Diagnostics() []Diagnostic { return r.catalog.Diagnostics() }

// ActiveNames returns activated skill names in activation order.
func (r *Runtime) ActiveNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.order...)
}

// Activate loads and retains one skill. The returned boolean is false when the
// same skill was already active and no content was reloaded.
func (r *Runtime) Activate(name string) (Activation, bool, error) {
	r.mu.RLock()
	activation, exists := r.active[name]
	r.mu.RUnlock()
	if exists {
		return cloneActivation(activation), false, nil
	}
	activation, err := r.catalog.Activate(name)
	if err != nil {
		return Activation{}, false, err
	}
	r.mu.Lock()
	if existing, raced := r.active[name]; raced {
		r.mu.Unlock()
		return cloneActivation(existing), false, nil
	}
	r.active[name] = cloneActivation(activation)
	r.order = append(r.order, name)
	r.mu.Unlock()
	return cloneActivation(activation), true, nil
}

// ReadResource returns one bounded supporting file from an already activated
// skill. Resources from inactive or unknown skills remain unavailable.
func (r *Runtime) ReadResource(name, resource string) ([]byte, error) {
	r.mu.RLock()
	activation, active := r.active[name]
	r.mu.RUnlock()
	if !active {
		return nil, fmt.Errorf("skill %q is not active", name)
	}
	return r.catalog.ReadResource(activation, resource)
}

// ReplaceCatalog atomically installs a refreshed discovery result. Existing
// activations are retained only when the refreshed catalog can revalidate and
// reactivate the same name; disabled, removed, or changed-invalid skills drop.
func (r *Runtime) ReplaceCatalog(catalog Catalog) {
	r.mu.Lock()
	defer r.mu.Unlock()
	active := make(map[string]Activation, len(r.active))
	order := make([]string, 0, len(r.order))
	for _, name := range r.order {
		activation, err := catalog.Activate(name)
		if err != nil {
			continue
		}
		active[name] = cloneActivation(activation)
		order = append(order, name)
	}
	r.catalog = catalog
	r.active = active
	r.order = order
}

// Reset removes session activations without changing the discovered catalog.
func (r *Runtime) Reset() {
	r.mu.Lock()
	r.active = make(map[string]Activation)
	r.order = nil
	r.mu.Unlock()
}

// SystemText returns the compact tier-one catalog followed by activated
// tier-two instructions. Supporting resources are named but never loaded.
func (r *Runtime) SystemText() string {
	entries := r.catalog.Entries()
	r.mu.RLock()
	order := append([]string(nil), r.order...)
	active := make(map[string]Activation, len(r.active))
	for name, activation := range r.active {
		active[name] = cloneActivation(activation)
	}
	r.mu.RUnlock()
	if len(entries) == 0 && len(order) == 0 {
		return ""
	}
	var out strings.Builder
	if len(entries) > 0 {
		out.WriteString("<available_skills>\n")
		included := 0
		for _, entry := range entries {
			line := fmt.Sprintf("  <skill><name>%s</name><description>%s</description></skill>\n", html.EscapeString(entry.Name), html.EscapeString(entry.Description))
			if out.Len()+len(line)+len("</available_skills>\n") > maxCatalogPromptBytes {
				break
			}
			out.WriteString(line)
			included++
		}
		if omitted := len(entries) - included; omitted > 0 {
			fmt.Fprintf(&out, "  <catalog_truncated omitted=\"%d\"/>\n", omitted)
		}
		out.WriteString("</available_skills>\nUse activate_skill when a task matches a description. Do not guess skill names.\n")
	}
	for _, name := range order {
		activation := active[name]
		fmt.Fprintf(&out, "<skill_content name=\"%s\">\n%s\nSkill directory: %s\nUse read_skill_resource with the listed relative path when a resource is needed.\n", html.EscapeString(name), activation.Instructions, html.EscapeString(activation.Directory))
		if len(activation.Resources) > 0 {
			out.WriteString("<skill_resources>\n")
			for _, resource := range activation.Resources {
				fmt.Fprintf(&out, "  <file>%s</file>\n", html.EscapeString(resource))
			}
			out.WriteString("</skill_resources>\n")
		}
		out.WriteString("</skill_content>\n")
	}
	return strings.TrimSpace(out.String())
}

func cloneActivation(value Activation) Activation {
	value.Resources = append([]string(nil), value.Resources...)
	return value
}
