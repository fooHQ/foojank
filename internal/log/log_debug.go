//go:build debug

package log

import "log"

func Debug(format string, v ...any) {
	log.Printf("<DEBUG> "+format, v...)
}
