import { redirect } from 'react-router'
import { href } from 'react-router'
import { captureMessage } from '@sentry/react-router'
import { CLEAR_SESSION_COOKIE_HEADER } from './kratos/kratos-client.server'
import logger from './logger.server'

const BACKEND_HTTP_URL = process.env.BACKEND_HTTP_URL || 'http://backend:8080'
const PLAID_API_PATH = '/api/plaid'

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

export function isPlaidError(response: unknown): response is PlaidError {
  return (
    response !== null &&
    typeof response === 'object' &&
    '_request' in response &&
    'errorCode' in response &&
    'status' in response
  )
}

/** Shape returned by POST /plaid/link-token. */
export interface PlaidLinkToken {
  link_token: string
  expiration: string
}

/** Shape returned by POST /plaid/link-to-fiant */
export interface PlaidLinkToFiantResult {
  linked_account_id: string
  payment_information_id: string
  already_linked: boolean
}

/** Body accepted by POST /plaid/link-to-fiant. */
export interface PlaidLinkToFiantArgs {
  public_token: string
  account_id: string
  account_name?: string
  account_mask?: string
}

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
  try {
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
  } catch (err) {
    if (err instanceof Response) throw err
    const message = err instanceof Error ? err.message : String(err)
    logger.error({ err, path }, 'plaidFetch: unexpected error')
    return new PlaidError(request, 500, 'NETWORK_ERROR', message)
  }
}


function createLinkToken(request: Request): Promise<PlaidLinkToken | PlaidError> {
  return plaidFetch<PlaidLinkToken>(request, `${PLAID_API_PATH}/link-token`, {
    method: 'POST'
  })
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
  createLinkToken,
  linkToFiant,
}