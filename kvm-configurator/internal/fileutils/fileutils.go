// fileutils/fileutils.go
package fileutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/* --------------------
	ListFiles
-------------------- */
func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.Type().IsRegular() {
			abs, err := filepath.Abs(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, err
			}
			files = append(files, abs)
		}
	}
	return files, nil
}

/* --------------------
	PromptSelection
-------------------- */
func PromptSelection(files []string) (int, error) {
	if len(files) == 0 {
		return -1, fmt.Errorf("\x1b[31mno files found\x1b[0m")
	}
	fmt.Println("\n\x1b[34m=== Select ISO ===\x1b[0m")
	for i, f := range files {
		fmt.Printf("[%d] %s\n", i+1, filepath.Base(f))
	}
	fmt.Print("\nPlease enter number (or 0 to cancel): ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return -1, err
	}
	line = strings.TrimSpace(line)

	choice, err := strconv.Atoi(line)
	if err != nil {
		return -1, fmt.Errorf("\x1b[31mInvalid selection!\x1b[0m")
	}
	if choice < 0 || choice > len(files) {
		return -1, fmt.Errorf("\x1b[31mPlease enter a valid number.\x1b[0m")
	}
	return choice, nil
}