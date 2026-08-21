# Robustness and Concurrency Hardening Report

## Resolution sweep

The stable B1 path was run on the real 30-sample MD1 image set at three canonical maximum sides.

| Max side | Completed | Errors | B1 regions | Mean B1 ms | Answer statuses |
|---:|---:|---:|---:|---:|---|
| 64 | 30 | 0 | 0 | 10.43 | 30 insufficient evidence |
| 256 | 30 | 0 | 38 | 29.67 | 24 answered, 6 insufficient evidence |
| 512 | 30 | 0 | 104 | 52.93 | 25 answered, 5 insufficient evidence |

The engine does not crash at max-side 64, but the current stable filter retains no regions. Therefore 64 is not a supported quality configuration even though it is operationally safe. The recommended development configuration remains max-side 256; max-side 512 is a slower high-detail option requiring separate quality measurement.

## Corrupt input isolation

A temporary manifest containing one valid JPEG and one malformed binary file produced one successful result and one explicit error. The evaluator completed the valid sample and recorded the corrupt sample as `status=error` with an `unknown format` message. The error did not terminate the whole run.

## Concurrency hardening

The original decode path used a mutable global `processor.MaxPixels` that was temporarily changed by B0/B1 adapter calls. The new path adds `DecodeWithMaxPixels`, passes limits per request through `facts.Extract` and B1, and removes adapter mutation of the global limit. A regression test confirms per-call decoding does not mutate the global value.

## Verification

The following checks pass after the change:

| Check | Result |
|---|---|
| `go test ./...` | Pass |
| `go vet ./...` | Pass |
| Race detector on processor, facts, and visionnext | Pass |
| 30-sample stable smoke after refactor | 30/30, zero errors |
| Invalid-input isolation | Pass |

## Remaining limits

This is engineering robustness, not proof of visual correctness. Resolution-dependent region quality, relation correctness, semantic understanding, and independent annotation remain open. The engine should reject or warn on unsupported low-resolution quality configurations at the API layer in a future release rather than silently presenting zero-region output as a successful visual interpretation.
