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
- OCR WASM integration

## PR
- Keep binary <12MB, CGO_ENABLED=0
- All logs to stderr, stdout only JSON-RPC
