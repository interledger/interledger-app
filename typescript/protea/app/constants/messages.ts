import type { ErrorMessagesByCode, OperationErrorMessage } from "~/services/error.service"

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
export const errorMessages: OperationErrorMessage = {
  defaultOperation: defaultErrorMessages,
  login: loginErrorMessages
}