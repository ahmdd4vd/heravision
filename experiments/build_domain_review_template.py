#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); src=json.loads(args.fixture.read_text()); samples=[]
    for s in src['samples']:
        samples.append({'sample_id':s['id'],'image_path':s['image_path'],'review_domain':'','review_answerability':'','object_presence':'','confidence':'','disagreement_reason':'','notes':''})
    result={'name':src['name']+'-independent-review','instructions':'Review original images without model predictions or proposed labels. Choose one domain: natural_photo, diagram_document, screenshot_document, ambiguous. Choose answerability: answerable or abstain. Mark confidence high, medium, or low. If uncertain, explain why. A second reviewer must complete a separate copy.','reviewer_name':'','reviewer_date':'','independence_statement':'I reviewed the original images independently of model predictions.','samples':samples}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps({'samples':len(samples),'output':str(args.output)},indent=2))
if __name__=='__main__': main()
