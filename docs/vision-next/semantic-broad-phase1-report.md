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

On the COCO development run, 800 semantic hypotheses were emitted across 126 samples with zero runtime errors. At the conservative threshold 0.65, no hypothesis was accepted because the evidence score remained below threshold. A development ablation at threshold 0.35 accepted 15 hypotheses across COCO.

On the 30-sample MD1 candidate run at threshold 0.35:

| Metric | Result |
|---|---:|
| Samples completed | 30/30 |
| Semantic hypotheses | 152 |
| Samples with accepted semantic | 3 |
| Coverage | 10.0% |
| Selective accuracy against metadata-derived target | 0.0% |
| Evidence missing | 0 |

The three accepted predictions did not match the provisional broad targets. This is a clear failure of semantic quality, not a success. The model is currently biased toward `vehicle` and its region geometry/chroma features are insufficient to distinguish the broad classes on MD1.

## Decision

Do **not** promote the semantic model or threshold 0.35 to default. Keep semantic integration opt-in. The phase successfully establishes the contract, training path, provenance, and failure metrics, but it does not yet produce a useful `dog` or broad-category recognizer.

## Next semantic engineering step

Add lightweight crop-pixel features computed identically during training and inference, such as coarse luminance/chroma histograms and spatial texture summaries. Then retrain on a larger balanced COCO development sample, add hard negatives, and require a holdout semantic fixture before any label is accepted as a product claim.
