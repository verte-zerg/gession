//nolint:forbidigo
package assert

import (
	"fmt"
	"os"
)

const (
	debug = true
)

func Assert(value bool, msg string, data ...any) {
	if !value {
		Fatal(msg, data...)
	}
}

func Fatal(msg string, data ...any) {
	if debug {
		panic(fmt.Sprintf(msg, data...))
	}

	fmt.Println("\r\n", fmt.Sprintf(msg, data...))
	os.Exit(1)
}
