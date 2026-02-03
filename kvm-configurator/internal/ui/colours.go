// ui/colours.go
// last modification: Feb 03 2026
package ui

import (
  "fmt"
  "os"
)

const (
    Reset  = "\033[0m"
    Red    = "\033[31m"
    Green  = "\033[32m"
    Yellow = "\033[33m"
    Blue   = "\033[34m"
    Magenta = "\033[35m"
    Cyan   = "\033[36m"
    White  = "\033[37m"
    Bold   = "\033[1m"
)

/* --------------------
  Colourise wraps a plain string in the given colour code
-------------------- */
func Colourise(text, colour string) string {
    return colour + text + Reset
}

/* --------------------
  Bold wraps a string in the “bold” escape sequence
-------------------- */
func MakeBold(text string) string {
    return Bold + text + Reset
}

/* --------------------
  combines colour & bold
-------------------- */
func ColouriseBold(text, colour string) string {
    return colour + Bold + text + Reset
}

// SimpleError
func SimpleError(prefix, ctx string, err error, colour string) {
	if err == nil {
		return
	}
	// Example: “❗️Config missing – while loading: <original error>”
	msg := fmt.Sprintf("❗️%s >%s %v", prefix, ctx, err)
	fmt.Fprintln(os.Stderr, Colourise(msg, colour))
}

// Convenience wrapper for the usual red error
func RedError(prefix, ctx string, err error) {
	SimpleError(prefix, ctx, err, Red)
}

// Success – displays a green success message.
// prefix = short title (e.g., “✅ VM created”)
// ctx = additional information (e.g., “my-vm-01”)
// extra = optional additional text (can be empty)
func Success(prefix, ctx, extra string) {
	msg := fmt.Sprintf("%s – %s", prefix, ctx)
	if extra != "" {
		msg = fmt.Sprintf("%s – %s", msg, extra)
	}
	fmt.Fprintln(os.Stdout, Colourise(msg, Green))
}

// Successf – lije fmt.Sprintf but colourised
func Successf(format string, a ...interface{}) {
	fmt.Fprintln(os.Stdout, Colourise(fmt.Sprintf(format, a...), Green))
}

// Info – neutral colour
func Info(prefix, ctx string) {
	fmt.Fprintln(os.Stdout, Colourise(fmt.Sprintf("%s – %s", prefix, ctx), Blue))
}
// EOF