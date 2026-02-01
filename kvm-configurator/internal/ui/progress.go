// ui/progress.go
// last modification: February 01 2026
package ui

import (
	"fmt"
	"time"
)

func SimpleProgress(msg string, stopChan <-chan struct{}) {
	go func() {
		chars := []rune{'⣾','⣽','⣻','⢿','⡿','⣟','⣯','⣷'}
		i := 0
		for {
			select {
			case <-stopChan:
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
}
// EOF