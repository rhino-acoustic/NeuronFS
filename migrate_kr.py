import os, shutil

brain = r"C:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4"

# Windows 파일명에 : 사용 불가 → 한자 접두어 사용 (기존 패턴 준수)
renames = {
    # cortex tools: 영어 → 한국어 (한자 접두어 패턴)
    r"cortex\frontend\coding\no_console_log": "禁console_log",
    r"cortex\tools\avoid_general_commands": "禁general_commands",
    r"cortex\tools\avoid_ls": "禁ls_usage",
    r"cortex\tools\avoid_g": "禁g_usage",
    r"cortex\tools\avoid_ls_and_g": "禁ls_and_g",
    r"cortex\tools\avoid_ls_for_directories": "禁ls_for_directories",
    r"cortex\tools\adopt_list_dir": "推list_dir",
    r"cortex\tools\adopt_precise_tools": "推precise_tools",
    r"cortex\tools\adopt_specific_tools": "推specific_tool_usage",
    r"cortex\tools\list_dir_instead_of_ls": "推list_dir_instead",
    r"cortex\tools\precise_tool_usage": "推precise_tool_usage",
    r"cortex\tools\specific_tool_usage": "推specific_tool_usage",
    r"cortex\tools\use_list_dir": "推use_list_dir",
    r"cortex\tools\use_precise_tools": "推use_precise_tools",
    r"cortex\tools\use_specific_tools": "推use_specific_tools",
    r"cortex\tool_usage\precise_tools_only": "推precise_tools_only",
    # cortex misc
    r"cortex\communication": "소통",
    r"cortex\thought": "사고",
    r"cortex\strategy\strategic_depth": "전략적_깊이",
    r"cortex\security\least_agency_principle": "최소권한_원칙",
    r"cortex\neuronfs\naming": "명명규칙",
    r"cortex\neuronfs\emit_kanji_dedup": "한자_중복제거",
    r"cortex\neuronfs\dual_gemini_sync": "듀얼_gemini_동기화",
    r"cortex\neuronfs\runtime\idle_auto_decay": "유휴자동감쇠",
    r"cortex\neuronfs\runtime\modtime_sync_fixed": "수정시간_동기화",
    r"cortex\skills\crawler\instagram_cdp_pipeline": "인스타_CDP_파이프라인",
    r"cortex\frontend\typography": "타이포그래피",
    r"cortex\neuronfs\design": "설계",
    r"cortex\neuronfs\ops": "운영",
    r"cortex\backend\devops": "데브옵스",
    # ego
    r"ego\proverbs": "격언",
    r"ego\communication\concise_execution": "간결_실행중심",
    r"ego\communication\structured_and_systematic": "구조화_체계적",
    r"ego\language\korean_thought": "한국어_사고",
    # hippocampus
    r"hippocampus\methodology": "방법론",
    r"hippocampus\quality": "품질",
    r"hippocampus\session_log": "세션로그",
    # sensors
    r"sensors\brand": "브랜드",
    r"sensors\environment": "환경",
    # prefrontal
    r"prefrontal\project\github_public_preparation": "깃허브_공개_준비",
    r"prefrontal\todo\groq_auto_neuronize": "Groq_자동뉴런화",
}

done = skip = merged = 0
for rel, new_name in renames.items():
    old = os.path.join(brain, rel)
    if not os.path.exists(old):
        skip += 1
        continue
    parent = os.path.dirname(old)
    new = os.path.join(parent, new_name)
    if os.path.exists(new):
        def max_counter(d):
            m = 0
            for f in os.listdir(d):
                if f.endswith('.neuron'):
                    try: m = max(m, int(f.replace('.neuron','')))
                    except: pass
            return m
        o = max_counter(old)
        n = max_counter(new)
        t = o + n
        for f in os.listdir(new):
            if f.endswith('.neuron'):
                os.remove(os.path.join(new, f))
        open(os.path.join(new, f"{t}.neuron"), 'w').close()
        shutil.rmtree(old)
        print(f"MERGE: {rel} -> {new_name} ({o}+{n}={t})")
        merged += 1
    else:
        os.rename(old, new)
        print(f"RENAME: {rel} -> {new_name}")
        done += 1

print(f"\nResult: {done} renamed, {merged} merged, {skip} skipped")
