import os

with open('runtime/ide_integration.go', 'r', encoding='utf-8') as f:
    text = f.read()

# 만약 UTF-16 BOM 등으로 오염되어 읽기 실패 시, 에러 방지를 위해 utf-16 으로 열어보고 다시 utf-8로 씁니다.
try:
    with open('runtime/ide_integration.go', 'r', encoding='utf-16') as f:
        text = f.read()
except:
    pass

text = text.replace('svLoadTelegram', 'hlLoadTelegram')

with open('runtime/ide_integration.go', 'w', encoding='utf-8') as f:
    f.write(text)
