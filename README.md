<h1 align="center">HeraVision</h1>
<p align="center"><i>UI structure extractor for blind LLMs — pure Go, offline, no API</i></p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white"/>
  <img src="https://img.shields.io/badge/binary-6.4MB-111827?style=flat-square"/>
  <img src="https://img.shields.io/badge/license-MIT-9CA3AF?style=flat-square"/>
</p>

```
 Text-only LLM (DeepSeek, GLM)  +  HeraVision (mata struktur)  =  paham layout UI
 "gak support vision"               "ada button biru #0000FF       "oh, tombolnya di
                                     di (98,158) 244x84"            situ, fix position"
```

---

### Apa ini?

HeraVision adalah **plugin vision struktural** untuk AI coding agent (Opencode, Claude Code, Codex, Cursor) via MCP. LLM text-only tidak bisa lihat gambar — HeraVision melihat **struktur layar**: posisi & ukuran elemen UI, warna dominan, peta layout — lalu mengirimnya sebagai JSON terstruktur untuk direasoning oleh LLM.

**Yang DIDETEKSI:** element boxes (`button` / `input` / `card` / `image` / `text_block`) dengan koordinat, ukuran, skor, dan warna rata-rata; palet warna Lab k-means; background; layout header/body/footer; graph mermaid sederhana.

**Yang BELUM:** OCR. Teks belum dibaca — field `texts` berisi placeholder bentuk seperti `[button]`, `[text]`. Ini roadmap utama (lihat bawah). Jangan pakai HeraVision untuk membaca isi teks gambar.

---

### Install

```bash
# Build dari source (butuh Go 1.25+)
git clone https://github.com/heravision/heravision
cd heravision
# Windows:
./build.ps1
# Linux/macOS:
make build

# hubungkan ke agent kamu
./heravision setup --all          # opencode + claude + codex + cursor
# atau satu-satu
./heravision setup --agent opencode
```

```bash
heravision doctor                   # cek setup
heravision extract ./ui.png --mode ui --json
heravision bench --n 20
```

---

### Cara Pakai

```bash
heravision extract ./screenshot.png --mode ui --json
heravision extract ./blurry.png     --mode blur --json    # sharpen + kontras dulu
heravision extract ./flow.png       --mode diagram --json # + mermaid graph
heravision compare a.png b.png --json                     # diff struktural
heravision mcp                                            # MCP stdio server
```

**Modes:** `general` • `ui` • `code` • `diagram` • `error` • `blur`

Output nyata (`extract testdata/blurry_ui.png --mode blur`, dipotong):

```json
{
  "meta": { "width": 800, "height": 400, "mode": "blur", "elapsed_ms": 98 },
  "boxes": [
    { "type": "button", "x": 98, "y": 158, "w": 244, "h": 84,
      "color": "#2B1E28", "score": 0.17 },
    { "type": "card", "x": 46, "y": 46, "w": 96, "h": 40,
      "color": "#AFC1F5", "score": 0.29 }
  ],
  "colors": {
    "dominant": ["#FFFFFF", "#0C05FE", "#AFC1F5", "#2B1E28", "#F8B84B"],
    "background": "#FFFFFF"
  },
  "layout": { "type": "root", "children": [ { "type": "header" }, { "type": "body" } ] },
  "mermaid": "",
  "markdown": "## Image Facts (blur)\n- Size: 800x400 px\n..."
}
```

---

### MCP Tools

| Tool | Input | Output |
|---|---|---|
| `heravision_extract` | `path`, `mode` | markdown + JSON: meta, boxes (posisi/warna/skor), colors Lab, layout tree |
| `heravision_compare` | `path_a`, `path_b` | added / removed / moved / color_changed boxes |
| `heravision_describe` | `path`, `mode` | markdown ringkas saja |

---

### Sistem — Pipeline

```
Input image (jpg/png/webp)
  → Decode (dengan limit dimensi anti decompression-bomb)
  → EXIF orientation tag 0x0112 → auto-rotate/flip
  → Preprocess: BlurMetric (Laplacian variance) → sharpen/kontras jika blur
  → Resize longest side ≤ max_side (default 1024)
  → Detector: Gaussian 3x3 → Sobel+Canny hysteresis (low/high) → morph close
              → connected components 8-connect → classify v2 (ar/area/edgeDensity)
              → warna rata-rata per box
  → Color: RGB→Lab → k-means (k=5) → merge ΔE<12 → dominant + background border
  → Layout: split header/body/footer by Y
  → Mermaid chain graph (mode diagram)
  → JSON + markdown
```

**Jujur soal batasan teknis saat ini:**
- Canny tanpa Non-Maximum Suppression (edge lebih tebal dari Canny standar)
- Layout split 3 zona Y, bukan recursive whitespace projection
- Mermaid = rantai node urut scan, bukan deteksi panah Hough
- Teks = placeholder bentuk, bukan OCR

Semua angka threshold bisa di-tune via `heravision.json` (copy dari `heravision.json.example`).

---

### Performa (diukur, bukan janji)

| Metrik | Nilai |
|---|---|
| Binary | 6.4 MB (`-ldflags "-s -w"`, CGO_ENABLED=0) |
| Latency | ~15 ms (gambar kecil), ~100 ms (mode blur 800x400) |
| RAM | dibatasi decode limit 12 MP default |
| Offline | 100%, tanpa API key, tanpa download model |

---

### Roadmap

- [ ] **OCR real** — kandidat: wazero + model WASM, atau ONNX Runtime; riset kompatibilitas CGO-free sedang berjalan
- [ ] Canny NMS + gradient direction
- [ ] Deteksi panah Hough → mermaid edges akurat
- [ ] Layout recursive rows/cols + reading order
- [ ] Table detection → JSON table

---

### Verifikasi

```bash
go vet ./... && go test ./...
./heravision doctor
./heravision bench --n 10 --mode blur
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}\n' | ./heravision mcp
```

### Konfigurasi

`heravision.json.example` — `max_side`, `max_pixels`, `detector.canny_low/high/min_area`, `preprocess.blur_threshold`, `color.k/deltaE_merge`. Letakkan di folder project atau `$HOME`.

---

<p align="center"><i>Built for blind LLMs — structure first, text later.</i> · <a href="LICENSE">MIT</a> · <a href="CONTRIBUTING.md">Contributing</a></p>
