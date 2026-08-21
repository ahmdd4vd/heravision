#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from pathlib import Path


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--input',type=Path,required=True); ap.add_argument('--threshold',type=float,required=True); ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); data=json.loads(args.input.read_text()); data['min_evidence']=args.threshold; data.setdefault('ablation',{})['threshold']=args.threshold; args.output.write_text(json.dumps(data,indent=2)+'\n'); print(args.output)
if __name__=='__main__': main()
