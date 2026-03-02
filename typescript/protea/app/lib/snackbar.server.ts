import { data as rrData, redirect, UNSAFE_DataWithResponseInit as DataWithResponseInit } from 'react-router';
import { v4 } from 'uuid'
import type { SnackbarType } from '~/lib/useScaffoldStore'
import { commitSession, getSession } from '~/session.server'

/**
 * Allows flashing a snackbar to the session storage.
 * @param request
 * @param snackbar The snackbar message and action to render.
 * @returns Promise<string> The Set-Cookie header to be used in the HTTP response.
 */
export async function flashSnackbar(
  request: Request,
  snackbar: Partial<SnackbarType>
): Promise<string> {
  const session = await getSession(request.headers.get('Cookie'))
  snackbar.fromServer = true
  if (typeof snackbar.id === 'undefined') snackbar.id = v4()
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

  const snackbar = session.get('snackbar')
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
  snackbar: Partial<SnackbarType>,
  init?: ResponseInit
) {
  const cookie = await flashSnackbar(request, snackbar)
  const newHeaders = new Headers(init?.headers)
  newHeaders.append('Set-Cookie', cookie)

  return redirect(url, {
    ...init,
    headers: newHeaders
  })
}

type JsonWithSnackbarFunction = <Data>(
  request: Request,
  data: Data,
  snackbar: Partial<SnackbarType>,
  init?: number | ResponseInit
) => Promise<DataWithResponseInit<(Data & object) | (Data & null)>>

/**
 * This is an extension of the data function from Remix.
 * This function will add a error snackbar to the response.
 * @param request
 * @param data
 * @param snackbar
 * @param init
 */
export const jsonWithSnackbar: JsonWithSnackbarFunction = async (
  request,
  data,
  snackbar,
  init
) => {
  let responseInit = typeof init === 'number' ? { status: init } : init

  const cookie = await flashSnackbar(request, snackbar)
  const newHeaders = new Headers(responseInit?.headers)
  newHeaders.append('Set-Cookie', cookie)

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
    { ...data },
    {
      ...responseInit,
      headers: newHeaders
    }
  )
}
