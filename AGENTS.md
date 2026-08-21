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
- `cmd/heravision/` — CLI entry (extract, compare, mcp, setup, bench, doctor, version)
- `internal/processor/` — decode (+limit anti-bomb), EXIF auto-rotate, preprocess, resize
- `internal/detector/` — Sobel+Canny hysteresis, components, classify v2, box color
- `internal/color/` — Lab k-means dominant + background
- `internal/ocr/` — engine interface + heuristic placeholder (OCR real = Phase 5)
- `internal/layout/` — header/body/footer tree
- `internal/diagram/` — mermaid chain graph
- `internal/facts/` — shared pipeline CLI+MCP (extract, compare, markdown)
- `internal/config/` — heravision.json loader
- `internal/buildinfo/` — single source version
- `mcp/` — MCP server stdio 3 tools
- `plugins/` — opencode/claude/codex configs
- `install.sh` / `install.ps1` / `npm/` — one-line installers; bare `setup` = interactive wizard (OCR → agent config)
- `plan.md` — single source of truth
- `testdata/` — fixtures png

## Workflow
- Sequential dev-team: architect → senior-dev → qa-engineer
- Phase 0: scaffolding → Phase 1: core native → Phase 2: OSS polish
- Verification: `go test ./...`, `golangci-lint`, `heravision doctor`, E2E MCP call di Opencode
- Semua log ke stderr, stdout hanya JSON-RPC

## Rules
- Binary <12MB core; OCR bundle boleh >12MB setelah Fase D sukses (keputusan 2c)
- RAM <80MB, latency <300ms (keputusan 3b)
- Pure Go tanpa CGO, tanpa API key
- Update plan.md jika scope berubah
- Heravision extract return JSON+markdown — jangan breaking change tanpa bump
- Dokumentasi hanya klaim yang terukur — output contoh = output asli (audit 2026-08-21)

## Child DOX Index
- (none yet — new durable boundaries create child AGENTS.md)

## Verification
- Sebelum close task: cek DOX chain, update AGENTS.md jika struktur berubah
