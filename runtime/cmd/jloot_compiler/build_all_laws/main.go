package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	fmt.Println("🏛️ [마스터 카트리지 빌더] 대한민국 법령 100% 딥-컴파일 개시...")
	start := time.Now()

	lawsDir := filepath.Join(".", "legalize-kr", "kr")
	outZip := "대한민국_통합법전_Master.jloot"

	zipFile, err := os.Create(outZip)
	if err != nil {
		fmt.Printf("❌ ZIP 생성 실패: %v\n", err)
		return
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filesProcessed := 0
	articlesProcessed := 0

	err = filepath.Walk(lawsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		rawBytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		lawName := strings.TrimSuffix(info.Name(), ".md")

		lines := strings.Split(string(rawBytes), "\n")
		var currentChapter, currentArticle string
		var content strings.Builder

		flush := func() {
			if currentArticle != "" {
				c := "제1장_기본"
				if currentChapter != "" {
					c = currentChapter
				}
				// Save path: cortex/대한민국_법령/[법이름]/[장]/[조]/본문.neuron
				zPath := filepath.ToSlash(filepath.Join("cortex", "대한민국_법령", lawName, c, currentArticle, "본문.neuron"))
				w, wErr := archive.Create(zPath)
				if wErr == nil {
					w.Write([]byte(strings.TrimSpace(content.String())))
					articlesProcessed++
				}
				content.Reset()
			}
		}

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "# ") {
				if currentArticle != "" {
					flush()
				}
				// # 근로기준법 -> 무시 (파일명으로 대체)
			} else if strings.HasPrefix(line, "## ") {
				if currentArticle != "" {
					flush()
				}
				currentChapter = strings.ReplaceAll(strings.TrimPrefix(line, "## "), " ", "_")
			} else if strings.HasPrefix(line, "### ") {
				if currentArticle != "" {
					flush()
				}
				art := strings.TrimPrefix(line, "### ")
				art = strings.ReplaceAll(art, "(", "_")
				art = strings.ReplaceAll(art, ")", "")
				currentArticle = strings.ReplaceAll(art, " ", "_")
			} else {
				content.WriteString(line + "\n")
			}
		}
		flush()
		filesProcessed++

		if filesProcessed%500 == 0 {
			fmt.Printf("⏳ 처리 중... (%d개 법령, %d개 조문)\n", filesProcessed, articlesProcessed)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ 스캔 오류: %v\n", err)
	}

	fmt.Printf("\n✅ [통합 완료] 총 %d개 법령, %d개 조항 압축 완료!\n", filesProcessed, articlesProcessed)
	fmt.Printf("⏱️ 소요 시간: %v\n", time.Since(start))

	vStat, _ := zipFile.Stat()
	fmt.Printf("📦 최종 카트리지 크기: %.2f MB (%s)\n", float64(vStat.Size())/1024/1024, outZip)
}
