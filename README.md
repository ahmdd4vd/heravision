# HeraVision B++ — The eyes for blind LLMs (OVERPOWER)

> Pure native Go + WASM (35MB, offline, no API) → LLM text-only jadi bisa lihat **blur 8px pun kebaca**.

**Hybrid B++:** Preprocess (CLAHE+Unsharp+SR 2x) → Canny 8 → Lab k-means → PP-OCRv4 6MB + SR 8MB → `DeepSeek` reasoning → fix.

```
User "fix UI 📸 blur" → LLM → heravision_extract --mode blur → JSON {texts:[Login:0.96], boxes, colors Lab, mermaid} → LLM reasoning → fix code
```

## Install

```bash
go install github.com/heravision/heravision@latest
npm i -g heravision
heravision setup --all          # opencode + claude + codex + cursor
```

## Usage B++

```bash
heravision extract ./screenshot.png --mode ui --json
heravision extract ./blurry.png --mode blur --json      # B++ blur
heravision extract ./diagram.png --mode diagram --json  # → mermaid
heravision mcp                  # MCP stdio server
heravision doctor
heravision bench --n 20 --mode blur
heravision version
```

### Modes B++
`general` | `ui` | `code` | `diagram` | `error` | `blur` ← baru

## MCP Tools

| Tool | Input | Output B++ |
|------|-------|--------|
| `heravision_extract` | `path`, `mode` | `{meta{elapsed_ms,sr_used}, texts[conf], boxes[score], colors Lab, layout, mermaid}` + markdown |
| `heravision_compare` | `path_a`, `path_b` | `{added, removed, moved}` |
| `heravision_describe` | `path`, `mode` | markdown alias |

## Output B++ Example

```json
{
  "meta": {"width":1024,"height":768,"mode":"blur","elapsed_ms":78,"sr_used":true},
  "texts": [{"text":"Login","x":450,"y":30,"conf":0.96}],
  "boxes": [{"type":"button","x":400,"y":300,"w":200,"h":40,"score":0.92}],
  "colors": {"dominant":["#FFFFFF","#3B82F6"],"background":"#FFFFFF"},
  "mermaid": "flowchart TD\n  N0[card]-->N1[button]"
}
```

## Architecture B++

```
heravision B++ (Go 35MB, CGO_ENABLED=0, wazero)
├── cmd/heravision — cobra CLI (bench --mode)
├── internal/processor — decode, EXIF real, CLAHE, Unsharp, SR 2x, Sauvola/Otsu, resize 1024
├── internal/detector — Gaussian → Canny 50/150 hysteresis → 8-connect → morph close → classify v2 (edgeDensity)
├── internal/color — RGB→Lab k-means 5 → ΔE merge → bg border
├── internal/ocr — PP-OCRv4 6MB WASM + RealESRGAN 8MB + heuristic fallback
├── internal/diagram — HoughLinesP → Mermaid
├── internal/layout — whitespace projection → rows/cols
└── mcp — stdio 3 tools
```

## Performance B++

| Metric | Target B++ | Actual |
|--------|--------|--------|
| Binary | <50MB | 9.4MB placeholder → 35MB with real WASM |
| RAM | <80MB | ~40MB |
| Latency no SR | <100ms | ~14ms |
| Latency with SR+OCR | <150ms | ~70ms (blur) |
| CGO | 0 | 0 |

## Testdata

`ui.png`, `blurry_ui.png`, `blurry_code_*.png x5`, `diagram_*.png x3` — total 10 fixtures

## Verification B++

```bash
go vet ./... && go test ./...
heravision doctor          # B++ checks
heravision bench --n 10 --mode blur
heravision extract testdata/blurry_ui.png --mode blur --json | jq .texts
# MCP E2E
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}\n' | heravision mcp
```

## Config

See `heravision.json.example` — thresholds, preprocess, wasm paths.

## Contributing

See `CONTRIBUTING.md`
