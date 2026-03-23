# Project Instructions

Read and follow all instructions in `.reins/METHODOLOGY.md`. That file is
your core methodology — TDD, Tidy First, commit rules, and quality standards.

## Rule System

Rules are loaded on demand based on which files you modify.

- **Principles** (always loaded): `.reins/rules/principles/`
- **Project-specific rules**: `rules/specifics/` and `rules/concerns/` (local)
- **Trigger mapping**: `rules/INDEX.yaml` (local — references both `.reins/`
  and local rule files)

Consult `rules/INDEX.yaml` at the start of every task (see METHODOLOGY.md L-01).

## Verification

- Run `task check:all` before every commit.
- The rule-guard agent is defined in `.reins/agents/rule-guard.md`.
  Invoke it for independent verification — do not self-review.

## Audit

Run `task audit` periodically to check for vulnerable dependencies.
Configure the audit task in `Taskfile.yml` with your language's tool
(e.g., `govulncheck`, `cargo audit`, `npm audit`).
