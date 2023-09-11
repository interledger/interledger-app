import { Code } from '@bufbuild/connect'
import type { ActionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'
import { connectClient } from '~/lib/connect.server'
import { validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { getUserSession } from '~/lib/kratos.server'

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const country = form.get('country') as string
  const phone = form.get('phone') as string

  await validateCSRFToken(request, form)

  // If the phone is missing we assume the user has a session and we can get it from there.
  const validation = form.has('phone')

  const errors = {
    form: '',
    country: '',
    phone: ''
  }
  const mapping = {
    phone: 'To'
  }

  let phoneNumber = ''

  if (validation) {
    try {
      phoneNumber = parsePhoneNumberWithError(
        phone,
        country as CountryCode
      ).number
    } catch (err) {
      switch ((err as ParseError).message) {
        case 'NOT_A_NUMBER':
          errors.phone = 'Phone number is invalid.'
          return error(request, { errors })
        case 'INVALID_COUNTRY':
          errors.phone = 'Country is invalid.'
          return error(request, { errors })
        case 'TOO_SHORT':
          errors.phone = 'Phone number is too short.'
          return error(request, { errors })
        case 'TOO_LONG':
          errors.phone = 'Phone number is too long.'
          return error(request, { errors })
        default:
          throw err
      }
    }
  } else {
    phoneNumber = await getUserSession(request).then(
      (v) => v.identity.traits.phone
    )
  }

  let response = await connectClient.sendPhoneVerification(request, {
    to: phoneNumber
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors }, mapping)
    } else
      return response.error({ errors }, mapping, { action: 'Contact support' })
  }

  return json({ phone: phoneNumber, success: true })
}
