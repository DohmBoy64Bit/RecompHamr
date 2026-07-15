.PHONY: verify baselinecheck docscheck coveragecheck archcheck fmtcheck test build

verify:
	pwsh -NoProfile -File ./scripts/verify.ps1

baselinecheck:
	pwsh -NoProfile -File ./scripts/check-baseline.ps1

docscheck:
	pwsh -NoProfile -File ./scripts/check-docs.ps1

coveragecheck:
	pwsh -NoProfile -File ./scripts/check-coverage.ps1

archcheck:
	pwsh -NoProfile -File ./scripts/check-architecture.ps1

fmtcheck:
	pwsh -NoProfile -File ./scripts/check-format.ps1

test:
	go test ./...

build:
	go build ./cmd/recomphamr
