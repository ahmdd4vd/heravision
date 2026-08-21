#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import defaultdict
from pathlib import Path


def metrics(tp: int, fp: int, fn: int) -> dict[str, float | int]:
    precision = tp / (tp + fp) if tp + fp else 0.0
    recall = tp / (tp + fn) if tp + fn else 0.0
    f1 = 2 * precision * recall / (precision + recall) if precision + recall else 0.0
    return {"tp": tp, "fp": fp, "fn": fn, "precision": precision, "recall": recall, "f1": f1}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--summary", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    summary = json.loads(args.summary.read_text())
    domains = {s["id"]: s["domain"] for s in manifest["samples"]}
    aggregate = defaultdict(lambda: defaultdict(lambda: {"tp": 0, "fp": 0, "fn": 0, "pred_count": 0, "gt_count": 0}))
    for sample in summary["per_sample"]:
        domain = domains[sample["sample_id"]]
        for engine in ("b0", "b1"):
            row = sample[engine]
            dest = aggregate[domain][engine]
            for key in ("tp", "fp", "fn", "pred_count", "gt_count"):
                dest[key] += row[key]
    result = {"manifest": str(args.manifest), "iou_threshold": summary["iou_threshold"], "samples": manifest["count"], "domains": {}}
    for domain, engines in aggregate.items():
        result["domains"][domain] = {}
        for engine, row in engines.items():
            result["domains"][domain][engine] = {**metrics(row["tp"], row["fp"], row["fn"]), "pred_count": row["pred_count"], "gt_count": row["gt_count"]}
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

