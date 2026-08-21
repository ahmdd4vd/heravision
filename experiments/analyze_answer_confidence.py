#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from collections import defaultdict
from pathlib import Path


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); f=json.loads(args.fixture.read_text()); p={json.loads(l)['sample_id']:json.loads(l) for l in args.predictions.read_text().splitlines() if l.strip()}
    rows=[]
    for s in f['samples']:
        r=p[s['id']]['b1']['answer']; rows.append({'id':s['id'],'domain':s['domain'],'expected':s['expected_answer_status'],'predicted':r['status'],'confidence':r['confidence']})
    args.output.write_text(json.dumps(rows,indent=2)+'\n')
    for expected in ('answered','abstain','insufficient_evidence'):
        vals=[r['confidence'] for r in rows if r['expected']==expected]
        print(expected, {'n':len(vals),'min':min(vals) if vals else None,'max':max(vals) if vals else None,'mean':sum(vals)/len(vals) if vals else None,'values':sorted(vals)})

if __name__=='__main__': main()
