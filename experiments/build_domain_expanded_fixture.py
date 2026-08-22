#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--base',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); args=ap.parse_args(); base=json.loads(args.base.read_text()); samples=[]
    for i,s in enumerate(base['samples']):
        item={'id':s['sample_id'],'image_path':s['image_path'],'visual_domain':'natural_photo' if i<13 else 'diagram_document','answerability':'answerable','source':'ai-review-base','source_overlap':'existing-candidate','split':'calibration' if s.get('split')=='calibration' else 'holdout','annotation_status':'ai-provisional'}; samples.append(item)
    screenshots=sorted(Path('data/domain-calibration/external-screenshots').glob('*'))
    ambiguous=sorted(Path('data/domain-calibration/external-ambiguous').glob('*'))
    for p in screenshots: samples.append({'id':'domain-extra-screen-'+p.stem,'image_path':str(p),'visual_domain':'screenshot_document','answerability':'answerable','source':'public-image-search','source_overlap':'external-source-review-pending','split':'calibration' if len(samples)%2 else 'holdout','annotation_status':'ai-provisional-pending-review'})
    for p in ambiguous: samples.append({'id':'domain-extra-ambiguous-'+p.stem,'image_path':str(p),'visual_domain':'ambiguous','answerability':'abstain','source':'public-image-search-or-local-scan','source_overlap':'external-source-review-pending','split':'calibration' if len(samples)%2 else 'holdout','annotation_status':'ai-provisional-pending-review'})
    args.output.write_text(json.dumps({'name':'domain-calibration-expanded-candidate','description':'Expanded candidate fixture; includes screenshot and ambiguous assets but remains provisional until review/source audit.','annotation_status':'ai-provisional-candidate','independence':'not-independent','samples':samples},indent=2)+'\n'); print(json.dumps({'samples':len(samples),'domains':{d:sum(x['visual_domain']==d for x in samples) for d in sorted(set(x['visual_domain'] for x in samples))}},indent=2))
if __name__=='__main__': main()
