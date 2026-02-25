package main

import (
	"flag"
	"fmt"
)

var (
	// debugOutput controls whether debug output is displayed
	debugOutput = flag.Bool("debug", true, "Enable debug output (default: true)")
)

// debugPrintf prints formatted output only if debug output is enabled
func debugPrintf(format string, args ...interface{}) {
	if *debugOutput {
		fmt.Printf(format, args...)
	}
}

// debugPrintln prints a line only if debug output is enabled
func debugPrintln(args ...interface{}) {
	if *debugOutput {
		fmt.Println(args...)
	}
}

// IsDebugEnabled returns whether debug output is enabled
func IsDebugEnabled() bool {
	return *debugOutput
}
