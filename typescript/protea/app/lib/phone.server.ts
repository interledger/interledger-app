import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'

type ParsedPhoneResult =
  | { phone: string; error?: never }
  | { phone?: never; error: string }

export function parseUserPhone(
  phone: string,
  country: string
): ParsedPhoneResult {
  try {
    return {
      phone: parsePhoneNumberWithError(phone, country as CountryCode).number
    }
  } catch (err) {
    switch ((err as ParseError).message) {
      case 'NOT_A_NUMBER':
        return { error: 'Phone number is invalid.' }
      case 'INVALID_COUNTRY':
        return { error: 'Country is invalid.' }
      case 'TOO_SHORT':
        return { error: 'Phone number is too short.' }
      case 'TOO_LONG':
        return { error: 'Phone number is too long.' }
      default:
        throw err
    }
  }
}
