package common

import (
	"fmt"
	"log"
	"os"
)

const (
	logFormat = log.Ldate | log.Ltime | log.Lmsgprefix | log.Lshortfile
)

type Logger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
}

func NewLogger(name string) *Logger {
	return &Logger{
		debug: log.New(os.Stdout, fmt.Sprintf("DEBUG: %s: ", name), logFormat),
		info:  log.New(os.Stdout, fmt.Sprintf("INFO: %s: ", name), logFormat),
		warn:  log.New(os.Stdout, fmt.Sprintf("WARNING: %s: ", name), logFormat),
		err:   log.New(os.Stdout, fmt.Sprintf("ERROR: %s: ", name), logFormat),
	}
}

func doFormat(format string, v ...any) string {
	if len(v) == 0 {
		return format
	}
	return fmt.Sprintf(format, v...)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.debug.Output(2, doFormat(format, v...))
}

func (l *Logger) Infof(format string, v ...any) {
	l.info.Output(2, doFormat(format, v...))
}

func (l *Logger) Warningf(format string, v ...any) {
	l.warn.Output(2, doFormat(format, v...))
}

func (l *Logger) Errorf(format string, v ...any) {
	l.err.Output(2, doFormat(format, v...))
}
