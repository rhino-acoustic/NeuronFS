package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// Entry represents a single knowledge entry in the AIA cartridge
type Entry struct {
	Path    string // cortex path: region/category/name
	Content string // neuron content
}

func main() {
	fmt.Println("[JLOOT BUILDER] 🏃 AIA Run Together 2026 카트리지 컴파일 시작...")

	entries := []Entry{
		// ─── 소스 인벤토리 ───
		{
			"cortex/aia/소스/R1_첫미팅_녹취",
			`소스: R1 — 1/20 첫 대면 미팅 (95분)
경로: c:\Users\BASEMENT_ADMIN\aia\docs\새로운 녹음 810.txt
날짜: 2026-01-20
참석: AIA 5명 (이서윤, 김태영, 심승부 외 2) / VEGA 2명 (김태경, 박정근)
장소: AIA 타워 21층
파싱:
- 14명 크루 계획 (MP반+고객반)
- 베가베리: 6,700명 커뮤니티/월9만 구독/17개 지역/유튜브 12만
- 삼성스토어 77회, 더현대닷컴, 파주마라톤 레퍼런스
- 트랙데이 월1회 목동 300-400명
- 촬영~편집 인하우스 (외주X)
- 월정액 모델 제안 (500~700만/월)
- "올해 파일럿, 27년에 크게"
- 하이원 정선 캠프 경험 (미즈노 라이트랩)`,
		},
		{
			"cortex/aia/소스/R2_후속통화",
			`소스: R2 — 1/24 후속 통화 (33분)
경로: c:\Users\BASEMENT_ADMIN\aia\archive\미팅메모\talk
날짜: 2026-01-24 토 18:03
참석: 이서윤, 류현정(첫등장), 김태경, 박정근
파싱:
- 류현정 첫 등장 (마케팅 팀원)
- 캠프 2박3일 확정, CMO 동행
- 콘텐츠: JTBC 직전/후 집중 방출
- 사장님까지 검수 ("셀프검수 없다")
- 제주 캠프 비용 견적 요청
- 브랜디드 콘텐츠 퀄리티 개선 필수`,
		},
		{
			"cortex/aia/소스/R3_3차미팅_녹취",
			`소스: R3 — 3/12 대면 미팅 (65분)
경로: c:\Users\BASEMENT_ADMIN\aia\archive\chat_logs\20260312 talk
날짜: 2026-03-12
참석: 이서윤, 김태경, 박정근, 류현정
파싱:
- 인원 31→20명 합의 (크루14+AIA6)
- 정승은 당일방문→이후 장동선 대체
- 콘텐츠 +α 추가 범위 합의
- 황싸부 6월까지만 (7월~ YG투어)
- 보험+앰뷸런스 필수
- 웰니스캠프 = AIA Only Experience 후보
- 인스타 AIA 비즈니스 계정 귀속 검토`,
		},
		{
			"cortex/aia/소스/K1_부장님방_카톡",
			`소스: K1 — 부장님방 카카오톡
경로: c:\Users\BASEMENT_ADMIN\aia\archive\chat_logs\KakaoTalk_20260413_0924_24_512_group.txt
기간: 2026-02-11 ~ 2026-04-10
참여: 이서윤, 김태영, 박정근, Ted, HJ(릴리)
핵심 이벤트:
- 2/11: 단톡방 개설, KPI 운영계획서
- 2/19: 2차 대면 미팅
- 2/26: 견적서 v1 (PDF+XLSX)
- 3/18: 4건 자료 정식 전달 (견적v11=133,200,000원)
- 4/1: 최종 승인 메일 (공식 GO)
- 4/10: 8개 아젠다, 4/16 Teams 미팅 확정`,
		},
		{
			"cortex/aia/소스/K2_과장별도방_카톡",
			`소스: K2 — 과장별도방 카카오톡
경로: c:\Users\BASEMENT_ADMIN\aia\archive\chat_logs\KakaoTalk_20260413_0924_39_231_group.txt
기간: 2026-03-19 ~ 2026-04-09
참여: 이서윤, 박정근, Ted
핵심 이벤트:
- 3/24: 박지혜 연락처, 장동선 500만, UGC 300건 확정
- 3/25: 트레이닝 일정, 션 하루만
- 3/26: 일정 3차 수정, 9/5 트랙데이 확정
- 3/27: 계약서 이메일, 캠프기획서, "예산은 저에게만"
- 4/6: Track Day 용어확정 (AIA 300명 vs 베가베리 기존)
- 4/9: 대관 현황 6곳, Track Day 9/12 확정`,
		},
		{
			"cortex/aia/소스/K3_인스타방_카톡",
			`소스: K3 — 인스타방 카카오톡
경로: c:\Users\BASEMENT_ADMIN\aia\archive\chat_logs\KakaoTalk_20260413_0924_51_952_group.txt
기간: 2026-03-25
참여: 이서윤, 송인재, 박정근, Ted
파싱:
- 송인재 첫 등장 (마케팅 파트너십&커뮤니케이션팀)
- 구글계정 기반 인스타 개설
- 기존 aia_together 삭제 → 공식계정 전환`,
		},
		{
			"cortex/aia/소스/M1_AIA피드백메일",
			`소스: M1 — AIA 공식 피드백 메일
경로: c:\Users\BASEMENT_ADMIN\aia\archive\미팅메모\20260217
날짜: 2/17 이후
6대 요청:
1. 크루장(안은태/박지혜) 3개 세션 모두 참여
2. HNW Wellness Club: 1회 파일럿 + JTBC VIP
3. SNS: Followers +3,000 / UGC +180
4. 캠프 제주 재제안 ("unrealistic experience" 부족)
5. 격주 1회 미팅 (목요일)
6. 견적서 전체 반영 + 옵션별 분리`,
		},
		{
			"cortex/aia/소스/M2_내부대응문서",
			`소스: M2 — 베가베리 내부 대응 문서
경로: c:\Users\BASEMENT_ADMIN\aia\archive\미팅메모\AIA_회신_대응정리.md
기준일: 2026-02-26
파싱:
- HNW VIP 여정 3단계: 러닝살롱→살롱런→JTBC 마라톤
- 캠프 A/B/C: 하이원 4,330만 / 파크로쉬 5,010만 / 알펜시아 4,680만
- 안은태: 6회+α 전체 참여 의사, 가민 콘텐츠 중
- 박지혜: AIA측 범위 미정, 계약 베가베리 일괄
- 황영조: 별도 500만`,
		},
		{
			"cortex/aia/소스/V1_초안제안서",
			`소스: V1 — 초안 제안서
경로: c:\Users\BASEMENT_ADMIN\aia\archive\aia_proposal_draft.md
작성일: 2026-01-21 (v1.0 내부검토)
파싱:
- 14명 (MP7+고객7), 6~11월 6개월
- 전용 웹앱: GPS 출석, 러닝일지, 체중기록
- 엘리트 코치 1:1 매칭 (격주 영상통화)
- 캠프: 제주 or 고산지대
- KPI: 완주율 95%, 만족도 4.7/5.0`,
		},
		{
			"cortex/aia/소스/A1_AIA내부PDF",
			`소스: A1 — AIA 내부 콘텐츠플랜 PDF
경로: c:\Users\BASEMENT_ADMIN\aia\archive\pdf_text.txt
파싱:
- "2026 Marathon Contents Plan" 영문 7페이지
- ⚠️ 크루 16명 (MP8+고객8) — 베가베리 14명과 불일치
- KPI: NPS +30 / Views +1M / MP Sharing 5,000
- 콘텐츠 8편 릴리즈 플랜
- 션(Sean) = AIA Health & Wellness Ambassador
- Running Camp 1N2D (베가베리 2박3일과 차이)`,
		},
		{
			"cortex/aia/소스/P1_전달가이드",
			`소스: P1 — 전달가이드 (3/18)
경로: c:\Users\BASEMENT_ADMIN\aia\reply_slides\0318_AIA전달\00_전달가이드_카톡_메일.md
파싱:
- 4건 전달 매뉴얼 (견적v11+장동선+ABC안+운영구조)
- 크루장: 안은태·박지혜 6회 인당 500만 합계 1,000만
- 총액: 133,200,000원 (VAT별도)`,
		},
		{
			"cortex/aia/소스/E1_견적서비교",
			`소스: E1 — 견적서 비교
경로: c:\Users\BASEMENT_ADMIN\aia\reply_slides\_archive\compare_output.txt
파싱: v1→v11 견적서 버전 비교`,
		},

		// ─── 타임라인 ───
		{
			"cortex/aia/타임라인/01월/0120_첫미팅",
			`날짜: 2026-01-20 (화)
유형: 대면 미팅 (95분)
소스: R1
장소: AIA 타워 21층
AIA: 이서윤, 김태영, 심승부 외 2
VEGA: 김태경, 박정근
합의: 14명 크루, 월1회 트레이닝, 캠프, 1/25 트랙데이 방문, 월정액 모델`,
		},
		{
			"cortex/aia/타임라인/01월/0121_초안제안서",
			`날짜: 2026-01-21 (수)
유형: 문서 작성
소스: V1
내용: aia_proposal_draft.md v1.0 내부검토용 작성
핵심: 14명, 웹앱, 1:1코치, 캠프 제주`,
		},
		{
			"cortex/aia/타임라인/01월/0124_후속통화",
			`날짜: 2026-01-24 (토)
유형: 통화 (33분)
소스: R2
참석: 이서윤, 류현정(첫등장), 김태경, 박정근
합의: 캠프 2박3일 확정, CMO 동행, 검수 경고`,
		},
		{
			"cortex/aia/타임라인/02월/0211_단톡방개설",
			`날짜: 2026-02-11 (수)
유형: 카톡
소스: K1
내용: 부장님방 개설, 김태영 합류, SNS제안서+KPI 요청, AIA_KPI_운영계획서.docx 전달`,
		},
		{
			"cortex/aia/타임라인/02월/0219_2차미팅",
			`날짜: 2026-02-19 (목)
유형: 대면 미팅
소스: K1
장소: AIA 타워
후속: AIA 6대 요청 피드백 메일 (M1)`,
		},
		{
			"cortex/aia/타임라인/02월/0226_견적v1",
			`날짜: 2026-02-26 (목)
유형: 문서 전달
소스: K1, M2
내용: 견적서 v1 PDF+XLSX, 내부 대응 문서 작성
부산: 신라대 or 아시아드`,
		},
		{
			"cortex/aia/타임라인/03월/0312_3차미팅",
			`날짜: 2026-03-12 (목)
유형: 대면 미팅 (65분)
소스: K1, R3
장소: AIA 타워
합의: 인원 31→20명, 정승은→장동선, +α, 보험 필수, AIA Only Experience`,
		},
		{
			"cortex/aia/타임라인/03월/0318_4건전달",
			`날짜: 2026-03-18 (수)
유형: 정식 자료 전달 (4건)
소스: K1, P1
문서: 견적v11(133,200,000원) + 장동선제안 + ABC안 + 운영구조
크루장: 안은태·박지혜 6회 인당 500만 = 1,000만`,
		},
		{
			"cortex/aia/타임라인/03월/0324_업체등록",
			`날짜: 2026-03-24 (화)
유형: 계약/서류
소스: K1, K2
내용: 업체등록서류+비밀유지, 박지혜 연락처, 장동선 500만(하루), UGC 300건 확정`,
		},
		{
			"cortex/aia/타임라인/03월/0326_일정확정",
			`날짜: 2026-03-26 (목)
유형: 일정 확정 (3차 수정)
소스: K1, K2
변경: 반포불가→목동, 9/26추석불가→9/5확정
확정: 6/27서울부산, 8/29서울, 9/5트랙데이, 10/31서울, 11/14부산, 6/19-21캠프`,
		},
		{
			"cortex/aia/타임라인/03월/0327_계약서_캠프",
			`날짜: 2026-03-27 (금)
유형: 계약서+캠프기획서
소스: K1, K2
내용: 계약서 이메일, aia_camp_master.html+PPTX, 장동선 확정, "예산은 저에게만"`,
		},
		{
			"cortex/aia/타임라인/04월/0401_최종승인",
			`날짜: 2026-04-01 (수)
유형: ★ 공식 승인
소스: K1
내용: 이서윤 "최종 승인 메일 전달" (GO), 류현정 10/17 황영조 요청`,
		},
		{
			"cortex/aia/타임라인/04월/0406_전환점",
			`날짜: 2026-04-06 (월)
유형: 핵심 유선통화
소스: K2
합의: Track Day 용어(AIA300 vs 베가베리기존), 장소요건(서울프라이빗300명), "스트럭처가 핵심"`,
		},
		{
			"cortex/aia/타임라인/04월/0409_미팅확정",
			`날짜: 2026-04-09 (목)
유형: 미팅 확정 + 대관 난항
소스: K1, K2
내용: 4/16 09:30 Teams, 대관 6곳 접촉, Track Day 9/12 확정`,
		},
		{
			"cortex/aia/타임라인/04월/0410_아젠다",
			`날짜: 2026-04-10 (금)
유형: 미팅 아젠다
소스: K1
8개 아젠다: 트레이닝확정안/SNS/캘린더/콘텐츠KPI/UGC카테고리/오프라인/OT/캠프`,
		},

		// ─── 핵심 데이터 ───
		{
			"cortex/aia/데이터/견적서",
			`견적서 변천:
v1 (2/26): 31명, 정승은, 제주 기준
v11 (3/18): 133,200,000원 (VAT별도), 장동선 1,000만, ~20명, +α

크루장 계약:
- 안은태: 6회(+α별도), 5,000,000원
- 박지혜: 6회(+α별도), 5,000,000원
- 합계: 10,000,000원, 계약서 별도`,
		},
		{
			"cortex/aia/데이터/인물",
			`AIA:
- 심승부 차장: 초기연결 (러너블 소개), 1/20
- 김태영 팀장: 의사결정 (부문장 보고선), 1/20→2/11
- 이서윤 과장: 실무총괄, 1/20→2/11
- 류현정(릴리): 영업MP+마케팅, 1/24
- 송인재: SM/인스타, 3/25

VEGAVERY:
- 김태경(Ted) 대표: 비즈니스, 1/20
- 박정근 PD: 실무/콘텐츠, 1/20

외부:
- 안은태 크루장: 러닝크리에이터, 6회 500만
- 박지혜 크루장: 아나운서(이영표 부인), 6회 500만
- 장동선 교수: 뇌과학·웰니스, 1일 500만
- 황영조 감독: 마라톤 금메달, 별도 500만
- 정승은 원장: 웰니스 → 장동선 대체
- 황싸부 코치: 러닝코치/YG전속`,
		},
		{
			"cortex/aia/데이터/확정일정",
			`확정 일정:
6/19-21: 강원 하이원 2박3일 웰니스 캠프
6/27(토): 서울 목동 08-11 + 부산 신라대 08-11 트레이닝
8/29(토): 서울 목동 08-11 트레이닝
9/12(토): 서울 미정 TBD AIA Track Day 300명 ← 장소 미확정
10/31(토): 서울 목동 09-12 트레이닝
11/14(토): 부산 아시아드 09-12 트레이닝`,
		},
		{
			"cortex/aia/데이터/미해결이슈",
			`미해결 이슈 (4/14 기준):
1. Track Day 장소 (9/12) — 육사/반포/한체대/한양대 접촉중
2. 장동선 최종 확정 — AIA 회신 미확인
3. 캠프 C안 승인 — AIA 회신 미확인
4. 크루장 계약서 양식 — AIA 표준 양식 대기
5. 하이원 예약 — 미확인
6. 업체등록 완료 — 3/27 서류 제출 후 대기
7. 10/17 황영조 장소 — 류현정 확인 중
8. 4/16 PPTX — 제작 중
9. ⚠️ 크루 인원 불일치 — AIA내부 16명 vs 베가베리 14명`,
		},
	}

	// ─── 컴파일 ───
	outDir := filepath.Join(".", "aia_cartridge_output")
	os.RemoveAll(outDir)

	zipName := "aia_runtogether_2026.jloot"
	zipFile, err := os.Create(zipName)
	if err != nil {
		fmt.Println("Error creating zip:", err)
		return
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	fmt.Println("\n[1단계] 뉴런 구조 컴파일 중...")

	for _, e := range entries {
		fullPath := filepath.Join(outDir, e.Path)
		if err := os.MkdirAll(fullPath, 0750); err != nil {
			fmt.Printf("Dir error: %v\n", err)
			continue
		}

		contentPath := filepath.Join(fullPath, "본문.neuron")
		if err := os.WriteFile(contentPath, []byte(e.Content), 0600); err != nil {
			fmt.Printf("File error: %v\n", err)
			continue
		}

		internalZipPath := filepath.ToSlash(filepath.Join(e.Path, "본문.neuron"))
		w, _ := archive.Create(internalZipPath)
		w.Write([]byte(e.Content))

		fmt.Printf("✅ %s\n", e.Path)
	}

	fmt.Printf("\n[완료] 총 %d개 뉴런이 컴파일되었습니다.\n", len(entries))
	fmt.Printf("📦 카트리지: %s\n", zipName)
	fmt.Printf("📂 물리 확인: %s\n", outDir)
	fmt.Println("\n사용법: cp", zipName, "tools/jloot/ (VFS 자동 마운트)")
}
