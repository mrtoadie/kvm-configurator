// utils/tabwriter.go
// last modification: January 25 2026
package utils

import (
	"text/tabwriter"
	"os"
	"strings"
)

/* --------------------
	NewTabWriter returns a ready‑to‑use tabwriter.Writer for stdout
-------------------- */
func NewTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 10, 2, ' ', 0)
}

func TableToLines(tableFunc func(w *tabwriter.Writer)) []string {
    var buf strings.Builder
    tw := tabwriter.NewWriter(&buf, 0, 10, 2, ' ', 0)
    tableFunc(tw)
    tw.Flush()
    return strings.Split(buf.String(), "\n")
}
// EOF