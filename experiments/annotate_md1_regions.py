#!/usr/bin/env python3
"""Create auditable provisional region annotations for the 30 MD1 fixtures.

These are analyst-reviewed provisional boxes, not independent human ground
truth. Every record carries method/confidence/status so later review can
replace or accept it without pretending the provenance is stronger than it is.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

# Normalized x, y, w, h boxes reviewed from the MD1 contact sheets.
BOXES = {
    "blind-md0-imagenette-0020e126190f": [(0.06, 0.15, 0.87, 0.62, "main_object", "high")],
    "blind-md0-imagenette-140e47321913": [(0.02, 0.12, 0.96, 0.72, "scene_structure", "medium")],
    "blind-md0-imagenette-3235df644822": [(0.25, 0.06, 0.56, 0.88, "main_object", "medium")],
    "blind-md0-imagenette-4c34e409eec9": [(0.35, 0.02, 0.30, 0.86, "main_object", "medium")],
    "blind-md0-imagenette-64159a40c4de": [(0.10, 0.02, 0.78, 0.95, "main_object", "medium")],
    "blind-md0-imagenette-81ae723b8b73": [(0.05, 0.10, 0.88, 0.82, "main_object", "medium")],
    "blind-md0-imagenette-a06d12b2ffd7": [(0.18, 0.03, 0.72, 0.94, "main_object", "medium")],
    "blind-md0-imagenette-bd0a96d53e4f": [(0.18, 0.27, 0.64, 0.46, "main_object", "high")],
    "blind-md0-imagenette-e757dbe33e0e": [(0.23, 0.20, 0.56, 0.64, "main_object", "high")],
    "blind-md0-imagenette-fff82b979667": [(0.04, 0.18, 0.92, 0.68, "main_object", "medium")],
    "blind-md0-imagewoof-007c38e50b6b": [(0.10, 0.16, 0.78, 0.70, "main_object", "high")],
    "blind-md0-imagewoof-24fca83147eb": [(0.02, 0.13, 0.96, 0.70, "main_object_partial", "medium")],
    "blind-md0-imagewoof-5089b36710c6": [(0.12, 0.12, 0.78, 0.75, "main_object", "high")],
    "blind-md0-imagewoof-5f83d078b0f8": [(0.05, 0.04, 0.86, 0.92, "main_object", "high")],
    "blind-md0-imagewoof-82cb64a1114a": [(0.34, 0.14, 0.60, 0.76, "main_object", "medium")],
    "blind-md0-imagewoof-a42c41aa9865": [(0.04, 0.30, 0.70, 0.58, "main_object", "medium")],
    "blind-md0-imagewoof-b8d750c4af6e": [(0.18, 0.20, 0.60, 0.64, "main_object", "medium")],
    "blind-md0-imagewoof-ce2d28126f82": [(0.28, 0.28, 0.50, 0.46, "main_object", "high")],
    "blind-md0-imagewoof-ea7ac7e143bc": [(0.22, 0.15, 0.62, 0.75, "main_object", "high")],
    "blind-md0-imagewoof-fee3361f1c41": [(0.13, 0.08, 0.72, 0.84, "main_object", "high")],
    "blind-md0-wikimedia-diagram-0c2b052b49a2": [(0.10, 0.03, 0.78, 0.94, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-2a6e30f0a1d1": [(0.08, 0.02, 0.84, 0.96, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-384fecf1094e": [(0.04, 0.04, 0.92, 0.90, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-47b1925795f7": [],
    "blind-md0-wikimedia-diagram-6b552e18e2d1": [(0.01, 0.01, 0.98, 0.98, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-7064f8675350": [(0.08, 0.03, 0.84, 0.94, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-80b81c884f3b": [(0.04, 0.04, 0.92, 0.92, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-874b0038f0f7": [(0.03, 0.14, 0.94, 0.72, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-c07bc8486d16": [(0.01, 0.01, 0.98, 0.98, "whole_diagram", "high")],
    "blind-md0-wikimedia-diagram-def1a5faff64": [(0.02, 0.04, 0.96, 0.90, "whole_diagram", "high")],
}


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    samples = []
    for sample in manifest["samples"]:
        sample_id = sample["id"]
        if sample_id not in BOXES:
            raise ValueError(f"missing reviewed fixture: {sample_id}")
        image_path = Path(sample["image_path"])
        if sha256(image_path) != sample["sha256"]:
            raise ValueError(f"hash mismatch: {sample_id}")
        regions = []
        for index, (x, y, w, h, region_type, confidence) in enumerate(BOXES[sample_id], 1):
            regions.append({
                "id": f"{sample_id}-gt-{index:02d}",
                "bbox": {
                    "x": round(x * sample["width"], 3),
                    "y": round(y * sample["height"], 3),
                    "w": round(w * sample["width"], 3),
                    "h": round(h * sample["height"], 3),
                },
                "region_type": region_type,
                "visibility": "visible" if confidence != "low" else "uncertain",
                "confidence": confidence,
                "status": "provisional-review",
                "source": "analyst-visual-review",
            })
        samples.append({
            "sample_id": sample_id,
            "image_path": sample["image_path"],
            "sha256": sample["sha256"],
            "width": sample["width"],
            "height": sample["height"],
            "domain": sample["domain"],
            "annotation_status": "provisional-review",
            "reviewer": "Manus AI",
            "review_method": "contact-sheet visual review; not independent human annotation",
            "ignore_reason": "no visible structure" if not regions else "",
            "regions": regions,
        })
    output = {
        "name": "md1-ground-truth-regions-provisional",
        "parent_manifest": str(args.manifest),
        "count": len(samples),
        "annotation_status": "provisional-review",
        "ground_truth_warning": "These are provisional analyst-reviewed boxes and must receive independent human/reviewer verification before final blind metrics.",
        "samples": samples,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    print(json.dumps({"samples": len(samples), "regions": sum(len(sample["regions"]) for sample in samples), "output": str(args.output)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
