<h1 align="center">HeraVision</h1>
<p align="center"><strong>UI structure extractor for text-only LLMs.</strong><br>
Pure Go · offline · <code>CGO_ENABLED=0</code> · MCP-native</p>

---

Text-only coding models (DeepSeek, GLM, and similar) cannot accept screenshots. HeraVision gives them eyes: it computes **measured facts about a UI screenshot** in a single static binary — element boxes with pixel coordinates, colors, layout tree, page-type guess, optional real OCR — and hands them to your agent through three MCP tools.

## Install

```bash
npm install -g heravision
```

The postinstall script downloads the prebuilt binary for your platform from [GitHub Releases](https://github.com/ahmdd4vd/heravision/releases) (no compilation), then you run the interactive wizard:

```
heravision setup
[1/2] Download real OCR engine (~30 MB)? [y/N]: y
[2/2] Configure MCP for your AI coding agent:
  1) opencode   2) claude   3) codex   4) cursor   a) all   s) skip
=== setup complete ===
```

Prefer no npm? Same result via:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/ahmdd4vd/heravision/main/install.sh | bash
# Windows (PowerShell)
irm https://raw.githubusercontent.com/ahmdd4vd/heravision/main/install.ps1 | iex
```

## What it detects

- **Element boxes** — `button`, `input`, `card`, `image`, `text_block`, `icon`, `checkbox`, `avatar` with position, size, average color, confidence, reading order, and captions ("blue button 101x32 large at center")
- **Page type** — login / dashboard / terminal / chat / form / general, with confidence
- **Grids & tables** — card grid columns×rows; ruled-table rows×columns
- **Colors** — dominant palette in CIE Lab (k-means) + background estimate
- **Layout tree** — recursive XY-cut structure of the screen
- **Mermaid graph** — real connector edges between shapes (`diagram` mode)
- **OCR (optional)** — ONNX Runtime via purego, PP-OCR mobile; without the bundle it honestly reports shape placeholders

## MCP tools

| Tool | Input | Returns |
|---|---|---|
| `heravision_extract` | `path`, `mode` | markdown summary + full JSON |
| `heravision_compare` | `path_a`, `path_b` | structural diff + one-line summary |
| `heravision_describe` | `path`, `mode` | markdown summary only |

## Usage

```bash
heravision extract ./screenshot.png --mode ui --json
heravision compare before.png after.png --json
heravision doctor
heravision bench --n 20
```

Modes: `general` · `ui` · `code` · `diagram` · `error` · `blur`

## Performance (measured)

| Metric | Value |
|---|---|
| Structure extraction | ≈ 15–35 ms per fixture |
| OCR (optional) | adds ≈ 0.5–0.9 s warm, tunable |
| Binary | ≈ 14 MB, single file, zero dependencies at runtime |

## Links

- Full documentation & the math behind it: [github.com/ahmdd4vd/heravision](https://github.com/ahmdd4vd/heravision)
- Releases: [github.com/ahmdd4vd/heravision/releases](https://github.com/ahmdd4vd/heravision/releases)
- Issues: [github.com/ahmdd4vd/heravision/issues](https://github.com/ahmdd4vd/heravision/issues)

MIT © David
