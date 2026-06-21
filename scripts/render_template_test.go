package scripts_test

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTemplateExcludesGeneratedReleaseArtifacts(t *testing.T) {
	contents, err := os.ReadFile("render_template.sh")
	if err != nil {
		t.Fatalf("read render_template.sh: %v", err)
	}

	script := string(contents)
	for _, required := range []string{
		"--exclude='./.codex'",
		"--exclude='./.omc'",
		"--exclude='./.omx'",
		"--exclude='./.worktree'",
		"--exclude='./release/manifest/latest.json'",
		"--exclude='./release/manifest/latest.json.sha256'",
		"--exclude='./release/standard-impact/latest.md'",
		"--exclude='./release/downstream-sync/latest.md'",
		"--exclude='./release/debt/latest.json'",
		"--exclude='./release/debt/latest.md'",
		"--exclude='./release/debt/latest.json.sha256'",
		`rm -rf "$out_dir/.codex"`,
		`rm -rf "$out_dir/.omc"`,
		`rm -rf "$out_dir/.omx"`,
		`rm -rf "$out_dir/.worktree"`,
		`rm -f "$out_dir/release/manifest/latest.json"`,
		`rm -f "$out_dir/release/manifest/latest.json.sha256"`,
		`rm -f "$out_dir/release/standard-impact/latest.md"`,
		`rm -f "$out_dir/release/downstream-sync/latest.md"`,
		`rm -f "$out_dir/release/debt/latest.json"`,
		`rm -f "$out_dir/release/debt/latest.md"`,
		`rm -f "$out_dir/release/debt/latest.json.sha256"`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("render_template.sh missing render omission rule %q", required)
		}
	}
}

func TestRenderTemplateIncludesGoalcliControlPlane(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "configx")
	cmd := exec.Command(
		"bash",
		"render_template.sh",
		"--module-name",
		"configx",
		"--module-path",
		"github.com/ZoneCNH/configx",
		"--package-name",
		"configx",
		"--out",
		outDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render template: %v\n%s", err, output)
	}

	for _, required := range []string{
		filepath.Join("cmd", "goalcli", "main.go"),
		filepath.Join("cmd", "goalcli", "main_test.go"),
		filepath.Join("cmd", "goalcli", "governance.go"),
		filepath.Join("internal", "goalcli", "README.md"),
		"Makefile",
		filepath.Join(".agent", "index.yaml"),
		filepath.Join(".agent", "harness", "harness.yaml"),
		filepath.Join(".agent", "harness", "gates.md"),
		filepath.Join(".agent", "registries", "command-registry.yaml"),
		filepath.Join(".agent", "registries", "command-implementation-status.yaml"),
		filepath.Join(".agent", "registries", "makefile-baseline.yaml"),
		filepath.Join(".agent", "registries", "makefile-target-registry.yaml"),
		filepath.Join("contracts", "goalcli-report.schema.json"),
		filepath.Join("docs", "standard", "goalcli-cli-contract.md"),
		filepath.Join("docs", "standard", "goalcli-runtime.md"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, required)); err != nil {
			t.Fatalf("rendered template missing goalcli control-plane path %s: %v", required, err)
		}
	}

	makefile, err := os.ReadFile(filepath.Join(outDir, "Makefile"))
	if err != nil {
		t.Fatalf("read rendered Makefile: %v", err)
	}
	if !strings.Contains(string(makefile), "GOALCLI ?= go run ./cmd/goalcli") {
		t.Fatalf("rendered Makefile missing GOALCLI entrypoint")
	}
}

func TestRenderTemplateRewritesSourceModuleIdentity(t *testing.T) {
	sourceModulePath := currentModulePath(t)
	sourceModuleName := path.Base(sourceModulePath)
	sourcePackageName := currentPackageName(t, sourceModuleName)
	targetModuleName := "rendered-identity"
	targetModulePath := "github.com/ZoneCNH/rendered-identity"
	targetPackageName := "rendered_identity"
	outDir := filepath.Join(t.TempDir(), targetPackageName)

	cmd := exec.Command(
		"bash",
		"render_template.sh",
		"--module-name",
		targetModuleName,
		"--module-path",
		targetModulePath,
		"--package-name",
		targetPackageName,
		"--out",
		outDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render template: %v\n%s", err, output)
	}

	moduleFile, err := os.ReadFile(filepath.Join(outDir, "go.mod"))
	if err != nil {
		t.Fatalf("read rendered go.mod: %v", err)
	}
	if strings.Contains(string(moduleFile), sourceModulePath) {
		t.Fatalf("rendered go.mod still references source module path")
	}
	if !strings.Contains(string(moduleFile), "module "+targetModulePath) {
		t.Fatalf("rendered go.mod missing target module path")
	}
	if strings.Contains(string(moduleFile), "github.com/taosdata/driver-go") {
		t.Fatalf("rendered go.mod still references source TDengine dependency")
	}

	sumFile, err := os.ReadFile(filepath.Join(outDir, "go.sum"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read rendered go.sum: %v", err)
	}
	if strings.Contains(string(sumFile), "github.com/taosdata/driver-go") {
		t.Fatalf("rendered go.sum still references source TDengine dependency")
	}

	if sourcePackageName != targetPackageName {
		if _, err := os.Stat(filepath.Join(outDir, "pkg", sourcePackageName)); !os.IsNotExist(err) {
			t.Fatalf("rendered output should omit source package directory, stat err=%v", err)
		}
	}

	makefile, err := os.ReadFile(filepath.Join(outDir, "Makefile"))
	if err != nil {
		t.Fatalf("read rendered Makefile: %v", err)
	}
	sourceCoverageTarget := sourcePackageName + "-coverage-check"
	targetCoverageTarget := targetPackageName + "-coverage-check"
	if sourceCoverageTarget != targetCoverageTarget && strings.Contains(string(makefile), sourceCoverageTarget) {
		t.Fatalf("rendered Makefile still references source coverage target")
	}
	if !strings.Contains(string(makefile), targetCoverageTarget) {
		t.Fatalf("rendered Makefile missing target coverage target")
	}

	nestedDir := filepath.Join(t.TempDir(), "nested_identity")
	nestedCmd := exec.Command(
		"bash",
		"scripts/render_template.sh",
		"--module-name",
		"nested-identity",
		"--module-path",
		"github.com/ZoneCNH/nested-identity",
		"--package-name",
		"nested_identity",
		"--out",
		nestedDir,
	)
	nestedCmd.Dir = outDir

	nestedOutput, err := nestedCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("nested render template: %v\n%s", err, nestedOutput)
	}

	nestedClient, err := os.ReadFile(filepath.Join(nestedDir, "pkg", "nested_identity", "client.go"))
	if err != nil {
		t.Fatalf("read nested client: %v", err)
	}
	if !strings.HasPrefix(string(nestedClient), "package nested_identity\n") {
		t.Fatalf("nested rendered client has invalid package declaration:\n%s", nestedClient)
	}
}

func currentModulePath(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list module path: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func currentPackageName(t *testing.T, moduleName string) string {
	t.Helper()

	goNameTemplate := "{" + "{.Name}" + "}"
	cmd := exec.Command("go", "list", "-f", goNameTemplate, "./pkg/"+moduleName)
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.Output()
	if err != nil {
		return strings.ReplaceAll(moduleName, "-", "_")
	}
	return strings.TrimSpace(string(output))
}

func TestRenderTemplateIncludesDockerContract(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "kernel")
	cmd := exec.Command(
		"bash",
		"render_template.sh",
		"--module-name",
		"kernel",
		"--module-path",
		"github.com/ZoneCNH/kernel",
		"--package-name",
		"kernel",
		"--out",
		outDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render template: %v\n%s", err, output)
	}

	for _, required := range []string{
		"Dockerfile",
		"docker-compose.yml",
		".dockerignore",
		filepath.Join(".devcontainer", "devcontainer.json"),
		filepath.Join("scripts", "docker", "check_toolchain.sh"),
		filepath.Join("scripts", "docker", "docker_gate.sh"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, required)); err != nil {
			t.Fatalf("rendered template missing Docker contract path %s: %v", required, err)
		}
	}

	dockerfile, err := os.ReadFile(filepath.Join(outDir, "Dockerfile"))
	if err != nil {
		t.Fatalf("read rendered Dockerfile: %v", err)
	}
	for _, required := range []string{
		"GOLANGCI_LINT_VERSION",
		"GOVULNCHECK_VERSION",
		"python3-yaml",
		"github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
		"golang.org/x/vuln/cmd/govulncheck",
		"safe.directory /workspace",
	} {
		if !strings.Contains(string(dockerfile), required) {
			t.Fatalf("rendered Dockerfile missing toolchain marker %s", required)
		}
	}

	makefile, err := os.ReadFile(filepath.Join(outDir, "Makefile"))
	if err != nil {
		t.Fatalf("read rendered Makefile: %v", err)
	}
	if !strings.Contains(string(makefile), `GITHUB_ACTIONS=$${GITHUB_ACTIONS:-}`) {
		t.Fatalf("rendered Makefile missing Docker CI environment passthrough")
	}
	if !strings.Contains(string(makefile), `GOLANGCI_LINT_VERSION=$${GOLANGCI_LINT_VERSION:-v2.1.6}`) {
		t.Fatalf("rendered Makefile missing Docker lint toolchain build arg")
	}
	if !strings.Contains(string(makefile), `GIT_CONFIG_VALUE_0=/workspace`) {
		t.Fatalf("rendered Makefile missing Docker Git workspace trust config")
	}
	for _, target := range []string{
		"docker-toolchain-check",
		"docker-build",
		"docker-build-check",
		"docker-shell",
		"docker-ci",
		"docker-release-check",
		"docker-release-final-check",
		"docker-goalcli",
		"docker-goalcli-image",
		"docker-goalcli-version",
		"docker-runtime-check",
		"docker-drift-check",
		"docker-contract",
	} {
		if !strings.Contains(string(makefile), target) {
			t.Fatalf("rendered Makefile missing Docker contract target %s", target)
		}
	}

	dockerGate, err := os.ReadFile(filepath.Join(outDir, "scripts", "docker", "docker_gate.sh"))
	if err != nil {
		t.Fatalf("read rendered Docker gate: %v", err)
	}
	if !strings.Contains(string(dockerGate), `GITHUB_ACTIONS=${GITHUB_ACTIONS:-}`) {
		t.Fatalf("rendered Docker gate missing GitHub Actions environment passthrough")
	}
	if !strings.Contains(string(dockerGate), `GOLANGCI_LINT_VERSION:-v2.1.6`) {
		t.Fatalf("rendered Docker gate missing lint toolchain build arg")
	}
	if !strings.Contains(string(dockerGate), `GOVULNCHECK_VERSION:-v1.1.4`) {
		t.Fatalf("rendered Docker gate missing govulncheck toolchain build arg")
	}
	if !strings.Contains(string(dockerGate), `GIT_CONFIG_VALUE_0=/workspace`) {
		t.Fatalf("rendered Docker gate missing Git workspace trust config")
	}
}

func TestRenderTemplatePrunesOmittedAgentInboxIndexEntries(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "kernel")
	cmd := exec.Command(
		"bash",
		"render_template.sh",
		"--module-name",
		"kernel",
		"--module-path",
		"github.com/ZoneCNH/kernel",
		"--package-name",
		"kernel",
		"--out",
		outDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render template: %v\n%s", err, output)
	}

	index, err := os.ReadFile(filepath.Join(outDir, ".agent", "index.yaml"))
	if err != nil {
		t.Fatalf("read rendered agent index: %v", err)
	}
	if strings.Contains(string(index), ".agent/inbox/") {
		t.Fatalf("rendered agent index still references omitted inbox entries")
	}
	if _, err := os.Stat(filepath.Join(outDir, ".agent", "inbox")); !os.IsNotExist(err) {
		t.Fatalf("rendered agent inbox should be omitted, stat err=%v", err)
	}
}

func TestRenderTemplateGitArchivePrunesRuntimeState(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "redisx")
	cmd := exec.Command(
		"bash",
		"render_template.sh",
		"--module-name",
		"redisx",
		"--module-path",
		"github.com/ZoneCNH/redisx",
		"--package-name",
		"redisx",
		"--out",
		outDir,
	)
	cmd.Env = append(os.Environ(), "XLIB_RENDER_FORCE_GIT_ARCHIVE=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render template from git archive: %v\n%s", err, output)
	}

	for _, omitted := range []string{
		".omc",
		".omx",
		".worktree",
		filepath.Join("release", "manifest", "latest.json"),
		filepath.Join("release", "manifest", "latest.json.sha256"),
		filepath.Join("release", "standard-impact", "latest.md"),
		filepath.Join("release", "downstream-sync", "latest.md"),
		filepath.Join("release", "debt", "latest.json"),
		filepath.Join("release", "debt", "latest.md"),
		filepath.Join("release", "debt", "latest.json.sha256"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, omitted)); !os.IsNotExist(err) {
			t.Fatalf("rendered git archive should omit %s, stat err=%v", omitted, err)
		}
	}
}

func TestRenderTemplateGitArchiveSkipsUntrackedFiles(t *testing.T) {
	markerPath := filepath.Join("..", ".xlib-render-untracked-marker-test")
	if err := os.WriteFile(markerPath, []byte("untracked marker"), 0o600); err != nil {
		t.Fatalf("write untracked marker: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(markerPath)
	})

	outDir := filepath.Join(t.TempDir(), "kernel")
	cmd := exec.Command(
		"bash",
		"render_template.sh",
		"--module-name",
		"kernel",
		"--module-path",
		"github.com/ZoneCNH/kernel",
		"--package-name",
		"kernel",
		"--out",
		outDir,
	)
	cmd.Env = append(os.Environ(), "XLIB_RENDER_FORCE_GIT_ARCHIVE=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render template: %v\n%s", err, output)
	}

	if _, err := os.Stat(filepath.Join(outDir, ".xlib-render-untracked-marker-test")); !os.IsNotExist(err) {
		t.Fatalf("expected git archive render to skip untracked marker, stat err=%v", err)
	}
}
