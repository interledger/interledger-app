import type { TypedResponse } from '@remix-run/node'
import { json } from '@remix-run/node'
import { captureMessage } from '@sentry/remix'
import { flashSnackbar } from '~/lib/snackbar.server'
import type { SnackbarType } from '~/lib/useScaffoldStore'

type JsonWithErrorFunction = <
  Data extends object | null | (object & { errors: object & { form: string } })
>(
  request: Request,
  data: Data,
  snackbar?: Partial<SnackbarType>,
  init?: number | ResponseInit
) => Promise<TypedResponse<(Data & object) | (Data & null)>>

/**
 * This is an extension of the json function from Remix.
 * This function will set the status of the response to 400, and can flash a snackbar if required.
 * It will also log the error to Sentry.
 * @param request
 * @param data
 * @param snackbar - Whether to show a snackbar on the client. Will use fieldErrors.form if set.
 * @param init
 */
export const error: JsonWithErrorFunction = async (
  request,
  data,
  snackbar,
  init
) => {
  let responseInit = typeof init === 'number' ? { status: init } : init
  const url = new URL(request.url)
  const newHeaders = new Headers(responseInit?.headers)

  if (data && 'errors' in data && 'form' in data.errors && data.errors.form) {
    snackbar = {
      ...snackbar,
      message: data.errors.form
    }
  }

  if (typeof snackbar !== 'undefined') {
    const cookie = await flashSnackbar(request, {
      ...snackbar,
      message:
        snackbar.message ||
        'There was an error with your request. Please retry. If the error continues, contact support.',
      icon: 'close'
    })
    newHeaders.append('Set-Cookie', cookie)
  }

  captureMessage('Invalid form submission', {
    extra: {
      url: url.pathname,
      message:
        snackbar?.message || 'There was a general error with the request.'
    }
  })

  if (typeof data !== 'object') {
    throw json(
      {},
      {
        status: 400,
        statusText: 'Only objects should be returned from loaders.'
      }
    )
  }

  return json(data, {
    ...responseInit,
    headers: newHeaders,
    status: 400
  })
}
