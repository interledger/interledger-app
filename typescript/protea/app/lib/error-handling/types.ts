type FormError = Record<string, string>
type UserFacingErrors<T extends FormError = any> = Partial<Record<keyof T, string>>

export type SuccessfulServerResponse<T = any> = {
    success: true
    data: T
}
export type FailedServerResponse<T extends FormError = any> = {
    success: false
    message?: string
    errors?: UserFacingErrors<T>
}

export type ServerResponse<T extends FormError = any> =
    SuccessfulServerResponse<T> |
    FailedServerResponse<T>
