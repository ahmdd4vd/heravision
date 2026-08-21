#!/usr/bin/env python3
"""Select a deterministic MD1 annotation subset and failure-gallery samples."""

from __future__ import annotations

import argparse
import json
from collections import defaultdict
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


def spaced(items: list[dict], count: int) -> list[dict]:
    items = sorted(items, key=lambda item: item["id"])
    if count >= len(items):
        return items
    return [items[round(i * (len(items) - 1) / (count - 1))] for i in range(count)]


def make_sheet(samples: list[dict], output: Path, title: str) -> None:
    tile_w, tile_h, label_h = 320, 260, 28
    cols = 5
    rows = (len(samples) + cols - 1) // cols
    sheet = Image.new("RGB", (cols * tile_w, rows * (tile_h + label_h)), "white")
    draw = ImageDraw.Draw(sheet)
    for index, sample in enumerate(samples):
        image = Image.open(sample["image_path"]).convert("RGB")
        image.thumbnail((tile_w - 12, tile_h - 12))
        x = (index % cols) * tile_w + (tile_w - image.width) // 2
        y = (index // cols) * (tile_h + label_h) + (tile_h - image.height) // 2
        sheet.paste(image, (x, y))
        draw.text(((index % cols) * tile_w + 5, (index // cols) * (tile_h + label_h) + tile_h + 4), sample["id"][-18:], fill="black")
    output.parent.mkdir(parents=True, exist_ok=True)
    sheet.save(output, quality=90)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--raw", type=Path, required=True)
    parser.add_argument("--output-manifest", type=Path, required=True)
    parser.add_argument("--gallery-dir", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    rows = {json.loads(line)["sample_id"]: json.loads(line) for line in args.raw.read_text().splitlines() if line.strip()}
    groups = defaultdict(list)
    for sample in manifest["samples"]:
        groups[sample["domain"]].append(sample)
    subset = []
    for domain in sorted(groups):
        subset.extend(spaced(groups[domain], 10))
    subset_payload = {
        "name": "md1-annotation-30",
        "description": "Deterministic annotation subset: 10 images per blind domain.",
        "parent_manifest": str(args.manifest),
        "annotation_status": "pending_manual_or_reviewer_annotation",
        "count": len(subset),
        "samples": subset,
    }
    args.output_manifest.parent.mkdir(parents=True, exist_ok=True)
    args.output_manifest.write_text(json.dumps(subset_payload, indent=2) + "\n")

    ranked = []
    for sample in manifest["samples"]:
        row = rows[sample["id"]]
        if row.get("status") != "ok":
            ranked.append((10**9, sample))
            continue
        b1_count = len(row["b1"]["graph"].get("nodes") or [])
        coverage = row["geometry"].get("coverage_a", 0.0)
        ranked.append((b1_count - 1000 * coverage, sample))
    failures = [sample for _, sample in sorted(ranked, key=lambda item: (item[0], item[1]["id"]), reverse=True)[:30]]
    args.gallery_dir.mkdir(parents=True, exist_ok=True)
    make_sheet(subset[:10], args.gallery_dir / "annotation-imagenette.jpg", "Imagenette annotation subset")
    make_sheet(subset[10:20], args.gallery_dir / "annotation-imagewoof.jpg", "Imagewoof annotation subset")
    make_sheet(subset[20:30], args.gallery_dir / "annotation-wikimedia.jpg", "Wikimedia annotation subset")
    make_sheet(failures[:15], args.gallery_dir / "failure-gallery-01.jpg", "Failure candidates")
    make_sheet(failures[15:30], args.gallery_dir / "failure-gallery-02.jpg", "Failure candidates")
    (args.gallery_dir / "failure-gallery.json").write_text(json.dumps([sample["id"] for sample in failures], indent=2) + "\n")
    print(json.dumps({"annotation_count": len(subset), "failure_gallery_count": len(failures), "output": str(args.output_manifest)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
