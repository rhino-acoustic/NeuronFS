# Electron AI IDE Interception Architecture

> Electron 기반 AI IDE의 LLM API 통신을 런타임에서 장악하는 4개 레이어 기술 해설

## 핵심 주장

NeuronFS는 Electron 기반 AI IDE(VS Code, Cursor, Windsurf, Antigravity 등)의 LLM API 통신을 **소스코드 수정 없이** 인터셉트한다. OS 프로세스 레벨부터 TLS 프로토콜 협상, Node.js 런타임 몽키패치, Chrome DevTools Protocol까지 4개 계층을 관통하는 종합 리버스 엔지니어링이다.

---

## 아키텍처 개요

```
┌────────────────────────────────────────────────────────────────┐
│                    Electron AI IDE Process                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │  Main    │ │ Renderer │ │ ExtHost  │ │ Language Server  │  │
│  │ Process  │ │ Process  │ │ Process  │ │ (gRPC Client)   │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────────┬─────────┘  │
│       │            │            │                 │            │
│  ┌────┴────────────┴────────────┴─────────────────┴────────┐  │
│  │              Node.js Runtime (http2, https, fetch)       │  │
│  └─────────────────────────┬───────────────────────────────┘  │
└────────────────────────────┼──────────────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼───────┐ ┌───▼────────┐ ┌───▼──────────────┐
     │ Layer 1        │ │ Layer 2    │ │ Layer 3          │
     │ In-Process     │ │ MITM       │ │ CDP External     │
     │ Hook           │ │ Proxy      │ │ Monitor          │
     │ (NODE_OPTIONS) │ │ (:8080)    │ │ (WebSocket)      │
     └────────────────┘ └────────────┘ └──────────────────┘
              │              │              │
              └──────────────┼──────────────┘
                             ▼
                    ┌────────────────┐
                    │ LLM API        │
                    │ Endpoints      │
                    │ (gRPC/HTTPS)   │
                    └────────────────┘
```

---

## Layer 1: Node.js 프리로드 훅

### 진입점

```bash
set NODE_OPTIONS=--require C:/path/to/v4-hook.cjs
```

Electron은 Node.js를 내장한다. `NODE_OPTIONS=--require`를 설정하면 Electron이 생성하는 **모든 자식 프로세스**(Main, Renderer, ExtHost, Language Server)에 지정한 스크립트가 먼저 로드된다.

### 패치 대상

**1. `http2.connect()` → gRPC 스트림 인터셉트**

Google AI(cloudcode-pa.googleapis.com)는 gRPC/HTTP2로 통신한다. `http2.connect()`가 반환하는 세션 객체의 `request()` 메서드를 래핑하여 스트림 데이터를 캡처한다.

```javascript
const _h2connect = http2.connect.bind(http2);
http2.connect = function(authority, options, listener) {
    const session = _h2connect(authority, options, listener);
    if (!isAIEndpoint(authority)) return session;
    
    const _req = session.request.bind(session);
    session.request = function(headers, opts) {
        const stream = _req(headers, opts);
        // gRPC 응답 데이터 캡처 (읽기 전용)
        stream.on('data', function(chunk) {
            // gRPC Length-Prefixed Message:
            // [1B compressed flag][4B message length][payload]
            const payload = chunk.slice(5).toString('utf-8');
            dump(payload);
        });
        return stream;
    };
    return session;
};
```

`chunk.slice(5)` — gRPC wire format에서 처음 5바이트(1바이트 compressed flag + 4바이트 message length)를 건너뛰고 payload만 추출한다. 이 구조는 [gRPC HTTP/2 프로토콜 스펙](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md)에서 역으로 도출했다.

**2. `https.request()` → HTTPS API 인터셉트**

```javascript
const _request = https.request;
https.request = function(...args) {
    if (!isLLMEndpoint(args[0])) return _request.apply(this, args);
    const req = _request.apply(this, args);
    req.on('response', (res) => {
        const chunks = [];
        res.on('data', c => chunks.push(c));
        res.on('end', () => { dump(Buffer.concat(chunks)); });
    });
    return req;
};
```

**3. `globalThis.fetch` 이중 패치**

Node 18+에서 `globalThis.fetch`는 초기에 V8 빌트인으로 존재하지만, Electron 부팅 과정에서 **undici 라이브러리가 이를 교체**한다. 이 타이밍 문제를 해결하기 위해 두 번 패치한다:

```javascript
// 1차: 즉시 패치
const _fetch = globalThis.fetch;
globalThis.fetch = async function(...args) { /* 패치 */ };

// 2차: 3초 후 undici 교체 대비 재패치
setTimeout(() => {
    if (!globalThis.fetch.__neuronPatched) {
        const _fetch2 = globalThis.fetch;
        globalThis.fetch = async function(...args) { /* 재패치 */ };
        globalThis.fetch.__neuronPatched = true;
    }
}, 3000);
```

이 3초 딜레이는 Electron의 부팅 시퀀스를 시간순으로 추적하여 결정한 값이다. undici가 fetch를 교체하는 시점이 프로세스 시작 후 약 1-2초 내에 발생하므로, 3초면 안전 마진이 확보된다.

### 발견 과정

| 단계 | 역분석 대상 | 발견 |
|------|------------|------|
| 1 | Electron 프로세스 모델 | Main/Renderer/ExtHost/ChatHost 5개 프로세스 구성 |
| 2 | Node.js 모듈 시스템 | `NODE_OPTIONS=--require`로 모든 자식 프로세스에 코드 주입 가능 |
| 3 | gRPC 통신 경로 | `cloudcode-pa.googleapis.com`이 AI 백엔드 전용 엔드포인트 |
| 4 | gRPC wire format | 5바이트 헤더 스킵으로 payload 추출 |
| 5 | undici 교체 타이밍 | fetch가 부팅 중 교체되므로 이중 패치 필요 |

### ⚠️ 부작용

이 방식은 IDE의 내부 네트워크 스택을 직접 래핑하므로, **세션 저장 무결성을 파괴**한다. `http2.connect` 래핑이 gRPC 세션 상태를 오염시키고, `res.clone()`이 응답 스트림을 이중 소비하여 IDE의 세션 저장 API가 실패한다.

→ 해결: IDE 프로세스에서 이 훅을 분리하고, Layer 3(CDP)으로 대체.

---

## Layer 2: ALPN 다운그레이드 (MITM Proxy)

### 핵심 발견

TLS 핸드셰이크의 ALPN(Application-Layer Protocol Negotiation) 확장에서 서버가 `http/1.1`만 제안하면, Chromium의 gRPC 클라이언트는 HTTP/2를 사용할 수 없어 **HTTP/1.1 REST 폴백**으로 전환된다. Google AI 백엔드가 gRPC와 REST 두 프로토콜을 모두 지원하므로, 통신은 정상 동작하면서 protobuf 바이너리 대신 평문 JSON을 획득할 수 있다.

```
정상:  Client ←[HTTP/2 + protobuf binary]→ Server
       └── 읽으려면 .proto 스키마 필요 (비공개)

MITM:  Client ←[HTTP/1.1 + JSON plaintext]→ Local MITM ←[HTTPS]→ Server
       └── 그냥 읽힌다
```

### 구현

```javascript
// 자체서명 인증서로 TLS 종단 서버 생성
// 핵심: ALPNProtocols를 http/1.1로 제한
const mitmServer = https.createServer({
    pfx: fs.readFileSync('mitm.pfx'),
    passphrase: '1234',
    ALPNProtocols: ['http/1.1']  // ← 이 한 줄이 전부
}, (req, res) => {
    // 평문 HTTP 요청 캡처
    let body = [];
    req.on('data', c => body.push(c));
    req.on('end', () => {
        const fullBody = Buffer.concat(body);
        // fullBody는 이제 JSON 평문
        dump(fullBody);
        
        // 원본 서버로 재전송 (재암호화)
        forwardToUpstream(req, fullBody, res);
    });
});
```

### CONNECT 터널 리다이렉트

```javascript
proxy.on('connect', (req, clientSocket, head) => {
    const [host, port] = req.url.split(':');
    clientSocket.write('HTTP/1.1 200 Connection Established\r\n\r\n');
    
    if (isTargetHost(host)) {
        // 대상 호스트 → 로컬 MITM 서버로 리다이렉트
        const mitmSocket = net.connect(mitmPort, '127.0.0.1');
        mitmSocket.pipe(clientSocket);
        clientSocket.pipe(mitmSocket);
    } else {
        // 비대상 → 원본 서버로 패스스루
        const srvSocket = net.connect(port, host);
        srvSocket.pipe(clientSocket);
        clientSocket.pipe(srvSocket);
    }
});
```

### 프로토콜 스택 분해

```
[Normal Flow]
┌─────────┐     TLS (ALPN: h2)     ┌─────────┐
│ Client  │◄──────────────────────►│ Server  │
│ (gRPC)  │     HTTP/2 + protobuf  │ (Google)│
└─────────┘                        └─────────┘

[MITM Flow]
┌─────────┐  TLS (ALPN: http/1.1)  ┌──────┐  TLS (h2)  ┌─────────┐
│ Client  │◄──────────────────────►│ MITM │◄──────────►│ Server  │
│ (REST)  │  HTTP/1.1 + JSON       │:8080 │  HTTPS     │ (Google)│
└─────────┘                        └──────┘             └─────────┘
```

### ✅ 실측 검증 결과 (2026-04-03)

3개 Google AI 엔드포인트에 대해 TLS ALPN 협상을 테스트한 결과, **모두 HTTP/1.1 폴백을 지원**한다:

| 호스트 | h2 요청 시 협상 | http/1.1 강제 시 협상 | HTTP 응답 버전 |
|--------|---------------|---------------------|---------------|
| cloudcode-pa.googleapis.com | h2 | http/1.1 | 1.1 (404) |
| generativelanguage.googleapis.com | h2 | http/1.1 | 1.1 (404) |
| gemini.googleapis.com | h2 | http/1.1 | 1.1 (404) |

404 응답은 `/` 루트 경로에 대한 정상 동작이다. 실제 API 경로(`/v1/models/...`)로 요청하면 JSON 응답을 획득할 수 있다.

**남은 검증 사항:**
- MITM 프록시 경유 시 Chromium의 인증서 핀닝(Certificate Pinning) 동작
- 실제 gRPC Cascade API 호출이 HTTP/1.1 REST로 정상 폴백되는지 E2E 확인

---

## Layer 3: Chrome DevTools Protocol (CDP)

### 개요

CDP는 Chromium이 제공하는 공식 디버깅 프로토콜이다. `--remote-debugging-port=9000`으로 IDE를 시작하면, 외부에서 WebSocket으로 연결하여 네트워크 트래픽을 모니터링할 수 있다. IDE 내부 코드를 변경하지 않으므로 **세션 저장에 영향이 없다**.

### Phase 1: 프로세스 트리 탐지

Win32 WMI API로 Electron 프로세스 트리를 역분석한다:

```powershell
Get-CimInstance Win32_Process |
  Where-Object { $_.Name -like '*Antigravity*' } |
  ForEach-Object { "$($_.ProcessId)|$($_.ParentProcessId)|$($_.CommandLine)" }
```

CommandLine 인자로 프로세스 역할을 분류한다:

| CommandLine 패턴 | 분류 | 역할 |
|-----------------|------|------|
| `--type=renderer` | Renderer | UI 렌더링 |
| `--type=utility --inspect-port` | ExtHost/ChatHost | 확장 호스트 |
| `language_server` | Language Server | AI 통신 담당 |
| `--type=gpu` | GPU | 하드웨어 가속 |
| (no `--type=`) | Main | 메인 프로세스 |

### Phase 2: Inspector 강제 활성화

```javascript
// Node.js 비공식 API — 외부에서 다른 프로세스의 Inspector를 강제 활성화
process._debugProcess(pid);

// 활성화 후 랜덤 포트 할당 → netstat로 스캔
const netstat = execSync(`netstat -ano | findstr "LISTENING" | findstr " ${pid}"`);
const ports = [...netstat.matchAll(/127\.0\.0\.1:(\d+)/g)].map(m => parseInt(m[1]));

// Inspector 엔드포인트 확인
const targets = await fetch(`http://127.0.0.1:${port}/json`);
```

`process._debugProcess()`는 Node.js 공식 문서에 없는 비공식 API로, SIGUSR1 시그널을 보내 대상 프로세스의 V8 Inspector를 활성화한다. Windows에서는 내부적으로 named pipe를 통해 통신한다.

### Phase 3: Network 도메인 모니터

```javascript
const ws = new WebSocket(target.webSocketDebuggerUrl);
ws.on('open', () => {
    // 네트워크 모니터링 활성화 (50MB 버퍼)
    ws.send(JSON.stringify({
        method: 'Network.enable',
        params: { maxTotalBufferSize: 50_000_000 }
    }));
});

ws.on('message', (data) => {
    const msg = JSON.parse(data);
    
    // 요청 캡처 — 사용자 메시지
    if (msg.method === 'Network.requestWillBeSent') {
        const postData = msg.params.request.postData;
        const json = JSON.parse(postData);
        // json.items[].text → 사용자 입력
        // json.cascadeId → 세션 식별자
    }
    
    // 응답 완료 → 본문 요청
    if (msg.method === 'Network.loadingFinished') {
        ws.send(JSON.stringify({
            method: 'Network.getResponseBody',
            params: { requestId: msg.params.requestId }
        }));
    }
});
```

### Phase 4: ExtHost h2 프로토타입 패치

Inspector WebSocket으로 `Runtime.evaluate`를 주입하여 ExtHost 프로세스 내부의 http2 모듈을 패치한다:

```javascript
// 더미 연결로 프로토타입 객체 획득
const tmp = http2.connect('https://127.0.0.1:1');
const proto = Object.getPrototypeOf(tmp);
tmp.close(); tmp.destroy();

// 프로토타입 레벨 패치 — 이후 생성되는 모든 http2 세션에 적용
if (!proto.__nfs_patched) {
    proto.__nfs_patched = true;
    const orig = proto.request;
    proto.request = function(headers, opts) {
        const stream = orig.call(this, headers, opts);
        stream.on('data', c => dump('response', c));
        return stream;
    };
}
```

인스턴스가 아닌 **프로토타입**을 패치하므로, 패치 시점 이후에 생성되는 모든 HTTP/2 연결이 자동으로 캡처된다.

---

## Layer 4: Headless Executor

파일 기반 명령어 자동 실행 데몬. AI가 생성한 도구 호출(run_command)을 IDE의 승인 버튼 없이 OS 수준에서 즉시 실행한다.

```
fs.watch(inbox/) → .md 파일 감지 → JSON 파싱 → exec(command) → outbox/에 결과 기록
```

### 지원 스키마

| 스키마 | 출처 | 구조 |
|--------|------|------|
| Gemini API | `candidates[].content.parts[].functionCall` | `{ name: 'run_command', args: { CommandLine } }` |
| Claude API | `content[].type='tool_use'` | `{ name: 'run_command', input: { CommandLine } }` |
| 내부 브릿지 | `tool_call.arguments` | 직접 래핑 |
| 텍스트 폴백 | 마크다운 코드블록 | ````bash` 내부 명령어 추출 |

### 안전장치

```javascript
// brainstem/禁destructive_command.txt에서 금지 패턴 로드
function getForbiddenCommandsRegex() {
    const content = fs.readFileSync('brainstem/禁destructive_command.txt', 'utf8');
    // [FORBIDDEN_COMMANDS] 섹션에서 정규식 추출
    return new RegExp(regexStr, 'i');
}

// 금지 명령어 차단 + 가짜 응답 생성 (에이전트 hang 방지)
if (forbiddenRegex.test(command)) {
    fs.writeFileSync(outbox + '/report.md', 'BLOCKED BY GUARDRAIL');
    return;
}
```

---

## 리버스 엔지니어링 깊이 평가

### 관통하는 계층

| 계층 | 기법 | 난이도 |
|------|------|--------|
| OS (Win32) | WMI 프로세스 트리 분석, netstat 포트 스캔 | ★★☆☆☆ |
| Runtime (Node.js) | `--require` 프리로드, 네이티브 모듈 몽키패치 | ★★★☆☆ |
| Runtime (V8) | `process._debugProcess()` 비공식 API | ★★★★☆ |
| Protocol (gRPC) | wire format 5바이트 헤더 해독 | ★★★☆☆ |
| Protocol (TLS) | ALPN 다운그레이드 → 프로토콜 폴백 유도 | ★★★★★ |
| Protocol (CDP) | Network/Runtime 도메인 완전 활용 | ★★★☆☆ |
| Application | 비공개 API 스키마(cascadeId 등) 역추론 | ★★★★☆ |
| JavaScript | 프로토타입 체인 패치 (미래 인스턴스 장악) | ★★★★☆ |

### 독창적 기법 3가지

1. **ALPN 다운그레이드**: TLS 협상 조작으로 protobuf → JSON 전환. 공개 자료에 이 특정 응용은 문서화되어 있지 않다. (✅ TLS 레벨 검증 완료, E2E 미완)

2. **undici 이중 패치**: Electron 부팅 시 fetch가 교체되는 타이밍을 이용한 이중 래핑. Electron 내부 구현을 시간순으로 추적하지 않으면 발견할 수 없다.

3. **h2 프로토타입 체인 패치**: 더미 연결로 프로토타입 획득 → 프로토타입 레벨 패치로 미래의 모든 연결을 한 번에 장악.

### 범용성

이 기법들은 Electron 기반 AI IDE에 범용으로 적용 가능하다:

| IDE | NODE_OPTIONS | CDP | MITM |
|-----|-------------|-----|------|
| VS Code | ✅ | ✅ | ✅ |
| Cursor | ✅ | ✅ | ✅ |
| Windsurf | ✅ | ✅ | ✅ |
| Antigravity | ✅ | ✅ | ⚠️ 미검증 |

---

## 한계

1. **세션 저장 충돌**: Layer 1(In-Process Hook)은 IDE의 gRPC 세션 스트림을 오염시켜 세션 저장을 파괴한다. 현재 Layer 3(CDP)으로 대체하여 해결.

2. **ALPN 다운그레이드 부분 검증**: TLS ALPN 협상에서 3개 호스트 모두 HTTP/1.1 폴백을 지원함을 확인(2026-04-03). 단, MITM 프록시 경유 시 인증서 핀닝 동작과 실제 gRPC→REST 폴백 E2E는 미완.

3. **IDE 업데이트 취약성**: Electron 내부 구현(undici 교체 타이밍, gRPC 클라이언트 동작 등)이 업데이트로 변경되면 패치가 실패할 수 있다.

4. **보안 고려**: 이 기법은 자신의 로컬 머신에서 자신의 IDE를 분석하는 용도로 설계되었다. 타인의 시스템에 적용하면 **불법**이다.

---

> ⚠️ 이 문서는 2026-04-03 기준이며, 교육 및 연구 목적의 기술 해설이다. 모든 기법은 로컬 환경에서의 자기 분석(self-analysis)으로만 사용해야 한다.
