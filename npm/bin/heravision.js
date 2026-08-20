#!/usr/bin/env node
const { spawn } = require('child_process');
const p = spawn('heravision', process.argv.slice(2), { stdio: 'inherit' });
p.on('exit', c => process.exit(c ?? 0));
