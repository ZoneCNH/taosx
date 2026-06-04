package scripts_test

import (
	"os"
	"os/exec"
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

	makefile, err := os.ReadFile(filepath.Join(outDir, "Makefile"))
	if err != nil {
		t.Fatalf("read rendered Makefile: %v", err)
	}
	for _, target := range []string{"docker-toolchain-check", "docker-ci", "docker-release-check"} {
		if !strings.Contains(string(makefile), target) {
			t.Fatalf("rendered Makefile missing Docker contract target %s", target)
		}
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
