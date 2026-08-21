#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--accepted-only", action="store_true")
    args = parser.parse_args()
    data = json.loads(args.input.read_text())
    samples = []
    for sample in data["samples"]:
        if args.accepted_only and sample.get("second_review", {}).get("status") != "accepted-provisional":
            continue
        samples.append({
            "id": sample["sample_id"],
            "image_path": sample["image_path"],
            "width": sample["width"],
            "height": sample["height"],
            "domain": sample["domain"],
            "sha256": sample["sha256"],
            "ground_truth": [region["bbox"] for region in sample["regions"] if region["visibility"] != "uncertain"],
        })
    payload = {"name": "md1-ground-truth-provisional", "count": len(samples), "samples": samples}
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(payload, indent=2) + "\n")
    print(json.dumps({"samples": len(samples), "ground_truth_boxes": sum(len(s["ground_truth"]) for s in samples)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
