import { UserFacingErrorType } from './bff-error'

export type SuccessfulServerResponse<T = any> = {
  success: true
  data: T
}
export type FailedServerResponse = {
  success: false
  error: UserFacingErrorType
}

export type ServerResponse<T = any> =
  | SuccessfulServerResponse<T>
  | FailedServerResponse
  | Promise<Response>
