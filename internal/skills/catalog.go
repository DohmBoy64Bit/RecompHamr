// Package skills discovers, validates, catalogs, and activates Agent Skills.
// It owns filesystem and trust policy below presentation and implements
// progressive disclosure: discovery reads metadata, while activation reads the
// selected instruction body and lists resources without loading them.
package skills

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const (
	maxSkillFileBytes = 512 << 10
	maxResourceBytes  = 2 << 20
	maxSkillsPerRoot  = 2000
	maxResources      = 2000
)

var validName = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var (
	absPath  = filepath.Abs
	lstat    = os.Lstat
	readDir  = os.ReadDir
	readFile = os.ReadFile
	sameFile = os.SameFile
	walkDir  = filepath.WalkDir
	relPath  = filepath.Rel
)

// Scope identifies the authority and precedence class of a discovery root.
type Scope uint8

const (
	// ScopeBundled contains application-authored skills shipped in the binary.
	ScopeBundled Scope = iota + 1
	// ScopeUser contains skills selected by the current user.
	ScopeUser
	// ScopeProject contains repository-provided skills and outranks user scope.
	ScopeProject
)

// Root is one explicit skills directory. Project roots are ignored unless
// Trusted is true. Earlier roots win collisions within the same scope.
type Root struct {
	Path    string
	Scope   Scope
	Trusted bool
}

// ReadResource loads one exact resource previously enumerated during
// activation. It refuses links, replacement races, unlisted paths, and files
// over the resource-size bound.
func (c Catalog) ReadResource(activation Activation, resource string) ([]byte, error) {
	if !slices.Contains(activation.Resources, resource) {
		return nil, fmt.Errorf("skill %q has no resource %q", activation.Name, resource)
	}
	path := filepath.Join(activation.Directory, filepath.FromSlash(resource))
	data, before, err := readRegular(path, maxResourceBytes)
	if err != nil {
		return nil, fmt.Errorf("read skill %q resource %q: %w", activation.Name, resource, err)
	}
	after, err := lstat(path)
	if err != nil || !sameFile(before, after) {
		return nil, fmt.Errorf("read skill %q resource %q: file changed while reading", activation.Name, resource)
	}
	return data, nil
}

// Severity classifies a display-safe discovery diagnostic.
type Severity uint8

const (
	// SeverityWarning reports a recoverable compatibility or shadowing issue.
	SeverityWarning Severity = iota + 1
	// SeverityError reports a skill that could not be cataloged.
	SeverityError
)

// Diagnostic records why a skill was skipped, accepted compatibly, or
// shadowed. Message contains no skill body or secret configuration.
type Diagnostic struct {
	Severity Severity
	Path     string
	Name     string
	Message  string
}

// Entry is immutable tier-one catalog metadata. Location is the absolute
// SKILL.md path used internally for controlled activation.
type Entry struct {
	Name          string
	Description   string
	Location      string
	Scope         Scope
	Compatibility string
}

// Activation is tier-two instruction content plus a tier-three resource list.
// Resources are relative slash-separated paths and are not read eagerly.
type Activation struct {
	Name          string
	Instructions  string
	Directory     string
	Resources     []string
	Compatibility string
}

// Catalog is an immutable discovery result keyed by canonical skill name.
type Catalog struct {
	entries     map[string]Entry
	order       []string
	diagnostics []Diagnostic
}

// Without removes configured skill names from disclosure and activation while
// retaining a display-safe diagnostic for each discovered disabled skill.
func (c Catalog) Without(names []string) Catalog {
	disabled := make(map[string]bool, len(names))
	for _, name := range names {
		disabled[name] = true
	}
	entries := make(map[string]Entry, len(c.entries))
	order := make([]string, 0, len(c.order))
	diagnostics := slices.Clone(c.diagnostics)
	for _, name := range c.order {
		entry := c.entries[name]
		if disabled[name] {
			diagnostics = append(diagnostics, diagnostic(SeverityWarning, entry.Location, name, "skill disabled by configuration"))
			continue
		}
		entries[name] = entry
		order = append(order, name)
	}
	return Catalog{entries: entries, order: order, diagnostics: diagnostics}
}

// Discover scans immediate child directories of explicit roots for files named
// exactly SKILL.md. It applies project-over-user precedence and stable
// first-root-wins precedence within a scope.
func Discover(roots []Root) Catalog {
	candidates := make(map[string]Entry)
	priority := make(map[string]int)
	diagnostics := make([]Diagnostic, 0)
	for rootIndex, root := range roots {
		if root.Scope != ScopeBundled && root.Scope != ScopeUser && root.Scope != ScopeProject {
			diagnostics = append(diagnostics, diagnostic(SeverityError, root.Path, "", "unsupported discovery scope"))
			continue
		}
		if root.Scope == ScopeProject && !root.Trusted {
			diagnostics = append(diagnostics, diagnostic(SeverityWarning, root.Path, "", "untrusted project skills skipped"))
			continue
		}
		abs, err := absPath(root.Path)
		if err != nil {
			diagnostics = append(diagnostics, diagnostic(SeverityError, root.Path, "", "invalid discovery root"))
			continue
		}
		rootInfo, err := lstat(abs)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil || rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
			diagnostics = append(diagnostics, diagnostic(SeverityError, abs, "", "discovery root is not a real directory"))
			continue
		}
		children, err := readDir(abs)
		if err != nil {
			diagnostics = append(diagnostics, diagnostic(SeverityError, abs, "", "discovery root cannot be read"))
			continue
		}
		if len(children) > maxSkillsPerRoot {
			diagnostics = append(diagnostics, diagnostic(SeverityError, abs, "", "discovery root exceeds the skill-directory limit"))
			continue
		}
		for _, child := range children {
			if !child.IsDir() || child.Type()&os.ModeSymlink != 0 {
				continue
			}
			dir := filepath.Join(abs, child.Name())
			location := filepath.Join(dir, "SKILL.md")
			entry, parseDiagnostics, ok := parseMetadata(location, child.Name(), root.Scope)
			diagnostics = append(diagnostics, parseDiagnostics...)
			if !ok {
				continue
			}
			candidatePriority := int(root.Scope)*len(roots) - rootIndex
			if old, exists := candidates[entry.Name]; exists {
				if candidatePriority <= priority[entry.Name] {
					diagnostics = append(diagnostics, diagnostic(SeverityWarning, entry.Location, entry.Name, "skill shadowed by a higher-precedence skill"))
					continue
				}
				diagnostics = append(diagnostics, diagnostic(SeverityWarning, old.Location, old.Name, "skill shadowed by a higher-precedence skill"))
			}
			candidates[entry.Name] = entry
			priority[entry.Name] = candidatePriority
		}
	}
	order := mapsKeys(candidates)
	slices.Sort(order)
	return Catalog{entries: candidates, order: order, diagnostics: diagnostics}
}

// Entries returns sorted copy-safe catalog metadata.
func (c Catalog) Entries() []Entry {
	out := make([]Entry, 0, len(c.order))
	for _, name := range c.order {
		out = append(out, c.entries[name])
	}
	return out
}

// Diagnostics returns a copy of every discovery diagnostic in observation
// order.
func (c Catalog) Diagnostics() []Diagnostic { return slices.Clone(c.diagnostics) }

// Activate loads one cataloged skill body and enumerates regular supporting
// files. It revalidates the file and refuses links or replacement races.
func (c Catalog) Activate(name string) (Activation, error) {
	entry, ok := c.entries[name]
	if !ok {
		return Activation{}, fmt.Errorf("unknown skill %q", name)
	}
	data, info, err := readRegular(entry.Location, maxSkillFileBytes)
	if err != nil {
		return Activation{}, fmt.Errorf("activate skill %q: %w", name, err)
	}
	meta, body, err := decodeSkill(data)
	if err != nil || meta.Name != entry.Name || meta.Description != entry.Description {
		return Activation{}, fmt.Errorf("activate skill %q: metadata changed since discovery", name)
	}
	after, err := lstat(entry.Location)
	if err != nil || !sameFile(info, after) {
		return Activation{}, fmt.Errorf("activate skill %q: file changed while reading", name)
	}
	resources, err := listResources(filepath.Dir(entry.Location))
	if err != nil {
		return Activation{}, fmt.Errorf("activate skill %q: %w", name, err)
	}
	return Activation{Name: name, Instructions: body, Directory: filepath.Dir(entry.Location), Resources: resources, Compatibility: meta.Compatibility}, nil
}

type metadata struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`
}

func parseMetadata(location, parent string, scope Scope) (Entry, []Diagnostic, bool) {
	data, _, err := readRegular(location, maxSkillFileBytes)
	if errors.Is(err, fs.ErrNotExist) {
		return Entry{}, nil, false
	}
	if err != nil {
		return Entry{}, []Diagnostic{diagnostic(SeverityError, location, "", "SKILL.md is unreadable, unsafe, or oversized")}, false
	}
	meta, _, err := decodeSkill(data)
	if err != nil {
		return Entry{}, []Diagnostic{diagnostic(SeverityError, location, "", "SKILL.md frontmatter is invalid")}, false
	}
	if err := validateMetadata(meta, parent); err != nil {
		return Entry{}, []Diagnostic{diagnostic(SeverityError, location, meta.Name, err.Error())}, false
	}
	abs, err := absPath(location)
	if err != nil {
		return Entry{}, []Diagnostic{diagnostic(SeverityError, location, meta.Name, "SKILL.md path cannot be resolved")}, false
	}
	return Entry{Name: meta.Name, Description: meta.Description, Location: abs, Scope: scope, Compatibility: meta.Compatibility}, nil, true
}

func validateMetadata(meta metadata, parent string) error {
	if len(meta.Name) == 0 || len(meta.Name) > 64 || !validName.MatchString(meta.Name) || meta.Name != parent {
		return errors.New("skill name must match its directory and use 1-64 lowercase letters, numbers, or single hyphens")
	}
	if len(meta.Description) == 0 || len(meta.Description) > 1024 {
		return errors.New("skill description must contain 1-1024 characters")
	}
	if len(meta.Compatibility) > 500 {
		return errors.New("skill compatibility must not exceed 500 characters")
	}
	return nil
}

func decodeSkill(data []byte) (metadata, string, error) {
	if !utf8.Valid(data) || !bytes.HasPrefix(data, []byte("---\n")) {
		return metadata{}, "", errors.New("missing UTF-8 YAML frontmatter")
	}
	end := bytes.Index(data[4:], []byte("\n---"))
	if end < 0 {
		return metadata{}, "", errors.New("unterminated YAML frontmatter")
	}
	end += 4
	var meta metadata
	decoder := yaml.NewDecoder(bytes.NewReader(data[4:end]))
	decoder.KnownFields(true)
	if err := decoder.Decode(&meta); err != nil {
		return metadata{}, "", err
	}
	bodyStart := end + len("\n---")
	if bodyStart < len(data) && data[bodyStart] == '\r' {
		bodyStart++
	}
	if bodyStart < len(data) && data[bodyStart] == '\n' {
		bodyStart++
	}
	return meta, strings.TrimSpace(string(data[bodyStart:])), nil
}

func readRegular(path string, limit int64) ([]byte, os.FileInfo, error) {
	before, err := lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() || before.Size() > limit {
		return nil, nil, errors.New("refuses link, non-regular, or oversized file")
	}
	data, err := readFile(path)
	if err != nil {
		return nil, nil, err
	}
	return data, before, nil
}

func listResources(root string) ([]string, error) {
	resources := make([]string, 0)
	for _, directory := range []string{"scripts", "references", "assets"} {
		base := filepath.Join(root, directory)
		info, err := lstat(base)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, errors.New("resource directory is unsafe")
		}
		err = walkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return errors.New("linked skill resource refused")
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return errors.New("non-regular skill resource refused")
			}
			if len(resources) >= maxResources {
				return errors.New("skill resource limit exceeded")
			}
			rel, relErr := relPath(root, path)
			if relErr != nil {
				return relErr
			}
			resources = append(resources, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	slices.Sort(resources)
	return resources, nil
}

func diagnostic(severity Severity, path, name, message string) Diagnostic {
	return Diagnostic{Severity: severity, Path: path, Name: name, Message: message}
}

func mapsKeys(values map[string]Entry) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
