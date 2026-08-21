# Domain Gate Phase 3 Report

## Scope

This phase adds an independent-review template, a named domain feature vector, an optional logistic calibration model, and an explicit `-domain-model` CLI flag. The heuristic gate remains the fallback when no calibrated model is supplied.

## Review protocol

A new review template hides proposed labels and model predictions. Reviewers must independently assign `visual_domain`, `answerability`, confidence, and disagreement notes. The current candidate fixture is not yet independent; it contains source overlap and pending license/review status. Therefore its labels are suitable only for plumbing and diagnostic experiments.

## Model contract

The optional domain model consumes seven bounded features:

```text
flatness_mean, edge_density, contrast_mean, chroma_mean,
chroma_std, orientation_entropy, luma_std
```

Its output includes label, score, margin, confidence, and routing action. The CLI accepts the model through `-domain-model`. If the flag is omitted, the previous heuristic path remains active.

## Candidate calibration result

The logistic model was trained on 28 candidate samples with metadata-derived labels and a deterministic sample split. It achieved high holdout scores on the tiny candidate split, but this is not meaningful evidence of generalization because the fixture is small, overlapping, and not independently reviewed.

When run on the same candidate fixture with the conservative default calibration threshold:

| Metric | Result |
|---|---:|
| Domain hypotheses | 28/30 |
| Selective coverage | 7.1% |
| Selective accuracy | 100% |
| Diagram/document → diagram/document | 1 |
| Natural photo → natural photo | 1 |
| Other samples | abstain/ambiguous |

This is **over-abstention**, not a production improvement. The calibrated model is precise on the few claims it makes but does not yet cover enough images to be useful.

## Decision

Keep `-domain-model` opt-in and do not replace the heuristic globally. Do not tune the model threshold on MD0 or MD1. Do not run a blind benchmark from this candidate calibration. The current result demonstrates the desired safety trade-off—wrong accepted domain claims are suppressed—but lacks coverage and independent evidence.

## Next step

Obtain a genuinely independent, reviewed domain set with at least 20 samples per domain, including `natural_photo`, `diagram_document`, `screenshot_document`, and `ambiguous`. Train and calibrate on one split, freeze the model and threshold, then evaluate on an untouched holdout. Report false-blocking, false-allowing, ambiguous abstention recall, coverage, P95 latency, and provenance completeness separately.
