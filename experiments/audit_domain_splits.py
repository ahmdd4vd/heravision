#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from collections import Counter
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--dir',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); seen={}; report={}; errors=[]
    for split in ('train','calibration','holdout'):
        data=json.loads((args.dir/f'domain-calibration-{split}.json').read_text()); ids=[s['id'] for s in data['samples']];
        for sid in ids:
            if sid in seen: errors.append({'sample_id':sid,'splits':[seen[sid],split]})
            seen[sid]=split
        report[split]={'count':len(ids),'domains':dict(Counter(s['visual_domain'] for s in data['samples'])),'source_overlap':dict(Counter(s.get('source_overlap','') for s in data['samples']))}
    result={'splits':report,'total_unique_ids':len(seen),'overlap_errors':errors,'status':'ok' if not errors else 'invalid','warning':'Metadata-derived candidate labels and source overlap remain; this is a split audit, not an accuracy claim.'}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result,indent=2))
if __name__=='__main__': main()
