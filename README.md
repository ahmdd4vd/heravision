# HeraVision — The eyes for blind LLMs

> Pure native Go vision (<10MB, no model/API) → LLM text-only jadi bisa lihat.

**Hybrid:** HeraVision ekstrak fakta mentah (boxes, colors, texts, layout) → DeepSeek V4 / GLM 5.3 reasoning → fix.

```
User "fix UI 📸" → LLM → heravision_extract → JSON {texts, boxes, colors, layout} → LLM reasoning → fix code
```

## Install

```bash
go install github.com/heravision/heravision@latest
# or npm
npm i -g heravision

heravision setup --all          # opencode + claude + codex + cursor
heravision setup --agent opencode
```

## Usage

```bash
heravision extract ./screenshot.png --mode ui --json
heravision extract ./error.png --mode error
heravision mcp                  # MCP stdio server
heravision doctor
heravision bench --n 20
heravision version
```

### Modes
`general` | `ui` | `code` | `diagram` | `error`

## MCP Tools

| Tool | Input | Output |
|------|-------|--------|
| `heravision_extract` | `path`, `mode` | JSON `{meta, texts, boxes, colors, layout, lines}` + markdown |
| `heravision_compare` | `path_a`, `path_b` | `{added, removed, moved, color_changed}` |
| `heravision_describe` | `path`, `mode` | markdown alias |

## Output Example

```json
{
  "meta": {"width":1024,"height":768,"mode":"ui","elapsed_ms":42},
  "texts": [{"text":"Login","x":450,"y":30}],
  "boxes": [{"type":"button","x":400,"y":300,"w":200,"h":40,"color":"#3B82F6"}],
  "colors": {"dominant":["#FFFFFF","#3B82F6"],"background":"#FFFFFF"},
  "layout": {"type":"root","children":[{"type":"header"},{"type":"body"}]}
}
```

## Architecture

```
heravision (Go, CGO_ENABLED=0)
├── cmd/heravision — cobra CLI
├── internal/processor — decode, EXIF, resize 1024
├── internal/detector — sobel → contours → classify
├── internal/color — histogram 64-bin → 5 dominant hex
├── internal/layout — header/body/footer tree
├── internal/ocr — interface (WASM tiny roadmap)
└── mcp — stdio server (mark3labs/mcp-go)
```

## Performance

| Metric | Target | Actual |
|--------|--------|--------|
| Binary | <12MB | 9.3MB |
| RAM | <50MB | ~15MB |
| Latency (no OCR) | <100ms | ~3ms (1024px) |
| CGO | 0 | 0 |

## Comparison

|  | HeraVision | Gemini API | Ollama moondream |
|--|-----------|------------|-----------------|
| Size | 9MB | API key | 1-2GB |
| Offline | ✅ | ❌ | ✅ |
| Cost | free | $$ | free |
| RAM | 15MB | 0 | 2GB |

## Config

See `heravision.json.example` — thresholds, max_side, ocr lang.

## Contributing

See `CONTRIBUTING.md` — `go test ./...`, `golangci-lint`.
