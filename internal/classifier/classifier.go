package classifier

import (
	"regexp"
	"strings"
)

// TemplateClass represents one of the three skill template types or "none".
type TemplateClass string

const (
	FullWorkflow TemplateClass = "full_workflow"
	MicroSkill   TemplateClass = "micro_skill"
	ToolBridge   TemplateClass = "tool_bridge"
	NoneClass    TemplateClass = "none"
)

// ClassificationResult holds the scored outcome of a skill classification.
type ClassificationResult struct {
	Skill        string
	Class        TemplateClass
	Confidence   float64
	Scores       map[TemplateClass]int
	Reasoning    []string
	Alternatives []TemplateClass // other classes within 1 point of the winner
}

// minMeaningfulLen is the minimum body length (chars) required for
// classification. Shorter bodies short-circuit to NoneClass.
const minMeaningfulLen = 60

// Classify reads a skill body and returns its template classification.
func Classify(name, body string) ClassificationResult {
	if len(body) < minMeaningfulLen || !hasContentSignals(body) {
		return ClassificationResult{
			Skill:      name,
			Class:      NoneClass,
			Confidence: 1.0,
			Scores: map[TemplateClass]int{
				FullWorkflow: 0,
				MicroSkill:   0,
				ToolBridge:   0,
			},
			Reasoning: []string{"body too short or no content signals — not classifiable"},
		}
	}
	fullScore := baseScore + scoreFull(body)
	microScore := baseScore + scoreMicro(body)
	bridgeScore := baseScore + scoreBridge(body)

	shared := sharedSignals(body)
	fullScore += shared
	microScore += shared
	bridgeScore += shared

	scores := map[TemplateClass]int{
		FullWorkflow: fullScore,
		MicroSkill:   microScore,
		ToolBridge:   bridgeScore,
	}

	// Determine winner and collect reasoning.
	winner, winnerScore := maxScore(scores)
	reasoning := collectReasoning(body, scores)

	// Confidence: normalised margin over the next-best class.
	nextBest := 0
	for c, s := range scores {
		if c != winner && s > nextBest {
			nextBest = s
		}
	}
	rangeTotal := maxScoreTotal(scores) - minScoreTotal(scores)
	var confidence float64
	if rangeTotal == 0 {
		confidence = 0.0
	} else {
		confidence = float64(winnerScore-nextBest) / float64(rangeTotal)
	}
	// Clamp.
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	// Check for no-template-fit: all scores ≤ 0 or winner score ≤ 0.
	// But also check: if at least one score > 0, there's a fit.
	class := winner
	if winnerScore <= 0 {
		class = NoneClass
	}

	// Gather alternatives within 1 point of winner.
	alternatives := []TemplateClass{}
	for c, s := range scores {
		if c != winner && winnerScore-s <= 1 && s > 0 {
			alternatives = append(alternatives, c)
		}
	}

	return ClassificationResult{
		Skill:        name,
		Class:        class,
		Confidence:   confidence,
		Scores:       scores,
		Reasoning:    reasoning,
		Alternatives: alternatives,
	}
}

// baseScore is added to every class so the total is never negative by default.
// A skill with zero strong signals still gets baseScore; the margin matters.
const baseScore = 1

// Feature detectors — each returns 0 or 1.

var (
	rePhase   = regexp.MustCompile(`(?im)^#{1,3}\s+(Phases?|Pipeline|Tracks)\b`)
	reBuild   = regexp.MustCompile(`(?im)\b(go\s+(build|test)|cmake|make\b|ninja|msbuild|cl\s|gcc\s|compile)\b`)
	reHW      = regexp.MustCompile(`(?im)^#{1,3}\s+(Hardware\s|Architecture\s|Hardware\sModel|CPU|Memory\sMap)\b`)
	reProhib  = regexp.MustCompile(`(?im)^#{1,3}\s+(Prohibitions|Guardrails|Do\sNot)\b`)
	reClose   = regexp.MustCompile(`(?im)^#{1,3}\s+(Session\sClose|Close|Shutdown)\b`)
	reMental  = regexp.MustCompile(`(?im)^#{1,3}\s+(Mental\sModel)\b`)
	reEvidence = regexp.MustCompile(`(?im)^#{1,3}\s+(Evidence\s+(Ladder|Protocol|Tiers?|Artifacts?))\b`)
	reFail    = regexp.MustCompile(`(?im)^#{1,3}\s+(Failure\s(Patterns?|Handling)|Crash\s(Pattern|Table))\b`)
	reOutput  = regexp.MustCompile(`(?im)^#{1,3}\s+(Output\s(Artifacts?|Files?))\b`)
	reMCP     = regexp.MustCompile(`[a-z][a-z0-9_-]*\.[a-z_][a-z0-9_]*`)
	reSetup   = regexp.MustCompile(`(?im)^#{1,3}\s+(Setup|Installation|Prerequisites)\b`)
	reConnect = regexp.MustCompile(`(?im)\b(connect|connection\s+(check|verify)|boot\s+check|verify\s+connection)\b`)
	reOps     = regexp.MustCompile(`(?im)\|\s*Operation\s*\|\s*Tool`)
	reRules   = regexp.MustCompile(`(?im)^#{1,3}\s+(Rules)\b`)
	reWorkflow = regexp.MustCompile(`(?im)^#{1,3}\s+(Workflow|Procedure|Methodology)\b`)
	reTrigger = regexp.MustCompile(`(?im)\b(use\s+this\s+skill\s+(when|for|to)|when\s+to\s+use)\b`)
	reStop    = regexp.MustCompile(`(?im)\b(stop\s+(condition|when)|when\s+to\s+stop|terminate\s+when)\b`)
	reNot     = regexp.MustCompile(`(?im)\b(do\s+not\s+use\s+this\s+skill|when\s+not\s+to\s+use)\b`)
)

// hasContentSignals returns true if any content-bearing regex matches the body.
// Bodies with no structural signals are not classifiable.
func hasContentSignals(body string) bool {
	for _, re := range []*regexp.Regexp{
		rePhase, reBuild, reHW, reProhib, reClose,
		reMental, reEvidence, reFail, reOutput,
		reMCP, reSetup, reConnect, reOps,
		reRules, reWorkflow, reTrigger, reStop, reNot,
	} {
		if re.MatchString(body) {
			return true
		}
	}
	return false
}

func scoreFull(body string) int {
	s := 0
	if rePhase.MatchString(body) {
		s += 3
	}
	if reBuild.MatchString(body) {
		s += 2
	}
	if reHW.MatchString(body) {
		s += 2
	}
	if reProhib.MatchString(body) {
		s += 1
	}
	if reClose.MatchString(body) {
		s += 1
	}
	if reTrigger.MatchString(body) {
		s += 1
	}
	if reNot.MatchString(body) {
		s += 2
	}
	if len(body) > 5000 {
		s += 2
	}
	if len(body) < 800 {
		s -= 2
	}
	return s
}

func scoreMicro(body string) int {
	s := 0
	if rePhase.MatchString(body) {
		s -= 2
	}
	if reBuild.MatchString(body) {
		s -= 1
	}
	if reHW.MatchString(body) {
		s -= 2
	}
	if reClose.MatchString(body) {
		s += 1
	}
	if reRules.MatchString(body) {
		s += 2
	}
	if reWorkflow.MatchString(body) {
		s += 2
	}
	if reTrigger.MatchString(body) {
		s += 2
	}
	if reStop.MatchString(body) {
		s += 2
	}
	if singleConcern(body) {
		s += 2
	}
	if len(body) < 800 {
		s += 2
	}
	return s
}

func scoreBridge(body string) int {
	s := 0
	if rePhase.MatchString(body) {
		s -= 2
	}
	if reBuild.MatchString(body) {
		s -= 1
	}
	if reHW.MatchString(body) {
		s -= 2
	}
	if reClose.MatchString(body) {
		s += 1
	}
	if reMCP.MatchString(body) {
		s += 4
	}
	if reSetup.MatchString(body) {
		s += 2
	}
	if reConnect.MatchString(body) {
		s += 2
	}
	if reOps.MatchString(body) {
		s += 2
	}
	if reTrigger.MatchString(body) {
		s += 1
	}
	if singleConcern(body) {
		s += 1
	}
	if len(body) > 5000 {
		s -= 2
	}
	if len(body) < 800 {
		s += 2
	}
	return s
}

// sharedSignals are quality signals common to well-structured skills of any
// class. They are added after per-class scoring so they don't bias the
// classification — they reward quality but don't change the winner.
func sharedSignals(body string) int {
	s := 0
	if reMental.MatchString(body) {
		s += 1
	}
	if reEvidence.MatchString(body) {
		s += 1
	}
	if reFail.MatchString(body) {
		s += 1
	}
	if reOutput.MatchString(body) {
		s += 1
	}
	return s
}

// singleConcern returns true if the body appears to be about one topic/platform.
// Heuristic: does not enumerate multiple unrelated platforms (N64 + Xbox + PS2 etc.).
func singleConcern(body string) bool {
	platforms := []string{"N64", "Xbox", "PlayStation", "GameCube", "Game Boy", "Genesis", "SNES", "PS2", "PS3", "PSX", "Dreamcast", "Saturn", "Wii", "DOS"}
	count := 0
	for _, p := range platforms {
		if strings.Contains(body, p) {
			count++
		}
	}
	return count <= 1
}

func maxScore(s map[TemplateClass]int) (TemplateClass, int) {
	// Deterministic tie-break: FullWorkflow > MicroSkill > ToolBridge.
	order := []TemplateClass{FullWorkflow, MicroSkill, ToolBridge}
	var best TemplateClass
	bestScore := -9999
	for _, c := range order {
		sc, ok := s[c]
		if !ok {
			continue
		}
		if sc > bestScore {
			best = c
			bestScore = sc
		}
	}
	return best, bestScore
}

func maxScoreTotal(s map[TemplateClass]int) int {
	m := -9999
	for _, v := range s {
		if v > m {
			m = v
		}
	}
	return m
}

func minScoreTotal(s map[TemplateClass]int) int {
	m := 9999
	for _, v := range s {
		if v < m {
			m = v
		}
	}
	return m
}

func collectReasoning(body string, scores map[TemplateClass]int) []string {
	r := []string{}

	winner, _ := maxScore(scores)

	// Positive signals from the winner class's features.
	switch winner {
	case FullWorkflow:
		if rePhase.MatchString(body) {
			r = append(r, "phases/pipeline/tracks section found")
		}
		if reBuild.MatchString(body) {
			r = append(r, "build/test commands found")
		}
		if reHW.MatchString(body) {
			r = append(r, "hardware/architecture section found")
		}
		if reProhib.MatchString(body) {
			r = append(r, "prohibitions/guardrails section found")
		}
		if reMental.MatchString(body) {
			r = append(r, "mental model section found")
		}
		if reEvidence.MatchString(body) {
			r = append(r, "evidence ladder/protocol section found")
		}
		if reFail.MatchString(body) {
			r = append(r, "failure patterns/crash table section found")
		}
		if reOutput.MatchString(body) {
			r = append(r, "output artifacts section found")
		}
		if reTrigger.MatchString(body) {
			r = append(r, "trigger statement found")
		}
		if reNot.MatchString(body) {
			r = append(r, "out-of-scope guard (do not use when) found")
		}
		if len(body) > 5000 {
			r = append(r, "length >5000 chars (full workflow range)")
		}
	case MicroSkill:
		if reRules.MatchString(body) {
			r = append(r, "rules section found")
		}
		if reWorkflow.MatchString(body) {
			r = append(r, "workflow/procedure section found")
		}
		if reTrigger.MatchString(body) {
			r = append(r, "trigger statement found")
		}
		if reStop.MatchString(body) {
			r = append(r, "stop conditions found")
		}
		if singleConcern(body) {
			r = append(r, "single-concern (single platform/topic)")
		}
		if len(body) < 800 {
			r = append(r, "length <800 chars (micro-skill range)")
		}
	case ToolBridge:
		if reMCP.MatchString(body) {
			r = append(r, "MCP server tool references detected")
		}
		if reSetup.MatchString(body) {
			r = append(r, "setup/install section found")
		}
		if reConnect.MatchString(body) {
			r = append(r, "connection check procedure found")
		}
		if reOps.MatchString(body) {
			r = append(r, "common operations table found")
		}
		if reTrigger.MatchString(body) {
			r = append(r, "trigger statement found")
		}
		if singleConcern(body) {
			r = append(r, "single-concern (single tool/bridge)")
		}
		if len(body) < 800 {
			r = append(r, "length <800 chars (tool bridge range)")
		}
	}

	// Negative signals (why NOT the other classes).
	for c, s := range scores {
		if c == winner {
			continue
		}
		switch c {
		case FullWorkflow:
			if s <= baseScore {
				if !rePhase.MatchString(body) {
					r = append(r, "no phases/pipeline section → not full workflow")
				}
				if len(body) < 800 {
					r = append(r, "too short for full workflow")
				}
			}
		case MicroSkill:
			if s <= baseScore {
				if !reRules.MatchString(body) {
					r = append(r, "no rules section → not micro-skill")
				}
				if len(body) > 5000 {
					r = append(r, "too long for micro-skill")
				}
			}
		case ToolBridge:
			if s <= baseScore {
				if !reMCP.MatchString(body) {
					r = append(r, "no MCP tool references → not tool bridge")
				}
				if !reSetup.MatchString(body) && !reConnect.MatchString(body) {
					r = append(r, "no setup or connection check → not tool bridge")
				}
			}
		}
	}

	if len(r) == 0 {
		r = append(r, "no strong signals — default classification")
	}
	return r
}

// TemplateName returns the human-readable name for a template class.
func (c TemplateClass) TemplateName() string {
	switch c {
	case FullWorkflow:
		return "Universal Template (Full Workflow)"
	case MicroSkill:
		return "Micro-Skill Method Template"
	case ToolBridge:
		return "Tool Bridge Card Template"
	default:
		return "No Template"
	}
}
