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
	case "boundary":
		return runExternal(stdin, stdout, stderr, "./scripts/check_boundary.sh")
	case "contracts":
		return runExternal(stdin, stdout, stderr, "./scripts/check_contracts.sh")
	case "dependency-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_dependency_diff.sh")
	case "docs-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_docs.sh")
	case "evidence":
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
	case "render-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_rendered_template.sh", args[1:]...)
	case "score":
		return runScore(args[1:], stdout, stderr)
	case "secrets":
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
  boundary
  contracts
  dependency-check
  docs-check
  evidence
  integration
  release-evidence-check
  release-evidence-checksum-check
  release-evidence-hash
  release-final-check
  render-check <rendered-dir>
  score [--min <score>]
  secrets
  standard-impact-check
`
