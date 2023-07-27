import type { Session } from '@remix-run/node'
import { randomUUID } from 'crypto'
import { commitSession, getSession } from '~/session.server'

type csrfSessionData = {
  ['csrf-token']: string
}

// commitSession must be called after this function and the Set-Cookie header set in the response.
export async function getSessionWithCSRFToken(
  request: Request
): Promise<Session<csrfSessionData>> {
  const session = await getSession(request.headers.get('Cookie'))
  let token = session.get('csrf-token')
  if (typeof token !== 'string') {
    session.set('csrf-token', randomUUID())
  }

  return session
}

// This will check that
export async function validateCSRFToken(
  request: Request,
  token: string
): Promise<void> {
  if (token === '') {
    throw new Error('CSRF token is empty.')
  }
  let session = await getSession(request.headers.get('Cookie'))
  let serverToken = session.get('csrf-token')
  if (typeof serverToken !== 'string' || serverToken === '') {
    throw new Error('No CSRF token set in session.')
  }

  if (serverToken !== token) {
    throw new Error('Invalid CSRF token.')
  }

  // invalidate old token
  session.set('csrf-token', randomUUID())

  await commitSession(session)
}
