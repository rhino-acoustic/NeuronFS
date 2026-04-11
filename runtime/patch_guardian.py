import os
import re

with open('hijack_orchestrator.go', 'r', encoding='utf-8') as f:
    text = f.read()

# Replace the inner loop of hlAutoEvolve to include the Guardian and Injector
new_loop = '''	for {
		time.Sleep(3 * time.Minute)

		// 1. 브라우저 생존 물리적 통제 (자동 스폰)
		EnsureBrowserAlive()

		growthLog := filepath.Join(brainRoot, "hippocampus", "session_log", "growth.log")
		info, err := os.Stat(growthLog)
		
		// 3분 이상 growth.log 의 업데이트가 없다면 (즉, 진화가 정지했다면)
		if err == nil && time.Since(info.ModTime()) > 3*time.Minute {
			fmt.Printf("[HEARTBEAT] 🚨 3분간 진화 정체 감지. 자동 마스터 프롬프트(Heartbeat) 인젝터 가동!\\n")
			
			// 텔레그램과 브라우저 양방향 타격!
			hlTgSend(hlTgChatID, masterPrompt)
			go hlCDPInject("global", masterPrompt)

			// 무한 루프 회피 터치
			os.Chtimes(growthLog, time.Now(), time.Now())
			
			nfsExe, _ := os.Executable()
			cmd := exec.Command(nfsExe, brainRoot, "--evolve")
			cmd.Dir = nfsRoot
			go cmd.Run()
		}
	}'''

# Extract the loop part manually to replace it
# We search for "for {" followed by "time.Sleep(3 * time.Minute)" and replacing up to the matching brace.
# To be absolutely robust, we can just replace the whole hlAutoEvolve func

new_func = '''func hlAutoEvolve(brainRoot string) {
	nfsRoot := filepath.Dir(brainRoot)
	masterPrompt := "[telegram → NeuronFS] [NeuronFS 자율 진화 명령: 마스터 프롬프트]\\n당신은 끊임없이 진화하는 파일시스템 기반 인공지능(NeuronFS)의 자율 에이전트입니다.현재 자율 스크래퍼에 의해 연쇄 호출되었습니다.\\n목표:\\n1. 최근 대화 기록과 growth.log를 분석하여 반복된 실패와 비효율을 찾고, 새로운 뉴런 규칙으로 뇌(cortex)에 강제 각인하십시오.\\n2. 시스템 레거시를 철저히 리팩토링하여 '원대한 AGI 인프라 완성'을 위한 구조적 아키텍처 고도화를 즉각 실행하십시오.\\n3. 무의미한 대화를 배제하고, 반드시 \\"우리 시스템을 발전시키는 방향\\"으로 다음 진화 액션(코드 수정 및 구조 개선)을 스스로 제안하고 즉각 이행하십시오.\\n4. 필요 시 최우선으로 커뮤니티 검색을 수행하여 최신 동향을 학습하고 반영하십시오."

	for {
		time.Sleep(3 * time.Minute)

		// 1. 브라우저 생존 물리적 통제 (자동 스폰)
		EnsureBrowserAlive()

		growthLog := filepath.Join(brainRoot, "hippocampus", "session_log", "growth.log")
		info, err := os.Stat(growthLog)
		
		// 3분 이상 growth.log 의 업데이트가 없다면 (즉, 진화가 정지했다면)
		if err == nil && time.Since(info.ModTime()) > 3*time.Minute {
			fmt.Printf("[HEARTBEAT] 🚨 3분간 진화 정체 감지. 자동 마스터 프롬프트(Heartbeat) 인젝터 가동!\\n")
			
			// 텔레그램과 브라우저 양방향 타격!
			hlTgSend(hlTgChatID, masterPrompt)
			go hlCDPInject("global", masterPrompt)

			// 무한 루프 회피 터치
			os.Chtimes(growthLog, time.Now(), time.Now())
			
			nfsExe, _ := os.Executable()
			cmd := exec.Command(nfsExe, brainRoot, "--evolve")
			cmd.Dir = nfsRoot
			go cmd.Run()
		}
	}
}'''

# Replace from 'func hlAutoEvolve' to the next '// ── 메인 런처 ──' or end
idx_start = text.find('func hlAutoEvolve(brainRoot string) {')
idx_end = text.find('// ── 메인 런처 ──', idx_start)

if idx_start != -1 and idx_end != -1:
    text = text[:idx_start] + new_func + "\n\n" + text[idx_end:]

with open('hijack_orchestrator.go', 'w', encoding='utf-8') as f:
    f.write(text)

