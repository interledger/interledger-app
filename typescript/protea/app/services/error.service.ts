export type ErrorCode = 100 | 101 | 400 | 404 | 500 | 503 | 504 | 505 | 506 | 507 | 508 | 510 // or use https://www.npmjs.com/package/http-status-codes

export type Operation = "defaultOperation" |"login" // add business logic operations

// Mapping of error codes to messages
export type ErrorMessagesByCode = { [key in ErrorCode]?: string } & { defaultErrorCode: string}
export type OperationErrorMessage = Record<Operation, ErrorMessagesByCode>


// Example usage
const loginErrorMessages: ErrorMessagesByCode = {
    defaultErrorCode: "a login error occurred",
    503: "service unavailable",
    504: "gateway timeout",
    505: "HTTP version not supported",
    506: "variant also negotiates",
    507: "insufficient storage",
    508: "loop detected",
    510: "not extended",
}

const defaultErrorMessages: ErrorMessagesByCode = {
    defaultErrorCode: "an error occurred",
    100: "continue",
    400: "bad request",
    404: "not found",
    500: "internal server error",
}


// Centralized error messages for all operations
const errorMessages: OperationErrorMessage = {
    defaultOperation: defaultErrorMessages,
    login: loginErrorMessages
}


export const ErrorMessageService = {
    getErrorMessage: (operation?: Operation, errorCode?: ErrorCode) => {
        const safeOperation: Operation = operation ?? "defaultOperation"
        const safeErrorCode: ErrorCode | "defaultErrorCode" = errorCode ?? "defaultErrorCode"

        return errorMessages[safeOperation][safeErrorCode];
    }
}

/** Example usage */
const errorMessage = ErrorMessageService.getErrorMessage("login", 503);
console.log(errorMessage);