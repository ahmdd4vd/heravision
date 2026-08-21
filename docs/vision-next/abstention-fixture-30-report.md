# MD1 Abstention Fixture Report

**Fixture:** `experiments/manifests/md1-abstention-fixture-30.json`  
**Samples:** 30  
**Status:** pending independent answerability review

## Purpose

This fixture separates three things that must not be confused: the original image, the provisional region-review status, and whether an answer is safe to provide. The 22 `accepted-provisional`, 7 `needs-review`, and 1 `accepted-ignore` statuses are copied as context only. They are not used as final answerability labels.

The reviewer must later fill `expected_answer_status` with `answered`, `abstain`, or `insufficient_evidence` based on an explicit answerability policy. Until that happens, the fixture is a test harness, not a quality benchmark.

## Candidate run

The stable B1 path with the stable COCO-trained filter and relation pruning completed all 30 samples without errors.

| Metric | Result |
|---|---:|
| Samples | 30 |
| Completed | 30 |
| Errors | 0 |
| B1 regions | 38 |
| Mean B1 runtime | 29.27 ms |
| Answered | 24 |
| Abstain | 0 |
| Insufficient evidence | 6 |

The absence of explicit `abstain` cases means the next review should include images with a retained but weak region. Otherwise, the system mostly separates images into “some stable structure” and “no retained region.”

## Review requirement

An independent reviewer must decide whether each image is answerable and whether the generic answer is safe. Semantic object names are intentionally excluded. The result should measure unsafe answered rate, answered coverage, and correct refusal rate—not merely the number of answers.
