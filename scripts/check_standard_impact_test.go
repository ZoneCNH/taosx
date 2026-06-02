package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStandardImpactRequiresDownstreamSyncForHarnessGeneratorEvidence(t *testing.T) {
	report := runStandardImpact(t, []string{
		"scripts/check_docs.sh",
		"scripts/render_template.sh",
		"internal/tools/releasemanifest/main.go",
	})

	assertReportContains(t, report,
		"- downstream_sync_required: `true`",
		"- changed_file_count: `3`",
		"## harness",
		"- `scripts/check_docs.sh`",
		"## generator",
		"- `scripts/render_template.sh`",
		"## evidence",
		"- `internal/tools/releasemanifest/main.go`",
		"- `downstream-sync-required`",
	)
}

func TestStandardImpactUsesUpstreamMergeBaseForCleanBranches(t *testing.T) {
	scriptsDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	scriptPath := filepath.Join(scriptsDir, "check_standard_impact.sh")

	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		t.Fatalf("create temp repo: %v", err)
	}

	runGit(t, repoDir, "init", "-b", "main")
	runGit(t, repoDir, "config", "user.name", "Standard Impact Test")
	runGit(t, repoDir, "config", "user.email", "standard-impact@example.com")
	writeFixtureFile(t, repoDir, "README.md")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "base")

	runGit(t, repoDir, "switch", "-c", "feature")
	runGit(t, repoDir, "branch", "--set-upstream-to=main")
	writeFixtureFile(t, repoDir, "scripts/check_docs.sh")
	writeFixtureFile(t, repoDir, "docs/standard/README.md")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "feature")

	reportPath := filepath.Join(tempDir, "standard-impact.md")
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(),
		"STANDARD_IMPACT_REPORT="+reportPath,
		"STANDARD_IMPACT_BASE=",
		"STANDARD_IMPACT_GENERATED_AT=2026-06-02T00:00:00Z",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("standard impact check failed: %v\n%s", err, output)
	}

	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	assertReportContains(t, string(report),
		"- generated_at: `2026-06-02T00:00:00Z`",
		"- downstream_sync_required: `true`",
		"- changed_file_count: `2`",
		"## docs",
		"- `docs/standard/README.md`",
		"## harness",
		"- `scripts/check_docs.sh`",
		"- `downstream-sync-required`",
	)
}

func TestStandardImpactIgnoresLocalAgentRuntimeState(t *testing.T) {
	report := runStandardImpact(t, []string{
		".omc/state/mission-state.json",
		".omx/state/ralph-progress.json",
		".worktree/scratch/README.md",
		"docs/standard/README.md",
	})

	assertReportContains(t, report,
		"- changed_file_count: `1`",
		"## docs",
		"- `docs/standard/README.md`",
		"- `downstream-sync-not-required`",
	)

	for _, localStatePath := range []string{".omc/", ".omx/", ".worktree/"} {
		if strings.Contains(report, localStatePath) {
			t.Fatalf("report included local runtime state %q:\n%s", localStatePath, report)
		}
	}
}

func TestStandardImpactSortsCommittedAndWorktreeChanges(t *testing.T) {
	scriptsDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	scriptPath := filepath.Join(scriptsDir, "check_standard_impact.sh")

	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		t.Fatalf("create temp repo: %v", err)
	}

	runGit(t, repoDir, "init", "-b", "main")
	runGit(t, repoDir, "config", "user.name", "Standard Impact Test")
	runGit(t, repoDir, "config", "user.email", "standard-impact@example.com")
	writeFixtureFile(t, repoDir, "README.md")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "base")

	runGit(t, repoDir, "switch", "-c", "feature")
	runGit(t, repoDir, "branch", "--set-upstream-to=main")
	writeFixtureFile(t, repoDir, "renovate.json")
	writeFixtureFile(t, repoDir, "scripts/check_dependency_diff_test.go")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "feature")
	writeFixtureFile(t, repoDir, ".github/dependabot.yml")
	writeFixtureFile(t, repoDir, ".gitignore")

	reportPath := filepath.Join(tempDir, "standard-impact.md")
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(),
		"STANDARD_IMPACT_REPORT="+reportPath,
		"STANDARD_IMPACT_BASE=",
		"STANDARD_IMPACT_GENERATED_AT=2026-06-02T00:00:00Z",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("standard impact check failed: %v\n%s", err, output)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	report := string(reportBytes)

	assertReportContains(t, report,
		"- changed_file_count: `4`",
		"## other",
	)
	assertReportOrder(t, report,
		"- `.github/dependabot.yml`",
		"- `.gitignore`",
		"- `renovate.json`",
		"- `scripts/check_dependency_diff_test.go`",
	)
}

func TestStandardImpactDoesNotRequireDownstreamSyncForDocsOnly(t *testing.T) {
	report := runStandardImpact(t, []string{
		"README.md",
		"docs/standard/README.md",
	})

	assertReportContains(t, report,
		"- downstream_sync_required: `false`",
		"- changed_file_count: `2`",
		"## docs",
		"- `README.md`",
		"- `docs/standard/README.md`",
		"- `downstream-sync-not-required`",
	)

	if strings.Contains(report, "- `downstream-sync-required`") {
		t.Fatalf("docs-only report unexpectedly required downstream sync:\n%s", report)
	}
}

func TestStandardImpactReportIncludesDecisionEvidence(t *testing.T) {
	report := runStandardImpact(t, []string{
		"contracts/template.md",
		"docs/supply-chain.md",
		"pkg/templatex/client.go",
	})

	assertReportContains(t, report,
		"- primary_downstream: `github.com/ZoneCNH/kernel`",
		"- changed_file_count: `3`",
		"## contracts",
		"- `contracts/template.md`",
		"## docs",
		"- `docs/supply-chain.md`",
		"## other",
		"- `pkg/templatex/client.go`",
		"## Sync Decision",
	)
}

func runStandardImpact(t *testing.T, changedFiles []string) string {
	t.Helper()

	scriptsDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	scriptPath := filepath.Join(scriptsDir, "check_standard_impact.sh")

	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		t.Fatalf("create temp repo: %v", err)
	}

	initCmd := exec.Command("git", "init")
	initCmd.Dir = repoDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}

	for _, file := range changedFiles {
		path := filepath.Join(repoDir, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create parent dir for %s: %v", file, err)
		}
		if err := os.WriteFile(path, []byte("test fixture\n"), 0o644); err != nil {
			t.Fatalf("write changed file %s: %v", file, err)
		}
	}

	reportPath := filepath.Join(tempDir, "standard-impact.md")
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(),
		"STANDARD_IMPACT_REPORT="+reportPath,
		"STANDARD_IMPACT_BASE=",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("standard impact check failed: %v\n%s", err, output)
	}

	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	return string(report)
}

func writeFixtureFile(t *testing.T, repoDir, file string) {
	t.Helper()

	path := filepath.Join(repoDir, filepath.FromSlash(file))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent dir for %s: %v", file, err)
	}
	if err := os.WriteFile(path, []byte("test fixture\n"), 0o644); err != nil {
		t.Fatalf("write fixture file %s: %v", file, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func assertReportContains(t *testing.T, report string, want ...string) {
	t.Helper()

	for _, text := range want {
		if !strings.Contains(report, text) {
			t.Fatalf("report missing %q:\n%s", text, report)
		}
	}
}

func assertReportOrder(t *testing.T, report string, want ...string) {
	t.Helper()

	previousIndex := -1
	for _, text := range want {
		index := strings.Index(report, text)
		if index < 0 {
			t.Fatalf("report missing %q:\n%s", text, report)
		}
		if index <= previousIndex {
			t.Fatalf("report order mismatch at %q:\n%s", text, report)
		}
		previousIndex = index
	}
}
