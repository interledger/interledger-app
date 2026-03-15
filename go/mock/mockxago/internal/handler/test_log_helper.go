package handler

import (
	"testing"
)

// TestHelperWithLogBuffer is a helper function for tests that want to see logs only on failure
// Usage in your test:
//
//	func TestMyFeature(t *testing.T) {
//	    // Start buffering logs (will be shown only if test fails)
//	    defer logger.FlushLogsOnFailure(t)()
//
//	    // Your test code here
//	    h := setupAuthHandler(t)
//	    // ... rest of test
//	}
//
// This pattern will:
// - Buffer all logs during test execution
// - Discard them if the test passes
// - Display them if the test fails
//
// This keeps the test output clean while maintaining debuggability.
func TestHelperWithLogBuffer(t *testing.T) {
	// This is not a real test, just a helper function
	// See the comment above for usage
	t.Skip("This is a helper function, not a test")
}

// ExampleTestWithLogBuffer shows how to use log buffering in tests
func ExampleTestWithLogBuffer() {
	// To enable log buffering in your test:
	//
	// func TestSomething(t *testing.T) {
	//     // Option 1: Buffer logs, show only on failure (recommended)
	//     defer logger.FlushLogsOnFailure(t)()
	//
	//     // Option 2: Just buffer logs (manual handling)
	//     defer logger.BufferLogs()()
	//
	//     // Option 3: Check logs manually during test
	//     logs := logger.GetBufferedLogs()
	//     t.Logf("Current buffered logs:\n%s", logs)
	//
	//     // Now run your test code...
	// }
}
