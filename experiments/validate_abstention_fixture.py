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
    assert data["annotation_status"] == "pending-independent-answerability-review"
    ids = set()
    for sample in data["samples"]:
        assert sample["id"] not in ids
        ids.add(sample["id"])
        path = Path(sample["image_path"])
        assert path.exists(), path
        assert sha256(path) == sample["sha256"], sample["id"]
        assert sample["expected_answer_status"] == ""
    print(json.dumps({"status": "valid-pending-abstention-fixture", "samples": len(ids)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

