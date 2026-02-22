// fileutils/fileutils.go
// last modified: Feb 22 2026
package utils

import (
	"bufio"
	//"configurator/internal/ui"
	"fmt"
	"os"
	"io"
	"path/filepath"
	"strconv"
	//"strings"
	// internal
	"configurator/internal/style"
)

const (
	CancelChoice   = 0  // user typed 0 → abort
	InvalidChoice  = -1 // parsing error or out‑of‑range
)

// ListFiles returns the absolute paths of regular files inside dir.
// If the directory cannot be read, the error is propagated to the caller.
func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// propagate – let the UI decide what to do
		return nil, fmt.Errorf("cannot read directory %q: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if e.Type().IsRegular() {
			abs, err := filepath.Abs(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("resolve absolute path for %q: %w", e.Name(), err)
			}
			files = append(files, abs)
		}
	}
	return files, nil
}

// PromptSelection shows a numbered menu of the supplied files and returns the
// user’s choice (1‑based). 0 means “cancel”. Errors are returned explicitly.
func PromptSelection(r *bufio.Reader, w io.Writer, files []string) (int, error) {
	if len(files) == 0 {
		return InvalidChoice, fmt.Errorf("no files to select from")
	}

	// Header – use the shared colour constant
	fmt.Fprintln(w, style.Colourise("\n=== Select ISO ===", style.ColBlue))

	for i, f := range files {
		fmt.Fprintf(w, "[%d] %s\n", i+1, filepath.Base(f))
	}
	prompt := style.Colourise("Please enter number (or 0 to cancel): ", style.ColYellow)

	// Re‑use the universal Prompt helper
	line, err := Prompt(r, w, prompt)
	if err != nil {
		return InvalidChoice, err
	}
	choice, err := strconv.Atoi(line)
	if err != nil {
		return InvalidChoice, fmt.Errorf("invalid number %q", line)
	}
	if choice < CancelChoice || choice > len(files) {
		return InvalidChoice, fmt.Errorf("choice %d out of range (0-%d)", choice, len(files))
	}
	return choice, nil
}