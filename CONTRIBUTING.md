# Contributing to HeraVision

## Dev
```
go test ./...
go vet ./...
heravision bench
```

## Good first issues
- add `mode: mobile`
- improve classifier thresholds
- OCR latency: rec-only pass per text_block crop (target <300ms total)

## PR
- Keep the core pure Go: CGO_ENABLED=0, binary ~14MB (purego binding included)
- All logs to stderr, stdout only JSON-RPC
