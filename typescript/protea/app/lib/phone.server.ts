import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'

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
