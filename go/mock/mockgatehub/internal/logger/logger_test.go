package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitialize_ValidLevel(t *testing.T) {
	err := Initialize("debug")
	require.NoError(t, err)

	l := Logger()
	assert.NotNil(t, l)
}

func TestInitialize_InvalidLevel(t *testing.T) {
	err := Initialize("bogus")
	assert.Error(t, err)
}

func TestLogger_ReturnsNonNil(t *testing.T) {
	l := Logger()
	assert.NotNil(t, l)
}

func TestConvenienceFunctions(t *testing.T) {
	// Just ensure they don't panic. These write to stdout/stderr which is fine in tests.
	assert.NotPanics(t, func() {
		Info("test info", zap.String("key", "val"))
		Warn("test warn")
		Error("test error")
		Debug("test debug")
	})
}

func TestBuildLogger_NilLevel(t *testing.T) {
	l := buildLogger(nil)
	assert.NotNil(t, l)
}
