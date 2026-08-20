# Task 003 — OSS Polish

**Phase:** 2
**Assignee:** senior-dev → qa-engineer
**Status:** pending
**Depends:** 002

## Goal
Compare, setup, goreleaser, docs

## Sub-tasks
- [ ] `heravision_compare` diff logic
- [ ] `heravision setup --all` (opencode/claude/codex/cursor)
- [ ] `heravision doctor` + `heravision bench`
- [ ] README + DEMO.gif + architecture diagram
- [ ] goreleaser.yaml + npm wrapper
- [ ] GitHub Action CI (test, lint, build)
- [ ] LICENSE MIT, CONTRIBUTING.md

## DoD
- `goreleaser build --snapshot` produce win/mac/linux
- `go test ./...` + `golangci-lint` pass
- README install verified
