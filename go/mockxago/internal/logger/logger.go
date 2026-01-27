package logger

import (
	"io"
	"log"
	"os"
)

var (
	Info  = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	Warn  = log.New(os.Stdout, "WARN: ", log.LstdFlags)
	Error = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	debugLog *log.Logger
)

func init() {
	if os.Getenv("LOG_LEVEL") == "debug" {
		debugLog = log.New(os.Stdout, "DEBUG: ", log.LstdFlags)
	} else {
		debugLog = log.New(io.Discard, "DEBUG: ", log.LstdFlags)
	}
}

// Fatal logs and exits
func Fatal(msg string, err error) {
	Error.Fatalf("%s: %v", msg, err)
}

// Infof logs an info message
func Infof(format string, v ...interface{}) {
	Info.Printf(format, v...)
}

// Warnf logs a warning message
func Warnf(format string, v ...interface{}) {
	Warn.Printf(format, v...)
}

// Errorf logs an error message
func Errorf(format string, v ...interface{}) {
	Error.Printf(format, v...)
}

// Debugf logs a debug message
func Debugf(format string, v ...interface{}) {
	debugLog.Printf(format, v...)
}
