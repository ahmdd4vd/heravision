#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from collections import Counter
from pathlib import Path


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); labels=Counter(); statuses=Counter(); samples=0; errors=0; semantic_count=0; missing_evidence=0; accepted_samples=set()
    for line in args.predictions.read_text().splitlines():
        if not line.strip(): continue
        row=json.loads(line); samples+=1
        if row.get('status')!='ok': errors+=1; continue
        found=False
        for node in row.get('b1',{}).get('graph',{}).get('nodes',[]):
            for h in node.get('hypotheses',[]):
                if '-sem-' not in h.get('id',''): continue
                found=True; semantic_count+=1; labels[h['label']]+=1; statuses[h.get('status','')]+=1
                if not h.get('evidence'): missing_evidence+=1
                if h.get('status')=='accepted': accepted_samples.add(row['sample_id'])
        if not found: pass
    result={'predictions':str(args.predictions),'samples':samples,'errors':errors,'semantic_hypotheses':semantic_count,'samples_with_accepted_semantic':len(accepted_samples),'labels':dict(labels),'statuses':dict(statuses),'missing_evidence':missing_evidence}
    args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result,indent=2))

if __name__=='__main__': main()
