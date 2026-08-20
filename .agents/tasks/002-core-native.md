# Task 002 — Core Native Pipeline

**Phase:** 1 (MVP)
**Assignee:** senior-dev
**Status:** pending
**Depends:** 001

## Goal
Detector + color + layout + OCR WASM → JSON real

## Sub-tasks
- [ ] `internal/detector/edge.go` — Canny 50/150 pure Go
- [ ] `internal/detector/contour.go` — findContours
- [ ] `internal/detector/classify.go` — button/input/card/image by aspect+area
- [ ] `internal/color/histogram.go` — 5 dominant hex
- [ ] `internal/layout/tree.go` — cluster → tree
- [ ] `internal/ocr/ocr.go` — interface + WASM tiny (fallback empty)
- [ ] `internal/prompt/modes.go` — 5 mode prompts
- [ ] MCP 3 tools return real JSON+markdown
- [ ] E2E test DeepSeek V4 via Opencode MCP

## DoD
- `heravision extract testdata/ui.png --json` <200ms, JSON valid
- Opencode tool call success
