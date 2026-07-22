package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	logger = buildLogger(nil)
}

func buildLogger(minLevel *zapcore.Level) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderConfig)

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

	stderrSink := zapcore.Lock(stderrBuffer)
	stderrCore := zapcore.NewCore(
		encoder,
		stderrSink,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		}),
	)

	core := zapcore.NewTee(stdoutCore, stderrCore)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// Initialize reconfigures the global logger with the specified log level.
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

func Info(msg string, fields ...zap.Field)  { logger.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { logger.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { logger.Error(msg, fields...) }
func Debug(msg string, fields ...zap.Field) { logger.Debug(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { logger.Fatal(msg, fields...) }

func Fatalln(err error) {
	logger.Fatal("fatal error occurred", zap.Error(err))
}

func Infof(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	logger.Error(fmt.Sprintf(format, v...))
}
