#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path


ALLOWED = {"", "correct", "incorrect", "uncertain"}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    args = parser.parse_args()
    data = json.loads(args.input.read_text())
    assert data["annotation_status"] in {"pending-independent-relation-review", "internal-third-review-complete"}
    edge_count = 0
    for sample in data["samples"]:
        assert Path(sample["image_path"]).exists()
        for edge in sample["edges"]:
            edge_count += 1
            assert edge["from"] and edge["to"] and edge["predicate"]
            assert edge["from_bbox"] and edge["to_bbox"]
            assert edge["review_decision"] in ALLOWED
    assert edge_count == data["total_predicted_edges"]
    print(json.dumps({"status": "valid-pending-relation-fixture", "samples": len(data["samples"]), "edges": edge_count}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

