// Server-side HTTP client for the backend's /plaid/* surface.
//
// Mirrors the cookie-forwarding pattern used by grpc.server.ts and the
// existing fetch-based routes (e.g. api_.statements_.*) — loaders/actions
// pass the inbound Request so we can lift the Kratos session cookie onto the
// backend call.

import { redirect } from 'react-router'
import { href } from 'react-router'
import { captureMessage } from '@sentry/react-router'
import { CLEAR_SESSION_COOKIE_HEADER } from './kratos/kratos-client.server'

const BACKEND_HTTP_URL = process.env.BACKEND_HTTP_URL || 'http://backend:8080'
const PLAID_API_PATH = '/api/plaid'

/**
 * PlaidError is the shape returned to callers on a non-2xx backend response.
 * It mirrors the BFF ConnectError pattern.
 */
export class PlaidError {
  public readonly _request: Request
  public readonly status: number
  public readonly errorCode: string
  public readonly message: string
  public readonly reqId?: string

  constructor(
    request: Request,
    status: number,
    errorCode: string,
    message: string,
    reqId?: string
  ) {
    this._request = request
    this.status = status
    this.errorCode = errorCode
    this.message = message
    this.reqId = reqId

    const url = new URL(request.url)

    captureMessage('Error received in Plaid API', {
      extra: {
        url: url.pathname,
        status,
        errorCode,
        message,
        reqId
      }
    })

    if (status === 401) {
      url.searchParams.set('returnTo', url.pathname + url.search)
      throw redirect(href('/login') + url.search, {
        headers: {
          'Set-Cookie': CLEAR_SESSION_COOKIE_HEADER
        }
      })
    }
  }
}

/** Type guard for PlaidError */
export function isPlaidError(response: unknown): response is PlaidError {
  return (
    response !== null &&
    typeof response === 'object' &&
    '_request' in response &&
    'errorCode' in response &&
    'status' in response
  )
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

/** Shape returned by GET /plaid/registered (Phase 2). */
export interface PlaidRegisteredResult {
  plaid_account_ids: string[]
}

/** Shape returned by POST /plaid/link-to-fiant (Phase 2). */
export interface PlaidLinkToFiantResult {
  linked_account_id: string
  payment_information_id: string
  already_linked: boolean
}

/** Body accepted by POST /plaid/link-to-fiant. */
export interface PlaidLinkToFiantArgs {
  account_id: string
  account_name?: string
  account_mask?: string
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
): Promise<T | PlaidError> {
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
      return new PlaidError(request, res.status, 'INTERNAL', text || res.statusText)
    }
  }

  if (!res.ok) {
    const errBody = (body ?? {}) as {
      error_code?: string
      message?: string
      req_id?: string
    }
    return new PlaidError(
      request,
      res.status,
      errBody.error_code || 'INTERNAL',
      errBody.message || res.statusText,
      errBody.req_id
    )
  }

  return body as T
}

/* ─── typed wrappers ─────────────────────────────────────────────────── */

function getState(request: Request): Promise<PlaidState | PlaidError> {
  return plaidFetch<PlaidState>(request, `${PLAID_API_PATH}/state`)
}

function createLinkToken(request: Request): Promise<PlaidLinkToken | PlaidError> {
  return plaidFetch<PlaidLinkToken>(request, `${PLAID_API_PATH}/link-token`, {
    method: 'POST'
  })
}

function exchangePublicToken(
  request: Request,
  publicToken: string
): Promise<PlaidExchangeResult | PlaidError> {
  return plaidFetch<PlaidExchangeResult>(request, `${PLAID_API_PATH}/exchange`, {
    method: 'POST',
    body: JSON.stringify({ public_token: publicToken })
  })
}

function getAccounts(request: Request): Promise<PlaidProductResponse | PlaidError> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/accounts`)
}

function getAuth(request: Request): Promise<PlaidProductResponse | PlaidError> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/auth`)
}

function getBalance(request: Request): Promise<PlaidProductResponse | PlaidError> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/balance`)
}

function getIdentity(request: Request): Promise<PlaidProductResponse | PlaidError> {
  return plaidFetch<PlaidProductResponse>(request, `${PLAID_API_PATH}/identity`)
}

function getTransactions(
  request: Request
): Promise<PlaidTransactionsResult | PlaidError> {
  return plaidFetch<PlaidTransactionsResult>(request, `${PLAID_API_PATH}/transactions`)
}

function disconnect(
  request: Request
): Promise<PlaidDisconnectResult | PlaidError> {
  return plaidFetch<PlaidDisconnectResult>(request, `${PLAID_API_PATH}/disconnect`, {
    method: 'DELETE'
  })
}

function getRegistered(
  request: Request
): Promise<PlaidRegisteredResult | PlaidError> {
  return plaidFetch<PlaidRegisteredResult>(request, `${PLAID_API_PATH}/registered`)
}

function linkToFiant(
  request: Request,
  args: PlaidLinkToFiantArgs
): Promise<PlaidLinkToFiantResult | PlaidError> {
  return plaidFetch<PlaidLinkToFiantResult>(request, `${PLAID_API_PATH}/link-to-fiant`, {
    method: 'POST',
    body: JSON.stringify(args)
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
  getRegistered,
  linkToFiant,
}