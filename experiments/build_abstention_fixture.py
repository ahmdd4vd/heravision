#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--annotation", type=Path, required=True)
    parser.add_argument("--review", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    annotation = json.loads(args.annotation.read_text())
    review = json.loads(args.review.read_text())
    statuses = {sample["sample_id"]: sample.get("second_review", {}).get("status", "unreviewed") for sample in review["samples"]}
    samples = []
    for source in annotation["samples"]:
        review_status = statuses.get(source["id"], "unreviewed")
        case_type = {
            "accepted-provisional": "candidate-answerable",
            "needs-review": "candidate-ambiguous",
            "accepted-ignore": "candidate-no-structure",
        }.get(review_status, "candidate-unreviewed")
        samples.append({
            "id": source["id"],
            "image_path": source["image_path"],
            "domain": source["domain"],
            "width": source["width"],
            "height": source["height"],
            "sha256": source["sha256"],
            "review_status": review_status,
            "candidate_case_type": case_type,
            "answerability": "pending-independent-review",
            "expected_answer_status": "",
            "reviewer_notes": "",
        })
    output = {
        "name": "md1-abstention-fixture-30",
        "description": "Candidate abstention fixture. Review status is not an answerability label.",
        "annotation_status": "pending-independent-answerability-review",
        "independence": "not-yet-reviewed",
        "count": len(samples),
        "samples": samples,
    }
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    print(json.dumps({"status": output["annotation_status"], "samples": len(samples)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

