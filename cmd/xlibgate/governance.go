package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const xlibgateVersion = "v2.9.3"

type gateReport struct {
	Command string   `json:"command"`
	Status  string   `json:"status"`
	Details []string `json:"details,omitempty"`
	Gaps    []string `json:"gaps,omitempty"`
}

func emitReport(stdout io.Writer, command, status string, details []string, gaps []string) int {
	report := gateReport{Command: command, Status: status, Details: details, Gaps: gaps}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		write(stdout, "{\"command\":%q,\"status\":%q}\n", command, status)
	} else {
		write(stdout, "%s\n", data)
	}
	if status == "passed" {
		return 0
	}
	return 1
}

func runVersion(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("version", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("version", err, stderr)
	}
	return emitReport(stdout, "version", "passed", []string{"xlib-standard goal " + xlibgateVersion, "xlibgate governance CLI available"}, nil)
}

func runDoctor(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("doctor", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("doctor", err, stderr)
	}
	required := []string{
		".agent/harness.yaml",
		".agent/issue-registry.yaml",
		".agent/command-registry.yaml",
		".agent/makefile-target-registry.yaml",
		".agent/makefile-baseline.yaml",
		"docs/standard/xlibgate-cli-contract.md",
		"contracts/xlibgate-report.schema.json",
		"Makefile",
	}
	if isXlibStandardSourceModule() {
		required = append([]string{"docs/goal.md"}, required...)
	}
	var gaps []string
	for _, path := range required {
		if !fileExists(path) {
			gaps = append(gaps, "missing "+path)
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: doctor found %d gap(s)\n", len(gaps))
		return emitReport(stdout, "doctor", "failed", nil, gaps)
	}
	return emitReport(stdout, "doctor", "passed", []string{"required governance files are present"}, nil)
}

func runMainGuard(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("xlibgate main-guard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	context := flags.String("context", envDefault("XLIB_CONTEXT", "local_write"), "execution context")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if !validContext(*context) {
		write(stderr, "ERROR: invalid context %q\n", *context)
		return 2
	}
	branch := gitOutput("rev-parse", "--abbrev-ref", "HEAD")
	if *context == "local_write" && (branch == "main" || branch == "master") {
		return emitReport(stdout, "main-guard", "failed", nil, []string{"local_write is forbidden on " + branch})
	}
	return emitReport(stdout, "main-guard", "passed", []string{"context=" + *context, "branch=" + fallback(branch, "unknown")}, nil)
}

func runWorktreeGuard(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("xlibgate worktree-guard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	context := flags.String("context", envDefault("XLIB_CONTEXT", "local_write"), "execution context")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if !validContext(*context) {
		write(stderr, "ERROR: invalid context %q\n", *context)
		return 2
	}
	top := gitOutput("rev-parse", "--show-toplevel")
	common := gitOutput("rev-parse", "--path-format=absolute", "--git-common-dir")
	isWorkerTree := strings.Contains(top, string(filepath.Separator)+".worktree"+string(filepath.Separator)) || strings.Contains(top, string(filepath.Separator)+".worktrees"+string(filepath.Separator)) || strings.Contains(common, string(filepath.Separator)+"worktrees"+string(filepath.Separator))
	if *context == "local_write" && !isWorkerTree {
		return emitReport(stdout, "worktree-guard", "failed", []string{"top=" + fallback(top, "unknown")}, []string{"local_write requires a worker worktree"})
	}
	return emitReport(stdout, "worktree-guard", "passed", []string{"context=" + *context, "top=" + fallback(top, "unknown")}, nil)
}

func runRegistryCheck(command string, required map[string][]string, stdout io.Writer, stderr io.Writer) int {
	var gaps []string
	for path, needles := range required {
		content, err := os.ReadFile(path)
		if err != nil {
			gaps = append(gaps, "missing "+path)
			continue
		}
		text := string(content)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				gaps = append(gaps, path+" missing "+needle)
			}
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: %s found %d gap(s)\n", command, len(gaps))
		return emitReport(stdout, command, "failed", nil, gaps)
	}
	return emitReport(stdout, command, "passed", []string{"registry contract satisfied"}, nil)
}

func runEvidenceCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("evidence-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("evidence-check", err, stderr)
	}
	return runRegistryCheck("evidence-check", map[string][]string{
		".agent/done-assertion.yaml":           {"DONE with evidence", "commit", "gates"},
		".agent/evidence-artifact-policy.yaml": {"redaction", "sha256", "release/manifest/latest.json"},
		".agent/harness.yaml":                  {"manifest", "checksum", "required_fields"},
	}, stdout, stderr)
}

func runCLIContract(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("cli-contract", args, internalCommandFlagSpec{boolFlags: []string{"json", "explain"}, stringFlags: []string{"output"}}); err != nil {
		return invalidInternalArgsExit("cli-contract", err, stderr)
	}
	return runRegistryCheck("cli-contract", map[string][]string{
		"docs/standard/xlibgate-cli-contract.md": commandRegistryRequiredCommands(),
		"contracts/xlibgate-report.schema.json":  {"command", "status", "details", "gaps"},
		".agent/command-registry.yaml":           requiredCommandRegistryNeedles(),
	}, stdout, stderr)
}

func runIssueRegistry(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("issue-registry", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("issue-registry", err, stderr)
	}
	return runRegistryCheck("issue-registry", map[string][]string{
		".agent/issue-registry.yaml": requiredIssueRegistryNeedles(),
	}, stdout, stderr)
}

func runCommandRegistry(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("command-registry", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("command-registry", err, stderr)
	}
	return runRegistryCheck("command-registry", map[string][]string{
		".agent/command-registry.yaml": requiredCommandRegistryNeedles(),
	}, stdout, stderr)
}

func runMakefileBaseline(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("makefile-baseline", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("makefile-baseline", err, stderr)
	}
	requiredTargets := []string{"fmt", "vet", "lint", "test", "race", "boundary", "security", "contracts", "docs-check", "evidence", "score-check", "main-guard", "worktree-guard", "evidence-check", "cli-contract", "issue-registry", "command-registry", "makefile-baseline", "context-profile", "context-profile-check", "context-schema-check", "context-lite", "context-standard", "context-full", "context-release", "context-fast-check", "context-standard-check", "context-full-check", "governance-check", "p1-governance-check", "execution-context", "p2-runtime-check", "release-check", "release-final-check"}
	required := map[string][]string{"Makefile": {}, ".agent/makefile-target-registry.yaml": requiredTargets, ".agent/makefile-baseline.yaml": requiredTargets}
	for _, target := range requiredTargets {
		required["Makefile"] = append(required["Makefile"], ".PHONY: "+target, target+":")
	}
	return runRegistryCheck("makefile-baseline", required, stdout, stderr)
}

var contextProfileGates = map[string][]string{
	"lite":     {"main-guard", "worktree-guard", "evidence-check", "cli-contract", "command-registry", "issue-registry", "makefile-baseline", "context-profile-check"},
	"standard": {"context-lite", "p1-governance-check", "docs-check"},
	"full":     {"context-standard", "p2-runtime-check"},
	"release":  {"context-standard", "standard-impact-check", "score-check", "evidence", "release-evidence-hash", "release-evidence-check", "release-evidence-checksum-check"},
}

func runContextProfile(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("xlibgate context-profile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profile := flags.String("profile", "standard", "context runtime profile")
	flags.Bool("json", false, "")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 {
		write(stderr, "ERROR: context-profile invalid arguments: unexpected positional argument %q\n", flags.Arg(0))
		return 2
	}
	normalized := normalizeContextProfile(*profile)
	gates, ok := contextProfileGates[normalized]
	if !ok {
		write(stderr, "ERROR: invalid context profile %q\n", *profile)
		return 2
	}
	return emitReport(stdout, "context-profile", "passed", contextProfileDetails(normalized, gates), nil)
}

func runContextProfileAlias(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs(command, args, internalCommandFlagSpec{boolFlags: []string{"json", "strict"}}); err != nil {
		return invalidInternalArgsExit(command, err, stderr)
	}
	profile := mapContextAliasToProfile(command)
	gates := contextProfileGates[profile]
	return emitReport(stdout, command, "passed", contextProfileDetails(profile, gates), nil)
}

func runContextProfileCheck(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs(command, args, internalCommandFlagSpec{boolFlags: []string{"json", "strict"}, stringFlags: []string{"profile"}}); err != nil {
		return invalidInternalArgsExit(command, err, stderr)
	}
	contextTargets := []string{"context-profile", "context-profile-check", "context-schema-check", "context-lite", "context-standard", "context-full", "context-release", "context-fast-check", "context-standard-check", "context-full-check"}
	required := map[string][]string{
		".agent/command-registry.yaml":        requiredCommandRegistryNeedles(),
		".agent/issue-registry.yaml":          requiredIssueRegistryNeedles(),
		".agent/makefile-target-registry.yaml": contextTargets,
		".agent/makefile-baseline.yaml":       contextTargets,
		"docs/standard/xlibgate-cli-contract.md": commandRegistryRequiredCommands(),
		"Makefile": {"release-final-check:", "$(MAKE) context-release"},
	}
	for _, target := range contextTargets {
		required["Makefile"] = append(required["Makefile"], ".PHONY: "+target, target+":")
	}
	var gaps []string
	for path, needles := range required {
		content, err := os.ReadFile(path)
		if err != nil {
			gaps = append(gaps, "missing "+path)
			continue
		}
		text := string(content)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				gaps = append(gaps, path+" missing "+needle)
			}
		}
	}
	if issueRegistry, err := os.ReadFile(".agent/issue-registry.yaml"); err == nil && strings.Contains(string(issueRegistry), "CTX-051") {
		gaps = append(gaps, ".agent/issue-registry.yaml must not claim CTX-051")
	}
	if makefile, err := os.ReadFile("Makefile"); err == nil {
		contextReleaseBlock := makefileTargetBlock(string(makefile), "context-release")
		if strings.Contains(contextReleaseBlock, "release-check") || strings.Contains(contextReleaseBlock, "release-final-check") {
			gaps = append(gaps, "context-release must not call release-check or release-final-check")
		}
		releaseFinalBlock := makefileTargetBlock(string(makefile), "release-final-check")
		if !strings.Contains(releaseFinalBlock, "$(MAKE) context-release") {
			gaps = append(gaps, "release-final-check must call context-release")
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: %s found %d gap(s)\n", command, len(gaps))
		return emitReport(stdout, command, "failed", nil, gaps)
	}
	return emitReport(stdout, command, "passed", []string{"context runtime v4.0 registry contract satisfied", ".agent/context not required or claimed", "context-release excludes release-check and release-final-check"}, nil)
}

func contextProfileDetails(profile string, gates []string) []string {
	return []string{
		"context_runtime=v4.0",
		"profile=" + profile,
		"gates=" + strings.Join(gates, ","),
		"legacy_aliases=context-fast-check,context-standard-check,context-full-check",
		"release_final_delegates=context-release",
	}
}

func normalizeContextProfile(profile string) string {
	switch profile {
	case "fast":
		return "lite"
	default:
		return profile
	}
}

func mapContextAliasToProfile(command string) string {
	switch command {
	case "context-lite", "context-fast-check":
		return "lite"
	case "context-full", "context-full-check":
		return "full"
	case "context-release":
		return "release"
	default:
		return "standard"
	}
}

func makefileTargetBlock(content, target string) string {
	lines := strings.Split(content, "\n")
	var block []string
	inBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, target+":") {
			inBlock = true
			block = append(block, line)
			continue
		}
		if inBlock {
			if line != "" && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") && strings.Contains(line, ":") {
				break
			}
			block = append(block, line)
		}
	}
	return strings.Join(block, "\n")
}

var plannedCommandFiles = map[string][]string{
	"minimal-kernel":          {".agent/minimal-kernel.yaml"},
	"done-assertion":          {".agent/done-assertion.yaml"},
	"agent-team-contract":     {".agent/team-contract.yaml"},
	"scope-lock":              {".agent/scope-locks.yaml"},
	"pr-template":             {".agent/pr-template-contract.yaml", ".github/pull_request_template.md"},
	"acceptance-matrix":       {".agent/acceptance-matrix.yaml"},
	"runtime-health":          {".agent/runtime-health.yaml"},
	"goal-runtime":            {".agent/goal-runtime.md", ".agent/harness.yaml"},
	"naming":                  {"docs/standard/repository-roles.md", "docs/standard/module-boundary.md"},
	"upgrade-standard":        {".agent/downstream-registry.yaml"},
	"conformance-profile":     {".agent/conformance-profiles.yaml"},
	"downstream-registry":     {".agent/downstream-registry.yaml"},
	"self-healing-skeleton":   {".agent/failure-taxonomy.yaml", ".agent/root-cause.yaml", ".agent/regression-memory.yaml"},
	"policy-schema":           {".agent/policy-schema.yaml"},
	"github-settings":         {".agent/github-settings.yaml"},
	"github-governance":       {".agent/github-governance.yaml"},
	"governance-fixture-test": {".agent/governance-fixture-test.yaml"},
	"toolchain":               {".agent/toolchain.yaml"},
	"evidence-artifacts":      {".agent/evidence-artifact-policy.yaml"},
	"install-runtime":         {".agent/runtime-install.yaml"},
	"upgrade-runtime":         {".agent/runtime-upgrade.yaml"},
	"release-ready":           {".agent/release-readiness-formula.yaml"},
	"evidence-replay":         {".agent/evidence-replay.yaml"},
	"attest-conformance":      {".agent/conformance-profiles.yaml"},
	"pack-standard":           {".agent/standard-pack.yaml"},
	"pack-gate":               {".agent/gate-pack.yaml"},
	"pack-evidence":           {".agent/evidence-pack.yaml"},
	"runtime-file-ownership":  {".agent/runtime-file-ownership.yaml"},
	"downstream-baseline":     {".agent/downstream-baseline-scan.yaml", ".agent/downstream-registry.yaml"},
	"downstream-adoption":     {".agent/downstream-adoption-modes.yaml", ".agent/downstream-registry.yaml"},
	"autoresearch":            {".agent/autoresearch.yaml"},
	"changelog":               {".agent/changelog.yaml"},
	"supply-chain":            {"docs/supply-chain.md"},
	"execution-context":       {".agent/execution-context.yaml", "contracts/execution-context.schema.json"},
}

func runPlannedCommand(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validatePlannedCommandArgs(command, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		write(stderr, "ERROR: %s invalid arguments: %v\n", command, err)
		return 2
	}

	var details []string
	var gaps []string
	files, ok := plannedCommandFiles[command]
	if !ok || len(files) == 0 {
		write(stderr, "ERROR: %s has no manifest coverage\n", command)
		return emitReport(stdout, command, "failed", []string{"args=" + strings.Join(args, " ")}, []string{"planned command has no manifest coverage: " + command})
	}
	for _, path := range files {
		if fileExists(path) {
			details = append(details, "found "+path)
		} else {
			gaps = append(gaps, "missing "+path)
		}
	}
	if command == "downstream-baseline" || command == "downstream-adoption" || command == "upgrade-standard" {
		mode := fallback(flagValue(args, "mode", ""), "patch-only")
		if flagProvided(args, "repo") {
			repo := flagValue(args, "repo", "")
			details = append(details, "repo="+repo, "mode="+mode, "dry_run=true")
			if !fileExists(repo) {
				gaps = append(gaps, "downstream repo unavailable in worker workspace: "+repo)
				return emitPlannedReport(stdout, stderr, command, "gap", details, gaps, args)
			}
		} else {
			details = append(details, "repo=manifest-only", "mode="+mode, "dry_run=true")
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: %s found %d gap(s)\n", command, len(gaps))
		return emitReport(stdout, command, "failed", details, gaps)
	}
	details = append(details, "args="+strings.Join(args, " "))
	if plannedCommandVerifyRequested(args) {
		details = append(details, "local dry-run verifier satisfied manifest coverage")
	}
	return emitReport(stdout, command, "passed", details, nil)
}

func flagProvided(args []string, name string) bool {
	for _, arg := range args {
		if arg == "--"+name || strings.HasPrefix(arg, "--"+name+"=") {
			return true
		}
	}
	return false
}

func validatePlannedCommandArgs(command string, args []string) error {
	flags := flag.NewFlagSet("xlibgate "+command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Bool("dry-run", false, "")
	flags.Bool("verify", false, "")
	flags.Bool("strict", false, "")
	flags.Bool("json", false, "")
	flags.String("repo", "", "")
	flags.String("mode", "", "")
	context := flags.String("context", "", "")
	flags.String("profile", "", "")
	flags.String("output", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected positional argument %q", flags.Arg(0))
	}
	if *context != "" && !validContext(*context) {
		return fmt.Errorf("invalid context %q", *context)
	}
	return nil
}

type internalCommandFlagSpec struct {
	boolFlags   []string
	stringFlags []string
}

func validateInternalCommandArgs(command string, args []string, spec internalCommandFlagSpec) error {
	flags := flag.NewFlagSet("xlibgate "+command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	for _, name := range spec.boolFlags {
		flags.Bool(name, false, "")
	}
	for _, name := range spec.stringFlags {
		flags.String(name, "", "")
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected positional argument %q", flags.Arg(0))
	}
	return nil
}

func invalidInternalArgsExit(command string, err error, stderr io.Writer) int {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	write(stderr, "ERROR: %s invalid arguments: %v\n", command, err)
	return 2
}

func isXlibStandardSourceModule() bool {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return true
	}
	sourceModule := strings.Join([]string{"github.com", "ZoneCNH", "xlib" + "-standard"}, "/")
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1] == sourceModule
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func envDefault(name, fallbackValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallbackValue
}

func fallback(value, fallbackValue string) string {
	if value == "" {
		return fallbackValue
	}
	return value
}

func validContext(value string) bool {
	for _, context := range validExecutionContexts {
		if value == context {
			return true
		}
	}
	return false
}

var validExecutionContexts = []string{"local_write", "local_readonly", "ci_pull_request", "ci_main_verify", "release_verify"}

var commandRegistryCommands = []string{
	"version",
	"doctor",
	"minimal-kernel",
	"main-guard",
	"worktree-guard",
	"evidence-check",
	"done-assertion",
	"cli-contract",
	"issue-registry",
	"command-registry",
	"makefile-baseline",
	"agent-team-contract",
	"scope-lock",
	"pr-template",
	"acceptance-matrix",
	"runtime-health",
	"goal-runtime",
	"naming",
	"upgrade-standard",
	"conformance-profile",
	"downstream-registry",
	"self-healing-skeleton",
	"policy-schema",
	"github-settings",
	"github-governance",
	"governance-fixture-test",
	"toolchain",
	"evidence-artifacts",
	"install-runtime",
	"upgrade-runtime",
	"release-ready",
	"evidence-replay",
	"attest-conformance",
	"pack-standard",
	"pack-gate",
	"pack-evidence",
	"runtime-file-ownership",
	"downstream-baseline",
	"downstream-adoption",
	"autoresearch",
	"changelog",
	"supply-chain",
	"execution-context",
	"boundary",
	"contracts",
	"dependency-check",
	"docs-check",
	"evidence",
	"manifest",
	"integration",
	"release-evidence-check",
	"release-evidence-checksum-check",
	"release-evidence-hash",
	"release-final-check",
	"render-check",
	"score",
	"secrets",
	"security",
	"standard-impact-check",
}

func commandRegistryRequiredCommands() []string {
	return append([]string(nil), commandRegistryCommands...)
}

func requiredCommandRegistryNeedles() []string {
	needles := make([]string, 0, len(commandRegistryCommands))
	for _, command := range commandRegistryCommands {
		needles = append(needles, "name: "+command)
	}
	return needles
}

func requiredIssueRegistryNeedles() []string {
	needles := []string{"status: implemented"}
	for _, prefixAndCount := range []struct {
		prefix string
		count  int
	}{
		{prefix: "P0", count: 16},
		{prefix: "P1", count: 21},
		{prefix: "P2", count: 15},
	} {
		for i := 1; i <= prefixAndCount.count; i++ {
			needles = append(needles, prefixAndCount.prefix+"-"+fmt.Sprintf("%03d", i))
		}
	}
	return needles
}

func gitOutput(args ...string) string {
	cmd := exec.Command("git", args...)
	data, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func flagValue(args []string, name string, fallbackValue string) string {
	for i, arg := range args {
		if arg == "--"+name && i+1 < len(args) {
			return args[i+1]
		}
		prefix := "--" + name + "="
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return fallbackValue
}

func plannedCommandVerifyRequested(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "--verify", "--strict", "--context=release_verify":
			return true
		}
		if arg == "--context" {
			continue
		}
	}
	for i, arg := range args {
		if arg == "--context" && i+1 < len(args) && args[i+1] == "release_verify" {
			return true
		}
	}
	return false
}

func emitPlannedReport(stdout io.Writer, stderr io.Writer, command, status string, details []string, gaps []string, args []string) int {
	exitCode := emitReport(stdout, command, status, details, gaps)
	if status == "planned" || status == "gap" {
		if plannedCommandVerifyRequested(args) {
			write(stderr, "ERROR: %s is %s under --verify/strict context\n", command, status)
		} else {
			write(stderr, "ERROR: %s is %s and cannot satisfy a release gate\n", command, status)
		}
	}
	return exitCode
}
