import fs from 'fs';

const file = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\hijack-launcher.mjs';
let code = fs.readFileSync(file, 'utf8');

// PS1 파일 방식으로 교체
const oldPattern = /const psCmd[\s\S]*?const output = execSync\(psCmd[^;]*;/;
const newCode = `const tmpPs = require('path').join(DUMP_DIR, '_proc.ps1');
        fs.writeFileSync(tmpPs, "Get-CimInstance Win32_Process | Where-Object { \\$_.Name -like '*Antigravity*' -or \\$_.Name -like '*language_server*' } | ForEach-Object { Write-Output \\"\\$( \\$_.ProcessId)|\\$( \\$_.ParentProcessId)|\\$( \\$_.CommandLine)\\" }");
        const output = execSync('powershell -NoProfile -ExecutionPolicy Bypass -File "' + tmpPs + '"', { encoding: 'utf8', timeout: 15000 });`;

code = code.replace(oldPattern, newCode);
fs.writeFileSync(file, code);
console.log('Launcher patched OK');
