#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--runs", type=Path, nargs="+")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    result = []
    for run in args.runs:
        summary = json.loads((run / "summary.json").read_text())
        statuses = {}
        rows = [json.loads(line) for line in (run / "predictions.jsonl").read_text().splitlines() if line.strip()]
        for row in rows:
            if row.get("status") != "ok":
                continue
            status = row["b1"].get("answer", {}).get("status", "missing")
            statuses[status] = statuses.get(status, 0) + 1
        result.append({
            "run": str(run),
            "max_side": summary.get("config", {}).get("max_side"),
            "samples": summary["samples"],
            "completed": summary["completed"],
            "errors": summary["errors"],
            "b1_regions": summary["b1_regions"],
            "mean_b1_ms": summary["mean_b1_ms"],
            "answer_statuses": statuses,
        })
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

