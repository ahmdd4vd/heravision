#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from collections import Counter
from pathlib import Path


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); rows=[]; status=Counter()
    for line in args.predictions.read_text().splitlines():
        if not line.strip(): continue
        r=json.loads(line)
        for n in r.get('b1',{}).get('graph',{}).get('nodes',[]):
            hs=[h for h in n.get('hypotheses',[]) if '-sem-' in h.get('id','')]
            if hs:
                best=max(hs,key=lambda h:h['score']); rows.append({'sample_id':r['sample_id'],'label':best['label'],'score':best['score'],'status':best['status']}); status[best['status']]+=1
    scores=[r['score'] for r in rows]; result={'n':len(rows),'min':min(scores) if scores else None,'max':max(scores) if scores else None,'mean':sum(scores)/len(scores) if scores else None,'status':dict(status),'top':sorted(rows,key=lambda r:r['score'],reverse=True)[:20]}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result,indent=2))
if __name__=='__main__': main()
