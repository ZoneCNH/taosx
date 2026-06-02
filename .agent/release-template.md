# Release Template

> 模板说明：本文件是发布证据填写模板，不是 DONE 证据本身。只有填写了
> workflow run、artifact URL、sha256、score 和 gate 结果的新鲜实例才可作为发布证据。

## Version

- Version:
- Commit:
- Tree SHA:

## Required Evidence

- [ ] `GOWORK=off make release-check`
- [ ] `go run ./cmd/xlibgate score --min 9.8`
- [ ] `GOWORK=off make release-final-check`
- [ ] `GOWORK=off make release-preflight VERSION=<version>`
- [ ] `release/manifest/latest.json` generated and validated
- [ ] `release/manifest/latest.json.sha256` generated and validated

## Score

- Score JSON:
- Threshold:
- Status:

## Notes

- Template behavior impact:
- Generated library impact:
- Known gaps:

## Artifact

- CI artifact:
- Workflow run ID:
- Artifact URL:
- Local manifest:
