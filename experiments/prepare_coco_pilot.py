#!/usr/bin/env python3
"""Prepare a reproducible COCO validation pilot for HeraVision Next.

The script downloads official COCO validation assets, selects image IDs by a
stable evenly-spaced rule, verifies dimensions, and writes a manifest with
SHA-256 digests. It intentionally does not download or execute any code from
the dataset archive.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
import urllib.request
from pathlib import Path

IMAGES_URL = "https://images.cocodataset.org/zips/val2017.zip"
ANNOTATIONS_URL = "https://images.cocodataset.org/annotations/annotations_trainval2017.zip"


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def download(url: str, destination: Path) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    if destination.exists() and destination.stat().st_size > 0:
        print(f"[skip] {destination}")
        return
    print(f"[get] {url}")
    request = urllib.request.Request(url, headers={"User-Agent": "heravision-next/0.1"})
    with urllib.request.urlopen(request, timeout=60) as response, destination.open("wb") as handle:
        total = response.headers.get("Content-Length")
        written = 0
        while True:
            chunk = response.read(1024 * 1024)
            if not chunk:
                break
            handle.write(chunk)
            written += len(chunk)
            if total:
                print(f"\r      {written / int(total):6.1%}", end="", flush=True)
        if total:
            print()


def evenly_spaced(items: list[dict], count: int) -> list[dict]:
    if count <= 0 or count > len(items):
        raise ValueError(f"count must be in [1, {len(items)}]")
    if count == 1:
        return [items[len(items) // 2]]
    positions = [round(i * (len(items) - 1) / (count - 1)) for i in range(count)]
    return [items[position] for position in positions]


def read_dimensions(path: Path) -> tuple[int, int]:
    try:
        from PIL import Image
    except ImportError as exc:
        raise RuntimeError("Pillow is required to verify image dimensions") from exc
    with Image.open(path) as image:
        image.verify()
    with Image.open(path) as image:
        return image.width, image.height


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path("data/coco"))
    parser.add_argument("--count", type=int, default=300)
    parser.add_argument("--manifest", type=Path, default=Path("experiments/manifests/coco-pilot-300.json"))
    args = parser.parse_args()

    root = args.root
    archive_dir = root / "archives"
    image_dir = root / "val2017"
    annotation_dir = root / "annotations"
    image_zip = archive_dir / "val2017.zip"
    annotation_zip = archive_dir / "annotations_trainval2017.zip"
    download(IMAGES_URL, image_zip)
    download(ANNOTATIONS_URL, annotation_zip)

    import zipfile

    image_dir.mkdir(parents=True, exist_ok=True)
    annotation_dir.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(image_zip) as archive:
        archive.extractall(root)
    with zipfile.ZipFile(annotation_zip) as archive:
        archive.extractall(root)

    annotation_path = root / "annotations" / "instances_val2017.json"
    if not annotation_path.exists():
        raise FileNotFoundError(annotation_path)
    with annotation_path.open(encoding="utf-8") as handle:
        payload = json.load(handle)
    images = sorted(payload["images"], key=lambda image: int(image["id"]))
    selected = evenly_spaced(images, args.count)

    samples = []
    for image in selected:
        filename = image["file_name"]
        path = image_dir / filename
        if not path.exists():
            raise FileNotFoundError(path)
        width, height = read_dimensions(path)
        if width != image["width"] or height != image["height"]:
            raise ValueError(f"dimension mismatch for {filename}")
        samples.append(
            {
                "id": f"coco-val-{int(image['id']):012d}",
                "image_path": str(path),
                "annotation_path": str(annotation_path),
                "split": "pilot",
                "tags": ["coco", "photo", "bbox-groundtruth"],
                "sha256": sha256(path),
                "width": width,
                "height": height,
                "source_url": f"https://images.cocodataset.org/val2017/{filename}",
                "source_image_id": int(image["id"]),
            }
        )

    args.manifest.parent.mkdir(parents=True, exist_ok=True)
    manifest = {
        "name": f"coco-val-pilot-{args.count}",
        "description": "Deterministic evenly-spaced COCO 2017 validation pilot for B0/B1 geometry evaluation.",
        "dataset_url": "https://cocodataset.org/",
        "images_url": IMAGES_URL,
        "annotations_url": ANNOTATIONS_URL,
        "count": len(samples),
        "samples": samples,
    }
    args.manifest.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    print(f"[ok] wrote {args.manifest} with {len(samples)} samples")
    return 0


if __name__ == "__main__":
    sys.exit(main())
