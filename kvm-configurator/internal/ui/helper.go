// ui/helpers.go
// last modified: Feb 24 2026
package ui

import (
	"bufio"
	"fmt"
	"strings"
	"os"

	"configurator/internal/model"
	"configurator/internal/style"
	"configurator/internal/utils"
)

// normalize trims whitespace and forces lower‑case.
func normalize(s string) string { return strings.TrimSpace(strings.ToLower(s)) }

// editEnum prompts for a value that must belong to an allow‑list.
// fieldPtr points to the struct field that shall be updated.
func editEnum(r *bufio.Reader, cfg *model.DomainConfig,
	fieldPtr *string, prompt, label string, allowed map[string]bool) {

	raw, err := utils.Prompt(r, os.Stdout, style.Hint(prompt))
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Err("Read error:"), err)
		return
	}
	if raw == "" { // user pressed <Enter> → keep current value
		return
	}

	val := normalize(raw)
	if !allowed[val] {
		// build a nice “allowed: a, b, c” string
		keys := make([]string, 0, len(allowed))
		for k := range allowed {
			keys = append(keys, k)
		}
		fmt.Fprintln(os.Stderr,
			style.Err(fmt.Sprintf("Invalid %s – allowed: %s (you entered: %s)",
				label, strings.Join(keys, ", "), val)))
		return
	}

	*fieldPtr = val
	style.Success(label, val, "")
}