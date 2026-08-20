# AGENTS.md — HeraVision (Child DOX Contract)

**Project:** HeraVision — The eyes for blind LLMs (Hybrid Native)
**Owner:** David (papengcepoko@gmail.com)
**Root Contract:** C:\Users\david\AGENTS.md
**Status:** Active — Phase 0 approved
**Date:** 2026-08-19

## Purpose
Plugin vision universal untuk AI agent coding (Opencode, Claude Code, Codex, Cursor) yang memberi kemampuan vision ke LLM text-only (DeepSeek V4, GLM 5.3, dll) via MCP tools. Arsitektur hybrid native: HeraVision ekstrak fakta mentah (Go pure, <10MB, tanpa model/AI/API) → LLM reasoning.

## Tech Stack
- Go 1.25, CGO_ENABLED=0, cobra, mark3labs/mcp-go, disintegration/imaging
- MCP stdio, 3 tools: heravision_extract, heravision_compare, heravision_describe
- Distribusi: go install + npm wrapper + goreleaser

## Structure
- `cmd/heravision/` — CLI entry
- `internal/processor|detector|color|ocr|layout/` — core native pipeline
- `mcp/` — MCP server
- `plugins/` — opencode/claude/codex configs
- `plan.md` — single source of truth
- `testdata/` — fixtures png

## Workflow
- Sequential dev-team: architect → senior-dev → qa-engineer
- Phase 0: scaffolding → Phase 1: core native → Phase 2: OSS polish
- Verification: `go test ./...`, `golangci-lint`, `heravision doctor`, E2E MCP call di Opencode
- Semua log ke stderr, stdout hanya JSON-RPC

## Rules
- Binary <12MB, RAM <50MB, latency <100ms (tanpa OCR)
- Pure Go tanpa CGO, tanpa model download, tanpa API key
- Update plan.md jika scope berubah
- Heravision extract return JSON+markdown — jangan breaking change tanpa bump

## Child DOX Index
- (none yet — new durable boundaries create child AGENTS.md)

## Verification
- Sebelum close task: cek DOX chain, update AGENTS.md jika struktur berubah
