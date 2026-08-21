#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--expected-count", type=int, required=True)
    args = parser.parse_args()
    data = json.loads(args.input.read_text())
    assert data["count"] == len(data["samples"]) == args.expected_count
    seen = set()
    total_boxes = 0
    for sample in data["samples"]:
        assert sample["id"] not in seen
        seen.add(sample["id"])
        path = Path(sample["image_path"])
        assert path.exists(), path
        assert sha256(path) == sample["sha256"], sample["id"]
        for box in sample["ground_truth"]:
            total_boxes += 1
            assert box["w"] > 0 and box["h"] > 0
            assert box["x"] >= 0 and box["y"] >= 0
            assert box["x"] + box["w"] <= sample["width"] + 1e-6
            assert box["y"] + box["h"] <= sample["height"] + 1e-6
    print(json.dumps({"status": "valid-eval-manifest", "samples": len(seen), "ground_truth_boxes": total_boxes}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

