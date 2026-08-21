#!/usr/bin/env python3
"""Create a verified HeraVision manifest from the official COCO128 archive."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


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


def read_labels(label_path: Path, width: int, height: int) -> list[dict]:
    boxes = []
    for line_no, line in enumerate(label_path.read_text(encoding="utf-8").splitlines(), 1):
        fields = line.split()
        if len(fields) != 5:
            raise ValueError(f"{label_path}:{line_no}: expected 5 fields")
        class_id, cx, cy, bw, bh = int(fields[0]), *(float(v) for v in fields[1:])
        if not (0 <= cx <= 1 and 0 <= cy <= 1 and 0 < bw <= 1 and 0 < bh <= 1):
            raise ValueError(f"{label_path}:{line_no}: normalized box out of range")
        x = (cx - bw / 2) * width
        y = (cy - bh / 2) * height
        boxes.append({
            "class_id": class_id,
            "x": round(x, 4),
            "y": round(y, 4),
            "w": round(bw * width, 4),
            "h": round(bh * height, 4),
        })
    return boxes


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path("data/coco128/coco128"))
    parser.add_argument("--manifest", type=Path, default=Path("experiments/manifests/coco128-verified.json"))
    args = parser.parse_args()

    image_dir = args.root / "images" / "train2017"
    label_dir = args.root / "labels" / "train2017"
    image_paths = sorted(image_dir.glob("*.jpg"))
    if len(image_paths) != 128:
        raise ValueError(f"expected 128 COCO128 images, found {len(image_paths)}")
    label_paths = sorted(label_dir.glob("*.txt"))
    image_stems = {path.stem for path in image_paths}
    label_stems = {path.stem for path in label_paths}
    common_stems = sorted(image_stems & label_stems)
    excluded = {
        "images_without_labels": sorted(image_stems - label_stems),
        "labels_without_images": sorted(label_stems - image_stems),
    }
    samples = []
    seen_hashes = set()
    for image_id in common_stems:
        image_path = image_dir / f"{image_id}.jpg"
        label_path = label_dir / f"{image_id}.txt"
        width, height = dimensions(image_path)
        digest = sha256(image_path)
        if digest in seen_hashes:
            raise ValueError(f"duplicate image hash: {image_path}")
        seen_hashes.add(digest)
        boxes = read_labels(label_path, width, height)
        samples.append({
            "id": f"coco128-{image_id}",
            "image_path": str(image_path),
            "annotation_path": str(label_path),
            "split": "pilot",
            "tags": ["coco128", "photo", "bbox-groundtruth"],
            "sha256": digest,
            "width": width,
            "height": height,
            "ground_truth": boxes,
            "source_url": "https://github.com/ultralytics/assets/releases/download/v0.0.0/coco128.zip",
        })
    args.manifest.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "name": "coco128-verified",
        "description": "Verified COCO128 detection pilot; image hashes and YOLO boxes are recorded.",
        "source": "https://docs.ultralytics.com/datasets/detect/coco128",
        "archive_sha256": sha256(args.root.parent / "coco128.zip"),
        "count": len(samples),
        "excluded": excluded,
        "samples": samples,
    }
    args.manifest.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    print(f"[ok] verified {len(samples)} images -> {args.manifest}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
