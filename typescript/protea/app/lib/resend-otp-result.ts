import type { handleResendOtp } from '~/lib/phone.server'

export type ResendOtpResult = Awaited<ReturnType<typeof handleResendOtp>>

export function isResendRateLimitedResult(
  result: ResendOtpResult | undefined
): result is Extract<ResendOtpResult, { error: 'rateLimited' }> {
  return Boolean(
    result &&
    typeof result === 'object' &&
    'error' in result &&
    result.error === 'rateLimited' &&
    'retryAfter' in result &&
    typeof result.retryAfter === 'number'
  )
}
