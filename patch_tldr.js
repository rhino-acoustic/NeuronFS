const fs = require('fs');

const koNewTLDR = `## [TL;DR] 파일시스템이 곧 뇌다

Mem0, Letta, MemGPT는 모두 별도의 서버가 필요합니다. 특정 모델에 종속되고 누군가 관리해야 합니다.
NeuronFS는 단순한 OS 파일시스템 위에 삽니다. 모든 규칙은 폴더입니다. 모든 교정은 새로운 뉴런을 자동으로 자라게 만듭니다. 모델을 교체하시겠습니까? \`cp -r brain/\` 1초면 끝납니다. 에이전트 간 공유는? NAS 공유 폴더. 버전 관리는? \`git diff\`. 도입 비용? $0.

팔란티어(Palantir)의 도박은 '더 똑똑한 AI'가 아니었습니다. '더 엄격한 파이프라인'이었습니다. NeuronFS는 똑같은 철학을 파일시스템 위에서, 0달러로, 누구나 쓸 수 있게 구현한 결과물입니다.

단순한 룰 매니저가 아닙니다. 어떤 AI 모델이 오든 그보다 오래 살아남는 자가 진화형 컨텍스트 레이어입니다.
**326 뉴런. 2026년 1월부터 매일 구동 중(Daily driver). 1인 기업의 모든 AI 업무를 통제하고 있습니다.**`;

const enNewTLDR = `## [TL;DR] The Filesystem is the Brain

Mem0, Letta, MemGPT all need a server. They're tied to specific models. Someone has to manage them.
NeuronFS lives on the filesystem. Every rule is a folder. Every correction grows a new neuron automatically. Switch models? \`cp -r brain/\`. Share across agents? NAS shared folder. Version control? \`git diff\`. Cost? $0.

Palantir's bet wasn't a smarter AI — it was a stricter pipeline. This is the same bet, on a filesystem, for anyone.

Not a rule manager. A self-evolving context layer that outlives every model it runs on.
**326 neurons. Daily driver since January 2026. One person, one company, every AI task.**`;

function patchKo() {
    let text = fs.readFileSync('README.ko.md', 'utf8');
    const tldrRegex = /## \[TL;DR\] 어떻게 0달러로 팔란티어급 AI 통제력을 얻는가[\s\S]*?(?=\r?\n---)/;
    if (tldrRegex.test(text)) {
        text = text.replace(tldrRegex, koNewTLDR);
        fs.writeFileSync('README.ko.md', text, 'utf8');
        console.log("README.ko.md TL;DR updated.");
    } else {
        console.log("Regex not matched in ko.md");
    }
}

function patchEn() {
    let text = fs.readFileSync('README.md', 'utf8');
    const tldrRegex = /## \[TL;DR\] How to get Palantir-level AI control for \$0[\s\S]*?(?=\r?\n---)/;
    if (tldrRegex.test(text)) {
        text = text.replace(tldrRegex, enNewTLDR);
        fs.writeFileSync('README.md', text, 'utf8');
        console.log("README.md TL;DR updated.");
    } else {
        console.log("Regex not matched in md");
    }
}

patchKo();
patchEn();
