# Scale Support Recall Ablation

**Purpose:** test whether allowing candidates supported by one scale recovers missed regions.  
**Training split:** COCO128 development only.  
**Evaluation:** 22 accepted-provisional MD1 samples, IoU 0.50.  
**Caveat:** MD1 still requires independent review.

## Results

| Variant | COCO held-out F1 | MD1 predictions | TP | FP | FN | Precision | Recall | F1 |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Stable support >=2 + stable filter | 0.2222 | 31 | 13 | 18 | 9 | 0.4194 | 0.5909 | 0.4906 |
| Stable support >=1 + support-1 filter | 0.1477 | 78 | 13 | 65 | 9 | 0.1667 | 0.5909 | 0.2600 |

Allowing one-scale candidates did not recover any additional true positives on the 22-sample fixture. It more than doubled output count and substantially increased false positives. The result is consistent with the scale-stability hypothesis: one-scale-only regions are weaker candidates in this configuration.

## Decision

Keep the default minimum support at two distinct scales. The support-1 path remains available only as an experimental ablation through `-scale-min-support 1`; it is not promoted and must not be used for headline metrics.

The next recall experiment should target small-object preservation using a dedicated high-resolution candidate channel rather than weakening the stability requirement globally.
