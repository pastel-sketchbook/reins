# Reins

A framework for tailored AI-assisted development processes.

## Purpose

Reins shapes how AI coding agents collaborate with developers on projects.
It codifies methodology, rules, and automation so that AI agents produce
consistent, high-quality work aligned with the team's standards --
regardless of context window limits, memory loss, or the AI's tendency to
infer rather than verify.

Reins is project-agnostic. It provides the scaffolding; each project
supplies its own language conventions, tooling, and domain rules.

## Design Principles

These principles are derived from recurring failure modes observed when
building large systems with AI agents. The evidence is primarily
practitioner experience, not controlled experiments. The principles are
directionally sound and consistent with known AI limitations, but should
be treated as engineering heuristics, not proven laws.

### Code is the source of truth

Documentation is a point-in-time snapshot. Code changes daily; docs drift.
When AI references stale docs it produces code in already-refactored
patterns. The framework treats well-structured, consistently named code as
the primary system description. Rules enforce the quality that makes this
possible. Narrative architecture docs are optional supplements, never the
authority.

### Rules over prose

AI agents follow explicit MUST/MUST NOT constraints more reliably than
nuanced architectural guidance. The framework expresses standards as
concrete, auditable rules rather than descriptive paragraphs.

### Separate do from check

The agent that writes code must not be the same one that validates it.
Self-policing fails because AI conflates inference with verification -- it
can "pass" its own work without actually checking. Independent verification
is a structural requirement, not a nice-to-have.

### Minimize what must survive compression

Long conversations trigger context compaction. Anything in the agent's
context will eventually be summarized and diluted. Core instructions must
be ruthlessly minimal. Detail lives in loadable files that are pulled in
on demand.

### Enforce at infrastructure level

Prompt-level rules are fragile -- they depend on the AI remembering and
choosing to follow them. Infrastructure-level mechanisms (hooks, automation,
separate agents) work regardless of the AI's memory state.

### Observe to control

If you cannot see whether the agent read files before modifying them,
whether it followed the prescribed workflow, or whether it invoked
verification, then you cannot control quality. Agent behavior must be
observable.

### Layer on existing engineering

Static analysis, linting, formatting, and automated tests are the
foundation. They handle what is mechanically verifiable. AI-based rules
handle what requires semantic understanding -- meaning and context that
AST parsing cannot reach. Different layers, different responsibilities.
The AI layer does not replace the traditional layer.

---

## Components

### 1. Core Agent Instructions (`METHODOLOGY.md`)

The minimal, always-loaded instruction set for the AI agent. Contains
methodology rules, the expected agent workflow, and a pointer to the
project's rule system.

**Design constraints:**
- Keep as short as possible. Minimize ruthlessly; every line competes for
  context that survives compaction.
- Express standards as MUST/MUST NOT rules with IDs, not prose guidance.
- Universal methodology (TDD cycle, commit rules, quality standards)
  lives here permanently.
- Project-specific conventions (language, frameworks, libraries) MUST NOT
  live here. Push them to `rules/`.
- Domain knowledge MUST NOT live here. Push it to `rules/` or `docs/`.

**Rationale:** Context compaction destroys long instructions. The smaller
the core, the higher the probability it survives compression intact.
Detail that the agent only needs situationally should be loaded on demand,
not permanently occupying context.

**Status:** Implemented. `.reins/METHODOLOGY.md` in downstream projects.
95 lines, 34 MUST/MUST NOT rules with IDs, zero project-specific content.

---

### 2. Layered Rule System (`rules/`)

A three-tier hierarchy of MUST/MUST NOT rules, dynamically loaded based
on what files the agent is modifying.

**Tiers:**

| Tier | Scope | Example |
|------|-------|---------|
| **Principles** | Always apply to every file | Single responsibility, dependency direction, explicit error handling |
| **Concerns** | Cross-cutting, activated by pattern | File structure conventions, singleton usage, hardcoding prohibition |
| **Specifics** | Domain/technology, activated by path | Framework-specific handler rules, database integration rules, UI component rules |

**Trigger mapping (`rules/INDEX.yaml`):**
- Maps file paths and content patterns to rule sets.
- **Principles** are listed unconditionally — always loaded.
- **Specifics** entries have a `trigger:` glob matched against modified
  file paths.
- **Concerns** entries have a `trigger:` glob and an optional
  `content_pattern:` regex. When `content_pattern:` is present, the
  concern loads only if both the path glob and the content regex match.
  When absent, the path glob alone is sufficient.
- When the agent modifies a file, Principles always apply. Relevant
  Concerns and Specifics are loaded based on matching triggers.
- The agent never needs all rules in context simultaneously.

**Rule file format:**
- Markdown with MUST/MUST NOT items.
- Each item has a unique ID (e.g., `P-01`, `C3-05`, `S-web-02`).
- IDs enable structured reporting and trend tracking.

**Consumption protocol:**
The agent reads `rules/INDEX.yaml` directly — there is no pre-processor.
On every task:
1. The agent opens `rules/INDEX.yaml`.
2. It loads all files listed under `principles:` (unconditionally).
3. For each file it is about to modify, it matches the path against
   `trigger:` globs in `concerns:` and `specifics:`. It loads the
   `rules:` files for every matching entry.
4. For each matching `concerns:` entry that has a `content_pattern:`,
   the agent scans the target file's content. The concern's rules are
   loaded only if the pattern matches.
5. The agent reads each loaded rule file in full before writing code.

The protocol relies on the AI agent being able to read YAML and follow
file references — a capability all current coding agents have. No
external tooling is required.

**Conflict resolution:**
When rules from different tiers contradict each other, the most-specific
tier wins:
- **Specifics** override **Concerns** override **Principles**.
- The overriding rule SHOULD include a comment referencing the Principle
  it relaxes (e.g., "Overrides P-03 for legacy migration context").
- If two rules within the same tier conflict, it is a rule authoring
  bug. Fix the rules — do not silently pick one.

**Rule freshness:** Rules are a form of documentation and subject to the
same drift problem. Mitigations:
- Rules SHOULD reference specific file paths so stale references are
  detectable.
- A `task check:rules` step MAY validate rule references against the
  current codebase.
- Rules SHOULD be updated in the same commit as the code change that
  makes them stale.

**Rationale:** AI cannot hold hundreds of rules in context and apply them
all correctly. Scoping rules to the current task keeps context small and
relevance high. The three-tier model ensures universal principles are
always active while domain details load only when needed.

**Status:** Partially implemented. `INDEX.yaml` with full schema,
consumption protocol, and conflict resolution policy. One Principles
file (`rules/principles/quality.md`, 8 rules). One Specifics file
(`rules/specifics/go.md`, 16 rules). Concerns tier has directory and
schema but no rule files yet.

---

### 3. Task Automation (`Taskfile.yml`)

A task runner configuration defining all verification, build, and
development commands. Serves as the mechanical verification layer.

**Required tasks:**
- `check:all` -- Run all quality gates (format, lint, test) in one command.
- `format` -- Auto-format source files.
- `lint` -- Run static analysis.
- `test:unit` -- Run unit tests.
- `build` -- Compile/build the project.

**Optional tasks:**
- `bench:*` -- Benchmarks.
- `pprof:*` -- Profiling.
- `loadtest:*` -- Load testing.
- `version:*` -- Semantic version management.

**Rationale:** The agent must have a single, unambiguous command to verify
its work before committing. `task check:all` is the mechanical safety net.
This layer catches what doesn't require semantic understanding -- syntax
errors, formatting violations, type errors, failing tests. It runs the
same way every time regardless of the AI's judgment.

**Status:** Exists. `Taskfile.yml` in repository root.

---

### 4. Independent Verification Agent (`agents/rule-guard.md`)

A separate agent that validates code against rules without modifying it.
The framework's answer to the self-policing problem.

**Properties:**
- Read-only. Never modifies source code.
- Must read actual rule files before judging. Inference from memory is
  forbidden.
- Restricted tool access (grep, file reading, reporting only).
- Invoked at two points:
  - **Pre-review:** Validates the modification plan against applicable rules
    before code is written.
  - **Post-review:** Validates the actual implementation against applicable
    rules after code is written.

**Reporting protocol:**
- One rule per report. Each report includes a `ruleId`.
- Verdict: `PASS` or `VIOLATION`.
- Violations include severity (`critical` | `high` | `medium` | `low`)
  and quoted MUST/MUST NOT text as evidence.

**Rationale:** AI hallucinates having verified things. It can claim
compliance without opening the rule file. Separating the verifier from the
implementor makes this significantly less likely -- the verifier's only job
is to read rules and compare them against code. This is not foolproof (the
verifier is also LLM-based), but it adds a structural barrier that
self-policing lacks.

**Status:** Implemented. `agents/rule-guard.md` defines role, constraints,
workflow, and reporting format. Ready for use.

---

## Extension Points (5-7)

Components 5-7 are not framework deliverables. They are extension
points — the framework defines *what* and *why*; each project decides
*whether* and *how* based on its tooling, scale, and maturity. A solo
developer on a small project may never need them. A team running
multi-agent sessions on a large codebase almost certainly will.

### 5. Memory Resilience (tool-specific)

A project MAY implement mechanisms to re-inject instructions when
context compaction occurs.

**Approaches:**
- **Re-injection hook:** Some tools support hooks triggered on context
  compaction. A hook MAY re-inject `METHODOLOGY.md` and active agent
  definitions mechanically.
- **Instruction minimization:** Already addressed by Component 1's
  design constraint (keep METHODOLOGY.md minimal).
- **Externalized detail:** Already addressed by Component 2 (rules
  loaded on demand, not persisted in context).

**Extend when:** Long sessions cause noticeable instruction drift, or
your AI tool provides a hook/event mechanism you can wire into.

---

### 6. Structured Reporting

A project MAY define a structured format for agents to report their
actions, enabling trend analysis and workflow auditing.

**Useful event types, if implemented:**

| Event | When |
|-------|------|
| `task_start` | New user instruction received |
| `plan` | Approach decided before implementation |
| `action` | Before any file read/write/search |
| `verify` | After test/build/lint execution |
| `rule_check` | Rule verification performed |
| `task_complete` | Work finished |

**Extend when:** You need to measure which rules are frequently
violated, which workflow steps are being skipped, or whether quality
trends are improving over time.

---

### 7. Observability

A project MAY implement tooling to visualize agent behavior patterns.
The expected workflow is:

```
Exploration → Planning → Implementation → Verification
```

An agent that skips Exploration is coding from inference. An agent
that skips Verification is committing unchecked work. Both patterns
should be visible — but making them visible requires tooling beyond
what a framework of markdown files can provide.

**Extend when:** Your team needs to audit agent sessions, diagnose
process failures (not just code failures), or track improvement trends
across sprints.

---

## Directory Structure

```
project-root/
├── AGENTS.md                  # Core agent instructions (minimal)
├── FRAMEWORK.md               # This file (framework design)
├── Taskfile.yml               # Task automation (mechanical verification)
├── AUTOPILOT.md               # Autopilot goal/constraints (optional)
├── .editorconfig              # Editor configuration
│
├── rules/                     # Layered rule system
│   ├── INDEX.yaml             # Trigger mapping (path → rule sets)
│   ├── principles/            # Tier 1: always-active rules
│   │   └── *.md
│   ├── concerns/              # Tier 2: cross-cutting rules
│   │   └── *.md
│   └── specifics/             # Tier 3: domain/technology rules
│       └── *.md
│
├── agents/                    # Agent role definitions
│   └── rule-guard.md          # Independent verification agent
│
└── [project source code]      # The actual project being built
```

---

## Implementation Priority

Build the core components in this order. Each layer depends on the one
before it.

| Phase | Component | Prerequisite | Outcome |
|-------|-----------|--------------|---------|
| 1 | Core Agent Instructions (`METHODOLOGY.md`) | None | Agent follows methodology |
| 2 | Task Automation (`Taskfile.yml`) | None | Mechanical verification works |
| 3 | Layered Rule System (`rules/`) | METHODOLOGY.md references rules | Agent loads project-specific constraints |
| 4 | Independent Verification Agent | Rules exist to verify against | Separate do-from-check enforcement |

Components 5-7 are extension points, not framework deliverables. Each
project extends them based on its tooling, scale, and needs. See the
"Extension Points" section above.

---

## Project Configuration

To apply Reins to a new project:

1. **AGENTS.md** -- The bridge file points to `.reins/METHODOLOGY.md`. Add a
   project context section that names the project and points to `rules/`.

2. **Taskfile.yml** -- Adapt tasks to your language and toolchain. The
   `check:all` task is mandatory. Everything else is project-specific.

   **Minimal template (language-agnostic):**
   ```yaml
   version: "3"
   tasks:
     check:all:
       desc: "Run all quality gates"
       deps: [format, lint, test:unit]
       cmds:
         - echo "All checks passed!"
     format:
       cmds:
         - echo "ERROR: configure the 'format' task" && exit 1
     lint:
       cmds:
         - echo "ERROR: configure the 'lint' task" && exit 1
     test:unit:
       cmds:
         - echo "ERROR: configure the 'test:unit' task" && exit 1
     build:
       cmds:
         - echo "ERROR: configure the 'build' task" && exit 1
   ```
   Replace each `exit 1` placeholder with your toolchain's command. The
   structure (`check:all` depending on `format`, `lint`, `test:unit`) is the
    contract METHODOLOGY.md rule V-01 relies on.

3. **rules/** -- Write rules specific to your project's architecture,
   frameworks, and domain. Start with Principles (universal to your
   codebase), then add Concerns and Specifics as patterns emerge.

4. **agents/rule-guard.md** -- Adapt the verification agent to reference
   your rule structure and reporting protocol.

---

## Distribution

Reins is distributed as a **standalone binary**. Install it once, then run
`reins init` in any project to bootstrap the framework. No git submodules,
no package manager dependencies, no Go toolchain required.

### Install

**Binary (recommended)** -- no Go toolchain needed:

```bash
curl -fsSL https://raw.githubusercontent.com/pastel-sketchbook/reins/main/install.sh | sh
```

Options:

```bash
# Install a specific version
curl -fsSL ... | sh -s -- --version v0.2.0

# Install to a custom directory
curl -fsSL ... | sh -s -- --dir /usr/local/bin
```

Or download directly from
[GitHub Releases](https://github.com/pastel-sketchbook/reins/releases).

**From source** (requires Go toolchain):

```bash
go install github.com/pastel-sketchbook/reins/cmd/reins@latest
```

### Supported platforms

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64, arm64 |

### Quick Start (downstream project)

```bash
cd your-project
reins init
# Customize Taskfile.yml, rules/INDEX.yaml, AGENTS.md
git add .reins AGENTS.md rules/ Taskfile.yml
git commit -m "chore: init reins framework"
```

### What the CLI embeds

The binary embeds three categories of content via `go:embed`:

| Category | Embedded path | Destination | Overwrite behavior |
|----------|--------------|-------------|-------------------|
| **Managed** | `content/managed/` | `.reins/` | Overwritten on `reins update` |
| **Scaffold** | `content/scaffold/` | Project root | Created once, never overwritten |
| **Templates** | `content/templates/` | `.reins/templates/` | Refreshed on `reins update` |

**Managed files** are owned by reins: `METHODOLOGY.md`, `agents/rule-guard.md`,
`rules/principles/quality.md`. These are refreshed to the latest version
when the user runs `reins update`.

**Scaffold files** are project-owned: `AGENTS.md`, `.editorconfig`,
`Taskfile.yml`, `rules/INDEX.yaml`, `AUTOPILOT.md`. Created once during
`reins init`, never touched again. The user customizes these for their project.

**Template files** are language/framework rule templates (e.g.,
`specifics/go.md`) that the user manually copies into `rules/specifics/`.

### What lives where (downstream project)

| Location | Content | Managed by |
|----------|---------|------------|
| `.reins/METHODOLOGY.md` | Core methodology (TDD, commits, quality) | Reins CLI |
| `.reins/rules/principles/` | Universal quality principles | Reins CLI |
| `.reins/agents/rule-guard.md` | Independent verification agent | Reins CLI |
| `.reins/templates/` | Language rule templates for manual copying | Reins CLI |
| `.reins/VERSION` | Installed reins version | Reins CLI |
| `AGENTS.md` | Bridge file — loads `.reins/METHODOLOGY.md` | Project |
| `.editorconfig` | Editor settings (indentation, charset, whitespace) | Project |
| `AUTOPILOT.md` | Autopilot goal, constraints, iteration protocol | Project |
| `rules/INDEX.yaml` | Trigger mapping for this project | Project |
| `rules/specifics/` | Language/framework rules | Project |
| `rules/concerns/` | Cross-cutting rules | Project |
| `Taskfile.yml` | Build/lint/test automation | Project |

### How AI tools discover reins

| Tool | Entry point | Mechanism |
|------|------------|-----------|
| OpenCode | `AGENTS.md` | Auto-read at project root; bridge points to `.reins/METHODOLOGY.md` |
| Claude Code | `CLAUDE.md` | Copy from `AGENTS.md`; Claude Code auto-reads CLAUDE.md at project root |
| Cursor | `.cursorrules` | Reference `.reins/METHODOLOGY.md` in project instructions |
| Any AI agent | System prompt | Load `.reins/METHODOLOGY.md` content manually |

### Updating reins

**Binary:**

```bash
curl -fsSL https://raw.githubusercontent.com/pastel-sketchbook/reins/main/install.sh | sh
reins update
```

**From source:**

```bash
go install github.com/pastel-sketchbook/reins/cmd/reins@latest
reins update
```

`reins update` overwrites managed files in `.reins/` and refreshes
templates, but never touches project-owned files. To pick up scaffold
improvements, diff manually:

```bash
diff AGENTS.md .reins/scaffold/AGENTS.md
diff rules/INDEX.yaml .reins/scaffold/rules/INDEX.yaml
diff AUTOPILOT.md .reins/scaffold/AUTOPILOT.md
```

### CLI reference

| Command | Description |
|---------|-------------|
| `reins init` | Bootstrap reins in the current project |
| `reins update` | Refresh managed files to latest version |
| `reins list` | List available language/framework templates |
| `reins version` | Print installed reins version |

---

## What Reins Is Not

- **Not a replacement for static analysis, linting, or tests.** Those are
  the foundation. Reins layers semantic verification on top.

- **Not an AI agent.** It is instructions, rules, and infrastructure that
  shape how an AI agent behaves.

- **Not documentation about the project.** METHODOLOGY.md tells the agent how
  to work. Rules tell it what constraints to follow. Neither describes
  the system architecture in narrative form.

- **Not prescriptive about AI tooling.** The components (hooks, reporting,
  observability) are described in terms of what they do, not which vendor
  implements them. Adapt to your tool.
