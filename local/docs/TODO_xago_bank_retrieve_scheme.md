# TODO: Migrate Backend to Official Xago API Format

**Status**: ⏳ Pending (separate PR required)  
**Priority**: Low (non-blocking, workaround in place)  
**Estimated Effort**: 4-6 hours  
**Created**: February 24, 2026

---

## Problem

Backend expects nested `bankingProviders` structure (from MockBOS legacy format), but [official Xago API](https://documenter.getpostman.com/view/49463771/2sB3QRo7pf) returns flat structure.

**Current workaround**: MockXago returns nested format to match backend expectations.  
**Long-term solution**: Update backend to parse official Xago API format.

---

## Backend Changes Required

### 1. Add Flat Format Support

**File**: `go/backend/providers/xago/external/types.go`

```go
// Flat format from official Xago API
type CurrencyFlat struct {
    CurrencyID    string `json:"currencyId"`
    BankName      string `json:"bankName"`
    AccountNumber string `json:"accountNumber"`
    BranchCode    string `json:"branchCode"`
    SwiftBIC      string `json:"swiftBIC"`
}

// Convert flat to nested for backward compatibility
func (c *CurrencyFlat) ToNested() *Currency {
    return &Currency{
        CurrencyCode: c.CurrencyID,
        DepositEnabled: true,
        BankingProviders: []BankProvider{
            {
                Name: c.BankName,
                DepositAvailable: true,
                DepositFields: DepositFields{
                    BankName:      c.BankName,
                    AccountNumber: c.AccountNumber,
                    BranchCode:    c.BranchCode,
                    SwiftBIC:      c.SwiftBIC,
                },
            },
        },
    }
}
```

### 2. Update API Client

**File**: `go/backend/providers/xago/external/client.go`

```go
func (c *client) BankAccounts(ctx context.Context) (*[]Currency, error) {
    // ... make request ...
    
    // Try nested format first (backward compatibility)
    var nested []Currency
    if err := json.Unmarshal(body, &nested); err == nil {
        return &nested, nil
    }
    
    // Parse flat format (official API)
    var flat []CurrencyFlat
    if err := json.Unmarshal(body, &flat); err != nil {
        return nil, fmt.Errorf("failed to parse currencies: %w", err)
    }
    
    // Convert to nested
    result := make([]Currency, len(flat))
    for i, f := range flat {
        result[i] = *f.ToNested()
    }
    return &result, nil
}
```

### 3. No Changes Needed

**File**: `go/backend/providers/xago/ops/ops.go`  
`GetBankAccount()` logic remains unchanged.

---

## Testing Checklist

- [ ] Add unit test for `CurrencyFlat.ToNested()` conversion
- [ ] Test `BankAccounts()` with both formats
- [ ] Verify `GetBankAccount()` finds ZAR account correctly
- [ ] Run E2E deposit test with MockXago
- [ ] Update MockXago to return flat format after backend deployed
- [ ] Verify E2E tests pass with flat format

---

## Migration Steps

1. **Add flat format support** (backward compatible)
2. **Deploy backend** to staging/production
3. **Update MockXago** to return flat format (matches official API)
4. **Remove nested format** support after migration complete

---

## Why This Matters

- Production Xago API uses flat format (per official docs)
- Backend currently coded to MockBOS custom nested format
- Risk: Backend may fail with real Xago if format differs
- Benefit: Future-proof against API changes, matches official specification

---

## Key Files

**Backend (modify)**:
- `go/backend/providers/xago/external/types.go` - Add flat format types
- `go/backend/providers/xago/external/client.go` - Add format detection

**MockXago (update after backend deployed)**:
- `go/mockxago/internal/handler/currency.go` - Switch to flat format

**Reference**:
- [Official Xago API Docs](https://documenter.getpostman.com/view/49463771/2sB3QRo7pf) - Target format
- [Backend ops.go#L395](../../go/backend/providers/xago/ops/ops.go#L395) - Current usage

---

**Last Updated**: February 24, 2026

