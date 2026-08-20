# HeraVision — Plan Hybrid Native

> **Tagline:** The eyes for blind LLMs. Pure native, no model, no API, no daemon.
> **Konsep:** HeraVision (mata native <8MB) + LLM buta (otak) = Vision tajam
> **Status:** Approved — gas hybrid
> **Tanggal:** 2026-08-19
> **Owner:** David (papengcepoko@gmail.com)

---

## 1. Visi & Problem

### Problem
- DeepSeek V4, GLM 5.3, banyak LLM coding **text-only** — dikasih gambar jawab "gak support vision"
- Solusi cloud (Gemini API) butuh API key, kuota, internet, cost
- Solusi local model (Ollama moondream) butuh 1-2GB RAM + download 500MB-2GB

### Visi HeraVision Hybrid Native
- **Super ringan:** Binary <10MB, RAM 20-50MB, startup <10ms, tanpa download model
- **Pure native:** 100% algoritma Go (image processing + OCR tiny), `CGO_ENABLED=0`, single binary cross-compile
- **Tajam via kolaborasi:** HeraVision ekstrak **fakta mentah terstruktur (JSON)** → LLM buta yang reasoning → hasil tajam
- **Universal plugin:** 1 binary work di Opencode, Claude Code, Codex, Cursor via MCP stdio

### Analogi SD
- HeraVision = mata yang cuma bisa bilang "ada kotak biru di (400,300), ada tulisan 'Error line 42'"
- LLM = otak yang simpulkan "oh itu button login error, fixnya cek null"

---

## 2. Arsitektur Hybrid Native

```
┌──────┐  "fix UI 📸"  ┌──────────┐  tools/call extract  ┌──────────────────┐
│ USER │ ───────────→ │ LLM Buta │ ──────────────────→ │ HeraVision Native│
└──────┘              │(DeepSeek)│                      │  (Go <8MB)       │
                      └────┬─────┘                      └────────┬─────────┘
                           │  JSON fakta mentah                  │ pure Go:
                           │  ←─────────────────────────────────┘  - decode + EXIF
                           │  {texts:[...], boxes:[...],         - Canny edge
                           │   colors:[...], lines:[...]}        - contour find
                           ↓                                     - color histogram
                      reasoning LLM:                              - OCR tiny (PP-OCR 3MB WASM)
                      "Ini login page,                            - layout tree
                       error line 42..."
                           ↓
                      "Ini fix code-nya" → USER
```

### Layer
```
heravision (Go binary)
├── cmd/heravision/        # CLI: describe, extract, compare, mcp, setup
├── internal/
│   ├── processor/         # decode, EXIF fix, resize 1024, JPEG 80%
│   ├── detector/          # edge (Canny) → contours → boxes classification
│   ├── color/             # k-means / histogram → dominant hex
│   ├── ocr/               # PP-OCR tiny WASM or gosseract tiny — pure Go path
│   ├── layout/            # segmentasi → tree (header/card/button/input)
│   └── prompt/            # template JSON → markdown untuk LLM (opsional)
├── mcp/                   # MCP server stdio — 3 tools
├── plugins/               # opencode.json, claude.json, codex.json
└── configs/               # heravision.json (thresholds, ocr lang)
```

### Kenapa Tidak Pakai Model AI di HeraVision?
- Model = berat (500MB-2GB), butuh RAM besar, download lama
- Hybrid = HeraVision fokus **ekstrak fakta** (yang bisa algoritma murni), LLM fokus **reasoning** (yang LLM sudah jago)
- Hasil tetap tajam karena LLM modern reasoningnya kuat kalau dikasih fakta terstruktur

---

## 3. Bahasa & Stack

| Pilihan | Keputusan |
|---------|-----------|
| **Bahasa core** | **Go 1.25** — `CGO_ENABLED=0`, single binary, cross-compile `GOOS=windows/darwin/linux` |
| **Kenapa Go** | `image` stdlib + `disintegration/imaging` pure Go, startup 5ms, distribusi `go install`/`npm`/`brew`, MCP SDK `mark3labs/mcp-go` mature |
| **Alternatif** | Rust + `imageproc` + `candle` lebih kencang 2x tapi build Windows ribet, dev velocity rendah — Go menang untuk OSS tool dev |
| **MCP SDK** | `github.com/mark3labs/mcp-go` — stdio transport, 900+ star |
| **Image lib** | `disintegration/imaging` + `golang.org/x/image` + `image/jpeg/png/webp` stdlib |
| **OCR** | **Opsi 1 (primary):** `gopaddle-ocr` WASM / `GO-OCR` tiny 3MB — pure Go/WASM tanpa CGO. **Opsi 2 fallback:** `otiai10/gosseract` jika user install tesseract (CGO) — detect otomatis |
| **Edge/Contour** | Implementasi pure Go: Gaussian blur → Sobel/Canny → findContours (port dari OpenCV, pure Go) atau `github.com/disintegration/gift` + custom |
| **Color** | Pure Go k-means 5 cluster atau histogram 64-bin → hex |
| **CLI** | `spf13/cobra` |
| **Distribusi** | `go install` + `npm i -g heravision` (wrapper download binary OS) + `goreleaser` (win/mac/linux/arm64) |

---

## 4. MCP Tools — Spesifikasi Hybrid

Semua tools return **JSON terstruktur + markdown ringkas** — LLM pilih mau pakai JSON atau markdown.

### Tool 1: `heravision_extract`
```json
{
  "name": "heravision_extract",
  "description": "Extract structured facts from image — texts, boxes, colors, layout. Use for any image when you are text-only and need to see.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {"type": "string", "description": "Absolute or relative path to image (png/jpg/webp)"},
      "mode": {"type": "string", "enum": ["general","ui","code","diagram","error"], "default": "general"}
    },
    "required": ["path"]
  }
}
```
**Output:**
```json
{
  "meta": {"width":1024,"height":768,"mode":"ui","elapsed_ms":42},
  "texts": [{"text":"Login","x":450,"y":30,"w":120,"h":28,"size":24},{"text":"Error: NullPointer line 42","x":20,"y":500,"w":400,"h":20}],
  "boxes": [{"type":"input","x":20,"y":100,"w":984,"h":48},{"type":"button","x":400,"y":300,"w":200,"h":40,"color":"#3B82F6","text":"Submit"}],
  "colors": {"dominant":["#FFFFFF","#3B82F6","#1F2937"],"background":"#FFFFFF"},
  "layout": {"type":"column","children":[{"type":"header","y":0,"h":80},{"type":"card","y":100,"h":400}]},
  "lines": [],
  "markdown": "## Image Facts\n- Size: 1024x768\n- Texts: Login (450,30), Error NullPointer line 42 (20,500)\n- Boxes: 2 inputs, 1 button #3B82F6 Submit\n..."
}
```

### Tool 2: `heravision_compare`
```json
{
  "name": "heravision_compare",
  "description": "Compare two images and return diff facts — added/removed/moved boxes and texts.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path_a": {"type":"string"},
      "path_b": {"type":"string"}
    },
    "required": ["path_a","path_b"]
  }
}
```
Output: `{diff: {added:[...], removed:[...], moved:[...], color_changed:[...]}}`

### Tool 3: `heravision_describe` (compat alias)
Alias ke `extract` tapi return markdown only untuk LLM yang mau langsung baca tanpa parse JSON.

---

## 5. Pipeline Internal Detail

```
Input: path/to/image.png (4K, 8MB)
  ↓
1. processor.Decode → image.Image (handle jpg/png/webp, EXIF orientation fix)
  ↓
2. processor.Resize → longest side 1024 (Lanczos), JPEG 80% in-memory (hemat token & speed)
  ↓
3. Parallel:
   ├── detector.Edges → Canny(50,150) → findContours → classify (aspect ratio, area → button/input/card/image)
   ├── color.Extract → histogram → 5 dominant hex
   ├── ocr.Extract → texts + bbox per word
   └── layout.Build → cluster boxes → tree (header/body/footer, row/column)
  ↓
4. Merge → JSON + markdown
  ↓
5. Return via MCP CallToolResult (content: [{type:"text", text: json_string}])
```

**Performance target:** <100ms untuk 1024px image di CPU laptop (tanpa OCR), <500ms dengan OCR tiny.

**Tanpa OCR fallback:** Jika OCR WASM gagal load, tetap return boxes+colors — LLM masih bisa reasoning layout.

---

## 6. Struktur Project

```
C:\Users\david\heravision\
├── plan.md
├── AGENTS.md
├── README.md
├── go.mod (module heravision)
├── go.sum
├── Makefile
├── .goreleaser.yaml
├── heravision.json.example
├── cmd/
│   └── heravision/
│       └── main.go
├── internal/
│   ├── processor/
│   │   ├── decode.go
│   │   ├── resize.go
│   │   └── exif.go
│   ├── detector/
│   │   ├── edge.go
│   │   ├── contour.go
│   │   └── classify.go
│   ├── color/
│   │   └── histogram.go
│   ├── ocr/
│   │   ├── ocr.go (interface)
│   │   ├── wasm.go (primary)
│   │   └── tesseract.go (fallback CGO)
│   ├── layout/
│   │   └── tree.go
│   └── mcp/
│       └── server.go
├── mcp/
│   └── server.go (mcp-go stdio, 3 tools)
├── plugins/
│   ├── opencode.json
│   ├── claude.json
│   └── codex.json
├── scripts/
│   ├── install.sh
│   └── setup.go (heravision setup --agent)
└── testdata/
    ├── ui.png
    ├── code_error.png
    └── diagram.png
```

---

## 7. Distribusi & Instalasi Universal

```bash
# Install
go install github.com/heravision/heravision@latest
# atau
npm i -g heravision
# atau
brew install heravision

# Setup MCP untuk semua agent
heravision setup --all
# atau manual:
heravision setup --agent opencode   # tulis ke C:\Users\david\.config\opencode\opencode.json
heravision setup --agent claude
heravision setup --agent codex
heravision setup --agent cursor

# Test
heravision extract ./screenshot.png --mode ui --json
heravision mcp  # run MCP stdio server

# Verifikasi MCP
heravision doctor  # cek binary, test extract, cek MCP config
```

**MCP config yang ditulis:**
```json
// opencode.json
{"mcp": {"heravision": {"type":"local","command":["heravision","mcp"],"enabled":true}}}
// claude.json
{"mcpServers": {"heravision": {"command":"heravision","args":["mcp"]}}}
// cursor mcp.json sama
```

---

## 8. Roadmap Phase

### Phase 0 — Scaffolding (Hari 1)
- [ ] `go mod init heravision`, `cobra` CLI skeleton: `heravision extract`, `heravision mcp`, `heravision doctor`, `heravision setup`
- [ ] `processor` — decode + EXIF + resize 1024
- [ ] `mcp/server.go` — stdio hello world, 1 tool `heravision_extract` return dummy JSON
- [ ] Test manual: `heravision extract testdata/ui.png`

### Phase 1 — Core Native (Hari 2-5) — MVP
- [ ] `detector` — Canny + contours + classify (button/input/card)
- [ ] `color` — histogram dominant
- [ ] `layout` — tree builder
- [ ] `ocr` — integrate PP-OCR WASM tiny (fallback: return texts kosong jika belum ready)
- [ ] MCP 3 tools full, return JSON+markdown real
- [ ] Plugin `opencode.json` + test di DeepSeek V4 (text-only) — buktikan LLM bisa fix UI dari JSON
- [ ] `heravision.json` config, `--mode` 5 variant

### Phase 2 — Polish OSS (Minggu 2)
- [ ] `heravision_compare` diff
- [ ] Benchmark `heravision bench` — latency, akurasi vs cloud vision (opsional)
- [ ] Docs: README dengan demo GIF, arsitektur diagram, comparison table
- [ ] `goreleaser` cross-compile win/mac/linux arm64, `npm` wrapper
- [ ] GitHub Action CI (lint, test, build)
- [ ] MIT LICENSE, CONTRIBUTING.md
- [ ] `heravision setup --all` auto-detect agent config

### Phase 3 — Future (Opsional, tidak di MVP)
- [ ] OCR WASM bundling lebih tajam (ganti PP-OCR dengan model 3MB ter-quantize)
- [ ] Shape detection untuk diagram → Mermaid
- [ ] VSCode extension wrapper

---

## 9. Kriteria Selesai (Definition of Done)

- [ ] Binary <12MB (Windows .exe), `heravision --version` jalan
- [ ] `heravision extract testdata/ui.png --json` return JSON valid <200ms (tanpa OCR) / <600ms (dengan OCR)
- [ ] MCP `tools/list` muncul 3 tools, `tools/call heravision_extract` return JSON
- [ ] Opencode + Claude Code bisa panggil tool dan DeepSeek V4 berhasil reasoning dari JSON (test E2E)
- [ ] `heravision setup --agent opencode` tulis config tanpa rusak existing `opencode.json`
- [ ] `go test ./...` pass, `golangci-lint` pass
- [ ] README + demo GIF + install instruction

---

## 10. Risiko & Mitigasi

| Risiko | Mitigasi |
|--------|----------|
| OCR pure Go akurasi rendah | Pakai PP-OCR WASM 3MB tiny (pre-trained), fallback ke tesseract jika user punya; tanpa OCR tetap berguna (boxes+colors) |
| Canny/contour false positive | Threshold tunable via `heravision.json`, mode `ui` vs `general` beda threshold |
| Windows path spasi | Selalu `LiteralPath`, test di `C:\Users\david\heravision\testdata\file with space.png` |
| MCP stdout tercemar log | Semua log ke `stderr`, `stdout` hanya JSON-RPC |

---

## 11. Open Source Strategy

- **License:** MIT
- **Nama:** `heravision` — cek `github.com/heravision/heravision` + `npm heravision` availability sebelum publish (fallback `@heravision/cli`)
- **Repo:** `github.com/heravision/heravision` (atau `github.com/<user>/heravision`)
- **BYO nothing:** Tanpa API key, tanpa server, tanpa download — benar-benar native
- **Contributing:** Good first issue: tambah `mode: mobile`, improve classifier

---

## 12. Keputusan Final

- **Bahasa:** Go
- **Arsitektur:** Hybrid Native — HeraVision ekstrak fakta, LLM reasoning
- **OCR:** WASM tiny primary, fallback no-OCR
- **Distribusi:** Go binary + MCP stdio universal

**Next:** Eksekusi Phase 0 via dev-team sequential: architect → senior-dev → qa-engineer
