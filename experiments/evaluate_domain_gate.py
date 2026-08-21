#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from collections import Counter
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--fixture',type=Path,required=True); ap.add_argument('--predictions',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args()
    fixture={s['id']:s for s in json.loads(args.fixture.read_text())['samples']}; rows=[]; matrix=Counter()
    for line in args.predictions.read_text().splitlines():
        if not line.strip(): continue
        row=json.loads(line); hs=[]
        for node in row.get('b1',{}).get('graph',{}).get('nodes',[]):
            hs.extend(h for h in node.get('hypotheses',[]) if h.get('id','').startswith('image-domain-'))
        if not hs: continue
        h=hs[0]; domain=fixture.get(row['sample_id'],{}).get('domain','missing'); target='diagram_document' if 'wikimedia' in domain or 'diagram' in domain else ('natural_photo' if domain in {'imagenette','imagewoof'} else 'missing'); pred=h['label']; matrix[(target,pred)]+=1; rows.append({'sample_id':row['sample_id'],'target':target,'predicted':pred,'score':h.get('score'),'status':h.get('status')})
    answered=[r for r in rows if r['predicted']!='ambiguous']; correct=sum(r['target']==r['predicted'] for r in answered); result={'samples':len(rows),'coverage':len(answered)/len(rows) if rows else 0,'selective_accuracy':correct/len(answered) if answered else 0,'matrix':{f'{a}->{b}':n for (a,b),n in sorted(matrix.items())},'rows':rows,'warning':'MD2 domain targets are provisional metadata-derived labels.'}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps({k:result[k] for k in ('samples','coverage','selective_accuracy','matrix')},indent=2))
if __name__=='__main__': main()
