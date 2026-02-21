# Agent Learnings And Guardrails

## Performance + Contract Rules

- Keep `--json` command paths fast and deterministic.
- Do not run update-check logic for machine-consumed `--json` calls.
- Keep async update checks enabled for interactive mode only, so machine callers (`zsync`) avoid unnecessary background work.
- Preserve output contracts for:
  - `zmosh-picker list --json`
  - `zmosh-picker check --json`

## Integration Rules (zsync dependency)

- `list --json` and `check --json` are consumed by the iOS app.
- Backward-compatible field stability is required.
- Avoid behavior changes that add latency or interactive prompts in JSON mode.

## Validation Commands

- `go test ./...`
- `go build -o zmosh-picker ./cmd/zmosh-picker`
