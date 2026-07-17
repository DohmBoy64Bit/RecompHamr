package tools

// ActivateSkillName is the model-facing standards-based skill activation tool.
const ActivateSkillName = "activate_skill"

// ReadSkillResourceName is the model-facing on-demand skill resource tool.
const ReadSkillResourceName = "read_skill_resource"

// ActivateSkillSchema describes the constrained activation tool. Names are
// copied into an enum so the model cannot request undiscovered skills.
func ActivateSkillSchema(names []string) map[string]any {
	values := append([]string(nil), names...)
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        ActivateSkillName,
			"description": "Load one discovered Agent Skill's full instructions when the current task matches its catalog description.",
			"parameters": map[string]any{
				"type":                 "object",
				"required":             []string{"name"},
				"additionalProperties": false,
				"properties": map[string]any{
					"name": map[string]any{"type": "string", "enum": values, "description": "Exact discovered skill name."},
				},
			},
		},
	}
}

// ReadSkillResourceSchema describes on-demand reads from activated skills.
// Runtime validation constrains both values to an activated, enumerated file.
func ReadSkillResourceSchema(names []string) map[string]any {
	values := append([]string(nil), names...)
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        ReadSkillResourceName,
			"description": "Read one listed supporting file from an already activated Agent Skill when its instructions require that resource.",
			"parameters": map[string]any{
				"type":                 "object",
				"required":             []string{"name", "path"},
				"additionalProperties": false,
				"properties": map[string]any{
					"name": map[string]any{"type": "string", "enum": values, "description": "Exact activated skill name."},
					"path": map[string]any{"type": "string", "description": "Exact relative resource path listed by activate_skill."},
				},
			},
		},
	}
}
