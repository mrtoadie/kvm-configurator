// ui/colours.go
// last modification: January 17 2026
package ui

const (
    reset  = "\033[0m"
    Red    = "\033[31m"
    Green  = "\033[32m"
    Yellow = "\033[33m"
    Cyan   = "\033[36m"
	Blue   = "\033[34m"
	//blue   = "\x1b[34m" ?
)

func Colourise(text, colour string) string {
    return colour + text + reset
}
// EOF