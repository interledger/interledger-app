import { Code } from "@bufbuild/connect"
import type { ConnectError } from '../error.server'
import { PlaidError } from '../plaid.server'
import logger from '../logger.server'
import { redirectWithSnackbar } from '../snackbar.server'
import type { FailedServerResponse } from './types'

/**
 * Standard error object that will be used in the BFF layer.
 */
export const UserFacingError = (
  message: string,
  status = 500,
  formErrors?: any
) => {
  return { message, status, formErrors }
}
export type UserFacingErrorType = ReturnType<typeof UserFacingError>

/**
 * Maps errors from different clients to UserFacingError to be then treated in a unified way by the handler,
 * regardless of the client they're coming from
 */
type Client = 'grpc' | 'plaid'
type ErrorMappingFn = {
  toUserFacingError: (data: any) => UserFacingErrorType
}

export const ErrorMapper: Record<Client, ErrorMappingFn> = {
  // kratosClient
  grpc: {
    toUserFacingError: (data: ConnectError) => {
      logger.error({ connectError: { ...data } }, 'Error from GRPC client.')

      if (data.code == Code.Unavailable) {
        // When BFF - BE Contract willl be in place,
        // on data.details.message, BE will tell us the exact error
        return UserFacingError(
          'Service not available, please try again later.',
          500
        )
      }

      if (data.code == Code.NotFound) {
        return UserFacingError('Not found, please try again.', 404)
      }

      if (data.code == Code.FailedPrecondition) {
        return UserFacingError(
          'There are preconditions that need to be met.',
          400,
          data.violations
        )
      }

      return UserFacingError("An error occured, please try again.", 500)
    }
  },
  plaid: {
    toUserFacingError: (data: PlaidError) => {
      logger.error({ plaidError: { status: data.status, errorCode: data.errorCode, message: data.message } }, 'Error from Plaid client.')
      return UserFacingError(data.message || 'Plaid error occurred', data.status)
    }
  }
}

/**
 * Handles errors by either returning them as is, redirecting with a snackbar, or throwing them.
 */
export const ErrorHandler = (
  request: Request,
  ufe: UserFacingErrorType,
  opts?: { cb?: () => FailedServerResponse; alwaysReturnUfe?: boolean }
): FailedServerResponse | Promise<Response> => {
  if (opts?.cb) {
    return opts.cb()
  }

  if (opts?.alwaysReturnUfe) {
    return {
      success: false,
      error: ufe
    }
  }

  if (ufe.status == 400) {
    return {
      success: false,
      error: ufe
    }
  }

  if ([500].includes(ufe.status)) {
    return redirectWithSnackbar(request, deleteLastPathArgument(request.url), {
      message: ufe.message
    })
  }

  throw ufe
}

const deleteLastPathArgument = (path: string) => {
  return path.substring(0, path.lastIndexOf('/'))
}
