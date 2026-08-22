#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--reconciliation',type=Path,required=True); ap.add_argument('--adjudication',type=Path,required=True); ap.add_argument('--output-dir',type=Path,required=True); args=ap.parse_args()
    fixture=json.loads(args.fixture.read_text()); rec=json.loads(args.reconciliation.read_text()); adj=json.loads(args.adjudication.read_text());
    if rec.get('status')!='consensus-ready': raise SystemExit('refusing to build consensus manifest: reconciliation is not consensus-ready')
    if any(r.get('status')!='resolved' for r in adj.get('samples',[])): raise SystemExit('refusing to build consensus manifest: adjudication is not fully resolved')
    labels={r['sample_id']:(r['consensus_domain'],r['consensus_answerability']) for r in rec.get('samples',[]) if r.get('status')=='consensus'}
    for r in adj.get('samples',[]): labels[r['sample_id']]=(r['adjudicated_domain'],r['adjudicated_answerability'])
    if len(labels)!=len(fixture.get('samples',[])): raise SystemExit('refusing to build consensus manifest: not every fixture sample has a resolved label')
    out=[]
    for s in fixture['samples']:
        domain,answer=labels[s['id']]; item=dict(s); item.update({'visual_domain':domain,'answerability':answer,'label_source':'independent-review-consensus','independent_review_status':'resolved'}); out.append(item)
    args.output_dir.mkdir(parents=True,exist_ok=True)
    for split in ('train','calibration','holdout'):
        items=[s for s in out if s.get('split')==split]; (args.output_dir/f'domain-consensus-{split}.json').write_text(json.dumps({'name':f'domain-consensus-{split}','split':split,'label_source':'independent-review-consensus','samples':items},indent=2)+'\n')
    print(json.dumps({'status':'built','samples':len(out),'splits':{split:sum(s.get('split')==split for s in out) for split in ('train','calibration','holdout')}},indent=2))
if __name__=='__main__': main()
