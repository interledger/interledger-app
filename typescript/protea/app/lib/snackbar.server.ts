import { redirect } from '@remix-run/node'
import { commitSession, getSession } from '~/session.server'

export type SnackbarType = {
  message: string
  show?: boolean
  action?: string
  icon?: string
}

/**
 * Allows flashing a snackbar to the session storage.
 * @param request
 * @param snackbar The snackbar message and action to render.
 * @returns Promise<string> The Set-Cookie header to be used in the HTTP response.
 */
export async function flashSnackbar(
  request: Request,
  snackbar: SnackbarType
): Promise<string> {
  const session = await getSession(request.headers.get('Cookie'))
  session.flash('snackbar', snackbar)
  return commitSession(session)
}

/**
 * Fetches the data for a snackbar, and commits the session so the snackbar has been spent.
 * @param request Request
 * @returns Promise<SnackbarType>
 */
export async function getSnackbar(request: Request): Promise<SnackbarType> {
  const session = await getSession(request.headers.get('Cookie'))

  const snackbar = {
    // NOTE: session.has must be called before userSettings.get
    show: session.has('snackbar'),
    ...session.get('snackbar')
  }
  await commitSession(session)

  return snackbar
}

/**
 * Helper method used to redirect the user to a new page with flash snackbar data.
 * @param request Request
 * @param url Url to redirect to
 * @param snackbar Snackbar values
 * @param init Response options
 * @returns Redirect response
 */
export async function redirectWithSnackbar(
  request: Request,
  url: string,
  snackbar: SnackbarType,
  init?: ResponseInit
) {
  const session = await getSession(request.headers.get('Cookie'))
  session.flash('snackbar', snackbar)
  
  const cookie = await commitSession(session)
  const newHeaders = new Headers(init?.headers)
  newHeaders.append('Set-Cookie', cookie)

  return redirect(url, {
    ...init,
    headers: newHeaders
  })
}
