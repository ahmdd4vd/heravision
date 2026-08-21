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
    parser.add_argument("--predictions", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    rows = {json.loads(line)["sample_id"]: json.loads(line) for line in args.predictions.read_text().splitlines() if line.strip()}
    groups = defaultdict(list)
    for sample in manifest["samples"]:
        row = rows.get(sample["id"])
        if row and row.get("status") == "ok":
            groups[sample["domain"]].append((sample, row))
    args.output_dir.mkdir(parents=True, exist_ok=True)
    font = ImageFont.load_default()
    for domain, entries in sorted(groups.items()):
        panel_w, panel_h = 420, 300
        cols = 2
        rows_n = (len(entries) + cols - 1) // cols
        sheet = Image.new("RGB", (cols * panel_w, rows_n * panel_h), "white")
        draw = ImageDraw.Draw(sheet)
        for i, (sample, row) in enumerate(entries):
            img = Image.open(sample["image_path"]).convert("RGB")
            scale = min((panel_w - 12) / img.width, 0.68 * panel_h / img.height)
            nw, nh = max(1, int(img.width * scale)), max(1, int(img.height * scale))
            img = img.resize((nw, nh))
            x0 = (i % cols) * panel_w + (panel_w - nw) // 2
            y0 = (i // cols) * panel_h + 4
            sheet.paste(img, (x0, y0))
            sx, sy = nw / sample["width"], nh / sample["height"]
            nodes = {n["id"]: n for n in row["b1"]["graph"].get("nodes", [])}
            edges = row["b1"]["graph"].get("edges") or []
            for node_id, node in nodes.items():
                b = node["region"]["bbox"]
                bx = x0 + int(b["x"] * sx); by = y0 + int(b["y"] * sy)
                bw = int(b["w"] * sx); bh = int(b["h"] * sy)
                draw.rectangle((bx, by, bx + bw, by + bh), outline="red", width=1)
                draw.text((bx, by), node_id.replace("r-", ""), fill="red", font=font)
            caption_y = (i // cols) * panel_h + int(0.70 * panel_h)
            draw.text(((i % cols) * panel_w + 6, caption_y), sample["id"][-12:], fill="black", font=font)
            edge_text = ", ".join(f"{e['predicate']}({e['from'].replace('r-','')}>{e['to'].replace('r-','')})" for e in edges)
            if not edge_text:
                edge_text = "no predicted edges"
            for line_idx in range(0, len(edge_text), 62):
                draw.text(((i % cols) * panel_w + 6, caption_y + 14 + (line_idx // 62) * 12), edge_text[line_idx:line_idx + 62], fill="black", font=font)
        path = args.output_dir / f"relation-review-{domain}.jpg"
        sheet.save(path, quality=94)
        print(path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

