# Task 001 — Scaffolding

**Phase:** 0
**Assignee:** architect → senior-dev
**Status:** pending
**Depends:** plan.md

## Goal
CLI skeleton + processor + MCP hello world

## Sub-tasks
- [ ] `go mod init heravision` + deps (cobra, mcp-go, imaging)
- [ ] `cmd/heravision/main.go` — cobra: `extract`, `mcp`, `doctor`, `setup`, `version`
- [ ] `internal/processor/decode.go` — jpg/png/webp + EXIF fix
- [ ] `internal/processor/resize.go` — longest 1024 Lanczos
- [ ] `mcp/server.go` — stdio, tool `heravision_extract` return dummy JSON
- [ ] `testdata/ui.png` fixture 1
- [ ] Verify: `go run ./cmd/heravision extract testdata/ui.png --json` + `go run ./cmd/heravision mcp` tools/list

## DoD
- Binary build `go build -o heravision ./cmd/heravision`
- `heravision doctor` green
