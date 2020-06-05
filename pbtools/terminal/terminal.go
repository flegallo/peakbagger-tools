package terminal

import (
	"fmt"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
	white  = "\033[97m"
)

// Error print error
func Error(err error, format string, a ...interface{}) {
	var message = format
	if err != nil {
		message = fmt.Sprintf("%s [%s]", format, err)
	}
	fmt.Printf("%s%s%s\n", red, fmt.Sprintf(message, a...), reset)
}
