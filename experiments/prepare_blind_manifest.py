#!/usr/bin/env python3
"""Build the first 300-image blind multi-domain manifest.

The blind set is intentionally separate from COCO128 development/training
artifacts. It contains 143 Imagenette validation images, 143 Imagewoof
validation images, and 14 Wikimedia Commons diagram/document images.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


WIKIMEDIA_SOURCES = {
    "blank_scan_image.jpg": ("File:Blank scan image.jpg", "https://commons.wikimedia.org/wiki/File:Blank_scan_image.jpg"),
    "capacitor_schematic.png": ("File:Capacitor schematic.svg", "https://commons.wikimedia.org/wiki/File:Capacitor_schematic.svg"),
    "drafting_table.jpg": ("File:Drafting table.jpg", "https://commons.wikimedia.org/wiki/File:Drafting_table.jpg"),
    "dunning_kruger_effect.png": ("File:Dunning–Kruger Effect 01.svg", "https://commons.wikimedia.org/wiki/File:Dunning%E2%80%93Kruger_Effect_01.svg"),
    "flow_chart_loop.png": ("Flow chart", "https://commons.wikimedia.org/wiki/Category:Flowcharts"),
    "flowchart_example.png": ("File:FlowchartExample.png", "https://commons.wikimedia.org/wiki/File:FlowchartExample.png"),
    "global_population_pie_chart.jpg": ("File:World population pie chart.JPG", "https://commons.wikimedia.org/wiki/File:World_population_pie_chart.JPG"),
    "human_eye_schematic.png": ("File:Schematic diagram of the human eye-es.svg", "https://commons.wikimedia.org/wiki/File:Schematic_diagram_of_the_human_eye-es.svg"),
    "image_file_formats.png": ("Wikimedia Commons image file formats result", "https://commons.wikimedia.org/"),
    "refinery_flow.png": ("File:RefineryFlow.png", "https://commons.wikimedia.org/wiki/File:RefineryFlow.png"),
    "sample_network_diagram.png": ("File:Sample-network-diagram.png", "https://commons.wikimedia.org/wiki/File:Sample-network-diagram.png"),
    "snellen_chart.png": ("File:Snellen chart.svg", "https://commons.wikimedia.org/wiki/File:Snellen_chart.svg"),
    "ssme_schematic.png": ("File:Ssme schematic.svg", "https://commons.wikimedia.org/wiki/File:Ssme_schematic.svg"),
    "tracking_data_flow_chart.jpg": ("File:Tracking-data flow chart.jpg", "https://commons.wikimedia.org/wiki/File:Tracking-data_flow_chart.jpg"),
}


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def dimensions(path: Path) -> tuple[int, int]:
    from PIL import Image

    with Image.open(path) as image:
        image.verify()
    with Image.open(path) as image:
        return image.width, image.height


def spaced(paths: list[Path], count: int) -> list[Path]:
    paths = sorted(paths)
    if count > len(paths):
        raise ValueError(f"requested {count}, only {len(paths)} available")
    if count == 1:
        return [paths[len(paths) // 2]]
    indices = [round(i * (len(paths) - 1) / (count - 1)) for i in range(count)]
    return [paths[i] for i in indices]


def balanced_class_sample(root: Path, count: int) -> list[tuple[Path, str]]:
    classes = sorted(path for path in root.iterdir() if path.is_dir())
    if len(classes) != 10:
        raise ValueError(f"expected 10 classes in {root}, found {len(classes)}")
    base, extra = divmod(count, len(classes))
    chosen = []
    for class_index, class_dir in enumerate(classes):
        paths = sorted(class_dir.glob("*.JPEG"))
        take = base + (1 if class_index < extra else 0)
        chosen.extend((path, class_dir.name) for path in spaced(paths, take))
    return chosen


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path("data"))
    parser.add_argument("--wikimedia", type=Path, default=Path("data/wikimedia14"))
    parser.add_argument("--output", type=Path, default=Path("experiments/manifests/blind-md0-300.json"))
    args = parser.parse_args()

    imagenette = balanced_class_sample(args.root / "imagenette/imagenette2-160/val", 143)
    imagewoof = balanced_class_sample(args.root / "imagewoof/imagewoof2-160/val", 143)
    wikimedia = sorted(path for path in args.wikimedia.iterdir() if path.suffix.lower() in {".jpg", ".jpeg", ".png"})
    if len(wikimedia) != 14:
        raise ValueError(f"expected 14 Wikimedia files, found {len(wikimedia)}")

    samples = []
    seen_hashes = set()
    entries = [
        *((path, "imagenette", class_name, "https://github.com/fastai/imagenette") for path, class_name in imagenette),
        *((path, "imagewoof", class_name, "https://github.com/fastai/imagenette") for path, class_name in imagewoof),
        *((path, "wikimedia-diagram", WIKIMEDIA_SOURCES[path.name][0], WIKIMEDIA_SOURCES[path.name][1]) for path in wikimedia),
    ]
    if len(entries) != 300:
        raise AssertionError(len(entries))
    for index, (path, domain, source_class, source_url) in enumerate(entries, 1):
        width, height = dimensions(path)
        digest = sha256(path)
        if digest in seen_hashes:
            raise ValueError(f"duplicate image hash: {path}")
        seen_hashes.add(digest)
        samples.append({
            "id": f"blind-md0-{domain}-{digest[:12]}",
            "image_path": str(path),
            "annotation_path": "",
            "split": "blind",
            "domain": domain,
            "source_class": source_class,
            "annotation_status": "unlabeled",
            "tags": ["blind", "md0", domain],
            "sha256": digest,
            "width": width,
            "height": height,
            "source_url": source_url,
        })

    by_domain = {}
    for sample in samples:
        by_domain[sample["domain"]] = by_domain.get(sample["domain"], 0) + 1
    payload = {
        "name": "blind-md0-300",
        "description": "First blind multi-domain benchmark: 143 Imagenette, 143 Imagewoof, 14 Wikimedia diagram/document images.",
        "blind_policy": "No sample in this manifest was used for B1 threshold tuning or region-filter training.",
        "source_urls": [
            "https://github.com/fastai/imagenette",
            "https://commons.wikimedia.org/",
        ],
        "count": len(samples),
        "by_domain": by_domain,
        "samples": samples,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"count": len(samples), "by_domain": by_domain, "output": str(args.output)}, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
