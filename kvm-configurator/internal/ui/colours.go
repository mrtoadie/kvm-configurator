// ui/colours.go
// last modification: January 18 2026
package ui

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
// EOF