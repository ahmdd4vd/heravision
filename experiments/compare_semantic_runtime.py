#!/usr/bin/env python3
from __future__ import annotations
import argparse, json, statistics
from pathlib import Path

def read(path):
    rows=[json.loads(line) for line in Path(path).read_text().splitlines() if line.strip()]
    ms=[r['b1']['elapsed_ms'] for r in rows if r.get('status')=='ok']
    accepted=sum(any(h.get('status')=='accepted' and '-sem-' in h.get('id','') for n in r['b1']['graph'].get('nodes',[]) for h in n.get('hypotheses',[])) for r in rows if r.get('status')=='ok')
    return {'samples':len(rows),'ok':len(ms),'mean_ms':statistics.mean(ms) if ms else None,'median_ms':statistics.median(ms) if ms else None,'p95_ms':sorted(ms)[max(0,int(len(ms)*.95)-1)] if ms else None,'accepted_samples':accepted}

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--baseline',type=Path,required=True); ap.add_argument('--semantic',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args()
    result={'baseline':read(args.baseline/'predictions.jsonl'),'semantic':read(args.semantic/'predictions.jsonl')}; result['delta_mean_ms']=result['semantic']['mean_ms']-result['baseline']['mean_ms'] if result['baseline']['mean_ms'] is not None and result['semantic']['mean_ms'] is not None else None; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result,indent=2))
if __name__=='__main__': main()
