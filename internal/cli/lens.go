// Package cli — lens subcommand: generate analysis-lens concern templates.
package cli

import (
	"fmt"
	"strings"
)

// Lens identifies one of the 10 research-analysis lenses.
type Lens int

const (
	ExpertSynthesizer Lens = iota
	StakeholderTranslator
	TimelineConstructor
	EvidenceMapper
	ContradictionHunter
	AssumptionExcavator
	WeaknessSpotter
	FrameworkBuilder
	ImplementationBlueprint
	QuestionGenerator
	lensCount // sentinel — must be last
)

// Preset is a curated bundle of lenses.
type Preset int

const (
	PresetQuickSynthesis Preset = iota
	PresetDueDiligence
	PresetStrategicPlanning
	PresetFullProtocol
)

// allLenses returns every lens in canonical order.
func allLenses() []Lens {
	out := make([]Lens, 0, int(lensCount))
	for l := range Lens(lensCount) {
		out = append(out, l)
	}
	return out
}

// expandPreset returns the lenses that make up a preset bundle.
func expandPreset(p Preset) []Lens {
	switch p {
	case PresetQuickSynthesis:
		return []Lens{ExpertSynthesizer, ImplementationBlueprint}
	case PresetDueDiligence:
		return []Lens{EvidenceMapper, ContradictionHunter, AssumptionExcavator, WeaknessSpotter}
	case PresetStrategicPlanning:
		return []Lens{ExpertSynthesizer, FrameworkBuilder, ImplementationBlueprint, QuestionGenerator}
	case PresetFullProtocol:
		return allLenses()
	default:
		return nil
	}
}

// resolveLenses merges a preset (if any) with individual lens selections.
// The result is deduplicated and sorted in canonical order (Synthesizers →
// Auditors → Architects).
func resolveLenses(preset *Preset, individual []Lens) []Lens {
	seen := make(map[Lens]bool)
	if preset != nil {
		for _, l := range expandPreset(*preset) {
			seen[l] = true
		}
	}
	for _, l := range individual {
		seen[l] = true
	}
	if len(seen) == 0 {
		return nil
	}
	var out []Lens
	for l := range Lens(lensCount) {
		if seen[l] {
			out = append(out, l)
		}
	}
	return out
}

// lensName returns the human-readable display name.
func lensName(l Lens) string {
	names := [...]string{
		"Expert Synthesizer",
		"Stakeholder Translator",
		"Timeline Constructor",
		"Evidence Mapper",
		"Contradiction Hunter",
		"Assumption Excavator",
		"Weakness Spotter",
		"Framework Builder",
		"Implementation Blueprint",
		"Question Generator",
	}
	if int(l) < len(names) {
		return names[l]
	}
	return "Unknown"
}

// lensAlias returns the short CLI alias (e.g., "synth", "weak").
func lensAlias(l Lens) string {
	aliases := [...]string{
		"synth", "stake", "time",
		"evidence", "contra", "assume", "weak",
		"frame", "impl", "quest",
	}
	if int(l) < len(aliases) {
		return aliases[l]
	}
	return ""
}

// lensCategory returns the category label for grouping.
func lensCategory(l Lens) string {
	switch {
	case l <= TimelineConstructor:
		return "Synthesizers"
	case l <= WeaknessSpotter:
		return "Auditors"
	default:
		return "Architects"
	}
}

// lensRuleID returns the C-LENS-NN rule ID for a lens.
func lensRuleID(l Lens) string {
	return fmt.Sprintf("C-LENS-%02d", int(l)+1)
}

// parseLensAlias resolves a CLI alias or full name to a Lens.
// Returns -1 if unknown.
func parseLensAlias(s string) Lens {
	s = strings.ToLower(strings.TrimSpace(s))
	for l := range Lens(lensCount) {
		if lensAlias(l) == s || strings.EqualFold(lensName(l), s) {
			return l
		}
	}
	return -1
}

// parsePresetAlias resolves a CLI alias to a Preset. Returns nil if unknown.
func parsePresetAlias(s string) *Preset {
	s = strings.ToLower(strings.TrimSpace(s))
	aliases := map[string]Preset{
		"quick-synthesis":    PresetQuickSynthesis,
		"quick":              PresetQuickSynthesis,
		"due-diligence":      PresetDueDiligence,
		"dd":                 PresetDueDiligence,
		"strategic-planning": PresetStrategicPlanning,
		"strat":              PresetStrategicPlanning,
		"full-protocol":      PresetFullProtocol,
		"all":                PresetFullProtocol,
	}
	if p, ok := aliases[s]; ok {
		return &p
	}
	return nil
}

// presetName returns the display name for a preset.
func presetName(p Preset) string {
	names := [...]string{
		"quick-synthesis",
		"due-diligence",
		"strategic-planning",
		"full-protocol",
	}
	if int(p) < len(names) {
		return names[p]
	}
	return "unknown"
}

// presetAlias returns the short alias for a preset.
func presetAlias(p Preset) string {
	aliases := [...]string{"quick", "dd", "strat", "all"}
	if int(p) < len(aliases) {
		return aliases[p]
	}
	return ""
}

// lensPrompt returns the concern rule text for a single lens.
func lensPrompt(l Lens) string {
	prompts := [...]string{
		// ExpertSynthesizer
		`Identify the 3 core groundbreaking insights that an experienced
  practitioner would immediately recognize. For each insight: state it in
  one sentence, explain why it matters (1-2 sentences), and describe what
  conventional wisdom it challenges (1-2 sentences). Prioritize depth
  over breadth. Focus on what makes this codebase genuinely interesting
  or unusual.`,

		// StakeholderTranslator
		`Translate the codebase's key insights for 3 distinct audiences:
  executives, engineers, and end-users. For each audience: reframe the
  core findings using that audience's language and priorities, use
  relevant examples, and focus on what they specifically care about.
  One codebase, three usable summaries.`,

		// TimelineConstructor
		`Extract every temporal signal: commit patterns, version markers,
  TODO/FIXME comments with dates, changelog entries, dependency version
  evolution, and architectural shifts visible in the code. Produce a
  chronological table with dates/periods, events, and significance.
  Mark acceleration points where progress increased dramatically. Note
  stagnation periods or abandoned directions.`,

		// EvidenceMapper
		`For every major architectural decision and design claim, extract the
  supporting evidence. Produce a table with columns: Claim/Decision,
  Evidence (tests, docs, benchmarks, usage patterns), Strength
  (anecdotal / correlational / experimental / well-tested), and
  Confidence Flag (flag claims with weak evidence stated with high
  confidence).`,

		// ContradictionHunter
		`Compare all modules, documentation, comments, and code patterns to
  find every point of contradiction. For each: state the contradiction,
  cite Location A and Location B (file:line or doc section), and
  determine which position has stronger evidence and why (or why both
  might be valid).`,

		// AssumptionExcavator
		`Identify every unstated assumption — architectural, environmental,
  about users, about scale, about dependencies. Produce a table sorted
  by criticality (descending) with columns: Assumption, Criticality
  (1-10), Likelihood Wrong (1-10), and What Changes If False (concrete
  consequences).`,

		// WeaknessSpotter
		`Act as a harsh peer reviewer. Identify every methodological flaw,
  logical gap, overclaim, and unsupported leap. For each: classify the
  flaw type (methodological / logical / overclaim / unsupported leap /
  missing test / security gap), cite location (file:line or module),
  rate severity (minor / moderate / critical), and state what evidence
  or change would fix it.`,

		// FrameworkBuilder
		`Create a comprehensive framework integrating all concepts. Document
  key components and their responsibilities, relationships between
  components (ASCII diagram encouraged), a decision tree for when to
  use which component, and edge cases where the framework breaks down
  or requires special handling.`,

		// ImplementationBlueprint
		`Extract every actionable insight and organize into a step-by-step
  implementation plan for someone building on or replicating this
  codebase's approach. For each step: state the action, list
  prerequisites, describe the expected outcome, and identify potential
  pitfalls.`,

		// QuestionGenerator
		`Generate 15 questions that an expert would ask about this codebase
  that the code itself does NOT answer. Prioritize by potential impact.
  For each question: state it clearly and provide a one-sentence
  rationale explaining why this question matters.`,
	}
	if int(l) < len(prompts) {
		return prompts[l]
	}
	return ""
}

// crossLensSynthesisPrompt returns the synthesis instruction appended
// when 3 or more lenses are active.
func crossLensSynthesisPrompt() string {
	return `After completing all individual lens analyses, synthesize the findings
  into a unified view:
  1. **Key Themes**: What patterns emerge across multiple lenses?
  2. **Critical Gaps**: What important areas were flagged by multiple
     lenses as problematic?
  3. **Recommended Next Steps**: Based on all analysis, what are the
     3-5 most important actions to take?
  Be opinionated — state which findings are most important and why.`
}

// buildConcernFile assembles a complete concern template from the selected
// lenses. The output follows the reins concern format with C-LENS-NN rule IDs.
func buildConcernFile(lenses []Lens, source string) string {
	var b strings.Builder

	b.WriteString("# Analysis Lenses — Cross-Cutting Concern\n\n")
	b.WriteString("Instructions for applying research-analysis lenses to codebase review.\n")
	if source != "" {
		b.WriteString("Generated by `reins lens " + source + "`. ")
	}
	b.WriteString("Language-agnostic.\n\n---\n")

	currentCategory := ""
	for _, l := range lenses {
		cat := lensCategory(l)
		if cat != currentCategory {
			b.WriteString("\n## " + strings.ToUpper(cat) + "\n")
			currentCategory = cat
		}
		b.WriteString("\n- **" + lensRuleID(l) + "** — " + lensPrompt(l) + "\n")
	}

	if len(lenses) >= 3 {
		b.WriteString("\n## CROSS-LENS SYNTHESIS\n\n")
		b.WriteString("- **C-LENS-11** — " + crossLensSynthesisPrompt() + "\n")
	}

	return b.String()
}

// printLensList prints available lenses and presets to stdout.
func printLensList() {
	fmt.Println("Available analysis lenses:")
	fmt.Println()

	currentCategory := ""
	for l := range Lens(lensCount) {
		cat := lensCategory(l)
		if cat != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			fmt.Printf("  %s\n", cat)
			currentCategory = cat
		}
		fmt.Printf("    %-28s  (%s)\n", lensName(l), lensAlias(l))
	}

	fmt.Println()
	fmt.Println("Presets:")
	fmt.Println()
	fmt.Printf("  %-22s  (quick)   Expert Synthesizer + Implementation Blueprint\n", "quick-synthesis")
	fmt.Printf("  %-22s  (dd)      Evidence Mapper, Contradiction Hunter, Assumption Excavator, Weakness Spotter\n", "due-diligence")
	fmt.Printf("  %-22s  (strat)   Expert Synthesizer, Framework Builder, Implementation Blueprint, Question Generator\n", "strategic-planning")
	fmt.Printf("  %-22s  (all)     All 10 lenses\n", "full-protocol")

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  reins lens --preset quick                # quick synthesis")
	fmt.Println("  reins lens --preset dd                   # due-diligence audit")
	fmt.Println("  reins lens --lens synth --lens weak      # cherry-pick lenses")
	fmt.Println("  reins lens --preset dd --lens synth      # merge preset + individual")
	fmt.Println("  reins lens --output path/to/file.md      # custom output path")
	fmt.Println()
}
