#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--predictions", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    rows = {json.loads(line)["sample_id"]: json.loads(line) for line in args.predictions.read_text().splitlines() if line.strip()}
    samples = []
    total_edges = 0
    for sample in manifest["samples"]:
        row = rows.get(sample["id"])
        if not row or row.get("status") != "ok":
            continue
        graph = row["b1"]["graph"]
        nodes = {node["id"]: node for node in graph.get("nodes", [])}
        edges = []
        for edge in graph.get("edges") or []:
            total_edges += 1
            edges.append({
                "from": edge["from"],
                "to": edge["to"],
                "predicate": edge["predicate"],
                "score": edge.get("score", 0),
                "from_bbox": nodes.get(edge["from"], {}).get("region", {}).get("bbox"),
                "to_bbox": nodes.get(edge["to"], {}).get("region", {}).get("bbox"),
                "review_decision": "",
                "review_notes": "",
            })
        samples.append({
            "sample_id": sample["id"],
            "image_path": sample["image_path"],
            "domain": sample["domain"],
            "edges": edges,
        })
    output = {
        "name": "md1-relation-review-fixture",
        "description": "Independent review template for visible relation correctness. This evaluates predicted-edge precision only; missed relations require separate annotation.",
        "annotation_status": "pending-independent-relation-review",
        "independence": "not-yet-reviewed",
        "samples": samples,
        "total_predicted_edges": total_edges,
    }
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    print(json.dumps({"status": output["annotation_status"], "samples": len(samples), "predicted_edges": total_edges}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

