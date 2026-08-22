# Domain Gate Review Workflow Phase 1

## Objective

This phase prepares the remaining human-review workflow for domain calibration. It does not fabricate reviewer labels and does not unlock model training from metadata-derived labels.

## Artifacts

| Artifact | Purpose |
|---|---|
| Reviewer A template | Independent domain/answerability review |
| Reviewer B template | Separate independent review |
| Reconciliation tool | Exact agreement only; disagreement remains unresolved |
| Adjudication template | Shows only disagreement samples for an adjudicator |
| Review validator | Rejects empty or invalid fields |
| Training guard | Refuses domain training unless reconciliation is consensus-ready |

The templates cover 22 calibration/holdout samples from the current 30-sample candidate fixture. The eight candidate train samples are not included in this review template because they remain unsuitable for training until a reviewed label source is established.

## Current status

Both reviewer templates are empty. Reconciliation therefore reports:

| State | Count |
|---|---:|
| Consensus | 0 |
| Disagreement/pending | 22 |
| Reviewer A complete | 0 |
| Reviewer B complete | 0 |
| Training unlocked | No |

The adjudication template contains the 22 unresolved rows and requires an adjudicator to inspect the original image, choose a domain and answerability, or mark the sample unresolved. It explicitly prevents consensus from being manufactured for the purpose of maximizing model coverage.

## Safety verification

Running the domain trainer with the pending reconciliation exits with `refusing to train: reviewer reconciliation is not consensus-ready` and creates no model artifact. This confirms that metadata-derived candidate labels cannot silently become training labels.

## Next action

Two independent reviewers must fill the A and B files separately. After that, run reconciliation, adjudicate only disagreements, remove unresolved samples from training and benchmark claims, and create a reviewed train/calibration/holdout manifest. Until then, the domain gate and semantic model remain experimental and opt-in.
