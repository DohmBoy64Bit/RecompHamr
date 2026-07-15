# Repository Lineage

## Canonical code baseline

The code baseline is the official `codehamr/codehamr` repository at:

```text
85409d167b97bec64ee330d51872d358d3ce2d57
```

The locally supplied official CodeHamr archive identified itself as:

```text
40c7cca098b8d53d97d1b4fe4b343a68f063be1e
```

The reconstruction process began from that fresh upstream archive. The five files changed between the supplied snapshot and the pinned revision were then replaced with their exact pinned upstream versions before any stripping occurred:

- `README.md`
- `internal/llm/llm.go`
- `internal/llm/llm_test.go`
- `internal/tui/model.go`
- `internal/tui/tui_test.go`

A SHA-256 manifest of the complete reconstructed pre-strip tree is stored at `docs/dev/reference/upstream-85409d1-tree.sha256`.

## Strip phase

Only after the pinned upstream tree was reconstructed was the barebones strip applied. The strip is recorded in `docs/dev/reference/strip-manifest.md`.

## RecompHamr-Legacy

The supplied Legacy repository is not the base tree for this template. It is reserved for later feature-by-feature behavioral evidence after the baseline and separation gates close.

## RecompHamr2 documentation reference

The supplied RecompHamr2 documentation archive influenced the idea of separating user docs, governance, durable memory, verification, architecture, workflows, roadmap, and reference material. Its numbering, wording, phase history, and implementation architecture were not copied one-for-one.

## License

The upstream MIT license remains in `LICENSE` and must be preserved with required attribution.
