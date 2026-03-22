# Autopilot

This file defines the goal, constraints, and iteration protocol for
autonomous agent sessions. When the agent is given this file as its
prompt, it operates in a self-directed loop until the goal is met or
the constraints are exhausted.

**This file is project-owned.** Reins creates it once during `reins init`.
Edit it freely to match your project's domain and workflow.

---

## Goal

<!-- Replace with your objective. Be specific and measurable. -->

Example: "Reduce test suite execution time below 30 seconds while
maintaining 100% pass rate."

## Constraints

1. **Fixed boundaries** — What the agent MUST NOT change.
   <!-- e.g., public API signatures, database schema, CI config -->

2. **Allowed changes** — What the agent CAN modify.
   <!-- e.g., internal implementation, test helpers, configuration -->

3. **Time/resource budget** — Optional hard limits.
   <!-- e.g., "Complete within 10 iterations", "Do not add dependencies" -->

4. **Quality gate** — `task check:all` MUST pass after every change.

## Iteration Protocol

For each iteration:

1. **Hypothesize** — State what you will change and the expected effect.
2. **Implement** — Make the smallest change that tests the hypothesis.
3. **Verify** — Run `task check:all`. If it fails, fix or revert before
   proceeding.
4. **Evaluate** — Did the change move toward the goal? Record the result.
5. **Decide** — Continue with the next hypothesis, or stop if the goal
   is met or no viable hypotheses remain.

## Success Criteria

<!-- Define what "done" looks like. Measurable is better than subjective. -->

Example: "Validation loss below 1.15 BPB within the time budget" or
"All endpoints respond under 200ms at p99 with zero test regressions."

## Notes

- If an iteration fails (tests break, build errors), revert to the last
  known-good state before trying a new approach.
- Prefer small, reversible changes over large speculative rewrites.
- The agent SHOULD consult `AGENTS.md` rules throughout — autopilot does
  not exempt the agent from TDD, commit discipline, or quality rules.
