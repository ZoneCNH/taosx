package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestBuildChecksUsesGlobalAndSpecificStatus(t *testing.T) {
	t.Setenv("CHECK_STATUS", "passed")
	t.Setenv("LINT_STATUS", "failed")
	t.Setenv("DOCS_CHECK_STATUS", "skipped")

	checks := buildChecks()

	if checks["fmt"] != "passed" {
		t.Fatalf("fmt status = %q, want passed", checks["fmt"])
	}
	if checks["lint"] != "failed" {
		t.Fatalf("lint status = %q, want failed", checks["lint"])
	}
	if checks["docs_check"] != "skipped" {
		t.Fatalf("docs_check status = %q, want skipped", checks["docs_check"])
	}
}

func TestCheckNamesCoverReleaseTargets(t *testing.T) {
	wantNames := []string{
		"fmt",
		"vet",
		"lint",
		"unit_test",
		"race_test",
		"boundary",
		"secret_scan",
		"security",
		"contract",
		"integration",
		"dependency_check",
		"standard_impact",
		"docs_check",
		"property",
		"golden",
		"fuzz_smoke",
	}

	if strings.Join(checkNames, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("checkNames = %v, want %v", checkNames, wantNames)
	}
	for _, name := range wantNames {
		if checkEnvNames[name] == "" {
			t.Fatalf("checkEnvNames[%q] is empty", name)
		}
	}
}

func TestValidateChecksRequiresPassedStatuses(t *testing.T) {
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = "passed"
	}
	checks["security"] = "unknown"

	failures := validateChecks(checks, true)

	if len(failures) != 1 {
		t.Fatalf("len(failures) = %d, want 1: %v", len(failures), failures)
	}
	if !strings.Contains(failures[0], "checks.security") {
		t.Fatalf("failure = %q, want security check failure", failures[0])
	}
}

func TestFileDigestRecordsPathAndSHA256(t *testing.T) {
	path := t.TempDir() + "/contract.json"
	if err := os.WriteFile(path, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}

	digest, err := fileDigest(path)
	if err != nil {
		t.Fatal(err)
	}

	if digest.Path != path {
		t.Fatalf("path = %q, want %q", digest.Path, path)
	}
	const want = "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if digest.SHA256 != want {
		t.Fatalf("sha256 = %q, want %q", digest.SHA256, want)
	}
}

func TestRunCLIGeneratesManifestToOut(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v1.2.3-cli")
	t.Setenv("GENERATED_BY", "releasemanifest-cli-test")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "custom", "latest.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-out", outPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if want := "generated " + outPath; !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("generated manifest is invalid JSON: %s", data)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Module != "example.com/releasefixture" {
		t.Fatalf("module = %q, want fixture module", manifest.Module)
	}
	if manifest.Version != "v1.2.3-cli" {
		t.Fatalf("version = %q, want v1.2.3-cli", manifest.Version)
	}
	if manifest.GeneratedBy != "releasemanifest-cli-test" {
		t.Fatalf("generated_by = %q, want releasemanifest-cli-test", manifest.GeneratedBy)
	}
	if manifest.Workflow.WorkflowRunID == "" || manifest.Workflow.ArtifactName == "" || manifest.Workflow.ArtifactURL == "" {
		t.Fatalf("workflow evidence is incomplete: %+v", manifest.Workflow)
	}
	if manifest.Score.Threshold != 9.8 || manifest.Score.Status == "" || len(manifest.Score.Dimensions) == 0 {
		t.Fatalf("score report is incomplete: %+v", manifest.Score)
	}
	if manifest.StandardImpact.ReportPath != standardImpactReportPath || manifest.StandardImpact.Status == "" {
		t.Fatalf("standard impact evidence is incomplete: %+v", manifest.StandardImpact)
	}
	if manifest.DownstreamSyncRequired != manifest.StandardImpact.DownstreamSyncRequired {
		t.Fatalf("downstream_sync_required = %t, want standard impact value %t", manifest.DownstreamSyncRequired, manifest.StandardImpact.DownstreamSyncRequired)
	}
	if !manifest.GeneratorEvidence.Required || manifest.GeneratorEvidence.Command == "" {
		t.Fatalf("generator evidence is incomplete: %+v", manifest.GeneratorEvidence)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "kernel", "github.com/ZoneCNH/kernel", "kernel") {
		t.Fatalf("generator targets = %+v, want kernel target", manifest.GeneratorEvidence.Targets)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "corekit", "example.com/acme/corekit", "corekit") {
		t.Fatalf("generator targets = %+v, want corekit target", manifest.GeneratorEvidence.Targets)
	}
	for _, name := range checkNames {
		if manifest.Checks[name] != "passed" {
			t.Fatalf("checks[%q] = %q, want passed", name, manifest.Checks[name])
		}
	}
}

func TestRunCLIGenerateReportsBuildManifestFailure(t *testing.T) {
	t.Setenv("GOWORK", "off")
	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/brokenmanifest\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repo, "git", "add", ".")
	chdir(t, repo)

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-out", outPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI generate exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	message := stderr.String()
	for _, want := range []string{"ERROR:", "contracts/config.schema.json"} {
		if !strings.Contains(message, want) {
			t.Fatalf("stderr = %q, want substring %q", message, want)
		}
	}
	if _, err := os.Stat(outPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("generated manifest exists after failed build: %v", err)
	}
}

func TestRunCLIGenerateReportsWriteManifestFailure(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-out", outPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI generate exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := "ERROR:"; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIVerifiesManifestWithRequirePassed(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v1.2.3")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-require-passed", "-expect-version", "v1.2.3"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI verify exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if want := "release evidence verified: " + outPath; !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
	}
}

func TestRunCLIVerifyRejectsScoreBelowMinimum(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-min-score", "9.8"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	for _, want := range []string{"release score", "below minimum"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
		}
	}
}

func TestBuildWorkflowEvidencePrefersExplicitEnvironment(t *testing.T) {
	t.Setenv("WORKFLOW_RUN_ID", "12345")
	t.Setenv("ARTIFACT_NAME", "manifest-artifact")
	t.Setenv("ARTIFACT_URL", "https://example.invalid/artifacts/manifest")

	got := buildWorkflowEvidence()

	if got.WorkflowRunID != "12345" || got.ArtifactName != "manifest-artifact" || got.ArtifactURL != "https://example.invalid/artifacts/manifest" {
		t.Fatalf("workflow evidence = %+v, want explicit env values", got)
	}
}

func TestRunCLIVerifyRejectsExpectedVersionMismatch(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v1.2.3")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-expect-version", "v9.9.9"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := `version mismatch: got "v1.2.3", want "v9.9.9"`; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIVerifyReportsDrift(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.SourceDigest = "sha256:stale"
	manifest.Checks["lint"] = "failed"
	manifest.StandardImpact.Status = "stale"
	manifest.DownstreamSyncRequired = !manifest.StandardImpact.DownstreamSyncRequired
	manifest.GeneratorEvidence.Targets = nil
	if err := writeManifest(outPath, manifest); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-require-passed"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	message := stderr.String()
	for _, want := range []string{
		"ERROR: release evidence verification failed",
		"source_digest does not match current tracked file contents",
		"standard_impact does not match current standard impact evidence",
		"generator_evidence does not match current integration evidence",
		`checks.lint must be passed, got "failed"`,
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("stderr = %q, want substring %q", message, want)
		}
	}
}

func TestRunCLIVerifyRequiresCleanTree(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.TreeState = "dirty"
	if err := writeManifest(outPath, manifest); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-require-clean"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := `tree_state must be clean, got "dirty"`; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIHelpReturnsSuccess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-h"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI help exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := "Usage of releasemanifest"; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("runCLI unknown flag exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := "flag provided but not defined"; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestPrintCLIMessageReportsWriterFailure(t *testing.T) {
	if code := printCLIStatus(errorWriter{}, "ok\n"); code != 1 {
		t.Fatalf("printCLIStatus exit code = %d, want 1", code)
	}
	if code := printCLIError(errorWriter{}, errors.New("boom")); code != 1 {
		t.Fatalf("printCLIError exit code = %d, want 1", code)
	}
}

func TestBuildManifestRecordsFixtureRepositoryFacts(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v9.9.9-test")
	t.Setenv("GENERATED_BY", "releasemanifest-test")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	if manifest.Module != "example.com/releasefixture" {
		t.Fatalf("module = %q, want example.com/releasefixture", manifest.Module)
	}
	if manifest.Version != "v9.9.9-test" {
		t.Fatalf("version = %q, want v9.9.9-test", manifest.Version)
	}
	if manifest.GeneratedBy != "releasemanifest-test" {
		t.Fatalf("generated_by = %q, want releasemanifest-test", manifest.GeneratedBy)
	}
	if _, err := time.Parse(time.RFC3339, manifest.GeneratedAt); err != nil {
		t.Fatalf("generated_at = %q, want RFC3339: %v", manifest.GeneratedAt, err)
	}
	if !strings.HasPrefix(manifest.SourceDigest, "sha256:") {
		t.Fatalf("source_digest = %q, want sha256 prefix", manifest.SourceDigest)
	}
	if manifest.TrackedFileCount == 0 {
		t.Fatal("tracked_file_count = 0, want tracked files")
	}
	if len(manifest.Contracts) != len(contractFiles) {
		t.Fatalf("len(contracts) = %d, want %d", len(manifest.Contracts), len(contractFiles))
	}
	if len(manifest.Dependencies) == 0 || manifest.Dependencies[0].Path != manifest.Module || !manifest.Dependencies[0].Main {
		t.Fatalf("dependencies[0] = %+v, want main module %q", manifest.Dependencies, manifest.Module)
	}
	if manifest.Tools["go"] == "" {
		t.Fatal("tools.go is empty")
	}
	if manifest.StandardImpact.ReportPath != standardImpactReportPath {
		t.Fatalf("standard_impact.report_path = %q, want %q", manifest.StandardImpact.ReportPath, standardImpactReportPath)
	}
	if manifest.StandardImpact.Status != "missing" && manifest.StandardImpact.Status != "present" {
		t.Fatalf("standard_impact.status = %q, want missing or present", manifest.StandardImpact.Status)
	}
	if manifest.DownstreamSyncRequired != manifest.StandardImpact.DownstreamSyncRequired {
		t.Fatalf("downstream_sync_required = %t, want standard impact value %t", manifest.DownstreamSyncRequired, manifest.StandardImpact.DownstreamSyncRequired)
	}
	if manifest.GeneratorEvidence.Command != "GOWORK=off make integration" || !manifest.GeneratorEvidence.Required {
		t.Fatalf("generator_evidence = %+v, want integration command and required=true", manifest.GeneratorEvidence)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "kernel", "github.com/ZoneCNH/kernel", "kernel") {
		t.Fatalf("generator targets = %+v, want kernel target", manifest.GeneratorEvidence.Targets)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "corekit", "example.com/acme/corekit", "corekit") {
		t.Fatalf("generator targets = %+v, want corekit target", manifest.GeneratorEvidence.Targets)
	}
	for _, artifact := range requiredArtifacts {
		if !contains(manifest.Artifacts, artifact) {
			t.Fatalf("artifacts = %v, want %s", manifest.Artifacts, artifact)
		}
	}
	for _, name := range checkNames {
		if manifest.Checks[name] != "passed" {
			t.Fatalf("checks[%q] = %q, want passed", name, manifest.Checks[name])
		}
	}
	if manifest.TreeState != "clean" && manifest.TreeState != "dirty" {
		t.Fatalf("tree_state = %q, want clean or dirty", manifest.TreeState)
	}
}

func TestVerifyManifestAcceptsFreshManifestAndRejectsDrift(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	goodPath := filepath.Join(t.TempDir(), "latest.json")
	if err := writeManifest(goodPath, manifest); err != nil {
		t.Fatal(err)
	}
	if err := verifyManifest(goodPath, true, false, "", 0); err != nil {
		t.Fatalf("verify fresh manifest: %v", err)
	}

	manifest.SourceDigest = "sha256:bad"
	manifest.Checks["lint"] = "unknown"
	manifest.Artifacts = []string{"release/manifest/latest.json"}
	manifest.StandardImpact.Status = "stale"
	manifest.DownstreamSyncRequired = !manifest.StandardImpact.DownstreamSyncRequired
	manifest.GeneratorEvidence.Command = "make old-integration"
	badPath := filepath.Join(t.TempDir(), "stale.json")
	if err := writeManifest(badPath, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(badPath, true, false, "", 0)
	if err == nil {
		t.Fatal("verify stale manifest succeeded, want error")
	}
	message := err.Error()
	for _, want := range []string{
		"source_digest does not match current tracked file contents",
		`checks.lint must be passed, got "unknown"`,
		"artifacts must include release/manifest/latest.json.sha256",
		"standard_impact does not match current standard impact evidence",
		"downstream_sync_required must match standard_impact.downstream_sync_required",
		"generator_evidence does not match current integration evidence",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error = %q, want substring %q", message, want)
		}
	}
}

func TestBuildStandardImpactEvidenceAllowsMissingReport(t *testing.T) {
	chdir(t, t.TempDir())

	got, err := buildStandardImpactEvidence()
	if err != nil {
		t.Fatal(err)
	}

	if got.ReportPath != standardImpactReportPath {
		t.Fatalf("report_path = %q, want %q", got.ReportPath, standardImpactReportPath)
	}
	if got.Status != "missing" {
		t.Fatalf("status = %q, want missing", got.Status)
	}
	if got.ReportSHA256 != "" {
		t.Fatalf("report_sha256 = %q, want empty", got.ReportSHA256)
	}
	if got.DownstreamSyncRequired {
		t.Fatal("downstream_sync_required = true, want false")
	}
}

func TestBuildStandardImpactEvidenceReadsReport(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, filepath.FromSlash(standardImpactReportPath))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "# Standard Impact\n\n- downstream_sync_required: `true`\n- primary_downstream: `github.com/ZoneCNH/kernel`\n"
	if err := os.WriteFile(reportPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)

	got, err := buildStandardImpactEvidence()
	if err != nil {
		t.Fatal(err)
	}

	if got.Status != "present" {
		t.Fatalf("status = %q, want present", got.Status)
	}
	if !strings.HasPrefix(got.ReportSHA256, "sha256:") {
		t.Fatalf("report_sha256 = %q, want sha256 prefix", got.ReportSHA256)
	}
	if !got.DownstreamSyncRequired {
		t.Fatal("downstream_sync_required = false, want true")
	}
	if got.PrimaryDownstream != "github.com/ZoneCNH/kernel" {
		t.Fatalf("primary_downstream = %q, want github.com/ZoneCNH/kernel", got.PrimaryDownstream)
	}
}

func TestBuildGeneratorEvidenceRecordsRepresentativeDownstreams(t *testing.T) {
	got := buildGeneratorEvidence()

	if got.Command != "GOWORK=off make integration" {
		t.Fatalf("command = %q, want integration command", got.Command)
	}
	if !got.Required {
		t.Fatal("required = false, want true")
	}
	if !hasGeneratorTarget(got.Targets, "kernel", "github.com/ZoneCNH/kernel", "kernel") {
		t.Fatalf("targets = %+v, want kernel target", got.Targets)
	}
	if !hasGeneratorTarget(got.Targets, "corekit", "example.com/acme/corekit", "corekit") {
		t.Fatalf("targets = %+v, want corekit target", got.Targets)
	}
}

func TestVerifyManifestRequiresCleanTree(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.TreeState = "dirty"

	path := filepath.Join(t.TempDir(), "dirty.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, true, true, "", 0)
	if err == nil {
		t.Fatal("verify dirty manifest with requireClean succeeded, want error")
	}
	if !strings.Contains(err.Error(), `tree_state must be clean, got "dirty"`) {
		t.Fatalf("error = %q, want require-clean failure", err)
	}
}

func TestSourceDigestUsesTrackedFileNamesAndContents(t *testing.T) {
	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")

	files := map[string]string{
		"a.txt":          "alpha\n",
		"nested/b.txt":   "bravo\n",
		"nested/cfg.yml": "name: charlie\n",
	}
	for path, content := range files {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestCommand(t, repo, "git", "add", ".")
	chdir(t, repo)

	gotDigest, gotCount, err := sourceDigest()
	if err != nil {
		t.Fatal(err)
	}

	if gotCount != len(files) {
		t.Fatalf("tracked file count = %d, want %d", gotCount, len(files))
	}
	if want := expectedSourceDigest(files); gotDigest != want {
		t.Fatalf("source digest = %q, want %q", gotDigest, want)
	}
}

func TestModuleDigestsIncludesReplaceMetadata(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(`module example.com/root

go 1.23

require example.com/dep v0.0.0

replace example.com/dep => ./dep
`), 0o644); err != nil {
		t.Fatal(err)
	}
	depDir := filepath.Join(root, "dep")
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(depDir, "go.mod"), []byte("module example.com/dep\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOWORK", "off")
	chdir(t, root)

	modules, err := moduleDigests()
	if err != nil {
		t.Fatal(err)
	}

	var foundMain bool
	var foundReplace bool
	for _, module := range modules {
		if module.Path == "example.com/root" && module.Main {
			foundMain = true
		}
		if module.Path == "example.com/dep" && module.Replace != nil && module.Replace.Path == "./dep" {
			foundReplace = true
		}
	}
	if !foundMain {
		t.Fatalf("modules = %+v, want main module", modules)
	}
	if !foundReplace {
		t.Fatalf("modules = %+v, want replace metadata for example.com/dep", modules)
	}
}

func TestWriteManifestCreatesParentAndWritesIndentedJSON(t *testing.T) {
	manifest := Manifest{
		Module:           "example.com/lib",
		Version:          "v1.2.3",
		Commit:           "abc123",
		TreeSHA:          "tree123",
		SourceDigest:     "sha256:source",
		TrackedFileCount: 1,
		GoVersion:        "go1.23.0",
		GeneratedAt:      "2026-01-02T03:04:05Z",
		GeneratedBy:      "test",
		TreeState:        "clean",
		Checks:           map[string]string{"fmt": "passed"},
		Tools:            map[string]string{"go": "go version"},
		Artifacts:        append([]string(nil), requiredArtifacts...),
	}
	path := filepath.Join(t.TempDir(), "release", "manifest", "latest.json")

	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("manifest JSON is invalid: %s", data)
	}
	if !strings.Contains(string(data), "\n  ") {
		t.Fatalf("manifest JSON is not indented: %s", data)
	}

	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Module != manifest.Module || got.Version != manifest.Version {
		t.Fatalf("round-trip manifest = %+v, want %+v", got, manifest)
	}
}

func TestToolVersionReportsMissingBinary(t *testing.T) {
	got := toolVersion("definitely-missing-releasemanifest-test-binary")
	if got != "missing" {
		t.Fatalf("toolVersion missing binary = %q, want missing", got)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func runTestCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
}

func releaseManifestFixtureRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/releasefixture\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, path := range contractFiles {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		content := "{}\n"
		if strings.HasSuffix(path, ".md") {
			content = "# Fixture Contract\n"
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFixtureOMCState(t, repo)
	runTestCommand(t, repo, "git", "add", ".")
	return repo
}

func writeFixtureOMCState(t *testing.T, repo string) {
	t.Helper()

	stateDir := filepath.Join(repo, ".omc", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(stateDir, "agent-replay-fixture.jsonl")
	content := []byte(`{"event":"fixture","source":"releasemanifest-test"}` + "\n")
	if err := os.WriteFile(statePath, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func expectedSourceDigest(files map[string]string) string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	digest := sha256.New()
	for _, path := range paths {
		sum := sha256.Sum256([]byte(files[path]))
		digest.Write([]byte(path))
		digest.Write([]byte{0})
		digest.Write([]byte(hex.EncodeToString(sum[:])))
		digest.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil))
}

func hasGeneratorTarget(targets []GeneratorTarget, name string, modulePath string, packageName string) bool {
	for _, target := range targets {
		if target.Name == name && target.ModulePath == modulePath && target.PackageName == packageName {
			return true
		}
	}
	return false
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
