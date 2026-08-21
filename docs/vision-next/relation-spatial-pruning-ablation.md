# Relation Spatial-Pruning Ablation

**Baseline commit:** `e846b0a`  
**Fixture:** 22 `accepted-provisional` MD1 samples  
**B1 mode:** filtered region proposals, threshold `0.95`  
**IoU threshold:** 0.50  
**Default prune gap:** 48 canonical pixels

## Implementation

`relation.Build` remains the exhaustive legacy relation builder. The new `relation.BuildPruned` path is opt-in through `-relation-prune`. It skips pairs whose bounding boxes are farther apart than the configured gap, but always retains containment, overlap, touching, and nearby pairs. The default gap is deliberately conservative at 48 pixels on the canonical view.

## Results

| Metric | B1 filtered baseline | B1 filtered + pruning | Change |
|---|---:|---:|---:|
| Samples completed | 22 | 22 | 0 |
| Region predictions | 64 | 64 | 0 |
| Region precision | 0.2344 | 0.2344 | 0 |
| Region recall | 0.6818 | 0.6818 | 0 |
| Region F1 | 0.3488 | 0.3488 | 0 |
| Total relation edges | 256 | 224 | -12.5% |
| Mean relation edges/sample | 11.64 | 10.18 | -12.5% |
| Mean B1 runtime | 16.00 ms | 17.18 ms | +1.18 ms |

Predicate counts changed as follows: `above` 31 to 23, `left_of` 37 to 13, while `contains` remained 39, `overlapping` remained 73, and `touching` remained 76. The relation evaluator currently has no independently annotated relation ground truth, so this is an efficiency ablation rather than a claim of improved semantic relation accuracy.

## Decision

The change passes the **graph-size and region-regression gate**: it removes 12.5% of relation edges without changing region precision, recall, or F1 on the accepted-only fixture. It remains opt-in until a relation-labeled fixture can measure false relation and missed relation rates. The next phase should train a new region filter for the scale-stable representation, because the old filter was trained with `ScaleStability=1` and is not a valid final scorer for the new proposal features.
