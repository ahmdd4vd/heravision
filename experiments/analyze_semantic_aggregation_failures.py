#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args()
    fixture={s['id']:s for s in json.loads(args.fixture.read_text())['samples']}; rows=[]
    for line in args.predictions.read_text().splitlines():
        if not line.strip(): continue
        row=json.loads(line); accepted=[]
        for node in row.get('b1',{}).get('graph',{}).get('nodes',[]):
            for h in node.get('hypotheses',[]):
                if h.get('status')=='accepted' and h.get('id','').startswith('image-sem-'): accepted.append(h)
        if accepted:
            h=max(accepted,key=lambda x:x['score']); target=fixture.get(row['sample_id'],{}).get('broad_label','missing'); rows.append({'sample_id':row['sample_id'],'target':target,'label':h['label'],'score':h['score'],'region_ids':h.get('region_ids',[]),'evidence':h.get('evidence',[])})
    result={'accepted_samples':len(rows),'correct':sum(r['target']==r['label'] for r in rows),'rows':rows}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps({'accepted_samples':result['accepted_samples'],'correct':result['correct']},indent=2))
if __name__=='__main__': main()
