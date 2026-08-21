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
    args = parser.parse_args()
    data = json.loads(args.input.read_text())
    assert data["count"] == len(data["samples"]) == 30
    assert data["annotation_status"] == "provisional-review"
    sample_ids, region_ids = set(), set()
    total_regions = 0
    for sample in data["samples"]:
        assert sample["sample_id"] not in sample_ids
        sample_ids.add(sample["sample_id"])
        path = Path(sample["image_path"])
        assert sha256(path) == sample["sha256"], sample["sample_id"]
        for region in sample["regions"]:
            total_regions += 1
            assert region["id"] not in region_ids
            region_ids.add(region["id"])
            box = region["bbox"]
            assert box["x"] >= 0 and box["y"] >= 0 and box["w"] > 0 and box["h"] > 0
            assert box["x"] + box["w"] <= sample["width"] + 1e-6
            assert box["y"] + box["h"] <= sample["height"] + 1e-6
            assert region["source"] == "analyst-visual-review"
            assert region["status"] == "provisional-review"
            assert region["region_type"] in {"main_object", "main_object_partial", "scene_structure", "whole_diagram"}
    print(json.dumps({"samples": len(sample_ids), "regions": total_regions, "unique_region_ids": len(region_ids), "status": "valid-provisional"}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
