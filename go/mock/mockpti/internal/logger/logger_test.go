package logger

import (
	"bytes"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestInitialize_ValidLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		if err := Initialize(level); err != nil {
			t.Errorf("Initialize(%q) returned unexpected error: %v", level, err)
		}
	}
}

func TestInitialize_InvalidLevel(t *testing.T) {
	if err := Initialize("invalid-level"); err == nil {
		t.Error("Initialize with invalid level should return error")
	}
}

func TestLogger_NotNil(t *testing.T) {
	if Logger() == nil {
		t.Error("Logger() should return non-nil logger")
	}
}

func TestInfof(t *testing.T) {
	defer FlushLogsOnFailure(t)()
	_ = Initialize("info")
	Infof("test info %s", "message")
}

func TestErrorf(t *testing.T) {
	defer FlushLogsOnFailure(t)()
	_ = Initialize("debug")
	Errorf("test error %s", "message")
}

func TestInfo(t *testing.T) {
	defer FlushLogsOnFailure(t)()
	_ = Initialize("info")
	Info("test info field message")
}

func TestWarn(t *testing.T) {
	defer FlushLogsOnFailure(t)()
	_ = Initialize("debug")
	Warn("test warn message")
}

func TestDebug(t *testing.T) {
	defer FlushLogsOnFailure(t)()
	_ = Initialize("debug")
	Debug("test debug message")
}

func TestBuildTestLogger(t *testing.T) {
	l := BuildTestLogger()
	if l == nil {
		t.Error("BuildTestLogger() should return non-nil logger")
	}
}

func TestBufferedSyncer_WriteAndFlush(t *testing.T) {
	bs := &bufferedSyncer{
		target: zapcore.AddSync(bytes.NewBuffer(nil)),
	}
	bs.StartBuffering()
	_, _ = bs.Write([]byte("hello"))
	bs.FlushBuffer(bytes.NewBuffer(nil))
}

func TestBufferedSyncer_ClearBuffer(t *testing.T) {
	bs := &bufferedSyncer{
		target: zapcore.AddSync(bytes.NewBuffer(nil)),
	}
	bs.StartBuffering()
	_, _ = bs.Write([]byte("data"))
	bs.ClearBuffer()
	bs.StopBuffering()
}

func TestBufferedSyncer_Sync(t *testing.T) {
	bs := &bufferedSyncer{
		target: zapcore.AddSync(bytes.NewBuffer(nil)),
	}
	_ = bs.Sync()
}
