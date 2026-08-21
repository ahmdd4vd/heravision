#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import Counter
from pathlib import Path


def analyze(path: Path) -> dict:
    total_edges = 0
    samples = 0
    errors = 0
    predicates = Counter()
    edge_counts = []
    for line in path.read_text().splitlines():
        if not line.strip():
            continue
        row = json.loads(line)
        samples += 1
        if row.get("status") != "ok":
            errors += 1
            continue
        edges = row["b1"]["graph"].get("edges") or []
        edge_counts.append(len(edges))
        total_edges += len(edges)
        predicates.update(edge["predicate"] for edge in edges)
    return {
        "predictions": str(path),
        "samples": samples,
        "errors": errors,
        "total_b1_edges": total_edges,
        "mean_b1_edges": total_edges / len(edge_counts) if edge_counts else 0.0,
        "predicates": dict(sorted(predicates.items())),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("predictions", type=Path, nargs="+")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    result = [analyze(path) for path in args.predictions]
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

