## Package Version

### Current State
```json
"@ory/kratos-client": "25.4.0"
```

This **is the latest version** of `@ory/kratos-client` for self-hosted Kratos. The versioning scheme (v25.x) reflects the API version/date rather than semantic versioning.

> [!NOTE]  
> The version numbering may seem unusual, but `25.4.0` corresponds to April 2025 API snapshot. For Ory Network users, `@ory/client` has different versioning (v1.x).

### Backend SDK (Go)
The backend uses `github.com/ory/kratos-client-go v0.13.1`, which is the Go SDK equivalent.

---

## SDK Initialization

### Create New Client Wrapper

Create a new file: [kratos-client.server.ts](file:///Users/antoniuneacsu/dev/interledger/interledger-app/typescript/protea/app/lib/kratos-client.server.ts)

```typescript
import { Configuration, FrontendApi, IdentityApi } from '@ory/kratos-client'

const KRATOS_PUBLIC_URL = process.env.KRATOS_URL
const KRATOS_ADMIN_URL = process.env.KRATOS_ADMIN_URL // For admin operations like isTotpSet

if (!KRATOS_PUBLIC_URL) {
  throw new Error('KRATOS_URL environment variable is not set')
}

// Frontend API client (public endpoints)
export const kratosPublicClient = new FrontendApi(
  new Configuration({
    basePath: KRATOS_PUBLIC_URL,
    baseOptions: {
      withCredentials: true
    }
  })
)

// Admin API client (for identity operations)
export const kratosAdminClient = KRATOS_ADMIN_URL
  ? new IdentityApi(
      new Configuration({
        basePath: KRATOS_ADMIN_URL
      })
    )
  : null

// Helper to extract cookies from Request
export function getCookieHeader(request: Request): string {
  return request.headers.get('cookie') ?? ''
}

// Helper to forward set-cookie headers
export function extractSetCookieHeader(headers: any): string | null {
  return headers?.['set-cookie'] ?? null
}
```

---

## SDK Method Mappings

The `FrontendApi` class provides typed methods for all self-service flows.

### Session Operations

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /sessions/whoami` | `frontend.toSession({ cookie })` |

### Login Flow

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /self-service/login/browser` | `frontend.createBrowserLoginFlow({ refresh?, aal?, returnTo? })` |
| `GET /self-service/login/flows?id=` | `frontend.getLoginFlow({ id, cookie })` |
| `POST /self-service/login?flow=` | `frontend.updateLoginFlow({ flow, updateLoginFlowBody, cookie })` |

### Registration Flow

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /self-service/registration/browser` | `frontend.createBrowserRegistrationFlow({ returnTo? })` |
| `GET /self-service/registration/flows?id=` | `frontend.getRegistrationFlow({ id, cookie })` |
| `POST /self-service/registration?flow=` | `frontend.updateRegistrationFlow({ flow, updateRegistrationFlowBody, cookie })` |

### Logout Flow

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /self-service/logout/browser` | `frontend.createBrowserLogoutFlow({ cookie })` |
| `GET /self-service/logout?token=` | `frontend.updateLogoutFlow({ token, returnTo?, cookie })` |

### Recovery Flow

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /self-service/recovery/browser` | `frontend.createBrowserRecoveryFlow({ returnTo? })` |
| `GET /self-service/recovery/flows?id=` | `frontend.getRecoveryFlow({ id, cookie })` |
| `POST /self-service/recovery?flow=` | `frontend.updateRecoveryFlow({ flow, updateRecoveryFlowBody, cookie })` |

### Verification Flow

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /self-service/verification/browser` | `frontend.createBrowserVerificationFlow({ returnTo? })` |
| `GET /self-service/verification/flows?id=` | `frontend.getVerificationFlow({ id, cookie })` |
| `POST /self-service/verification?flow=` | `frontend.updateVerificationFlow({ flow, updateVerificationFlowBody, cookie })` |

### Settings Flow

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /self-service/settings/browser` | `frontend.createBrowserSettingsFlow({ returnTo?, cookie })` |
| `GET /self-service/settings/flows?id=` | `frontend.getSettingsFlow({ id, cookie })` |
| `POST /self-service/settings?flow=` | `frontend.updateSettingsFlow({ flow, updateSettingsFlowBody, cookie })` |

### Identity API (Admin)

| Current HTTP Call | SDK Method |
|-------------------|------------|
| `GET /admin/identities/{id}` | `identityApi.getIdentity({ id })` |

---

## Detailed Replacement Examples

### 1. Session Check (`getUserSession`)

#### Current Implementation
```typescript
// kratos.server.ts
export async function getUserSession(request: Request, allowAal1 = false): Promise<Session> {
  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })
  // ... status handling
  return session.json()
}
```

#### SDK Replacement
```typescript
import { kratosPublicClient, getCookieHeader } from './kratos-client.server'
import type { Session } from '@ory/kratos-client'

export async function getUserSession(request: Request, allowAal1 = false): Promise<Session> {
  const cookie = getCookieHeader(request)
  
  try {
    const { data: session } = await kratosPublicClient.toSession({ cookie })
    
    // AAL2 check
    if (!allowAal1 && session.authenticator_assurance_level === 'aal1') {
      const requestUrl = new URL(request.url)
      const returnTo = safeReturnTo(requestUrl.pathname + requestUrl.search)
      const searchParams = new URLSearchParams()
      searchParams.set('returnTo', returnTo)
      throw redirect(`${route('/login')}?${searchParams.toString()}`)
    }
    
    return session
  } catch (error) {
    if (error instanceof Response) throw error // Already a redirect
    
    const requestUrl = new URL(request.url)
    const returnTo = safeReturnTo(requestUrl.pathname + requestUrl.search)
    const searchParams = new URLSearchParams()
    searchParams.set('returnTo', returnTo)
    
    // Handle API errors
    if (error.response?.status === 401 || error.response?.status === 500) {
      throw redirect(`${route('/login')}?${searchParams.toString()}`)
    }
    if (error.response?.status === 403 || error.response?.status === 422) {
      if (!allowAal1) {
        throw redirect(`${route('/login')}?${searchParams.toString()}`)
      }
    }
    throw error
  }
}
```

---

### 2. Login Flow (Loader)

#### Current Implementation
```typescript
// login.tsx loader
if (flowId) {
  const flowRes = await fetch(
    `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
    { headers: { cookie, Accept: 'application/json' } }
  )
  flow = await flowRes.json()
  if (flowRes.status >= 400) handleFlowError(flow, 'login')
} else {
  const flowRes = await fetch(
    `${KRATOS_URL}/self-service/login/browser${url.search}`,
    { headers: { Accept: 'application/json' } }
  )
  flow = await flowRes.json()
  if (flowRes.status >= 400) handleFlowError(flow, 'login')
  headers = trimHeaders(flowRes.headers, ['set-cookie'])
}
```

#### SDK Replacement
```typescript
import { kratosPublicClient, getCookieHeader, extractSetCookieHeader } from '~/lib/kratos-client.server'

// login.tsx loader
const cookie = getCookieHeader(request)

let flow: LoginFlow
let responseHeaders: Headers | undefined

try {
  if (flowId) {
    const { data } = await kratosPublicClient.getLoginFlow({ id: flowId, cookie })
    flow = data
  } else {
    const returnTo = url.searchParams.get('return_to') ?? undefined
    const aal = url.searchParams.get('aal') as 'aal1' | 'aal2' | undefined
    const refresh = url.searchParams.get('refresh') === 'true'
    
    const response = await kratosPublicClient.createBrowserLoginFlow({ returnTo, aal, refresh })
    flow = response.data
    
    // Extract set-cookie header
    const setCookie = extractSetCookieHeader(response.headers)
    if (setCookie) {
      responseHeaders = new Headers()
      responseHeaders.set('Set-Cookie', setCookie)
    }
  }
} catch (error) {
  handleFlowError(error, 'login')
}

return json(
  { flowId: flow.id, csrfToken: getCsrfTokenFromFlow(flow) },
  { headers: responseHeaders }
)
```

---

### 3. Login Flow (Action - Submit)

#### Current Implementation
```typescript
// login.tsx action
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

if (res.status >= 400 && res.status !== 422) {
  const errors = await kratosErrorMapping(res, fieldErrors)
  // ...
}
```

#### SDK Replacement
```typescript
import { kratosPublicClient, getCookieHeader, extractSetCookieHeader } from '~/lib/kratos-client.server'
import { UpdateLoginFlowWithPasswordMethod } from '@ory/kratos-client'

// login.tsx action
const cookie = getCookieHeader(request)

try {
  const response = await kratosPublicClient.updateLoginFlow({
    flow: flowId as string,
    updateLoginFlowBody: {
      method: 'password',
      identifier: email as string,
      password: password as string,
      csrf_token: csrfToken as string
    } as UpdateLoginFlowWithPasswordMethod,
    cookie
  })
  
  const headers = new Headers()
  const setCookie = extractSetCookieHeader(response.headers)
  if (setCookie) {
    headers.set('Set-Cookie', setCookie)
  }
  
  // Check for AAL1 (needs TOTP)
  if (response.data.session?.authenticator_assurance_level === 'aal1') {
    return redirect(`${route('/totp/two-factor-authentication')}?${searchParams.toString()}`, { headers })
  }
  
  return redirect(returnTo || '/', { headers })
  
} catch (error) {
  // Status 400 = validation errors
  if (error.response?.status === 400) {
    const flowData = error.response.data
    const errors = mapFlowToFieldErrors(flowData, fieldErrors)
    
    if (isSessionAlreadyExitsMessage(errors.form)) {
      return redirect(returnTo || '/')
    }
    return error(request, { errors }, { action: 'Contact support' })
  }
  
  // Status 422 = needs AAL2
  if (error.response?.status === 422) {
    const headers = new Headers()
    const setCookie = extractSetCookieHeader(error.response.headers)
    if (setCookie) headers.set('Set-Cookie', setCookie)
    return redirect(`${route('/totp/challenge')}?${searchParams.toString()}`, { headers })
  }
  
  throw error
}
```

---

### 4. Registration Flow

#### Current Implementation
```typescript
// signup/route.tsx
const flowRes = await fetch(
  `${KRATOS_URL}/self-service/registration/browser?${url.searchParams}`,
  { headers: { Accept: 'application/json' } }
)
const kratosFlow = await flowRes.json()
if (flowRes.status >= 400) handleFlowError(kratosFlow, 'signup')

// Action
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
    headers: { 'Content-type': 'application/json', cookie }
  }
)
```

#### SDK Replacement
```typescript
import { kratosPublicClient, getCookieHeader, extractSetCookieHeader } from '~/lib/kratos-client.server'
import { UpdateRegistrationFlowWithPasswordMethod } from '@ory/kratos-client'

// Loader
const returnTo = url.searchParams.get('return_to') ?? undefined
const response = await kratosPublicClient.createBrowserRegistrationFlow({ returnTo })
const kratosFlow = response.data

// Action
const cookie = getCookieHeader(request)

try {
  const response = await kratosPublicClient.updateRegistrationFlow({
    flow: kratosFlowId as string,
    updateRegistrationFlowBody: {
      method: 'password',
      traits: { email, phone, firstName, lastName, countryCode },
      password,
      csrf_token: kratosCsrfToken
    } as UpdateRegistrationFlowWithPasswordMethod,
    cookie
  })
  
  // response.data contains the session and identity
  const successData = response.data
  const headers = new Headers()
  const setCookie = extractSetCookieHeader(response.headers)
  if (setCookie) headers.set('Set-Cookie', setCookie)
  
  return redirectWithSnackbar(request, route('/wallet-address'), {...}, { headers })
  
} catch (error) {
  if (error.response?.status >= 400) {
    const flowData = error.response.data
    const errs = mapFlowToFieldErrors(flowData, errors)
    return error(request, { errors: errs })
  }
  throw error
}
```

---

### 5. Logout Flow

#### Current Implementation
```typescript
// logout.tsx
// Loader
const flowRes = await fetch(`${KRATOS_URL}/self-service/logout/browser`, {
  headers: { cookie, Accept: 'application/json' }
})
flow = await flowRes.json()

// Action
const res = await fetch(`${KRATOS_URL}/self-service/logout?token=${token}`, {
  method: 'GET',
  headers: { Accept: 'application/json', cookie }
})
```

#### SDK Replacement
```typescript
import { kratosPublicClient, getCookieHeader, extractSetCookieHeader } from '~/lib/kratos-client.server'

// Loader
const cookie = getCookieHeader(request)
const { data: logoutFlow } = await kratosPublicClient.createBrowserLogoutFlow({ cookie })
return jsonWithCSRF(request, { logoutToken: logoutFlow.logout_token })

// Action
const cookie = getCookieHeader(request)
try {
  const response = await kratosPublicClient.updateLogoutFlow({
    token: token as string,
    cookie
  })
  
  const session = await getSession(cookie)
  const sessionHeaders = await destroySession(session)
  
  const headers = new Headers()
  const setCookie = extractSetCookieHeader(response.headers)
  if (setCookie) headers.append('Set-Cookie', setCookie)
  headers.append('Set-Cookie', sessionHeaders)
  
  return redirect(route('/'), { headers })
  
} catch (error) {
  return json({ errors: { form: 'Something went wrong trying to logout.' } }, { status: 400 })
}
```

---

### 6. Settings Flow (Password Update)

#### Current Implementation
```typescript
// settings_.password.tsx
const res = await fetch(`${KRATOS_URL}/self-service/settings?flow=${flowId}`, {
  method: 'POST',
  body: JSON.stringify({
    method: 'password',
    password,
    csrf_token: csrfToken
  }),
  headers: { 'Content-type': 'application/json', Accept: 'application/json', cookie }
})
```

#### SDK Replacement
```typescript
import { kratosPublicClient, getCookieHeader } from '~/lib/kratos-client.server'
import { UpdateSettingsFlowWithPasswordMethod } from '@ory/kratos-client'

const cookie = getCookieHeader(request)

try {
  await kratosPublicClient.updateSettingsFlow({
    flow: flowId as string,
    updateSettingsFlowBody: {
      method: 'password',
      password,
      csrf_token: csrfToken
    } as UpdateSettingsFlowWithPasswordMethod,
    cookie
  })
  
  return redirectWithSnackbar(request, route('/settings'), {
    message: 'New password successfully saved.',
    icon: 'close'
  })
  
} catch (error) {
  if (error.response?.status > 400) {
    handleFlowError(error.response.data, 'settings/password')
  }
  if (error.response?.status === 400) {
    const errs = mapFlowToFieldErrors(error.response.data, fieldErrors)
    return error(request, { errors: errs })
  }
  throw error
}
```

---

### 7. TOTP Flow

#### Current Implementation
```typescript
// totp_.two-factor-authentication.tsx
const response = await fetch(`${KRATOS_URL}/self-service/settings/browser`, {
  headers: { Accept: 'application/json', cookie }
})

// Action
const res = await fetch(`${KRATOS_URL}/self-service/settings?flow=${flowId}`, {
  method: 'POST',
  body: JSON.stringify({ method: 'totp', totp_code: totpCode, csrf_token: csrfToken })
})
```

#### SDK Replacement
```typescript
import { kratosPublicClient, getCookieHeader, extractSetCookieHeader } from '~/lib/kratos-client.server'
import { UpdateSettingsFlowWithTotpMethod } from '@ory/kratos-client'

// Loader
const cookie = getCookieHeader(request)
const { data: flow } = await kratosPublicClient.createBrowserSettingsFlow({ cookie })

// Action
try {
  await kratosPublicClient.updateSettingsFlow({
    flow: flowId as string,
    updateSettingsFlowBody: totpUnlink
      ? { method: 'totp', totp_unlink: true, csrf_token: csrfToken }
      : { method: 'totp', totp_code: totpCode, csrf_token: csrfToken } as UpdateSettingsFlowWithTotpMethod,
    cookie
  })
  
  return redirectDocument(returnTo ?? '/')
  
} catch (error) {
  if (error.response?.status === 400) {
    return json({ errors: { totpCode: 'Invalid code...' } }, { status: 400 })
  }
  throw error
}
```

---

### 8. Identity API (Admin) - Check TOTP Status

#### Current Implementation
```typescript
// totp.server.ts
export async function isTotpSet(session: Session, headers: Headers): Promise<boolean> {
  const response = await fetch(`${KRATOS_URL}/admin/identities/${session.identity.id}`, {
    headers
  })
  const identity = await response.json()
  return !!identity.credentials?.totp
}
```

#### SDK Replacement
```typescript
import { kratosAdminClient } from './kratos-client.server'

export async function isTotpSet(session: Session): Promise<boolean> {
  if (session?.authenticator_assurance_level === 'aal2') {
    return true
  }
  
  if (!kratosAdminClient) {
    console.warn('Kratos Admin URL not configured, cannot check TOTP status')
    return false
  }
  
  try {
    const { data: identity } = await kratosAdminClient.getIdentity({
      id: session.identity.id
    })
    return !!identity.credentials?.totp
  } catch (error) {
    console.error('Failed to fetch identity:', error)
    return false
  }
}
```

---

## Error Handling Updates

### Current Error Mapping

The current `kratosErrorMapping()` function parses raw JSON responses. With the SDK, errors are thrown as exceptions with structured data.

#### Updated Error Mapping Helper

```typescript
import type { FlowError, UiNode } from '@ory/kratos-client'

// Helper to map SDK flow errors to field errors
export function mapFlowToFieldErrors<T extends object>(
  flowData: any,
  fieldErrors: T
): T {
  if (!flowData?.ui?.nodes) return fieldErrors
  
  for (const node of flowData.ui.nodes as UiNode[]) {
    if (node.messages && node.messages.length > 0) {
      const attrName = (node.attributes as any).name
      if (attrName) {
        (fieldErrors as any)[attrName] = kratosErrorMessage(node.messages[0])
      }
    }
  }
  
  if (flowData.ui.messages?.length > 0) {
    (fieldErrors as any).form = kratosErrorMessage(flowData.ui.messages[0])
  }
  
  return fieldErrors
}

// Updated handleFlowError for SDK exceptions
export function handleFlowError(error: any, flowType: string, flowId?: string): void {
  const flowData = error.response?.data ?? error
  
  if (!flowData?.error?.id) return
  
  let redirectRoute = `/${flowType}`
  
  switch (flowData.error.id) {
    case 'session_inactive':
      throw redirect(route('/login'), {
        headers: { 'Clear-Site-Data': 'cookies' }
      })
    case 'session_aal2_required':
      throw redirect(`/totp/challenge?returnTo=${redirectRoute}`)
    case 'session_already_available':
      throw redirect(route('/'))
    case 'session_refresh_required':
      throw redirect(`/login/challenge`)
    case 'self_service_flow_expired':
    case 'security_csrf_violation':
    case 'security_identity_mismatch':
      throw redirect(redirectRoute, {
        headers: { 'Clear-Site-Data': 'cookies' }
      })
    case 'browser_location_change_required':
      throw redirect(flowData.error.redirect_browser_to)
  }
}
```

---

## Cookie Handling in Remix SSR

> [!WARNING]
> The SDK uses axios under the hood. For SSR in Remix, we must manually forward cookies.

### Key Pattern: Always Pass Cookies

```typescript
// Every SDK call that needs authentication must include the cookie parameter
const cookie = getCookieHeader(request)

await kratosPublicClient.toSession({ cookie })
await kratosPublicClient.getLoginFlow({ id: flowId, cookie }) 
await kratosPublicClient.updateLoginFlow({ flow, updateLoginFlowBody, cookie })
```

### Key Pattern: Extract and Forward Set-Cookie Headers

```typescript
const response = await kratosPublicClient.createBrowserLoginFlow({ returnTo })

// The response.headers contains set-cookie for CSRF token
const setCookie = response.headers?.['set-cookie']
if (setCookie) {
  return json(data, { 
    headers: { 'Set-Cookie': Array.isArray(setCookie) ? setCookie.join(', ') : setCookie }
  })
}
```

---

## Type Updates

### Old Types (v25.4.0)
```typescript
import type {
  SelfServiceLoginFlow,
  SelfServiceRecoveryFlow,
  SelfServiceRegistrationFlow,
  SelfServiceSettingsFlow,
  SelfServiceVerificationFlow,
  Session,
  UiNodeInputAttributes
} from '@ory/kratos-client'
```

### New Types (Latest)
```typescript
import type {
  LoginFlow,
  RecoveryFlow,
  RegistrationFlow,
  SettingsFlow,
  VerificationFlow,
  Session,
  UiNodeInputAttributes,
  UpdateLoginFlowWithPasswordMethod,
  UpdateRegistrationFlowWithPasswordMethod,
  UpdateSettingsFlowWithPasswordMethod,
  UpdateSettingsFlowWithTotpMethod,
  UpdateRecoveryFlowWithLinkMethod
} from '@ory/kratos-client'
```

---

## Migration Checklist by File

Use this checklist when migrating each file:

- [ ] Update imports (types and client)
- [ ] Replace `fetch()` calls with SDK methods
- [ ] Add cookie parameter to all authenticated calls
- [ ] Extract set-cookie headers from responses
- [ ] Update error handling to catch SDK exceptions
- [ ] Update type annotations to new SDK types
- [ ] Test the specific flow thoroughly
