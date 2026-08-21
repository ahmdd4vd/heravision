# B1 Scale-Stability and Boundary-Aware Proposal Ablation

**Baseline commit:** `3637cd4`  
**Ablation:** multi-scale proposal with normalized IoU consensus and boundary-gap guard  
**Fixture:** 22 `accepted-provisional` MD1 samples  
**IoU threshold:** 0.50  
**Runtime:** `GOMAXPROCS=1`, `GOMEMLIMIT=1800MiB`, `max-side=256`

## Implementation

The new path is behind the CLI flag `-scale-stable`; the default remains B1-old. It evaluates bounded views at fractions `0.65`, `0.85`, and `1.0` of the configured maximum side. Candidate boxes are matched in normalized coordinates. A cluster is retained when it is supported by at least two distinct scales, unless only one unique view exists. Boundary-strength disagreement blocks low-overlap cross-scale matches. The output contains `scale-consensus` and `scale-region` evidence references, plus `scale_support` and `scale_count` features.

## Aggregate results

| Variant | Predictions | TP | FP | FN | Precision | Recall | F1 | Mean predictions/sample | Mean B1 ms |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| B1-old raw | 1,385 | 15 | 1,370 | 7 | 0.0108 | 0.6818 | 0.0213 | 62.95 | 16.55 |
| B1-stable raw | 682 | 14 | 668 | 8 | 0.0205 | 0.6364 | 0.0398 | 31.00 | 31.27 |
| B1-old filtered | 64 | 15 | 49 | 7 | 0.2344 | 0.6818 | 0.3488 | 2.91 | 16.00 |
| B1-stable + old filter | 74 | 14 | 60 | 8 | 0.1892 | 0.6364 | 0.2917 | 3.36 | 33.14 |

## Interpretation

The proposal change reduces raw proposal count by 50.8% and improves raw precision and F1. It also increases B1 runtime by approximately 1.9x because it computes three views. However, recall falls from 0.6818 to 0.6364.

The old COCO-trained filter is not a valid final filter for the new feature distribution: `ScaleStability` was previously hardcoded to `1`, while the new path emits measured support ratios. Applying the old filter therefore produces a lower F1 and is not evidence that the proposal idea itself is invalid. A new filter must be trained only after the proposal representation is frozen.

Per-domain B1-stable raw F1 is 0.0258 for Imagenette, 0.0335 for Imagewoof, and 0.0516 for Wikimedia diagrams. The raw proposal F1 increase is therefore not sufficient to pass the stronger two-domain quality gate. The implementation remains behind an explicit flag and is not promoted to the default path.

## Decision

This is a **measured but non-promoted ablation**. The next valid experiment is a small region filter retrained on the stable proposal representation using the COCO development split, followed by the same MD1 accepted-only evaluation. No blind tuning is allowed.
