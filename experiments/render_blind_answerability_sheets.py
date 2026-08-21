#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import defaultdict
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    data = json.loads(args.manifest.read_text())
    groups = defaultdict(list)
    for sample in data["samples"]:
        groups[sample["domain"]].append(sample)
    args.output_dir.mkdir(parents=True, exist_ok=True)
    font = ImageFont.load_default()
    for domain, samples in sorted(groups.items()):
        thumb_w, thumb_h = 260, 210
        cols = 2
        rows = (len(samples) + cols - 1) // cols
        sheet = Image.new("RGB", (cols * thumb_w, rows * thumb_h), "white")
        draw = ImageDraw.Draw(sheet)
        for i, sample in enumerate(samples):
            img = Image.open(sample["image_path"]).convert("RGB")
            img.thumbnail((thumb_w - 12, thumb_h - 35))
            x = (i % cols) * thumb_w + (thumb_w - img.width) // 2
            y = (i // cols) * thumb_h + 4
            sheet.paste(img, (x, y))
            label = f"{i+1}: {sample['id'][-12:]}"
            draw.text(((i % cols) * thumb_w + 6, (i // cols) * thumb_h + thumb_h - 24), label, fill="black", font=font)
        path = args.output_dir / f"blind-answerability-{domain}.jpg"
        sheet.save(path, quality=94)
        print(path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

