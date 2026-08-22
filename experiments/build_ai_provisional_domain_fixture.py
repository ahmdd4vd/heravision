#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--review',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); fixture=json.loads(args.fixture.read_text()); review=json.loads(args.review.read_text()); labels={r['sample_id']:r for r in review['samples']}; out=[]
    for s in fixture['samples']:
        r=labels.get(s['id']);
        if not r or r.get('review_domain') not in {'natural_photo','diagram_document','screenshot_document','ambiguous'}: continue
        item=dict(s); item['visual_domain']=r['review_domain']; item['answerability']=r['review_answerability']; item['label_source']='ai-two-pass-provisional'; item['independent_review_status']='not-independent'; out.append(item)
    result={'name':'domain-calibration-ai-provisional-22','description':'AI pass 1 plus same-process self-audit; use only for provisional experiments.','annotation_status':'ai-two-pass-provisional','independence':'not-independent-human','samples':out}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps({'samples':len(out),'domains':{d:sum(s['visual_domain']==d for s in out) for d in sorted(set(s['visual_domain'] for s in out))}},indent=2))
if __name__=='__main__': main()
