package main

import (
	"io"
	"os"
	"strings"
)

type contextProfile struct {
	Name        string   `json:"name"`
	Target      string   `json:"target"`
	LegacyAlias string   `json:"legacy_alias,omitempty"`
	Components  []string `json:"components"`
}

var contextProfileBaseline = []contextProfile{
	{
		Name:        "lite",
		Target:      "context-lite",
		LegacyAlias: "context-fast-check",
		Components:  []string{"governance-check", "context-profile-check"},
	},
	{
		Name:        "standard",
		Target:      "context-standard",
		LegacyAlias: "context-standard-check",
		Components:  []string{"governance-check", "p1-governance-check", "context-profile-check"},
	},
	{
		Name:        "full",
		Target:      "context-full",
		LegacyAlias: "context-full-check",
		Components:  []string{"governance-check", "p1-governance-check", "p2-runtime-check", "context-profile-check"},
	},
	{
		Name:       "release",
		Target:     "context-release",
		Components: []string{"context-full", "integration", "dependency-check", "standard-impact-check", "score-check", "evidence", "release-evidence-hash", "release-evidence-check", "release-evidence-checksum-check"},
	},
}

func runContextProfile(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("context-profile", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("context-profile", err, stderr)
	}
	details := []string{"Context Runtime v4.0 Profile Baseline MVA", "registries remain SSOT", agentContextDetail()}
	for _, profile := range contextProfileBaseline {
		details = append(details, profile.Target+" => "+strings.Join(profile.Components, ","))
		if profile.LegacyAlias != "" {
			details = append(details, profile.LegacyAlias+" aliases "+profile.Target)
		}
	}
	return emitReport(stdout, "context-profile", "passed", details, nil)
}

func runContextProfileCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("context-profile-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("context-profile-check", err, stderr)
	}
	return runContextRuntimeCheck("context-profile-check", stdout, stderr)
}

func runContextSchemaCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("context-schema-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("context-schema-check", err, stderr)
	}
	return runContextRuntimeCheck("context-schema-check", stdout, stderr)
}

func runContextRuntimeCheck(command string, stdout io.Writer, stderr io.Writer) int {
	requiredTargets := contextRuntimeTargets()
	requiredCommands := []string{"context-profile", "context-profile-check", "context-schema-check", "context-fast-check", "context-standard-check", "context-full-check"}
	required := map[string][]string{
		"Makefile":                             {},
		".agent/makefile-target-registry.yaml": requiredTargets,
		".agent/makefile-baseline.yaml":        requiredTargets,
		".agent/command-registry.yaml":         requiredCommands,
	}
	for _, target := range requiredTargets {
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
	if makefile, err := os.ReadFile("Makefile"); err == nil {
		if releaseBody := makeTargetBody(string(makefile), "context-release"); strings.Contains(releaseBody, "release-check") || strings.Contains(releaseBody, "release-final-check") {
			gaps = append(gaps, "Makefile context-release must not call release-check or release-final-check")
		}
		if finalBody := makeTargetBody(string(makefile), "release-final-check"); !strings.Contains(finalBody, "context-release") {
			gaps = append(gaps, "Makefile release-final-check must call context-release")
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: %s found %d gap(s)\n", command, len(gaps))
		return emitReport(stdout, command, "failed", []string{agentContextDetail()}, gaps)
	}
	return emitReport(stdout, command, "passed", []string{"Context Runtime v4.0 profile baseline satisfied", "context-release is acyclic from release-check/release-final-check", "release-final-check delegates to context-release", agentContextDetail()}, nil)
}

func contextRuntimeTargets() []string {
	return []string{"context-profile", "context-profile-check", "context-schema-check", "context-lite", "context-standard", "context-full", "context-release", "context-fast-check", "context-standard-check", "context-full-check"}
}

func agentContextDetail() string {
	if fileExists(".agent/context") {
		return ".agent/context materialized"
	}
	return ".agent/context not materialized; registry SSOT active"
}

func makeTargetBody(makefile string, target string) string {
	lines := strings.Split(makefile, "\n")
	prefix := target + ":"
	inTarget := false
	var body []string
	for _, line := range lines {
		if !inTarget {
			if line == prefix || strings.HasPrefix(line, prefix+" ") {
				inTarget = true
				body = append(body, line)
			}
			continue
		}
		if line == "" || strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") {
			body = append(body, line)
			continue
		}
		if strings.Contains(line, ":") && !strings.HasPrefix(line, ".PHONY:") {
			break
		}
		body = append(body, line)
	}
	return strings.Join(body, "\n")
}
