#!/usr/bin/env bash
set -euo pipefail

required_files=(
  "README.md"
  "docs/standard/README.md"
  "docs/standard/xlib-standard.md"
  "docs/standard/repository-roles.md"
  "docs/standard/layering.md"
  "docs/standard/module-boundary.md"
  "docs/standard/harness-gates.md"
  "docs/standard/release-standard.md"
  "docs/standard/security-and-secret-policy.md"
  "docs/standard/retrospective-and-patches.md"
  "docs/standard/evidence-protocol.md"
  "docs/standard/template-generation-contract.md"
  "docs/standard/dod.md"
  "docs/standard/downstream-compatibility.md"
  "docs/downstream-sync-policy.md"
  "docs/scorecard.md"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: required documentation file missing: $file" >&2
    exit 1
  fi
done

require_text() {
  local file="$1"
  local needle="$2"

  if ! grep -Fq -- "$needle" "$file"; then
    echo "ERROR: $file must mention: $needle" >&2
    exit 1
  fi
}

require_text "README.md" "GOWORK=off make docs-check"
require_text "README.md" "GOWORK=off make dependency-check"
require_text "README.md" "GOWORK=off make standard-impact-check"
require_text "README.md" "GOWORK=off make release-check"
require_text "README.md" "DONE with evidence:"
require_text "README.md" "release/manifest/latest.json"
require_text "README.md" "release/manifest/latest.json.sha256"
require_text "README.md" "release/standard-impact/latest.md"
require_text "README.md" "renovate.json"
require_text "README.md" ".github/dependabot.yml"
require_text "README.md" "downstream_sync_required"
require_text "README.md" "FUZZ_SMOKE_TIME"
require_text "README.md" "docs/downstream-sync-policy.md"
require_text "README.md" "kernel"
require_text "docs/standard/README.md" "GOWORK=off make docs-check"
require_text "docs/standard/README.md" "GOWORK=off make dependency-check"
require_text "docs/standard/README.md" "GOWORK=off make standard-impact-check"
require_text "docs/standard/README.md" "GOWORK=off make release-check"
require_text "docs/standard/README.md" "release/manifest/latest.json"
require_text "docs/standard/README.md" "release/manifest/latest.json.sha256"
require_text "docs/standard/README.md" "FUZZ_SMOKE_TIME"
require_text "docs/standard/README.md" "../downstream-sync-policy.md"
require_text "docs/downstream-sync-policy.md" "xlib-standard"
require_text "docs/downstream-sync-policy.md" "kernel"
require_text "docs/downstream-sync-policy.md" "corekit"
require_text "docs/downstream-sync-policy.md" "L1 基础库"
require_text "docs/downstream-sync-policy.md" "x.go 仅作为基础库消费方"
require_text "docs/downstream-sync-policy.md" "downstream-sync-required"
require_text "docs/downstream-sync-policy.md" "release/standard-impact/latest.md"
require_text "docs/downstream-sync-policy.md" "downstream_sync_required"
require_text "docs/supply-chain.md" "kernel"
require_text "docs/supply-chain.md" '旧 `foundationx` 只作为迁移兼容扫描项'
require_text "docs/standard/evidence-protocol.md" "release/manifest/template.json"
require_text "docs/standard/evidence-protocol.md" "release/manifest/latest.json"
require_text "docs/standard/evidence-protocol.md" "artifact_url"
require_text "docs/standard/evidence-protocol.md" "sha256"
require_text "docs/standard/evidence-protocol.md" "workflow_run_id"
require_text "docs/standard/evidence-protocol.md" "standard_impact"
require_text "docs/standard/evidence-protocol.md" "downstream_sync_required"
require_text "docs/standard/evidence-protocol.md" "generator_evidence"
require_text "docs/standard/evidence-protocol.md" "dependency_check"
require_text "docs/standard/evidence-protocol.md" "GOWORK=off make dependency-check"
require_text "docs/standard/evidence-protocol.md" "GOWORK=off make standard-impact-check"
require_text "docs/standard/release-standard.md" "release/manifest/latest.json.sha256"
require_text "release/manifest/template.json" "release/manifest/latest.json.sha256"
require_text "release/manifest/template.json" '"dependencies"'
require_text "release/manifest/template.json" '"standard_impact"'
require_text "release/manifest/template.json" '"downstream_sync_required"'
require_text "release/manifest/template.json" '"generator_evidence"'
require_text "release/manifest/template.json" '"dependency_check"'
require_text "docs/scorecard.md" "go run ./cmd/xlibgate score --min 9.8"
require_text "docs/scorecard.md" "RELEASE_EVIDENCE_MIN_SCORE=9.5"
require_text "release/manifest/template.json" '"score"'
require_text "release/manifest/template.json" '"workflow_run_id"'
require_text "internal/tools/releasemanifest/main.go" "min-score"
require_text "Makefile" "go run ./cmd/xlibgate score --min 9.8"
require_text "Makefile" "RELEASE_EVIDENCE_MIN_SCORE=9.5"
require_text ".agent/release-template.md" "go run ./cmd/xlibgate score --min 9.8"
require_text ".agent/retrospective-template.md" "Score"
require_text ".agent/harness.yaml" "go run ./cmd/xlibgate score --min 9.8"
require_text "internal/tools/releasemanifest/main.go" "release/manifest/latest.json.sha256"
require_text "cmd/xlibgate/main.go" "docs-check"
require_text "cmd/xlibgate/main.go" "dependency-check"
require_text "cmd/xlibgate/main.go" "standard-impact-check"
require_text "cmd/xlibgate/main.go" "boundary"
require_text "cmd/xlibgate/main.go" "contracts"
require_text "cmd/xlibgate/main.go" "render-check"
require_text "cmd/xlibgate/main.go" "release-final-check"
require_text "cmd/xlibgate/main.go" "score"
require_text "cmd/xlibgate/main.go" "--min"
require_text "Makefile" "GOWORK=off is required for release targets"
require_text "Makefile" "XLIBGATE ?= go run ./cmd/xlibgate"
require_text "Makefile" '$(XLIBGATE) docs-check'
require_text "Makefile" '$(XLIBGATE) dependency-check'
require_text "Makefile" '$(XLIBGATE) standard-impact-check'
require_text "Makefile" '$(XLIBGATE) boundary'
require_text "Makefile" '$(XLIBGATE) contracts'
require_text "Makefile" '$(XLIBGATE) integration'
require_text "Makefile" '$(XLIBGATE) score --min 9.8'
require_text "Makefile" '$(XLIBGATE) release-evidence-checksum-check'
require_text "scripts/run_fuzz_smoke.sh" 'fuzz_time="${FUZZ_SMOKE_TIME:-10s}"'
require_text ".github/workflows/ci.yml" "make release-check"
require_text ".github/workflows/ci.yml" "release/manifest/latest.json.sha256"
require_text ".github/workflows/release.yml" "make release-final-check"
require_text ".github/workflows/release.yml" "release/manifest/latest.json.sha256"
require_text ".github/workflows/release.yml" "ARTIFACT_URL"
require_text ".github/workflows/ci.yml" "ARTIFACT_URL"

xlib_standard_url="https://github.com/ZoneCNH/xlib-standard"
require_text "README.md" "$xlib_standard_url"
require_text "docs/standard/README.md" "$xlib_standard_url"
require_text "docs/spec.md" "$xlib_standard_url"
require_text "docs/design.md" "$xlib_standard_url"
require_text "docs/generation.md" "$xlib_standard_url"
require_text "docs/standard/xlib-standard.md" "$xlib_standard_url"
require_text "docs/standard/repository-roles.md" "$xlib_standard_url"
require_text "docs/standard/harness-gates.md" "GOWORK=off make dependency-check"
require_text "docs/standard/harness-gates.md" "GOWORK=off make standard-impact-check"
require_text "renovate.json" '"gomod"'
require_text "renovate.json" '"github-actions"'
require_text ".github/dependabot.yml" 'package-ecosystem: "gomod"'
require_text ".github/dependabot.yml" 'package-ecosystem: "github-actions"'

python3 - "$PWD" <<'PY'
import sys
from pathlib import Path

root = Path(sys.argv[1])
requirements = {
    "docs/standard/xlib-standard.md": [
        "xlib-standard",
        "baselib-template",
        "模板",
        "generator",
        "Harness",
        "Evidence",
    ],
    "docs/standard/repository-roles.md": [
        "xlib-standard",
        "baselib-template",
        "标准权威源",
        "模板、generator、Harness、Evidence 实现仓库",
    ],
    "docs/standard/layering.md": [
        "xlib-standard",
        "baselib-template",
        "Standard 规则的独立来源",
        "Go 基础库模板中的实现仓库",
    ],
    "docs/standard/module-boundary.md": [
        "xlib-standard",
        "baselib-template",
        "go.mod",
        "module path",
    ],
    "docs/downstream-sync-policy.md": [
        "Standard Source",
        "Go Reference Template",
        "Generator",
        "L0 代表下游",
        "L1 基础库",
        "x.go 仅作为基础库消费方",
        "kernel",
        "corekit",
        "downstream-sync-required",
    ],
}

errors = []
for rel, needles in requirements.items():
    text = (root / rel).read_text(encoding="utf-8")
    for needle in needles:
        if needle not in text:
            errors.append(f"{rel} must mention: {needle}")

if errors:
    for error in errors:
        print(f"ERROR: {error}", file=sys.stderr)
    sys.exit(1)
PY

python3 - "$PWD" <<'PY'
import re
import sys
from pathlib import Path

root = Path(sys.argv[1])
scan_files = [
    root / "README.md",
    root / "docs/supply-chain.md",
    *sorted((root / "docs/standard").glob("*.md")),
]

bad_current_name_patterns = [
    (re.compile(r"渲染\s*`?foundationx`?"), "current downstream render target must be kernel/corekit"),
    (re.compile(r"生成\s*`?foundationx`?"), "current generated downstream target must be kernel/corekit"),
    (re.compile(r"默认下游[^。\n]*foundationx"), "default downstream must be kernel"),
    (re.compile(r"foundationx[^。\n]*默认下游"), "default downstream must be kernel"),
    (re.compile(r"foundationx[^。\n]*代表下游"), "representative downstream must be kernel/corekit"),
]

errors = []
for path in scan_files:
    text = path.read_text(encoding="utf-8")
    rel = path.relative_to(root)
    for pattern, message in bad_current_name_patterns:
        for match in pattern.finditer(text):
            sentence_start = max(text.rfind("。", 0, match.start()), text.rfind("\n", 0, match.start())) + 1
            sentence_end_candidates = [idx for idx in (text.find("。", match.end()), text.find("\n", match.end())) if idx != -1]
            sentence_end = min(sentence_end_candidates) if sentence_end_candidates else len(text)
            sentence = text[sentence_start:sentence_end]
            if any(word in sentence for word in ("旧", "迁移", "兼容", "历史")):
                continue
            errors.append(f"{rel}: {message}: {match.group(0)}")

if errors:
    for error in errors:
        print(f"ERROR: {error}", file=sys.stderr)
    sys.exit(1)
PY

if ! git check-ignore -q release/manifest/latest.json; then
  echo "ERROR: release/manifest/latest.json must remain ignored because it is generated Evidence" >&2
  exit 1
fi

if ! git check-ignore -q release/manifest/latest.json.sha256; then
  echo "ERROR: release/manifest/latest.json.sha256 must remain ignored because it is generated Evidence" >&2
  exit 1
fi

python3 - "$PWD/Makefile" <<'PY'
import re
import sys
from pathlib import Path

makefile = Path(sys.argv[1]).read_text(encoding="utf-8")
errors = []

for target in ("release-check", "release-check-extended"):
    match = re.search(rf"^{re.escape(target)}:([^\n]*)", makefile, re.MULTILINE)
    if not match:
        errors.append(f"Makefile target missing: {target}")
        continue
    deps = match.group(1).split()
    body_match = re.search(
        rf"^{re.escape(target)}:[^\n]*\n((?:\t.*\n)+)",
        makefile,
        re.MULTILINE,
    )
    body = body_match.group(1) if body_match else ""
    for dep in ("dependency-check", "standard-impact-check", "docs-check", "require-gowork-off"):
        if dep not in deps:
            errors.append(f"Makefile {target} must depend on {dep}")
    if "release-evidence-hash" not in body:
        errors.append(f"Makefile {target} must generate release Evidence checksum")
    if "release-evidence-checksum-check" not in body:
        errors.append(f"Makefile {target} must verify release Evidence checksum")

if errors:
    for error in errors:
        print(f"ERROR: {error}", file=sys.stderr)
    sys.exit(1)
PY

scan_files=("README.md")
while IFS= read -r -d '' file; do
  scan_files+=("$file")
done < <(find docs/standard -maxdepth 1 -type f -name '*.md' -print0)

if command -v rg >/dev/null 2>&1; then
  if rg -n '\{\{[^}]+\}\}|TODO_TEMPLATE' "${scan_files[@]}"; then
    echo "ERROR: unresolved template placeholder found in README.md or docs/standard/*.md" >&2
    exit 1
  fi
else
  if grep -nE '\{\{[^}]+\}\}|TODO_TEMPLATE' "${scan_files[@]}"; then
    echo "ERROR: unresolved template placeholder found in README.md or docs/standard/*.md" >&2
    exit 1
  fi
fi

mapfile -t markdown_files < <(find docs -type f -name '*.md' | sort)
python3 - "$PWD" "README.md" "${markdown_files[@]}" <<'PY'
import os
import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlparse

root = Path(sys.argv[1])
files = [Path(p) for p in sys.argv[2:]]
link_re = re.compile(r'!?\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)')
errors = []

for rel in files:
    path = root / rel
    text = path.read_text(encoding="utf-8")
    for match in link_re.finditer(text):
        raw_target = match.group(1).strip()
        if not raw_target or raw_target.startswith("#"):
            continue
        target = raw_target.strip("<>")
        parsed = urlparse(target)
        if parsed.scheme or target.startswith("//") or target.startswith("mailto:"):
            continue
        target_path = unquote(target.split("#", 1)[0])
        if not target_path:
            continue
        resolved = (path.parent / target_path).resolve()
        try:
            resolved.relative_to(root)
        except ValueError:
            errors.append(f"{rel}: local link escapes repository: {raw_target}")
            continue
        if not resolved.exists():
            errors.append(f"{rel}: broken local link: {raw_target}")

if errors:
    for error in errors:
        print(error, file=sys.stderr)
    sys.exit(1)
PY

echo "docs-check passed"
