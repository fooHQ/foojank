//go:build dev

package log

import "log"

func Debug(format string, v ...any) {
	log.Printf("<DEBUG> "+format, v...)
}
