#!/usr/bin/env node
// HeraVision npm installer: fetches the prebuilt binary from GitHub Releases,
// extracts it next to this package, and launches the interactive setup wizard.
'use strict';
const { execFileSync, spawnSync } = require('child_process');
const fs = require('fs');
const os = require('os');
const path = require('path');

const REPO = 'ahmdd4vd/heravision';
const BIN_DIR = path.join(__dirname, 'bin', 'native');
const EXE = process.platform === 'win32' ? 'heravision.exe' : 'heravision';

function pickAsset(name) {
  const p = process.platform;
  const a = process.arch;
  const osName = p === 'win32' ? 'windows' : p === 'darwin' ? 'darwin' : p === 'linux' ? 'linux' : null;
  if (!osName || name.indexOf('heravision_') !== 0) return false;
  const archNames = a === 'x64' ? ['x86_64', 'amd64'] : a === 'arm64' ? ['arm64'] : [];
  const ext = p === 'win32' ? '.zip' : '.tar.gz';
  return name.indexOf('_' + osName + '_') !== -1 &&
    archNames.some(x => name.indexOf('_' + x) !== -1) &&
    name.slice(-ext.length) === ext;
}

async function main() {
  if (typeof fetch !== 'function') throw new Error('node >= 18 required');
  const res = await fetch('https://api.github.com/repos/' + REPO + '/releases/latest');
  if (!res.ok) throw new Error('github api ' + res.status);
  const rel = await res.json();
  const asset = (rel.assets || []).find(a => pickAsset(a.name));
  if (!asset) throw new Error('no matching release asset for this platform');
  console.log('heravision: downloading ' + asset.name + ' ...');

  const bin = await fetch(asset.browser_download_url);
  if (!bin.ok) throw new Error('download failed: http ' + bin.status);
  const buf = Buffer.from(await bin.arrayBuffer());
  console.log('heravision: downloaded ' + Math.round(buf.length / 1024) + ' kB');

  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'heravision-'));
  const pkgPath = path.join(tmp, asset.name);
  fs.writeFileSync(pkgPath, buf);

  fs.mkdirSync(BIN_DIR, { recursive: true });
  let extractExit;
  if (/\.zip$/.test(pkgPath)) {
    // Windows ships bsdtar as tar.exe; it reads zip archives fine.
    extractExit = spawnSync('tar', ['-xf', pkgPath, '-C', BIN_DIR], { stdio: 'inherit' }).status;
  } else {
    extractExit = spawnSync('tar', ['-xzf', pkgPath, '-C', BIN_DIR], { stdio: 'inherit' }).status;
  }
  if (extractExit !== 0) throw new Error('extraction failed (exit ' + extractExit + ')');

  const out = path.join(BIN_DIR, EXE);
  if (!fs.existsSync(out)) throw new Error('binary missing after extraction');
  fs.chmodSync(out, 0o755);
  console.log('heravision installed: ' + out);
  return out;
}

main().then(exe => {
  if (process.stdin.isTTY) {
    const r = spawnSync(exe, ['setup'], { stdio: 'inherit' });
    if (r.status !== 0) console.log('finish later with: heravision setup');
  } else {
    console.log('done. run: heravision setup   (enable OCR + connect your AI agent)');
  }
}).catch(err => {
  console.warn('heravision npm installer: ' + err.message);
  console.warn('fallback: build from source — https://github.com/ahmdd4vd/heravision');
  process.exitCode = 1;
});
