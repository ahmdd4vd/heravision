#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import math
from pathlib import Path

import numpy as np
from PIL import Image

FEATURES = ["area_ratio", "aspect_ratio", "compactness", "solidity", "boundary_strength", "scale_stability", "color0", "color1", "color2", "texture0", "crop_luma_mean", "crop_luma_std", "crop_chroma_r_mean", "crop_chroma_b_mean", "crop_chroma_mag", "crop_edge_density", "crop_dark_fraction", "crop_bright_fraction", "ring_luma_mean", "ring_luma_std", "ring_chroma_mag", "object_ring_luma_delta", "object_ring_chroma_delta", "x", "y", "w", "h"]
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

def crop_features(arr, b):
    h, w = arr.shape[:2]; grid=8; vals=[]; edge=0.0; dark=0; bright=0
    previous=[0.0]*grid; seen=[False]*grid
    for gy in range(grid):
        for gx in range(grid):
            x=min(w-1, max(0, b['x']+(gx*b['w']+b['w']//2)//grid)); y=min(h-1, max(0, b['y']+(gy*b['h']+b['h']//2)//grid))
            px=arr[y,x].astype(float)/255.0; l=0.2126*px[0]+0.7152*px[1]+0.0722*px[2]
            vr=np.log(px[0]+1e-4)-np.log(l+1e-4); vb=np.log(px[2]+1e-4)-np.log(l+1e-4); vals.append((l,vr,vb))
            if l<0.25: dark+=1
            if l>0.75: bright+=1
            if gx>0: edge+=abs(l-previous[gx-1])
            if seen[gx]: edge+=abs(l-previous[gx])
            previous[gx]=l; seen[gx]=True
    a=np.asarray(vals); mean=float(a[:,0].mean()); std=float(a[:,0].std()); cr=float(np.clip((a[:,1].mean()+4)/8,0,1)); cb=float(np.clip((a[:,2].mean()+4)/8,0,1)); mag=float(np.clip(np.sqrt(a[:,1]**2+a[:,2]**2).mean(),0,1)); return [mean,std,cr,cb,mag,float(np.clip(edge/(2*len(vals)),0,1)),dark/len(vals),bright/len(vals)]

def ring_features(arr, b, crop):
    h,w=arr.shape[:2]; pad_x=max(2,b['w']//2); pad_y=max(2,b['h']//2); outer={'x':b['x']-pad_x,'y':b['y']-pad_y,'w':b['w']+2*pad_x,'h':b['h']+2*pad_y}; vals=[]; grid=8
    for gy in range(grid):
        for gx in range(grid):
            x=outer['x']+(gx*outer['w']+outer['w']//2)//grid; y=outer['y']+(gy*outer['h']+outer['h']//2)//grid
            if b['x']<=x<b['x']+b['w'] and b['y']<=y<b['y']+b['h']: continue
            if not (0<=x<w and 0<=y<h): continue
            px=arr[y,x].astype(float)/255.0; l=0.2126*px[0]+0.7152*px[1]+0.0722*px[2]; vr=np.log(px[0]+1e-4)-np.log(l+1e-4); vb=np.log(px[2]+1e-4)-np.log(l+1e-4); vals.append((l,vr,vb))
    if not vals: return [0.0]*5
    a=np.asarray(vals); mean=float(a[:,0].mean()); std=float(a[:,0].std()); chroma=float(np.clip(np.sqrt(a[:,1]**2+a[:,2]**2).mean(),0,1)); return [mean,std,chroma,float(np.clip(((crop[0]-mean)*4+4)/8,0,1)),float(np.clip(((crop[4]-chroma)*4+4)/8,0,1))]

def node_features(node, width, height, arr):
    r=node["region"]; f=r.get("features",{}); b=r["bbox"]; color=f.get("color",[0.0]); texture=f.get("texture",[0.0]); crop=crop_features(arr,b); ring=ring_features(arr,b,crop)
    return [float(f.get("area_ratio",0)), min(1.0,float(f.get("aspect_ratio",0))/8), min(1.0,float(f.get("compactness",0))), min(1.0,float(f.get("solidity",0))), float(f.get("boundary_strength",0)), float(f.get("scale_stability",0)), min(1.0,(float(color[0])+4)/8), min(1.0,(float(color[1])+4)/8), min(1.0,(float(color[2])+4)/8), min(1.0,abs(float(texture[0]))*20), *crop, *ring, b["x"]/max(1,width), b["y"]/max(1,height), b["w"]/max(1,width), b["h"]/max(1,height)]

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
    ap=argparse.ArgumentParser(); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--coco-root',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); ap.add_argument('--iou',type=float,default=0.25); ap.add_argument('--pseudo-predictions',type=Path); ap.add_argument('--pseudo-manifest',type=Path)
    args=ap.parse_args(); rows=[]; x=[]; y=[]; image_ids=[]
    for line in args.predictions.read_text().splitlines():
        if not line.strip(): continue
        row=json.loads(line); image=Path(row['b1']['image_path']); stem=image.stem
        label_path=args.coco_root/'labels'/'train2017'/f'{stem}.txt'
        if not label_path.exists(): continue
        with Image.open(image).convert('RGB') as im:
            scale=min(1.0,256.0/max(im.size)); size=(max(1,int(im.width*scale+0.5)),max(1,int(im.height*scale+0.5))); arr=np.asarray(im.resize(size,Image.Resampling.NEAREST))
        gt=[]
        for ln in label_path.read_text().splitlines():
            c,cx,cy,w,h=map(float,ln.split()); gt.append((int(c), (cx-w/2)*row['b1']['width'], (cy-h/2)*row['b1']['height'], w*row['b1']['width'], h*row['b1']['height']))
        for node in row['b1']['graph'].get('nodes',[]):
            b=node['region']['bbox']; box=(b['x'],b['y'],b['w'],b['h']); best=max(gt,key=lambda g:iou(box,g[1:]),default=None)
            if best is None or iou(box,best[1:]) < args.iou: continue
            label=broad(COCO_NAMES[best[0]])
            x.append(node_features(node,row['b1']['width'],row['b1']['height'],arr)); y.append(label); image_ids.append(row['sample_id'])
    if args.pseudo_predictions and args.pseudo_manifest:
        manifest=json.loads(args.pseudo_manifest.read_text()); pseudo_labels={s['id']: next((t for t in s.get('tags',[]) if t in {'animal','artifact','vehicle','person'}), 'unknown') for s in manifest.get('samples',[])}
        for line in args.pseudo_predictions.read_text().splitlines():
            if not line.strip(): continue
            row=json.loads(line); label=pseudo_labels.get(row['sample_id']); image=Path(row['b1']['image_path'])
            if label not in {'animal','artifact','vehicle','person'} or not image.exists(): continue
            with Image.open(image).convert('RGB') as im:
                scale=min(1.0,256.0/max(im.size)); size=(max(1,int(im.width*scale+0.5)),max(1,int(im.height*scale+0.5))); arr=np.asarray(im.resize(size,Image.Resampling.NEAREST))
            nodes=sorted(row['b1']['graph'].get('nodes',[]), key=lambda n: float(n.get('region',{}).get('features',{}).get('scale_stability',0))*float(n.get('region',{}).get('features',{}).get('area_ratio',0)), reverse=True)[:12]
            for node in nodes:
                x.append(node_features(node,row['b1']['width'],row['b1']['height'],arr)); y.append(label); image_ids.append(row['sample_id'])
    labels=sorted(set(y)); X=np.asarray(x,dtype=float); split=np.asarray([int.from_bytes(__import__('hashlib').sha256(s.encode()).digest()[:8], 'big') % 5 != 0 for s in image_ids])
    weights={}; bias={}; metrics={}
    for label in labels:
        tr=split; te=~split; yy=(np.asarray(y)==label).astype(float); w,b=fit(X[tr],yy[tr]); weights[label]=w.tolist(); bias[label]=float(b); pred=(sigmoid(X[te]@w+b)>=0.5).astype(int) if te.any() else np.array([]); metrics[label]={"test_n":int(te.sum()),"test_positive":int(((yy[te]==1)).sum()),"test_accuracy":float((pred==(yy[te])).mean()) if te.any() else None}
    model={"name":"semantic-broad-coco-dev-iou25-v4","feature_names":FEATURES,"features":FEATURES,"labels":labels,"weights":weights,"bias":bias,"min_evidence":0.65,"min_margin":0.10,"min_support":2,"training":{"source":"COCO128 development split plus optional train-only pseudo image labels","regions":len(y),"images":len(set(image_ids)),"iou":args.iou,"split":"deterministic image hash"},"test_metrics":metrics}
    args.output.write_text(json.dumps(model,indent=2)+'\n'); print(json.dumps({"labels":labels,"regions":len(y),"images":len(set(image_ids)),"metrics":metrics},indent=2))

if __name__=='__main__': main()
