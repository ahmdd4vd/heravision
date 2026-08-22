# Expanded Domain Gate Phase 1 Report

## Scope

This phase adds provisional screenshot/document and ambiguous candidates, performs two AI review passes, and evaluates the existing heuristic and an optional four-class calibration model. These results are explicitly provisional and are not independent human benchmark claims.

## Expanded candidate fixture

The fixture contains 34 samples:

| Domain | Count | Status |
|---|---:|---|
| `natural_photo` | 13 | AI-reviewed provisional; existing candidate source |
| `diagram_document` | 9 | AI-reviewed provisional; existing candidate source |
| `screenshot_document` | 8 | AI-reviewed provisional; external source/license review pending |
| `ambiguous` | 4 | AI-reviewed provisional; external/blank-scan source review pending |

The two AI passes agreed on all 34 samples. Their agreement is recorded as `ai-consensus-provisional`, not independent consensus.

## Heuristic baseline

On 33 successfully emitted domain hypotheses, the heuristic achieved 78.8% non-ambiguous coverage and 26.9% selective accuracy against the provisional labels. The dominant errors were natural photos routed to `diagram_document` and screenshots routed to `diagram_document`. This confirms that the heuristic does not provide a reliable four-domain routing boundary.

## Four-class calibration experiment

A logistic model was trained on 33 available evidence samples with the provisional AI labels. Its deterministic test subset accuracy was 0.714 for ambiguous, 0.857 for diagram/document, 0.857 for natural photo, and 0.571 for screenshot/document. These values are not independent evidence because the labels are AI-generated and the dataset is small.

When run through the runtime gate on the expanded fixture, the model accepted only 3 of 33 domain hypotheses, giving 9.1% coverage and 0% selective accuracy against the provisional target. Most inputs were abstained. This is an example of a model that is safe by being excessively conservative but not useful or generalizable.

## Decision

Do not promote the expanded model or heuristic to default. Do not run blind MD0. The experiment successfully exposed two dataset and model problems: the screenshot class was absent from earlier fixtures, and the four-class model does not generalize from the small feature set to the expanded assets.

## Next step

The project needs a reviewed, source-audited dataset with at least 20 samples per domain, including genuine ambiguous examples and screenshots. Features should be computed at image level using region-count, region-size distribution, OCR/layout cues where available, and robust edge/flatness statistics. Calibration must be performed on one split and measured once on a frozen holdout. Human-independent review remains required for any final benchmark claim.
