# HeraVision — The eyes for blind LLMs

> Pure native Go vision (<10MB, no model/API) → LLM text-only jadi bisa lihat.

**Hybrid:** HeraVision ekstrak fakta mentah (boxes, colors, texts, layout) → DeepSeek/GLM reasoning.

## Install
```
go install github.com/heravision/heravision@latest
# or
npm i -g heravision
heravision setup --all
```

## Usage
```
heravision extract ./screenshot.png --mode ui --json
heravision mcp          # MCP stdio server
heravision doctor
heravision version
```

## MCP Tools
- `heravision_extract` — path, mode → JSON+markdown
- `heravision_compare` — path_a, path_b → diff
- `heravision_describe` — markdown alias

## Performance
Binary 9.3MB, <50ms tanpa OCR, CGO_ENABLED=0.

## Config
See `heravision.json.example`
