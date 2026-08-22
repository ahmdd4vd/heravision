# Domain Gate AI Review Phase 1 Report

## Scope

At the user's request, the 22 calibration/holdout review samples received two AI passes: a visual first pass and a separate self-audit pass. This is explicitly **AI-provisional annotation**, not independent human review.

## Review result

The two passes agreed on all 22 samples:

| Provisional domain | Count |
|---|---:|
| `natural_photo` | 13 |
| `diagram_document` | 9 |
| `screenshot_document` | 0 |
| `ambiguous` | 0 |

All 22 images were judged visually clear and answerable in the AI pass. The contact sheet showed fish/dog/person photos and diagrams/technical drawings/flowcharts. No genuine ambiguous sample and no screenshot-document sample was identified in this subset.

## Methodological status

Reviewer A is recorded as `AI reviewer — pass 1`; reviewer B is recorded as `AI reviewer — self-audit pass 2`. Reconciliation therefore reports 22 `ai-consensus-provisional` rows and zero independent consensus rows. The consensus-manifest generator correctly refuses to build an independent consensus manifest from these labels.

The explicit `--allow-ai-provisional` flag permits a separate calibration experiment only. The resulting model artifact is marked `ai-provisional-experiment` and must not be used for production or final benchmark claims.

## Provisional model result

A logistic domain model trained on the 21 samples for which evidence predictions were available produced 100% accuracy on its tiny deterministic test subset (4 test samples per class). This number is not evidence of generalization because the labels were produced by the same AI process and the dataset is small and source-overlapping.

When run on the candidate fixture with the calibrated model, it accepted only 3 of 28 available domain hypotheses, giving 10.7% selective coverage and 100% selective accuracy against the provisional labels. The dominant behavior was abstention. This is a safety-oriented experiment, not a useful production model.

## Decision

The AI review successfully unblocked provisional plumbing experiments while preserving the independent-review guard. It does not unlock a benchmark claim, default routing, or production semantic training. The dataset quality issue is also clear: the current review subset lacks ambiguous and screenshot-document examples.

## Required next step

Add genuinely ambiguous and screenshot-document samples from a separately sourced dataset, obtain human-independent review if a final benchmark claim is required, and only then train/calibrate a domain model on reviewed labels. Keep the AI-provisional model opt-in and clearly marked.
