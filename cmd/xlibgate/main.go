package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ZoneCNH/xlib-standard/internal/releasequality"
)

func main() {
	exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

var exit = os.Exit

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		write(stderr, usage)
		return 2
	}
	switch args[0] {
	case "version":
		return runVersion(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "main-guard":
		return runMainGuard(args[1:], stdout, stderr)
	case "worktree-guard":
		return runWorktreeGuard(args[1:], stdout, stderr)
	case "evidence-check":
		return runEvidenceCheck(args[1:], stdout, stderr)
	case "cli-contract":
		return runCLIContract(args[1:], stdout, stderr)
	case "issue-registry":
		return runIssueRegistry(args[1:], stdout, stderr)
	case "command-registry":
		return runCommandRegistry(args[1:], stdout, stderr)
	case "makefile-baseline":
		return runMakefileBaseline(args[1:], stdout, stderr)
	case "context-profile":
		return runContextProfile(args[1:], stdout, stderr)
	case "context-profile-check":
		return runContextProfileCheck(args[1:], stdout, stderr)
	case "context-schema-check":
		return runContextSchemaCheck(args[1:], stdout, stderr)
	case "minimal-kernel", "done-assertion", "agent-team-contract", "scope-lock", "pr-template", "acceptance-matrix", "runtime-health", "goal-runtime", "naming", "upgrade-standard", "conformance-profile", "downstream-registry", "self-healing-skeleton", "policy-schema", "github-settings", "toolchain", "evidence-artifacts", "install-runtime", "upgrade-runtime", "release-ready", "evidence-replay", "attest-conformance", "pack-standard", "pack-gate", "pack-evidence", "runtime-file-ownership", "downstream-baseline", "downstream-adoption", "autoresearch", "changelog", "github-governance", "governance-fixture-test", "supply-chain", "execution-context":
		return runPlannedCommand(args[0], args[1:], stdout, stderr)
	case "boundary":
		return runExternal(stdin, stdout, stderr, "./scripts/check_boundary.sh")
	case "contracts":
		return runExternal(stdin, stdout, stderr, "./scripts/check_contracts.sh")
	case "dependency-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_dependency_diff.sh")
	case "docs-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_docs.sh")
	case "evidence", "manifest":
		return runExternal(stdin, stdout, stderr, "go", "run", "./internal/tools/releasemanifest", "--out", "release/manifest/latest.json")
	case "integration":
		return runExternal(stdin, stdout, stderr, "./scripts/run_integration.sh")
	case "release-evidence-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_release_evidence.sh")
	case "release-evidence-checksum-check":
		return runExternal(stdin, stdout, stderr, "./scripts/hash_release_evidence.sh", "--check")
	case "release-evidence-hash":
		return runExternal(stdin, stdout, stderr, "./scripts/hash_release_evidence.sh")
	case "release-final-check":
		return runExternal(stdin, stdout, stderr, "make", "release-final-check")
	case "context-fast-check":
		return runExternal(stdin, stdout, stderr, "make", "context-fast-check")
	case "context-standard-check":
		return runExternal(stdin, stdout, stderr, "make", "context-standard-check")
	case "context-full-check":
		return runExternal(stdin, stdout, stderr, "make", "context-full-check")
	case "render-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_rendered_template.sh", args[1:]...)
	case "score":
		return runScore(args[1:], stdout, stderr)
	case "secrets", "security":
		return runExternal(stdin, stdout, stderr, "./scripts/check_secrets.sh")
	case "standard-impact-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_standard_impact.sh")
	case "help", "-h", "--help":
		write(stdout, usage)
		return 0
	default:
		write(stderr, "unknown command %q\n", args[0])
		return 2
	}
}

func runScore(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("xlibgate score", flag.ContinueOnError)
	flags.SetOutput(stderr)
	minimum := flags.Float64("min", releasequality.DefaultMinimum, "minimum acceptable release score")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	report := computeReleaseQuality(*minimum)
	data, err := marshalReleaseQuality(report)
	if err != nil {
		write(stderr, "ERROR: %v\n", err)
		return 1
	}
	write(stdout, "%s\n", data)
	if err := verifyReleaseQuality(report, *minimum); err != nil {
		write(stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}

var (
	computeReleaseQuality = releasequality.Compute
	marshalReleaseQuality = releasequality.Marshal
	verifyReleaseQuality  = releasequality.Verify
)

func runExternal(stdin io.Reader, stdout io.Writer, stderr io.Writer, name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		write(stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}

func write(writer io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(writer, format, args...)
}

const usage = `usage: xlibgate <command> [args]

commands:
  agent-team-contract [--dry-run]
  acceptance-matrix
  attest-conformance [--profile <name>]
  autoresearch
  boundary
  changelog
  cli-contract [--json|--output <path>|--explain]
  command-registry
  conformance-profile [--profile <name>]
  context-fast-check
  context-full-check
  context-profile [--json]
  context-profile-check [--json]
  context-schema-check [--json]
  context-standard-check
  contracts
  dependency-check
  doctor [--json]
  docs-check
  downstream-adoption
  downstream-baseline
  downstream-registry
  evidence
  evidence-artifacts
  evidence-check
  evidence-replay
  execution-context
  github-governance
  github-settings [--verify]
  goal-runtime
  governance-fixture-test
  install-runtime [--dry-run]
  integration
  issue-registry
  main-guard [--context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify]
  makefile-baseline
  manifest
  minimal-kernel
  done-assertion
  naming
  pack-evidence
  pack-gate
  pack-standard
  policy-schema
  pr-template
  release-evidence-check
  release-evidence-checksum-check
  release-evidence-hash
  release-final-check
  release-ready
  render-check <rendered-dir>
  runtime-file-ownership
  runtime-health
  scope-lock
  score [--min <score>]
  secrets
  security
  self-healing-skeleton
  standard-impact-check
  supply-chain
  toolchain
  upgrade-runtime [--dry-run]
  upgrade-standard [--dry-run]
  version [--json]
  worktree-guard [--context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify]
`
