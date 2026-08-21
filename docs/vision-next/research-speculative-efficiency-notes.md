# External Research Notes — Speculative Efficiency

## Sources

1. Google Research, “Looking back at speculative decoding,” 2024-12-06: https://research.google/blog/looking-back-at-speculative-decoding/
2. Yan et al., “Decoding Speculative Decoding,” arXiv:2402.01528v3: https://arxiv.org/html/2402.01528v3
3. DeepSeek AI, DeepSpec repository: https://github.com/deepseek-ai/DeepSpec
4. Cheng et al., “DSpark: Confidence-Scheduled Speculative Decoding with Semi-Autoregressive Generation,” arXiv:2607.05147: https://arxiv.org/abs/2607.05147

## Key findings

Google describes speculative decoding as using a fast draft to propose work and a slower target to verify it. When verification agrees, computation is reused; when it disagrees, the speculative result is discarded. The original method preserves the target output distribution. The key transfer principle is not “guess without checking,” but “cheap proposal plus authoritative verification.”

The Google explanation also identifies that autoregressive LLM inference can be memory-bandwidth limited and that speculative decoding increases concurrency by turning sequential generation into a more parallel verification-like operation. This is relevant to HeraVision only as an analogy: the visual engine can propose cheap regions or labels and then verify them using stronger evidence. It is not the same algorithm because B1 is not autoregressive token generation.

“Decoding Speculative Decoding” reports that draft-model latency is a major bottleneck and that draft-model task accuracy is not the same as speculative acceptance performance. The lesson for HeraVision is to measure proposal latency, acceptance rate, and verification cost directly; a more accurate but slow proposal stage may lose overall throughput.

DeepSeek’s DeepSpec repository describes a workflow that prepares target outputs, trains draft models, and evaluates acceptance against a target model. It warns that its default setup assumes substantial GPU resources. Therefore, HeraVision should borrow the draft/verify and confidence-scheduling ideas, not claim that the same infrastructure is CPU-kentang compatible.

DSpark’s arXiv abstract describes semi-autoregressive drafting with a lightweight sequential module and confidence-scheduled verification that adapts verification length using estimated survival probabilities and throughput profiles. The direct vision analogue is selective region verification: verify only high-value or uncertain regions, choose verification depth from confidence, and early-exit when evidence survives. It does not justify skipping semantic verification or inventing labels.

## Transferable principles for HeraVision

- Cheap proposal plus evidence-backed verification.
- Confidence-dependent amount of computation.
- Measure proposal latency separately from verifier latency.
- Use survival/acceptance statistics to decide whether speculation saves time.
- Reject bad proposals without changing the authoritative result.
- Optimize the bottleneck actually measured on the target hardware.

## Non-transferable claims

Speculative decoding provides distribution-preserving acceleration for autoregressive language generation. It does not magically create visual knowledge, solve unknown-object recognition, or guarantee expression understanding. In HeraVision, all semantic claims still require labeled training/evaluation data, region provenance, and abstention when evidence is insufficient.
