package logger

import (
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// bufferedSyncer wraps a write syncer and captures output to a buffer.
type bufferedSyncer struct {
	sync.Mutex
	target      zapcore.WriteSyncer
	buffer      []byte
	isBuffering bool
}

func (bs *bufferedSyncer) Write(p []byte) (n int, err error) {
	bs.Lock()
	defer bs.Unlock()

	if bs.isBuffering {
		bs.buffer = append(bs.buffer, p...)
		return len(p), nil
	}
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
		_, _ = w.Write(bs.buffer)
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

// BuildTestLogger creates a logger that buffers output for test failure scenarios.
func BuildTestLogger() *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	stdoutCore := zapcore.NewCore(
		encoder,
		stdoutBuffer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel
		}),
	)

	stderrCore := zapcore.NewCore(
		encoder,
		stderrBuffer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		}),
	)

	core := zapcore.NewTee(stdoutCore, stderrCore)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// FlushLogsOnFailure should be called at the end of a test.
func FlushLogsOnFailure(t interface {
	Failed() bool
}) func() {
	stdoutBuffer.StartBuffering()
	stderrBuffer.StartBuffering()

	return func() {
		stdoutBuffer.StopBuffering()
		stderrBuffer.StopBuffering()

		if t.Failed() {
			stdoutBuffer.FlushBuffer(os.Stdout)
			stderrBuffer.FlushBuffer(os.Stderr)
		} else {
			stdoutBuffer.ClearBuffer()
			stderrBuffer.ClearBuffer()
		}
	}
}
