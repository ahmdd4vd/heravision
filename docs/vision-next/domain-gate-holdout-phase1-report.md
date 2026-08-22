# Domain Gate Holdout Phase 1 Report

## Objective

This phase prepares a reviewable domain calibration and holdout structure without claiming that the current metadata labels are independent ground truth. The work separates train, calibration, and holdout IDs deterministically and adds a validation tool that rejects incomplete reviewer annotations.

## Split design

The 30 candidate samples are split stratified by the current provisional `visual_domain` label using a deterministic round-robin assignment. No sample ID occurs in more than one split.

| Split | Count | Ambiguous | Diagram/document | Natural photo |
|---|---:|---:|---:|---:|
| Train | 8 | 2 | 3 | 3 |
| Calibration | 12 | 2 | 5 | 5 |
| Holdout | 10 | 2 | 4 | 4 |
| **Total** | **30** | **6** | **12** | **12** |

The split audit reports 30 unique IDs and zero split-overlap errors. The distribution is suitable for plumbing tests, but the sample count is too small for a production claim.

## Independence status

The fixture remains **candidate metadata-derived**, not a clean benchmark. Train and natural-photo samples overlap the semantic pseudo-training source. Some diagram samples overlap existing MD0-family material or use external assets whose source/license audit is pending. Ambiguous samples overlap the ImageNet-style family used in MD0. These facts are recorded in `source_overlap` fields and must not be hidden in later reports.

## Review protocol

The generated review template contains no proposed labels or model predictions. A reviewer must fill `review_domain`, `review_answerability`, and `review_confidence`, then provide reviewer identity, date, and an independence statement. The validator returns `pending` until all required fields are complete. A valid schema alone does not prove independent review; the process and reviewer separation must still be verified.

## Baseline artifacts

B1 processed all 30 candidate images successfully after removing a non-image `.DS_Store` entry from manifest generation. The existing heuristic domain gate and optional calibrated model can be rerun against the split manifests. No threshold was tuned from the holdout output, and no blind MD0 benchmark was run.

## Decision

The engineering split and review workflow are ready. The data is not yet ready to calibrate a production domain model. The next required action is to complete two independent reviews and reconcile disagreements without showing either reviewer the model predictions. Only then can a reviewed train/calibration set and a locked holdout set be created.
