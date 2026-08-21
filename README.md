<h1 align="center">HeraVision</h1>
<p align="center"><strong>UI structure extractor for text-only LLMs.</strong><br>
Pure Go · offline · <code>CGO_ENABLED=0</code> · MCP-native</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.25"/>
  <img src="https://img.shields.io/badge/binary-%7E14MB-111827?style=flat-square" alt="binary ~14MB"/>
  <img src="https://img.shields.io/badge/CGO-disabled-success?style=flat-square" alt="CGO disabled"/>
  <img src="https://img.shields.io/badge/license-MIT-6B7280?style=flat-square" alt="MIT"/>
</p>

---

Text-only coding models (DeepSeek, GLM, and similar) cannot accept screenshots. When a user pastes a UI image and asks "fix this layout", the model has nothing to work with. Cloud vision APIs solve this with API keys, quotas, and network round-trips; local vision-language models need 1–2 GB of RAM and a model download.

HeraVision takes a different position: an LLM does not need to *see* a screenshot — it needs **measured facts about it**. HeraVision computes those facts deterministically in a single ~14 MB static binary:

- element boxes (`button`, `input`, `card`, `image`, `text_block`, `icon`, `checkbox`, `avatar`) with pixel coordinates, average color, confidence score, reading order, and a plain-language caption ("blue button 101x32 large at center");
- page-type guess (`login`, `dashboard`, `terminal`, `chat`, `form`, `general`) with confidence;
- grid detection (e.g. a 3×2 card grid) and table detection (rows × columns from ruled lines);
- dominant colors in CIE L\*a\*b\* space plus a background estimate;
- a recursive XY-cut layout tree, and a Mermaid graph with real connector edges in `diagram` mode;
- optional real OCR (ONNX Runtime via purego, still zero-CGO) — without the bundle it honestly reports shape placeholders instead of pretending to read text.

The JSON and markdown land in the model context through three MCP tools, and the model does what it is good at: reasoning over structured input.

## How it works

```
 JPEG / PNG / WebP
       │
       ▼
 decode ──────────── limits enforced before allocation: ≤16384 px/side, ≤12 MP
       │
       ▼
 EXIF orientation ── manual TIFF/IFD parse of tag 0x0112, auto-rotate (all 8 cases)
       │
       ▼
 preprocess ──────── Laplacian-variance blur score < 80
       │              → contrast +15, sharpen 0.8, 2× Lanczos if the image is small
       ▼
 resize ──────────── longest side → max_side (default 1024, Lanczos)
       │
       ▼
 multi-scale pyramid   ½× · 1× · 2× (2× only when source ≥ 1.5× base)
       │               boxes mapped back to base coordinates, merged at IoU > 0.55
       ▼
 detector (per scale)  Gaussian 3×3 → Sobel → non-maximum suppression
       │               → Canny hysteresis → morphological close
       │               → 8-connected components → classify v3 → box color
       ▼
 analysis              tables (line geometry) · OCR (ONNX or placeholders)
       │               colors (Lab k-means) · grids · page type · captions · order
       ▼
 structure             XY-cut layout tree · mermaid graph (diagram mode)
       │
       ▼
 Result                { meta, page_type, texts, boxes, grids, tables,
                         colors, layout, mermaid } + markdown summary
```

Every stage runs locally, allocates against hard limits, and degrades gracefully: OCR missing → placeholders; connectors absent → hierarchy fallback; config absent → defaults.

## The math

This section documents the actual algorithms and constants in `internal/`. Thresholds are defaults; all are tunable through `heravision.json`.

### 1. Blur score — Laplacian variance (`processor/preprocess.go`)

The gray image *I* is convolved with the 4-neighbour Laplacian ΔI = I(x−1,y) + I(x+1,y) + I(x,y−1) + I(x,y+1) − 4·I(x,y). The blur score is the variance of |ΔI| over interior pixels:

```
BM = Var(|ΔI|)
```

Sharp images have high-frequency content, so their Laplacian responses vary strongly; blurred ones collapse toward zero. `BM < blur_threshold` (default **80**) triggers enhancement. A second band (`BM < 2·threshold`) gets a light sharpen only.

### 2. Edges — Sobel, NMS, hysteresis (`detector/detector.go`)

3×3 Sobel kernels produce gradients *gx*, *gy*. Magnitude and direction:

```
m = min(255, |gx| + |gy|)
θ ∈ {0°, 45°, 90°, 135°}   quantized by dominance tests (e.g. |gx| ≥ 2|gy| ⇒ horizontal)
```

**Non-maximum suppression** keeps a pixel only if *m* is a local maximum along its quantized gradient axis — edges come out exactly 1 px wide, which is what makes box coordinates precise. **Canny hysteresis** then accepts strong pixels (*m > high*, default **150**) as seeds and grows into weak pixels (*m > low*, default **50**) that touch an accepted region, iterated to a fixed point.

### 3. Closing and components (`detector/detector.go`)

Morphological closing (dilate → erode, 3×3 structuring element, 2 iterations) reconnects outlines broken by NMS. Connected regions are extracted by BFS flood fill with 8-connectivity and reduced to bounding boxes, filtered by `w ≥ 12`, `h ≥ 10`, `area ≥ max(min_area, W·H/20000)` and `area ≤ 0.9·W·H`.

### 4. Shape features and classification (`detector/detector.go`)

Each candidate box yields five features:

| Feature | Definition |
|---|---|
| `ar` | aspect ratio `w/h` |
| `area` | `w·h` |
| `ed` | edge density — edge pixels ÷ box area |
| `ring` | share of edge pixels within 2 px of the box border (outline shapes score high) |
| `cv` | mean per-channel color variance of interior samples ÷ 255² |

Classification is an ordered rule cascade — first match wins:

| Order | Type | Condition |
|---|---|---|
| 1 | `checkbox` | 12–28 px square, ar 0.7–1.4, ring > 0.45, cv < 0.02 |
| 2 | `icon` | < 40 px both sides, ed > 0.15 |
| 3 | `avatar` | square 30–140 px, cv > 0.04 |
| 4 | `input` | ar > 3.2, height 14–70, ed < 0.3 |
| 5 | `button` | ar 1.4–6, w > 50, h 20–85, area < 40000, ed 0.08–0.5 |
| 6 | `image` | ar < 1.25, 35–400 px both sides, ed < 0.4 |
| 7 | `card` | width > 50% of image, h > 50 |
| 8 | `text_block` | ed 0.02–0.25, h < 35, w > 25 |
| — | `card` | fallback |

Confidence: `score = min(1, ½·area/(W·H) + ½·ed)`. Duplicate boxes are removed greedily at IoU > 0.45, keeping the higher score.

### 5. Perceptual color — Lab k-means (`color/histogram.go`)

Up to 8000 pixels are sampled (stride ⌈√(total/8000)⌉) and converted sRGB → linear (IEC gamma) → XYZ (D65) → CIE L\*a\*b\*. Clustering is k-means with **k = 5**, 8 iterations, deterministic spread initialisation. Centers closer than **ΔE(CIE76) < 12** merge with count-weighted averages; survivors are ranked by cluster size and emitted as hex. The background is the mean color of the two topmost and two bottommost pixel rows.

### 6. Scale space (`facts/facts.go`)

Detection runs at up to three scales — `maxSide/2`, `maxSide`, and `min(2·maxSide, 2048)` (the last only when the source is at least 1.5× the base) — because a 16 px icon that vanishes at base scale survives at 2×. Boxes are rescaled into base coordinates and duplicates removed greedily:

```
IoU(A, B) = |A ∩ B| / |A ∪ B|      merge threshold 0.55
```

### 7. Layout — recursive XY-cut (`layout/tree.go`)

Boxes are projected onto Y, split wherever consecutive intervals leave a gap ≥ **12 px**, then each group is re-split along X, alternating down to depth **6**. Single-child chains collapse; groups that cannot be separated become flat `region` nodes. Leaves carry the global reading order (Y-band, then X) and caption.

### 8. Tables from line geometry (`detector/table.go`)

Run-length scans find straight edge segments ≥ `max(dim/6, 40)` px; collinear fragments merge when within 2 px in position and 14 px in gap. Horizontal segments form bands when they overlap > 0.6 within half the image height; vertical segments crossing both band edges (± 8 px tolerance) count as column separators. Then `rows = bands − 1`, `cols = crossings − 1`, and any 1×1 result is rejected — a plain rectangle is not a table.

### 9. Diagram connectivity (`diagram/graph.go`)

An edge-map component (≥ 12 px, span ≥ 30 px) whose extreme endpoints map to two different boxes (gap ≤ 90 px) is a connector; direction goes toward the less dense endpoint — a proxy for arrowhead mass. Elongated bars (ar ≥ 6, thickness ≤ 30 px) connect their axis neighbours within 140 px. With no connectors found, the graph falls back to the layout-tree hierarchy rather than inventing edges.

### 10. Page typing (`facts/analyze.go`)

Additive evidence scores compete: terminal (dark background + many text rows), login (1–3 centered inputs), dashboard (card grid ≥ 4 cells), chat (left/right bubbles + bottom input), form (≥ 3 inputs). The winner needs ≥ 0.45 confidence, otherwise the honest answer is `general` at 0.40; reported confidence caps at 0.95.

## Output example

Actual output of `heravision extract testdata/ui.png --mode ui --json` (abridged):

```json
{
  "meta":     { "width": 200, "height": 100, "mode": "ui", "version": "0.1.0", "elapsed_ms": 18 },
  "page_type": "general",
  "page_confidence": 0.4,
  "texts":    [{ "text": "[button]", "x": 50, "y": 29, "w": 101, "h": 32, "size": 32 }],
  "boxes": [{
      "type": "button", "x": 50, "y": 29, "w": 101, "h": 32,
      "color": "#0000FF", "score": 0.121, "order": 1,
      "caption": "blue button 101x32 large at center"
  }],
  "grids": [], "tables": [],
  "colors":   { "dominant": ["#FFFFFF", "#0000FF"], "background": "#FFFFFF" },
  "layout":   { "type": "root", "w": 200, "h": 100, "children": [ "…box leaf…" ] }
}
```

Without the OCR bundle, `texts` contains shape placeholders (`[button]`, `[text]`) and the markdown says so explicitly — the output never pretends to have read text it has not.

## Install

One line per platform — prebuilt binaries, no Go toolchain required:

```bash
# npm — any platform (recommended)
npm install -g heravision

# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/ahmdd4vd/heravision/main/install.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/ahmdd4vd/heravision/main/install.ps1 | iex
```

The package lives at [npmjs.com/package/heravision](https://www.npmjs.com/package/heravision).

Every channel ends in the same interactive setup wizard:

```text
[1/2] Download real OCR engine (~28 MB: onnxruntime + PP-OCR mobile)? [y/N]: y
[2/2] Configure MCP for your AI coding agent:
  1) opencode   2) claude   3) codex   4) cursor   a) all   s) skip
choice [1/2/3/4/a/s]: a

=== setup complete ===
```

Answering `y` downloads the OCR sidecar to `~/.heravision/ocr` **and enables it automatically** in `~/heravision.json`. Answering `a` writes MCP config for all four agents; a number configures just that one. Non-interactive equivalents: `heravision setup --ocr`, `heravision setup --agent opencode`, `heravision setup --all`.

Build from source instead (Go 1.25+, no cgo toolchain needed):

```bash
git clone https://github.com/ahmdd4vd/heravision.git && cd heravision
./build.ps1        # Windows
make build         # Linux / macOS
```

Verify:

```bash
heravision doctor             # environment check
heravision extract ./shot.png --mode ui --json
heravision bench --n 20
```

Manual OCR configuration (only if you skipped setup): set `ocr.enabled` plus `lib_path` / `det_path` / `rec_path` / `dict_path` in `heravision.json` — see `heravision.json.example`.

### Modes

| Mode | Behavior |
|---|---|
| `general` | balanced defaults |
| `ui` | element detection emphasis |
| `code` | editor and terminal screenshots |
| `diagram` | additionally emits a Mermaid graph with connector edges |
| `error` | error dialogs and stack traces |
| `blur` | forces contrast + sharpen preprocessing |

## MCP tools

| Tool | Input | Returns |
|---|---|---|
| `heravision_extract` | `path`, `mode` | markdown summary + full JSON |
| `heravision_compare` | `path_a`, `path_b` | structural diff: added / removed / moved / color-changed, with a one-line summary |
| `heravision_describe` | `path`, `mode` | markdown summary only |

Compare resizes both images to a 512 px long side, matches boxes at IoU > 0.5, flags movement beyond 5 px and color changes beyond RGB distance 25.

## Configuration

`heravision.json` is searched in the working directory, then `$HOME`; an absent file means defaults. Copy `heravision.json.example` as a starting point.

| Key | Default | Notes |
|---|---|---|
| `max_side` | 1024 | clamped to 64–4096 |
| `max_pixels` | 12000000 | decode guard |
| `multiscale` | true | pyramid on/off |
| `detector.canny_low` / `canny_high` | 50 / 150 | high ≤ low resets to low × 3 |
| `detector.min_area` | 200 | effective floor scales with image size |
| `preprocess.blur_threshold` | 80 | Laplacian variance |
| `color.k` / `color.deltaE_merge` | 5 / 12 | k clamped 1–16 |
| `ocr.*` | off | see Install |

## Performance

Measured on the bundled fixtures, local machine, release flags:

| Metric | Value |
|---|---|
| Structure extraction | ≈ 15–35 ms per fixture |
| OCR (when enabled) | adds ≈ 520–900 ms warm, tunable via `ocr.det_max_side_len` / `ocr.num_threads` (measured on the blurry fixture; was ≈ 1.2 s before the parallelized pipeline) |
| Binary size | ≈ 14 MB (`-s -w`, includes purego binding) |
| Memory | bounded by the 12 MP decode limit |
| Network / API keys | none, ever |

OCR latency pass 1 (done): OCR now runs **in parallel** with the structural pipeline and exposes tuning knobs — measured ≈ 1.2 s → ≈ 0.5–0.9 s on the blurry fixture with identical output. The < 300 ms total target is not reached yet; next step is rec-only recognition per text crop.

## Honest limitations

- Table detection requires ruled lines; borderless tables are invisible to it.
- Mermaid edge direction is inferred from endpoint density, not arrowhead shape analysis; arrows fused into a shape by morphological closing cannot be recovered.
- Mobile-tier OCR reads clear UI text at 0.99+ confidence on fixtures but struggles with heavy blur; spaces occasionally surface as "?".
- Classification thresholds are tuned on synthetic and captured fixtures; unusual design languages may misclassify.

## Repository layout

```
cmd/heravision/      CLI entry (extract, compare, mcp, setup, bench, doctor, version)
internal/processor/  decode limits, EXIF rotate, blur preprocess, resize
internal/detector/   Sobel/Canny/NMS pipeline, classify v3, table geometry
internal/color/      Lab conversion, k-means, background
internal/ocr/        engine interface, heuristic placeholders, ONNX engine
internal/layout/     recursive XY-cut tree
internal/diagram/    mermaid nodes, connector discovery, hierarchy fallback
internal/facts/      shared pipeline: extract, compare, captions, grids, page type
internal/config/     heravision.json loader with normalization
internal/buildinfo/  single-source version
mcp/                 MCP stdio server (mark3labs/mcp-go)
plugins/             ready-made opencode / claude configs
install.sh / .ps1    one-line installers (curl / PowerShell)
npm/                 installer package: fetches release binary, wraps CLI
vscode-extension/    command-palette bridge to the CLI
testdata/            png fixtures used by tests and examples
```

## Verify

```bash
go vet ./... && go test ./... -count=1
./heravision doctor
./heravision bench --n 10 --mode blur
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}\n' | ./heravision mcp
```

CI runs the same suite on every push.

## Contributing & license

PRs welcome — keep the core pure Go (`CGO_ENABLED=0`), log to stderr only (stdout is JSON-RPC), and document only what you can measure. MIT — see [LICENSE](LICENSE) and [CONTRIBUTING.md](CONTRIBUTING.md).

<p align="center"><sub>HeraVision — structure first, reasoning where it belongs: in the model.</sub></p>
