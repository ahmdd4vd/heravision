#!/usr/bin/env python3
from __future__ import annotations
import argparse,hashlib,json
from pathlib import Path

def bucket(sample_id: str) -> int:
    return int.from_bytes(hashlib.sha256(sample_id.encode()).digest()[:8], 'big') % 10

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--input',type=Path,required=True); ap.add_argument('--out-dir',type=Path,required=True); args=ap.parse_args(); src=json.loads(args.input.read_text());     groups={'train':[],'calibration':[],'holdout':[]}; by_domain={}
    for s in src['samples']: by_domain.setdefault(s['visual_domain'],[]).append(s)
    for domain,items in sorted(by_domain.items()):
        items=sorted(items,key=lambda s:s['id'])
        for idx,s in enumerate(items):
            split=('train','calibration','holdout')[idx%5 if idx%5<3 else 1 if idx%5==3 else 2]
            item=dict(s); item['split']=split; item['split_assignment']='stratified_domain_round_robin'; groups[split].append(item)
    args.out_dir.mkdir(parents=True,exist_ok=True)
    for split,items in groups.items():
        payload={'name':f"{src['name']}-{split}",'description':'Candidate only; labels require independent review before calibration claims.','split':split,'independence':src.get('independence','pending'),'samples':items}; (args.out_dir/f'domain-calibration-{split}.json').write_text(json.dumps(payload,indent=2)+'\n')
    review=[]
    for s in groups['calibration']+groups['holdout']:
        review.append({'sample_id':s['id'],'image_path':s['image_path'],'split':s['split'],'review_domain':'','review_answerability':'','review_confidence':'','reviewer_notes':''})
    (args.out_dir/'domain-calibration-independent-review-calibration-holdout.json').write_text(json.dumps({'name':'domain-calibration-independent-review-calibration-holdout','instructions':'Review original images without model predictions or proposed labels. Use natural_photo, diagram_document, screenshot_document, or ambiguous. Mark answerability and confidence. Complete a separate copy for each reviewer.','reviewer_name':'','reviewer_date':'','independence_statement':'','samples':review},indent=2)+'\n')
    print(json.dumps({k:len(v) for k,v in groups.items()},indent=2))
if __name__=='__main__': main()
