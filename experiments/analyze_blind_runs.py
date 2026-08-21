#!/usr/bin/env python3
"""Aggregate blind MD0 raw and filtered runs by domain."""

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
        errors = [item for item in items if item.get("status") != "ok"]
        def mean(values): return sum(values) / len(values) if values else 0.0
        result[domain] = {
            "samples": len(items),
            "completed": len(completed),
            "errors": len(errors),
            "mean_b0_regions": mean([len(item["b0"]["graph"].get("nodes") or []) for item in completed]),
            "mean_b1_regions": mean([len(item["b1"]["graph"].get("nodes") or []) for item in completed]),
            "mean_b0_edges": mean([len(item["b0"]["graph"].get("edges") or []) for item in completed]),
            "mean_b1_edges": mean([len(item["b1"]["graph"].get("edges") or []) for item in completed]),
            "mean_b0_ms": mean([item["b0"]["elapsed_ms"] for item in completed]),
            "mean_b1_ms": mean([item["b1"]["elapsed_ms"] for item in completed]),
            "mean_b0_b1_coverage": mean([item["geometry"]["coverage_a"] for item in completed]),
            "mean_matched_iou": mean([item["geometry"]["mean_iou"] for item in completed]),
            "failure_ids": [item.get("sample_id", "") for item in errors],
        }
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--raw", type=Path, required=True)
    parser.add_argument("--filtered", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    raw_rows = load_rows(args.raw / "predictions.jsonl")
    filtered_rows = load_rows(args.filtered / "predictions.jsonl")
    output = {
        "manifest": str(args.manifest),
        "blind_policy": manifest.get("blind_policy", ""),
        "raw": aggregate(manifest, raw_rows),
        "filtered": aggregate(manifest, filtered_rows),
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    print(json.dumps(output, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
