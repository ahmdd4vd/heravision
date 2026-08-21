# MD1 Accepted-Only Baseline Report

**Branch:** `vision-next`  
**Baseline commit:** `0fa00de`  
**Dataset:** 22 `accepted-provisional` samples from the 30-sample MD1 fixture  
**Excluded:** 7 `needs-review` samples and 1 `accepted-ignore` sample  
**IoU threshold:** 0.50  
**Runtime configuration:** CPU, `GOMAXPROCS=1`, `GOMEMLIMIT=1800MiB`, `max-side=256`, `legacy-max-pixels=24000000`

> This is a development baseline, not a final benchmark claim. The accepted status was produced by a same-agent second pass and still requires independent human review before publication.

## Aggregate results

| Engine | Predictions | TP | FP | FN | Precision | Recall | F1 | Mean predictions/sample |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| B0 legacy | 337 | 9 | 328 | 13 | 0.0267 | 0.4091 | 0.0501 | 15.32 |
| B1 raw | 1,385 | 15 | 1,370 | 7 | 0.0108 | 0.6818 | 0.0213 | 62.95 |
| B1 filtered | 64 | 15 | 49 | 7 | 0.2344 | 0.6818 | 0.3488 | 2.91 |

The region filter is therefore essential in the current pipeline. It preserves the same aggregate B1 recall as raw proposals while reducing the mean proposal count from 62.95 to 2.91 per sample. Precision rises from 0.0108 to 0.2344 and F1 rises from 0.0213 to 0.3488.

## Per-domain results

| Domain | Engine | TP | FP | FN | Precision | Recall | F1 | Predictions |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| Imagenette | B0 | 1 | 13 | 5 | 0.0714 | 0.1667 | 0.1000 | 14 |
| Imagenette | B1 filtered | 3 | 7 | 3 | 0.3000 | 0.5000 | 0.3750 | 10 |
| Imagewoof | B0 | 2 | 15 | 6 | 0.1176 | 0.2500 | 0.1600 | 17 |
| Imagewoof | B1 filtered | 4 | 8 | 4 | 0.3333 | 0.5000 | 0.4000 | 12 |
| Wikimedia diagrams | B0 | 6 | 300 | 2 | 0.0196 | 0.7500 | 0.0382 | 306 |
| Wikimedia diagrams | B1 filtered | 8 | 34 | 0 | 0.1905 | 1.0000 | 0.3200 | 42 |

## Interpretation

B1 filtered is materially better than B0 on this provisional accepted subset in aggregate F1, and it improves F1 in all three domains. The improvement is primarily due to suppressing proposal explosion rather than understanding object semantics. B1 still misses 7 of 22 provisional regions, so it is not yet a general visual understanding engine.

The most important engineering problem revealed by this baseline is proposal quality: raw B1 produces too many regions, while the filter can suppress useful candidates in some natural-image cases. The next controlled change is therefore scale-stable, boundary-aware merging followed by the same evaluation protocol.

## Reproducibility artifacts

- Accepted-only manifest: `experiments/manifests/md1-ground-truth-eval-accepted-22.json`
- Raw run: `experiments/runs/md1-accepted-22-b1-raw-baseline/`
- Filtered run: `experiments/runs/md1-accepted-22-b1-filtered-baseline/`
- Accepted-only manifest builder: `experiments/build_md1_eval_manifest.py --accepted-only`
- Evaluation manifest validator: `experiments/validate_eval_manifest.py`
- Per-domain summarizer: `experiments/summarize_gt_by_domain.py`

## Gate decision

The baseline passes the gate to implement one algorithmic ablation: **scale stability and boundary-aware merge**. It does not pass a production-release gate, a semantic-understanding gate, or a final-publication benchmark gate.
