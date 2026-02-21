// utils/tabwriter.go
// last modification: Feb 21 2026
package utils

import (
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

const (
	// DefaultTabOpts mirrors the historic parameters (0,10,2,' ')
	DefaultMinWidth = 0
	DefaultTabWidth = 10
	DefaultPadding  = 2
	DefaultPadChar  = ' '
)

// newTabWriter is the single place where we configure a tabwriter.
// Keeping it private avoids accidental drift between callers.
func newTabWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, DefaultMinWidth, DefaultTabWidth,
		DefaultPadding, DefaultPadChar, 0)
}

// NewTabWriter is the public helper for interactive console output.
func NewTabWriter() *tabwriter.Writer { return newTabWriter(os.Stdout) }

// TableToLines renders a table‑writer callback into a slice of strings.
// It returns the lines *and* any flush‑error, so callers can decide what to do.
func TableToLines(tableFunc func(w *tabwriter.Writer)) ([]string, error) {
	var buf strings.Builder
	tw := newTabWriter(&buf)

	tableFunc(tw) // fill the writer
	if err := tw.Flush(); err != nil {
		return nil, err
	}
	return strings.Split(buf.String(), "\n"), nil
}

// MustTableToLines is a convenience wrapper that panics on error.
// Handy for UI code where a failure is truly unexpected.
func MustTableToLines(tableFunc func(w *tabwriter.Writer)) []string {
	lines, err := TableToLines(tableFunc)
	if err != nil {
		panic(err) // or log.Fatalf – whichever fits your error policy
	}
	return lines
}
// EOF