# Abstention Policy v2 and Touching-Safe Phase

## Conservative answer policy

The default generic answer threshold is now `answer-min-score=0.65`. The threshold remains an evidence score, not a calibrated probability. The CLI flag permits controlled ablations.

On the same-agent internal answerability fixture, threshold 0.45 produced 79.17% coverage on internally expected-answerable samples and 83.33% unsafe answered rate on internally expected-nonanswerable samples. Threshold 0.65 produced 75.00% coverage and 33.33% unsafe answered rate. Threshold 0.70 reduced unsafe answered rate to zero but coverage to 33.33%, so it was rejected as too conservative for the current development stage.

The final blind 300 with threshold 0.65 completed 300/300 with zero errors:

| Status | Count |
|---|---:|
| Answered | 197 |
| Abstain | 36 |
| Insufficient evidence | 67 |

These blind counts are operational diagnostics only because no answerability ground truth exists for the 300 samples.

## Touching-safe relation policy

The opt-in `-relation-touching-safe` mode suppresses touching when one bbox contains the other and attaches `boundary-contact` evidence to retained boundary-adjacent touching relations.

| Run | Total edges | Touching | Other predicates |
|---|---:|---:|---|
| MD1 30 baseline | 50 | 17 | contains 12, overlap 16, above 3, left_of 2 |
| MD1 30 safe | 38 | 5 | unchanged |
| Blind 300 baseline | 172 | 59 | contains 44, overlap 57, above 8, left_of 4 |
| Blind 300 safe | 128 | 15 | unchanged |

Both MD1 and blind 300 completed with zero errors. Touching-safe remains an opt-in candidate because independent relation-complete labels do not yet exist.

## Decision

Promote answer threshold 0.65 as the conservative candidate default, with explicit warning that it is based on same-agent development review. Keep touching-safe available as opt-in until relation annotation confirms the removed containment-derived touching edges were false. Do not add semantic labels yet.
