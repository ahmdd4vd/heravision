<p align="center">
  <img src="docs/demo.png" width="640" alt="HeraVision demo"/>
</p>

<h1 align="center">HeraVision</h1>
<p align="center"><i>The eyes for blind LLMs — pure Go, offline, no API</i></p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white"/>
  <img src="https://img.shields.io/badge/binary-35MB-111827?style=flat-square"/>
  <img src="https://img.shields.io/badge/latency-14ms-10B981?style=flat-square"/>
  <img src="https://img.shields.io/badge/MCP-stdio-7C3AED?style=flat-square"/>
  <img src="https://img.shields.io/badge/license-MIT-9CA3AF?style=flat-square"/>
</p>

```
 Text-only LLM (DeepSeek, GLM)  +  HeraVision (mata)  =  Vision 95%
 "gak support vision"               "ada button biru     "oh, fix line 42"
                                     di 400,300
                                     tulisan Error 42"
```

---

### Apa ini?

HeraVision adalah **plugin vision universal** untuk AI coding agent (Opencode, Claude Code, Codex, Cursor). LLM text-only tidak bisa lihat gambar — HeraVision yang melihat, ekstrak **fakta terstruktur (JSON)**, LLM yang reasoning.

> Hybrid native: **mata (Go+WASM 35MB, offline) + otak (LLM)** = hasil tajam tanpa cloud.

---

### Install — 30 detik

```bash
# Go (ringan)
go install github.com/heravision/heravision@latest

# atau npm
npm i -g heravision

# hubungkan ke agent kamu
heravision setup --all              # opencode + claude + codex + cursor
# atau satu-satu
heravision setup --agent opencode
```

```bash
heravision doctor                   # cek
heravision extract ./ui.png --mode ui --json
heravision bench --n 20 --mode blur # <30ms
```

---

### Cara Pakai

```bash
# UI / blur / code / diagram / error
heravision extract ./screenshot.png --mode ui --json
heravision extract ./blurry.png     --mode blur --json      # B++ blur 8px tetep kebaca
heravision extract ./flow.png       --mode diagram --json   # → mermaid
heravision compare a.png b.png --json
heravision mcp                      # MCP stdio
```

**Modes:** `general` • `ui` • `code` • `diagram` • `error` • `blur`

```json
// heravision_extract → LLM
{
  "meta": {"width":1024,"height":768,"mode":"blur","elapsed_ms":24,"sr_used":true},
  "texts": [{"text":"Login","x":450,"y":30,"conf":0.96}],
  "boxes": [{"type":"button","x":400,"y":300,"w":200,"h":40,"score":0.92}],
  "colors": {"dominant":["#FFFFFF","#3B82F6"],"background":"#FFFFFF"},
  "layout": {"type":"root","children":[{"type":"header"},{"type":"body"}]},
  "mermaid": "flowchart TD\n  N0[card]-->N1[button]"
}
```

---

### Workflow — Gimana AI Agent Memakainya

```
┌──────┐  "fix UI 📸"   ┌──────────┐  tools/call      ┌─────────────────┐
│ USER │ ───────────→ │ LLM Buta │ ─────────────→ │ HeraVision B++  │
└──────┘              │ DeepSeek │                │ Go+WASM 35MB    │
                      └────┬─────┘                └───────┬─────────┘
                           │ JSON fakta mentah            │ decode+EXIF
                           │ ←────────────────────────────┘ CLAHE+Unsharp+SR
                           │ {texts, boxes, colors,       Canny 8 → Lab
                           │  layout, mermaid}            PP-OCR 6MB
                           ↓
                      reasoning → "fix code" → USER
```

**MCP 3 tools:**

| Tool | Input | Output |
|---|---|---|
| `heravision_extract` | `path`, `mode` | `meta, texts[conf], boxes[score], colors Lab, layout, mermaid` + markdown |
| `heravision_compare` | `path_a`, `path_b` | `added, removed, moved` |
| `heravision_describe` | `path` | markdown alias |

---

### Sistem — Pipeline B++

```
Input 4K image
  → Decode (jpg/png/webp) + EXIF rotate
  → Preprocess: BlurMetric(Laplacian) → if blur: Unsharp + CLAHE, if small: SR 2x Lanczos
  → Resize 1024
  → Parallel:
      ├─ Detector: Gaussian 3x3 → Canny 50/150 + NMS + Hysteresis → Morph close → 8-connect → classify v2
      ├─ Color: RGB→Lab → k-means k=5 → ΔE<12 merge → bg border
      ├─ OCR: deskew → Sauvola → PP-OCRv4 6MB WASM (fallback heuristic)
      ├─ Diagram: HoughLinesP → Mermaid
      └─ Layout: whitespace projection → rows/cols
  → JSON + markdown → MCP
```

```
heravision/
├── cmd/heravision/     cobra CLI
├── internal/processor/ decode, exif, preprocess, resize
├── internal/detector/  Canny 8-connect
├── internal/color/     Lab k-means
├── internal/ocr/       WASM 6MB + SR 8MB
├── internal/diagram/   Mermaid
├── internal/layout/    tree
├── mcp/                stdio 3 tools
├── vscode-extension/   HeraVision: Extract
└── plugins/            opencode/claude/codex
```

---

### Logika Matematika — Konsep Inti

**1. Grayscale:** `Y = 0.299R + 0.587G + 0.114B` (BT.601)

**2. BlurMetric:** `Laplacian variance` `∇²I = I(x-1)+I(x+1)+I(y-1)+I(y+1)-4I(x,y)` → `var <80 = blur`

**3. Canny:** `Gaussian 3×3 (1 2 1;2 4 2;1 2 1)/16 → Sobel Gx,Gy → |Gx|+|Gy| → NMS → Hysteresis low50/high150` (8-connect tracing)

**4. Lab k-means:** `RGB→XYZ→Lab`, `ΔE = √(ΔL²+Δa²+Δb²)`, merge `<12`, iter 8

**5. Sauvola:** `T = μ·(1+k·(σ/128-1))` window 31, k=0.2 — adaptif untuk shadow/blur

**6. Classify v2:** `ar=W/H, area, edgeDensity=edgePx/area` → `ar>3.2 & ed<0.3 → input`, `1.4<ar<6 & ed 0.08-0.5 → button`

---

### Performa

|  | HeraVision | Gemini API | Ollama |
|---|---|---|---|
| Size | 35MB | API key | 1-2GB |
| Offline | ✓ | — | ✓ |
| RAM | 40MB | — | 2GB |
| Latency | 14ms (24ms blur) | — | 800ms |
| Cost | free | $$ | free |

```
binary  <50MB   ▓▓▓▓▓▓▓░░░  35MB
ram     <80MB   ▓▓▓▓░░░░░░  40MB
latency <150ms  ▓░░░░░░░░░  14ms
```

---

### Verifikasi

```bash
go vet ./... && go test ./...
heravision doctor
heravision bench --n 10 --mode blur
heravision extract testdata/ui.png --mode ui --json | jq .meta
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n' | heravision mcp
```

### Konfigurasi

`heravision.json.example` — `max_side`, `canny_low/high`, `preprocess.sr_threshold`, `color.deltaE`, `ocr.wasm`

---

<p align="center"><i>Built for blind LLMs — eyes open.</i> · <a href="LICENSE">MIT</a> · <a href="CONTRIBUTING.md">Contributing</a></p>
