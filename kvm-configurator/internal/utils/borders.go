// utils/display.go
// last modified: Feb 22 2026
package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	borderBlue  = "\033[34m"
	borderReset = "\033[0m"
	BoxStdWidth = 60
)

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
