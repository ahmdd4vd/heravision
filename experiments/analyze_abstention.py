#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import Counter, defaultdict
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--predictions", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    domains = {sample["id"]: sample.get("domain", "unknown") for sample in manifest["samples"]}
    all_counts = Counter()
    by_domain = defaultdict(Counter)
    errors = []
    rows = [json.loads(line) for line in args.predictions.read_text().splitlines() if line.strip()]
    for row in rows:
        if row.get("status") != "ok":
            errors.append({"sample_id": row.get("sample_id", ""), "error": row.get("error", "")})
            continue
        domain = domains.get(row["sample_id"], "unknown")
        status = row["b1"].get("answer", {}).get("status", "missing")
        all_counts[status] += 1
        by_domain[domain][status] += 1
    result = {
        "manifest": str(args.manifest),
        "predictions": str(args.predictions),
        "samples": len(rows),
        "errors": errors,
        "overall": dict(sorted(all_counts.items())),
        "by_domain": {domain: dict(sorted(counts.items())) for domain, counts in sorted(by_domain.items())},
    }
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

