package util

import (
    "fmt"
    "log"    
    "os"
)

const (
    logFormat = log.Ldate | log. Ltime | log.Lmsgprefix | log.Lshortfile
)

type Logger struct {
    info  *log.Logger
    warn  *log.Logger
    err   *log.Logger
}

func NewLogger(name string) *Logger {
    return  &Logger {
        info: log.New(os.Stdout, "INFO: ", logFormat),
        warn: log.New(os.Stdout, "WARNING: ", logFormat),
        err:  log.New(os.Stdout, "ERROR: ", logFormat),
    }
}

func doFormat(format string, v ...any) string {
    if len(v) == 0 {
        return format 
    }
    return fmt.Sprintf(format, v...)
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
