export type SuccessfulActionResponse<T = any> = {
    success: true
    data: T
}

export type UserFacingErrors<T extends Record<string, string> = any> = Partial<Record<keyof T, string>>

export type FailedActionResponse<T = any> = {
    success: false
    errors: UserFacingErrors
}

export type ActionResponse<T = any> = SuccessfulActionResponse<T> | FailedActionResponse<T>
