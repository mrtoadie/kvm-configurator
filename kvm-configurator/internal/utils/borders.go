// utils/borders.go
// last modification: Feb 06 2026
package utils

import (
	"fmt"
	"strings"
)

func Box(width int, lines []string) string {
	// kills the last line
	var cleanLines []string
  for _, l := range lines {
    if strings.TrimSpace(l) != "" {
      cleanLines = append(cleanLines, l)
    }
  }
  lines = cleanLines
	// max length for the box
  maxLen := width
  for _, l := range lines {
    if len(l) > maxLen {
            maxLen = len(l)
    }
  }
	// top and bottom style
  top    := "\033[34m╭\033[0m" + strings.Repeat("\033[34m─\033[0m", maxLen+2) + "\033[34m╮\033[0m"
  bottom := "\033[34m╰\033[0m" + strings.Repeat("\033[34m─\033[0m", maxLen+2) + "\033[34m╯\033[0m"

  var b strings.Builder
  b.WriteString(top + "\n")
  for _, l := range lines {
    b.WriteString(fmt.Sprintf("\033[34m│\033[0m %-*s \033[34m│\033[0m\n", maxLen, l))
  }
  b.WriteString(bottom)
  return b.String()
}

func BoxCenter(width int, lines []string) string {
    maxLen := width
    for _, l := range lines {
        if len(l) > maxLen {
            maxLen = len(l)
        }
    }

    top    := "\033[34m╭\033[0m" + strings.Repeat("\033[34m─\033[0m", maxLen+2) + "\033[34m╮\033[0m"
    bottom := "\033[34m╰\033[0m" + strings.Repeat("\033[34m─\033[0m", maxLen+2) + "\033[34m╯\033[0m"

    var b strings.Builder
    b.WriteString(top + "\n")
    
    for _, l := range lines {
        // Zentrierung: Leerzeichen links + Text + Leerzeichen rechts
        padding := (maxLen - len(l)) / 2
        centered := strings.Repeat(" ", padding) + l + strings.Repeat(" ", maxLen-len(l)-padding)
        b.WriteString(fmt.Sprintf("\033[34m│ %s │\033[0m\n", centered))
    }
    
    b.WriteString(bottom)
    return b.String()
}