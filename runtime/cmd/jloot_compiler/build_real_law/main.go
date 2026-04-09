package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🔥 [실테스트 가동] 대규모 법령 텍스트 Jloot 카트리지 변환 시작...")

	// 실제 근로기준법 텍스트 파싱을 견고하게 테스트하기 위한 실제 텍스트 덤프 (1장 총칙 전면)
	rawText := `
# 제1장 총칙
## 제1조(목적)
이 법은 헌법에 따라 근로조건의 기준을 정함으로써 근로자의 기본적 생활을 보장, 향상시키며 균형 있는 국민경제의 발전을 꾀하는 것을 목적으로 한다.
## 제2조(정의)
이 법에서 사용하는 용어의 뜻은 다음과 같다.
1. "근로자"란 직업의 종류와 관계없이 임금을 목적으로 사업이나 사업장에 근로를 제공하는 사람을 말한다.
2. "사용자"란 사업주 또는 사업 경영 담당자, 그 밖에 근로자에 관한 사항에 대하여 사업주를 위하여 행위하는 자를 말한다.
## 제3조(근로조건의 기준)
이 법에서 정하는 근로조건은 최저기준이므로 근로 관계 당사자는 이 기준을 이유로 근로조건을 낮출 수 없다.
## 제4조(근로조건의 결정)
근로조건은 근로자와 사용자가 동등한 지위에서 자유의사에 따라 결정하여야 한다.
## 제5조(균등한 처우)
사용자는 근로자에 대하여 남녀의 성(性)을 이유로 차별적 대우를 하지 못하며, 국적ㆍ신앙 또는 사회적 신분을 이유로 근로조건에 대한 차별적 처우를 하지 못한다.
## 제6조(강제 근로의 금지)
사용자는 폭행, 협박, 감금, 그 밖에 정신상 또는 신체상의 자유를 부당하게 구속하는 수단으로써 근로자의 자유의사에 어긋나는 근로를 강요하지 못한다.
## 제7조(폭행의 금지)
사용자는 사고의 발생이나 그 밖의 어떠한 이유로도 근로자에게 폭행을 하지 못한다.
## 제8조(중간착취의 배제)
누구든지 법률에 따르지 아니하고는 영리로 다른 사람의 취업에 개입하거나 중간인으로서 이익을 취득하지 못한다.
## 제9조(공민권 행사의 보장)
사용자는 근로자가 근로시간 중에 선거권, 그 밖의 공민권 행사 또는 공(公)의 직무를 집행하기 위하여 필요한 시간을 청구하면 거부하지 못한다.
`

	outZip := "근로기준법_Master.jloot"
	zipFile, err := os.Create(outZip)
	if err != nil {
		panic(err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	lines := strings.Split(rawText, "\n")
	var currentChapter string
	var currentArticle string
	var articleContent strings.Builder

	// Write existing content before sweeping
	flushArticle := func() {
		if currentArticle != "" && currentChapter != "" {
			internalZipPath := filepath.ToSlash(filepath.Join("cortex", "노무법", "근로기준법", currentChapter, currentArticle, "본문.neuron"))
			w, _ := archive.Create(internalZipPath)
			w.Write([]byte(strings.TrimSpace(articleContent.String())))
			articleContent.Reset()
			fmt.Printf("✔ 굽기 완료: %s/%s\n", currentChapter, currentArticle)
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "# ") {
			if currentArticle != "" {
				flushArticle()
			}
			currentChapter = strings.ReplaceAll(strings.TrimPrefix(line, "# "), " ", "_")
		} else if strings.HasPrefix(line, "## ") {
			if currentArticle != "" {
				flushArticle()
			}
			// 예: "제1조(목적)" -> "제1조_목적"
			rawArticle := strings.TrimPrefix(line, "## ")
			rawArticle = strings.ReplaceAll(rawArticle, "(", "_")
			rawArticle = strings.ReplaceAll(rawArticle, ")", "")
			currentArticle = strings.ReplaceAll(rawArticle, " ", "_")
		} else {
			articleContent.WriteString(line + "\n")
		}
	}
	flushArticle() // 마지막 남은 조항 처리

	fmt.Printf("\n✅ [실테스트 완료] 실제 근로기준법 텍스트를 파싱하여 %s 로 패키징 완료했습니다.\n", outZip)
	info, _ := os.Stat(outZip)
	fmt.Printf("📦 생성된 jloot 파일 용량: %d 바이트 (극도로 경량화됨)\n", info.Size())
}
