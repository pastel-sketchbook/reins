# Reins

A framework for consistent, high-quality AI-assisted development.

Reins shapes how AI coding agents collaborate with developers. It codifies
TDD methodology, layered code rules, and independent verification so that
agents produce reliable work -- regardless of context window limits or
memory loss.

## Install

### Binary (recommended)

Prerequisites: `curl`, `tar` (or `unzip` on Windows)

```bash
curl -fsSL https://raw.githubusercontent.com/pastel-sketchbook/reins/main/install.sh | sh
```

Pre-built binaries are available for Linux, macOS, and Windows on both
amd64 and arm64. See [GitHub Releases](https://github.com/pastel-sketchbook/reins/releases)
for manual download.

### From source

Prerequisites: Go 1.26+

```bash
go install github.com/pastel-sketchbook/reins/cmd/reins@latest
```

## Quick Start

```bash
cd your-project

# Bootstrap reins (generic -- language-agnostic TODO placeholders)
reins init

# -- or -- Bootstrap with a language preset (pre-configured for Go)
reins init --lang go

# Commit
git add .reins AGENTS.md rules/ Taskfile.yml
git commit -m "chore: init reins framework"
```

`reins init` creates:

| File | Purpose |
|------|---------|
| `.reins/` | Managed directory -- METHODOLOGY.md, quality principles, rule-guard agent |
| `.editorconfig` | Editor settings -- consistent indentation, charset, trailing whitespace |
| `AGENTS.md` | Bridge file -- points AI agents to `.reins/METHODOLOGY.md` |
| `rules/INDEX.yaml` | Rule trigger mapping -- references `.reins/` principles and your local rules |
| `Taskfile.yml` | Task automation skeleton -- replace the TODO placeholders with your toolchain |
| `AUTOPILOT.md` | Autopilot template -- define a goal for autonomous agent sessions |

It also creates empty `rules/specifics/`, `rules/concerns/`, and
`rules/principles/` directories for your project-specific rules.

With `--lang go`, the preset additionally:

- Replaces `Taskfile.yml` with a fully configured Go toolchain (gofmt, go vet, staticcheck, go test, go build)
- Replaces `AGENTS.md` with a Go-specific bridge file including Tech Stack and ADR conventions
- Activates `**/*.go` trigger in `rules/INDEX.yaml`
- Copies `go.md` language rules into `rules/specifics/`
- Creates `docs/rationale/` for Architecture Decision Records

## Language Presets

The `--lang` flag replaces generic TODO placeholders with a fully configured
scaffold for a specific language. Presets only configure reins artifacts
(Taskfile, AGENTS.md, INDEX.yaml, rules) -- they do not generate application
code.

| Preset | Flag | What it configures |
|--------|------|--------------------|
| Go | `--lang go` | Taskfile with gofmt/vet/staticcheck/test, Go-specific AGENTS.md, `**/*.go` trigger, `rules/specifics/go.md`, `docs/rationale/` |
| Rust | `--lang rust` | Taskfile with cargo fmt/clippy/test/build, Rust TUI AGENTS.md (ratatui + crossterm), `**/*.rs` trigger, `rules/specifics/rust.md`, `docs/rationale/` |
| Zig | `--lang zig` | Taskfile with zig fmt/build/test, Zig AGENTS.md (allocators + std.log), `**/*.zig` trigger, `rules/specifics/zig.md`, `docs/rationale/` |

Adding a new preset requires only a `content/presets/<lang>/` directory and
an entry in the `presetRuleTemplates` map. Presets for TypeScript and
Python are planned.

## Analysis Lenses

The `reins lens` subcommand generates analysis-lens concern templates that
instruct AI agents to apply structured research lenses during codebase
review. The generated file is placed in `rules/concerns/analysis-lenses.md`
and loaded by agents via `INDEX.yaml` triggers.

```bash
# List available lenses and presets
reins lens

# Generate a due-diligence concern (4 auditor lenses)
reins lens --preset dd

# Cherry-pick individual lenses
reins lens --lens synth --lens weak

# Merge a preset with additional lenses
reins lens --preset dd --lens synth

# Custom output path
reins lens --preset all --output docs/analysis-lenses.md
```

The 10 lenses are organized into three categories:

| Category | Lenses |
|----------|--------|
| **Synthesizers** | Expert Synthesizer, Stakeholder Translator, Timeline Constructor |
| **Auditors** | Evidence Mapper, Contradiction Hunter, Assumption Excavator, Weakness Spotter |
| **Architects** | Framework Builder, Implementation Blueprint, Question Generator |

Four presets bundle common combinations:

| Preset | Alias | Lenses |
|--------|-------|--------|
| `quick-synthesis` | `quick` | Expert Synthesizer + Implementation Blueprint |
| `due-diligence` | `dd` | Evidence Mapper, Contradiction Hunter, Assumption Excavator, Weakness Spotter |
| `strategic-planning` | `strat` | Expert Synthesizer, Framework Builder, Implementation Blueprint, Question Generator |
| `full-protocol` | `all` | All 10 lenses |

When 3 or more lenses are active, a cross-lens synthesis section is
appended that instructs the agent to identify themes, gaps, and
recommended actions across all analyses.

## What You Get

### Core methodology (`.reins/METHODOLOGY.md`)

34 MUST/MUST NOT rules covering:

- **TDD** -- Red, Green, Refactor cycle
- **Tidy First** -- structural changes separated from behavioral changes
- **Commits** -- Conventional Commits, one logical unit per commit
- **Quality** -- single responsibility, explicit dependencies, no duplication
- **Verification** -- `task check:all` before every commit

### Layered rule system (`rules/`)

Three tiers of rules, loaded on demand based on which files the agent modifies:

| Tier | Scope | When loaded |
|------|-------|-------------|
| **Principles** | Universal quality rules | Always |
| **Concerns** | Cross-cutting constraints | When file path + optional content pattern matches |
| **Specifics** | Language/framework rules | When file path matches |

Rules use MUST/MUST NOT format with unique IDs (e.g., `P-01`, `S-GO-03`)
for structured reporting. Only relevant rules enter the agent's context.

### Independent verification (`.reins/agents/rule-guard.md`)

A read-only agent that validates code against rules without modifying it.
Separates "do" from "check" -- the agent that writes code is not the one
that verifies it.

### Task automation (`Taskfile.yml`)

A single `task check:all` command runs all quality gates (format, lint,
test). The agent runs this before every commit.

### Autopilot support (`program.md` + backlog)

Two lightweight conventions enable autonomous agent sessions:

- **`AUTOPILOT.md`** -- A scaffold template defining goal, constraints,
  iteration protocol, and success criteria. The agent uses this as its
  prompt for self-directed work loops.
- **"What to do next"** -- A backlog section in `AGENTS.md` that
  persists across sessions. The agent reads it at session start to pick
  up where the previous session left off.

## Project Layout

After `reins init`, your project looks like this:

```
your-project/
├── .reins/                          # managed by reins CLI
│   ├── AGENTS.md                    # core methodology rules
│   ├── agents/rule-guard.md         # verification agent
│   ├── rules/principles/quality.md  # universal quality principles
│   ├── templates/specifics/         # language rule templates
│   └── VERSION                      # installed reins version
│
├── AGENTS.md                        # bridge file (generated, customizable)
├── Taskfile.yml                     # your project's task automation
├── AUTOPILOT.md                     # autopilot goal/constraints (optional)
│
├── docs/rationale/                  # ADRs (created by --lang presets)
│
└── rules/                           # your project-specific rules
    ├── INDEX.yaml                   # trigger mapping
    ├── principles/                  # (optional) additional principles
    ├── concerns/                    # cross-cutting rules
    └── specifics/                   # language/framework rules
```

The `docs/rationale/` directory is only created when using a language
preset (`--lang`). It follows the `000n_<slug>.md` naming convention
for Architecture Decision Records.

## Customization

> **Tip:** If you used `reins init --lang go`, steps 1 and 2 below are
> already done -- the Taskfile is pre-configured and language rules are
> copied into `rules/specifics/`. Skip to step 3.

### 1. Taskfile.yml

Replace the TODO placeholders with your toolchain. The `check:all` task
is the only mandatory contract -- it must run format, lint, and test.

**Go example:**

```yaml
tasks:
  check:all:
    deps: [format, lint, test:unit]
    cmds:
      - echo "All checks passed!"
  format:
    cmds:
      - gofmt -w .
      - gci write -s standard -s default -s localmodule .
  lint:
    cmds:
      - golangci-lint run -v
  test:unit:
    cmds:
      - go test -v ./...
```

### 2. Language rules

List and copy a template:

```bash
reins list
cp .reins/templates/specifics/go.md rules/specifics/go.md
```

Then uncomment the matching trigger in `rules/INDEX.yaml`:

```yaml
specifics:
  - trigger: "**/*.go"
    rules:
      - rules/specifics/go.md
```

### 3. Add concerns

Create rules for cross-cutting patterns:

```yaml
concerns:
  - trigger: "**/*.{yaml,yml,json,toml,env}"
    rules:
      - rules/concerns/no-hardcoded-secrets.md

  - trigger: "**/*.{go,ts,js,py}"
    content_pattern: "[Ss]ingleton|getInstance"
    rules:
      - rules/concerns/no-singletons.md
```

Concerns with `content_pattern` only load when both the path glob and
the content regex match.

### 4. Write custom rules

Rule files are Markdown with MUST/MUST NOT items and unique IDs:

```markdown
# No Hardcoded Secrets

- **C-SEC-01** -- MUST NOT hardcode API keys, passwords, or tokens
  in source files. Use environment variables or a secrets manager.

- **C-SEC-02** -- MUST NOT commit `.env` files containing real
  credentials. Use `.env.example` with placeholder values.
```

## AI Tool Compatibility

| Tool | Entry point | Setup |
|------|------------|-------|
| OpenCode | `AGENTS.md` | Works automatically -- OpenCode reads it at project root |
| Claude Code | `CLAUDE.md` | Create: `cp AGENTS.md CLAUDE.md` (Claude Code auto-reads CLAUDE.md) |
| Cursor | `.cursorrules` | Add: "Read and follow `.reins/METHODOLOGY.md`" |
| Any AI agent | System prompt | Paste `.reins/METHODOLOGY.md` contents into the system prompt |

## Updating Reins

Re-run the same install method, then refresh managed files:

```bash
# Binary
curl -fsSL https://raw.githubusercontent.com/pastel-sketchbook/reins/main/install.sh | sh

# -- or from source --
go install github.com/pastel-sketchbook/reins/cmd/reins@latest
```

```bash
reins update
```

`reins update` overwrites managed files in `.reins/` but never touches
project-owned files (`AGENTS.md`, `Taskfile.yml`, `rules/INDEX.yaml`).

To check for scaffold changes after an update:

```bash
diff AGENTS.md .reins/scaffold/AGENTS.md
diff rules/INDEX.yaml .reins/scaffold/rules/INDEX.yaml
```

## CLI Reference

| Command | Description |
|---------|-------------|
| `reins init [--lang <name>]` | Bootstrap reins in the current project. With `--lang`, apply a language preset (available: `go`, `rust`, `zig`) |
| `reins update` | Refresh managed files to the latest version |
| `reins list` | List available language/framework templates |
| `reins lens [--preset <name>] [--lens <alias>...] [--output <path>]` | Generate analysis-lens concern templates. No flags prints available lenses |
| `reins version` | Print installed reins version |

## License

This project is licensed under the [MIT License](LICENSE).

## Design

For the full rationale, design principles, and component details, see
[FRAMEWORK.md](FRAMEWORK.md).
