# AGENTS

Guidance for humans and coding agents working in this repository.

## Purpose

Keep `engine` correctness first, keep `gui` thin, and preserve accounting invariants under all workflows.

## Architecture rules (must follow)

- `engine` contains all business/accounting rules.
- `gui` is presentation + input only.
- `engine` must not import or depend on `gui` or TUI libraries.
- Store access must go through `engine.Store`.
- Keep behavior consistent across store backends (`memory`, `sqlite`).

## Accounting invariants (do not break)

- Double-entry per event: debits == credits.
- Global per-currency net must be zero.
- Integer amounts only; no float math in posted entries.
- Liquidity decomposition must remain interpretable:
  - `float = liquidity balance - sum(position balances)`
- Bilateral position mirrors must net to zero by counterparty/currency pair.

## FX behavior expectations

- FX rates are rational integers (`num/den`).
- Default seed: `EUR/ZAR = 15/1` at app boot.
- Cross-provider auto conversion uses current FX rate.
- After successful conversion only: mutate by exactly `+5%` or `-5%`.
- Failed conversion must not mutate FX state.
- Event metadata must preserve applied FX rate.

## Storage expectations

- Default runtime: SQLite (`modernc.org/sqlite`), no CGO dependency.
- DB path from `TTT_DB`, default `ttt.db`.
- `Reset()` semantics: wipe providers/accounts/entries safely.

## UI expectations

- Main table must remain navigable with frozen metadata columns.
- Workflow menu should keep dangerous actions explicit (`CLEAR` confirmation).
- Balances panel reflects state at highlighted row (time-travel style view).
- Keep error messages short and actionable.

## Dev commands

Build:

```bash
go build ./...
```

Test:

```bash
go test ./...
```

Lint + tests (requires golangci-lint):

```bash
make test
```

Run app:

```bash
go run .
```

## Coding conventions

- Prefer small, focused changes.
- Avoid changing public behavior unless required by feature/spec.
- Add tests for new engine behavior and edge cases.
- Keep comments concise and useful; avoid obvious narration.
- Preserve existing naming/style patterns in each package.

## When modifying workflows

- Update engine tests first or in same change.
- Ensure metadata keys (`workflow`, `step`, FX keys) remain populated.
- Re-check GUI rendering assumptions for new metadata.
- Validate both backends if store behavior is touched.

## Suggested change checklist

1. `go build ./...`
2. `go test ./...`
3. If engine/store touched, run targeted tests under `engine/...`.
4. Verify main UI path manually with `go run .` when view/menu behavior changed.
