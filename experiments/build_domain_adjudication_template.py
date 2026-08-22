#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--reconciliation',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); src=json.loads(args.reconciliation.read_text()); rows=[]
    for r in src.get('samples',[]):
        if r.get('status')!='disagreement': continue
        rows.append({'sample_id':r['sample_id'],'split':r.get('split',''),'review_a':r.get('review_a',{}),'review_b':r.get('review_b',{}),'adjudicated_domain':'','adjudicated_answerability':'','adjudicator_name':'','adjudication_reason':'','status':'pending'})
    out={'name':'domain-calibration-adjudication','instructions':'Review only disagreement samples using the original image. Do not select a consensus merely to maximize agreement. Choose one domain and answerability, or mark unresolved and exclude from training/benchmark. Record the adjudicator and reason.','samples':rows,'status':'pending'}; args.output.write_text(json.dumps(out,indent=2)+'\n'); print(json.dumps({'disagreements':len(rows),'status':'pending'},indent=2))
if __name__=='__main__': main()
