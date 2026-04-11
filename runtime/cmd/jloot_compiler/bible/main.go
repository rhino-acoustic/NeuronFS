package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// BibleVerse represents a single verse (= 1 neuron)
type BibleVerse struct {
	Testament string // old_testament / new_testament
	Book      string
	Chapter   string
	Verse     string
	Text      string
	IsBan     bool // 禁 = 십계명
	IsMust    bool // 必 = 지상명령
}

func main() {
	fmt.Println("[JLOOT BUILDER] ✝ 성경 카트리지 컴파일을 시작합니다...")

	verses := []BibleVerse{
		// ── 구약 ──
		// 창세기
		{Testament: "old_testament", Book: "창세기", Chapter: "1장", Verse: "1절_태초에_하나님이_천지를_창조하시니라",
			Text: "태초에 하나님이 천지를 창조하시니라"},
		{Testament: "old_testament", Book: "창세기", Chapter: "1장", Verse: "3절_빛이_있으라",
			Text: "하나님이 이르시되 빛이 있으라 하시니 빛이 있었고"},
		{Testament: "old_testament", Book: "창세기", Chapter: "1장", Verse: "27절_하나님의_형상대로",
			Text: "하나님이 자기 형상 곧 하나님의 형상대로 사람을 창조하시되 남자와 여자를 창조하시고"},

		// 출애굽기 20장 — 십계명 (禁)
		{Testament: "old_testament", Book: "출애굽기", Chapter: "20장", Verse: "다른신을_두지말라",
			Text: "너는 나 외에는 다른 신들을 네게 두지 말라", IsBan: true},
		{Testament: "old_testament", Book: "출애굽기", Chapter: "20장", Verse: "우상을_만들지말라",
			Text: "너를 위하여 새긴 우상을 만들지 말고", IsBan: true},
		{Testament: "old_testament", Book: "출애굽기", Chapter: "20장", Verse: "살인하지말라",
			Text: "살인하지 말라", IsBan: true},
		{Testament: "old_testament", Book: "출애굽기", Chapter: "20장", Verse: "간음하지말라",
			Text: "간음하지 말라", IsBan: true},
		{Testament: "old_testament", Book: "출애굽기", Chapter: "20장", Verse: "도둑질하지말라",
			Text: "도둑질하지 말라", IsBan: true},
		{Testament: "old_testament", Book: "출애굽기", Chapter: "20장", Verse: "거짓증거하지말라",
			Text: "네 이웃에 대하여 거짓 증거하지 말라", IsBan: true},
		{Testament: "old_testament", Book: "출애굽기", Chapter: "20장", Verse: "탐내지말라",
			Text: "네 이웃의 집을 탐내지 말라", IsBan: true},

		// 시편
		{Testament: "old_testament", Book: "시편", Chapter: "23장", Verse: "1절_여호와는_나의_목자",
			Text: "여호와는 나의 목자시니 내게 부족함이 없으리로다"},
		{Testament: "old_testament", Book: "시편", Chapter: "23장", Verse: "4절_사망의_골짜기",
			Text: "내가 사망의 음침한 골짜기로 다닐지라도 해를 두려워하지 않을 것은 주께서 나와 함께 하심이라"},
		{Testament: "old_testament", Book: "시편", Chapter: "119장", Verse: "105절_말씀은_등불",
			Text: "주의 말씀은 내 발에 등이요 내 길에 빛이니이다"},

		// 잠언
		{Testament: "old_testament", Book: "잠언", Chapter: "3장", Verse: "5절_마음을다하여",
			Text: "너는 마음을 다하여 여호와를 신뢰하고 네 명철을 의지하지 말라"},

		// ── 신약 ──
		// 마태복음
		{Testament: "new_testament", Book: "마태복음", Chapter: "5장", Verse: "3절_심령이_가난한자",
			Text: "심령이 가난한 자는 복이 있나니 천국이 그들의 것임이요"},
		{Testament: "new_testament", Book: "마태복음", Chapter: "5장", Verse: "14절_세상의_빛",
			Text: "너희는 세상의 빛이라 산 위에 있는 동네가 숨겨지지 못할 것이요"},
		{Testament: "new_testament", Book: "마태복음", Chapter: "6장", Verse: "9절_하늘에_계신_아버지",
			Text: "하늘에 계신 우리 아버지여 이름이 거룩히 여김을 받으시오며"},
		{Testament: "new_testament", Book: "마태복음", Chapter: "28장", Verse: "모든민족을_제자로삼아",
			Text: "그러므로 너희는 가서 모든 민족을 제자로 삼아 아버지와 아들과 성령의 이름으로 세례를 베풀고", IsMust: true},

		// 요한복음
		{Testament: "new_testament", Book: "요한복음", Chapter: "1장", Verse: "1절_태초에_말씀이",
			Text: "태초에 말씀이 계시니라 이 말씀이 하나님과 함께 계셨으니 이 말씀은 곧 하나님이시니라"},
		{Testament: "new_testament", Book: "요한복음", Chapter: "3장", Verse: "16절_하나님이_세상을_사랑하사",
			Text: "하나님이 세상을 이처럼 사랑하사 독생자를 주셨으니 이는 그를 믿는 자마다 멸망하지 않고 영생을 얻게 하려 하심이라"},
		{Testament: "new_testament", Book: "요한복음", Chapter: "14장", Verse: "6절_길이요_진리요_생명",
			Text: "내가 곧 길이요 진리요 생명이니 나로 말미암지 않고는 아버지께로 올 자가 없느니라"},

		// 로마서
		{Testament: "new_testament", Book: "로마서", Chapter: "8장", Verse: "28절_합력하여_선을",
			Text: "하나님을 사랑하는 자 곧 그의 뜻대로 부르심을 입은 자들에게는 모든 것이 합력하여 선을 이루느니라"},

		// 빌립보서
		{Testament: "new_testament", Book: "빌립보서", Chapter: "4장", Verse: "13절_능력주시는자",
			Text: "내게 능력 주시는 자 안에서 내가 모든 것을 할 수 있느니라"},

		// 요한계시록
		{Testament: "new_testament", Book: "요한계시록", Chapter: "22장", Verse: "13절_알파와오메가",
			Text: "나는 알파와 오메가요 처음과 마지막이요 시작과 마침이라"},
	}

	outDir := filepath.Join(".", "cartridge_bible_output", "cortex", "bible")
	os.RemoveAll(filepath.Join(".", "cartridge_bible_output"))

	zipName := "bible.jloot"
	zipFile, err := os.Create(zipName)
	if err != nil {
		fmt.Println("Error creating zip:", err)
		return
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	fmt.Println("\n[1단계] 물리 경로 (폴더) 생성 중...")

	banCount := 0
	mustCount := 0

	for _, v := range verses {
		// 禁 경로 삽입
		var fullPath string
		if v.IsBan {
			fullPath = filepath.Join(outDir, v.Testament, v.Book, v.Chapter, "禁", v.Verse)
			banCount++
		} else if v.IsMust {
			fullPath = filepath.Join(outDir, v.Testament, v.Book, v.Chapter, "必", v.Verse)
			mustCount++
		} else {
			fullPath = filepath.Join(outDir, v.Testament, v.Book, v.Chapter, v.Verse)
		}

		if err := os.MkdirAll(fullPath, 0750); err != nil {
			fmt.Printf("Dir create error: %v\n", err)
			continue
		}

		// 카운터: 禁=99, 必=20, 일반=랜덤
		counter := 1
		if v.IsBan {
			counter = 99
		} else if v.IsMust {
			counter = 20
		}

		neuronFile := filepath.Join(fullPath, fmt.Sprintf("%d.neuron", counter))
		if err := os.WriteFile(neuronFile, []byte(v.Text), 0600); err != nil {
			fmt.Printf("File write error: %v\n", err)
			continue
		}

		// Zip에도 기록
		zipPath := filepath.ToSlash(filepath.Join("cortex", "bible", v.Testament, v.Book, v.Chapter))
		if v.IsBan {
			zipPath += "/禁"
		} else if v.IsMust {
			zipPath += "/必"
		}
		zipPath += "/" + v.Verse + "/" + fmt.Sprintf("%d.neuron", counter)
		w, _ := archive.Create(zipPath)
		w.Write([]byte(v.Text))

		fmt.Printf("✅ %s/%s/%s → %s\n", v.Book, v.Chapter, v.Verse, neuronFile)
	}

	fmt.Printf("\n[완료] 총 %d절 컴파일 (禁 %d / 必 %d / 일반 %d)\n",
		len(verses), banCount, mustCount, len(verses)-banCount-mustCount)
	fmt.Printf("1. 물리 확인 폴더: %s\n", filepath.Join("cartridge_bible_output"))
	fmt.Printf("2. 패키징 본체: %s\n", zipName)
	fmt.Printf("\n[장착 방법]\n")
	fmt.Printf("  1) 폴더 직접 복사: cp -r cartridge_bible_output/cortex/bible brain_v4/cortex/bible\n")
	fmt.Printf("  2) .jloot 마운트: cp %s brain_v4/bible.jloot (VFS 자동 마운트)\n", zipName)
}
