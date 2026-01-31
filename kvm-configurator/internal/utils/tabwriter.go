// utils/tabwriter.go
// last modification: January 25 2026
package utils

import (
	"text/tabwriter"
	"os"
)

/* --------------------
	NewTabWriter returns a ready‑to‑use tabwriter.Writer for stdout
-------------------- */
func NewTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 10, 2, ' ', 0)
}
// EOF