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
