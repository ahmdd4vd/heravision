# Extra-Scale Recall Ablation

**Experiment:** add a 1.25x canonical scale to the stable proposal path.  
**Training:** dedicated filter trained on COCO128 development split.  
**Evaluation:** 22 accepted-provisional MD1 samples, IoU 0.50.

## Results

| Variant | MD1 predictions | TP | FP | FN | Precision | Recall | F1 | Mean B1 ms |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Stable default, support >=2 | 31 | 13 | 18 | 9 | 0.4194 | 0.5909 | 0.4906 | 30.59 |
| Stable + extra scale 1.25 | 62 | 13 | 49 | 9 | 0.2097 | 0.5909 | 0.3095 | 38.55 |

The extra scale increases computation and candidate output but does not recover any additional provisional true positives. Precision and F1 fall materially. Per-domain F1 also fails to improve: Imagenette changes from 0.4444 to 0.2857, Imagewoof remains 0.3333 to 0.3000, and Wikimedia diagrams falls from 0.6154 to 0.3200.

## Decision

Do not promote the extra scale. Keep it opt-in through `-scale-extra-fraction 1.25` for future small-object experiments, but retain the default three-scale stable path. The next recall work should improve candidate formation or evidence aggregation rather than simply increasing resolution.
