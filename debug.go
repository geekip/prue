package prue

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
)

var (
	DefaultWriter      io.Writer = os.Stdout
	DefaultErrorWriter io.Writer = os.Stderr
	debugPrintFunc     func(format string, values ...interface{})
	debugPrefix        string = "[Prue-debug] "
	debugErrorPrefix   string = debugPrefix + "[ERROR] "
)

func IsDebug() bool {
	return atomic.LoadInt32(&systemMode) == debugCode
}

func debugPrint(format string, values ...any) {
	if !IsDebug() {
		return
	}

	if debugPrintFunc != nil {
		debugPrintFunc(format, values...)
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(DefaultWriter, debugPrefix+format, values...)
}

func debugPrintError(err error) {
	if err != nil && IsDebug() {
		fmt.Fprintf(DefaultErrorWriter, debugErrorPrefix+"%v\n", err)
	}
}
