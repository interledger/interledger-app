//go:build ruleguard

// Package gorules holds custom ruleguard rules enforced via gocritic in
// golangci-lint. The build tag keeps this file out of normal `go build`; it is
// loaded directly by ruleguard. See go/.golangci.yml for wiring.
package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// environmentMode forbids comparing an EnvironmentConfig's Mode field against a
// string literal. Use the helper methods instead (IsModeProd, IsModeSandbox,
// IsModeDev, IsModeLocal, IsModeTest) so the set of valid modes stays defined in
// one place and call sites read intent rather than a magic string.
func environmentMode(m dsl.Matcher) {
	m.Match(
		`$x.Environment.Mode == $y`,
		`$x.Environment.Mode != $y`,
	).Report(`compare environment mode via the helper methods (IsModeProd, IsModeSandbox, IsModeDev, IsModeLocal, IsModeTest) instead of comparing .Environment.Mode directly`)
}
