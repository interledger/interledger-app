import { errorMessages } from "~/constants/messages"

export type ErrorCode = 100 | 101 | 400 | 404 | 500 | 503 | 504 | 505 | 506 | 507 | 508 | 510 // or use https://www.npmjs.com/package/http-status-codes

export type Operation = "defaultOperation" |"login" // add business logic operations

export type ErrorMessagesByCode = { [key in ErrorCode]?: string } & { defaultErrorCode: string}
export type OperationErrorMessage = Record<Operation, ErrorMessagesByCode>

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