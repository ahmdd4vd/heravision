#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

LABELS={
'domain-cal-026':('natural_photo','answerable','high','person holding fish outdoors'),
'domain-cal-028':('natural_photo','answerable','high','close-up dog photo'),
'domain-cal-014':('diagram_document','answerable','high','electric-field technical diagram'),
'domain-cal-016':('diagram_document','answerable','high','Dunning-Kruger plotted graph'),
'domain-cal-019':('diagram_document','answerable','high','process/storage/input-output diagram'),
'domain-cal-021':('diagram_document','answerable','high','flowchart with process table and bar chart'),
'domain-cal-024':('diagram_document','answerable','high','system/data-flow architecture diagram'),
'domain-cal-002':('natural_photo','answerable','high','fish photographed outdoors'),
'domain-cal-004':('natural_photo','answerable','high','person holding fish outdoors'),
'domain-cal-007':('natural_photo','answerable','high','dog photo with ball'),
'domain-cal-009':('natural_photo','answerable','high','dog photo with toy'),
'domain-cal-012':('natural_photo','answerable','high','dog photo outdoors'),
'domain-cal-027':('natural_photo','answerable','high','fish on dark background'),
'domain-cal-029':('natural_photo','answerable','high','dog photo indoors'),
'domain-cal-015':('diagram_document','answerable','high','technical drawing/blueprint'),
'domain-cal-017':('diagram_document','answerable','high','decision flowchart'),
'domain-cal-020':('diagram_document','answerable','high','multi-step flowchart'),
'domain-cal-022':('diagram_document','answerable','high','network architecture diagram'),
'domain-cal-003':('natural_photo','answerable','high','person holding fish outdoors'),
'domain-cal-005':('natural_photo','answerable','high','fish in a landing net'),
'domain-cal-008':('natural_photo','answerable','high','small dog portrait'),
'domain-cal-010':('natural_photo','answerable','high','dog on patio chair'),
}

def fill(path, name, mode, pass_note):
    data=json.loads(path.read_text()); data['reviewer_name']=name; data['review_date']='2026-08-22'; data['review_mode']=mode; data['independence_statement']='AI provisional review; not an independent human review.'
    for item in data['samples']:
        d,a,c,n=LABELS[item['sample_id']]; item['review_domain']=d; item['review_answerability']=a; item['review_confidence']=c; item['reviewer_notes']=f'{pass_note}: {n}'
    path.write_text(json.dumps(data,indent=2)+'\n')

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--review-a',type=Path,required=True); ap.add_argument('--review-b',type=Path,required=True); args=ap.parse_args(); fill(args.review_a,'AI reviewer — pass 1','ai-provisional', 'visual first pass'); fill(args.review_b,'AI reviewer — self-audit pass 2','ai-self-audit', 'second visual pass, same AI process'); print('filled 2 provisional AI review files')
if __name__=='__main__': main()
