package log

import (
	"log"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level.SetLevel(zap.DebugLevel)

	var err error
	// AddCallerSkip(1) moves the stack 1 higher than the actual call to logger.*.
	// This means the file and line number will be displayed the called the global lib i.e.
	// WARN	ledger/service.go:121 instead of WARN	log/log.go:37
	logger, err = cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		log.Fatalln(err)
	}
}

func Setup(newLogger *zap.Logger) {
	if newLogger == nil {
		return
	}
	logger = newLogger
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}
