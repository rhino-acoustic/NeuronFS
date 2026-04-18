# 올리브영 2026 DOM 구조 (CDP 실측 + 성공 패턴)

## 실측일: 2026-04-08
## 방법: CDP Page.captureScreenshot + Runtime.evaluate

---

## ✅ 성공 추출 패턴 (검증 완료)

### 진입점: `.prd_name a` (절대 `.prd_info`로 순회하지 말 것!)

```
❌ 실패: document.querySelectorAll('.prd_info')  → 24개지만 빈 요소 포함 → 0개 추출
✅ 성공: document.querySelectorAll('.prd_name a') → 정확히 제품 수만큼 → 6/24개 추출
```

### 성공 JS 패턴:
```javascript
const nameEls = document.querySelectorAll('.prd_name a');
nameEls.forEach(a => {
    let name = a.textContent?.trim();
    let parent = a.closest('.prd_info') || a.closest('li');
    const price = parent?.querySelector('.tx_cur .tx_num')?.textContent?.replace(/[^0-9]/g,'');
    const img = parent?.querySelector('.prd_thumb img')?.src;
    const link = parent?.querySelector('.prd_thumb')?.href;
});
```

### 핵심 교훈:
- `.prd_info` 24개 중 18개는 사이드바/최근 본 상품 등의 빈 래퍼
- `.prd_name a`가 유일하게 실제 제품만 매칭되는 셀렉터
- `a.closest('.prd_info')`로 부모를 역추적하면 안전

---

## ✅ 브랜드 일괄 전략 (14분 → 1분)

### 문제:
제품별 개별 검색(8회 CDP 호출) → TimeoutError 반복 → 14분 소요

### 해결:
브랜드명 1회 검색으로 전 제품 일괄 추출 → fuzzy 매칭 → **1분 완료**

```
Step 1: 고유 브랜드명 추출 (누그레이, 롬앤, 페리페라)
Step 2: 브랜드별 1회 올리브영 검색 → .prd_name a 기반 전 제품 일괄 추출
Step 3: 각 TARGET_PRODUCT에 fuzzy 매칭 (쿼리 키워드 ∩ 올리브영 제품명)
```

### 결과:
- **누그레이 6개** 전 제품: 가격 + 이미지 ✅
- **롬앤 24개** 전 제품 ✅
- **페리페라 24개** 전 제품 ✅

---

## 핵심 셀렉터 (실측 적중률 순)

| 셀렉터 | 개수 | 용도 | 상태 |
|--------|------|------|------|
| `.prd_name a` | 6 | ⭐⭐ 제품 진입점 (성공) | ✅ 최우선 |
| `.prd_info` | 24 (6 유효) | 제품 래퍼 (빈 요소 많음) | ⚠ 직접 순회 금지 |
| `.tx_cur .tx_num` | - | 할인가 | ✅ 정확 |
| `.prd_thumb img` | - | 제품 이미지 | ✅ 정확 |
| `.prd_thumb` href | - | 제품 상세 링크 | ✅ 정확 |
| `[class*="goods"]` | 0 | ❌ 폐기 | ❌ 2026 없음 |

---

## 검색 URL 패턴
`https://www.oliveyoung.co.kr/store/search/getSearchMain.do?query={브랜드명}`

## 제품 상세 URL 패턴
`https://www.oliveyoung.co.kr/store/goods/getGoodsDetail.do?goodsNo={goodsNo}`

## CDP 설정
- navigate 후 5~7초 대기
- scroll 3회
- recv timeout 60초
