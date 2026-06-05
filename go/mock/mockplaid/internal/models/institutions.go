package models

// Mock institution IDs.
const (
	// InstitutionA — deterministic: each account always yields the same
	// account_id, so re-selecting drives the duplicate-link path.
	InstitutionA = "ins_mock_a"
	// InstitutionB — always-new: every selection mints a fresh account_id,
	// so it never duplicates (multi-account/stress path).
	InstitutionB = "ins_mock_b"
)

// CatalogAccount is a selectable account template in a mock institution.
type CatalogAccount struct {
	Key     string // "checking" | "savings"
	Name    string
	Mask    string
	Subtype string
}

// Institution is a mock bank with a fixed set of selectable accounts.
type Institution struct {
	ID       string
	Name     string
	Accounts []CatalogAccount
}

// Account returns the catalog account for a key.
func (i Institution) Account(key string) (CatalogAccount, bool) {
	for _, a := range i.Accounts {
		if a.Key == key {
			return a, true
		}
	}
	return CatalogAccount{}, false
}

// Catalog is the fixed set of mock institutions.
var Catalog = map[string]Institution{
	InstitutionA: {
		ID:   InstitutionA,
		Name: "Tartan Bank",
		Accounts: []CatalogAccount{
			{Key: "checking", Name: "Plaid Checking", Mask: "0000", Subtype: "checking"},
			{Key: "savings", Name: "Plaid Saving", Mask: "1111", Subtype: "savings"},
		},
	},
	InstitutionB: {
		ID:   InstitutionB,
		Name: "Platypus Bank",
		Accounts: []CatalogAccount{
			{Key: "checking", Name: "Plaid Checking", Mask: "2222", Subtype: "checking"},
			{Key: "savings", Name: "Plaid Saving", Mask: "3333", Subtype: "savings"},
		},
	},
}
