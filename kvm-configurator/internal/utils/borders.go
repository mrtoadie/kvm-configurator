// utils/borders.go
// last modification: Feb 12 2026
package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	BorderColorReset = "\033[0m"
	BorderColorBlue  = "\033[34m"
	BoxStdWidth      = 60
)

func Box(width int, lines []string) string {
	if width <= 0 {
		width = BoxStdWidth
	}

	// wrap / truncate
	var wrapped []string
	for _, l := range lines {
		// skip blank lines
		if strings.TrimSpace(l) == "" {
			continue
		}
		wrapped = append(wrapped, truncateOrWrap(l, width)...)
	}
	lines = wrapped

	// determine maxLen (can be smaller than width)
	maxLen := width
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	// top and bottom style
	top := BorderColorBlue + "╭" + BorderColorReset + strings.Repeat(BorderColorBlue+"─"+BorderColorReset, maxLen+2) + BorderColorBlue + "╮" + BorderColorReset
	bottom := BorderColorBlue + "╰" + BorderColorReset + strings.Repeat(BorderColorBlue+"─"+BorderColorReset, maxLen+2) + BorderColorBlue + "╯" + BorderColorReset

	var b strings.Builder
	b.WriteString(top + "\n")
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("\033[34m│\033[0m %-*s \033[34m│\033[0m\n", maxLen, l))
	}
	b.WriteString(bottom)
	return b.String()
}

func BoxCenter(width int, lines []string) string {
	if width <= 0 {
		width = BoxStdWidth
	}

	// wrap / truncate
	var wrapped []string
	for _, l := range lines {
		// skip blank lines
		if strings.TrimSpace(l) == "" {
			continue
		}
		wrapped = append(wrapped, truncateOrWrap(l, width)...)
	}
	lines = wrapped

	// determine maxLen (can be smaller than width)
	maxLen := width
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}

	top := BorderColorBlue + "╭" + BorderColorReset +
		strings.Repeat(BorderColorBlue+"─"+BorderColorReset, maxLen+2) +
		BorderColorBlue + "╮" + BorderColorReset
	bottom := BorderColorBlue + "╰" + BorderColorReset +
		strings.Repeat(BorderColorBlue+"─"+BorderColorReset, maxLen+2) +
		BorderColorBlue + "╯" + BorderColorReset

	var b strings.Builder
	b.WriteString(top + "\n")
	for _, l := range lines {
		// centering – calculate left and right padding
		padding := (maxLen - len(l)) / 2
		rightPad := maxLen - len(l) - padding
		centered := strings.Repeat(" ", padding) + l + strings.Repeat(" ", rightPad)
		b.WriteString(fmt.Sprintf("\033[34m│ %s │\033[0m\n", centered))
	}
	b.WriteString(bottom)
	return b.String()
}

func truncateOrWrap(s string, max int) []string {
	if max <= 0 {
		return []string{s}
	}

	// quick check: is it okay?
	if utf8.RuneCountInString(s) <= max {
		return []string{s}
	}

	var parts []string
	runes := []rune(s)
	for len(runes) > max {
		// cut off exactly `max` runes
		part := string(runes[:max])
		parts = append(parts, part)
		runes = runes[max:]
	}
	// rest (can be shorter)
	if len(runes) > 0 {
		// if already had at least one section, mark the last one as cut off
		if len(parts) > 0 {
			parts = append(parts, string(runes)+"…")
		} else {
			// no previous section → simple shortening
			parts = []string{string(runes[:max-1]) + "…"}
		}
	}
	return parts
}

// EOF
