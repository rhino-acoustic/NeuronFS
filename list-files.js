const fs = require('fs');
const path = require('path');

const targetDir = process.argv[2] || '.';

try {
  const files = fs.readdirSync(targetDir);
  console.log(`📂 [${path.resolve(targetDir)}] 파일 목록 (${files.length}개):\n`);
  files.forEach((file) => {
    const fullPath = path.join(targetDir, file);
    const stat = fs.statSync(fullPath);
    const type = stat.isDirectory() ? '📁' : '📄';
    console.log(`  ${type} ${file}`);
  });
} catch (err) {
  console.error(`❌ 오류: ${err.message}`);
  process.exit(1);
}
