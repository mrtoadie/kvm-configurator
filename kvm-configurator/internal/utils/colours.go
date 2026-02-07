// utils/colours.go
// last modification: Feb 07 2026
package utils

import (
  "fmt"
  "os"
)

const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorYellow = "\033[33m"
    ColorBlue   = "\033[34m"
    ColorMagenta = "\033[35m"
    ColorCyan   = "\033[36m"
    ColorWhite  = "\033[37m"
    ColorBold   = "\033[1m"
)

// Colourise wraps a plain string in the given colour code
func Colourise(text, colour string) string {
    return colour + text + ColorReset
}

// Bold wraps a string in the “bold” escape sequence
func MakeBold(text string) string {
    return ColorBold + text + ColorReset
}

// combines colour & bold
func ColouriseBold(text, colour string) string {
    return colour + ColorBold + text + ColorReset
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
	SimpleError(prefix, ctx, err, ColorRed)
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
	fmt.Fprintln(os.Stdout, Colourise(msg, ColorGreen))
}

// Successf – lije fmt.Sprintf but colourised
func Successf(format string, a ...interface{}) {
	fmt.Fprintln(os.Stdout, Colourise(fmt.Sprintf(format, a...), ColorGreen))
}

// Info – neutral colour
func Info(prefix, ctx string) {
	fmt.Fprintln(os.Stdout, Colourise(fmt.Sprintf("%s – %s", prefix, ctx), ColorBlue))
}
// EOF