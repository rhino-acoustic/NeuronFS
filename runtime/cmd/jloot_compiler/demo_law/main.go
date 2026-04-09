package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// LawArticle represents a single law definition
type LawArticle struct {
	LawName string
	Chapter string
	Article string
	Content string
}

func main() {
	fmt.Println("[JLOOT BUILDER] 대한민국 법령 카트리지 컴파일을 시작합니다...")

	// 데모용: 국세기본법 및 근로기준법 핵심 조항 8개
	laws := []LawArticle{
		// 국세기본법
		{"세법/국세기본법", "제1장_총칙", "제1조_목적", "이 법은 국세에 관한 기본적이고 공통적인 사항과 위법 또는 부당한 국세처분에 대한 불복 절차를 규정함으로써 국세 수입을 원활하게 확보하고 납세자의 권리를 보호함을 목적으로 한다."},
		{"세법/국세기본법", "제1장_총칙", "제2조_정의", "국세란 국가가 부과하는 다음 각 목의 세금을 말한다. 1. 소득세 2. 법인세 3. 상속세 ..."},
		{"세법/국세기본법", "제2장_국세부과와_세법적용", "제14조_실질과세", "① 과세의 대상이 되는 소득, 수익, 재산, 행위 또는 거래의 귀속이 명의일 뿐이고 사실상 귀속되는 자가 따로 있을 때에는 사실상 귀속되는 자를 납세의무자로 하여 세법을 적용한다."},
		{"세법/국세기본법", "제2장_국세부과와_세법적용", "제15조_신의성실", "납세자가 그 의무를 이행할 때에는 신의에 따라 성실하게 하여야 한다. 세무공무원이 직무를 수행할 때에도 또한 같다."},
		
		// 근로기준법
		{"노무법/근로기준법", "제1장_총칙", "제1조_목적", "이 법은 헌법에 따라 근로조건의 기준을 정함으로써 근로자의 기본적 생활을 보장, 향상시키며 균형 있는 국민경제의 발전을 꾀하는 것을 목적으로 한다."},
		{"노무법/근로기준법", "제1장_총칙", "제2조_정의", "이 법에서 사용하는 용어의 뜻은 다음과 같다. 1. '근로자'란 직업의 종류와 관계없이 임금을 목적으로 사업이나 사업장에 근로를 제공하는 사람을 말한다."},
		{"노무법/근로기준법", "제1장_총칙", "제5조_균등한처우", "사용자는 근로자에 대하여 남녀의 성을 이유로 차별적 대우를 하지 못하며, 국적, 신앙 또는 사회적 신분을 이유로 근로조건에 대한 차별적 처우를 하지 못한다."},
		{"노무법/근로기준법", "제4장_근로시간과휴식", "제50조_근로시간", "① 1주 간의 근로시간은 휴게시간을 제외하고 40시간을 초과할 수 없다. ② 1일의 근로시간은 휴게시간을 제외하고 8시간을 초과할 수 없다."},
	}

	outDir := filepath.Join(".", "cartridge_demo_output", "cortex")
	os.RemoveAll(outDir)

	zipName := "korea_law_demo.jloot"
	zipFile, err := os.Create(zipName)
	if err != nil {
		fmt.Println("Error creating zip:", err)
		return
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	fmt.Println("\n[1단계] 물리 경로 (폴더) 생성 중...")

	for _, law := range laws {
		// 경로 설계: cortex / 세법/국세기본법 / 제1장_총칙 / 제1조_목적
		fullPath := filepath.Join(outDir, law.LawName, law.Chapter, law.Article)
		
		// 물리적 OS 경로 생성 
		if err := os.MkdirAll(fullPath, 0750); err != nil {
			fmt.Printf("Dir create error: %v\n", err)
			continue
		}

		// 내용 쓰기 (본문.neuron)
		contentPath := filepath.Join(fullPath, "본문.neuron")
		if err := os.WriteFile(contentPath, []byte(law.Content), 0600); err != nil {
			fmt.Printf("File write error: %v\n", err)
			continue
		}

		// Zip (에뮬레이트된 jloot) 파일에도 기록
		internalZipPath := filepath.ToSlash(filepath.Join("cortex", law.LawName, law.Chapter, law.Article, "본문.neuron"))
		w, _ := archive.Create(internalZipPath)
		w.Write([]byte(law.Content))

		fmt.Printf("✅ 폴더 렌더링 완료: %s\n", filepath.ToSlash(filepath.Join(law.LawName, law.Chapter, law.Article)))
	}

	fmt.Printf("\n[완료] 총 %d개의 조항이 OS B-Tree 구조로 컴파일되었습니다.\n", len(laws))
	fmt.Printf("1. 물리 확인 폴더: %s\n", filepath.Join("cartridge_demo_output"))
	fmt.Printf("2. 패키징 본체: %s (이 파일 하나만 VFS 엔진에 마운트하면 끝납니다!)\n", zipName)
}
