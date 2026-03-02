import { data as rrData, UNSAFE_DataWithResponseInit as DataWithResponseInit } from 'react-router';
import { captureMessage } from '@sentry/react-router'
import { randomUUID } from 'crypto'
import { commitSession, getSession } from '~/session.server'

async function getCSRFToken(
  request: Request,
  headers?: ResponseInit['headers']
) {
  const session = await getSession(request.headers.get('Cookie'))

  // It's important that we always get the already set CSRF token from the session.
  // Loaders can run in parallel and we don't want to override the CSRF token.
  let csrfToken = session.get('csrf-token')
  if (typeof csrfToken === 'undefined') {
    csrfToken = randomUUID()
    session.set('csrf-token', csrfToken)
    const url = new URL(request.url)
    captureMessage('Generating new CSRF token', {
      extra: {
        url: url.pathname,
        csrfToken,
        session: session.id
      }
    })
  }

  const cookie = await commitSession(session)
  const newHeaders = new Headers(headers)
  newHeaders.append('Set-Cookie', cookie)

  return { csrfToken, newHeaders }
}

type JsonWithCSRFFunction = <Data>(
  request: Request,
  data: Data,
  init?: number | ResponseInit
) => Promise<
  DataWithResponseInit<
    Data &
    object & {
      csrfToken: `${string}-${string}-${string}-${string}-${string}`
    }
  >
>

/**
 * This is an extension of the data function from Remix.
 * This function will add a CSRF token to the response data.
 * And ensure that a new CSRF token is always set in the session storage.
 * @param request
 * @param data
 * @param init
 */
export const jsonWithCSRF: JsonWithCSRFFunction = async (
  request,
  data,
  init
) => {
  let responseInit = typeof init === 'number' ? { status: init } : init

  const { csrfToken, newHeaders } = await getCSRFToken(
    request,
    responseInit?.headers
  )

  if (typeof data !== 'object') {
    throw rrData(
      {},
      {
        status: 400,
        statusText: 'Only objects should be returned from loaders.'
      }
    )
  }

  return rrData(
    { ...data, csrfToken },
    {
      ...responseInit,
      headers: newHeaders
    }
  )
}

/**
 * This function will validate the CSRF token in the request.
 * It will throw an error if:
 *  - there is no CSRF token stored in the user's session
 *  - the token is different to the CSRF token stored in the user's session.
 *  - the token is missing from the FormData.
 * @param request
 * @param form
 */
export async function validateCSRFToken(
  request: Request,
  form: FormData
): Promise<void> {
  const csrfToken = form.get('csrfToken') as string

  let session = await getSession(request.headers.get('Cookie'))
  let serverToken = session.get('csrf-token')

  if (
    !form.has('csrfToken') ||
    csrfToken === '' ||
    typeof serverToken !== 'string' ||
    serverToken === '' ||
    serverToken !== csrfToken
  ) {
    // throw new Error('No CSRF token set in session.')
    const url = new URL(request.url)
    captureMessage('Invalid CSRF token', {
      extra: {
        url: url.pathname,
        serverToken,
        csrfToken,
        session: session.id
      }
    })
    // throw rrData(
    //   {
    //     action: {
    //       route: url.pathname,
    //       text: 'Try again'
    //     }
    //   },
    //   { status: 422, statusText: 'Invalid CSRF token.' }
    // )
  }

  // Invalidate old token by setting a new one.
  // Ensures we never use the same token twice.
  // And that there is always a token set in the session.
  // If we simply unset the token, we run into possible race conditions in loaders fetching data in parallel.
  session.set('csrf-token', randomUUID())

  await commitSession(session)
}
