package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ZoneCNH/xlib-standard/internal/releasequality"
)

func TestMainDispatchesUsageHelpAndUnknownCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "no command",
			wantCode:   2,
			wantStderr: "usage: xlibgate <command>",
		},
		{
			name:       "help",
			args:       []string{"help"},
			wantCode:   0,
			wantStdout: "commands:",
		},
		{
			name:       "short help",
			args:       []string{"-h"},
			wantCode:   0,
			wantStdout: "commands:",
		},
		{
			name:       "long help",
			args:       []string{"--help"},
			wantCode:   0,
			wantStdout: "commands:",
		},
		{
			name:       "unknown",
			args:       []string{"missing"},
			wantCode:   2,
			wantStderr: `unknown command "missing"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			got := run(tt.args, strings.NewReader(""), &stdout, &stderr)

			if got != tt.wantCode {
				t.Fatalf("run(%v) = %d; want %d", tt.args, got, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Fatalf("stdout = %q; want containing %q", stdout.String(), tt.wantStdout)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q; want containing %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestMainUsesRunExitCode(t *testing.T) {
	originalArgs := os.Args
	originalExit := exit
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	t.Cleanup(func() {
		os.Args = originalArgs
		exit = originalExit
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open os.DevNull: %v", err)
	}
	t.Cleanup(func() { _ = devNull.Close() })
	os.Stdout = devNull
	os.Stderr = devNull
	os.Args = []string{"xlibgate", "help"}

	var got int
	exit = func(code int) {
		got = code
	}

	main()

	if got != 0 {
		t.Fatalf("main exit code = %d; want 0", got)
	}
}

func TestRunDispatchesExternalCommands(t *testing.T) {
	root := t.TempDir()
	writeGateScript(t, root, "scripts/check_boundary.sh")
	writeGateScript(t, root, "scripts/check_contracts.sh")
	writeGateScript(t, root, "scripts/check_dependency_diff.sh")
	writeGateScript(t, root, "scripts/check_docs.sh")
	writeGateScript(t, root, "scripts/run_integration.sh")
	writeGateScript(t, root, "scripts/check_release_evidence.sh")
	writeGateScript(t, root, "scripts/hash_release_evidence.sh")
	writeGateScript(t, root, "scripts/check_secrets.sh")
	writeGateScript(t, root, "scripts/check_standard_impact.sh")
	writeGateScript(t, root, "scripts/check_rendered_template.sh")
	writePathTool(t, root, "go")
	writePathTool(t, root, "make")
	chdir(t, root)
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))

	tests := []struct {
		name       string
		args       []string
		wantStdout string
	}{
		{name: "boundary", args: []string{"boundary"}, wantStdout: "check_boundary.sh"},
		{name: "contracts", args: []string{"contracts"}, wantStdout: "check_contracts.sh"},
		{name: "dependency-check", args: []string{"dependency-check"}, wantStdout: "check_dependency_diff.sh"},
		{name: "docs-check", args: []string{"docs-check"}, wantStdout: "check_docs.sh"},
		{name: "evidence", args: []string{"evidence"}, wantStdout: "go run ./internal/tools/releasemanifest --out release/manifest/latest.json"},
		{name: "integration", args: []string{"integration"}, wantStdout: "run_integration.sh"},
		{name: "release-evidence-check", args: []string{"release-evidence-check"}, wantStdout: "check_release_evidence.sh"},
		{name: "release-evidence-checksum-check", args: []string{"release-evidence-checksum-check"}, wantStdout: "hash_release_evidence.sh --check"},
		{name: "release-evidence-hash", args: []string{"release-evidence-hash"}, wantStdout: "hash_release_evidence.sh"},
		{name: "release-final-check", args: []string{"release-final-check"}, wantStdout: "make release-final-check"},
		{name: "render-check", args: []string{"render-check", "rendered"}, wantStdout: "check_rendered_template.sh rendered"},
		{name: "secrets", args: []string{"secrets"}, wantStdout: "check_secrets.sh"},
		{name: "standard-impact-check", args: []string{"standard-impact-check"}, wantStdout: "check_standard_impact.sh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			got := run(tt.args, strings.NewReader("input"), &stdout, &stderr)

			if got != 0 {
				t.Fatalf("run(%v) = %d, stderr %q; want 0", tt.args, got, stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Fatalf("stdout = %q; want containing %q", stdout.String(), tt.wantStdout)
			}
		})
	}
}

func TestRunScore(t *testing.T) {
	resetReleaseQuality := func() {
		computeReleaseQuality = releasequality.Compute
		marshalReleaseQuality = releasequality.Marshal
		verifyReleaseQuality = releasequality.Verify
	}
	t.Cleanup(resetReleaseQuality)

	t.Run("success", func(t *testing.T) {
		resetReleaseQuality()
		var gotMinimum float64
		computeReleaseQuality = func(minimum float64) releasequality.Report {
			gotMinimum = minimum
			return releasequality.Report{Value: 9.9, Threshold: minimum, Status: "passed"}
		}
		marshalReleaseQuality = func(report releasequality.Report) ([]byte, error) {
			if report.Value != 9.9 {
				t.Fatalf("marshal report Value = %.1f; want 9.9", report.Value)
			}
			return []byte(`{"status":"passed"}`), nil
		}
		verifyReleaseQuality = func(report releasequality.Report, minimum float64) error {
			if minimum != 9.7 {
				t.Fatalf("verify minimum = %.1f; want 9.7", minimum)
			}
			return nil
		}

		var stdout, stderr bytes.Buffer
		got := run([]string{"score", "--min", "9.7"}, strings.NewReader(""), &stdout, &stderr)

		if got != 0 {
			t.Fatalf("score exit = %d, stderr %q; want 0", got, stderr.String())
		}
		if gotMinimum != 9.7 {
			t.Fatalf("compute minimum = %.1f; want 9.7", gotMinimum)
		}
		if strings.TrimSpace(stdout.String()) != `{"status":"passed"}` {
			t.Fatalf("stdout = %q; want JSON", stdout.String())
		}
	})

	t.Run("flag parse error", func(t *testing.T) {
		resetReleaseQuality()
		var stdout, stderr bytes.Buffer
		got := run([]string{"score", "--min", "nope"}, strings.NewReader(""), &stdout, &stderr)
		if got != 2 {
			t.Fatalf("score parse exit = %d; want 2", got)
		}
		if !strings.Contains(stderr.String(), "invalid value") {
			t.Fatalf("stderr = %q; want invalid value", stderr.String())
		}
	})

	t.Run("flag help", func(t *testing.T) {
		resetReleaseQuality()
		var stdout, stderr bytes.Buffer
		got := run([]string{"score", "-h"}, strings.NewReader(""), &stdout, &stderr)
		if got != 0 {
			t.Fatalf("score help exit = %d; want 0", got)
		}
		if !strings.Contains(stderr.String(), "minimum acceptable release score") {
			t.Fatalf("stderr = %q; want help", stderr.String())
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		resetReleaseQuality()
		computeReleaseQuality = func(minimum float64) releasequality.Report {
			return releasequality.Report{Value: 10, Threshold: minimum, Status: "passed"}
		}
		marshalReleaseQuality = func(report releasequality.Report) ([]byte, error) {
			return nil, errors.New("marshal boom")
		}

		var stdout, stderr bytes.Buffer
		got := run([]string{"score"}, strings.NewReader(""), &stdout, &stderr)

		if got != 1 {
			t.Fatalf("score marshal exit = %d; want 1", got)
		}
		if !strings.Contains(stderr.String(), "marshal boom") {
			t.Fatalf("stderr = %q; want marshal error", stderr.String())
		}
	})

	t.Run("verify error", func(t *testing.T) {
		resetReleaseQuality()
		computeReleaseQuality = func(minimum float64) releasequality.Report {
			return releasequality.Report{Value: 1, Threshold: minimum, Status: "failed"}
		}
		marshalReleaseQuality = func(report releasequality.Report) ([]byte, error) {
			return []byte(`{"status":"failed"}`), nil
		}
		verifyReleaseQuality = func(report releasequality.Report, minimum float64) error {
			return fmt.Errorf("score too low")
		}

		var stdout, stderr bytes.Buffer
		got := run([]string{"score"}, strings.NewReader(""), &stdout, &stderr)

		if got != 1 {
			t.Fatalf("score verify exit = %d; want 1", got)
		}
		if !strings.Contains(stdout.String(), `"failed"`) {
			t.Fatalf("stdout = %q; want failed report", stdout.String())
		}
		if !strings.Contains(stderr.String(), "score too low") {
			t.Fatalf("stderr = %q; want verify error", stderr.String())
		}
	})
}

func TestRunExternalErrorPaths(t *testing.T) {
	t.Run("exit error returns command status", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		got := runExternal(strings.NewReader(""), &stdout, &stderr, shellPath(t), "-c", "exit 7")
		if got != 7 {
			t.Fatalf("runExternal exit status = %d; want 7", got)
		}
	})

	t.Run("non exit error is reported", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		got := runExternal(strings.NewReader(""), &stdout, &stderr, filepath.Join(t.TempDir(), "missing-command"))
		if got != 1 {
			t.Fatalf("runExternal missing command = %d; want 1", got)
		}
		if !strings.Contains(stderr.String(), "ERROR:") {
			t.Fatalf("stderr = %q; want ERROR", stderr.String())
		}
	})
}

func writeGateScript(t *testing.T, root string, relative string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	body := "#!/bin/sh\nprintf '%s' \"$(basename \"$0\")\"\nfor arg in \"$@\"; do printf ' %s' \"$arg\"; done\nprintf '\\n'\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writePathTool(t *testing.T, root string, name string) {
	t.Helper()
	path := filepath.Join(root, name)
	body := "#!/bin/sh\nprintf '%s' \"$(basename \"$0\")\"\nfor arg in \"$@\"; do printf ' %s' \"$arg\"; done\nprintf '\\n'\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd %s: %v", old, err)
		}
	})
}

func shellPath(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	for _, candidate := range []string{"/bin/sh", "/usr/bin/sh"} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	t.Fatal("no POSIX shell found")
	return ""
}
