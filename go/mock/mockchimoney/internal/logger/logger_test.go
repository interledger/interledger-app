package logger

import (
	"testing"
)

func TestBuildLoggerReturnsInstance(t *testing.T) {
	l := buildLogger(nil)
	if l == nil {
		t.Fatal("buildLogger(nil) returned nil")
	}
}

func TestInitializeRejectsInvalidLevel(t *testing.T) {
	err := Initialize("invalid-level")
	if err == nil {
		t.Fatal("Initialize() expected error for invalid log level")
	}
}

func TestInitializeAcceptsValidLevelAndLoggerIsUsable(t *testing.T) {
	if err := Initialize("debug"); err != nil {
		t.Fatalf("Initialize() unexpected error: %v", err)
	}

	if Logger() == nil {
		t.Fatal("Logger() returned nil after initialization")
	}

	Info("info message")
	Warn("warn message")
	Error("error message")
	Debug("debug message")
}
