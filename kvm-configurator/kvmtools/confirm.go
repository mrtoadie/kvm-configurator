// kvmtools/confirm.go
package kvmtools

import (
	"bufio"
	"fmt"
	"strings"
)

// AskYesNo fragt den Nutzer nach einer Bestätigung.
// Gibt true zurück, wenn die Antwort mit „y“ oder „yes“ beginnt (case‑insensitive).
func AskYesNo(r *bufio.Reader, prompt string) (bool, error) {
	fmt.Print(prompt + " [y/N]: ")
	line, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return strings.HasPrefix(answer, "y"), nil
}