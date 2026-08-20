# HeraVision — Plan Hybrid Native B++ Overpower

> **Tagline:** The eyes for blind LLMs. Pure native, no API, no daemon — now OVERPOWER 35MB.
> **Konsep:** HeraVision B++ (mata super 35MB) + LLM buta (otak) = Vision 95% realistis blur tetep kebaca
> **Status:** Approved — gas B++ (20-50MB allowed, makin tajam makin overpower wkwk)
> **Tanggal:** 2026-08-19 → Update B++ 2026-08-20
> **Owner:** David (papengcepoko@gmail.com)

---

## 1. Visi & Problem

### Problem
- DeepSeek V4, GLM 5.3, banyak LLM coding **text-only** — dikasih gambar jawab "gak support vision"
- Solusi cloud (Gemini API) butuh API key, kuota, internet, cost
- Solusi local model (Ollama moondream) butuh 1-2GB RAM + download 500MB-2GB

### Visi HeraVision Hybrid Native B++
- **Overpower tapi tetep native:** Binary 22-35MB (was 9MB), RAM 40-60MB (was 15MB), startup <15ms, offline 100%, `CGO_ENABLED=0`
- **Batas baru:** 20-50MB GPP — tuker 75% → 95% akurat text buram 8px — worth!
- **Pure Go/WASM:** 100% Go + WASM embed (wazero), single binary cross-compile, tanpa download model eksternal
- **Tajam via kolaborasi:** HeraVision ekstrak **fakta mentah terstruktur (JSON) + mermaid** → LLM buta yang reasoning → hasil tajam
- **Universal plugin:** 1 binary work di Opencode, Claude Code, Codex, Cursor via MCP stdio

### Analogi SD — Before vs After B++
- **Before (9MB):** Mata minus 3 tanpa kacamata — liat blur → tebak `[text]` doang
- **After B++ (35MB):** Mata pakai kacamata super + mikroskop + otak AI kecil — blur 8px di-tajemin 2x dulu baru baca `Login` beneran
- HeraVision = mata yang bilang "ada button biru #3B82F6 di (400,300), ada tulisan 'Error line 42' size 14"
- LLM = otak yang simpulkan "oh itu button login error, fixnya cek null"

---

## 2. Arsitektur Hybrid Native B++

```
┌──────┐  "fix UI 📸 blur"  ┌──────────┐  tools/call extract  ┌──────────────────────┐
│ USER │ ───────────────→ │ LLM Buta │ ──────────────────→ │ HeraVision B++ Native│
└──────┘                  │(DeepSeek)│                      │  (Go 35MB)           │
                          └────┬─────┘                      └────────┬─────────────┘
                               │  JSON fakta mentah                  │ Go+WASM:
                               │  ←─────────────────────────────────┘  - EXIF real
                               │  {texts:[Login:95%],                - CLAHE+Unsharp+SR 2x
                               │   boxes:[...],                      - Canny hysteresis
                               │   colors:Lab k-means,               - 8-connect + Hough
                               │   mermaid: flowchart}               - PP-OCRv4 6MB WASM
                               ↓                                     - RealESRGAN 8MB
                          reasoning LLM:                              - Lab color + layout
                          "Ini login page blur,
                           error line 42..."
                               ↓
                          "Ini fix code-nya" → USER
```

### Layer B++
```
heravision (Go binary 35MB)
├── cmd/heravision/        # CLI: extract, compare, mcp, setup, bench, doctor, version
├── internal/
│   ├── processor/         # decode, EXIF real rotate, resize 1024, JPEG 80%, preprocess (CLAHE, Unsharp, SR 2x, Sauvola/Otsu)
│   ├── detector/          # Gaussian 3x3 → Canny 50/150 hysteresis → morph close → 8-connect → classify v2 (ar+edgeDensity+colorVar)
│   ├── color/             # Lab k-means 5 + ΔE merge → dominant hex + bg/fg split
│   ├── ocr/               # PP-OCRv4 6MB WASM (wazero) + RealESRGAN 8MB SR + heuristic fallback — pure Go path
│   ├── diagram/           # HoughLinesP + arrow head → graph → Mermaid flowchart
│   ├── layout/            # whitespace projection → rows/cols recursive → reading order
│   └── prompt/            # template JSON → markdown untuk LLM (opsional)
├── mcp/                   # MCP server stdio — 3 tools (extract/compare/describe) + mermaid
├── vscode-extension/      # VSCode: HeraVision: Extract
├── plugins/               # opencode.json, claude.json, codex.json
└── configs/               # heravision.json (thresholds, ocr lang, sr enable)
```

### Kenapa B++ Pakai Model Kecil (Tapi Tetap Native)?
- Dulu `<12MB` = nggak boleh model → heuristic 75% mentok
- Sekarang `20-50MB GPP` → bundle `PP-OCR 6MB + SR 8MB` WASM embed (bukan download) — tetap offline, tetap Go, tapi akurat 95% blur — worth wkwkw
- Hybrid tetep: HeraVision fokus **ekstrak fakta** (dengan AI kecil), LLM fokus **reasoning**

---

## 3. Bahasa & Stack B++

| Pilihan | Keputusan B++ |
|---------|---------------|
| **Bahasa core** | **Go 1.25** — `CGO_ENABLED=0`, single binary, cross-compile `GOOS=windows/darwin/linux` |
| **MCP SDK** | `github.com/mark3labs/mcp-go` — stdio transport |
| **Image lib** | `disintegration/imaging` + `golang.org/x/image` + `image/jpeg/png/webp` + `wazero` WASM runtime |
| **Preprocess** | `CLAHE clip 2.0 + Unsharp r1.0 amt0.6 + Lanczos 2x + Sauvola w31 k0.2 / Otsu` pure Go |
| **Edge/Contour** | Gaussian 3x3 → Canny 50/150 + NMS + Hysteresis → Morph close 3x3 → 8-connect → `edgeDensity` classify v2 |
| **Color** | `RGB→Lab` + `k-means 5 iter8` + `ΔE<12` merge → hex + bg dari border |
| **OCR** | **Primary:** `PP-OCRv4 6MB WASM` via wazero (quantized, 95%) + **SR:** `RealESRGAN tiny 8MB WASM` untuk text <16px blur → **Fallback:** heuristic template 36 char |
| **Diagram** | `HoughLinesP` + triangle arrow head → graph → Mermaid |
| **Layout** | Whitespace projection gap>40px → recursive row/col |
| **CLI** | `spf13/cobra` |
| **Distribusi** | `go install` + `npm i -g heravision` + `goreleaser` (win/mac/linux/arm64) — embed.gz WASM |

---

## 4. MCP Tools — Spesifikasi Hybrid B++

Semua tools return **JSON terstruktur + markdown + mermaid** — LLM pilih mau pakai JSON atau markdown.

### Tool 1: `heravision_extract` (B++: tambah mermaid + confidence)
```json
{
  "name": "heravision_extract",
  "inputSchema": {
    "properties": {
      "path": {"type": "string"},
      "mode": {"type": "string", "enum": ["general","ui","code","diagram","error","blur"], "default": "general"}
    },
    "required": ["path"]
  }
}
```
**Output B++:**
```json
{
  "meta": {"width":1024,"height":768,"mode":"ui","elapsed_ms":78,"sr_used":true},
  "texts": [{"text":"Login","x":450,"y":30,"w":120,"h":28,"size":24,"conf":0.96},{"text":"Error: NullPointer line 42","x":20,"y":500,"w":400,"h":20,"conf":0.93}],
  "boxes": [{"type":"button","x":400,"y":300,"w":200,"h":40,"color":"#3B82F6","score":0.92}],
  "colors": {"dominant":["#FFFFFF","#3B82F6","#1F2937"],"background":"#FFFFFF"},
  "layout": {"type":"root","children":[{"type":"header","children":[...]},{"type":"body"}]},
  "mermaid": "flowchart TD\n  N0[card Login]-->N1[button Submit]",
  "markdown": "## Image Facts\n- Size: 1024x768\n- Texts: Login 0.96, Error 0.93\n..."
}
```

### Tool 2: `heravision_compare` — sama, tambah `text_changed`
### Tool 3: `heravision_describe` — alias markdown only

---

## 5. Pipeline Internal Detail B++

```
Input: path/to/image.png (4K, 8MB, blur 8px, miring 5°)
  ↓
1. processor.Decode → image.Image (jpg/png/webp, truncated recovery)
  ↓
2. processor.FixOrientation → baca EXIF 0x0112 → rotate real
  ↓
3. processor.Preprocess (B++ baru):
   ├── blurMetric = Laplacian variance → jika <100 → blur
   ├── if blur or mode==blur: UnsharpMask + CLAHE
   ├── if text H<16px: RealESRGAN 8MB WASM 2x upscale crop
   └── Sauvola w31 k0.2 vs Otsu auto (pilih variance terkecil)
  ↓
4. processor.Resize → longest 1024 (jika SR: 512→1024)
  ↓
5. Parallel B++:
   ├── detector: Gaussian 3x3 → Canny(50,150)+NMS+Hysteresis → Morph close → 8-connect → classify v2 (ar+area+edgeDensity+colorVar+fillRatio → button/input/card/image/text_block)
   ├── color: RGB→Lab → k-means 5 iter8 → ΔE merge → 5 dominant hex + bg border
   ├── ocr: per text_block box → deskew Hough angle → 2x upscale → Sauvola → PP-OCRv4 6MB WASM → texts+conf (fallback heuristic)
   ├── diagram: HoughLinesP → arrow head triangle → graph
   └── layout: whitespace projection gap>40px → recursive row/col → reading order Y cluster → X
  ↓
6. Merge → JSON + markdown + mermaid
  ↓
7. Return via MCP CallToolResult
```

**Performance B++:** <80ms tanpa SR, <150ms dengan SR+OCR (was <100ms tanpa OCR) — masih <200ms, binary 35MB <50MB.

**Fallback:** Jika WASM gagal load, heuristic `[input]/[text]` + boxes+colors tetap — LLM masih bisa reasoning.

---

## 6. Struktur Project B++

```
C:\Users\david\heravision\
├── plan.md
├── AGENTS.md
├── README.md
├── go.mod (module heravision + wazero)
├── go.sum
├── Makefile
├── .goreleaser.yaml (embed.gz WASM)
├── heravision.json.example
├── cmd/heravision/main.go
├── internal/
│   ├── processor/ (decode, resize, exif real, preprocess: clahe, unsharp, sr, sauvola)
│   ├── detector/ (edge Canny, contour 8, classify v2)
│   ├── color/ (Lab k-means)
│   ├── ocr/ (ocr.go interface, wasm.go PP-OCR 6MB, sr.go RealESRGAN 8MB)
│   ├── diagram/ (mermaid.go Hough)
│   ├── layout/ (tree.go whitespace)
│   └── prompt/
├── mcp/server.go (3 tools + mermaid)
├── vscode-extension/ (package.json, extension.js)
├── plugins/ (opencode/claude/codex)
└── testdata/ (ui.png, blurry_*.png x20, diagram.png)
```

---

## 7. Distribusi & Instalasi Universal — sama, binary 35MB

```bash
go install github.com/heravision/heravision@latest
npm i -g heravision
heravision setup --all
heravision extract ./screenshot.png --mode blur --json  # B++ blur mode
heravision bench --n 20
```

---

## 8. Roadmap Phase B++

### Phase 0 — Scaffolding ✅ DONE (Hari 1)
- [x] cobra CLI, processor decode/resize, mcp hello world, testdata/ui.png

### Phase 1 — Core Native ✅ DONE (Hari 2-5)
- [x] detector Sobel→Canny awal, color histogram, layout tree, ocr heuristic, MCP 3 tools

### Phase 2 — Polish OSS ✅ DONE (Minggu 2)
- [x] compare diff, bench, README demo, goreleaser, npm wrapper, CI, LICENSE, setup merge fix

### Phase 3 — B++ Overpower (Minggu 3) — CURRENT
- [x] OCR wasm stub + heuristic fallback + diagram Mermaid + VSCode (done 22MB)
- [ ] **B++ Upgrade:** EXIF real rotate, preprocess CLAHE+Unsharp+SR 2x, Canny hysteresis 8-connect, Lab k-means, PP-OCR 6MB + SR 8MB embed.gz
- [ ] Test blurry_*.png x20 fixtures, bench <150ms, binary 35MB

### Phase 3++ — Future Overpower+ (Opsional)
- [ ] DocLayout 5MB WASM, angle classifier, table detection → JSON table
- [ ] Video frame extract

---

## 9. Kriteria Selesai B++

- [ ] Binary 22-35MB (was <12MB), `heravision --version` jalan
- [ ] `heravision extract testdata/blurry_ui.png --mode blur --json` → texts real `Login` conf>0.9, <150ms
- [ ] `heravision extract --mode diagram --json` → mermaid `flowchart TD` valid
- [ ] MCP `tools/list` 3 tools, `tools/call` return JSON + mermaid
- [ ] `heravision setup --agent opencode` merge tanpa rusak existing (fix backslash done)
- [ ] `go test ./...` pass, `go vet` pass, `heravision bench --n 20` avg <100ms (no SR) / <150ms (SR)
- [ ] README demo blur + architecture B++ + comparison table
- [ ] Binary <50MB, RAM <80MB

---

## 10. Risiko & Mitigasi B++

| Risiko | Mitigasi |
|--------|----------|
| WASM 14MB bikin binary 35MB | `embed.gz` + `goreleaser` compress, masih <50MB GPP — worth |
| SR 80ms latency | Hanya jika `blurMetric<100` atau `mode==blur` atau `H<16px`, else skip |
| Canny false positive blur | Threshold tunable `heravision.json`, mode `blur` vs `general` beda |
| Windows path spasi | `LiteralPath`, test `file with space.png` |
| MCP stdout tercemar log | Semua log `stderr`, `stdout` hanya JSON-RPC |
| OCR WASM load gagal | Fallback heuristic `[text]` + boxes+colors tetap |

---

## 11. Open Source Strategy — sama MIT, repo `heravision`, BYO nothing

---

## 12. Keputusan Final B++

- **Bahasa:** Go + WASM (wazero)
- **Arsitektur:** Hybrid Native B++ — preprocess + SR + PP-OCR 6MB + LLM reasoning
- **Size:** 35MB <50MB (was 9MB <12MB) — overpower wkwk
- **OCR:** PP-OCRv4 6MB + SR 8MB primary, heuristic fallback
- **Distribusi:** Go binary embed.gz + MCP stdio universal

**Next:** Eksekusi Phase 3++ B++ — preprocess + Canny 8 + Lab + OCR SR
