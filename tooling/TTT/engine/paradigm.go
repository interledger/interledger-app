package engine

import (
	"fmt"
	"strings"
)

// Paradigm identifies the account topology used when seeding the simulation.
type Paradigm int

const (
	// ParadigmPOSTwo is the two-POS setup.
	// Both GateHub and Xago carry EUR system + liquidity accounts so that a
	// position account is created on each side during cross-provider transfers.
	ParadigmPOSTwo Paradigm = 1

	// ParadigmSingleGHEUR is the legacy single-POS setup.
	// Only GateHub carries EUR accounts; Xago operates in ZAR only.
	ParadigmSingleGHEUR Paradigm = 2

	// ParadigmStandard is an alias for the currently recommended paradigm.
	// It always points to the latest accepted setup.
	ParadigmStandard = ParadigmPOSTwo
)

// ValidParadigms lists every paradigm available at init time.
var ValidParadigms = []Paradigm{ParadigmPOSTwo, ParadigmSingleGHEUR}

// Name returns the human-readable display name for the paradigm.
func (p Paradigm) Name() string {
	switch p {
	case ParadigmPOSTwo:
		return "Standard — two-POS setup (recommended)"
	case ParadigmSingleGHEUR:
		return "Single GateHub EUR POS (legacy)"
	default:
		return fmt.Sprintf("unknown(%d)", int(p))
	}
}

// IsValid reports whether p is a known paradigm.
func (p Paradigm) IsValid() bool {
	for _, v := range ValidParadigms {
		if p == v {
			return true
		}
	}
	return false
}

// SeedParadigm seeds the engine with the providers and accounts required by p.
// Duplicate-creation errors (i.e. accounts that already exist) are silently
// ignored so that calling SeedParadigm on an already-seeded store is idempotent.
func SeedParadigm(p Paradigm, eng *Engine) error {
	switch p {
	case ParadigmPOSTwo:
		return seedPOSTwo(eng)
	case ParadigmSingleGHEUR:
		return seedSingleGHEUR(eng)
	default:
		return fmt.Errorf("SeedParadigm: unknown paradigm %d", int(p))
	}
}

// skipAlreadyExists returns nil for "already exists" errors; any other error
// is returned unchanged.
func skipAlreadyExists(err error) error {
	if err == nil || strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}

// seedPOSTwo seeds the Standard / two-POS paradigm.
//
// Account topology:
//   - GateHub: system(EUR), liquidity(EUR)
//   - Xago:    system(ZAR), liquidity(ZAR), system(EUR), liquidity(EUR)
//
// Position accounts are created on demand during cross-provider transfers;
// having EUR liquidity on both sides ensures both positions are established.
func seedPOSTwo(eng *Engine) error {
	type providerSeed struct{ id, name string }
	for _, ps := range []providerSeed{
		{"gatehub", "GateHub"},
		{"xago", "Xago"},
	} {
		if _, err := eng.CreateProvider(ps.id, ps.name); err != nil {
			if err := skipAlreadyExists(err); err != nil {
				return err
			}
		}
	}

	type acctSeed struct {
		fn func() (Account, error)
	}
	seeds := []acctSeed{
		{func() (Account, error) { return eng.CreateSystemAccount("gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateLiquidityAccount("gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateSystemAccount("xago", ZAR) }},
		{func() (Account, error) { return eng.CreateLiquidityAccount("xago", ZAR) }},
		{func() (Account, error) { return eng.CreateSystemAccount("xago", EUR) }},
		{func() (Account, error) { return eng.CreateLiquidityAccount("xago", EUR) }},
		{func() (Account, error) { return eng.CreateUserAccount("alice", "gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateUserAccount("bob", "gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateUserAccount("carlos", "xago", ZAR) }},
	}
	for _, s := range seeds {
		if _, err := s.fn(); err != nil {
			if err := skipAlreadyExists(err); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedSingleGHEUR seeds the legacy single-POS paradigm.
//
// Account topology:
//   - GateHub: system(EUR), liquidity(EUR)
//   - Xago:    system(ZAR), liquidity(ZAR)
//
// Xago has no EUR accounts, so cross-provider transfers use only the GateHub
// EUR position account (single POS on the GateHub side).
func seedSingleGHEUR(eng *Engine) error {
	type providerSeed struct{ id, name string }
	for _, ps := range []providerSeed{
		{"gatehub", "GateHub"},
		{"xago", "Xago"},
	} {
		if _, err := eng.CreateProvider(ps.id, ps.name); err != nil {
			if err := skipAlreadyExists(err); err != nil {
				return err
			}
		}
	}

	type acctSeed struct {
		fn func() (Account, error)
	}
	seeds := []acctSeed{
		{func() (Account, error) { return eng.CreateSystemAccount("gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateLiquidityAccount("gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateSystemAccount("xago", ZAR) }},
		{func() (Account, error) { return eng.CreateLiquidityAccount("xago", ZAR) }},
		{func() (Account, error) { return eng.CreateUserAccount("alice", "gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateUserAccount("bob", "gatehub", EUR) }},
		{func() (Account, error) { return eng.CreateUserAccount("carlos", "xago", ZAR) }},
	}
	for _, s := range seeds {
		if _, err := s.fn(); err != nil {
			if err := skipAlreadyExists(err); err != nil {
				return err
			}
		}
	}
	return nil
}
