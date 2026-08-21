# Semantic Broad Phase 2 Report

## Objective

This phase implemented the first semantic v4 changes proposed after the v3 failure: image-level aggregation, top-1-per-region support selection, a minimum consensus guard, ring-background features, and additional training-only pseudo-labeled images from ImageNet-style train splits.

The semantic layer remains opt-in. The default B1 engine and default answer threshold were not changed.

## Implemented changes

| Change | Purpose |
|---|---|
| Top-k image aggregation | Combine region evidence instead of treating one crop as an image claim |
| Top-1 region support | Prevent weak non-winning labels from accumulating across regions |
| Minimum support of 2 | A single region cannot become an accepted image-level claim |
| Ring-background features | Compare crop statistics against surrounding context |
| Train-only pseudo data | Add animal, artifact, and vehicle diversity without using MD1/MD0 holdout samples for training |
| Provenance expansion | Aggregated hypotheses list contributing region IDs and support scores |

The crop and ring sampler remains bounded to an 8×8 grid per region. The runtime model stores `min_support=2`, and the aggregation note records the required and observed support.

## Training data

The original COCO region training set contained 341 matched regions from 106 images. The new optional training run added 800 images from local `train` splits: 480 animal, 280 artifact, and 40 vehicle image-level pseudo labels. The pseudo-label run contributed up to 12 stable regions per image. These labels are noisy because an image-level label is assigned to selected regions; therefore the result is an experiment, not independently annotated region ground truth.

No blind MD0 sample or MD1 fixture sample was used to train this model.

## Results

The v4 pseudo model improved the provisional MD1 broad fixture relative to the earlier v4 diagnostic run, but remains far below the promotion bar.

| Configuration | Samples with accepted semantic | Coverage | Selective accuracy vs metadata-derived target |
|---|---:|---:|---:|
| v4 COCO-only, threshold 0.35 | 15/30 | 50.0% | 0.0% |
| v4 pseudo train-only, threshold 0.35 | 20/30 | 66.7% | 35.0% |
| v4 pseudo train-only, default threshold 0.65 | 0/30 | 0.0% | Not applicable |

The COCO development one-vs-rest holdout accuracy of the pseudo model was 0.734 animal, 0.680 artifact, 0.685 person, and 0.694 vehicle. These numbers are not equivalent to image-level semantic accuracy and should not be used as a product claim.

The pseudo model's provisional confusion matrix still shows systematic errors:

| Target | Predicted outcome |
|---|---|
| animal | 7 correct animal, 2 artifact, 3 abstain |
| artifact | 4 abstain, 3 animal |
| diagram | 5 animal, 2 artifact, 2 abstain |
| unknown | 1 animal |
| vehicle | 1 abstain |

The result is a real improvement over 0% selective accuracy, but it is not safe enough. The fixture itself is metadata-derived and not independently reviewed, so the number is directional only.

## Decision

Do **not** run blind MD0 retest yet. Do **not** promote the v4 pseudo model or threshold 0.35 to default. The default threshold 0.65 continues to abstain on MD1, which is the correct safety behavior while semantic quality is weak.

The main remaining blocker is domain coverage: the classifier was trained on natural-object categories but has no independently labeled diagram/document class. Consequently, diagrams are forced into animal/artifact/vehicle/person and produce false semantic claims at diagnostic thresholds.

## Next step

Acquire or create a separate, independently reviewed semantic training/holdout fixture containing natural photos, diagrams, documents, screenshots, and ambiguous images. Add a `diagram_or_document` gate before broad object classification. The gate should be trained and evaluated on non-blind data, with calibration and an explicit unknown class. Only after the independent holdout reaches the safety target should a frozen model be evaluated on a new blind benchmark.
