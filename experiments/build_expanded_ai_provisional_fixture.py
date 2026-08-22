#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--review',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); fixture=json.loads(args.fixture.read_text()); review={r['sample_id']:r for r in json.loads(args.review.read_text())['samples']}; out=[]
    for s in fixture['samples']:
        r=review.get(s['id']);
        if not r or not r.get('review_domain'): continue
        item=dict(s); item.update({'visual_domain':r['review_domain'],'answerability':r['review_answerability'],'label_source':'ai-two-pass-provisional','independent_review_status':'not-independent'}); out.append(item)
    args.output.write_text(json.dumps({'name':'domain-calibration-expanded-ai-provisional','description':'AI two-pass labels only; not independent human ground truth.','annotation_status':'ai-two-pass-provisional','independence':'not-independent-human','samples':out},indent=2)+'\n'); print(json.dumps({'samples':len(out),'domains':{d:sum(s['visual_domain']==d for s in out) for d in sorted(set(s['visual_domain'] for s in out))}},indent=2))
if __name__=='__main__': main()
