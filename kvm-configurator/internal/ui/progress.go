// ui/progress.go
// last modification: Feb 03 2026
package ui

import (
	"fmt"
	"time"
)

// Progress encapsulates a Spinner go-routine
type Progress struct {
	stop chan struct{}
}

// NewProgress starts a spinner with the message passed
func NewProgress(msg string) *Progress {
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

// Stop
func (p *Progress) Stop() {
	close(p.stop)
}
// EOF