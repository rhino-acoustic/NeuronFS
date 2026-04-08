package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var brainDir = "C:\\Users\\BASEMENT_ADMIN\\.gemini\\antigravity\\brain\\850c6a4e-f69a-49ed-81cf-b2257b0c29fe"
var outDir = "C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\github_wiki_ready"

func compileAct(outFile, engTitle, korTitle string, startIdx, endIdx int) {
	content := fmt.Sprintf("# [ENG] %s\n# [KOR] %s\n\n", engTitle, korTitle)
	for i := startIdx; i <= endIdx; i++ {
		num := fmt.Sprintf("%02d", i)
		file := filepath.Join(brainDir, fmt.Sprintf("blog_%s.md", num))
		if _, err := os.Stat(file); err == nil {
			text, _ := ioutil.ReadFile(file)
			content += fmt.Sprintf("## Episode %s\n\n%s\n\n---\n\n", num, string(text))
		}
	}
	ioutil.WriteFile(filepath.Join(outDir, outFile), []byte(content), 0644)
}

func main() {
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)

	compileAct("Act-1-Suspicion.md", "Act I: Suspicion & Origin", "제1막: 의심과 기원", 1, 7)
	compileAct("Act-3-Trial.md", "Act III: Trial & Wargames", "제3막: 시련과 방어", 8, 11)
	compileAct("Act-4-Proof.md", "Act IV: Proof & Embody", "제4막: 증명과 체현", 12, 16)
	compileAct("Act-5-Declaration.md", "Act V: Declaration & Ultraplan", "제5막: 선언과 울트라플랜", 17, 22)

	vfsText, _ := ioutil.ReadFile(filepath.Join(brainDir, "walkthrough.md"))
	ioutil.WriteFile(filepath.Join(outDir, "Jloot-VFS-Architecture.md"), []byte("# [ENG] Jloot VFS Architecture\n# [KOR] Jloot VFS 아키텍처 명세\n\n"+string(vfsText)), 0644)

	potText, _ := ioutil.ReadFile(filepath.Join(brainDir, "blog_23_the_100_potentials.md"))
	ioutil.WriteFile(filepath.Join(outDir, "The-100-Potentials.md"), []byte("# [ENG] 100 Potentials (Ultraplan)\n# [KOR] 100가지 잠재력 (울트라플랜)\n\n"+string(potText)), 0644)

	sidebar := `# 🌐 Navigation (한/영 목차)

## 📖 Chronicles (연대기)
* [Act 1: Suspicion (의심)](Act-1-Suspicion)
* [Act 3: Trial (시련)](Act-3-Trial)
* [Act 4: Proof (증명)](Act-4-Proof)
* [Act 5: Declaration (선언)](Act-5-Declaration)

## ⚙️ Core Architecture (코어 엔진)
* [Jloot VFS & Sandbox (가상 파일 시스템)](Jloot-VFS-Architecture)
* [Ultraplan Potentials (비즈니스 잠재력)](The-100-Potentials)
`
	ioutil.WriteFile(filepath.Join(outDir, "_Sidebar.md"), []byte(sidebar), 0644)

	home := `# Welcome to NeuronFS (OS for AGI)
## 환영합니다. NeuronFS 글로벌 저장소입니다.

**[ENG]** NeuronFS is an isolated, zero-trust virtual filesystem designed entirely for Advanced General Intelligence (AGI). It physically manifests AI synapses onto encrypted layer cartridges (Jloot OverlayFS) and establishes deterministic "File-as-Neuron" structures. By combining XChaCha20 cryptography with an O(1) routing brainstem, NeuronFS ensures complete psychological safety and uncompromised neuroplasticity for agentic memory models.

**[KOR]** NeuronFS는 AGI 환경을 위해 설계된 제로 트러스트(Zero-Trust) 가상 파일 시스템 기법입니다. 암호화된 계층형 카트리지(Jloot OverlayFS) 상에 인공지능의 시냅스를 물리적인 폴더와 파일 구조로 발현시킵니다. 시스템은 XChaCha20 암호화와 O(1) 라우팅 뇌간 로직을 통해 에이전틱 메모리 모델의 완전한 심리적 안전(Psychological Safety)과 투명한 신경가소성을 보장합니다.

---
### 📚 Start Exploring (탐색 시작)
Please use the navigation **Sidebar** on the right to dive into the technical chronicles and proofs. 
우측의 **사이드바(Sidebar)**를 통해 기술 탄생 연대기와 아키텍처 명세로 진입해 주십시오.
`
	ioutil.WriteFile(filepath.Join(outDir, "Home.md"), []byte(home), 0644)

	fmt.Println("[SUCCESS] Wiki files compiled in Go.")
}
