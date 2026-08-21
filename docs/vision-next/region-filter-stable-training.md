# Stable Region Filter Training and MD1 Ablation

**Proposal representation:** B1 multi-scale stable path  
**Training data:** COCO128 verified development split only  
**Blind/MD1 data used for training:** none  
**IoU target:** 0.50

## Method correction

The trainer contained an IoU bug in its ground-truth intersection calculation: the x-axis upper bound used the truth box height instead of its width. The calculation was corrected before both comparison filters were retrained. Existing filter JSON artifacts remain unchanged; the corrected filters are new artifacts under `experiments/runs/coco128-verified/`.

## COCO training outputs

| Filter | Proposal source | Train-selected threshold | Held-out test F1 | Held-out test precision | Held-out test recall |
|---|---|---:|---:|---:|---:|
| Corrected B1-old | B1-old raw | 0.95 | 0.1920 | 0.1176 | 0.5217 |
| B1-stable | B1-stable raw | 0.90 | 0.2222 | 0.1489 | 0.4375 |

The threshold `0.90` for the stable filter was selected from the COCO training split and then held fixed for MD1. It was not tuned on MD1 or blind data.

## Accepted-only MD1 results

| Variant | Predictions | TP | FP | FN | Precision | Recall | F1 | Mean predictions/sample |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| B1-old + corrected filter | 66 | 14 | 52 | 8 | 0.2121 | 0.6364 | 0.3182 | 3.00 |
| B1-stable + stable filter | 31 | 13 | 18 | 9 | 0.4194 | 0.5909 | 0.4906 | 1.41 |

Per-domain F1 changed as follows:

| Domain | Corrected B1-old | B1-stable | Change |
|---|---:|---:|---:|
| Imagenette | 0.4000 | 0.4444 | +0.0444 |
| Imagewoof | 0.2857 | 0.3333 | +0.0476 |
| Wikimedia diagrams | 0.3077 | 0.6154 | +0.3077 |

Stable filtering improves precision and F1 in all three provisional domains, and reduces the mean output count by 53.0%. Recall falls by 4.55 percentage points overall, from 0.6364 to 0.5909. Because the fixture is provisional and only 22 samples, this passes the **development ablation gate** but not a final benchmark or production gate. The blind 300 retest is required.

## Artifacts

- Corrected filter: `experiments/runs/coco128-verified/region-filter-corrected-iou.json`
- Stable filter: `experiments/runs/coco128-verified/region-filter-stable.json`
- Corrected B1-old run: `experiments/runs/md1-accepted-22-b1-corrected-filter-ablation/`
- Stable filtered run: `experiments/runs/md1-accepted-22-b1-stable-new-filter-ablation/`
- Trainer: `experiments/train_region_filter.py`
