#!/usr/bin/env bash
set -euo pipefail

report_path="${STANDARD_IMPACT_REPORT:-release/standard-impact/latest.md}"
base_ref="${STANDARD_IMPACT_BASE:-}"

downstreams=(
  "github.com/ZoneCNH/kernel"
  "github.com/ZoneCNH/configx"
  "github.com/ZoneCNH/observex"
  "github.com/ZoneCNH/testkitx"
  "github.com/ZoneCNH/postgresx"
  "github.com/ZoneCNH/redisx"
  "github.com/ZoneCNH/kafkax"
  "github.com/ZoneCNH/taosx"
  "github.com/ZoneCNH/ossx"
  "github.com/ZoneCNH/clickhousex"
  "github.com/ZoneCNH/x.go"
)

changed_files=()

add_file() {
  local file="$1"

  [[ -n "$file" ]] || return 0
  [[ "$file" == .git/* ]] && return 0
  [[ "$file" == .omc/* || "$file" == .omx/* || "$file" == .worktree/* ]] && return 0

  local existing
  for existing in "${changed_files[@]}"; do
    if [[ "$existing" == "$file" ]]; then
      return 0
    fi
  done

  changed_files+=("$file")
}

collect_git_diff() {
  local range="$1"
  local file

  while IFS= read -r file; do
    add_file "$file"
  done < <(git diff --name-only --diff-filter=ACDMRTUXB "$range" --)
}

collect_worktree_changes() {
  local status path

  while IFS= read -r status; do
    path="${status:3}"
    if [[ "$path" == *" -> "* ]]; then
      path="${path##* -> }"
    fi
    add_file "$path"
  done < <(git status --porcelain --untracked-files=all)
}

collect_upstream_diff() {
  local upstream="$1"
  local merge_base

  merge_base="$(git merge-base "$upstream" HEAD)"
  collect_git_diff "${merge_base}...HEAD"
}

sort_changed_files() {
  if (( ${#changed_files[@]} == 0 )); then
    return 0
  fi

  local sorted_files=()
  local file
  while IFS= read -r file; do
    sorted_files+=("$file")
  done < <(printf '%s\n' "${changed_files[@]}" | LC_ALL=C sort)

  changed_files=("${sorted_files[@]}")
}

report_generated_at() {
  if [[ -n "${STANDARD_IMPACT_GENERATED_AT:-}" ]]; then
    printf '%s\n' "$STANDARD_IMPACT_GENERATED_AT"
    return 0
  fi

  local author_date
  if author_date="$(git show -s --format=%aI HEAD 2>/dev/null)" && [[ -n "$author_date" ]]; then
    date -u -d "$author_date" +%Y-%m-%dT%H:%M:%SZ
    return 0
  fi

  date -u +%Y-%m-%dT%H:%M:%SZ
}

if [[ -n "$base_ref" ]]; then
  collect_git_diff "$base_ref"
elif [[ -n "${GITHUB_BASE_REF:-}" ]] && git rev-parse --verify "origin/${GITHUB_BASE_REF}" >/dev/null 2>&1; then
  collect_upstream_diff "origin/${GITHUB_BASE_REF}"
  collect_worktree_changes
elif upstream_ref="$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)" && [[ -n "$upstream_ref" ]]; then
  collect_upstream_diff "$upstream_ref"
  collect_worktree_changes
else
  collect_worktree_changes
fi

sort_changed_files

docs_files=()
contracts_files=()
context_runtime_files=()
governance_registry_files=()
harness_files=()
repository_rules_files=()
generator_files=()
downstream_context_files=()
evidence_files=()
other_files=()

add_category_file() {
  local category="$1"
  local file="$2"

  case "$category" in
    docs) docs_files+=("$file") ;;
    contracts) contracts_files+=("$file") ;;
    context_runtime) context_runtime_files+=("$file") ;;
    governance_registry) governance_registry_files+=("$file") ;;
    harness) harness_files+=("$file") ;;
    repository_rules) repository_rules_files+=("$file") ;;
    generator) generator_files+=("$file") ;;
    downstream_context) downstream_context_files+=("$file") ;;
    evidence) evidence_files+=("$file") ;;
    other) other_files+=("$file") ;;
  esac
}

classify_file() {
  local file="$1"

  case "$file" in
    .agent/command-registry.yaml|.agent/issue-registry.yaml|.agent/makefile-baseline.yaml|.agent/makefile-target-registry.yaml)
      add_category_file "governance_registry" "$file"
      ;;
    templates/context-consumer/*)
      add_category_file "downstream_context" "$file"
      ;;
    AGENTS.md|Makefile|.github/CODEOWNERS|.github/rulesets/*|.github/workflows/*|infra/github-rules/*|.agent/gates.md)
      add_category_file "repository_rules" "$file"
      ;;
    cmd/xlibgate/*|.agent/context/*|docs/standard/xlibgate-cli-contract.md|docs/standard/release-standard.md|docs/standard/harness-gates.md|docs/standard/evidence-protocol.md)
      add_category_file "context_runtime" "$file"
      ;;
    contracts/*)
      add_category_file "contracts" "$file"
      ;;
    release/manifest/*|release/standard-impact/*|internal/tools/releasemanifest/*|scripts/generate_manifest.sh|scripts/hash_release_evidence.sh|scripts/check_release_evidence.sh|scripts/check_release_preflight.sh)
      add_category_file "evidence" "$file"
      ;;
    scripts/render_template.sh|scripts/check_rendered_template.sh|examples/*|testkit/*)
      add_category_file "generator" "$file"
      ;;
    scripts/check_*.sh|scripts/run_*.sh|.agent/harness*)
      add_category_file "harness" "$file"
      ;;
    docs/*|README.md|.agent/*)
      add_category_file "docs" "$file"
      ;;
    *)
      add_category_file "other" "$file"
      ;;
  esac
}

for file in "${changed_files[@]}"; do
  classify_file "$file"
done

downstream_sync_required="false"
if (( ${#contracts_files[@]} > 0 || ${#context_runtime_files[@]} > 0 || ${#governance_registry_files[@]} > 0 || ${#harness_files[@]} > 0 || ${#repository_rules_files[@]} > 0 || ${#generator_files[@]} > 0 || ${#downstream_context_files[@]} > 0 || ${#evidence_files[@]} > 0 )); then
  downstream_sync_required="true"
fi
context_runtime_change="false"
if (( ${#context_runtime_files[@]} > 0 )); then
  context_runtime_change="true"
fi
governance_registry_change="false"
if (( ${#governance_registry_files[@]} > 0 )); then
  governance_registry_change="true"
fi
downstream_release_decision="downstream-sync-not-required"
if [[ "$downstream_sync_required" == "true" ]]; then
  downstream_release_decision="downstream-sync-required"
fi
repository_rules_release_decision="repository-rules-review-not-required"
if (( ${#repository_rules_files[@]} > 0 )); then
  repository_rules_release_decision="repository-rules-review-required"
fi

context_runtime_change="false"
if (( ${#context_runtime_files[@]} > 0 )); then
  context_runtime_change="true"
fi

governance_registry_change="false"
if (( ${#governance_registry_files[@]} > 0 )); then
  governance_registry_change="true"
fi

downstream_release_decision="downstream-sync-not-required"
if [[ "$downstream_sync_required" == "true" ]]; then
  downstream_release_decision="downstream-sync-required"
fi

repository_rules_release_decision="repository-rules-release-not-required"
if (( ${#repository_rules_files[@]} > 0 )); then
  repository_rules_release_decision="repository-rules-release-required"
fi

context_runtime_change="false"
if (( ${#context_runtime_files[@]} > 0 )); then
  context_runtime_change="true"
fi

governance_registry_change="false"
if (( ${#governance_registry_files[@]} > 0 )); then
  governance_registry_change="true"
fi

downstream_release_decision="downstream-sync-not-required"
if [[ "$downstream_sync_required" == "true" ]]; then
  downstream_release_decision="downstream-sync-required"
fi

repository_rules_release_decision="repository-rules-release-not-required"
if (( ${#repository_rules_files[@]} > 0 )); then
  repository_rules_release_decision="repository-rules-release-required"
fi

context_runtime_change="unchanged"
if (( ${#context_runtime_files[@]} > 0 )); then
  context_runtime_change="changed"
fi

governance_registry_change="unchanged"
if (( ${#governance_registry_files[@]} > 0 )); then
  governance_registry_change="changed"
fi

downstream_release_decision="sync-not-required"
if [[ "$downstream_sync_required" == "true" ]]; then
  downstream_release_decision="sync-required"
fi

repository_rules_release_decision="release-review-not-required"
if (( ${#repository_rules_files[@]} > 0 )); then
  repository_rules_release_decision="release-review-required"
fi

context_runtime_change="unchanged"
if (( ${#context_runtime_files[@]} > 0 )); then
  context_runtime_change="changed"
fi

governance_registry_change="unchanged"
if (( ${#governance_registry_files[@]} > 0 )); then
  governance_registry_change="changed"
fi

downstream_release_decision="sync-not-required"
if [[ "$downstream_sync_required" == "true" ]]; then
  downstream_release_decision="sync-required"
fi

repository_rules_release_decision="release-review-not-required"
if (( ${#repository_rules_files[@]} > 0 )); then
  repository_rules_release_decision="release-review-required"
fi

context_runtime_change="unchanged"
if (( ${#context_runtime_files[@]} > 0 )); then
  context_runtime_change="changed"
fi

governance_registry_change="unchanged"
if (( ${#governance_registry_files[@]} > 0 )); then
  governance_registry_change="changed"
fi

downstream_release_decision="sync-not-required"
if [[ "$downstream_sync_required" == "true" ]]; then
  downstream_release_decision="sync-required"
fi

repository_rules_release_decision="release-review-not-required"
if (( ${#repository_rules_files[@]} > 0 )); then
  repository_rules_release_decision="release-review-required"
fi

mkdir -p "$(dirname "$report_path")"

write_file_list() {
  local title="$1"
  shift
  local files=("$@")

  printf '## %s\n\n' "$title"
  if (( ${#files[@]} == 0 )); then
    printf -- '- 无变化\n\n'
    return 0
  fi

  local file
  for file in "${files[@]}"; do
    printf -- '- `%s`\n' "$file"
  done
  printf '\n'
}

{
  printf '# Standard Impact Report\n\n'
  printf -- '- generated_at: `%s`\n' "$(report_generated_at)"
  printf -- '- downstream_sync_required: `%s`\n' "$downstream_sync_required"
  printf -- '- context_runtime_change: `%s`\n' "$context_runtime_change"
  printf -- '- governance_registry_change: `%s`\n' "$governance_registry_change"
  printf -- '- downstream_release_decision: `%s`\n' "$downstream_release_decision"
  printf -- '- repository_rules_release_decision: `%s`\n' "$repository_rules_release_decision"
  printf -- '- primary_downstream: `%s`\n' "github.com/ZoneCNH/kernel"
  printf -- '- context_runtime_change: `%s`\n' "$context_runtime_change"
  printf -- '- governance_registry_change: `%s`\n' "$governance_registry_change"
  printf -- '- downstream_release_decision: `%s`\n' "$downstream_release_decision"
  printf -- '- repository_rules_release_decision: `%s`\n' "$repository_rules_release_decision"
  printf -- '- changed_file_count: `%s`\n\n' "${#changed_files[@]}"

  printf '## Downstream\n\n'
  for downstream in "${downstreams[@]}"; do
    printf -- '- `%s`\n' "$downstream"
  done
  printf '\n'

  write_file_list "docs" "${docs_files[@]}"
  write_file_list "contracts" "${contracts_files[@]}"
  write_file_list "context_runtime" "${context_runtime_files[@]}"
  write_file_list "governance_registry" "${governance_registry_files[@]}"
  write_file_list "harness" "${harness_files[@]}"
  write_file_list "repository_rules" "${repository_rules_files[@]}"
  write_file_list "generator" "${generator_files[@]}"
  write_file_list "downstream_context" "${downstream_context_files[@]}"
  write_file_list "evidence" "${evidence_files[@]}"
  write_file_list "other" "${other_files[@]}"

  printf '## Sync Decision\n\n'
  if [[ "$downstream_sync_required" == "true" ]]; then
    printf -- '- `downstream-sync-required`\n'
    printf -- '- 原因：contracts、context_runtime、governance_registry、harness、repository_rules、generator、downstream_context 或 evidence 影响面发生变化。\n'
  else
    printf -- '- `downstream-sync-not-required`\n'
    printf -- '- 原因：未发现 contracts、context_runtime、governance_registry、harness、repository_rules、generator、downstream_context 或 evidence 影响面变化。\n'
  fi
  printf -- '- `%s`\n' "$repository_rules_release_decision"
} > "$report_path"

echo "standard impact report generated: $report_path"
