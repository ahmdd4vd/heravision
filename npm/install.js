const { execSync } = require('child_process');
const os = require('os');
const plat = os.platform(), arch = os.arch();
console.log(`heravision npm wrapper: ${plat} ${arch} - download binary from GitHub releases`);
console.log(`or build from source: go install github.com/heravision/heravision@latest`);
