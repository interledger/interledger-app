# Bug 3: Persona iframe submits `"local-inquiry-id"` instead of the wallet UUID

## Summary

In local mode the backend hardcodes the Persona inquiry ID to `"local-inquiry-id"`:

```go
// go/backend/kyc/ops/persona.go
if env.IsLocal() {
    return &kyc.PersonaInquiry{ID: "local-inquiry-id", ...}
}
// ...
if env.IsLocal() {
    return fmt.Sprintf(urlFormat, "local-inquiry-id"), nil
}
```

The frontend therefore opens the mockxago iframe at:
```
GET /v1/inquiries/local-inquiry-id/iframe
```

`PersonaGetInquiryIframe` passes `inquiryID` (the path param `"local-inquiry-id"`) directly
as `UserID` to the HTML template. The form submits `user_id=local-inquiry-id`.

The Persona webhook (`sendPersonaInquiryApproved`) fires with `"id": "local-inquiry-id"`.
The backend's webhook handler tries to query the DB:

```sql
-- wallet_id is a UUID column
SELECT ... FROM kyc_persona_inquiries WHERE wallet_id = 'local-inquiry-id'
```

Postgres rejects it:

```
pq: invalid input syntax for type uuid: "local-inquiry-id"
```

The KYC approval therefore never propagates; the `CreateSubAccount` Temporal workflow
never gets signalled.

## Sequence diagram (current broken flow)

```mermaid
sequenceDiagram
    participant B   as backend
    participant MX  as mockxago
    participant UI  as browser

    B->>B: GetPersonaInquiry (env.IsLocal()) → ID = "local-inquiry-id"
    B-->>UI: persona widget config {id: "local-inquiry-id"}
    UI->>MX: GET /v1/inquiries/local-inquiry-id/iframe
    note over MX: inquiryID = "local-inquiry-id", UserID = "local-inquiry-id"
    MX-->>UI: KYC form (user_id hidden = "local-inquiry-id")
    UI->>MX: POST /v1/inquiries/local-inquiry-id/submit {user_id="local-inquiry-id", ...}
    MX->>B: POST /webhooks/persona {"data":{"id":"local-inquiry-id",...}}
    B->>B: DB lookup WHERE wallet_id='local-inquiry-id'
    note over B: pq: invalid input syntax for type uuid
    B-->>MX: 500
    MX->>MX: ERROR: Persona webhook failed with status 500
```

## Sequence diagram (desired flow after fix)

```mermaid
sequenceDiagram
    participant B   as backend
    participant MX  as mockxago
    participant UI  as browser

    B->>B: GetPersonaInquiry (env.IsLocal()) → ID = walletID (real UUID)
    B-->>UI: persona widget config {id: walletID}
    UI->>MX: GET /v1/inquiries/{walletID}/iframe
    MX-->>UI: KYC form (user_id hidden = walletID)
    UI->>MX: POST /v1/inquiries/{walletID}/submit {user_id=walletID, ...}
    MX->>B: POST /webhooks/persona {"data":{"id":walletID,...}}
    B->>B: DB lookup WHERE wallet_id=walletID ✓
    B->>B: trigger CreateSubAccount workflow
```

## Root cause location

The literal string `"local-inquiry-id"` appears in two separate places in
`go/backend/kyc/ops/persona.go`. These are the canonical fix points; mockxago is a
consequence.

**Option A — fix the backend (preferred, authoritative):**

Replace the hardcoded constant with the real walletID in both local branches:

```go
// GetPersonaInquiry — local branch
if env.IsLocal() {
    err := GenerateKycData(ctx, b, walletID)
    return &kyc.PersonaInquiry{
        ID:     walletID,   // ← was "local-inquiry-id"
        Status: persona.InquiryStatus(persona.InquiryApproved),
    }, err
}

// GetApprovedPersonaInquiryURL — local branch
if env.IsLocal() {
    return fmt.Sprintf(urlFormat, walletID), nil   // ← was "local-inquiry-id"
}
```

With this change the Persona SDK opens the iframe at `/v1/inquiries/{walletID}/iframe`.
`PersonaGetInquiryIframe` already passes the URL path param as `UserID`, so the form
correctly submits the wallet UUID — no changes needed on the mockxago side for this bug.

**Option B — fix in mockxago only (workaround):**

Pass the wallet UUID as a separate query param (e.g. `?reference_id=<walletID>`) when
opening the iframe URL, and have `PersonaGetInquiryIframe` prefer that over the inquiry
ID. This is more fragile since it requires a coordinated change between the frontend and
mockxago.

Option A is strongly preferred because the backend is authoritative (per the brief).
