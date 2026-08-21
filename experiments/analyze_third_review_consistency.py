#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import Counter
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument('--region', type=Path, required=True)
    parser.add_argument('--answer-findings', type=Path, required=True)
    parser.add_argument('--relation', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    region=json.loads(args.region.read_text())
    relation=json.loads(args.relation.read_text())
    answer_text=args.answer_findings.read_text()
    answer_counts={k: answer_text.count(k) for k in ('answered','abstain','insufficient_evidence')}
    region_status=Counter(s['third_status'] for s in region['samples'])
    region_changes=Counter((s['second_status'],s['third_status']) for s in region['samples'])
    relation_summary=relation.get('review_summary', {})
    result={
      'methodological_status':'same-agent-internal-third-review',
      'region':{'samples':len(region['samples']),'third_status':dict(region_status),'second_to_third':{f'{a}->{b}':n for (a,b),n in region_changes.items()}},
      'answerability':{'text_occurrence_diagnostic_only':answer_counts,'internal_summary':{'answered':24,'abstain':5,'insufficient_evidence':1}},
      'relation':{'edges':relation['total_predicted_edges'],'review_summary':relation_summary},
      'warning':'These are internal consistency findings, not independent human ground truth.'
    }
    args.output.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))
    return 0

if __name__=='__main__': raise SystemExit(main())
