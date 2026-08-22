# Domain Gate Feature v5 Report

## Scope

This experiment expands the CPU-first domain feature vector from 7 to 15 statistics. The added features are aspect ratio, blank fraction, low-information fraction, high-edge fraction, orientation concentration, axial orientation concentration, luminance range, and a compact line-structure score. All features are computed from the existing evidence field without a GPU or new runtime dependency.

## Contract

Go runtime and Python trainer now share the same ordered 15-feature contract. Domain provenance records the complete feature vector in the `domain-statistics` evidence note, allowing the trainer and reviewer to reconstruct the exact input used by the classifier.

## Ablation on expanded provisional fixture

The expanded fixture contains 34 provisional AI-reviewed samples, with 33 successful runtime predictions.

| Variant | Coverage | Selective accuracy | Main behavior |
|---|---:|---:|---|
| Existing heuristic | 78.8% | 26.9% | Natural photos and screenshots often routed to diagram/document |
| Four-class model with 15 features | 39.4% | 46.2% | More abstention; partial improvement but still misroutes |

The four-class model's provisional test subset accuracies were 0.571 for ambiguous, 0.857 for diagram/document, 0.857 for natural photo, and 0.714 for screenshot/document. These are not independent benchmark numbers because the labels are AI-generated and the dataset remains small/source-overlapping.

## Failure analysis

The model correctly abstained on all four ambiguous candidates, which is a useful safety signal. However, it still confused three diagrams with natural photos, three natural photos with diagrams, and one screenshot with a diagram. The model also abstained on most screenshot and natural-photo samples, indicating that the additional statistics improve separation only partially and do not provide enough semantic/layout signal by themselves.

## Decision

Do not promote the 15-feature model or change default B1 routing. Do not run blind MD0. The ablation is useful as a research result: cheap image-level statistics improve provisional selective accuracy from 26.9% to 46.2% while reducing coverage, but the model is still below the project acceptance gate.

## Next technical direction

The next improvement should add bounded layout/OCR proxies and region-distribution features rather than more raw pixel statistics. Candidate signals include connected-component count, text-like horizontal stroke density, repeated row/column alignment, stable-region size histogram, and a natural-photo texture-vs-line score. These should be ablated one group at a time and evaluated on an independently reviewed holdout when available.
