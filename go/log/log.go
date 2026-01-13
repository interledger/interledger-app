package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	// Create encoder config for JSON logging
	encoderConfig := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create stdout sink for info/debug/warn
	stdoutSink := zapcore.Lock(zapcore.AddSync(os.Stdout))
	stdoutCore := zapcore.NewCore(
		encoder,
		stdoutSink,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel
		}),
	)

	// Create stderr sink for error/fatal
	stderrSink := zapcore.Lock(zapcore.AddSync(os.Stderr))
	stderrCore := zapcore.NewCore(
		encoder,
		stderrSink,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		}),
	)

	// Combine cores with tee
	// AddCallerSkip(1) moves the stack 1 higher than the actual call to logger.*.
	// This means the file and line number will be displayed the called the global lib i.e.
	// WARN	ledger/service.go:121 instead of WARN	log/log.go:37
	core := zapcore.NewTee(stdoutCore, stderrCore)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
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

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Fatalln(err error) {
	logger.Fatal("fatal error occurred", zap.Error(err))
}
