package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitializeRejectsInvalidLevel(t *testing.T) {
	err := Initialize("not-a-level")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestLoggerWritesToBuffer(t *testing.T) {
	requireNoError := func(err error) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	requireNoError(Initialize("debug"))
	StartBufferingLogs()
	defer DiscardBufferedLogs()

	Info("info message", zap.String("k", "v"))
	Warn("warn message")
	Error("error message")
	Debug("debug message")
	Infof("formatted %s", "info")
	Warnf("formatted %s", "warn")
	Errorf("formatted %s", "error")
	Debugf("formatted %s", "debug")
	Logger().Sync()

	logs := GetBufferedLogs()
	assert.Contains(t, logs, "info message")
	assert.Contains(t, logs, "warn message")
	assert.Contains(t, logs, "error message")
	assert.Contains(t, logs, "debug message")
	assert.Contains(t, logs, "formatted info")
	assert.Contains(t, logs, "formatted warn")
	assert.Contains(t, logs, "formatted error")
	assert.Contains(t, logs, "formatted debug")
}

func TestBufferingHelpers(t *testing.T) {
	requireNoError := func(err error) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	requireNoError(Initialize("info"))
	cleanup := BufferLogs()
	Info("buffered message")
	cleanup()
	assert.Contains(t, GetBufferedLogs(), "buffered message")

	tCleanup := FlushLogsOnFailure(t)
	Info("message in FlushLogsOnFailure")
	tCleanup()
}
