# TUI Subtree Instructions

These rules add to the root `AGENTS.md` for work under `internal/tui/`.

## Stage A freeze

Until the baseline gate is manually accepted:

- do not redesign, rearrange, modernize, restyle, or recompose the inherited TUI;
- do not upgrade Bubble Tea, Bubbles, Lip Gloss, or Glamour;
- preserve layout composition, spacing, inline terminal behavior, composer mechanics, transcript behavior, resize handling, model picker, cancellation, and clean exit behavior;
- make only evidence-backed defect repairs needed to restore or secure the stripped baseline.

## Ownership

Do not add new filesystem, process, networking, persistence, MCP, or skills lifecycle responsibilities to the TUI.

During Stage C, move orchestration and side effects behind typed application contracts while preserving accepted observable behavior. Presentation should end with rendering, presentation state, and input-to-intent translation.

## Verification

TUI changes require focused tests plus the canonical repository gate. When the active work affects visible behavior, automated render tests are regression evidence only; perform the target-terminal manual acceptance required by the stage gate.
