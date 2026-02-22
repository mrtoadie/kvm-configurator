// ui/input.go
// last modified: Feb 22 2026
package utils

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Prompt reads a line from r after printing prompt to w.
// Returns the trimmed line and any I/O error (including io.EOF).
func Prompt(r *bufio.Reader, w io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(w, prompt); err != nil {
		return "", err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err // propagate EOF / other errors
	}
	return strings.TrimSpace(line), nil
}