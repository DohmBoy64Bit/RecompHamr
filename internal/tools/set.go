package tools

import (
	"context"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
)

// Set is an immutable, application-configured collection of local tools. It
// keeps cache authority and injectable process/network boundaries below the
// agent and presentation layers.
type Set struct {
	privateRoot    string
	restrict       func(string, bool) error
	runGit         func(context.Context, string, []string, string) ([]byte, error)
	httpClient     *http.Client
	now            func() time.Time
	repomixTimeout time.Duration
	activateSkill  func(string) (string, error)
	readSkill      func(string, string) ([]byte, error)
}

// WithSkillResourceReader returns a copy that serves bounded resources from
// the application-owned active-skill set.
func (s Set) WithSkillResourceReader(read func(string, string) ([]byte, error)) Set {
	s.readSkill = read
	return s
}

// WithSkillActivator returns a copy that dispatches activate_skill through the
// application-owned skills runtime. A nil callback leaves the capability
// unavailable and unknown-tool behavior unchanged.
func (s Set) WithSkillActivator(activate func(string) (string, error)) Set {
	s.activateSkill = activate
	return s
}

// NewSet configures the production tool collection beneath an already-secured
// .rehamr directory. restrict applies platform-native current-user protection
// to every cache directory and file created by the set.
func NewSet(privateRoot string, restrict func(string, bool) error) Set {
	if privateRoot != "" {
		privateRoot = filepath.Clean(privateRoot)
	}
	return Set{
		privateRoot:    privateRoot,
		restrict:       restrict,
		runGit:         runGitCommand,
		httpClient:     newPublicHTTPClient(),
		now:            time.Now,
		repomixTimeout: 5 * time.Minute,
	}
}

// Execute dispatches a tool call through this configured set and returns the
// bounded model-facing tool message.
func (s Set) Execute(parent context.Context, call chmctx.ToolCall) chmctx.Message {
	raw := s.runRaw(parent, call)
	return chmctx.Message{Role: chmctx.RoleTool, Content: chmctx.Truncate(raw), ToolCallID: call.ID, ToolName: call.Name}
}

func (s Set) runRaw(parent context.Context, call chmctx.ToolCall) string {
	if msg, ok := call.Arguments["_parse_error"].(string); ok {
		return parseErrorResult(msg)
	}
	switch call.Name {
	case ActivateSkillName:
		name, _ := call.Arguments["name"].(string)
		if s.activateSkill == nil {
			return "(activate_skill: unavailable)"
		}
		result, err := s.activateSkill(name)
		if err != nil {
			return "(activate_skill: " + boundedDiagnostic([]byte(err.Error())) + ")"
		}
		return result
	case ReadSkillResourceName:
		name, _ := call.Arguments["name"].(string)
		path, _ := call.Arguments["path"].(string)
		if s.readSkill == nil {
			return "(read_skill_resource: unavailable)"
		}
		result, err := s.readSkill(name, path)
		if err != nil {
			return "(read_skill_resource: " + boundedDiagnostic([]byte(err.Error())) + ")"
		}
		return string(result)
	case RepomixrName:
		return s.repomix(parent, repomixArgsFrom(call.Arguments))
	case RecompReferenceName:
		refURL, _ := call.Arguments["url"].(string)
		return s.recompReference(parent, refURL)
	default:
		return runBaselineRaw(parent, call)
	}
}

func runGitCommand(ctx context.Context, executable string, args []string, dir string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	configureProcessTree(cmd)
	cmd.WaitDelay = 100 * time.Millisecond
	buffer := &headTailBuffer{}
	cmd.Stdout, cmd.Stderr = buffer, buffer
	err := cmd.Run()
	return []byte(buffer.String()), err
}
