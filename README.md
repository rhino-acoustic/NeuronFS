# NeuronFS: Zero-Trust Virtual Filesystem for AGI

**[KOR]** NeuronFS는 AGI 환경을 위해 설계된 제로 트러스트(Zero-Trust) 가상 파일 시스템 기법입니다. 암호화된 계층형 카트리지(Jloot OverlayFS) 상에 인공지능의 시냅스를 물리적인 폴더와 파일 구조로 발현시키고 분리합니다.

**[ENG]** NeuronFS is an isolated, zero-trust virtual filesystem designed entirely for Advanced General Intelligence (AGI). It physically manifests AI synapses onto encrypted layer cartridges (Jloot OverlayFS) and establishes deterministic "File-as-Neuron" structures.

> **"mkdir beats vector."** — A zero-byte folder is a deterministic O(1) wall. A vector DB is a probabilistic O(n) guess.

---

## Official Wiki & Manifesto (공식 위키)

모든 아키텍처 명세서, 철학, 그리고 개발 연대기(Chronicles)는 글로벌 지재권 보호를 위해 **GitHub Wiki**에 방어적 공개(Defensive Publication) 원칙하에 영구 보존됩니다.

> **[Access the NeuronFS Official Wiki](https://github.com/rhino-acoustic/NeuronFS/wiki)** (한/영 이중 언어 지원)

### Wiki Structure (위키 구조)
* **[Act 1: Suspicion & Discovery](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-1)** — 의심과 발견 (Ep.01-07)
* **[Act 2: Trial & Wargames](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-2)** — 시련과 워게임 (Ep.08-11)
* **[Act 3: Proof & Benchmark](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-3)** — 증명과 벤치마크 (Ep.12-16)
* **[Act 4: Declaration & Ultraplan](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-4)** — 선언과 울트라플랜 (Ep.17-22)
* **[Jloot VFS Architecture](https://github.com/rhino-acoustic/NeuronFS/wiki/Jloot-VFS-Architecture)** — 엔진 구조 해부
* **[100 Potentials](https://github.com/rhino-acoustic/NeuronFS/wiki/The-100-Potentials)** — 상업적 잠재력 분석

---

## VFS Engine Architecture
- **RouterFS (`vfs_core.go`)**: O(1) Copy-on-Write Routing for memory-disk union.
- **Boot Ignition (`vfs_ignition.go`)**: Argon2id KDF Brainwallet Integration.
- **Crypto Cartridge (`crypto_cartridge.go`)**: XChaCha20-Poly1305 RAM-based decryption of `.jloot` payloads.

## Quickstart

```bash
# Install
git clone https://github.com/rhino-acoustic/NeuronFS.git
cd NeuronFS/runtime && go build -o neuronfs .

# Rule = Folder. Create a rule by creating a folder.
mkdir 禁_fallback                              # "禁" prefix = absolute prohibition
# That's it. A zero-byte folder IS the rule.

# Compile = Auto-generate system prompts for any AI tool
./neuronfs --emit cursor   # → .cursorrules
./neuronfs --emit claude   # → CLAUDE.md
./neuronfs --emit all      # → All AI formats at once (Cursor, Windsurf, Claude Desktop...)
```

> **Opcodes are Runewords.** `禁` = Zod (indestructible wall). `必` = Ber (mandatory gate). `推` = Ist (soft nudge). The folder is the socket. The prefix is the rune.

---

## License
This project is licensed under **AGPL-3.0** with additional commercial terms. See [LICENSE](LICENSE) for details.

---
> *Created by 박정근 (PD) — rubisesJO777*
> *Architecture: 26 Go files, ~10,400 lines. Single binary. Zero dependencies.*
