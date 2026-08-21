#!/usr/bin/env node
const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

const exe = process.platform === 'win32' ? 'heravision.exe' : 'heravision';
const local = path.join(__dirname, 'native', exe);
const cmd = fs.existsSync(local) ? local : 'heravision';
const p = spawn(cmd, process.argv.slice(2), { stdio: 'inherit' });
p.on('exit', c => process.exit(c ?? 0));
p.on('error', () => {
  console.error('heravision binary not found. reinstall: npm install -g heravision');
  process.exit(1);
});
