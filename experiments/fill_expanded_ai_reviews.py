#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def label(item):
    sid=item['sample_id']
    if sid.startswith('domain-extra-screen-'): return ('screenshot_document','answerable','high','screenshot/document UI or dashboard')
    if sid.startswith('domain-extra-ambiguous-'): return ('ambiguous','abstain','medium','low-information or motion-blurred candidate')
    # Existing 22 were already visually reviewed in pass 1.
    source=json.loads(Path('experiments/manifests/domain-calibration-reviewer-a.json').read_text())
    for r in source['samples']:
        if r['sample_id']==sid: return (r['review_domain'],r['review_answerability'],r['review_confidence'],r['reviewer_notes'])
    raise SystemExit(f'missing prior AI label for {sid}')

def fill(path,name,mode,note):
    data=json.loads(path.read_text()); data['reviewer_name']=name; data['review_date']='2026-08-22'; data['review_mode']=mode; data['independence_statement']='AI provisional review; not an independent human review.'
    for item in data['samples']:
        d,a,c,n=label(item); item['review_domain']=d; item['review_answerability']=a; item['review_confidence']=c; item['reviewer_notes']=f'{note}: {n}'
    path.write_text(json.dumps(data,indent=2)+'\n')

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--review-a',type=Path,required=True); ap.add_argument('--review-b',type=Path,required=True); args=ap.parse_args(); fill(args.review_a,'AI reviewer — expanded pass 1','ai-provisional','expanded visual first pass'); fill(args.review_b,'AI reviewer — expanded self-audit pass 2','ai-self-audit','expanded second visual pass, same AI process'); print('filled expanded AI review A/B')
if __name__=='__main__': main()
