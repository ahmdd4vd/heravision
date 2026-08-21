#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from collections import Counter, defaultdict
from pathlib import Path


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); fixture=json.loads(args.fixture.read_text())
    predictions={json.loads(line)['sample_id']: json.loads(line) for line in args.predictions.read_text().splitlines() if line.strip()}
    matrix=Counter(); by_domain=defaultdict(Counter)
    for s in fixture['samples']:
        expected=s['expected_answer_status']; row=predictions[s['id']]
        predicted=row.get('b1',{}).get('answer',{}).get('status','missing') if row.get('status')=='ok' else 'error'
        matrix[(expected,predicted)]+=1; by_domain[s['domain']][(expected,predicted)]+=1
    nonanswerable=sum(n for (e,p),n in matrix.items() if e!='answered')
    unsafe=sum(n for (e,p),n in matrix.items() if e!='answered' and p=='answered')
    answerable=sum(n for (e,p),n in matrix.items() if e=='answered')
    covered=sum(n for (e,p),n in matrix.items() if e=='answered' and p=='answered')
    result={
      'status':'internal-development-only',
      'matrix':{f'{e}->{p}':n for (e,p),n in sorted(matrix.items())},
      'coverage_on_expected_answerable': covered/answerable if answerable else 0,
      'unsafe_answer_rate_on_expected_nonanswerable': unsafe/nonanswerable if nonanswerable else 0,
      'by_domain':{d:{f'{e}->{p}':n for (e,p),n in sorted(c.items())} for d,c in sorted(by_domain.items())},
      'warning':'Expected statuses are same-agent internal review labels, not independent human ground truth.'
    }
    args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result,indent=2))

if __name__=='__main__': main()
