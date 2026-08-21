# Domain Gate Phase 2 Report

## Scope

This phase adds a candidate calibration fixture and an explicit routing action to the domain gate. The fixture is designed to expose false-blocking and false-allowing before any future promotion.

The fixture is intentionally **not claimed as independent**. It contains overlap with existing local training or blind-family sources, and external diagram assets whose licensing/source review is still pending.

## Fixture

| Candidate domain | Count | Source status |
|---|---:|---|
| `natural_photo` | 12 | Imagenette/Imagewoof train images; overlaps semantic pseudo-training |
| `diagram_document` | 12 | Wikimedia14 overlap plus six external diagram/flowchart assets; review pending |
| `ambiguous` | 6 | ImageNet-style validation candidates; family overlap with MD0 |
| **Total** | **30** | Candidate only; independent review pending |

Each sample records `visual_domain`, `answerability`, `source`, `source_overlap`, `independent_review_status`, and notes. This prevents the fixture from being accidentally treated as a clean benchmark.

## Explicit routing schema

The domain result now contains an `Action` field:

| Action | Meaning |
|---|---|
| `allow_object_semantic` | Object semantic may run, subject to its own evidence gate |
| `block_object_semantic` | High-confidence diagram/document; object classifier is blocked |
| `abstain_domain` | Domain evidence is insufficient; object classifier is blocked |

The action is emitted inside the domain evidence note and is attached to an image-level hypothesis with region provenance.

## Candidate evaluation

The candidate fixture was processed with B1 and the opt-in v4 pseudo semantic model. The heuristic gate produced domain hypotheses for 28/30 samples.

| Metric | Result |
|---|---:|
| Domain hypotheses | 28/30 |
| Non-ambiguous selective coverage | 78.6% |
| Selective accuracy against candidate metadata | 50.0% |
| Ambiguous → diagram_document | 5 |
| Diagram/document → diagram_document | 11 |
| Natural photo → diagram_document | 6 |
| Natural photo → ambiguous | 4 |

These values are diagnostic only. The six natural-photo false positives and five ambiguous false-allowing cases show that the heuristic is not ready to become a default hard gate.

## Decision

Keep the domain gate opt-in and conservative. The explicit action schema is accepted as an engineering contract, but the heuristic classifier is not accepted as a production-quality domain model. No blind benchmark is run from this phase.

The next phase requires a genuinely independent review set or newly sourced/licensed images, followed by a small calibration model trained only on that domain fixture. The calibration report must separate false-blocking of answerable natural photos from false-allowing of diagrams/documents and must include P95 CPU latency.
