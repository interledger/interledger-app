package logger

import (
	"testing"
)

// BufferLogsForTest buffers all logger output during test execution.
// Returns a cleanup function that should be deferred.
//
// Example usage:
//
//	func TestSomething(t *testing.T) {
//		defer logger.FlushLogsOnFailure(t)()
//
//		// ... test code here ...
//		// Logs are buffered during test execution.
//		// If test fails: buffered logs are flushed to stdout for debugging.
//		// If test passes: buffered logs are silently discarded.
//	}
func BufferLogsForTest(t *testing.T) func() {
	stdoutBuffer.StartBuffering()
	stderrBuffer.StartBuffering()

	return func() {
		stdoutBuffer.StopBuffering()
		stderrBuffer.StopBuffering()
	}
}

// GetBufferedLogsAsString returns all buffered logs as a single string.
// Useful for custom assertions or debugging.
func GetBufferedLogsAsString() string {
	stdoutBuffer.Lock()
	defer stdoutBuffer.Unlock()
	return string(stdoutBuffer.buffer)
}
