# Phase 7 Implementation Plan: Wallet E2E Integration for Xago Signup Processes

## Overview

Phase 7 brings Xago signup scenarios into the E2E test suite. This phase focuses on testing the user signup flow for South Africa (ZA) with Xago as the wallet provider.

**Phase Outcome**: E2E feature tests for Xago user signup processes pass, with minimal changes to backend service configuration.

## Guiding Principles

1. **Avoid changes to `go/backend/` and `typescript/protea/` if at all possible.** Phase 7 is an E2E-only phase. Backend and frontend should already support ZA signup. If a backend or protea change turns out to be unavoidable (e.g., ZA `Supported` flag), document why and keep the change minimal.
2. **Port, don't rewrite.** The `bosbaber/mockxago` branch already has working code for everything in this phase. Cherry-pick and adapt rather than writing from scratch.
3. **SA phone numbers must follow South African format.** This is already solved on `bosbaber/mockxago` — see `getPhoneBaseForCountry("south africa") → "+27710000000"`. The `+27` country code and `71` mobile prefix must be preserved; only trailing digits vary per test run.

## Reference Branch: `bosbaber/mockxago`

The `bosbaber/mockxago` branch contains a complete, working implementation of all Xago E2E phases (signup, KYC, deposit, withdrawal, linked accounts). It was diff'd against the current branch (`stephan/wal-639`) to determine exactly what to cherry-pick for Phase 7 and what to defer.

**Branch diff summary** (`e2e/` only — 25 files, +2849 / -382 lines):

| Category | Files | Phase |
|----------|-------|-------|
| Signup feature + steps | `000-signup.feature`, `gatehub_signup_clean.go`, `gatehub_workflows.go` | **7 (this phase)** |
| Health check + context | `context.go`, `godog_test.go` | **7 (this phase)** |
| Phone generation | `gatehub_forms.go` (new `generateDeterministicPhone`) | **7 (this phase)** |
| Scenario tagging | All `.feature` files (`@gatehub`/`@xago` tags) | **7 (this phase)** |
| DNS + local infra | `local/Makefile` (add `mockxago.interledger.test`) | **7 (this phase)** |
| Local env vars | `local/wallet.yaml` (add `GATEHUB_WIDGET_BASE_URL`) | **7 (this phase)** |
| KYC iframe (Persona) | `gatehub_kyc.go` (+112 lines), `001-kyc.feature` | Phase 8 |
| Deposit steps | `xago_deposit.go` (840 lines, new file) | Phase 9 |
| Withdrawal steps | `xago_withdrawal.go` (194 lines, new file) | Phase 11 |
| Linked accounts | `xago_link_account.go` (471 lines, new file), `005-linked-account-management.feature` | Phase 12 |
| DB helpers | `db_helpers.go` (+118 lines: `getWalletIDByEmail`, `getXagoAccountIDByWalletID`, `xagoLinkedAccountExists`) | Phase 8+ |
| Backend source changes | `go/backend/cli/cli.go` (rename + add Xago env vars) | **Deferred** (Phase 9+) |

## Current Status (verified 11 March 2026)

- ✓ Phase 5 complete: MockXago with 65%+ coverage and full feature tests
- ✓ Phase 6 complete: MockXago deployed in local Docker Compose environment
- ✓ `local/mockxago.yaml` exists with correct configuration (Redis DB 4, Traefik routing)
- ✓ `local/wallet.yaml` contains XAGO_* environment variables
- ✓ `local/docker-compose.yaml` includes `mockxago.yaml`
- ✓ Backend depends_on includes `mockxago`
- ✓ E2E test infrastructure operational (Playwright, Godog)
- ✓ GateHub German signup scenario exists and passes as reference
- ✓ Frontend already supports ZA in signup route
- ✓ Backend Xago provider fully initialized

## Scope

### What's Included (from `bosbaber/mockxago` diff)
- E2E signup scenario for South Africa (`@signup @xago` tag)
- `mockxago is running at` health check step definition (`context.go`)
- Country-aware phone generation refactor (`gatehub_forms.go` + `gatehub_signup_clean.go`)
- `@gatehub` tags on existing German scenarios (to enable isolated runs)
- Feature file restructuring (proper tagging, shared Background, scenario ordering)
- DNS entry for `mockxago.interledger.test` in `local/Makefile`
- Removal of `prerequisite()` GateHub-specific Temporal workflow from `godog_test.go`
- Backend env var alignment in `local/wallet.yaml` only (Docker Compose config, NOT `go/backend/` source)

### What's NOT Included (Defer to Later Phases)
- KYC/Persona iframe workflows — Phase 8
- Xago deposit steps (`xago_deposit.go`) — Phase 9
- P2P payments with ZAR — Phase 10
- Xago withdrawal steps (`xago_withdrawal.go`) — Phase 11
- Linked account management (`xago_link_account.go`) — Phase 12
- Changes to `go/backend/` source code (e.g., `cli.go` field renames, ops account fields) — defer until a phase actually needs them
- Changes to `typescript/protea/` — frontend already supports ZA signup

## Detailed Changes from `bosbaber/mockxago` (Phase 7 only)

### 1. `e2e/features/000-signup.feature`

**Exact diff to apply:**
- Add `And mockxago is running at "https://mockxago.interledger.test"` to Background
- Add `@signup @xago` scenario for South Africa (BEFORE the German scenario)
- Tag existing German scenario with `@signup @gatehub`
- Tag validation scenario with `@signup @gatehub`

```feature
@signup @xago
Scenario: Successfully sign up as a South-African user
  Given that my "country" is "South Africa"
  And I completed the signup workflow
  And I completed the account verification workflow
  And I finished the TOTP registration workflow
  And I finished the wallet address creation workflow
  Then I should be navigated back to the dashboard with reserved wallet status
  And I take a screenshot "signup-complete"
```

### 2. `e2e/context.go`

**Changes to apply:**
- Add step registration for `mockxago is running at` (line ~100, after mockgatehub step)
- Add `theMockxagoIsRunningAt` function (mirror of `theMockgatehubIsRunningAt`)
- Rename phone step: `"unique valid German number"` → `"unique valid phone number"`
- Register renamed step: `iFillInWithUniquePhoneNumber` instead of `iFillInWithUniqueGermanPhoneNumber`

**Not yet:** Skip the Xago deposit/withdrawal/linked-account step registrations (30+ new steps). Those are for later phases.

### 3. `e2e/gatehub_forms.go`

**New functions to add** (from `bosbaber/mockxago`):

```go
// getPhoneBaseForCountry returns the base phone number template for a given country.
// The trailing zeros in each base are replaced by a deterministic overlay.
func getPhoneBaseForCountry(country string) string {
    switch strings.ToLower(country) {
    case "south africa":
        // +27 country code, 71 mobile prefix → +27710000000 (11 digits)
        // South African mobile numbers: +27 XX XXXX XXX (9 digits after country code)
        return "+27710000000"
    default:
        // Germany / fallback: +49 country code, 1700 prefix → +491700000000 (12 digits)
        return "+491700000000"
    }
}

func generateDeterministicPhone(country, testEmailPrefix, emailSuffix string) string {
    // Uses getPhoneBaseForCountry + overlay logic:
    //   1. Get base for country (e.g. "+27710000000")
    //   2. Build 6-digit overlay: testPrefix[:3] + hash(emailUsername)[:3]
    //   3. Replace trailing zeros in base with overlay
    // Result for ZA: "+2771" + 6 deterministic digits (e.g. "+27711234567"... wait, that's too many)
    // Actually: base is 11 chars, overlay replaces last 6 → "+27710" stays, last 6 vary
}
```

**SA phone format constraint**: South African mobile numbers are `+27` + 9 digits. The `71` prefix in the base ensures a valid Vodacom/MTN range. The generator preserves `+2771` and only varies the last 6 digits, producing numbers like `+27710XXXXXX` — always valid SA format.

Also update `iFillInPhoneWithRandomNumber` to use `generateDeterministicPhone(sc.country, ...)`.

### 4. `e2e/gatehub_signup_clean.go`

**Changes to apply:**
- Rename `iFillInWithUniqueGermanPhoneNumber` → `iFillInWithUniquePhoneNumber`
- Replace inline phone generation with call to `generateDeterministicPhone(sc.country, sc.testIdentifier, emailSuffix)`
- In `iCompletedTheSignupWorkflow`: change `phonePrefix := "+49"` → `phonePrefix := ""`
  (phone prefix is now ignored; the generator picks the right base from `sc.country`)

### 5. `e2e/godog_test.go`

**Changes to apply** (from `bosbaber/mockxago`):
- Remove `prerequisite()` function and its call from `TestFeatures`
  - This ran `UpdateGateHubOrganizationConfig` Temporal workflow before all tests
  - On `bosbaber/mockxago`, this was removed because it's GateHub-specific and not needed when running `@xago` tests
  - **Decision needed**: Remove entirely (matching `bosbaber/mockxago`), or guard with `@gatehub` tag check?
- Remove `temporalWorkflowArgs` var and `workflowArgs` type
- Remove `"time"`, `"github.com/google/uuid"`, `"go.temporal.io/sdk/client"` imports
- Change report format from `pretty,pretty:%s` → `pretty,cucumber:%s`

### 6. `local/Makefile`

**Change:** Replace `docs.interledger.test` with `mockxago.interledger.test` in HOSTS list (matching `bosbaber/mockxago`).

### 7. `local/wallet.yaml`

**Current branch** (Phase 6) has XAGO env vars at lines 139-155 with `XAGO_WEBHOOK_SECRET`.
**`bosbaber/mockxago`** has them at lines 126-136 (no `XAGO_WEBHOOK_SECRET`, slightly different grouping).

Key differences:
| Var | Current (`stephan/wal-639`) | `bosbaber/mockxago` | Phase 7 Action |
|-----|---------------------------|---------------------|----------------|
| `XAGO_WEBHOOK_SECRET` | Present | **Absent** | Keep (safe default) |
| `MOCKXAGO_ENDPOINT` | Present (line ~157) | Present (line ~113, earlier) | Already done |
| `GATEHUB_WIDGET_BASE_URL` | **Absent** | Present | **Add** |
| `GATEHUB_ORGANIZATION_ID` | Present | **Removed** | Keep (avoid GateHub regression) |
| Xago vars position | After PTI block | After GateHub EUR ops block | Leave as-is |

**Conservative approach**: Only ADD `GATEHUB_WIDGET_BASE_URL`. Don't remove or move existing vars to minimize risk.

### 8. `go/backend/cli/cli.go` (DEFERRED)

**`bosbaber/mockxago` renames and adds:**
- `XagoPublicKey` → `XagoAPIPublicKey` (struct field name change)
- `XagoSecret` → `XagoAPISecret` (struct field name change)
- Added: `XagoUSDOpsAccount`, `XagoZAROpsAccount`, `XagoLedgerIDZAR`, `XagoLedgerIDUSD`
- Removed: `GatehubOrganizationID`

> **NOT in Phase 7 scope.** These changes are needed for Xago deposit/withdrawal (balance account operations), not for signup. The current `cli.go` field names work fine for E2E signup. Defer to Phase 9 when deposit tests actually need the ops account and ledger ID fields.

## Implementation Tasks

### Task 1: Cherry-Pick Structural E2E Changes (30 min)
**Objective**: Apply the refactored phone generation and step registration from `bosbaber/mockxago`

**Files**:
- `e2e/context.go` — add `mockxago is running at` step, rename phone step
- `e2e/gatehub_forms.go` — add `getPhoneBaseForCountry`, `generateDeterministicPhone`, update `iFillInPhoneWithRandomNumber`
- `e2e/gatehub_signup_clean.go` — rename phone function, use new generator, clear `phonePrefix`

**Implementation approach**: Port the exact code from `bosbaber/mockxago` for these functions. The `generateDeterministicPhone` function extracts/replaces the inline phone logic that was in `iFillInWithUniqueGermanPhoneNumber`.

### Task 2: Add MockXago Health Check Step (10 min)
**Objective**: Implement `theMockxagoIsRunningAt` in `e2e/context.go`

**Copy from `bosbaber/mockxago`** — the implementation is identical to `theMockgatehubIsRunningAt`:
```go
func (sc *E2EContext) theMockxagoIsRunningAt(urlStr string) error {
    parsedURL, err := url.Parse(urlStr)
    if err != nil { return fmt.Errorf("failed to parse mockxago URL: %w", err) }
    if err := sc.ensureHostsResolve([]string{parsedURL.Hostname()}); err != nil { return err }
    debugPrintf("🔍 Verifying mockxago health endpoint at %s...\n", urlStr)
    healthURL := strings.TrimSuffix(urlStr, "/") + "/health"
    return sc.waitForHealthEndpoint(healthURL, 30*time.Second)
}
```

### Task 3: Restructure Feature Files (30 min)
**Objective**: Restructure feature files to support both GateHub and Xago scenarios cleanly

The current feature files have no provider tags and assume GateHub-only. They need restructuring to:
1. Tag every existing scenario with `@gatehub` (enables filtered runs)
2. Add `@xago` scenarios where applicable (signup for Phase 7; KYC/deposit/etc. in later phases)
3. Add `mockxago` health check to shared Background blocks
4. Place `@xago` scenarios BEFORE `@gatehub` scenarios (matching `bosbaber/mockxago` convention)

**`e2e/features/000-signup.feature`** (full restructure):

Target state (matches `bosbaber/mockxago` exactly):
```feature
Feature: User Signup
  ...
  Background:
    Given a random test identifier is generated
    And the frontend is running at "https://interledger.test"
    And mockgatehub is running at "https://mockgatehub.interledger.test"
    And mockxago is running at "https://mockxago.interledger.test"
    Given the details of 'signup-user' are
      | field       | value                        |
      | emailSuffix | signup@example.com           |
      | password    | InterlEdger2025!TestPassword |
      | firstName   | Signupper                    |
      | lastName    | Donaldson                    |
      | dateOfBirth | 1995-03-20                   |
    And I impersonate 'signup-user'

  @signup @xago
  Scenario: Successfully sign up as a South-African user
    Given that my "country" is "South Africa"
    ...shared steps...

  @signup @gatehub
  Scenario: Successfully sign up as a German user
    Given that my "country" is "germany"
    ...shared steps...

  @signup @gatehub
  Scenario: Signup form validates required fields
    ...existing validation scenario unchanged...
```

Key structural changes vs current:
- `mockxago` health check line added to Background (after `mockgatehub`)
- New `@signup @xago` scenario inserted BEFORE the German one
- Existing `@signup` tags become `@signup @gatehub`
- Scenario steps are identical (provider-agnostic) — only country differs

**Other feature files** (light touch for Phase 7 — add mockxago to Background, tag existing scenarios):
- `e2e/features/001-kyc.feature` — add `mockxago` to Background, tag existing scenarios `@gatehub`
- `e2e/features/002-deposit.feature` — add `mockxago` to Background, tag existing scenarios `@gatehub`
- `e2e/features/003-p2p-payment.feature` — add `mockxago` to Background, tag existing scenarios `@gatehub`
- `e2e/features/004-withdrawal.feature` — add `mockxago` to Background, tag existing scenarios `@gatehub`

Do NOT add `@xago` scenarios to files 001-004 yet — those are for later phases.

### Task 4: Update `godog_test.go` (10 min)
**Objective**: Remove GateHub-specific prerequisite and simplify test setup

**From `bosbaber/mockxago`**:
- Remove `prerequisite()` function (runs `UpdateGateHubOrganizationConfig` Temporal workflow)
- Remove associated imports (`time`, `uuid`, `temporal client`)
- Remove `temporalWorkflowArgs` and `workflowArgs` type
- Change report format to `pretty,cucumber:%s`

**Risk**: Removing `prerequisite()` may break GateHub scenarios if GateHub org config is still needed. Check if the config was already set during `make all` startup.

### Task 5: Align Local Configuration (20 min)
**Objective**: Ensure `local/Makefile` and `local/wallet.yaml` are correct for Xago E2E

> **Principle: No `go/backend/` or `typescript/protea/` changes.** The `cli.go` field renames and new Xago ops account fields on `bosbaber/mockxago` are NOT needed for signup. Defer those to the phase that actually requires them (likely Phase 9 deposit or Phase 11 withdrawal). If signup fails without a backend change, investigate and document the specific reason before touching backend code.

**`local/Makefile`**: Add `mockxago.interledger.test` to HOSTS (replace `docs.interledger.test`)

**`local/wallet.yaml`** (Docker Compose env vars — this is infrastructure config, not backend source):
- Add `GATEHUB_WIDGET_BASE_URL=https://mockgatehub.interledger.test`
- Verify `MOCKXAGO_ENDPOINT=http://mockxago:8080` is present (already added in Phase 6)
- Decide on `XAGO_WEBHOOK_SECRET` (current branch has it, `bosbaber/mockxago` doesn't — keep it for now, remove only if it causes issues)
- Decide on `GATEHUB_ORGANIZATION_ID` removal (current branch has it, `bosbaber/mockxago` doesn't — keep it for now to avoid GateHub regression)

**NOT in scope for this task:**
- `go/backend/cli/cli.go` — field renames (`XagoPublicKey` → `XagoAPIPublicKey`) and new fields (`XagoLedgerIDZAR` etc.) are for later phases
- `go/backend/main.go` — no changes
- `typescript/protea/` — no changes (ZA already supported in signup route)

### Task 6: Verify ZA `Supported` Flag (10 min)
**Objective**: Check if `IsSupported()` blocks ZA signup

In `go/backend/country/types.go` line 588: `ZA: {Name: "South Africa", Numeric: "710"}` — lacks `Supported: true`.

**Action**: Search for `IsSupported()` usage in signup/wallet path. If it gates the flow, this is one of the rare cases where a `go/backend/` change is unavoidable — add `Supported: true` to the ZA entry.

> If this change IS needed, it's a one-line change to a data declaration, not a logic change. Document it clearly.

### Task 7: Run and Debug E2E Signup (1-2 hours)
**Objective**: Execute the Xago signup scenario end-to-end

```bash
cd local && make hosts  # After Task 5
cd local && make all
# Wait ~30s
cd local/scripts && ./local-dev-tool rafiki --skip-ui --wait-for-ready 120

cd e2e
go test -v -timeout 15m -run TestFeatures -args -tags="@xago&&@signup"
```

**Also run GateHub regression**:
```bash
go test -v -timeout 15m -run TestFeatures -args -tags="@gatehub&&@signup"
```

## Architecture Reference

### E2E Signup Flow (Provider-Agnostic)
```
Feature Background
  → Health-check frontend, mockgatehub, mockxago

Scenario: South-African signup
  → Set country = "South Africa"
  → iCompletedTheSignupWorkflow (gatehub_signup_clean.go)
    → Navigate to /signup
    → Fill form: name, email, country, phone (generateDeterministicPhone → +27710XXXXXX)
    → Submit → Kratos identity created
  → iCompletedTheAccountVerificationWorkflow (gatehub_workflows.go)
    → DB: verify Kratos identity email
  → iFinishedTheTOTPRegistrationWorkflow (gatehub_workflows.go)
    → Login → TOTP setup → Dashboard
  → iFinishedTheWalletAddressCreationWorkflow (gatehub_workflows.go)
    → Fill wallet address form → Submit → Reserved status
```

### Phone Number Generation (from `bosbaber/mockxago`)

**South African mobile format**: `+27` (country code) + 9 digits. Valid mobile prefixes: 71, 72, 73, 74, 76, 78, 79, 81, 82, 84. The generator uses `71` (Vodacom range).

```
getPhoneBaseForCountry("South Africa") → "+27710000000"  (11 chars total)
getPhoneBaseForCountry("Germany")      → "+491700000000" (12 chars total)

generateDeterministicPhone(country, testIdentifier, emailSuffix):
  base = getPhoneBaseForCountry(country)
  overlay = testIdentifier[:3] + hash(emailUsername)[:3]  → 6 deterministic digits
  result = base[:len(base)-len(overlay)] + overlay
  e.g. ZA: "+2771" + "038291" → "+27710382910"... trimmed to 11 chars → "+2771038291"
```

This is already fully implemented on `bosbaber/mockxago` in `e2e/gatehub_forms.go`. Port it directly — do not reinvent.

### Backend Country → Provider Routing
```
Signup: stores country_code=ZA in signups table
Wallet creation: wallet.Country=ZA in wallets table
KYC initiation: ZA → Persona inquiry → Phase 8
KYC approval: ZA → CreateKYCWallets → xago.CreateBalanceAccount(ZAR) → Phase 8
```

### Key Files Changed in Phase 7
| File | Change |
|------|--------|
| `e2e/features/000-signup.feature` | Add `@xago` scenario, `@gatehub` tags, mockxago Background |
| `e2e/context.go` | `theMockxagoIsRunningAt`, rename phone step |
| `e2e/gatehub_forms.go` | `getPhoneBaseForCountry`, `generateDeterministicPhone` |
| `e2e/gatehub_signup_clean.go` | Rename `iFillInWithUniquePhoneNumber`, use new generator |
| `e2e/godog_test.go` | Remove `prerequisite()`, simplify imports |
| `local/Makefile` | Add `mockxago.interledger.test` to HOSTS |
| `local/wallet.yaml` | Add `GATEHUB_WIDGET_BASE_URL`, verify Xago env vars |
| `local/Makefile` | Add `mockxago.interledger.test` to HOSTS |

## Testing Strategy

### Before Proceeding to Phase 8
```bash
cd e2e

# 1. Xago signup passes
go test -v -timeout 15m -run TestFeatures -args -tags="@xago&&@signup"

# 2. GateHub signup still passes (regression)
go test -v -timeout 15m -run TestFeatures -args -tags="@gatehub&&@signup"

# 3. All signup scenarios pass
go test -v -timeout 15m -run TestFeatures -args -tags="@signup"

# 4. No compilation errors
go build ./...
```

## Success Criteria

Phase 7 is **COMPLETE** when:

1. ✅ `mockxago.interledger.test` resolves locally (via `make hosts`)
2. ✅ `e2e/features/000-signup.feature` restructured with `@xago` South-African scenario and `@gatehub` tags on existing scenarios
3. ✅ All feature files (001-004) have `mockxago` in Background and `@gatehub` tags on existing scenarios
4. ✅ `mockxago is running at` step registered and working
5. ✅ Phone generation is country-aware (ZA → `+2771XXXXXXX` valid SA mobile format)
6. ✅ `godog_test.go` simplified (no GateHub-specific prerequisite)
7. ✅ No changes to `go/backend/` source (unless ZA `Supported` flag is proven necessary)
8. ✅ No changes to `typescript/protea/`
9. ✅ `@xago&&@signup` scenarios pass
10. ✅ `@gatehub&&@signup` scenarios still pass (no regressions)

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Removing `prerequisite()` breaks GateHub | Test GateHub scenarios after removal; if broken, make it tag-conditional |
| ZA not in `Supported` countries | Verify `IsSupported()` usage; add flag only if proven necessary (minimal backend touch) |
| SA phone validation rejects generated numbers | Format already validated on `bosbaber/mockxago`; port `getPhoneBaseForCountry` exactly |
| `wallet.yaml` env var changes break backend startup | Only add vars, don't remove existing ones unless confirmed safe |
| Phone generation regression for Germany | Run both `@gatehub` and `@xago` signup tests |
| Temptation to touch `go/backend/` | Resist — `cli.go` renames and new fields are for later phases; signup doesn't need them |

## Implementation Tips

- **Start by reading the `bosbaber/mockxago` diffs**, not writing from scratch. Most code is already written there.
- The `generateDeterministicPhone` function in `gatehub_forms.go` is a **direct extract** of the old inline code from `iFillInWithUniqueGermanPhoneNumber`, plus country dispatch. The SA phone format (`+2771XXXXXXX`) is already validated there — port it exactly.
- The `godog_test.go` changes (removing `prerequisite`) are the riskiest change — test GateHub scenarios immediately after.
- **Do NOT port** `xago_deposit.go`, `xago_link_account.go`, `xago_withdrawal.go` — those are for later phases.
- The `gatehub_kyc.go` changes on `bosbaber/mockxago` are extensive (Persona iframe support) — defer entirely to Phase 8.
- **Do NOT touch `go/backend/` or `typescript/protea/`** unless a specific test failure proves it's necessary. The `cli.go` field renames (`XagoPublicKey` → `XagoAPIPublicKey`) and new ops account fields on `bosbaber/mockxago` are for deposit/withdrawal phases, not signup.
- Feature file restructuring should follow the `bosbaber/mockxago` pattern exactly: `@xago` scenarios first, `@gatehub` scenarios second, shared Background with both mock health checks.

---

**Phase**: 7 (Wallet E2E Integration)
**Prerequisite**: Phase 5 (MockXago Foundation), Phase 6 (Infrastructure Setup)
**Reference**: `bosbaber/mockxago` branch
**Next Phase**: Phase 8 (Xago KYC Processes)
**Last Updated**: March 11, 2026
