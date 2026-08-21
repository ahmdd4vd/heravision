#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from pathlib import Path

IMAGENETTE={"n01440764":"animal","n02102040":"animal","n02979186":"artifact","n03000684":"artifact","n03028079":"artifact","n03394916":"artifact","n03417042":"vehicle","n03425413":"artifact","n03445777":"artifact","n03888257":"artifact"}

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--imagenette',type=Path,required=True); ap.add_argument('--imagewoof',type=Path,required=True); ap.add_argument('--per-class',type=int,default=40); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); samples=[]
    for root in (args.imagenette,args.imagewoof):
        for cls in sorted(p for p in root.iterdir() if p.is_dir()):
            label=IMAGENETTE.get(cls.name,'animal') if 'imagenette' in str(root) else 'animal'; files=sorted([p for p in cls.iterdir() if p.is_file()])[:args.per_class]
            for p in files: samples.append({'id':f'semtrain-{cls.name}-{p.stem}','image_path':str(p),'split':'semantic-train','tags':['pseudo_image_label',label]})
    args.output.write_text(json.dumps({'name':'semantic-train-imagenet-style','description':'Training-only pseudo image labels; no blind validation samples.','samples':samples},indent=2)+'\n'); print(json.dumps({'samples':len(samples),'labels':{'animal':sum('animal' in s['tags'] for s in samples),'artifact':sum('artifact' in s['tags'] for s in samples),'vehicle':sum('vehicle' in s['tags'] for s in samples)}},indent=2))
if __name__=='__main__': main()
