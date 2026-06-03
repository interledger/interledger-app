import { Code } from '@bufbuild/connect'
import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'
import { href, redirect } from 'react-router'
import { ErrorDescriptions } from '~/lib/error.constants'
import type { TwillioError } from '~/lib/error.mappers'
import { isConnectError, isOtpValidationError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { RateLimitKeys, getKey, rateLimit } from '~/lib/rateLimit.server'

type ParsedPhoneResult =
  | { success: true; phone: string }
  | { success: false; error: string }

export function parseUserPhone(
  phone: string,
  country: string
): ParsedPhoneResult {
  try {
    return {
      success: true,
      phone: parsePhoneNumberWithError(phone, country as CountryCode).number
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
  const rateLimitError = await rateLimit(
    getKey(RateLimitKeys.PhoneOTP, phone),
    { limit: 1, ttlSeconds: 60 }
  )
  if (rateLimitError) {
    return {
      codeSent: false as const,
      error: 'rateLimited' as const,
      retryAfter: 60
    }
  }

  const response = await grpc.sendPhoneVerification(request, { to: phone })
  if (isConnectError(response)) throw response.errorResponse

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
