const ErrorCodes = {
  OTP_EXPIRED: 'OTP_EXPIRED',
  USER_INVALIDATED: 'USER_INVALIDATED'
} as const

type ErrorCode = keyof typeof ErrorCodes
