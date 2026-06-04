export const ErrorCodes = {
  OTP_EXPIRED: 'OTP_EXPIRED',
    USER_INVALIDATED: 'USER_INVALIDATED'
} as const

export type ErrorCode = keyof typeof ErrorCodes
