# Phase 8 Implementation Plan: Xago KYC Processes

## Overview

Phase 8 integrates Xago KYC (Know Your Customer) workflows into the E2E test suite. The key insight from analyzing `bosbaber/mockxago` is that this requires changes across **four layers**: the MockXago service (Persona API routes and a Persona-compatible SDK script), the Go backend (remove `env.IsLocal()` bypasses, route Persona client to MockXago), the Protea frontend (minimal config seam for Persona SDK URL), and the E2E tests (new scenario + step implementation).

**Phase Outcome**: E2E feature tests for Xago KYC processes pass, with South Africa users properly routed through MockXago's mocked Persona integration.

**Key Correction from Analysis (March 13, 2026)**: The original plan assumed this was E2E-only work. In reality, the `bosbaber/mockxago` branch shows that the KYC flow requires:
1. **MockXago service** must expose Persona SDK-compatible API endpoints (`/v1/inquiries/*`) — these routes exist in `persona.go` handler but are **NOT registered** in the current `main.go`
2. **Backend** `kyc/persona/client.go` must route to MockXago when `PERSONA_TOKEN` is empty and `MOCKXAGO_ENDPOINT` is set
3. **Backend** `kyc/ops/persona.go` and `grpc/kyc.go` must remove `env.IsLocal()` bypasses that short-circuit the real flow
4. **Protea** `personal-details.tsx` should keep standard Persona flow and load SDK from configurable `PERSONA_SDK_URL`
5. **Docker config** must pass `MOCKXAGO_ENDPOINT` to backend and `PERSONA_SDK_URL` to protea
6. **E2E tests** need the new scenario and step implementations

## Analysis: Current State vs bosbaber/mockxago (March 13, 2026)

### Feature File Differences (`e2e/features/001-kyc.feature`)

| Aspect | Current (`stephan/wal-599`) | `bosbaber/mockxago` |
|--------|----------------------------|---------------------|
| Background | Missing `mockxago is running at` line | Includes MockXago service check |
| GateHub tag | `@kyc @gatehub` | `@kyc @germany` |
| GateHub scenario name | "Successfully activate account and complete KYC as verified user" | "Successfully complete KYC as a verified user in Germany" |
| Xago scenario | **Missing** | Present: `@kyc @xago` "Successfully complete KYC as a verified user in South Africa" |
| Xago KYC steps | N/A | Uses `I fill and submit the mockxago KYC iframe` step |

### Backend Changes Required (from `bosbaber/mockxago` diff)

#### `go/backend/kyc/persona/client.go` — Route Persona API to MockXago
The `New()` constructor must detect local dev mode and redirect API calls to MockXago:
```go
baseURL := "https://api.withpersona.com/api/v1/"
if os.Getenv("PERSONA_TOKEN") == "" && os.Getenv("MOCKXAGO_ENDPOINT") != "" {
    baseURL = os.Getenv("MOCKXAGO_ENDPOINT") + "/v1/"
}
```
This is the **critical wiring** — it makes the backend's Persona client talk to MockXago instead of real Persona. Without this, `GetPersonaInquiry` will either hit the old `env.IsLocal()` bypass (returning a fake ID) or fail against the real Persona API.

#### `go/backend/kyc/ops/persona.go` — Remove local bypasses
Three `env.IsLocal()` / `!env.IsProd()` blocks must be removed:
- `GetPersonaInquiry()`: Remove early return with `"local-inquiry-id"` (line 21) — let real flow execute against MockXago
- `GetApprovedPersonaInquiryURL()`: Remove local bypass (line 117) — let DB lookup run
- `GetPersonaIDNumbers()`: Remove fake ID number generation (line 138) — let real inquiry flow return data
- `GetZAIDNumber()`: Remove fake ID bypass (line 186) — same reason
- **Delete** helper functions: `generateLocalIDNumbers()`, `generateLocalTestIdNumber()` — no longer needed

#### `go/backend/grpc/kyc.go` — Remove local provider bypass
Remove `env.IsLocal()` block at line 99 that returns `Provider: "local"` instead of `Provider: "persona"`. This bypass prevented the frontend from rendering the Persona/MockXago iframe.

### Frontend Changes Required (`typescript/protea/app/routes/personal-details.tsx`)

The `PersonaPage` component should keep the standard Persona flow and only add a minimal configuration seam:
1. **Read** `personaSdkUrl` from loader data (passed via `process.env.PERSONA_SDK_URL`)
2. **Load SDK** via `useScript(personaSdkUrl)` instead of hard-coding Persona CDN URL
3. **Keep existing flow**: `new Persona.Client({...})` + `KycIntro` + `Continue` button

The loader should pass the `personaSdkUrl` value to the client:
```typescript
return json({
  ...existingData,
  personaSdkUrl:
    process.env.PERSONA_SDK_URL ||
    'https://cdn.withpersona.com/dist/persona-v4.8.0-alpha.js'
})
```

### MockXago Service Changes (`go/mock/mockxago`)

#### Route Registration (`cmd/mockxago/main.go`)
Persona SDK-compatible routes must be registered. The handler functions (`persona.go`) already exist and are tested, but the routes are **NOT wired up** in `main.go`. Add after the KYC endpoints:
```go
// Persona SDK compatible endpoints - no auth required
router.Post("/v1/inquiries", h.PersonaCreateInquiry)
router.Get("/v1/inquiries/{inquiryId}", h.PersonaGetInquiry)
router.Get("/v1/inquiries/{inquiryId}/iframe", h.PersonaGetInquiryIframe)
router.Post("/v1/inquiries/{inquiryId}/submit", h.PersonaInquirySubmit)
router.Get("/v1/accounts/{accountId}", h.PersonaGetAccount)
router.Post("/v1/accounts/{accountId}/remove-tag", h.PersonaRemoveTag)
```

**Important**: These routes must go **outside** the authenticated `/v1` group since the Persona client and the iframe use bearer token auth differently (or no auth for iframe). In `bosbaber/mockxago`, these are registered at the top level after the main `/v1` route group.

#### Existing Assets (already present, no changes needed)
- `internal/handler/persona.go` — All Persona handler functions ✅ (only import path differs)
- `internal/handler/persona_test.go` — Tests ✅  
- `web/kyc-iframe.html` — iframe HTML template ✅

#### Additional Asset Needed for Option 2
- `web/persona-sdk.js` — Persona-compatible mock SDK served by MockXago at `/v1/persona-sdk.js`

### Docker Configuration Changes

#### `local/wallet.yaml` — Add `MOCKXAGO_ENDPOINT` to backend
```yaml
- MOCKXAGO_ENDPOINT=http://mockxago:8080
```
Currently missing. Without this, `persona/client.go`'s `New()` won't redirect to MockXago.

#### `local/protea.yaml` — Add `PERSONA_SDK_URL` to protea
```yaml
- PERSONA_SDK_URL=https://mockxago.interledger.test/v1/persona-sdk.js
```
Currently missing. The protea frontend needs to load a Persona-compatible SDK from MockXago (browser-accessible URL via Traefik).

### E2E Test Changes

#### `e2e/features/001-kyc.feature`
- Add `mockxago is running at` to Background
- Rename GateHub scenario tag from `@gatehub` to `@germany`
- Add Xago KYC scenario with `@kyc @xago` tags

#### `e2e/gatehub_kyc.go` (+318 lines in bosbaber/mockxago)
Key additions:
- `iFillAndSubmitTheMockxagoiframe()` — Interacts with Persona iframe from MockXago: finds form fields (first_name, last_name, dob, address, city, country), fills them with test data, clicks submit
- `submitMockXagoKYCForm()` — Fallback: direct HTTP POST to `https://mockxago.interledger.test/kyc/submit` if iframe form elements aren't found
- Provider detection in `iCompletedTheKYCFlowFor()` — `useMockXago := strings.EqualFold(sc.country, "South Africa")`
- Enhanced `iShouldSeeTheActivateWalletButton()` — Now also checks for iframe presence (MockXago renders iframe directly)
- Enhanced `iWaitForTheKYCCompletion()` — Polls `getKYCStatusByWalletID()` every 10 iterations as fallback
- Various debug logging improvements

#### `e2e/context.go`
- Register new step: `I fill and submit the mockxago KYC iframe`
- Minor: `waitForHealthEndpoint` becomes a method on `E2EContext` (from standalone function)

#### `e2e/db_helpers.go` (+121 lines)
New DB helper functions needed for KYC status polling:
- `getWalletIDByEmail()` — Looks up wallet ID via kratos user ID
- `getKYCStatusByWalletID()` — Reads `wallet_kyc_status` table
- `getXagoAccountIDByWalletID()` — Reads `xago_sub_accounts` table (used more in later phases)
- `xagoLinkedAccountExists()` — Checks `linked_accounts` for xago provider (used more in later phases)

#### `e2e/godog_test.go` (removals only)
- Removes `prerequisite()` function (GateHub org config workflow) — this was GateHub-specific setup that blocked Xago tests
- Removes `waitForWorkers()` and Temporal client imports
- Simplifies `TestFeatures` and `TestMain`

## Implementation Tasks (Revised)

### Task 1: Register Persona Routes in MockXago (15 min)
**File**: `go/mock/mockxago/cmd/mockxago/main.go`

Add 6 Persona SDK-compatible routes after the existing KYC endpoints. The handler functions already exist in `persona.go`. This is purely route registration.

### Task 2: Backend Persona Client → MockXago Routing (30 min)
**Files**:
- `go/backend/kyc/persona/client.go` — Add `MOCKXAGO_ENDPOINT` env var detection in `New()`
- `go/backend/kyc/ops/persona.go` — Remove all `env.IsLocal()` and `!env.IsProd()` bypasses, delete `generateLocalIDNumbers()` and `generateLocalTestIdNumber()` helper functions
- `go/backend/grpc/kyc.go` — Remove `env.IsLocal()` block that returns `Provider: "local"`

### Task 3: Protea Persona SDK URL Seam (45 min)
**File**: `typescript/protea/app/routes/personal-details.tsx`

- Add `personaSdkUrl` to loader return data
- Use `useScript(personaSdkUrl)` in `PersonaPage`
- Keep standard `Persona.Client` + `KycIntro` flow (no provider-specific iframe branch)

### Task 4: Docker Configuration (10 min)
**Files**:
- `local/wallet.yaml` — Add `MOCKXAGO_ENDPOINT=http://mockxago:8080`
- `local/protea.yaml` — Add `PERSONA_SDK_URL=https://mockxago.interledger.test/v1/persona-sdk.js`

### Task 5: E2E Feature File Update (15 min)
**File**: `e2e/features/001-kyc.feature`

- Add `mockxago is running at` to Background
- Rename `@gatehub` tag to `@germany`  
- Add `@kyc @xago` scenario for South Africa

### Task 6: E2E Step Implementation (1-1.5 hours)
**Files**:
- `e2e/gatehub_kyc.go` — Add `iFillAndSubmitTheMockxagoiframe()`, `submitMockXagoKYCForm()`, provider detection in `iCompletedTheKYCFlowFor()`, enhanced waiters
- `e2e/context.go` — Register `I fill and submit the mockxago KYC iframe` step
- `e2e/db_helpers.go` — Add `getWalletIDByEmail()`, `getKYCStatusByWalletID()` (and optionally xago-specific helpers for later phases)

### Task 7: Simplify godog_test.go (15 min)
**File**: `e2e/godog_test.go`

Remove the `prerequisite()` function that runs `UpdateGateHubOrganizationConfig` Temporal workflow before tests. This blocks Xago-only test runs. Also remove `waitForWorkers()` and related Temporal imports.

### Task 8: Run & Debug (1-2 hours)

1. Rebuild local environment: `cd local && make down && make all`
2. Run Xago KYC tests: `cd e2e && go test -v -timeout 20m -args -tags="@xago&&@kyc"`
3. Also run GateHub KYC to verify no regression: `go test -v -timeout 20m -args -tags="@germany&&@kyc"`
4. Debug any issues with screenshots, logs, and DB queries

## How the KYC Flow Works End-to-End (from bosbaber/mockxago)

```
User clicks "Activate wallet" in Protea
  → Protea loader calls grpc.getKYCProviderWidget()
  → Backend GetKYCProviderWidget() calls GetPersonaInquiry()
    → persona/client.go New() sees MOCKXAGO_ENDPOINT, sets baseURL to http://mockxago:8080/v1/
    → GetPersonaInquiry() calls CreateInquiry() → POST http://mockxago:8080/v1/inquiries
    → MockXago creates inquiry record, returns inquiry ID
  → Backend returns Provider: "persona", PersonaInquiry: {ID: "..."}
  → Protea loads Persona SDK from PERSONA_SDK_URL (local: https://mockxago.interledger.test/v1/persona-sdk.js)
    → MockXago `persona-sdk.js` creates Persona-compatible modal and iframe
    → User clicks Continue (standard KycIntro flow)
    → SDK opens iframe at /v1/inquiries/{id}/iframe
    → MockXago serves kyc-iframe.html template with form fields
    → User fills form, clicks submit → POST /v1/inquiries/{id}/submit
    → MockXago marks inquiry as "approved", sends postMessage to SDK parent window
    → SDK triggers Persona `onComplete` callback
    → MockXago sends webhook to backend: id.verification.accepted
  → Protea `onComplete` submits form POST to /personal-details
  → Backend processes webhook, updates kyc_status in DB
  → User redirected to dashboard with approved status
```

## Success Criteria

Phase 8 is **COMPLETE** when:

1. ✅ MockXago Persona routes registered and serving iframe at `/v1/inquiries/{id}/iframe`
2. ✅ Backend persona client routes to MockXago when `MOCKXAGO_ENDPOINT` is set
3. ✅ Backend `env.IsLocal()` KYC bypasses removed from `persona.go` and `grpc/kyc.go`
4. ✅ Protea uses configurable `PERSONA_SDK_URL` and keeps standard Persona UX flow
5. ✅ Docker config passes `MOCKXAGO_ENDPOINT` to backend and `PERSONA_SDK_URL` to protea
6. ✅ `e2e/features/001-kyc.feature` contains `@xago` tagged KYC scenario
7. ✅ `e2e/gatehub_kyc.go` has `iFillAndSubmitTheMockxagoiframe()` and provider detection
8. ✅ All `@xago&&@kyc` scenarios pass: `go test ./e2e -v -args -tags="@xago&&@kyc"`
9. ✅ GateHub KYC scenarios still pass (no regression): `go test ./e2e -v -args -tags="@germany&&@kyc"`
10. ✅ No new standalone `xago_kyc.go` file created

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Removing `env.IsLocal()` breaks other local flows | MockXago Persona endpoints replicate the exact same behavior; backend tests should still pass with test migrations |
| Persona client routing wrong | The conditional is simple: no `PERSONA_TOKEN` + has `MOCKXAGO_ENDPOINT` = use MockXago |
| Mock Persona SDK script doesn't load | Ensure MockXago serves `/v1/persona-sdk.js` and `PERSONA_SDK_URL` points to Traefik URL |
| postMessage not received | MockXago iframe template already includes postMessage calls; debug with browser console |
| godog prerequisite removal breaks GateHub tests | The `UpdateGateHubOrganizationConfig` workflow should be moved to scenario-level setup if needed, not test-suite prereq |

## Dependency on Previous Phases

- **Phase 5**: MockXago foundation (mock service, handler code, KYC endpoints)
- **Phase 7**: Xago signup (Xago accounts created for South Africa users)
- **Local environment**: MockXago, backend, and protea running

---

## Post-Implementation Notes (March 17, 2026)

Phase 8 was implemented and all success criteria are met. The following reviewer comments were addressed during PR review:

| Thread | Finding | Resolution |
|--------|---------|------------|
| Thread 1: persona-sdk.js postMessage origin | Origin validation already present (`event.origin !== expectedOrigin`) — no change needed | Already correct |
| Thread 2: Missing PersonaSDK handler test | Added `TestPersonaSDK_ServesScript` to `persona_test.go` | Fixed |
| Thread 3: MOCKXAGO_ENDPOINT fallback masking config errors | Implementation uses dedicated `PERSONA_BASE_URL` env var (not the plan's MOCKXAGO_ENDPOINT+PERSONA_TOKEN approach) — already the cleaner design | Already correct |
| Thread 4: Typo "fro" → "for" in `kyc/ops/persona.go` comment | Fixed | Fixed |
| Thread 5: Fill/Click errors silently ignored in `gatehub_kyc.go` | Errors now propagated from Fill/Click calls in both KYC iframe helpers | Fixed |
| Thread 6: `PersonaInquirySubmit` treats all `GetSubAccountByWalletID` errors as not-found | Now branches on `storage.ErrSubAccountNotFound`; other errors return 500 | Fixed |
| Thread 7: `PersonaGetAccount` returns 400 `missing_dob` for missing accounts | Already returns 404 `account_not_found` for missing accounts — no change needed | Already correct |

**E2E test results** (post-fix):
- `@xago&&@kyc`: 1 scenario, 20 steps — PASS (39s)
- `@germany&&@kyc`: 1 scenario, 20 steps — PASS (46s)

---

**Phase**: 8 (Xago KYC Processes)
**Status**: COMPLETE
**Prerequisite**: Phase 7 (Xago Signup)
**Next Phase**: Phase 9 (Xago Deposits)
**Last Updated**: March 17, 2026
