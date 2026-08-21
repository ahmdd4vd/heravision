#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from collections import Counter, defaultdict
from pathlib import Path


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); f=json.loads(args.fixture.read_text()); p={json.loads(l)['sample_id']:json.loads(l) for l in args.predictions.read_text().splitlines() if l.strip()}
    matrix=Counter(); rows=[]
    for s in f['samples']:
        row=p.get(s['id']); candidates=[]
        if row and row.get('status')=='ok':
            for n in row.get('b1',{}).get('graph',{}).get('nodes',[]):
                candidates.extend(h for h in n.get('hypotheses',[]) if '-sem-' in h.get('id','') and h.get('status')=='accepted')
        pred=max(candidates,key=lambda h:h['score'])['label'] if candidates else 'abstain'
        expected=s['broad_label']; matrix[(expected,pred)]+=1; rows.append({'id':s['id'],'expected':expected,'predicted':pred,'answerability':s['expected_answer_status']})
    answered=[r for r in rows if r['predicted']!='abstain']; correct=sum(r['expected']==r['predicted'] for r in answered); result={'status':'provisional-metadata-target','samples':len(rows),'coverage':len(answered)/len(rows) if rows else 0,'selective_accuracy':correct/len(answered) if answered else 0,'matrix':{f'{a}->{b}':n for (a,b),n in sorted(matrix.items())},'rows':rows,'warning':'Targets are metadata-derived and not independent semantic annotation.'}
    args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps({k:result[k] for k in ('status','samples','coverage','selective_accuracy','matrix')},indent=2))
if __name__=='__main__': main()
