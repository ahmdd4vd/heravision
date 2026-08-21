# Evidence-First Answer and Abstention Report

**Branch:** `vision-next`  
**Scope:** generic visual evidence only; no semantic object labels

## What changed

B1 and B0 observations now include a schema-level `Answer`. The answer builder is deliberately conservative. It can return `answered` with the text `stable visual structure detected`, `abstain` when a region exists but its evidence score is weak, or `insufficient_evidence` when no supported region is available. Every non-empty answer copies region provenance evidence into the answer.

The builder does not emit labels such as `person`, `dog`, `car`, or `diagram`. This preserves the evidence-first contract while semantic ground truth is still unavailable.

## Evidence score

The current development score is deterministic and interpretable:

```text
0.45 × scale stability
+ 0.40 × boundary strength
+ 0.15 × bounded area support
```

The score is a routing signal for answer status, not a calibrated probability. It must not be described as accuracy or semantic confidence until an independently labeled calibration set exists.

## Observed statuses

| Run | Samples | Errors | Answered | Abstain | Insufficient evidence |
|---|---:|---:|---:|---:|---:|
| Accepted MD1 stable | 22 | 0 | 17 | 0 | 5 |
| Blind MD0 stable | 300 | 0 | 233 | 0 | 67 |

The current feature distribution produced no `abstain` cases in these runs; samples either had no retained region or exceeded the weak-evidence threshold. This is itself useful diagnostic information: the next calibration phase should include intentionally ambiguous and low-evidence images so the `abstain` path is exercised rather than merely present in code.

## Tests

The answer package has unit tests for empty input, weak evidence, generic answered output, and provenance retention. The full visionnext suite and `go vet` pass after integration.

## Limitations

These statuses do not yet measure whether the engine abstained correctly. Correct abstention requires an independently annotated set containing answerable and unanswerable cases. Until that exists, the answer status is an auditable routing decision, not a validated quality metric.

## Next experiment

Create an abstention fixture with clear-answer, ambiguous, blank, heavily occluded, and out-of-scope images. Review it independently, then measure coverage, answered precision, answered recall, and unsafe-answer rate. The engine should be rewarded for refusing unsupported claims, not only for producing more outputs.
