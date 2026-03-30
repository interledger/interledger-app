import { Code } from "@bufbuild/connect"
import { ConnectError } from "../error.server"
import { redirectWithSnackbar } from "../snackbar.server"
import logger from "../logger.server"

/**
 * Standard error object that will be used in the BFF layer.
 */
export const UserFacingError = (message: string, status = 500, formErrors?: any) => {
    return { message, status, formErrors }
}
export type UserFacingErrorType = ReturnType<typeof UserFacingError>

/**
 * Maps errors from different clients to UserFacingError to be then treated in a unified way by the handler,
 * regardless of the client they're coming from
 */
type Client = 'grpc'
type ErrorMappingFn = {
    toUserFacingError: (data: any) => UserFacingErrorType
}

export const ErrorMapper: Record<Client, ErrorMappingFn> = {
    // kratosClient
    grpc: {
        toUserFacingError: (data: ConnectError) => {
            logger.error({ connectError: { ...data } }, 'Error from GRPC client.')

            if (data.code == Code.Internal) {
                // When BFF - BE Contract willl be in place,
                // on data.details.message, BE will tell us the exact error
                return UserFacingError("Personal information not available, please contact support.", 500)
            }

            if (data.code == Code.Unavailable) {
                return UserFacingError("Service not available, please try again later.", 500)
            }

            if (data.code == Code.NotFound) {
                return UserFacingError("Not found, please try again.", 404)
            }

            if (data.code == Code.FailedPrecondition) {
                return UserFacingError("There are preconditions that need to be met.", 400, data.violations)
            }

            return UserFacingError("An error occured, please try again.", 500)
        }
    }
}

/**
 * Handles errors by either returning them as is, redirecting with a snackbar, or throwing them.
 */
export const ErrorHandler = (request: Request, ufe: UserFacingErrorType, cb?: () => UserFacingErrorType) => {
    if (cb) {
        return cb();
    }

    if (ufe.status == 400) {
        return ufe
    }

    if ([500].includes(ufe.status)) {
        return redirectWithSnackbar(request, deleteLastPathArgument(request.url), {
            message: ufe.message
        })
    }

    throw ufe
}


















/**
type Fields = {
  email: string
  password: string
}

type LoginError = UserFacingError<Fields>
// => { errors: Record<'email' | 'password', string> }
 */
// export type UserFacingErrorType<T extends Record<string, any> = any> = {
//     errors: Partial<Record<keyof T, any>>
//     status: number
// }
const deleteLastPathArgument = (path: string) => {
    return path.substring(0, path.lastIndexOf('/'));
}


// Utility function to be used in Bff to check for errors in client responses
export const isClientError = (data: any): data is ReturnType<typeof UserFacingError> => {
    return data && typeof data === 'object' && 'status' in data && 'message' in data
}