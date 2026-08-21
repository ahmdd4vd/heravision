#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--log", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    log = json.loads(args.log.read_text())
    assert manifest["count"] == len(manifest["samples"]) == 30
    assert manifest["annotation_status"] == "second-review-complete-pending-independent-human"
    statuses = {"accepted-provisional", "needs-review", "accepted-ignore"}
    ids = set()
    for sample in manifest["samples"]:
        assert sample["sample_id"] not in ids
        ids.add(sample["sample_id"])
        status = sample["second_review"]["status"]
        assert status in statuses
        assert sample["second_review"]["reviewer"] == "Manus AI"
        if status == "needs-review":
            for region in sample["regions"]:
                assert region["second_review"] == "needs-review"
                assert region["second_review_category"] != "none"
    counts = log["counts"]
    assert counts["accepted"] + counts["needs_review"] + counts["accepted_ignore"] == 30
    assert len(log["disagreements"]) == counts["needs_review"]
    print(json.dumps({"status": "valid-second-review", "samples": len(ids), "counts": counts}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
