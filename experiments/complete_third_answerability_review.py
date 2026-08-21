#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from pathlib import Path

DECISIONS = {
    'imagenette': ['answered','abstain','answered','abstain','abstain','answered','abstain','answered','answered','answered'],
    'imagewoof': ['answered','answered','answered','answered','abstain','answered','answered','answered','answered','answered'],
    'wikimedia-diagram': ['answered','answered','answered','insufficient_evidence','answered','answered','answered','answered','answered','answered'],
}


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--input',type=Path,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); data=json.loads(args.input.read_text()); cursors={d:0 for d in DECISIONS}
    for sample in data['samples']:
        domain=sample['domain']; i=cursors[domain]; sample['expected_answer_status']=DECISIONS[domain][i]
        sample['answerability']='internal-third-review'; sample['reviewer_notes']='blind same-agent third pass from original image contact sheet'
        cursors[domain]+=1
    assert all(cursors[d]==10 for d in DECISIONS)
    data['annotation_status']='internal-third-review-complete'
    data['independence']='same-agent-internal-review; not independent human annotation'
    data['review_policy']='generic visual structure only; no semantic object claim'
    data['review_summary']={k:sum(v.count(k) for v in DECISIONS.values()) for k in ('answered','abstain','insufficient_evidence')}
    args.output.write_text(json.dumps(data,indent=2)+'\n')
    print(json.dumps({'status':data['annotation_status'],'summary':data['review_summary']}))

if __name__=='__main__': main()
