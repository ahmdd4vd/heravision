# Domain Gate Phase 1 Report

## Scope

This phase adds a CPU-first domain gate before the opt-in semantic object classifier. The gate estimates whether the canonical evidence field looks more like a natural photo, a diagram/document, or an ambiguous input. It is a routing gate, not an object recognizer.

The gate uses existing B1 evidence arrays and computes bounded global statistics: flatness mean, edge density, local contrast mean, chroma mean and variance, luminance variance, and orientation entropy. No new runtime dependency or neural model was added.

## Routing policy

| Gate result | Object semantic action |
|---|---|
| `natural_photo` | Run region semantic and image aggregation |
| `diagram_document`, medium confidence | Keep object semantic advisory; do not block solely on medium gate evidence |
| `diagram_document`, high confidence | Block object semantic classification |
| `ambiguous` | Block object semantic classification and emit `unknown` domain hypothesis |

Every domain result is emitted as an image-level hypothesis with evidence references containing feature summaries, domain scores, margin, and the contributing region IDs.

## Provisional evaluation

The gate was evaluated on the existing 30-sample MD1-derived fixture. Its domain targets are metadata-derived and not independent human labels; therefore the results are directional only.

| Metric | Result |
|---|---:|
| Samples with domain hypothesis | 24/30 |
| Non-ambiguous coverage | 66.7% |
| Selective accuracy against provisional domain target | 43.8% |
| Diagram/document → diagram/document | 7 |
| Diagram/document → ambiguous | 3 |
| Natural photo → diagram/document | 9 |
| Natural photo → ambiguous | 5 |

The false-positive rate on natural photos is too high for the gate to be a hard blocker at medium confidence. The implementation therefore blocks only high-confidence diagram/document results and all ambiguous results. This is safer than allowing every domain prediction to suppress object semantics, but it is not yet a production-quality domain classifier.

With the advisory gate and the v4 pseudo semantic model, all 30 samples completed without runtime errors. The output contained 85 object semantic hypotheses and no accepted object semantic claim under the default evidence threshold of 0.65. This preserves evidence-first abstention while domain quality is still weak.

## Decision

Do not run a blind benchmark yet. Do not call the domain gate accurate. Keep it opt-in behind `-semantic-model` and preserve the default B1 behavior when no semantic model is supplied.

## Next step

The heuristic should be replaced or calibrated using a separately reviewed domain dataset containing balanced natural photos, diagrams, documents, screenshots, and ambiguous images. The next implementation should add a domain calibration artifact and an explicit `block_object_semantic` decision field, then evaluate false blocking and false allowing separately. Only after those metrics stabilize should the gate become a default routing policy.
