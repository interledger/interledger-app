import { Code } from '@bufbuild/connect'
import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'
import { href, redirect } from 'react-router'
import { ErrorDescriptions } from '~/lib/error.constants'
import type { TwillioError } from '~/lib/error.mappers'
import { isConnectError, isOtpValidationError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getUserSession } from '~/lib/kratos/session.server'
import {
  RateLimitKeys,
  getKey,
  getRateLimit,
  rateLimit
} from '~/lib/rateLimit.server'

type ParsedPhoneResult =
  | { success: true; phone: string }
  | { success: false; error: string }

const PHONE_OTP_LIMITS = {
  perMinute: { limit: 1, ttlSeconds: 60, retryAfter: 60 },
  perHour: { limit: 3, ttlSeconds: 60 * 60, retryAfter: 60 * 60 }
} as const

type OtpRateLimitResult = {
  error: 'rateLimited'
  retryAfter: number
  message: string
}

const PHONE_OTP_RATE_LIMIT_MESSAGES = {
  perMinute:
    'Please wait 1 minute before requesting another verification code.',
  perHour: 'Too many verification code requests. Please try again in 1 hour.'
} as const

async function getRateLimitIdentity(request: Request): Promise<string> {
  const session = await getUserSession(request)

  if (!session?.identity?.id) {
    throw redirect(href('/login'))
  }

  return session.identity.id
}

async function applyPhoneOtpRateLimit(
  request: Request
): Promise<OtpRateLimitResult | null> {
  const identity = await getRateLimitIdentity(request)
  const hourKey = getKey(RateLimitKeys.PhoneOTP, `${identity}:1h`)
  const minuteKey = getKey(RateLimitKeys.PhoneOTP, `${identity}:1m`)

  // Check first (without incrementing) so we can control message precedence.
  const hourRateLimitError = await getRateLimit(
    hourKey,
    PHONE_OTP_LIMITS.perHour
  )
  if (hourRateLimitError) {
    return {
      error: 'rateLimited',
      retryAfter: PHONE_OTP_LIMITS.perHour.retryAfter,
      message: PHONE_OTP_RATE_LIMIT_MESSAGES.perHour
    }
  }

  const minuteIncrementError = await rateLimit(
    minuteKey,
    PHONE_OTP_LIMITS.perMinute
  )
  if (minuteIncrementError) {
    return {
      error: 'rateLimited',
      retryAfter: PHONE_OTP_LIMITS.perMinute.retryAfter,
      message: PHONE_OTP_RATE_LIMIT_MESSAGES.perMinute
    }
  }

  // Increment both counters only after checks pass.
  // Keep hour first so a race condition prefers the more restrictive bucket.
  const hourIncrementError = await rateLimit(hourKey, PHONE_OTP_LIMITS.perHour)
  if (hourIncrementError) {
    return {
      error: 'rateLimited',
      retryAfter: PHONE_OTP_LIMITS.perHour.retryAfter,
      message: PHONE_OTP_RATE_LIMIT_MESSAGES.perHour
    }
  }
  return null
}

export function parseUserPhone(
  phone: string,
  country: string
): ParsedPhoneResult {
  try {
    const parsedPhone = parsePhoneNumberWithError(
      phone,
      country as CountryCode
    ).number

    return {
      success: true,
      phone: parsedPhone.replace(/[\s\-()]/g, '')
    }
  } catch (err) {
    switch ((err as ParseError).message) {
      case 'NOT_A_NUMBER':
        return { success: false, error: 'Phone number is invalid.' }
      case 'INVALID_COUNTRY':
        return { success: false, error: 'Country is invalid.' }
      case 'TOO_SHORT':
        return { success: false, error: 'Phone number is too short.' }
      case 'TOO_LONG':
        return { success: false, error: 'Phone number is too long.' }
      default:
        throw err
    }
  }
}

/**
 * Handles the `updatePhone` form intent: validates, updates the phone number via
 * gRPC, and sends an OTP to the new number.
 */
export async function handleUpdatePhone(request: Request, form: FormData) {
  const newPhone = form.get('phone') as string
  const country = form.get('country') as string
  const parsedPhone = parseUserPhone(newPhone, country)

  if (!parsedPhone.success) {
    return { errors: { phone: parsedPhone.error } }
  }

  const otpRateLimitResult = await applyPhoneOtpRateLimit(request)
  if (otpRateLimitResult) {
    return {
      errors: {
        phone: otpRateLimitResult.message
      }
    }
  }

  const updateResponse = await grpc.updateUserPhone(request, {
    phone: parsedPhone.phone
  })
  if (isConnectError(updateResponse)) {
    if (updateResponse.code === Code.InvalidArgument) {
      return {
        errors: { phone: 'Invalid phone number. Please check the format.' }
      }
    }
    if (
      updateResponse.code === Code.AlreadyExists ||
      updateResponse.hasAppErrorCode('SIGNUP_DUPLICATE_PHONE')
    ) {
      return {
        errors: {
          phone: 'This mobile number is already in use. Try a different number.'
        }
      }
    }
    if (updateResponse.code === Code.Unauthenticated) {
      throw redirect(href('/logout'))
    }
    throw updateResponse.errorResponse
  }

  const sendResponse = await grpc.sendPhoneVerification(request, {
    to: parsedPhone.phone
  })
  if (isConnectError(sendResponse)) throw sendResponse.errorResponse

  return { codeSent: true as const, phone: parsedPhone.phone }
}

/**
 * Handles the `resend` form intent: rate-limits and re-sends an OTP to `phone`.
 */
export async function handleResendOtp(request: Request, phone: string) {
  const errors: Partial<TwillioError> = { otp: '' }
  const normalizedPhone = phone.replace(/[\s\-()]/g, '')
  const otpRateLimitResult = await applyPhoneOtpRateLimit(request)

  if (otpRateLimitResult) {
    return {
      codeSent: false as const,
      error: 'rateLimited' as const,
      retryAfter: otpRateLimitResult.retryAfter,
      message: otpRateLimitResult.message
    }
  }

  const response = await grpc.sendPhoneVerification(request, {
    to: normalizedPhone
  })

  if (isConnectError(response)) {
    if (response.code === Code.InvalidArgument) {
      return response.error(
        { errors },
        {},
        {
          action: 'Update mobile number',
          message:
            'Your mobile number format is invalid. Please update it and try again.'
        }
      )
    } else {
      return response.error(
        { errors },
        {},
        { action: 'Contact support', message: ErrorDescriptions.DEFAULT }
      )
    }
  }

  return { codeSent: true as const }
}

/**
 * Handles the `verify` form intent: confirms the OTP.
 * Returns `null` on success (caller is responsible for the redirect).
 */
export async function handleVerifyOtp(request: Request, otp: string) {
  const errors: Partial<TwillioError> = { otp: '' }

  const response = await grpc.confirmUserPhone(request, { otp })

  if (isConnectError(response)) {
    if (isOtpValidationError(response)) {
      return response.error(
        { errors },
        {},
        { action: 'Contact support', message: ErrorDescriptions.INVALID_OTP }
      )
    } else if (response.code === Code.InvalidArgument) {
      return response.error({ errors })
    } else {
      return response.error(
        { errors },
        {},
        { action: 'Contact support', message: ErrorDescriptions.DEFAULT }
      )
    }
  }

  return null // success — caller handles redirect
}
