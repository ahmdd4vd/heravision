# B1 MD0-300 Stable Retest Report

**Manifest:** `experiments/manifests/blind-md0-300.json`  
**Samples:** 300 total: 143 Imagenette, 143 Imagewoof, 14 Wikimedia diagrams  
**Runtime:** CPU, `GOMAXPROCS=1`, `GOMEMLIMIT=1800MiB`, `max-side=256`, `legacy-max-pixels=24000000`  
**Important:** MD0 is blind and has no complete region ground truth. These results measure completion, graph size, timing, and geometry overlap diagnostics, not final precision/recall.

## Completion and aggregate runtime

| Configuration | Completed | Errors | B1 regions | Mean B1 ms | Mean matched IoU diagnostic |
|---|---:|---:|---:|---:|---:|
| B1-old + corrected filter | 300 | 0 | 576 | 8.04 | 0.7256 |
| B1-stable + stable filter | 300 | 0 | 287 | 21.52 | 0.7466 |
| B1-stable + stable filter + relation prune | 300 | 0 | 287 | 20.37 | 0.7466 |

The stable candidate reduces region output by 50.2% relative to corrected B1-old. Its mean runtime is approximately 2.7 times higher because it computes three views. Relation pruning leaves region output and geometry diagnostics unchanged and reduces graph work slightly.

## Per-domain comparison

| Domain | Configuration | Mean regions | Mean edges | Mean B1 ms | Mean coverage | Mean matched IoU |
|---|---|---:|---:|---:|---:|---:|
| Imagenette | Corrected B1-old | 2.16 | 4.12 | 7.11 | 0.1188 | 0.1762 |
| Imagenette | Stable + filter | 0.97 | 0.71 | 20.84 | 0.1220 | 0.1820 |
| Imagewoof | Corrected B1-old | 1.41 | 1.46 | 6.39 | 0.1325 | 0.2428 |
| Imagewoof | Stable + filter | 0.86 | 0.22 | 19.50 | 0.1184 | 0.2312 |
| Wikimedia diagrams | Corrected B1-old | 4.64 | 19.36 | 34.43 | 0.0012 | 0.0419 |
| Wikimedia diagrams | Stable + filter | 1.79 | 2.93 | 49.14 | 0.0012 | 0.0419 |

## Interpretation and limitations

All 300 images completed without runtime errors under the CPU budget. The stable candidate is substantially sparser and has a slightly higher aggregate geometry diagnostic, but it is slower and often emits fewer than one region per image on average for natural-image domains. Because MD0 has no verified ground truth, this cannot establish that the missing regions are harmless or that the stable candidate is more accurate.

The relation-pruned combined configuration changes no region result and reduces only a small number of edges on this blind run. It remains safe as a graph-size optimization, but relation accuracy still needs relation-labeled data.

## Gate decision

The stable filter passes an **engineering retest gate**: 300/300 completion, lower graph size, and no observed errors. It does not pass a final quality gate because blind precision/recall are unavailable. The stable path remains experimental and must not replace the default B1-old path until an independent reviewer finalizes MD1 and a larger labeled benchmark confirms recall.
