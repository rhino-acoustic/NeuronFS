package main
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)
func main() {
	filepath.Walk("C:\\Users\\BASEMENT_ADMIN\\NeuronFS", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() { return nil }
		if strings.Contains(path, "scratch") || strings.Contains(path, ".git") || strings.Contains(path, "node_modules") { return nil }
		data, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(data), "마스터 프롬프트 v2") {
			fmt.Println("MATCH IN FILE:", path)
		}
		return nil
	})
}
