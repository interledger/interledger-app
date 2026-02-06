# Email Verification Flow Investigation

**Date**: 3 February 2026

## Executive Summary

The email verification flow differs **significantly** between production (clicking email link) and the e2e test (direct database update). The production flow includes Kratos redirect chains and potentially missing backend webhook handlers, but **no wallet-related operations occur during email verification itself**. Wallet auto-provisioning happens separately via gRPC middleware.

---

## 1. Kratos Email Verification Endpoint

### Production Flow URL

When a user clicks the verification link from their email in production, they are directed to:

```
https://interledger.app/verify?flow=<flow_id>&code=<verification_code>
```

This is a **self-service verification flow** in Kratos that leads to:

```
POST /self-service/verification/flows?id=<flow_id>  [Kratos Public API]
```

### Kratos Configuration

From [local/config/kratos.yml](local/config/kratos.yml):

```yaml
verification:
  after:
    default_browser_return_url: http://interledger.test
  enabled: true
  use: link  # Email link verification (not TOTP or SMS)
  lifespan: 24h
  ui_url: http://interledger.test/verify  # Protea UI endpoint
```

**Key points**:
- **No hooks configured for verification completion** (unlike registration which has `password.hooks.session`)
- Verification is done via email `link` method only
- After verification completes, Kratos redirects to `http://interledger.test` (or the configured `default_browser_return_url`)
- **No backend webhook is triggered** by Kratos after email verification

---

## 2. What Happens in Production When User Clicks Verification Link

### Step-by-Step Production Flow

```
1. User receives email with link:
   https://interledger.app/verify?flow=<flow_id>&code=<verification_code>

2. User clicks link → Browser redirects to Protea /verify page
   [typescript/protea/app/routes/verify.tsx]

3. verify.tsx loader() executes:
   - Checks if user already verified (line 43-45)
   - If NOT verified:
     a) Fetches verification flow from Kratos:
        GET /self-service/verification/flows?id=<flow_id>
     b) Returns flow and email to UI

4. User sees "Verify your email" page with two options:
   - Click embedded "Verify" button (submits form)
   - Click "Resend verification" button

5. User clicks "Verify" button → verify.tsx action() executes:
   POST /self-service/verification?flow=<flow_id>
   Body: {
     method: 'link',
     email: '<email>',
     csrf_token: '<token>'
   }

6. Kratos processes verification:
   - Validates the link/code
   - Updates identity_verifiable_addresses.verified = true
   - Updates identity_verifiable_addresses.verified_at = NOW()
   - Returns 200 OK or redirect

7. Browser receives response:
   - If successful, Kratos redirects to http://interledger.test
   - verify.tsx returns success to frontend

8. **NO backend webhook is called**
9. **NO wallet operations occur**
```

### Key Code Location

[typescript/protea/app/routes/verify.tsx](typescript/protea/app/routes/verify.tsx) lines 175-216:

```tsx
export async function action({ request }: ActionFunctionArgs) {
  const verificationResponse = await fetch(
    `${KRATOS_URL}/self-service/verification?flow=${flowId}`,
    {
      method: 'POST',
      redirect: 'manual',
      body: JSON.stringify({
        method: 'link',
        email,
        csrf_token: csrfToken
      }),
      headers: {
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )

  if (verificationResponse.status >= 400) {
    throw new Error('Could not send verification email')
  }

  return json<ActionData>({ success: true }, { status: 200 })
}
```

---

## 3. Backend Code Triggered by Email Verification

### What the Backend Does

**⚠️ IMPORTANT**: There is **NO backend webhook handler for email verification** in the current codebase.

From [go/backend/main.go](go/backend/main.go) lines 185-198:

```go
router.Handle("/kratos/signup", analytics_webhook.NewHandleSignup(b))     // ✓ Exists
router.Handle("/kratos/login", analytics_webhook.NewHandleLogin(b))       // ✓ Exists
router.Handle("/kratos/logout", analytics_webhook.NewHandleLogout(b))     // ✓ Exists
// router.Handle("/kratos/verification", ...)  ❌ NO HANDLER
```

The backend has webhooks for signup/login/logout events, **but NOT for email verification**.

### What Email Verification Actually Does

1. **Kratos-side only**: Updates `identity_verifiable_addresses` table
2. **No async operations triggered**
3. **No backend notifications**
4. Backend remains unaware that email was verified

The verification status is only checked when:
- User attempts to access features requiring verified email
- Backend queries `verifiable_addresses[0].verified` from Kratos session

---

## 4. Wallet-Related Operations During Email Verification

### Key Finding: **Wallets Are NOT Created During Email Verification**

Wallets are auto-provisioned separately via **gRPC middleware**, NOT by email verification:

From [go/backend/wallets/middleware/middleware.go](go/backend/wallets/middleware/middleware.go):

```go
// Create a default wallet for the user if they don't already have one
if len(walletList) == 0 {
  _, err = wc.Create(ctx, wallets.CreateArgs{
    UserID:  u.ID,
    Country: u.Country,
  })
  if err != nil && !errors.Is(err, wallets.ErrDuplicateWallet) {
    log.Warn("failed to create default wallet for user", ...)
  }
}
```

**When is this middleware called?**
- Every gRPC request to the backend
- After user is authenticated (has valid session/cookies)
- **NOT tied to email verification status**

**Timeline**:
1. User signs up → Kratos creates identity
2. Middleware runs on first gRPC call → Creates "default" wallet
3. User verifies email → **No wallet operations**
4. User already has "default" wallet from step 2

---

## 5. Test Flow vs Production Flow Comparison

### E2E Test Flow (Current Implementation)

[local/e2e-playwright/signup_assertions.go](local/e2e-playwright/signup_assertions.go) lines 155-220:

```go
func (sc *SignupContext) iTriggerUserVerificationFor(email string) error {
  // Step 1: Check if Kratos identity exists
  // Step 2: If not found, FAIL
  // Step 3: Direct database update to mark email as verified:
  
  kratosDB.Exec(`
    UPDATE identity_verifiable_addresses 
    SET verified = TRUE, verified_at = NOW()
    WHERE value = $1
  `, prefixedEmail)
  
  return nil
}
```

### Key Differences

| Aspect | Production | E2E Test |
|--------|-----------|----------|
| **Kratos endpoint called** | `/self-service/verification/flows` + `/self-service/verification` | ❌ None |
| **Kratos validation** | Full flow validation, code verification | ❌ Skipped |
| **Database update** | Via Kratos internal logic | Direct SQL query |
| **Email link** | User clicks link | ❌ Bypassed |
| **Redirect chains** | `/verify?flow=...` → Kratos → redirect back | ❌ None |
| **Backend webhook** | ❌ Not called (no handler exists) | ❌ Not called |
| **Wallet operations** | ❌ None (happened at signup) | ❌ None (happened at signup) |
| **CSRF token** | Validated | ❌ Skipped |

---

## 6. Missing Backend Webhook Handler

### Current State

Kratos has **no webhook hook configured** for email verification completion in [local/config/kratos.yml](local/config/kratos.yml):

```yaml
verification:
  after:
    default_browser_return_url: http://interledger.test
    # ❌ No hooks configured (like registration has)
    # hooks:
    #   - hook: web
    #     config:
    #       url: http://backend:8080/kratos/verification
```

### Why This Matters

1. **Backend is unaware** when users verify their email
2. **No async jobs triggered** (e.g., send verification complete notification)
3. **No wallet state changes** (not needed, already created)
4. **No analytics events** (signup/login tracked, but not verification)

### To Enable Backend Webhooks for Verification

Would need:
1. Add hook to Kratos config
2. Create backend handler: `/kratos/verification`
3. Implement async operations if needed

---

## 7. Why Duplicate Wallets Aren't Created

The investigation notes were concerned about duplicate wallet creation. Here's why it doesn't happen:

### Timeline

```
1. [Signup] User registers
   → Kratos creates identity
   → User has session cookie
   
2. [First gRPC call] Middleware runs
   → Checks wallet list: empty
   → Creates "default" wallet
   → Adds wallet to context
   
3. [Email Verification] User verifies email
   → Only updates identity_verifiable_addresses
   → Middleware still runs on next gRPC call
   → Checks wallet list: already has 1 wallet
   → Skips creation (if len(walletList) == 0)
   
4. [No duplicates] Because:
   - Wallet creation is idempotent (ErrDuplicateWallet handled)
   - Only created if list is empty
   - Email verification doesn't trigger wallet creation
```

---

## 8. Comparison with Gatehub MockGatehub KYC Flow

For reference, the parallel KYC verification flow (mockgatehub) works differently:

1. `POST /id/v1/users/{userID}/hubs/{gatewayID}` – Initiates KYC
2. `GET /?paymentType=onboarding&bearer={token}` – Serves iframe
3. `POST /iframe/submit` – Form submission
4. **Webhook `id.verification.accepted` emitted asynchronously**

This is **separate from email verification** and is handled by mockgatehub webhooks.

---

## 9. Flow Diagram: Production vs Test

### Production Flow (Email Link)

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Email verification link                                   │
│    https://interledger.app/verify?flow=<id>&code=<code>      │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│ 2. Browser requests /verify?flow=<id>                        │
│    [typescript/protea/app/routes/verify.tsx]                 │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│ 3. Protea UI checks Kratos session                           │
│    GET /self-service/verification/flows?id=<id>             │
│    (via loader)                                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│ 4. User clicks "Verify" button on UI                         │
│    POST /self-service/verification?flow=<id>                │
│    Body: {method: 'link', email, csrf_token}                │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│ 5. Kratos processes verification                             │
│    - Validates code/signature                               │
│    - Updates identity_verifiable_addresses                  │
│    - Returns 200 OK or redirect                             │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│ 6. Browser redirected to http://interledger.test             │
│    User is verified, can access features                    │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ❌ NO BACKEND WEBHOOK
        ❌ NO WALLET OPERATIONS
```

### Test Flow (Direct Database Update)

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Test calls iTriggerUserVerificationFor(email)             │
│    [local/e2e-playwright/signup_assertions.go]               │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│ 2. Direct Kratos database query                              │
│    UPDATE identity_verifiable_addresses                      │
│    SET verified=TRUE, verified_at=NOW()                      │
│    WHERE value=<email>                                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│ 3. Email marked as verified                                  │
│    User can access verified email features                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ❌ NO KRATOS API CALLS
        ❌ NO CSRF VALIDATION
        ❌ NO KRATOS FLOW VALIDATION
        ❌ NO BACKEND WEBHOOK
        ❌ NO WALLET OPERATIONS
```

---

## 10. Key Findings & Recommendations

### Findings

1. **✓ Production email verification is complete** – Kratos handles it fully
2. **✗ No backend webhook for email verification** – No handler in backend
3. **✗ No wallet operations during email verification** – Wallets created at signup via middleware
4. **✗ E2E test bypasses Kratos flow** – Direct database update, not realistic
5. **✓ No duplicate wallet risk** – Wallet middleware is idempotent

### Potential Issues with Current Test Flow

The test flow misses:
- Kratos flow validation
- CSRF token verification
- Email code verification
- Any async operations that might be configured

### Recommendations

1. **Add Kratos verification webhook** (if needed for backend operations)
   - Uncomment/configure in [local/config/kratos.yml](local/config/kratos.yml)
   - Create handler in backend (e.g., `/kratos/verification`)

2. **Make E2E test more realistic** (optional)
   - Instead of direct DB update, could:
     - Call Kratos `/self-service/verification` endpoint
     - Extract verification code from Kratos logs
     - Simulate user clicking email link

3. **Document the separation**
   - Email verification = Kratos responsibility
   - Wallet creation = gRPC middleware responsibility
   - They are independent operations

---

## 11. Kratos Endpoint Summary

| Endpoint | Method | Purpose | Handler | Called By |
|----------|--------|---------|---------|-----------|
| `/self-service/verification/browser` | GET | Initialize verification flow | Kratos | Protea loader |
| `/self-service/verification/flows?id=<id>` | GET | Fetch verification flow details | Kratos | Protea loader |
| `/self-service/verification?flow=<id>` | POST | Submit verification (email or code) | Kratos | Protea action |
| `/kratos/verification` | POST | Webhook notification (if configured) | Backend | Kratos (if hooked) |

---

## 12. Test Data Flow

When running `I trigger user verification for "email@example.com"`:

```
1. Prefix email: "test-<random>-email@example.com"
2. Find Kratos identity by email (retry 60 times)
3. Connect to Kratos database directly
4. Execute SQL: UPDATE identity_verifiable_addresses SET verified=TRUE
5. User is now verified in Kratos database
6. Next session query will see verified=TRUE
```

---

## Conclusion

The email verification flow is **complete in production** but **there is no backend webhook handler**. The flow is:

1. **User clicks email link** → Kratos validation flow
2. **Kratos updates database** → User is verified
3. **No backend notification** → Backend unaware (by design, currently)
4. **Wallets already exist** → Created at signup, not verification

The E2E test **simplifies this by skipping Kratos** but achieves the same end state (verified email) directly. This is acceptable for testing but less realistic than simulating the actual user flow.

---

## Files Affected

- [typescript/protea/app/routes/verify.tsx](typescript/protea/app/routes/verify.tsx) – Frontend verification UI
- [go/backend/main.go](go/backend/main.go) – Backend router (no verification webhook)
- [go/backend/wallets/middleware/middleware.go](go/backend/wallets/middleware/middleware.go) – Wallet auto-provisioning
- [local/config/kratos.yml](local/config/kratos.yml) – Kratos configuration
- [local/e2e-playwright/signup_assertions.go](local/e2e-playwright/signup_assertions.go) – E2E test helper

