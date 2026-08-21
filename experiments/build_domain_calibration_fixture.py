#!/usr/bin/env python3
from __future__ import annotations
import argparse,json
from pathlib import Path

def sample_files(root, n): return sorted([p for p in Path(root).rglob('*') if p.is_file() and p.suffix.lower() in {'.jpg','.jpeg','.png','.webp'}])[:n]
def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--output',type=Path,required=True); ap.add_argument('--md2',type=Path,required=True); args=ap.parse_args(); samples=[]
    def add(path, domain, answer, source, overlap, note): samples.append({'id':'domain-cal-'+str(len(samples)+1).zfill(3),'image_path':str(path),'visual_domain':domain,'answerability':answer,'source':source,'source_overlap':overlap,'independent_review_status':'pending','notes':note})
    for p in sample_files('data/imagenette/imagenette2-160/train',6): add(p,'natural_photo','answerable','imagenette-train','semantic-training-pseudo','natural photo candidate')
    for p in sample_files('data/imagewoof/imagewoof2-160/train',6): add(p,'natural_photo','answerable','imagewoof-train','semantic-training-pseudo','animal photo candidate')
    for p in sample_files('data/wikimedia14',6): add(p,'diagram_document','answerable','wikimedia14','blind-md0-overlap','diagram/document candidate')
    for p in sample_files('data/domain-calibration/external-diagrams',6): add(p,'diagram_document','answerable','public-image-search','external-asset-review-pending','external diagram or flowchart; license/source review pending')
    for p in sample_files('data/imagenette/imagenette2-160/val',3): add(p,'ambiguous','abstain','imagenette-val','md0-family-overlap','ambiguous candidate requiring independent review')
    for p in sample_files('data/imagewoof/imagewoof2-160/val',3): add(p,'ambiguous','abstain','imagewoof-val','md0-family-overlap','ambiguous candidate requiring independent review')
    result={'name':'domain-calibration-candidate-30','description':'Candidate domain calibration fixture; not independent until second review and source/license audit are complete.','count':len(samples),'annotation_status':'candidate-metadata-derived','independence':'pending-independent-review','samples':samples}; args.output.write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps({'count':len(samples),'domains':{d:sum(s['visual_domain']==d for s in samples) for d in sorted(set(s['visual_domain'] for s in samples))}},indent=2))
if __name__=='__main__': main()
