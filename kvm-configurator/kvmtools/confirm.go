// kvmtools/confirm.go
// last modification: Feb 10 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"strings"
)

// ask prompt yes/no
func AskYesNo(r *bufio.Reader, prompt string) (bool, error) {
	fmt.Print(prompt + " [y/N]: ")
	line, err := r.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return strings.HasPrefix(answer, "y"), nil
}