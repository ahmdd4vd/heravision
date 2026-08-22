#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

DOMAINS={'natural_photo','diagram_document','screenshot_document','ambiguous'}
ANS={'answerable','abstain'}
CONF={'high','medium','low'}

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--review',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); data=json.loads(args.review.read_text()); errors=[]; complete=[]
    for item in data.get('samples',[]):
        sid=item.get('sample_id','')
        if item.get('review_domain') not in DOMAINS: errors.append({'sample_id':sid,'field':'review_domain','value':item.get('review_domain')})
        if item.get('review_answerability') not in ANS: errors.append({'sample_id':sid,'field':'review_answerability','value':item.get('review_answerability')})
        if item.get('review_confidence') not in CONF: errors.append({'sample_id':sid,'field':'review_confidence','value':item.get('review_confidence')})
        if item.get('review_domain') in DOMAINS and item.get('review_answerability') in ANS and item.get('review_confidence') in CONF: complete.append(sid)
    result={'reviewer_name':data.get('reviewer_name',''),'reviewer_date':data.get('reviewer_date',''),'samples':len(data.get('samples',[])),'complete':len(complete),'errors':errors,'status':'valid' if not errors and data.get('reviewer_name') and data.get('independence_statement') else 'pending','warning':'A valid schema does not prove reviewer independence; identity and process must still be checked.'}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result,indent=2))
if __name__=='__main__': main()
