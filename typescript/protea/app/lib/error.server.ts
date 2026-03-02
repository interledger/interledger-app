import { Code } from '@bufbuild/connect'
import type { ConnectError as BufConnectError } from '@bufbuild/connect'
import type { PlainMessage } from '@bufbuild/protobuf'
import { data as rrData, redirect, UNSAFE_DataWithResponseInit as DataWithResponseInit } from 'react-router';
import { captureMessage } from '@sentry/react-router'
import { href } from 'react-router'
import type {
  BadRequest_FieldViolation,
  PreconditionFailure_Violation
} from '~/generated/connect/google/rpc/error_details_pb'
import {
  BadRequest,
  ErrorInfo,
  PreconditionFailure
} from '~/generated/connect/google/rpc/error_details_pb'
import { flashSnackbar } from '~/lib/snackbar.server'
import type { SnackbarType } from '~/lib/useScaffoldStore'

type JsonWithErrorFunction = <
  Data extends
    | null
    | (Record<string, unknown> & Record<'errors', Record<string, string>>)
>(
  request: Request,
  data: Data,
  snackbar?: Partial<SnackbarType>,
  init?: number | ResponseInit,
  isConnectError?: boolean
) => Promise<DataWithResponseInit<Data & object | Data & null>>

type JsonWithConnectErrorFunction = <
  Data extends Record<string, unknown> &
    Record<'errors', Record<string, string>>
>(
  data: Data,
  mapping?: Partial<Data['errors']>,
  snackbar?: Partial<SnackbarType>,
  init?: number | ResponseInit
) => Promise<DataWithResponseInit<Data & object | Data & null>>

/**
 * This is an extension of the data function from Remix.
 * This function will set the status of the response to 400, and can flash a snackbar if required.
 * It will also log the error to Sentry.
 * @param request
 * @param data
 * @param snackbar - Whether to show a snackbar on the client. Will use fieldErrors.form if set.
 * @param init
 * @param isConnectError - Whether the error is a ConnectError. This is used to determine whether to log the error to Sentry.
 */
export const error: JsonWithErrorFunction = async (
  request,
  data,
  snackbar,
  init,
  isConnectError = false
) => {
  let responseInit = typeof init === 'number' ? { status: init } : init
  const newHeaders = new Headers(responseInit?.headers)

  if (!isConnectError) {
    const url = new URL(request.url)
    captureMessage('Non GRPC error returned to user', {
      extra: {
        url: url.pathname,
        data
      }
    })
  }

  if (
    data &&
    'errors' in data &&
    typeof data.errors === 'object' &&
    data.errors &&
    'form' in data.errors &&
    data.errors.form &&
    typeof data.errors.form === 'string'
  ) {
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

  if (typeof data !== 'object') {
    throw rrData(
      {},
      {
        status: 400,
        statusText: 'Only objects should be returned from loaders.'
      }
    )
  }

  return rrData(data, {
    ...responseInit,
    headers: newHeaders,
    status: 400
  })
}

/**
 * ConnectError is an extension of the ConnectError from @bufbuild/connect.
 *
 * It doesn't explicitly extend the class, because we don't need all the
 * properties and methods it provides.
 *
 * https://github.com/connectrpc/connect-es/blob/main/packages/connect/src/connect-error.ts
 */
export class ConnectError {
  readonly _request: Request
  readonly _err: BufConnectError
  readonly code: Code
  // The HTTP status code for this error.
  readonly status: ResponseInit | undefined
  readonly errorResponse: DataWithResponseInit<Record<string, never>> | undefined

  violations: PlainMessage<PreconditionFailure_Violation>[] = []
  fieldViolations: PlainMessage<BadRequest_FieldViolation>[] = []

  constructor(request: Request, err: BufConnectError) {
    this._request = request
    this._err = err
    this.code = err.code

    const mappedStatus = httpMapping(err.code)
    this.status = mappedStatus
    this.errorResponse = rrData({}, mappedStatus)

    const url = new URL(request.url)

    captureMessage('Error received in connect client', {
      extra: {
        url: url.pathname,
        code: err.code,
        cause: err.cause
      }
    })

    if (err.code === Code.Unauthenticated) {
      url.searchParams.set('returnTo', url.pathname + url.search)

      const errorDetails = err.findDetails(ErrorInfo)
      const reason = errorDetails?.[0]?.reason
      if (reason === 'aal2_required') {
        throw redirect(href('/totp/challenge') + url.search)
      }

      throw redirect(href('/login') + url.search, {
        headers: {
          'Set-Cookie':
            'ory_kratos_session=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT'
        }
      })
    }

    // If Code.FailedPrecondition, store violations
    if (err.code === Code.FailedPrecondition) {
      this.violations = err
        .findDetails(PreconditionFailure)
        .reduce(
          (
            accumulator: PlainMessage<PreconditionFailure_Violation>[],
            currentValue: PlainMessage<PreconditionFailure>
          ) => {
            currentValue.violations.forEach((val: PlainMessage<PreconditionFailure_Violation>) => accumulator.push(val))
            return accumulator
          },
          []
        )
    }

    // If Code.BadRequest, store fieldViolations
    if (err.code === Code.InvalidArgument) {
      this.fieldViolations = err
        .findDetails(BadRequest)
        .reduce(
          (
            accumulator: PlainMessage<BadRequest_FieldViolation>[],
            currentValue: PlainMessage<BadRequest>
          ) => {
            currentValue.fieldViolations.forEach((val: PlainMessage<BadRequest_FieldViolation>) => accumulator.push(val))
            return accumulator
          },
          []
        )
    }
  }

  /**
   * Helper to return error (function above) with mapped field errors.
   * @param data
   * @param mapping
   * @param snackbar
   * @param init
   */
  public error: JsonWithConnectErrorFunction = async (
    data,
    mapping,
    snackbar,
    init
  ) => {
    data.errors = this.fieldErrors(data.errors, mapping)
    return error(this._request, data, snackbar, init, true)
  }

  /**
   * Returns a new object with the BadRequest_FieldViolations mapped to the passed in fieldErrors.
   * @param fieldErrors
   * @param mapping An object that allows overriding fieldErrors with a different key. This is useful when the field name on the backend is different from the key in the fieldErrors object.
   */
  private fieldErrors<T extends Record<string, string>>(
    fieldErrors: T,
    mapping?: Partial<T>
  ): T {
    if (this.fieldViolations.length == 0) return fieldErrors

    const fieldNames: Record<string, string> = {}

    Object.keys(fieldErrors).forEach(
      (key) => (fieldNames[key.toLowerCase()] = key)
    )

    if (mapping) {
      Object.entries(mapping).forEach(
        ([key, value]) => (fieldNames[value.toLowerCase()] = key)
      )
    }

    return this.fieldViolations.reduce((accumulator, current) => {
      let fieldName = fieldNames[current.field.toLowerCase()]
      if (fieldName) {
        ;(accumulator as Record<string, string>)[fieldName] =
          current.description
      }
      return accumulator
    }, fieldErrors)
  }
}

/**
 * httpMapping maps the grpc error code to an HTTP status code.
 * @param code Connect error code
 */
function httpMapping(code: Code): ResponseInit | undefined {
  switch (code) {
    case Code.InvalidArgument:
    case Code.FailedPrecondition:
    case Code.OutOfRange:
      return { status: 400, statusText: 'Bad Request' }
    case Code.Unauthenticated:
      return { status: 401, statusText: 'Unauthorized' }
    case Code.PermissionDenied:
      return { status: 403, statusText: 'Forbidden' }
    case Code.NotFound:
      return { status: 404, statusText: 'Not Found' }
    case Code.AlreadyExists:
    case Code.Aborted:
      return { status: 409, statusText: 'Conflict' }
    case Code.ResourceExhausted:
      return { status: 429, statusText: 'Too Many Requests' }
    case Code.Canceled:
      return { status: 499, statusText: 'Client Closed Request' }
    case Code.Unknown:
    case Code.Internal:
    case Code.DataLoss:
      return { status: 500, statusText: 'Internal Server Error' }
    case Code.Unimplemented:
      return { status: 501, statusText: 'Not Implemented' }
    case Code.Unavailable:
      return { status: 503, statusText: 'Service Unavailable' }
    case Code.DeadlineExceeded:
      return { status: 504, statusText: 'Gateway Timeout' }
  }
}

/**
 * Type guard for ConnectError
 * @param response The response from the grpc call using the connect client.
 */
export function isConnectError(response: unknown): response is ConnectError {
  // NOTE `response instanceof ConnectError` doesn't work here because the class
  //  is only instantiated once when the grpc client is created.
  return (
    (response as ConnectError).code !== undefined &&
    (response as ConnectError).status !== undefined &&
    (response as ConnectError)._err !== undefined &&
    (response as ConnectError)._request !== undefined
  )
}

export function isTwilioCodeError(error: ConnectError): boolean {
  return error.fieldViolations.some((violation) => violation.field === 'Code')
}

export function isTwilioError(error: ConnectError): boolean {
  const errorInfo = error._err.findDetails(ErrorInfo)
  return errorInfo.some((info: PlainMessage<ErrorInfo>) => info.reason === 'TwilioError')
}
