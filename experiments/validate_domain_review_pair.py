#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--review-a',type=Path,required=True); ap.add_argument('--review-b',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); a=json.loads(args.review_a.read_text()); b=json.loads(args.review_b.read_text()); sa={x['sample_id']:x for x in a.get('samples',[])}; sb={x['sample_id']:x for x in b.get('samples',[])}; errors=[]
    if set(sa)!=set(sb): errors.append({'type':'sample_id_mismatch','only_a':sorted(set(sa)-set(sb)),'only_b':sorted(set(sb)-set(sa))})
    if a.get('reviewer_name') and a.get('reviewer_name')==b.get('reviewer_name'): errors.append({'type':'reviewer_identity_not_distinct'})
    for sid in sorted(set(sa)&set(sb)):
        if sa[sid].get('image_path')!=sb[sid].get('image_path'): errors.append({'type':'image_path_mismatch','sample_id':sid})
        for key in ('review_domain','review_answerability','review_confidence'):
            if key not in sa[sid] or key not in sb[sid]: errors.append({'type':'missing_review_field','sample_id':sid,'field':key})
    result={'samples_a':len(sa),'samples_b':len(sb),'reviewer_a':a.get('reviewer_name',''),'reviewer_b':b.get('reviewer_name',''),'errors':errors,'status':'valid-pair' if not errors and a.get('reviewer_name') and b.get('reviewer_name') and a.get('independence_statement') and b.get('independence_statement') else 'pending','warning':'Pair validity does not prove the reviews were actually performed independently.'}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result,indent=2))
if __name__=='__main__': main()
