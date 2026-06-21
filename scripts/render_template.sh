#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/render_template.sh --module-name NAME --module-path PATH --package-name NAME --out DIR

Renders the current repository into a concrete base library by copying the
repository, moving the template package to pkg/<package>, and replacing
template identifiers.
USAGE
}

module_name=""
module_path=""
package_name=""
out_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module-name)
      module_name="${2:-}"
      shift 2
      ;;
    --module-path)
      module_path="${2:-}"
      shift 2
      ;;
    --package-name)
      package_name="${2:-}"
      shift 2
      ;;
    --out)
      out_dir="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$module_name" || -z "$module_path" || -z "$package_name" || -z "$out_dir" ]]; then
  echo "ERROR: --module-name, --module-path, --package-name and --out are required" >&2
  usage >&2
  exit 2
fi

if [[ "$package_name" =~ [^a-zA-Z0-9_] || "$package_name" =~ ^[0-9] ]]; then
  echo "ERROR: --package-name must be a valid Go package identifier" >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repo_abs="$(realpath "$repo_root")"
out_abs="$(realpath -m "$out_dir")"
source_module_path="$(cd "$repo_root" && GOWORK=off go list -m)"
source_module_name="${source_module_path##*/}"
go_name_template="{"
go_name_template="${go_name_template}{.Name}"
go_name_template="${go_name_template}}"
source_package_name="$(cd "$repo_root" && GOWORK=off go list -f "$go_name_template" "./pkg/$source_module_name" 2>/dev/null || true)"
if [[ -z "$source_package_name" ]]; then
  source_package_name="$(printf '%s' "$source_module_name" | tr '-' '_')"
fi
template_package_name="template"
template_package_name="${template_package_name}x"
if [[ ! -d "$repo_root/pkg/$template_package_name" ]]; then
  template_package_name="$source_package_name"
fi

if [[ "$out_abs" == "$repo_abs" || "$out_abs" == "$repo_abs"/* ]]; then
  echo "ERROR: output directory must be outside the template repository: $out_abs" >&2
  exit 2
fi

if [[ -e "$out_abs" && ! -d "$out_abs" ]]; then
  echo "ERROR: output path exists but is not a directory: $out_abs" >&2
  exit 2
fi

if [[ -e "$out_abs/.git" || -e "$out_abs/go.mod" ]]; then
  echo "ERROR: output directory looks like an existing repository: $out_abs" >&2
  exit 2
fi

if [[ -d "$out_abs" ]] && find "$out_abs" -mindepth 1 -maxdepth 1 | read -r _; then
  echo "ERROR: output directory must be empty: $out_abs" >&2
  exit 2
fi

mkdir -p "$out_abs"
out_dir="$out_abs"

copy_from_live_tree() {
  (
    cd "$repo_root"
    tar \
    --exclude='./.git' \
    --exclude='./.codex' \
    --exclude='./.omc' \
    --exclude='./.omx' \
    --exclude='./.worktree' \
    --exclude='./.agent/inbox' \
    --exclude='./docs/adr' \
    --exclude='./docs/goal.md' \
    --exclude='./tmp' \
    --exclude='./dist' \
    --exclude='./node_modules' \
    --exclude='./coverage.out' \
    --exclude='./coverage.*' \
    --exclude='./*.coverprofile' \
    --exclude='./profile.cov' \
    --exclude='./release/manifest/latest.json' \
    --exclude='./release/manifest/latest.json.sha256' \
    --exclude='./release/standard-impact/latest.md' \
    --exclude='./release/downstream-sync/latest.md' \
    --exclude='./release/debt/latest.json' \
    --exclude='./release/debt/latest.md' \
    --exclude='./release/debt/latest.json.sha256' \
    -cf - .
  ) | (
    cd "$out_dir"
    tar -xf -
  )
}

prune_render_omissions() {
  rm -rf "$out_dir/.codex"
  rm -rf "$out_dir/.omc"
  rm -rf "$out_dir/.omx"
  rm -rf "$out_dir/.worktree"
  rm -rf "$out_dir/.agent/inbox"
  rm -rf "$out_dir/docs/adr"
  rm -f "$out_dir/docs/goal.md"
  rm -f "$out_dir/release/manifest/latest.json"
  rm -f "$out_dir/release/manifest/latest.json.sha256"
  rm -f "$out_dir/release/standard-impact/latest.md"
  rm -f "$out_dir/release/downstream-sync/latest.md"
  rm -f "$out_dir/release/debt/latest.json"
  rm -f "$out_dir/release/debt/latest.md"
  rm -f "$out_dir/release/debt/latest.json.sha256"
}

copy_from_git_archive() {
  git -C "$repo_root" archive --format=tar HEAD | (
    cd "$out_dir"
    tar -xf -
  )
  prune_render_omissions
}

use_git_archive=0
if [[ "${XLIB_RENDER_FORCE_GIT_ARCHIVE:-0}" == "1" ]]; then
  use_git_archive=1
elif git -C "$repo_root" rev-parse --is-inside-work-tree >/dev/null 2>&1 && \
  [[ -z "$(git -C "$repo_root" status --porcelain=v1 --untracked-files=no)" ]]; then
  use_git_archive=1
fi

if [[ "$use_git_archive" == "1" ]]; then
  copy_from_git_archive
else
  copy_from_live_tree
fi

# Raw inbox archives are intentionally omitted from rendered downstream repos.
# Keep the rendered control-plane index aligned with that reduced file set.
index_path="$out_dir/.agent/index.yaml"
if [[ -f "$index_path" ]]; then
  awk '
    /^  - path: \.agent\/inbox\// {
      skip = 1
      next
    }
    skip && /^    / {
      next
    }
    {
      skip = 0
      print
    }
  ' "$index_path" > "$index_path.tmp"
  mv "$index_path.tmp" "$index_path"
fi

if [[ "$package_name" != "$template_package_name" ]]; then
  mkdir -p "$out_dir/pkg"
  mv "$out_dir/pkg/$template_package_name" "$out_dir/pkg/$package_name"
fi

if [[ "$source_package_name" != "$template_package_name" && "$source_package_name" != "$package_name" ]]; then
  rm -rf "$out_dir/pkg/$source_package_name"
fi

replace_in_text_files() {
  local find_text="$1"
  local replace_text="$2"

  while IFS= read -r -d '' file; do
    FIND_TEXT="$find_text" REPLACE_TEXT="$replace_text" perl -0pi -e 's/\Q$ENV{FIND_TEXT}\E/$ENV{REPLACE_TEXT}/g' "$file"
  done < <(
    find "$out_dir" -type f \( \
      -name '*.go' -o \
      -name '*.md' -o \
      -name '*.json' -o \
      -name '*.sh' -o \
      -name '*.yml' -o \
      -name '*.yaml' -o \
      -name 'Makefile' -o \
      -name 'go.mod' \
    \) ! -path "$out_dir/scripts/render_template.sh" -print0
  )
}

replace_package_identity() {
  local source_name="$1"
  local source_title
  local source_upper

  if [[ "$source_name" == "$package_name" ]]; then
    return
  fi

  source_title="$(printf '%s%s' "$(printf '%s' "${source_name:0:1}" | tr '[:lower:]' '[:upper:]')" "${source_name:1}")"
  source_upper="$(printf '%s' "$source_name" | tr '[:lower:]' '[:upper:]')"

  replace_in_text_files "${source_name}_" "${package_name}_"
  replace_in_text_files "$source_title" "$package_title"
  replace_in_text_files "$source_upper" "$package_upper"
  replace_in_text_files "pkg/$source_name" "pkg/$package_name"
  replace_in_text_files "$source_name" "$package_name"
}

template_token() {
  local open="{"
  local close="}"
  printf '%s%s%s%s%s' "$open" "$open" "$1" "$close" "$close"
}

replace_in_text_files "$(template_token MODULE_NAME)" "$module_name"
replace_in_text_files "$(template_token MODULE_PATH)" "$module_path"
replace_in_text_files "$(template_token PACKAGE_NAME)" "$package_name"
standard_module_name="xlib"
standard_module_name="${standard_module_name}-standard"
legacy_module_name="baselib"
legacy_module_name="${legacy_module_name}-template"
replace_in_text_files "github.com/ZoneCNH/$standard_module_name" "$module_path"
replace_in_text_files "github.com/ZoneCNH/$legacy_module_name" "$module_path"
if [[ "$source_module_path" != "$module_path" ]]; then
  replace_in_text_files "$source_module_path" "$module_path"
fi
replace_in_text_files "$standard_module_name" "$module_name"
replace_in_text_files "$legacy_module_name" "$module_name"
package_title="$(printf '%s%s' "$(printf '%s' "${package_name:0:1}" | tr '[:lower:]' '[:upper:]')" "${package_name:1}")"
package_upper="$(printf '%s' "$package_name" | tr '[:lower:]' '[:upper:]')"
replace_package_identity "$template_package_name"
replace_package_identity "$source_package_name"

write_rendered_generic_samples() {
  rm -rf "$out_dir/examples/basic" "$out_dir/examples/config" "$out_dir/examples/health"
  mkdir -p "$out_dir/examples/basic" "$out_dir/contracts"

  cat > "$out_dir/examples/basic/main.go" <<EXAMPLE_GO
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"$module_path/pkg/$package_name"
)

func main() {
	run(os.Stdout, os.Stderr, $package_name.Config{
		Name:    "$module_name",
		Timeout: time.Second,
	})
}

func run(stdout, stderr io.Writer, cfg $package_name.Config) {
	client, err := $package_name.New(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "create client: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close(context.Background())
	}()

	_, _ = fmt.Fprintln(stdout, $package_name.ModuleName)
}
EXAMPLE_GO

  cat > "$out_dir/examples/basic/main_test.go" <<EXAMPLE_TEST_GO
package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"$module_path/pkg/$package_name"
)

func TestMainPrintsModuleName(t *testing.T) {
	output := captureStdout(t, main)
	if output != "$module_path\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunReportsInvalidConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer

	run(&stdout, &stderr, $package_name.Config{})

	if stdout.String() != "" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if stderr.String() != "create client: validation: Config.Validate: name is required\n" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = original
	})

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = original

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return buf.String()
}
EXAMPLE_TEST_GO

  cat > "$out_dir/contracts/contracts_test.go" <<'CONTRACT_TEST_GO'
package contracts

import (
	"encoding/json"
	"os"
	"testing"
)

func TestRequiredContractSchemasAreValidJSON(t *testing.T) {
	for _, path := range []string{
		"config.schema.json",
		"health.schema.json",
		"error.schema.json",
		"goalcli-report.schema.json",
		"issue-registry.schema.json",
		"command-registry.schema.json",
		"execution-context.schema.json",
		"conformance-attestation.schema.json",
		"policy.schema.json",
	} {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			var schema map[string]any
			if err := json.Unmarshal(content, &schema); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			if schema["$schema"] == "" || schema["type"] != "object" {
				t.Fatalf("%s must declare object JSON schema, got %#v", path, schema)
			}
		})
	}
}

func TestMetricsContractIsPresent(t *testing.T) {
	content, err := os.ReadFile("metrics.md")
	if err != nil {
		t.Fatalf("read metrics contract: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("metrics contract must not be empty")
	}
}
CONTRACT_TEST_GO
}

write_rendered_generic_samples

(
  cd "$out_dir"
  gofmt -w ./pkg ./internal ./contracts ./examples ./testkit
  GOWORK=off go mod tidy
)

echo "rendered $module_name at $out_dir"
