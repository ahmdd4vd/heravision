# Semantic Broad Phase 1 Report

## Scope

This phase adds an opt-in broad semantic hypothesis layer. It consumes B1 regions, computes a small CPU feature vector, applies a logistic model trained only from the COCO128 development split, attaches semantic and visual-quality evidence, and emits `accepted`, `candidate`, or `unknown` hypotheses.

The default B1 geometry behavior is unchanged when `-semantic-model` is omitted.

## Fixture

A 30-sample provisional broad semantic fixture was generated from real MD1 images. The broad targets are metadata-derived development targets, not independent semantic annotations.

| Broad target | Count |
|---|---:|
| animal | 12 |
| diagram | 9 |
| artifact | 7 |
| vehicle | 1 |
| unknown | 1 |

The fixture deliberately records `independent_review_status=pending` and keeps semantic bbox annotation empty.

## Training

The scorer was trained from 341 B1 regions across 106 COCO development images. The training labels map COCO classes into `animal`, `person`, `vehicle`, and `artifact`. The trainer now uses a deterministic image split and class-balanced logistic loss. No blind MD0 or MD1 image was used for training.

## Results

The first geometry/chroma model used 341 matched B1 regions across 106 COCO development images. On the 30-sample MD1 candidate run at diagnostic threshold 0.35 it achieved 10.0% accepted-sample coverage and 0.0% selective accuracy against the metadata-derived target. This was a failure and was not promoted.

A v3 model added a fixed 8×8 crop-pixel sampler. The sampler computes luminance mean and standard deviation, two chroma means, chroma magnitude, edge density, dark-pixel fraction, and bright-pixel fraction. The same formulas are used in Python training and Go inference. On the COCO development split, the per-label holdout accuracies were 0.657 animal, 0.700 artifact, 0.686 person, and 0.714 vehicle; these are binary one-vs-rest region metrics with severe class imbalance and are not a semantic product accuracy.

On MD1, v3 produced the following results:

| Configuration | Accepted samples | Coverage | Selective accuracy vs metadata target | Runtime mean |
|---|---:|---:|---:|---:|
| B1 baseline, no semantic | 0/30 | 0.0% | Not applicable | 30.7 ms |
| v3, default evidence threshold 0.65 | 0/30 | 0.0% | Not applicable | 29.7 ms |
| v3, diagnostic threshold 0.35 | 17/30 | 56.7% | 0.0% | Not used for product claim |

The threshold-0.35 result demonstrates why lowering a threshold is not a solution: coverage increased, but the accepted labels were wrong, predominantly `vehicle`. The default 0.65 policy correctly abstained from making unsupported semantic claims. All 30 MD1 samples completed without runtime errors, and crop features added no measured mean latency in this small noisy run; the timing result is not a performance certification.

## Decision

Do **not** promote v3 or threshold 0.35 to the default semantic model. Keep semantic integration opt-in and keep the default evidence threshold at 0.65. The implementation phase is successful as engineering infrastructure: the scorer now has a bounded pixel sampler, matching training/runtime features, unit tests, provenance fields, and an evaluator. It is not yet a useful broad-category recognizer and must not claim to understand “this is a dog.”

## Next semantic engineering step

The next bottleneck is not another threshold. It is dataset and objective design. Build an independently reviewed image-level holdout with balanced animal, person, vehicle, artifact, diagram, and unknown classes; aggregate region evidence at image level with max-pooling/noisy-OR plus an abstention margin; and add hard negatives from diagrams, documents, and visually confusing objects. No new semantic label should be promoted until the independent holdout reports coverage, selective accuracy, unsafe-answer rate, and per-domain confusion.
