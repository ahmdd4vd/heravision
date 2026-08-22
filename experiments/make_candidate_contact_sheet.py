#!/usr/bin/env python3
from __future__ import annotations
import argparse
from pathlib import Path
from PIL import Image,ImageDraw

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--root',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); ap.add_argument('--limit',type=int,default=20); args=ap.parse_args(); files=sorted([p for p in args.root.rglob('*') if p.is_file() and p.suffix.lower() in {'.jpg','.jpeg','.png','.webp'}])[:args.limit]; thumb=220; cols=4; rows=(len(files)+cols-1)//cols; canvas=Image.new('RGB',(cols*thumb,rows*250),'white'); d=ImageDraw.Draw(canvas)
    for i,p in enumerate(files):
        x=i%cols*thumb; y=i//cols*250
        with Image.open(p).convert('RGB') as im: im.thumbnail((212,212)); canvas.paste(im,(x+(thumb-im.width)//2,y+4))
        d.rectangle((x,y+216,x+thumb,y+250),fill='#eee'); d.text((x+5,y+220),f'{i+1:02d} {p.name[:20]}',fill='black')
    args.output.parent.mkdir(parents=True,exist_ok=True); canvas.save(args.output); print(f'wrote {args.output} ({len(files)} samples)')
if __name__=='__main__': main()
