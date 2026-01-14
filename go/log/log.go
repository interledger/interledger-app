package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	// Default logger: info/debug/warn -> stdout, error+ -> stderr.
	// No configured minimum level filter (matches your current init behavior).
	logger = buildLogger(nil)
}

func buildLogger(minLevel *zapcore.Level) *zap.Logger {
	// Create encoder config for JSON logging
	encoderConfig := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create stdout sink for info/debug/warn
	stdoutSink := zapcore.Lock(zapcore.AddSync(os.Stdout))
	stdoutCore := zapcore.NewCore(
		encoder,
		stdoutSink,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			if lvl >= zapcore.ErrorLevel {
				return false
			}
			if minLevel == nil {
				return true
			}
			return lvl >= *minLevel
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
	// This means the file and line number will be displayed at the call site, e.g.
	// WARN	ledger/service.go:121 instead of WARN	log/log.go:37
	core := zapcore.NewTee(stdoutCore, stderrCore)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// Initialize reconfigures the global logger with the specified log level.
// This should be called during application startup with the configured log level.
func Initialize(logLevel string) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	logger = buildLogger(&level)
	return nil
}

// Setup overwrites the global logger. Deprecated: use Initialize() instead.
func Setup(newLogger *zap.Logger) {
	if newLogger == nil {
		return
	}
	logger = newLogger
}

// Logger returns the global logger instance.
func Logger() *zap.Logger {
	return logger
}

func Info(msg string, fields ...zap.Field)  { logger.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { logger.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { logger.Error(msg, fields...) }
func Debug(msg string, fields ...zap.Field) { logger.Debug(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { logger.Fatal(msg, fields...) }

func Fatalln(err error) {
	logger.Fatal("fatal error occurred", zap.Error(err))
}
