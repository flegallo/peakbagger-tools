package terminal

import (
	"fmt"
	"time"
)

const spinner = `|/-\`

// Operation represents a long running operation
type Operation struct {
	channel chan bool
}

// NewOperation starts a long running operation
func NewOperation(format string, a ...interface{}) *Operation {
	c := make(chan bool)
	spinFrames := []rune(spinner)
	spinFramesSize := len(spinFrames)

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		pos := 0

	L:
		for {
			select {
			case <-c:
				break L
			case <-ticker.C:
				fmt.Printf("\r  %s%s%s %s ", yellow, fmt.Sprintf(format, a...), reset, string(spinFrames[pos%spinFramesSize]))
				pos++
			}
		}
	}()

	return &Operation{
		channel: c,
	}
}

// Success informs that the operation is over
func (o *Operation) Success(format string, a ...interface{}) {
	o.finished("\u2713", green, format, a...)
}

// Success informs that the operation is over
func (o *Operation) Error(err error, format string, a ...interface{}) {
	var message = format
	if err != nil {
		message = fmt.Sprintf("%s [%s]", format, err)
	}
	o.finished("\u2717", red, message, a...)
}

// Success informs that the operation is over
func (o *Operation) finished(symbol string, color string, format string, a ...interface{}) {
	o.channel <- true

	fmt.Printf("\033[2K")
	fmt.Printf("\r%s %s%s%s \n", symbol, color, fmt.Sprintf(format, a...), reset)
}
