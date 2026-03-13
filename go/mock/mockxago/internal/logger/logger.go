package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	// Default logger: info/debug/warn -> stdout, error+ -> stderr.
	logger = buildLogger(nil)
}

func buildLogger(minLevel *zapcore.Level) *zap.Logger {
	// Create encoder config for JSON logging
	encoderConfig := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Use buffered sinks for test-friendly output
	stdoutSink := zapcore.Lock(stdoutBuffer)
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
	stderrSink := zapcore.Lock(stderrBuffer)
	stderrCore := zapcore.NewCore(
		encoder,
		stderrSink,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		}),
	)

	// Combine cores with tee
	// AddCallerSkip(1) moves the stack 1 higher than the actual call to logger.*.
	// This means the file and line number will be displayed at the call site.
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

// Logger returns the global logger instance.
func Logger() *zap.Logger {
	return logger
}

// Helper functions for convenience
func Info(msg string, fields ...zap.Field)  { logger.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { logger.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { logger.Error(msg, fields...) }
func Debug(msg string, fields ...zap.Field) { logger.Debug(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { logger.Fatal(msg, fields...) }

func Fatalln(err error) {
	logger.Fatal("fatal error occurred", zap.Error(err))
}

// Legacy compatibility functions (convert string format to structured logging)
func Infof(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
}

func Warnf(format string, v ...interface{}) {
	logger.Warn(fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	logger.Error(fmt.Sprintf(format, v...))
}

func Debugf(format string, v ...interface{}) {
	logger.Debug(fmt.Sprintf(format, v...))
}
