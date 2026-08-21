#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import defaultdict
from pathlib import Path


def load_rows(path: Path) -> dict[str, dict]:
    return {row["sample_id"]: row for row in (json.loads(line) for line in path.read_text().splitlines() if line.strip())}


def aggregate(manifest: dict, rows: dict[str, dict]) -> dict:
    groups = defaultdict(list)
    for sample in manifest["samples"]:
        groups[sample["domain"]].append(rows.get(sample["id"], {"status": "missing"}))
    result = {}
    for domain, items in sorted(groups.items()):
        completed = [item for item in items if item.get("status") == "ok"]
        mean = lambda values: sum(values) / len(values) if values else 0.0
        result[domain] = {
            "samples": len(items),
            "completed": len(completed),
            "errors": len(items) - len(completed),
            "mean_b1_regions": mean([len(item["b1"]["graph"].get("nodes") or []) for item in completed]),
            "mean_b1_edges": mean([len(item["b1"]["graph"].get("edges") or []) for item in completed]),
            "mean_b1_ms": mean([item["b1"]["elapsed_ms"] for item in completed]),
            "mean_coverage": mean([item["geometry"]["coverage_a"] for item in completed]),
            "mean_matched_iou": mean([item["geometry"]["mean_iou"] for item in completed]),
        }
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--reference", type=Path, required=True)
    parser.add_argument("--candidate", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    result = {
        "manifest": str(args.manifest),
        "reference": aggregate(manifest, load_rows(args.reference / "predictions.jsonl")),
        "candidate": aggregate(manifest, load_rows(args.candidate / "predictions.jsonl")),
    }
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

