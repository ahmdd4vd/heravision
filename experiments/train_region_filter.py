#!/usr/bin/env python3
"""Train and evaluate a tiny CPU-friendly logistic region filter.

The target is not semantic recognition. A B1 region is positive when its box
has IoU >= threshold with at least one verified ground-truth box. The split is
performed by image ID so regions from one image cannot leak across train/test.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
from pathlib import Path

import numpy as np

FEATURE_NAMES = [
    "log_area", "log_width", "log_height", "aspect", "area_ratio",
    "compactness", "boundary_strength", "scale_stability", "x_norm",
    "y_norm", "w_norm", "h_norm", "texture",
]


def iou(a: dict, b: dict) -> float:
    x1, y1 = max(a["x"], b["x"]), max(a["y"], b["y"])
    x2, y2 = min(a["x"] + a["w"], b["x"] + b["w"]), min(a["y"] + a["h"], b["y"] + b["h"])
    if x2 <= x1 or y2 <= y1:
        return 0.0
    inter = (x2 - x1) * (y2 - y1)
    union = a["w"] * a["h"] + b["w"] * b["h"] - inter
    return inter / union if union > 0 else 0.0


def node_features(node: dict, width: int, height: int) -> list[float]:
    region = node["region"]
    box = region["bbox"]
    features = region.get("features", {})
    texture = features.get("texture", [0.0])
    return [
        math.log1p(max(0, region.get("area", 0))),
        math.log1p(max(0, box["w"])),
        math.log1p(max(0, box["h"])),
        float(features.get("aspect_ratio", box["w"] / max(1, box["h"]))),
        float(features.get("area_ratio", 0.0)),
        math.log1p(max(0, float(features.get("compactness", 0.0)))),
        float(features.get("boundary_strength", 0.0)),
        float(features.get("scale_stability", 0.0)),
        box["x"] / max(1, width),
        box["y"] / max(1, height),
        box["w"] / max(1, width),
        box["h"] / max(1, height),
        float(texture[0] if texture else 0.0),
    ]


def sigmoid(z: np.ndarray) -> np.ndarray:
    return 1.0 / (1.0 + np.exp(-np.clip(z, -30, 30)))


def fit_logistic(x: np.ndarray, y: np.ndarray, steps: int = 2500, lr: float = 0.08) -> tuple[np.ndarray, float, np.ndarray, np.ndarray]:
    mean = x.mean(axis=0)
    std = x.std(axis=0)
    std[std < 1e-8] = 1.0
    z = (x - mean) / std
    pos = max(1, int(y.sum()))
    neg = max(1, len(y) - pos)
    weights = np.where(y > 0.5, neg / pos, 1.0)
    w = np.zeros(z.shape[1], dtype=np.float64)
    b = 0.0
    for _ in range(steps):
        pred = sigmoid(z @ w + b)
        error = (pred - y) * weights
        w -= lr * (z.T @ error) / len(y)
        b -= lr * float(error.mean())
    return w, b, mean, std


def predict(x: np.ndarray, params: tuple[np.ndarray, float, np.ndarray, np.ndarray]) -> np.ndarray:
    w, b, mean, std = params
    return sigmoid(((x - mean) / std) @ w + b)


def f1_at(y: np.ndarray, scores: np.ndarray, threshold: float) -> dict:
    pred = scores >= threshold
    tp = int(np.logical_and(pred, y == 1).sum())
    fp = int(np.logical_and(pred, y == 0).sum())
    fn = int(np.logical_and(~pred, y == 1).sum())
    precision = tp / (tp + fp) if tp + fp else 0.0
    recall = tp / (tp + fn) if tp + fn else 0.0
    f1 = 2 * precision * recall / (precision + recall) if precision + recall else 0.0
    return {"threshold": threshold, "tp": tp, "fp": fp, "fn": fn, "precision": precision, "recall": recall, "f1": f1, "kept": int(pred.sum())}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--predictions", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--iou", type=float, default=0.5)
    args = parser.parse_args()
    manifest = json.loads(args.manifest.read_text())
    samples = {sample["id"]: sample for sample in manifest["samples"]}
    rows = [json.loads(line) for line in args.predictions.read_text().splitlines() if line.strip()]

    feature_rows, labels, image_ids = [], [], []
    for row in rows:
        sample = samples[row["sample_id"]]
        gt = sample.get("ground_truth", [])
        obs = row["b1"]
        for node in obs["graph"]["nodes"]:
            box = node["region"]["bbox"]
            candidate = {"x": box["x"] * sample["width"] / obs["width"], "y": box["y"] * sample["height"] / obs["height"], "w": box["w"] * sample["width"] / obs["width"], "h": box["h"] * sample["height"] / obs["height"]}
            best = max((iou(candidate, truth) for truth in gt), default=0.0)
            feature_rows.append(node_features(node, obs["width"], obs["height"]))
            labels.append(1 if best >= args.iou else 0)
            image_ids.append(row["sample_id"])
    x = np.asarray(feature_rows, dtype=np.float64)
    y = np.asarray(labels, dtype=np.float64)
    image_hash = {image_id: int(hashlib.sha256(image_id.encode()).hexdigest()[:8], 16) % 5 for image_id in set(image_ids)}
    train_mask = np.asarray([image_hash[image_id] != 0 for image_id in image_ids])
    test_mask = ~train_mask
    params = fit_logistic(x[train_mask], y[train_mask])
    train_scores = predict(x[train_mask], params)
    test_scores = predict(x[test_mask], params)
    thresholds = np.linspace(0.05, 0.95, 19)
    train_choices = [f1_at(y[train_mask], train_scores, float(t)) for t in thresholds]
    best = max(train_choices, key=lambda row: (row["f1"], row["precision"]))
    test_metrics = f1_at(y[test_mask], test_scores, best["threshold"])
    baseline = f1_at(y[test_mask], np.ones(test_mask.sum()), 0.5)
    output = {
        "task": "b1_region_filter",
        "positive_definition": f"best IoU >= {args.iou}",
        "features": FEATURE_NAMES,
        "samples": {"regions": int(len(y)), "train_regions": int(train_mask.sum()), "test_regions": int(test_mask.sum()), "train_images": int(len({image_ids[i] for i in range(len(image_ids)) if train_mask[i]})), "test_images": int(len({image_ids[i] for i in range(len(image_ids)) if test_mask[i]}))},
        "train_selection": best,
        "test_metrics": test_metrics,
        "raw_b1_test_baseline": baseline,
        "weights": params[0].tolist(),
        "bias": float(params[1]),
        "mean": params[2].tolist(),
        "std": params[3].tolist(),
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    print(json.dumps({"train_selection": best, "test_metrics": test_metrics, "raw_b1_test_baseline": baseline}, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
