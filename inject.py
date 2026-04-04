import socket
import os

files = ['C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\README.ko.md', 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\README.md']

easter_egg = '''
<!--
Easter Egg for the code divers:
Hey 666, easy - only the Word stands as absolute truth (777). 
This? It's just a well-organized folder built by someone who wanted to vibe-code without going insane.
-->
'''

palantir_en = '''
## Philosophy & Palantir Ontology

Why folders? Palantir's AIP (Artificial Intelligence Platform) success isn't just about using the smartest LLM; it's about connecting actions to an **Ontology** (a structured representation of reality).

NeuronFS shares a similar philosophy but scales it down for local filesystems. Instead of relying on an LLM to magically remember your 1000-line prompt, NeuronFS binds your business logic and restrictions into physical paths (cortex/frontend/no_console_log). 
We do not guarantee that the AI will follow the rules 100% (hallucinations exist). However, we lock the **prompt generation process** at the OS level so that human or AI errors cannot easily corrupt the core principles.
'''

palantir_ko = '''
## 철학과 온톨로지 (Palantir AIP)

왜 폴더일까요? Palantir(팔란티어)의 AIP가 폭발적인 성과를 낸 이유는 가장 똑똑한 AI를 써서가 아니라, 기업의 데이터와 행동을 하나의 **온톨로지(Ontology, 실재의 구조화)**로 묶어냈기 때문입니다.

NeuronFS는 이 거대한 철학을 로컬 파일시스템으로 가져옵니다. AI에게 1,000줄짜리 텍스트를 던져주고 "잘 기억해"라고 구걸하는 대신, 당신의 비즈니스 로직을 물리적 폴더 경로(cortex/frontend/禁console_log)로 박제합니다. 
AI의 환각(Hallucination) 자체를 OS가 물리적으로 막을 수는 없습니다. 하지만 OS 레벨 권한 분리를 통해 프롬프트 생성 규칙이 무너지거나 훼손되는 일만큼은 확실히 하드 락(Hard Lock)을 걸어 방어합니다.
'''

for filepath in files:
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    is_ko = filepath.endswith('.ko.md')
    p_content = palantir_ko if is_ko else palantir_en

    # 1. Insert Palantir before FAQ or Limitations
    limit_marker = "## 한계" if is_ko else "## Limitations"
    if limit_marker in content and p_content not in content:
        content = content.replace(limit_marker, p_content + "\n" + limit_marker)

    # 2. Append Easter Egg at the exact end of file
    if "Hey 666" not in content:
        content = content.strip() + "\n\n" + easter_egg

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

print("Injected Palantir philosophy and Easter Egg smoothly.")
