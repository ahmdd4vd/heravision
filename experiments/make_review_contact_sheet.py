#!/usr/bin/env python3
from __future__ import annotations
import argparse,json,math
from pathlib import Path
from PIL import Image,ImageDraw,ImageFont

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--manifest',type=Path,required=True); ap.add_argument('--output',type=Path,required=True); ap.add_argument('--thumb',type=int,default=240); args=ap.parse_args(); data=json.loads(args.manifest.read_text()); items=data['samples']; cols=4; rows=math.ceil(len(items)/cols); label_h=44; canvas=Image.new('RGB',(cols*args.thumb,rows*(args.thumb+label_h)),'white'); draw=ImageDraw.Draw(canvas)
    for i,item in enumerate(items):
        x=(i%cols)*args.thumb; y=(i//cols)*(args.thumb+label_h); p=Path(item['image_path']);
        try:
            with Image.open(p).convert('RGB') as im:
                im.thumbnail((args.thumb-8,args.thumb-8)); ox=x+(args.thumb-im.width)//2; oy=y+4+(args.thumb-8-im.height)//2; canvas.paste(im,(ox,oy))
        except Exception:
            draw.rectangle((x+4,y+4,x+args.thumb-4,y+args.thumb-4),outline='red',width=2); draw.text((x+10,y+args.thumb//2), 'LOAD ERROR', fill='red')
        draw.rectangle((x,y+args.thumb,x+args.thumb,y+args.thumb+label_h),fill='#eeeeee'); draw.text((x+6,y+args.thumb+5),f"{i+1:02d} {item.get('sample_id', item.get('id', 'unknown'))}",fill='black'); draw.text((x+6,y+args.thumb+22),item.get('split',''),fill='#444444')
    args.output.parent.mkdir(parents=True,exist_ok=True); canvas.save(args.output); print(f'wrote {args.output} ({len(items)} samples)')
if __name__=='__main__': main()
