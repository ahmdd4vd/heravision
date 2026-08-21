#!/usr/bin/env python3
"""Score B0/B1 normalized regions against verified COCO128 boxes."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


def iou(a: dict, b: dict) -> float:
    x1 = max(a["x"], b["x"])
    y1 = max(a["y"], b["y"])
    x2 = min(a["x"] + a["w"], b["x"] + b["w"])
    y2 = min(a["y"] + a["h"], b["y"] + b["h"])
    if x2 <= x1 or y2 <= y1:
        return 0.0
    inter = (x2 - x1) * (y2 - y1)
    union = a["w"] * a["h"] + b["w"] * b["h"] - inter
    return inter / union if union > 0 else 0.0


def scale_box(box: dict, width: int, height: int, target_w: int, target_h: int) -> dict:
    sx = target_w / width
    sy = target_h / height
    return {"x": box["x"] * sx, "y": box["y"] * sy, "w": box["w"] * sx, "h": box["h"] * sy}


def pred_boxes(observation: dict, target_w: int, target_h: int) -> list[dict]:
    boxes = []
    for node in observation["graph"]["nodes"]:
        bbox = node["region"]["bbox"]
        boxes.append(scale_box(bbox, observation["width"], observation["height"], target_w, target_h))
    return boxes


def score(gt: list[dict], pred: list[dict], threshold: float) -> dict:
    candidates = sorted(((iou(g, p), gi, pi) for gi, g in enumerate(gt) for pi, p in enumerate(pred)), reverse=True)
    used_g, used_p, matches = set(), set(), []
    for value, gi, pi in candidates:
        if value < threshold or gi in used_g or pi in used_p:
            continue
        used_g.add(gi)
        used_p.add(pi)
        matches.append(value)
    tp = len(matches)
    fp = len(pred) - tp
    fn = len(gt) - tp
    precision = tp / len(pred) if pred else 0.0
    recall = tp / len(gt) if gt else 0.0
    f1 = 2 * precision * recall / (precision + recall) if precision + recall else 0.0
    return {"tp": tp, "fp": fp, "fn": fn, "precision": precision, "recall": recall, "f1": f1, "mean_iou": sum(matches) / tp if tp else 0.0, "pred_count": len(pred), "gt_count": len(gt)}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--predictions", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--iou", type=float, default=0.5)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    by_id = {sample["id"]: sample for sample in manifest["samples"]}
    rows = [json.loads(line) for line in args.predictions.read_text().splitlines() if line.strip()]
    per_engine = {}
    per_sample = []
    for row in rows:
        sample = by_id[row["sample_id"]]
        gt = [{k: float(box[k]) for k in ("x", "y", "w", "h")} for box in sample.get("ground_truth", [])]
        item = {"sample_id": row["sample_id"]}
        for engine in ("b0", "b1"):
            pred = pred_boxes(row[engine], sample["width"], sample["height"])
            item[engine] = score(gt, pred, args.iou)
            per_engine.setdefault(engine, []).append(item[engine])
        per_sample.append(item)
    summary = {"manifest": str(args.manifest), "predictions": str(args.predictions), "iou_threshold": args.iou, "samples": len(per_sample), "engines": {}}
    for engine, metrics in per_engine.items():
        totals = {key: sum(metric[key] for metric in metrics) for key in ("tp", "fp", "fn", "pred_count", "gt_count")}
        totals["precision"] = totals["tp"] / (totals["tp"] + totals["fp"]) if totals["tp"] + totals["fp"] else 0.0
        totals["recall"] = totals["tp"] / (totals["tp"] + totals["fn"]) if totals["tp"] + totals["fn"] else 0.0
        p, r = totals["precision"], totals["recall"]
        totals["f1"] = 2 * p * r / (p + r) if p + r else 0.0
        totals["mean_iou_matched"] = sum(metric["mean_iou"] * metric["tp"] for metric in metrics) / totals["tp"] if totals["tp"] else 0.0
        totals["mean_pred_count"] = totals["pred_count"] / len(metrics) if metrics else 0.0
        summary["engines"][engine] = totals
    summary["per_sample"] = per_sample
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(summary, indent=2) + "\n")
    for engine, metrics in summary["engines"].items():
        print(engine, json.dumps(metrics, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
