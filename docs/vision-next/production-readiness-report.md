# HeraVision Next Production Readiness Report

**Branch:** `vision-next`  
**Latest hardening commit:** `77bc2794c4071829aba4e81adc516a9b4b8f2578`

## Hardening completed

The evaluator now validates both B0 and B1 `SceneGraph` contracts before a sample is marked successful. Run summaries record the manifest, git SHA, Go version, operating system, architecture, `GOMAXPROCS`, `GOMEMLIMIT`, and all relevant engine flags. The CLI exposes experimental scale-stability and relation-pruning paths without changing the B1-old default.

The code passes ordinary unit tests, `go vet`, and the Go race detector with `CGO_ENABLED=1` and `GOMAXPROCS=1`. The final accepted-only smoke run completed 22/22 samples with zero errors using the combined stable configuration.

## Release test matrix

| Check | Result |
|---|---|
| Unit tests `go test ./internal/visionnext/... ./cmd/vision-eval` | Pass |
| Static analysis `go vet ./internal/visionnext/... ./cmd/vision-eval` | Pass |
| Race detector `CGO_ENABLED=1 go test -race ./internal/visionnext/...` | Pass |
| Blind MD0-300 corrected B1 filter | 300/300, zero errors |
| Blind MD0-300 stable filter | 300/300, zero errors |
| Accepted MD1 combined stable smoke | 22/22, zero errors |
| Independent final annotation | Not complete |
| Semantic-label benchmark | Not available |
| Relation-labeled benchmark | Not available |
| Public production API/load test | Not complete |

## Current recommended modes

The default B1-old mode remains the safest reproducibility baseline. The stable mode should be treated as an experimental candidate because it is slower and its quality evidence is still provisional. The combined command used for the final smoke is:

```bash
GOMAXPROCS=1 GOMEMLIMIT=1800MiB go run ./cmd/vision-eval \
  -manifest experiments/manifests/md1-ground-truth-eval-accepted-22.json \
  -output experiments/runs/md1-accepted-22-production-smoke-final \
  -mode general -max-side 256 -legacy-max-pixels 24000000 \
  -scale-stable -relation-prune \
  -region-filter experiments/runs/coco128-verified/region-filter-stable.json \
  -region-filter-threshold 0.90
```

## Why this is not a production release

The engine is operationally robust enough for continued research and controlled internal experiments, but it is not yet a general-purpose production vision product. The main blockers are quality and scope rather than compilation: only provisional region labels exist for MD1, seven samples still need independent decisions, semantic labels are intentionally absent, relation accuracy has no labeled benchmark, and the stable path increases CPU time.

The branch must therefore remain separate from `main`. A production release requires independent annotation, a larger multi-domain labeled benchmark, explicit abstention metrics, load and memory tests at supported resolutions, a stable API contract, and a documented decision about whether the slower stable mode is worth its quality trade-off.
