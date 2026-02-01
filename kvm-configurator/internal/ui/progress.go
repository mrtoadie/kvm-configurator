// ui/progress.go
package ui

import (
	"fmt"
	"time"
)

// Progress kapselt einen Spinner‑Goroutine.
type Progress struct {
	stop chan struct{}
}

// NewProgress startet einen Spinner mit der übergebenen Meldung.
func NewProgress(msg string) *Progress {
	p := &Progress{stop: make(chan struct{})}
	go func() {
		chars := []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}
		i := 0
		for {
			select {
			case <-p.stop:
				fmt.Print("\r")
				fmt.Printf("%s ... done!\n", msg)
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

// Stop beendet den Spinner.
func (p *Progress) Stop() {
	close(p.stop)
}