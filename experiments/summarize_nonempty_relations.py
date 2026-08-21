#!/usr/bin/env python3
from __future__ import annotations
import argparse, json
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    data = json.loads(args.input.read_text())
    rows = []
    for sample in data['samples']:
        if sample['edges']:
            rows.append({
                'sample_id': sample['sample_id'],
                'domain': sample['domain'],
                'edges': [{k: e[k] for k in ('from','to','predicate','score','from_bbox','to_bbox')} for e in sample['edges']],
            })
    args.output.write_text(json.dumps(rows, indent=2) + '\n')
    print(json.dumps({'samples': len(rows), 'edges': sum(len(x['edges']) for x in rows)}))
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
