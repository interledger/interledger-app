// Server-side HTTP client for the backend's /plaid/* surface.
//
// Mirrors the cookie-forwarding pattern used by grpc.server.ts and the
// existing fetch-based routes (e.g. api_.statements_.*) — loaders/actions
// pass the inbound Request so we can lift the Kratos session cookie onto the
// backend call.
//
// All endpoints are documented in
// /Users/antoniuneacsu/dev/interledger/interledger-app/documentation/poc/plaid/architecture.md#4-api-contract-phase-1.
//
// NOTE: this client currently throws PlaidError on non-2xx responses. A
// follow-up task (`FX` in tasks.md) aligns it with the BFF error pattern used
// by `grpc.server.ts` / `error.server.ts` (returns `T | PlaidError`, Sentry
// capture, 401 → /login). Kept simple for the POC.

const BACKEND_HTTP_URL = process.env.BACKEND_HTTP_URL || 'http://backend:8080'
const PLAID_API_PATH = '/api/plaid'

/**
 * PlaidError is the shape returned to callers on a non-2xx backend response.
 * It mirrors the apperrors.WriteAppError JSON envelope so existing snackbar
 * helpers can render a useful message.
 */
export class PlaidError extends Error {
  constructor(
    public readonly status: number,
    public readonly errorCode: string,
    message: string,
    public readonly reqId?: string
  ) {
    super(message)
    this.name = 'PlaidError'
  }
}

/** Shape returned by GET /plaid/state. */
export interface PlaidState {
  linked: boolean
  item_id?: string
  institution_name?: string
  linked_at?: string
}

/** Shape returned by POST /plaid/link-token. */
export interface PlaidLinkToken {
  link_token: string
  expiration: string
}

/** Shape returned by POST /plaid/exchange. */
export interface PlaidExchangeResult {
  item_id: string
  institution_name: string
}

/** Shape returned by GET /plaid/transactions. */
export interface PlaidTransactionsResult {
  added: unknown[]
  modified: unknown[]
  removed: unknown[]
  next_cursor: string
}

/** Shape returned by DELETE /plaid/disconnect. */
export interface PlaidDisconnectResult {
  disconnected: boolean
}

/** Plaid product responses are passed through verbatim (SDK shapes). */
export type PlaidProductResponse = unknown

/** Endpoints whose response shape we don't model strictly. */
export type PlaidProduct =
  | 'accounts'
  | 'auth'
  | 'balance'
  | 'identity'
  | 'transactions'

/**
 * plaidFetch is the single transport used by every wrapper below. It forwards
 * the inbound cookie header so the backend's MakeUserMiddleware recognises
 * the Kratos session, decodes the JSON body, and converts any non-2xx
 * response into a structured PlaidError.
 */
async function plaidFetch<T>(
  request: Request,
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const cookie = request.headers.get('cookie') || ''
  const headers = new Headers(init.headers)
  if (cookie) headers.set('cookie', cookie)
  if (init.body && !headers.has('content-type')) {
    headers.set('content-type', 'application/json')
  }

  const res = await fetch(`${BACKEND_HTTP_URL}${path}`, {
    ...init,
    headers
  })

  const text = await res.text()
  let body: unknown = null
  if (text.length > 0) {
    try {
      body = JSON.parse(text)
    } catch {
      // Surface as a PlaidError so callers don't have to special-case non-JSON
      // 5xx pages.
      throw new PlaidError(res.status, 'INTERNAL', text || res.statusText)
    }
  }

  if (!res.ok) {
    const errBody = (body ?? {}) as {
      error_code?: string
      message?: string
      req_id?: string
    }
    throw new PlaidError(
      res.status,
      errBody.error_code || 'INTERNAL',
      errBody.message || res.statusText,
      errBody.req_id
    )
  }

  return body as T
}

/* ─── typed wrappers ─────────────────────────────────────────────────── */

function getState(request: Request): Promise<PlaidState> {
  return plaidFetch<PlaidState>(request, `${PLAID_API_PATH}/state`)
}

function createLinkToken(request: Request): Promise<PlaidLinkToken> {
  return plaidFetch<PlaidLinkToken>(request, `${PLAID_API_PATH}/link-token`, {
    method: 'POST'
  })
}

function exchangePublicToken(
  request: Request,
  publicToken: string
): Promise<PlaidExchangeResult> {
  return plaidFetch<PlaidExchangeResult>(request, `${PLAID_API_PATH}/exchange`, {
    method: 'POST',
    body: JSON.stringify({ public_token: publicToken })
  })
}

function getAccounts(request: Request): Promise<PlaidProductResponse> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/accounts`)
}

function getAuth(request: Request): Promise<PlaidProductResponse> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/auth`)
}

function getBalance(request: Request): Promise<PlaidProductResponse> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/balance`)
}

function getIdentity(request: Request): Promise<PlaidProductResponse> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/identity`)
}

function getTransactions(
  request: Request
): Promise<PlaidTransactionsResult> {
  return plaidFetch<PlaidTransactionsResult>(request, `${PLAID_API_PATH}/transactions`)
}

function disconnect(
  request: Request
): Promise<PlaidDisconnectResult> {
  return plaidFetch<PlaidDisconnectResult>(request, `${PLAID_API_PATH}/disconnect`, {
    method: 'DELETE'
  })
}

export default {
  getState,
  createLinkToken,
  exchangePublicToken,
  getAccounts,
  getAuth,
  getBalance,
  getIdentity,
  getTransactions,
  disconnect,
}