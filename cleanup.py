import re

files = ['README.ko.md', 'README.md']

# List of common emojis to remove
emoji_pattern = re.compile(r'[\U00010000-\U0010ffff]|\u26A0\uFE0F|\u2728|\u2705|\u274C|\u2699\uFE0F|[\u2700-\u27BF]|[\uE000-\uF8FF]|\uD83C[\uDF00-\uDFFF]|\uD83D[\uDC00-\uDE4F]|\uD83D[\uDE80-\uDEF6]|\uD83E[\uDD10-\uDDFF]')

for filepath in files:
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # 1. Remove emojis
    content = emoji_pattern.sub('', content)
    # Cleanup extra spaces left by emoji removal
    content = content.replace('  ', ' ')
    content = content.replace('> v4.3', '> v4.3') # Just ensuring
    content = content.replace(' NeuronFS', 'NeuronFS')

    # 2. Move Quickstart to top
    # We need to find the Quickstart section and move it right below the 릴리즈 노트 or TL;DR. 
    # Actually, the user asked "설치 관련 내용이 맨위로 올려줘".
    # Let's place it right after the title (# NeuronFS ... compiler)
    
    # Identify the Quickstart section
    if filepath == 'README.ko.md':
        start_marker = "### 30초 시작"
        end_marker = "---"
        target_marker = "## 요약 (TL;DR)"
    else:
        start_marker = "### Quickstart"
        end_marker = "---"
        target_marker = "## TL;DR"

    if start_marker in content:
        parts = content.split(start_marker)
        before_quickstart = parts[0]
        quickstart_and_after = parts[1]
        
        # split quickstart_and_after by the next "---"
        qs_parts = quickstart_and_after.split(end_marker, 1)
        quickstart_content = start_marker + qs_parts[0] + end_marker + "\n"
        after_quickstart = qs_parts[1]
        
        # Remove the original quickstart content from the before part just in case (already split)
        content_without_qs = before_quickstart + after_quickstart
        
        # Insert target
        if target_marker in content_without_qs:
            insert_parts = content_without_qs.split(target_marker)
            new_content = insert_parts[0] + quickstart_content + "\n" + target_marker + insert_parts[1]
            content = new_content

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

print('Done cleaning and restructuring READMEs.')
