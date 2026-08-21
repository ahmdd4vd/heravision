#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from pathlib import Path

from PIL import Image, ImageDraw


def sheet(samples: list[dict], output: Path) -> None:
    tile_w, tile_h, label_h = 360, 290, 32
    cols = 5
    rows = (len(samples) + cols - 1) // cols
    canvas = Image.new("RGB", (cols * tile_w, rows * (tile_h + label_h)), "white")
    draw = ImageDraw.Draw(canvas)
    for index, sample in enumerate(samples):
        image = Image.open(sample["image_path"]).convert("RGB")
        sx = (tile_w - 12) / image.width
        sy = (tile_h - 12) / image.height
        scale = min(sx, sy)
        resized = image.resize((max(1, round(image.width * scale)), max(1, round(image.height * scale))))
        ox = (tile_w - resized.width) // 2 + (index % cols) * tile_w
        oy = (tile_h - resized.height) // 2 + (index // cols) * (tile_h + label_h)
        canvas.paste(resized, (ox, oy))
        for region in sample["regions"]:
            box = region["bbox"]
            x1 = ox + box["x"] * scale
            y1 = oy + box["y"] * scale
            x2 = x1 + box["w"] * scale
            y2 = y1 + box["h"] * scale
            draw.rectangle((x1, y1, x2, y2), outline=(255, 0, 0), width=3)
        label = f"{sample['sample_id'][-18:]} [{len(sample['regions'])}]"
        draw.text(((index % cols) * tile_w + 5, (index // cols) * (tile_h + label_h) + tile_h + 6), label, fill="black")
    output.parent.mkdir(parents=True, exist_ok=True)
    canvas.save(output, quality=92)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    data = json.loads(args.input.read_text())
    groups = {"imagenette": [], "imagewoof": [], "wikimedia-diagram": []}
    for sample in data["samples"]:
        groups[sample["domain"]].append(sample)
    for domain, samples in groups.items():
        sheet(samples, args.output_dir / f"second-review-{domain}.jpg")
    print(json.dumps({"sheets": len(groups), "samples": len(data["samples"])}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
