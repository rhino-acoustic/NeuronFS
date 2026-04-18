# 올리브영 상세 페이지 DOM (CDP 실측 성공)

## URL 패턴
`https://www.oliveyoung.co.kr/store/goods/getGoodsDetail.do?goodsNo={goodsNo}`

## 핵심 셀렉터 (검증 완료)

| 데이터 | 셀렉터 | 예시 값 |
|--------|--------|---------|
| **평점** | `.rating` 또는 `[class*="rating-star"]` | `4.7` |
| **리뷰 수** | `[class*="review-count"]` | `리뷰 108건` |
| **보는 중** | `[class*="viewer-count"]` | `67명이 보고 있어요` |
| **리뷰탭 수** | `.GoodsDetailTabs_count__nz2tF` | `108` |

## 성공 JS 패턴:
```javascript
const ratingEl = document.querySelector('.rating, [class*="rating-star"]');
const rating = ratingEl?.innerText?.match(/(\d\.\d)/)?.[1];

const reviewEl = document.querySelector('[class*="review-count"]');
const reviewCount = reviewEl?.innerText?.match(/(\d[\d,]*)/)?.[1];

const viewerEl = document.querySelector('[class*="viewer-count"]');
const viewerCount = viewerEl?.innerText?.match(/(\d[\d,]*)/)?.[1];
```

## CDP 설정
- navigate 후 4~6초 대기 (상세 페이지는 React 렌더링)
- 스크롤 불필요 (상단에 평점/리뷰 표시)

## 실측 결과 (2026-04-08)
- 누그레이 세모 탭 치크: ⭐4.7, 108건, 73명 보는중
- 누그레이 에센셜 아이팔레트: ⭐4.9, 97건
- 누그레이 데일리 무드 마뜨: ⭐4.9, 337건
- 롬앤 쥬시 래스팅 틴트: ⭐4.7, 250,860건, 322명 보는중
