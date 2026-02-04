# Current Kratos HTTP Call Usages

This document catalogs all direct HTTP `fetch()` calls made to Kratos API endpoints in the Protea application. The goal is to replace these with the official Ory Kratos TypeScript SDK.

## Current Package Version

```json
"@ory/kratos-client": "25.4.0"
```

> [!CAUTION]
> This is an extremely outdated version (from ~2021). The latest version is **1.22.15** of `@ory/client` or matching `@ory/kratos-client` version for self-hosted Kratos.

---

## Core Library: `app/lib/kratos.server.ts`

The central Kratos utility file that exports helper functions and the `KRATOS_URL` constant.

### Exports Used:
- `KRATOS_URL` - Base URL for Kratos API (env variable)
- `getCsrfTokenFromFlow()` - Extracts CSRF token from flow UI nodes
- `handleFlowError()` - Handles Kratos flow error responses with redirects
- `getUserSession()` - Fetches current user session
- `hasUserSession()` - Checks if session cookie exists
- `requireNoUserSession()` - Ensures user is not logged in
- `kratosErrorMapping()` - Maps Kratos errors to form field errors
- `kratosErrorMessage()` - Custom error message overrides

### HTTP Calls in this file:

#### 1. `getUserSession()` - Session Check
**Purpose:** Validate user session and enforce AAL2 when required

```typescript
// Line 48
const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
  headers: request.headers
})
```

#### 2. `requireNoUserSession()` - Session Guard
**Purpose:** Verify user doesn't have an active session

```typescript
// Line 90
const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
  headers: request.headers
})
```

---

## Core Library: `app/lib/totp.server.ts`

TOTP/2FA related utilities.

### HTTP Calls:

#### 1. `isTotpAvailable()` - Check TOTP Settings Availability
**Purpose:** Check if TOTP is available in the settings flow

```typescript
// Line 38
const response = await fetch(`${KRATOS_URL}/self-service/settings/browser`, {
  headers: {
    Accept: 'application/json',
    cookie
  }
})
```

#### 2. `isTotpSet()` - Check if User Has TOTP Configured
**Purpose:** Check if user identity has TOTP credentials (uses Admin API)

```typescript
// Line 72
const response = await fetch(`${KRATOS_URL}/admin/identities/${session.identity.id}`, {
  headers: headers
})
```

---

## Route: `app/routes/login.tsx`

### HTTP Calls:

#### 1. Loader - Get Login Flow
**Purpose:** Initialize or fetch existing login flow

```typescript
// Line 47-48 - Fetch existing flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
  {
    headers: { cookie, Accept: 'application/json' }
  }
)

// Line 60-61 - Initialize new flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/login/browser${url.search}`,
  { headers: { Accept: 'application/json' } }
)
```

#### 2. Action - Submit Login
**Purpose:** Submit login credentials

```typescript
// Line 181-194
const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flowId}`, {
  method: 'POST',
  body: JSON.stringify({
    method: 'password',
    identifier: email,
    password: password,
    csrf_token: csrfToken
  }),
  headers: {
    Accept: 'application/json',
    'Content-type': 'application/json',
    cookie: String(request.headers.get('cookie'))
  }
})
```

---

## Route: `app/routes/signup/route.tsx`

### HTTP Calls:

#### 1. Loader - Initialize Registration Flow
**Purpose:** Create new registration flow

```typescript
// Line 47-49
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/registration/browser?${url.searchParams}`,
  { headers: { Accept: 'application/json' } }
)
```

#### 2. `passwordAction()` - Submit Registration
**Purpose:** Complete user registration

```typescript
// Line 281-302
const response = await fetch(
  `${KRATOS_URL}/self-service/registration?flow=${kratosFlowId}`,
  {
    method: 'POST',
    body: JSON.stringify({
      method: 'password',
      traits: { email, phone, firstName, lastName, countryCode },
      password,
      csrf_token: kratosCsrfToken
    }),
    headers: {
      'Content-type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    }
  }
)
```

---

## Route: `app/routes/logout.tsx`

### HTTP Calls:

#### 1. Loader - Create Logout Flow
**Purpose:** Initialize logout flow to get logout token

```typescript
// Line 20-25
const flowRes = await fetch(`${KRATOS_URL}/self-service/logout/browser`, {
  headers: {
    cookie,
    Accept: 'application/json'
  }
})
```

#### 2. Action - Execute Logout
**Purpose:** Perform the actual logout

```typescript
// Line 78-84
const res = await fetch(`${KRATOS_URL}/self-service/logout?token=${token}`, {
  method: 'GET',
  headers: {
    Accept: 'application/json',
    cookie
  }
})
```

---

## Route: `app/routes/recovery.tsx`

### HTTP Calls:

#### 1. Loader - Get/Initialize Recovery Flow
**Purpose:** Manage recovery flow state

```typescript
// Line 34-35 - Fetch existing flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/recovery/flows?id=${flowId}`,
  { headers: { cookie, Accept: 'application/json' } }
)

// Line 47-48 - Initialize new flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/recovery/browser?${url.searchParams}`,
  { headers: { Accept: 'application/json' } }
)
```

#### 2. Action - Submit Recovery Email
**Purpose:** Send recovery email

```typescript
// Line 155-168
const res = await fetch(
  `${KRATOS_URL}/self-service/recovery?flow=${flowId}`,
  {
    method: 'POST',
    body: JSON.stringify({
      method: 'link',
      email,
      csrf_token: csrfToken
    }),
    headers: {
      'Content-Type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    }
  }
)
```

---

## Route: `app/routes/recovery_.password.tsx`

### HTTP Calls:

#### 1. Loader - Get/Initialize Settings Flow (during recovery)
**Purpose:** Manage settings flow after recovery link clicked

```typescript
// Line 57-58 - Fetch existing flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/settings/flows?id=${flowId}`,
  { headers: { cookie, Accept: 'application/json' } }
)

// Line 70-71 - Initialize new flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/settings/browser?${url.searchParams}`,
  { headers: { cookie, Accept: 'application/json' } }
)
```

#### 2. Action - Update Password During Recovery
**Purpose:** Set new password

```typescript
// Line 179-193
const res = await fetch(
  `${KRATOS_URL}/self-service/settings?flow=${flowId}`,
  {
    method: 'POST',
    body: JSON.stringify({
      method: 'password',
      password,
      csrf_token: csrfToken
    }),
    headers: {
      'Content-type': 'application/json',
      Accept: 'application/json',
      cookie
    }
  }
)
```

---

## Route: `app/routes/verify.tsx`

### HTTP Calls:

#### 1. Loader - Get/Initialize Verification Flow
**Purpose:** Manage email verification flow

```typescript
// Line 55-56 - Fetch existing flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/verification/flows?id=${flowId}`,
  { headers: { cookie, Accept: 'application/json' } }
)

// Line 68-69 - Initialize new flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/verification/browser?${url.searchParams}`,
  { headers: { cookie, Accept: 'application/json' } }
)
```

#### 2. Action - Resend Verification Email
**Purpose:** Submit verification request

```typescript
// Line 193-207
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
```

---

## Route: `app/routes/settings_.password.tsx`

### HTTP Calls:

#### 1. Loader - Get/Initialize Settings Flow
**Purpose:** Manage password settings flow

```typescript
// Line 30-31 - Fetch existing flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/settings/flows?id=${flowId}`,
  { headers: { cookie, Accept: 'application/json' } }
)

// Line 43-44 - Initialize new flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/settings/browser?${url.searchParams}`,
  { headers: { cookie, Accept: 'application/json' } }
)
```

#### 2. Action - Update Password
**Purpose:** Change password

```typescript
// Line 133-147
const res = await fetch(
  `${KRATOS_URL}/self-service/settings?flow=${flowId}`,
  {
    method: 'POST',
    body: JSON.stringify({
      method: 'password',
      password,
      csrf_token: csrfToken
    }),
    headers: {
      'Content-type': 'application/json',
      Accept: 'application/json',
      cookie
    }
  }
)
```

---

## Route: `app/routes/settings_.phone.tsx`

### HTTP Calls:

#### 1. Loader - Get Settings Flow
**Purpose:** Fetch settings flow for phone update

```typescript
// Line 69-70
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/settings/flows?id=${flowId}`,
  { headers: { cookie, Accept: 'application/json' } }
)
```

#### 2. Action - Update Phone Number (Profile Method)
**Purpose:** Update user's phone trait

```typescript
// Line 262-279
const res = await fetch(
  `${KRATOS_URL}/self-service/settings?flow=${flowId}`,
  {
    method: 'POST',
    body: JSON.stringify({
      method: 'profile',
      traits: { ...session.identity.traits, phone },
      csrf_token: csrfToken
    }),
    headers: {
      'Content-type': 'application/json',
      Accept: 'application/json',
      cookie
    }
  }
)
```

---

## Route: `app/routes/login_.challenge.tsx`

### HTTP Calls:

#### 1. Loader - Get/Initialize Refresh Login Flow
**Purpose:** Session refresh (aal1 re-authentication)

```typescript
// Line 32-33 - Fetch existing flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
  { headers: { cookie, Accept: 'application/json' } }
)

// Line 45-46 - Initialize refresh flow
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/login/browser?refresh=true`,
  { headers: { Accept: 'application/json' } }
)
```

#### 2. Action - Submit Login Challenge
**Purpose:** Re-authenticate for privileged actions

```typescript
// Line 143-155
const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flowId}`, {
  method: 'POST',
  body: JSON.stringify({
    method: 'password',
    identifier: email,
    password,
    csrf_token: csrfToken
  }),
  headers: {
    'Content-type': 'application/json',
    Accept: 'application/json',
    Cookie: String(request.headers.get('cookie'))
  }
})
```

---

## Route: `app/routes/totp_.challenge.tsx`

### HTTP Calls:

#### 1. Loader - Initialize AAL2 Login Flow
**Purpose:** Create TOTP challenge flow

```typescript
// Line 37-46 - Initialize AAL2 flow (with redirect handling)
const initRes = await fetch(
  `${KRATOS_URL}/self-service/login/browser?aal=aal2${refresh ? '&refresh=true' : ''}`,
  {
    headers: { cookie },
    redirect: 'manual'
  }
)

// Line 69-76 - Fetch flow details
const kratosFlow = await fetch(
  `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
  { headers: { cookie, Accept: 'application/json' } }
)
```

#### 2. Action - Submit TOTP Code
**Purpose:** Verify TOTP code for AAL2

```typescript
// Line 127-138
const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flow}`, {
  method: 'POST',
  headers: {
    Accept: 'application/json',
    'Content-type': 'application/json',
    cookie: String(request.headers.get('cookie'))
  },
  body: JSON.stringify({
    method: 'totp',
    totp_code,
    csrf_token
  })
})
```

---

## Route: `app/routes/totp_.two-factor-authentication.tsx`

### HTTP Calls:

#### 1. Loader - Initialize Settings Flow for TOTP Setup
**Purpose:** Get TOTP QR code and secret

```typescript
// Line 39-46
const response = await fetch(
  `${KRATOS_URL}/self-service/settings/browser`,
  {
    headers: {
      Accept: 'application/json',
      cookie
    }
  }
)
```

#### 2. Action - Submit TOTP Code or Unlink
**Purpose:** Enable/disable TOTP

```typescript
// Line 213-236
const res = await fetch(
  `${KRATOS_URL}/self-service/settings?flow=${flowId}`,
  {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    },
    body: JSON.stringify(
      !totpUnlink
        ? { method: 'totp', totp_code: totpCode, csrf_token: csrfToken }
        : { method: 'totp', totp_unlink: true, csrf_token: csrfToken }
    )
  }
)
```

---

## Route: `app/routes/otp_.challenge.tsx`

### HTTP Calls:

#### 1. Action - Initialize Settings Flow
**Purpose:** After OTP verification, create settings flow for phone change

```typescript
// Line 171-173
const flowRes = await fetch(`${KRATOS_URL}/self-service/settings/browser`, {
  headers: { cookie, Accept: 'application/json' }
})
```

---

## Route: `app/routes/api.totp-challenge-init.tsx`

### HTTP Calls:

#### 1. Action - Initialize TOTP Challenge (Inline)
**Purpose:** Start AAL2 flow without page redirects

```typescript
// Line 16-24 - Initialize AAL2 flow
const initResponse = await fetch(
  `${KRATOS_URL}/self-service/login/browser?aal=aal2&refresh=true`,
  {
    method: 'GET',
    headers: { Accept: 'application/json', cookie },
    redirect: 'manual'
  }
)

// Line 45-52 - Fetch flow details
const flowResponse = await fetch(
  `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
  { headers: { Accept: 'application/json', cookie } }
)
```

---

## Route: `app/routes/api.totp-challenge-verify.tsx`

### HTTP Calls:

#### 1. Action - Verify TOTP Code (Inline)
**Purpose:** Verify TOTP without page redirects

```typescript
// Line 25-39
const verifyTotpCodeResponse = await fetch(
  `${KRATOS_URL}/self-service/login?flow=${flowId}`,
  {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      cookie
    },
    body: JSON.stringify({
      method: 'totp',
      totp_code: totpCode,
      csrf_token: csrfToken
    })
  }
)
```

---

## Route: `app/routes/api.check-totp-enabled.tsx`

### HTTP Calls:

#### 1. Loader - Check Session for TOTP Status
**Purpose:** Check if user has TOTP enabled

```typescript
// Line 13-14
const sessionResponse = await fetch(`${KRATOS_URL}/sessions/whoami`, {
  headers: { cookie }
})
```

---

## Route: `app/routes/deposit/route.tsx`

### HTTP Calls:

#### 1. Loader - Session Check
**Purpose:** Verify user is authenticated before deposit

```typescript
// Line 35-36
const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
  headers: args.request.headers
})
```

---

## Summary of Endpoints Used

| Endpoint | Method | Usage Count | Purpose |
|----------|--------|-------------|---------|
| `/sessions/whoami` | GET | 5 | Session validation |
| `/self-service/login/browser` | GET | 5 | Initialize login flow |
| `/self-service/login/flows?id=` | GET | 4 | Fetch login flow |
| `/self-service/login?flow=` | POST | 4 | Submit login |
| `/self-service/registration/browser` | GET | 1 | Initialize registration |
| `/self-service/registration?flow=` | POST | 1 | Submit registration |
| `/self-service/logout/browser` | GET | 1 | Initialize logout |
| `/self-service/logout?token=` | GET | 1 | Execute logout |
| `/self-service/recovery/browser` | GET | 1 | Initialize recovery |
| `/self-service/recovery/flows?id=` | GET | 1 | Fetch recovery flow |
| `/self-service/recovery?flow=` | POST | 1 | Submit recovery |
| `/self-service/verification/browser` | GET | 1 | Initialize verification |
| `/self-service/verification/flows?id=` | GET | 1 | Fetch verification flow |
| `/self-service/verification?flow=` | POST | 1 | Submit verification |
| `/self-service/settings/browser` | GET | 5 | Initialize settings flow |
| `/self-service/settings/flows?id=` | GET | 4 | Fetch settings flow |
| `/self-service/settings?flow=` | POST | 5 | Submit settings |
| `/admin/identities/{id}` | GET | 1 | Get identity (Admin API) |

---

## Files Importing from kratos.server.ts

| File | Imports Used |
|------|-------------|
| `root.tsx` | `hasUserSession` |
| `_index/route.tsx` | `hasUserSession` |
| `login.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `handleFlowError`, `isSessionAlreadyExitsMessage`, `kratosErrorMapping`, `requireNoUserSession` |
| `signup/route.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `handleFlowError`, `kratosErrorMapping`, `requireNoUserSession` |
| `logout.tsx` | `KRATOS_URL`, `handleFlowError` |
| `recovery.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `handleFlowError`, `kratosErrorMapping`, `requireNoUserSession` |
| `recovery_.password.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `handleFlowError`, `hasUserSession`, `kratosErrorMapping` |
| `verify.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `getUserSession`, `handleFlowError` |
| `settings_.password.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `handleFlowError`, `kratosErrorMapping` |
| `settings_.phone.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `getUserSession`, `handleFlowError`, `kratosErrorMapping` |
| `settings.profile-contact.tsx` | `getUserSession` |
| `login_.challenge.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `getUserSession`, `handleFlowError`, `kratosErrorMapping` |
| `totp_.challenge.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow`, `isSessionAlreadyExitsMessage` |
| `totp_.two-factor-authentication.tsx` | `KRATOS_URL`, `getCsrfTokenFromFlow` |
| `otp_.challenge.tsx` | `KRATOS_URL`, `getUserSession`, `handleFlowError` |
| `api.totp-challenge-init.tsx` | `KRATOS_URL` |
| `api.totp-challenge-verify.tsx` | `KRATOS_URL` |
| `api.check-totp-enabled.tsx` | `KRATOS_URL` |
| `api_.sendOtp.ts` | `getUserSession` |
| `deposit/route.tsx` | `KRATOS_URL` |
| `pay_.$paymentId/route.tsx` | `getUserSession` |
| `consent.tsx` | `getUserSession` |
| `support.tsx` | `getUserSession` |
| `waitlist.tsx` | `requireNoUserSession` |
| `confirmations.tsx` | `hasUserSession` |
| `me_.$.tsx` | `hasUserSession` |
