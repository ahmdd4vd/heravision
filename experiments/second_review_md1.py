#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path

# Same-agent second pass: deliberately explicit that this is not an independent human labeler.
DECISIONS = {
    "blind-md0-imagenette-140e47321913": ("needs-review", "scene box includes city/background; object-vs-scene scope is ambiguous", "scope-ambiguity"),
    "blind-md0-imagenette-3235df644822": ("needs-review", "foreground person and held fish create two plausible targets", "multiple-foreground"),
    "blind-md0-imagenette-64159a40c4de": ("needs-review", "person and large board/equipment overlap the candidate box", "occlusion-scope"),
    "blind-md0-imagenette-a06d12b2ffd7": ("needs-review", "person and equipment are visually connected but object scope is uncertain", "scope-ambiguity"),
    "blind-md0-imagewoof-007c38e50b6b": ("needs-review", "two dogs share one candidate region", "multiple-foreground"),
    "blind-md0-imagewoof-82cb64a1114a": ("needs-review", "dog and human overlap; foreground ownership needs reviewer agreement", "occlusion-scope"),
    "blind-md0-wikimedia-diagram-874b0038f0f7": ("needs-review", "whole-diagram box includes network whitespace and connector extent", "diagram-scope"),
}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--log", type=Path, required=True)
    args = parser.parse_args()
    data = json.loads(args.input.read_text())
    disagreements = []
    accepted = 0
    needs_review = 0
    ignore = 0
    for sample in data["samples"]:
        sample_id = sample["sample_id"]
        if not sample["regions"]:
            sample["second_review"] = {
                "status": "accepted-ignore",
                "reason": "no visible structure in blank scan",
                "reviewer": "Manus AI",
                "independence": "same-agent-second-pass",
            }
            ignore += 1
            continue
        decision, reason, category = DECISIONS.get(sample_id, ("accepted-provisional", "box scope is visually consistent with declared region_type", "none"))
        sample["second_review"] = {
            "status": decision,
            "reason": reason,
            "reviewer": "Manus AI",
            "independence": "same-agent-second-pass",
        }
        for region in sample["regions"]:
            region["second_review"] = decision
            region["second_review_category"] = category
        if decision == "needs-review":
            needs_review += 1
            disagreements.append({
                "sample_id": sample_id,
                "region_ids": [region["id"] for region in sample["regions"]],
                "category": category,
                "provisional_status": sample["annotation_status"],
                "second_status": decision,
                "reason": reason,
                "reviewer": "Manus AI",
                "independence": "same-agent-second-pass; independent human confirmation still required",
            })
        else:
            accepted += 1
    output = dict(data)
    output["name"] = "md1-ground-truth-regions-second-review"
    output["annotation_status"] = "second-review-complete-pending-independent-human"
    output["second_review"] = {
        "reviewer": "Manus AI",
        "independence": "same-agent-second-pass",
        "warning": "This pass is not an independent human annotation. Needs-review samples must be resolved before final benchmark claims.",
        "accepted_samples": accepted,
        "needs_review_samples": needs_review,
        "accepted_ignore_samples": ignore,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    args.log.parent.mkdir(parents=True, exist_ok=True)
    args.log.write_text(json.dumps({
        "name": "md1-disagreement-log",
        "reviewer": "Manus AI",
        "independence": "same-agent-second-pass",
        "disagreements": disagreements,
        "counts": {"accepted": accepted, "needs_review": needs_review, "accepted_ignore": ignore},
    }, indent=2) + "\n")
    print(json.dumps({"accepted": accepted, "needs_review": needs_review, "accepted_ignore": ignore, "disagreements": len(disagreements)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
