package cli

import (
	"fmt"
	"strings"
	"testing"
)

// ── expandPreset ──────────────────────────────────────────────────────────

func TestExpandPreset_QuickSynthesis(t *testing.T) {
	got := expandPreset(PresetQuickSynthesis)
	want := []Lens{ExpertSynthesizer, ImplementationBlueprint}
	assertLensSlice(t, got, want)
}

func TestExpandPreset_DueDiligence(t *testing.T) {
	got := expandPreset(PresetDueDiligence)
	want := []Lens{EvidenceMapper, ContradictionHunter, AssumptionExcavator, WeaknessSpotter}
	assertLensSlice(t, got, want)
}

func TestExpandPreset_StrategicPlanning(t *testing.T) {
	got := expandPreset(PresetStrategicPlanning)
	want := []Lens{ExpertSynthesizer, FrameworkBuilder, ImplementationBlueprint, QuestionGenerator}
	assertLensSlice(t, got, want)
}

func TestExpandPreset_FullProtocol_HasAll10(t *testing.T) {
	got := expandPreset(PresetFullProtocol)
	if len(got) != 10 {
		t.Errorf("FullProtocol has %d lenses, want 10", len(got))
	}
}

// ── resolveLenses ─────────────────────────────────────────────────────────

func TestResolveLenses_Empty(t *testing.T) {
	got := resolveLenses(nil, nil)
	if got != nil {
		t.Errorf("resolveLenses(nil, nil) = %v, want nil", got)
	}
}

func TestResolveLenses_PresetOnly(t *testing.T) {
	p := PresetQuickSynthesis
	got := resolveLenses(&p, nil)
	want := []Lens{ExpertSynthesizer, ImplementationBlueprint}
	assertLensSlice(t, got, want)
}

func TestResolveLenses_IndividualOnly_CanonicalOrder(t *testing.T) {
	got := resolveLenses(nil, []Lens{WeaknessSpotter, ExpertSynthesizer})
	want := []Lens{ExpertSynthesizer, WeaknessSpotter}
	assertLensSlice(t, got, want)
}

func TestResolveLenses_MergeDedup(t *testing.T) {
	p := PresetQuickSynthesis
	// ExpertSynthesizer is already in QuickSynthesis — should not duplicate.
	got := resolveLenses(&p, []Lens{ExpertSynthesizer, WeaknessSpotter})
	want := []Lens{ExpertSynthesizer, WeaknessSpotter, ImplementationBlueprint}
	assertLensSlice(t, got, want)
}

func TestResolveLenses_CanonicalOrderStable(t *testing.T) {
	// Add lenses in reverse order; should come out canonical.
	got := resolveLenses(nil, []Lens{QuestionGenerator, WeaknessSpotter, ExpertSynthesizer})
	want := []Lens{ExpertSynthesizer, WeaknessSpotter, QuestionGenerator}
	assertLensSlice(t, got, want)
}

// ── lensName / lensAlias / lensCategory ──────────────────────────────────

func TestLensName_AllDefined(t *testing.T) {
	for l := range Lens(lensCount) {
		name := lensName(l)
		if name == "" || name == "Unknown" {
			t.Errorf("lens %d has no name", l)
		}
	}
}

func TestLensAlias_AllDefined(t *testing.T) {
	for l := range Lens(lensCount) {
		alias := lensAlias(l)
		if alias == "" {
			t.Errorf("lens %d has no alias", l)
		}
	}
}

func TestLensCategory_Correct(t *testing.T) {
	cases := []struct {
		lens Lens
		want string
	}{
		{ExpertSynthesizer, "Synthesizers"},
		{StakeholderTranslator, "Synthesizers"},
		{TimelineConstructor, "Synthesizers"},
		{EvidenceMapper, "Auditors"},
		{ContradictionHunter, "Auditors"},
		{AssumptionExcavator, "Auditors"},
		{WeaknessSpotter, "Auditors"},
		{FrameworkBuilder, "Architects"},
		{ImplementationBlueprint, "Architects"},
		{QuestionGenerator, "Architects"},
	}
	for _, tc := range cases {
		got := lensCategory(tc.lens)
		if got != tc.want {
			t.Errorf("lensCategory(%s) = %q, want %q", lensName(tc.lens), got, tc.want)
		}
	}
}

func TestLensRuleID_Format(t *testing.T) {
	if got := lensRuleID(ExpertSynthesizer); got != "C-LENS-01" {
		t.Errorf("got %q, want C-LENS-01", got)
	}
	if got := lensRuleID(QuestionGenerator); got != "C-LENS-10" {
		t.Errorf("got %q, want C-LENS-10", got)
	}
}

// ── parseLensAlias ────────────────────────────────────────────────────────

func TestParseLensAlias_Aliases(t *testing.T) {
	cases := []struct {
		input string
		want  Lens
	}{
		{"synth", ExpertSynthesizer},
		{"stake", StakeholderTranslator},
		{"time", TimelineConstructor},
		{"evidence", EvidenceMapper},
		{"contra", ContradictionHunter},
		{"assume", AssumptionExcavator},
		{"weak", WeaknessSpotter},
		{"frame", FrameworkBuilder},
		{"impl", ImplementationBlueprint},
		{"quest", QuestionGenerator},
	}
	for _, tc := range cases {
		got := parseLensAlias(tc.input)
		if got != tc.want {
			t.Errorf("parseLensAlias(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseLensAlias_Unknown(t *testing.T) {
	got := parseLensAlias("nonexistent")
	if got != -1 {
		t.Errorf("parseLensAlias(nonexistent) = %d, want -1", got)
	}
}

// ── parsePresetAlias ──────────────────────────────────────────────────────

func TestParsePresetAlias_All(t *testing.T) {
	cases := []struct {
		input string
		want  Preset
	}{
		{"quick-synthesis", PresetQuickSynthesis},
		{"quick", PresetQuickSynthesis},
		{"due-diligence", PresetDueDiligence},
		{"dd", PresetDueDiligence},
		{"strategic-planning", PresetStrategicPlanning},
		{"strat", PresetStrategicPlanning},
		{"full-protocol", PresetFullProtocol},
		{"all", PresetFullProtocol},
	}
	for _, tc := range cases {
		got := parsePresetAlias(tc.input)
		if got == nil {
			t.Errorf("parsePresetAlias(%q) = nil, want %d", tc.input, tc.want)
			continue
		}
		if *got != tc.want {
			t.Errorf("parsePresetAlias(%q) = %d, want %d", tc.input, *got, tc.want)
		}
	}
}

func TestParsePresetAlias_Unknown(t *testing.T) {
	got := parsePresetAlias("nonexistent")
	if got != nil {
		t.Errorf("parsePresetAlias(nonexistent) = %v, want nil", got)
	}
}

// ── presetName / presetAlias ──────────────────────────────────────────────

func TestPresetName(t *testing.T) {
	if got := presetName(PresetDueDiligence); got != "due-diligence" {
		t.Errorf("got %q, want due-diligence", got)
	}
}

func TestPresetAlias(t *testing.T) {
	if got := presetAlias(PresetDueDiligence); got != "dd" {
		t.Errorf("got %q, want dd", got)
	}
}

// ── lensPrompt ────────────────────────────────────────────────────────────

func TestLensPrompt_AllNonEmpty(t *testing.T) {
	for l := range Lens(lensCount) {
		prompt := lensPrompt(l)
		if prompt == "" {
			t.Errorf("lens %d (%s) has empty prompt", l, lensName(l))
		}
	}
}

// ── buildConcernFile ──────────────────────────────────────────────────────

func TestBuildConcernFile_Header(t *testing.T) {
	lenses := expandPreset(PresetDueDiligence)
	got := buildConcernFile(lenses, "--preset dd")
	if !strings.Contains(got, "# Analysis Lenses") {
		t.Error("missing header")
	}
	if !strings.Contains(got, "reins lens --preset dd") {
		t.Error("missing source attribution")
	}
}

func TestBuildConcernFile_ContainsAllSelectedLenses(t *testing.T) {
	lenses := expandPreset(PresetDueDiligence)
	got := buildConcernFile(lenses, "")
	for _, l := range lenses {
		if !strings.Contains(got, lensRuleID(l)) {
			t.Errorf("missing rule ID %s for %s", lensRuleID(l), lensName(l))
		}
	}
}

func TestBuildConcernFile_CategoryHeaders(t *testing.T) {
	lenses := []Lens{ExpertSynthesizer, WeaknessSpotter, FrameworkBuilder}
	got := buildConcernFile(lenses, "")
	for _, cat := range []string{"SYNTHESIZERS", "AUDITORS", "ARCHITECTS"} {
		if !strings.Contains(got, cat) {
			t.Errorf("missing category header %s", cat)
		}
	}
}

func TestBuildConcernFile_CrossLensSynthesis_3Plus(t *testing.T) {
	lenses := []Lens{ExpertSynthesizer, WeaknessSpotter, FrameworkBuilder}
	got := buildConcernFile(lenses, "")
	if !strings.Contains(got, "CROSS-LENS SYNTHESIS") {
		t.Error("3+ lenses should include cross-lens synthesis")
	}
	if !strings.Contains(got, "C-LENS-11") {
		t.Error("cross-lens synthesis should have rule ID C-LENS-11")
	}
}

func TestBuildConcernFile_NoCrossLensForFewerThan3(t *testing.T) {
	lenses := expandPreset(PresetQuickSynthesis)
	got := buildConcernFile(lenses, "")
	if strings.Contains(got, "CROSS-LENS SYNTHESIS") {
		t.Error("fewer than 3 lenses should not include cross-lens synthesis")
	}
}

func TestBuildConcernFile_FullProtocol_AllRuleIDs(t *testing.T) {
	lenses := expandPreset(PresetFullProtocol)
	got := buildConcernFile(lenses, "")
	for i := 1; i <= 10; i++ {
		id := fmt.Sprintf("C-LENS-%02d", i)
		if !strings.Contains(got, id) {
			t.Errorf("full protocol missing rule ID %s", id)
		}
	}
	// Plus cross-lens synthesis.
	if !strings.Contains(got, "C-LENS-11") {
		t.Error("full protocol should include C-LENS-11")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

func assertLensSlice(t *testing.T, got, want []Lens) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %d (%s), want %d (%s)", i, got[i], lensName(got[i]), want[i], lensName(want[i]))
		}
	}
}
