package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_CreatesExpectedFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	code := runInit()
	if code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	wantFiles := []string{
		// Managed files
		".reins/AGENTS.md",
		".reins/agents/rule-guard.md",
		".reins/rules/principles/quality.md",
		".reins/VERSION",
		".reins/templates/specifics/go.md",
		// Scaffold files
		"CLAUDE.md",
		"Taskfile.yml",
		"rules/INDEX.yaml",
	}

	for _, f := range wantFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}

	wantDirs := []string{
		"rules/principles",
		"rules/concerns",
		"rules/specifics",
	}

	for _, d := range wantDirs {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}
}

func TestRunInit_RejectsDoubleInit(t *testing.T) {
	t.Chdir(t.TempDir())

	if code := runInit(); code != 0 {
		t.Fatalf("first runInit() = %d, want 0", code)
	}

	code := runInit()
	if code != 1 {
		t.Errorf("second runInit() = %d, want 1", code)
	}
}

func TestRunInit_WritesVersion(t *testing.T) {
	t.Chdir(t.TempDir())

	SetVersion("1.2.3")
	t.Cleanup(func() { SetVersion("dev") })

	if code := runInit(); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	data, err := os.ReadFile(".reins/VERSION")
	if err != nil {
		t.Fatalf("failed to read VERSION: %v", err)
	}
	if got := string(data); got != "1.2.3\n" {
		t.Errorf("VERSION = %q, want %q", got, "1.2.3\n")
	}
}

func TestRunUpdate_RefreshesManagedFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	if code := runInit(); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Tamper with a managed file.
	agentsPath := filepath.Join(".reins", "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("tampered"), 0o644); err != nil {
		t.Fatalf("failed to tamper with AGENTS.md: %v", err)
	}

	code := runUpdate()
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read AGENTS.md after update: %v", err)
	}
	if string(data) == "tampered" {
		t.Error("AGENTS.md was not refreshed by update")
	}
}

func TestRunUpdate_DoesNotOverwriteProjectFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	if code := runInit(); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Tamper with a project-owned file.
	customContent := []byte("# My custom CLAUDE.md\n")
	if err := os.WriteFile("CLAUDE.md", customContent, 0o644); err != nil {
		t.Fatalf("failed to write custom CLAUDE.md: %v", err)
	}

	code := runUpdate()
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	// Project file should be untouched.
	data, err := os.ReadFile("CLAUDE.md")
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md after update: %v", err)
	}
	if string(data) != string(customContent) {
		t.Errorf("CLAUDE.md was modified by update:\ngot:  %q\nwant: %q", string(data), string(customContent))
	}
}

func TestRunUpdate_FailsWithoutInit(t *testing.T) {
	t.Chdir(t.TempDir())

	code := runUpdate()
	if code != 1 {
		t.Errorf("runUpdate() without init = %d, want 1", code)
	}
}

func TestRunList_Succeeds(t *testing.T) {
	code := runList()
	if code != 0 {
		t.Errorf("runList() = %d, want 0", code)
	}
}

func TestRun_Version(t *testing.T) {
	code := Run([]string{"reins", "version"})
	if code != 0 {
		t.Errorf("Run version = %d, want 0", code)
	}
}

func TestRun_Help(t *testing.T) {
	code := Run([]string{"reins", "help"})
	if code != 0 {
		t.Errorf("Run help = %d, want 0", code)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	code := Run([]string{"reins", "bogus"})
	if code != 1 {
		t.Errorf("Run bogus = %d, want 1", code)
	}
}

func TestRun_NoArgs(t *testing.T) {
	code := Run([]string{"reins"})
	if code != 0 {
		t.Errorf("Run no args = %d, want 0", code)
	}
}
