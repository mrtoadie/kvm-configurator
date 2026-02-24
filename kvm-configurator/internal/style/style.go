package style

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"
)

const (
	// borders
	borderBlue  = "\033[34m"
	borderReset = "\033[0m"
	BoxStdWidth = 60
	// colours
	ColReset   = "\033[0m"
	ColRed     = "\x1b[31m"
	ColGreen   = "\x1b[32m"
	ColYellow  = "\x1b[33m"
	ColBlue    = "\x1b[34m"
	ColMagenta = "\x1b[35m"
	ColCyan    = "\x1b[36m"
	ColWhite   = "\x1b[37m"
	ColBold    = "\x1b[1m"

	// Kombinierte Stile für häufige UI‑Elemente
	ColError   = ColRed
	ColSuccess = ColGreen
	ColInfo    = ColBlue
	ColPrompt  = ColYellow
	ColHeader  = ColCyan
	ColBorder  = ColBlue // bereits in style.go als borderBlue definiert – jetzt alias
	// DefaultTabOpts mirrors the historic parameters (0,10,2,' ')
	DefaultMinWidth = 0
	DefaultTabWidth = 10
	DefaultPadding  = 2
	DefaultPadChar  = ' '
)

// BORDERS / BOXES

// Box prints a normal left‑aligned box.
func Box(width int, lines []string) string {
	return drawBox(width, lines, false)
}

// BoxCenter prints a box where each line is horizontally centred.
func BoxCenter(width int, lines []string) string {
	return drawBox(width, lines, true)
}

// drawBox prints a box with optional centering.
// width ≤0 → uses BoxStdWidth.
func drawBox(width int, lines []string, center bool) string {
	if width <= 0 {
		width = BoxStdWidth
	}

	// wrap / truncate
	var wrapped []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		wrapped = append(wrapped, truncateOrWrap(l, width)...)
	}

	// longest line (Unicode‑aware)
	maxLen := width
	for _, l := range wrapped {
		if utf8.RuneCountInString(l) > maxLen {
			maxLen = utf8.RuneCountInString(l)
		}
	}

	// top / bottom border
	top := borderBlue + "╭" + borderReset + strings.Repeat(borderBlue+"─"+borderReset, maxLen+2) + borderBlue + "╮" + borderReset
	bottom := borderBlue + "╰" + borderReset + strings.Repeat(borderBlue+"─"+borderReset, maxLen+2) + borderBlue + "╯" + borderReset

	var b strings.Builder
	b.WriteString(top + "\n")
	for _, l := range wrapped {
		if center {
			// centre the line (Unicode‑aware length)
			padding := maxLen - utf8.RuneCountInString(l)
			left := padding / 2
			right := padding - left
			centered := strings.Repeat(" ", left) + l + strings.Repeat(" ", right)
			b.WriteString(fmt.Sprintf("%s│ %s │%s\n", borderBlue, centered, borderReset))
		} else {
			b.WriteString(fmt.Sprintf("%s│%s %-*s %s│%s\n", borderBlue, borderReset, maxLen, l, borderBlue, borderReset))
		}
	}
	b.WriteString(bottom)
	return b.String()
}

// truncateOrWrap cuts a string to `max` runes, returning one or more parts.
// if the original fits, it is returned unchanged.
func truncateOrWrap(s string, max int) []string {
	if max <= 0 {
		return []string{s}
	}
	if utf8.RuneCountInString(s) <= max {
		return []string{s}
	}

	var parts []string
	runes := []rune(s)

	for len(runes) > max {
		parts = append(parts, string(runes[:max]))
		runes = runes[max:]
	}

	if len(parts) > 0 && len(runes) > 0 {
		parts = append(parts, string(runes)+"…")
	} else if len(runes) > 0 {
		if max > 1 {
			parts = []string{string(runes[:max-1]) + "…"}
		} else {
			parts = []string{"…"}
		}
	}
	return parts
}

// COLOURS
// Colourise wraps a plain string in the given colour code
func Colourise(text, colour string) string {
	return colour + text + ColReset
}

// combines colour & bold
func ColouriseBold(text, colour string) string {
	return colour + ColBold + text + ColReset
}

// SimpleError
func SimpleError(prefix, ctx string, err error) {
	if err == nil {
		return
	}
	// Example: “❗️Config missing – while loading: <original error>”
	msg := fmt.Sprintf("❗️%s >%s %v", prefix, ctx, err)
	fmt.Fprintln(os.Stderr, Colourise(msg, ColRed))
}

// Convenience wrapper for the usual red error
func RedError(prefix, ctx string, err error) {
	if err == nil {
		msg := fmt.Sprintf("❗️%s >%s", prefix, ctx)
		fmt.Fprintln(os.Stderr, Colourise(msg, ColRed))
		return
	}
	SimpleError(prefix, ctx, err)
}

/*
Success – displays a green success message.
prefix = short title (e.g., “✅ VM created”)
ctx = additional information (e.g., “my-vm-01”)
extra = optional additional text (can be empty)
*/
func Success(prefix, ctx, extra string) {
	msg := fmt.Sprintf("%s – %s", prefix, ctx)
	if extra != "" {
		msg = fmt.Sprintf("%s – %s", msg, extra)
	}
	fmt.Fprintln(os.Stdout, Colourise(msg, ColGreen))
}

// Successf – fmt.Sprintf but colourised
func Successf(format string, a ...interface{}) {
	fmt.Fprintln(os.Stdout, Colourise(fmt.Sprintf(format, a...), ColGreen))
}

// Info – neutral colour
func Info(prefix, ctx string) {
	fmt.Fprintln(os.Stdout, Colourise(fmt.Sprintf("%s – %s", prefix, ctx), ColBlue))
}

// Fehler‑Message (rot)
func Err(text string) string { return Colourise(text, ColError) }

// Erfolg‑Message (grün)
func Ok(text string) string { return Colourise(text, ColSuccess) }

// Hinweis‑Message (blau)
func Hint(text string) string { return Colourise(text, ColInfo) }

// Prompt‑Message (gelb)
func PromptMsg(text string) string { return Colourise(text, ColPrompt) }

// Header‑Zeile (cyan)
func Header(text string) string { return Colourise(text, ColHeader) }

// SPINNER / PROGRESS
// Progress encapsulates a Spinner go-routine
type Progress struct {
	stop chan struct{}
}

// NewProgress starts a spinner with the message passed
func SpinnerProgress(msg string) *Progress {
	p := &Progress{stop: make(chan struct{})}
	go func() {
		chars := []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}
		i := 0
		for {
			select {
			case <-p.stop:
				fmt.Print("\r")
				//fmt.Printf("%s ... done!\n", msg)
				return
			default:
				fmt.Printf("\r%s %c ", msg, chars[i%len(chars)])
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
	return p
}

// stops the spinner and releases the goroutine
func (p *Progress) Stop() {
	close(p.stop)
}

// STATUS
/* --------------------
	NormalizeStatus converts German/English VM‑states to the internal canonical forms
-------------------- */
func NormalizeStatus(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "running", "laufend":
		return "running"
	case "shut", "off", "shutoff", "ausgeschaltet":
		return "shut off"
	default:
		// unknown – keep as‑is so we never treat it as “running”
		return s
	}
}

// TABWRITER / TABLES
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
