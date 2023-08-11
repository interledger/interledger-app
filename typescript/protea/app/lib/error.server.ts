import type { TypedResponse } from '@remix-run/node'
import { json } from '@remix-run/node'
import { captureMessage } from '@sentry/remix'
import { flashSnackbar } from '~/lib/snackbar.server'
import type { SnackbarAction } from '~/lib/useScaffoldStore'

type JsonWithErrorFunction = <T extends { form?: string }>(
  request: Request,
  fieldErrors: T,
  showSnackbar?: boolean,
  snackbarAction?: SnackbarAction,
  init?: number | ResponseInit
) => Promise<TypedResponse<{ errors: T }>>

// NOTE: This doesn't work with the typings on actiondata
// https://github.com/epicweb-dev/epic-stack/blob/main/app/routes/resources%2B/login.tsx#L175
/**
 * This is an extension of the json function from Remix.
 * This function will set the status of the response to 400, and can flash a snackbar if required.
 * It will also log the error to Sentry.
 * Should only be used for errors we want to surface to the user.
 * @param request
 * @param fieldErrors
 * @param showSnackbar - Whether to show a snackbar on the client. Will use fieldErrors.form if set.
 * @param snackbarAction - Override the action to show on the snackbar. Default is undefined.
 * @param init
 */
export const error: JsonWithErrorFunction = async (
  request,
  fieldErrors,
  showSnackbar,
  snackbarAction,
  init
) => {
  let responseInit = typeof init === 'number' ? { status: init } : init
  const newHeaders = new Headers(responseInit?.headers)

  if (fieldErrors.form || showSnackbar) {
    const url = new URL(request.url)
    captureMessage('Invalid form submission', {
      extra: {
        url: url.pathname,
        message:
          fieldErrors.form || 'There was a general error with the request.'
      }
    })
  }

  if (fieldErrors.form || showSnackbar) {
    const cookie = await flashSnackbar(request, {
      message:
        fieldErrors.form ||
        'There was an error with your request. Please retry. If the error continues, contact support.',
      action: snackbarAction,
      icon: 'close'
    })

    newHeaders.append('Set-Cookie', cookie)
  }

  if (typeof fieldErrors !== 'object') {
    throw json(
      {},
      {
        status: 400,
        statusText: 'Only objects should be returned from loaders.'
      }
    )
  }

  return json(
    { errors: { ...fieldErrors } },
    {
      ...responseInit,
      headers: newHeaders,
      status: 400
    }
  )
}
