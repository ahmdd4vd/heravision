# Semantic Prior Gate Decision

## Decision

No semantic labeler, object-name predictor, or semantic prior is added in this phase.

## Reason

HeraVision Next currently has verified evidence for luminance, chroma, local contrast, boundaries, region geometry, and visible spatial relations. It does not yet have an independently reviewed semantic-label dataset covering the target domains. Adding labels such as `person`, `dog`, `car`, or `diagram component` without such evidence would violate the evidence-first contract and would make the system appear more capable without a reproducible quality measurement.

The stable proposal/filter ablation improves provisional MD1 region F1, but the blind 300 run has no complete region ground truth and therefore cannot authorize semantic claims. The semantic gate is explicitly **not passed**.

## Allowed behavior

The current engine may emit generic hypotheses such as `region`, `compact_region`, `elongated_region`, or `flat_surface_region` when their geometry and evidence support them. It must abstain from object names and activity claims. Any future semantic layer must attach pixel or region provenance, a confidence/uncertainty value, and a benchmark with independent annotations.

## Next prerequisite

Before adding a semantic prior, create an independent semantic fixture with at least two review passes, define label and abstention metrics, and prove that the semantic layer improves at least two domains without increasing unsupported claims.
