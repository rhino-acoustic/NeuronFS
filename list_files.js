const fs = require('fs');
const path = require('path');

const targetDir = process.argv[2] || '.';

fs.readdir(targetDir, { withFileTypes: true }, (err, entries) => {
  if (err) {
    process.stderr.write(`디렉토리 읽기 실패: ${err.message}\n`);
    process.exit(1);
  }

  entries.forEach((entry) => {
    const type = entry.isDirectory() ? '[DIR] ' : '[FILE]';
    process.stdout.write(`${type} ${entry.name}\n`);
  });
});
