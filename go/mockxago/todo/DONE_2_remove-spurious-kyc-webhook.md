# Bug 2: Spurious Xago KYC webhook fires at the wrong endpoint

## Summary

After KYC iframe submission, `KYCIframeSubmit` fires **two** goroutines:

```go
go h.sendKYCWebhook(walletID)          // ← wrong: posts a GateHub-style payload
go h.sendPersonaInquiryApproved(walletID) // ← correct: posts to /webhooks/persona
```

`sendKYCWebhook` POSTs a `{"event_type": "id.verification.accepted", ...}` payload to
`WEBHOOK_URL=http://backend:8080/webhooks/xago`. The backend's Xago deposit webhook
handler (`EventWebhook`) speaks a completely different schema and has no concept of KYC
events. It checks `hook.Code != 104` (always true here) and returns 501.

## Evidence

```
mockxago:  ERROR: Failed to send KYC webhook: webhook returned status 501:
backend:   "msg":"unsupported xago webhook received",
           "webhook":"{\"event_type\":\"id.verification.accepted\",
                       \"wallet_id\":\"local-inquiry-id\",...}"
```

## Sequence diagram (current broken flow)

```mermaid
sequenceDiagram
    participant UI as browser (KYC iframe)
    participant MX as mockxago
    participant B  as backend

    UI->>MX: POST /kyc/submit {user_id, first_name, ...}
    MX->>MX: save sub-account
    par Correct path
        MX->>B: POST /webhooks/persona  (Persona inquiry.approved payload)
        B->>B: handle persona webhook → trigger Xago CreateSubAccount workflow
    and Wrong path
        MX->>B: POST /webhooks/xago  {event_type: "id.verification.accepted", wallet_id: "..."}
        note over B: EventWebhook checks hook.Code != 104 → true
        B-->>MX: 501 Not Implemented
        MX->>MX: log "Failed to send KYC webhook"
    end
```

## Sequence diagram (desired flow after fix)

```mermaid
sequenceDiagram
    participant UI as browser (KYC iframe)
    participant MX as mockxago
    participant B  as backend

    UI->>MX: POST /kyc/submit {user_id, first_name, ...}
    MX->>MX: save sub-account
    MX->>B: POST /webhooks/persona  (Persona inquiry.approved payload)
    B->>B: handle persona webhook → trigger Xago CreateSubAccount workflow
```

## Fix

In `internal/handler/kyc.go`, remove the spurious call inside `KYCIframeSubmit`:

```go
// Before
go h.sendKYCWebhook(walletID)         // ← delete this line
go h.sendPersonaInquiryApproved(walletID)

// After
go h.sendPersonaInquiryApproved(walletID)
```

Delete (or keep private and unexported) the `sendKYCWebhook` method entirely — there is
no matching handler on the backend side and the `WEBHOOK_URL` env var is unambiguously
only for deposit webhooks.

The same spurious call exists inside `PersonaInquirySubmit` in `internal/handler/persona.go`:

```go
go h.sendKYCWebhook(inquiryID)  // ← also remove
```
