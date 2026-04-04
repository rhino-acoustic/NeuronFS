const fs = require('fs');
const path = require('path');

const dir = process.cwd();
const files = fs.readdirSync(dir).filter(f => f.endsWith('.go'));

if (files.length === 0) {
  process.stdout.write('No .go files found.\n');
  process.exit(0);
}

for (const file of files) {
  const content = fs.readFileSync(path.join(dir, file), 'utf-8');
  const lines = content.split('\n').length;
  process.stdout.write(`${file}: ${lines} lines\n`);
}
