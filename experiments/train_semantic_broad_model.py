#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import math
from pathlib import Path

import numpy as np

FEATURES = ["area_ratio", "aspect_ratio", "compactness", "solidity", "boundary_strength", "scale_stability", "color0", "color1", "color2", "texture0", "x", "y", "w", "h"]
COCO_NAMES = "person bicycle car motorcycle airplane bus train truck boat traffic light fire hydrant stop sign parking meter bench bird cat dog horse sheep cow elephant bear zebra giraffe backpack umbrella handbag tie suitcase frisbee skis snowboard sports ball kite baseball bat baseball glove skateboard surfboard tennis racket bottle wine glass cup fork knife spoon bowl banana apple sandwich orange broccoli carrot hot dog pizza donut cake chair couch potted plant bed dining table toilet tv laptop mouse remote keyboard cell phone microwave oven toaster sink refrigerator book clock vase scissors teddy bear hair drier toothbrush".split()
ANIMAL = {"bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe"}
VEHICLE = {"bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat"}

def broad(name):
    if name == "person": return "person"
    if name in ANIMAL: return "animal"
    if name in VEHICLE: return "vehicle"
    return "artifact"

def iou(a, b):
    x1=max(a[0],b[0]); y1=max(a[1],b[1]); x2=min(a[0]+a[2],b[0]+b[2]); y2=min(a[1]+a[3],b[1]+b[3])
    inter=max(0,x2-x1)*max(0,y2-y1); union=a[2]*a[3]+b[2]*b[3]-inter
    return inter/union if union else 0.0

def node_features(node, width, height):
    r=node["region"]; f=r.get("features",{}); b=r["bbox"]; color=f.get("color",[0.0]); texture=f.get("texture",[0.0])
    return [float(f.get("area_ratio",0)), min(1.0,float(f.get("aspect_ratio",0))/8), min(1.0,float(f.get("compactness",0))), min(1.0,float(f.get("solidity",0))), float(f.get("boundary_strength",0)), float(f.get("scale_stability",0)), min(1.0,(float(color[0])+4)/8), min(1.0,(float(color[1])+4)/8), min(1.0,(float(color[2])+4)/8), min(1.0,abs(float(texture[0]))*20), b["x"]/max(1,width), b["y"]/max(1,height), b["w"]/max(1,width), b["h"]/max(1,height)]

def sigmoid(x):
    x=np.clip(x,-30,30); return 1/(1+np.exp(-x))

def fit(X,y,steps=1800,lr=0.12,l2=0.02):
    w=np.zeros(X.shape[1]); b=0.0
    positive=max(1.0,float(y.sum())); negative=max(1.0,float(len(y)-y.sum()))
    weights=np.where(y>0.5, 0.5/positive, 0.5/negative)
    for _ in range(steps):
        p=sigmoid(X@w+b); err=(p-y)*weights
        w-=lr*((X.T@err)+l2*w); b-=lr*float(err.sum())
    return w,b

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--coco-root',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); ap.add_argument('--iou',type=float,default=0.25)
    args=ap.parse_args(); rows=[]; x=[]; y=[]; image_ids=[]
    for line in args.predictions.read_text().splitlines():
        if not line.strip(): continue
        row=json.loads(line); image=Path(row['b1']['image_path']); stem=image.stem
        label_path=args.coco_root/'labels'/'train2017'/f'{stem}.txt'
        if not label_path.exists(): continue
        gt=[]
        for ln in label_path.read_text().splitlines():
            c,cx,cy,w,h=map(float,ln.split()); gt.append((int(c), (cx-w/2)*row['b1']['width'], (cy-h/2)*row['b1']['height'], w*row['b1']['width'], h*row['b1']['height']))
        for node in row['b1']['graph'].get('nodes',[]):
            b=node['region']['bbox']; box=(b['x'],b['y'],b['w'],b['h']); best=max(gt,key=lambda g:iou(box,g[1:]),default=None)
            if best is None or iou(box,best[1:]) < args.iou: continue
            label=broad(COCO_NAMES[best[0]])
            x.append(node_features(node,row['b1']['width'],row['b1']['height'])); y.append(label); image_ids.append(row['sample_id'])
    labels=sorted(set(y)); X=np.asarray(x,dtype=float); split=np.asarray([int.from_bytes(__import__('hashlib').sha256(s.encode()).digest()[:8], 'big') % 5 != 0 for s in image_ids])
    weights={}; bias={}; metrics={}
    for label in labels:
        tr=split; te=~split; yy=(np.asarray(y)==label).astype(float); w,b=fit(X[tr],yy[tr]); weights[label]=w.tolist(); bias[label]=float(b); pred=(sigmoid(X[te]@w+b)>=0.5).astype(int) if te.any() else np.array([]); metrics[label]={"test_n":int(te.sum()),"test_positive":int(((yy[te]==1)).sum()),"test_accuracy":float((pred==(yy[te])).mean()) if te.any() else None}
    model={"name":"semantic-broad-coco-dev-iou25","feature_names":FEATURES,"features":FEATURES,"labels":labels,"weights":weights,"bias":bias,"min_evidence":0.65,"min_margin":0.10,"training":{"source":"COCO128 development split","regions":len(y),"images":len(set(image_ids)),"iou":args.iou,"split":"deterministic image hash"},"test_metrics":metrics}
    args.output.write_text(json.dumps(model,indent=2)+'\n'); print(json.dumps({"labels":labels,"regions":len(y),"images":len(set(image_ids)),"metrics":metrics},indent=2))

if __name__=='__main__': main()
