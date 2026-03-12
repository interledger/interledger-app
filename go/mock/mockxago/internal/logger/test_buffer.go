package logger

import (
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// bufferedSyncer wraps a write syncer and captures output to a buffer
type bufferedSyncer struct {
	sync.Mutex
	target      zapcore.WriteSyncer
	buffer      []byte
	isBuffering bool
}

func (bs *bufferedSyncer) Write(p []byte) (n int, err error) {
	bs.Lock()
	defer bs.Unlock()

	// If buffering is enabled, capture output in memory.
	if bs.isBuffering {
		bs.buffer = append(bs.buffer, p...)
		return len(p), nil
	}

	// Otherwise pass through directly.
	return bs.target.Write(p)
}

func (bs *bufferedSyncer) Sync() error {
	bs.Lock()
	defer bs.Unlock()
	return bs.target.Sync()
}

func (bs *bufferedSyncer) StartBuffering() {
	bs.Lock()
	defer bs.Unlock()
	bs.buffer = []byte{}
	bs.isBuffering = true
}

func (bs *bufferedSyncer) FlushBuffer(w io.Writer) {
	bs.Lock()
	defer bs.Unlock()
	if len(bs.buffer) > 0 {
		w.Write(bs.buffer)
		bs.buffer = []byte{}
	}
}

func (bs *bufferedSyncer) ClearBuffer() {
	bs.Lock()
	defer bs.Unlock()
	bs.buffer = []byte{}
}

func (bs *bufferedSyncer) StopBuffering() {
	bs.Lock()
	defer bs.Unlock()
	bs.isBuffering = false
}

// Initialize buffers with direct assignment to ensure they're ready before logger.go's init() runs
var (
	stdoutBuffer = &bufferedSyncer{
		target:      zapcore.Lock(zapcore.AddSync(os.Stdout)),
		isBuffering: false,
	}

	stderrBuffer = &bufferedSyncer{
		target:      zapcore.Lock(zapcore.AddSync(os.Stderr)),
		isBuffering: false,
	}
)

// BuildTestLogger creates a logger that buffers output for test failure scenarios
// Call this from your test setup, and use defer BufferLogs()() to auto-flush on failure
func BuildTestLogger() *zap.Logger {
	// Create encoder config for JSON logging
	encoderConfig := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create stdout sink for info/debug/warn with buffer
	stdoutCore := zapcore.NewCore(
		encoder,
		stdoutBuffer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel
		}),
	)

	// Create stderr sink for error/fatal with buffer
	stderrCore := zapcore.NewCore(
		encoder,
		stderrBuffer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		}),
	)

	// Combine cores with tee
	core := zapcore.NewTee(stdoutCore, stderrCore)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// BufferLogs enables log buffering and returns a cleanup function
// Use: defer test.BufferLogs()()
// Logs will be buffered and only output if the test fails
func BufferLogs() func() {
	stdoutBuffer.StartBuffering()
	stderrBuffer.StartBuffering()

	return func() {
		stdoutBuffer.StopBuffering()
		stderrBuffer.StopBuffering()

		// Note: Logs are automatically flushed when test ends
		// For explicit failure detection, see FlushLogsOnFailure
	}
}

// FlushLogsOnFailure should be called at the end of a test
// It checks if the test failed and flushes the buffered logs if so
// Usage: defer test.FlushLogsOnFailure(t)()
func FlushLogsOnFailure(t interface {
	Failed() bool
	Log(...interface{})
}) func() {
	startBuffering := func() {
		stdoutBuffer.StartBuffering()
		stderrBuffer.StartBuffering()
	}
	startBuffering()

	return func() {
		stdoutBuffer.StopBuffering()
		stderrBuffer.StopBuffering()

		if t.Failed() {
			// Flush to output on failure
			stdoutBuffer.FlushBuffer(os.Stdout)
			stderrBuffer.FlushBuffer(os.Stderr)
		} else {
			// Discard on success
			stdoutBuffer.ClearBuffer()
			stderrBuffer.ClearBuffer()
		}
	}
}

// GetBufferedLogs returns the current buffered logs as a string
// This is useful for manual inspection during debugging
func GetBufferedLogs() string {
	stdoutBuffer.Lock()
	stderrBuffer.Lock()
	defer stdoutBuffer.Unlock()
	defer stderrBuffer.Unlock()

	return string(append(append([]byte{}, stdoutBuffer.buffer...), stderrBuffer.buffer...))
}

// StartBufferingLogs starts buffering all log output (for godog/E2E tests)
func StartBufferingLogs() {
	stdoutBuffer.StartBuffering()
	stderrBuffer.StartBuffering()
}

// FlushBufferedLogs outputs all buffered logs to stdout/stderr (for failed scenarios)
func FlushBufferedLogs() {
	stdoutBuffer.FlushBuffer(os.Stdout)
	stderrBuffer.FlushBuffer(os.Stderr)
	stdoutBuffer.StopBuffering()
	stderrBuffer.StopBuffering()
}

// DiscardBufferedLogs clears all buffered logs without outputting (for passed scenarios)
func DiscardBufferedLogs() {
	stdoutBuffer.ClearBuffer()
	stderrBuffer.ClearBuffer()
	stdoutBuffer.StopBuffering()
	stderrBuffer.StopBuffering()
}
