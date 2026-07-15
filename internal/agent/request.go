// Package agent owns RecompHamr's model-facing turn lifecycle and tool-loop
// policy independently of terminal presentation.
package agent

import (
	chmctx "github.com/DohmBoy64Bit/RecompHamr/internal/ctx"
	"github.com/DohmBoy64Bit/RecompHamr/internal/llm"
	"github.com/DohmBoy64Bit/RecompHamr/internal/tools"
)

// BuildMessages packs history to the supplied context window and prepends the
// system prompt. The returned slice is independent of history's backing array.
func BuildMessages(system string, history []chmctx.Message, contextSize int) []chmctx.Message {
	packed := chmctx.Pack(history, chmctx.Budget(contextSize))
	messages := make([]chmctx.Message, 0, len(packed.Messages)+1)
	messages = append(messages, chmctx.Message{Role: chmctx.RoleSystem, Content: system})
	return append(messages, packed.Messages...)
}

// Tools returns the four local tool definitions exposed on every model round
// in their accepted order: powershell, read_file, write_file, and edit_file.
func Tools() []llm.Tool {
	return []llm.Tool{
		schemaToTool(tools.PowerShellSchema()),
		schemaToTool(tools.ReadFileSchema()),
		schemaToTool(tools.WriteFileSchema()),
		schemaToTool(tools.EditFileSchema()),
	}
}

// ToolNames returns the four exposed local tool names in request order.
func ToolNames() []string {
	return []string{tools.PowerShellName, tools.ReadFileName, tools.WriteFileName, tools.EditFileName}
}

func schemaToTool(schema map[string]any) llm.Tool {
	fn := schema["function"].(map[string]any)
	return llm.Tool{
		Type: schema["type"].(string),
		Function: llm.FunctionDef{
			Name:        fn["name"].(string),
			Description: fn["description"].(string),
			Parameters:  fn["parameters"].(map[string]any),
		},
	}
}
