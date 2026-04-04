import puppeteer from 'puppeteer-core';
import fs from 'fs';

// [V4 Architecture] CDP Starter Motor (초소형 방아쇠)
// 텍스트를 입력하고 전송 버튼만 누릅니다. (결과 스크래핑 X, 승인 버튼 클릭 X)

const WS_ENDPOINT_FILE = "C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\bot1_ws.txt";

async function injectAndTrigger(promptText) {
    if (!fs.existsSync(WS_ENDPOINT_FILE)) {
        console.error("CDP Endpoint file not found.");
        return;
    }

    const wsUrl = fs.readFileSync(WS_ENDPOINT_FILE, 'utf-8').trim();
    const browser = await puppeteer.connect({ browserWSEndpoint: wsUrl, defaultViewport: null });
    
    // 대상 패널(Antigravity 등) 찾기 (가장 흔한 웹뷰 패턴)
    const targets = await browser.targets();
    let targetPageTarget = targets.find(t => t.url().includes('webview') || t.title().includes('Antigravity'));
    
    if (!targetPageTarget) {
        console.log("No valid Antigravity Webview Target Found.");
        browser.disconnect();
        return;
    }

    const page = await targetPageTarget.page();
    
    console.log("[V4 Starter] Tapping into target to inject prompt...");

    // 1. 에러 발생 시 자동 Retry 클릭 (안전장치)
    try {
        const retryBtn = await page.$('button[title*="Retry"], button:contains("Retry")');
        if (retryBtn) {
            console.log("[!] Retry button detected. Clicking...");
            await retryBtn.click();
            await new Promise(r => setTimeout(r, 2000));
        }
    } catch(e) {}

    // 2. 텍스트 입력 및 전송
    // (셀렉터는 환경에 맞게 조정 필요: textarea, input 등)
    try {
        await page.evaluate((text) => {
            const el = document.querySelector('textarea, div[contenteditable="true"]');
            if (el) {
                if (el.tagName === 'TEXTAREA') el.value = text;
                else el.innerText = text;
                // React/Vue 이벤트 트리거
                el.dispatchEvent(new Event('input', { bubbles: true }));
                
                // 전송 버튼 클릭 코어 로직
                setTimeout(() => {
                    const sendBtn = document.querySelector('button[type="submit"], button[title*="Send"]');
                    if (sendBtn) sendBtn.click();
                }, 500);
            }
        }, promptText);

        console.log("[V4 Starter] Prompt Injected & Disconnected. The Hijacker will handle the rest.");
    } catch (e) {
        console.error("[X] Injection Failed: ", e.message);
    }
    
    browser.disconnect();
}

// 프로세스 인자로 받은 메시지 전송
const msg = process.argv[2] || "Proceed with Next Batch";
injectAndTrigger(msg);
