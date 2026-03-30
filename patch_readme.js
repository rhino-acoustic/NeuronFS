const fs = require('fs');

function patchKo() {
    let text = fs.readFileSync('README.ko.md', 'utf8');
    
    // 1. 벤치마크 수정
    text = text.replace(
        '| 전체 스캔 (326 뉴런) | ~1ms |', 
        '| 전체 스캔 (326 뉴런) | ~1ms (vs. 일반 RAG 평균 300ms, Mem0 150ms) |'
    );
    
    // 2. 94.9% 설명 추가
    text = text.replace(
        '하네스가 위반 감지 → 교정 루프. **실측 94.9%** (brainstem 위반 18회 / 전체 353회 fire)',
        '하네스가 위반 감지 → 교정 루프. **실측 94.9%** (brainstem 위반 18회 / 전체 353회 fire)* <br><sub>* 353회 fire: 에이전트가 단독 판단 후 실제 행동을 실행(Fire)한 개별 독립 세션. 즉, 353회의 자율 행동 중 시스템 최상위 계층 무시가 단 18회에 불과함.</sub>'
    );
    
    // 3. 팔란티어 부분 뽑아서 올리기
    const palantirRegex = /## 팔란티어 인사이트[\s\S]*?\*\*현재 운영 환경:\*\* Windows 11, Google Antigravity \(DeepMind\), 326 뉴런, 2026-01부터 매일 실전 운영\.\r?\n/;
    const match = text.match(palantirRegex);
    if (match) {
        let palantirContent = match[0];
        text = text.replace(match[0], ''); // 뜯어내기
        // 제목 교체
        palantirContent = palantirContent.replace('## 팔란티어 인사이트', '## [TL;DR] 어떻게 0달러로 팔란티어급 AI 통제력을 얻는가');
        // 삽입 (### *...* 문자열 바로 밑)
        const targ1 = '### *폴더로 만든 자가 진화하는 AI 뇌. 인프라 제로. 종속성 제로.*\n\n';
        const targ2 = '### *폴더로 만든 자가 진화하는 AI 뇌. 인프라 제로. 종속성 제로.*\r\n\r\n';
        if (text.includes(targ1)) {
            text = text.replace(targ1, targ1 + palantirContent + '\n---\n\n');
        } else if (text.includes(targ2)) {
            text = text.replace(targ2, targ2 + palantirContent + '\r\n---\r\n\r\n');
        }
    }
    
    fs.writeFileSync('README.ko.md', text, 'utf8');
    console.log("README.ko.md OK");
}

function patchEn() {
    let text = fs.readFileSync('README.md', 'utf8');
    
    text = text.replace(
        '| Full scan (326 neurons) | ~1ms |', 
        '| Full scan (326 neurons) | ~1ms (vs. RAG avg 300ms, Mem0 150ms) |'
    );
    
    text = text.replace(
        'Harness detects violations → correction loop. **Measured 94.9%** (18 brainstem violations / 353 total fires)',
        'Harness detects violations → correction loop. **Measured 94.9%** (18 brainstem violations / 353 total fires)* <br><sub>* 353 total fires: The number of times the agent independently made decisions and executed autonomous actions. Only 18 core principle violations out of 353 total autonomous loops.</sub>'
    );

    const palantirRegex = /## The Palantir Insight[\s\S]*?\*\*Current production environment:\*\* Windows 11, Google Antigravity \(DeepMind\), 326 neurons, daily operation since 2026-01\.\r?\n/;
    const match = text.match(palantirRegex);
    if (match) {
        let palantirContent = match[0];
        text = text.replace(match[0], '');
        palantirContent = palantirContent.replace('## The Palantir Insight', '## [TL;DR] How to get Palantir-level AI control for $0');
        
        const targ1 = '### *A self-evolving AI brain made of folders. Zero infra. Zero dependencies.*\n\n';
        const targ2 = '### *A self-evolving AI brain made of folders. Zero infra. Zero dependencies.*\r\n\r\n';
        
        // Sometimes the EN file has slightly different sub-header, let me check just the 'A self-evolving'
        if (text.includes(targ1)) {
            text = text.replace(targ1, targ1 + palantirContent + '\n---\n\n');
        } else if (text.includes(targ2)) {
            text = text.replace(targ2, targ2 + palantirContent + '\r\n---\r\n\r\n');
        } else {
            console.log("fallback insert target");
            // just put it after "# 🧠 NeuronFS\r\n### ..."
            text = text.replace(/(# 🧠 NeuronFS\r?\n### [^\r\n]+\r?\n\r?\n)/, '$1' + palantirContent + '\r\n---\r\n\r\n');
        }
    }
    
    fs.writeFileSync('README.md', text, 'utf8');
    console.log("README.md OK");
}

patchKo();
patchEn();
