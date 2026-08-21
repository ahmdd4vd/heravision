#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

IMAGENETTE_BROAD = {
    "n01440764": "animal",
    "n02979186": "artifact",
    "n03000684": "artifact",
    "n03417042": "vehicle",
    "n03425413": "artifact",
    "n03888257": "artifact",
}

ANSWERABILITY = {
    "answered": "answerable-broad-semantic",
    "abstain": "ambiguous-broad-semantic",
    "insufficient_evidence": "insufficient-evidence",
}


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--answerability", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    source = json.loads(args.input.read_text())
    answer = json.loads(args.answerability.read_text())
    answer_by_id = {s["id"]: s for s in answer["samples"]}
    samples = []
    for item in source["samples"]:
        path = Path(item["image_path"])
        review = answer_by_id.get(item["id"], {})
        status = review.get("expected_answer_status", "")
        if item["domain"] == "imagewoof":
            label = "animal"
            label_source = "dataset-family-metadata-imagewoof"
        elif item["domain"] == "wikimedia-diagram":
            label = "diagram" if item["source_class"] != "File:Blank scan image.jpg" else "unknown"
            label_source = "source-filename-metadata-plus-visual-review"
        else:
            label = IMAGENETTE_BROAD.get(item["source_class"], "unknown")
            label_source = "dataset-class-metadata-mapping"
        samples.append({
            "id": item["id"],
            "image_path": item["image_path"],
            "domain": item["domain"],
            "source_class": item["source_class"],
            "width": item["width"],
            "height": item["height"],
            "sha256": sha256(path),
            "broad_label": label,
            "label_source": label_source,
            "answerability": ANSWERABILITY.get(status, "pending-review"),
            "expected_answer_status": status,
            "regions": [],
            "required_evidence": ["region", "boundary", "scale-stability"],
            "review_status": "same-agent-internal-development-review",
            "independent_review_status": "pending",
            "notes": "Broad label is a development target derived from dataset/source metadata; semantic bbox annotation is pending."
        })
    output = {
        "name": "md2-semantic-broad-fixture-30",
        "description": "Broad semantic pilot fixture built from real MD1 images; labels are metadata-derived development targets and require independent review before headline claims.",
        "count": len(samples),
        "annotation_status": "provisional-metadata-derived",
        "semantic_levels": ["broad-category"],
        "independence": "same-agent internal development fixture; not independent human semantic annotation",
        "samples": samples,
    }
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    print(json.dumps({"count": len(samples), "domains": sorted({s['domain'] for s in samples}), "labels": sorted({s['broad_label'] for s in samples})}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
