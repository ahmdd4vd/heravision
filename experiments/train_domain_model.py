#!/usr/bin/env python3
from __future__ import annotations
import argparse, hashlib, json, re
from pathlib import Path
import numpy as np

FEATURES = ['flatness_mean','edge_density','contrast_mean','chroma_mean','chroma_std','orientation_entropy','luma_std','aspect_ratio','blank_fraction','low_info_fraction','edge_high_fraction','orientation_concentration','axis_concentration','luma_range','line_structure']
PAT = re.compile(r'flatness_mean=([0-9.]+) edge_density=([0-9.]+) contrast_mean=([0-9.]+) chroma_mean=([0-9.]+) chroma_std=([0-9.]+) orientation_entropy=([0-9.]+) luma_std=([0-9.]+) aspect_ratio=([0-9.]+) blank_fraction=([0-9.]+) low_info_fraction=([0-9.]+) edge_high_fraction=([0-9.]+) orientation_concentration=([0-9.]+) axis_concentration=([0-9.]+) luma_range=([0-9.]+) line_structure=([0-9.]+)')

def sigmoid(x):
    x=np.clip(x,-30,30); return 1/(1+np.exp(-x))

def fit(X,y,steps=1600,lr=.12,l2=.02):
    w=np.zeros(X.shape[1]); b=0.; pos=max(1,float(y.sum())); neg=max(1,float(len(y)-y.sum())); weights=np.where(y>.5,.5/pos,.5/neg)
    for _ in range(steps):
        p=sigmoid(X@w+b); err=(p-y)*weights; w-=lr*(X.T@err+l2*w); b-=lr*float(err.sum())
    return w,b

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); ap.add_argument('--reconciliation',type=Path,required=True); ap.add_argument('--allow-ai-provisional',action='store_true'); args=ap.parse_args()
    reconciliation=json.loads(args.reconciliation.read_text())
    allowed = reconciliation.get('status')=='consensus-ready' or (args.allow_ai_provisional and reconciliation.get('status')=='ai-consensus-provisional')
    if not allowed or reconciliation.get('counts',{}).get('disagreement',1)!=0:
        raise SystemExit('refusing to train: reviewer reconciliation is not eligible for this training mode')
    fixture={s['id']:s for s in json.loads(args.fixture.read_text())['samples']}; X=[]; y=[]; ids=[]
    for line in args.predictions.read_text().splitlines():
        if not line.strip(): continue
        row=json.loads(line); item=fixture.get(row.get('sample_id'))
        if not item: continue
        feats=None
        for node in (row.get('b1') or {}).get('graph',{}).get('nodes') or []:
            for h in node.get('hypotheses',[]):
                if h.get('id','').startswith('image-domain-'):
                    for e in h.get('evidence',[]):
                        m=PAT.search(e.get('note',''))
                        if m: feats=[float(v) for v in m.groups()]; break
                if feats: break
            if feats: break
        if feats:
            X.append(feats); y.append(item['visual_domain']); ids.append(row['sample_id'])
    labels=sorted(set(y)); A=np.asarray(X,float); split=np.asarray([int.from_bytes(hashlib.sha256(i.encode()).digest()[:8],'big')%5!=0 for i in ids]); weights={}; bias={}; metrics={}
    for label in labels:
        yy=(np.asarray(y)==label).astype(float); w,b=fit(A[split],yy[split]); weights[label]=w.tolist(); bias[label]=float(b); pred=(sigmoid(A[~split]@w+b)>=.5) if (~split).any() else np.array([]); metrics[label]={'test_n':int((~split).sum()),'test_accuracy':float((pred==yy[~split]).mean()) if (~split).any() else None}
    model={'name':'domain-calibration-reviewed-logistic','features':FEATURES,'labels':labels,'weights':weights,'bias':bias,'min_score':.65,'min_margin':.10,'training':{'fixture':str(args.fixture),'samples':len(y),'split':'deterministic sample hash','status':'ai-provisional-experiment' if args.allow_ai_provisional else 'independent-review-consensus-required'},'test_metrics':metrics}; args.output.write_text(json.dumps(model,indent=2)+'\n'); print(json.dumps({'samples':len(y),'labels':labels,'metrics':metrics},indent=2))
if __name__=='__main__': main()
