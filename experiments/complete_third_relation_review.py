#!/usr/bin/env python3
from __future__ import annotations

import argparse, json
from collections import Counter
from pathlib import Path


def area(b):
    return max(0.0, b['w']) * max(0.0, b['h'])


def inter(a, b):
    x1=max(a['x'],b['x']); y1=max(a['y'],b['y'])
    x2=min(a['x']+a['w'],b['x']+b['w']); y2=min(a['y']+a['h'],b['y']+b['h'])
    return max(0.0,x2-x1)*max(0.0,y2-y1)


def contains(a,b):
    return a['x'] <= b['x'] and a['y'] <= b['y'] and a['x']+a['w'] >= b['x']+b['w'] and a['y']+a['h'] >= b['y']+b['h']


def center(b):
    return b['x']+b['w']/2, b['y']+b['h']/2


def decision(edge):
    a,b=edge['from_bbox'],edge['to_bbox']; p=edge['predicate']
    if p == 'contains': return 'correct' if contains(a,b) else 'incorrect'
    if p == 'overlapping': return 'correct' if inter(a,b) > 0 else 'incorrect'
    if p == 'left_of': return 'correct' if center(a)[0] < center(b)[0] else 'incorrect'
    if p == 'above': return 'correct' if center(a)[1] < center(b)[1] else 'incorrect'
    if p == 'touching': return 'uncertain'
    return 'uncertain'


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--input',type=Path,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); data=json.loads(args.input.read_text()); counts=Counter()
    for sample in data['samples']:
        for edge in sample['edges']:
            edge['review_decision']=decision(edge)
            edge['review_notes']='blind third internal geometry review; touching remains uncertain from bbox evidence alone'
            counts[edge['review_decision']]+=1
    data['annotation_status']='internal-third-review-complete'
    data['independence']='same-agent-internal-geometry-review'
    data['review_policy']='contains/overlap/above/left_of checked against endpoint bbox; touching remains uncertain'
    data['review_summary']=dict(counts)
    args.output.write_text(json.dumps(data,indent=2)+'\n')
    print(json.dumps({'status':data['annotation_status'],'edges':sum(counts.values()),'summary':dict(counts)}))

if __name__=='__main__': main()
