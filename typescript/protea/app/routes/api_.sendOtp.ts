import type { Route } from './+types/api_.sendOtp'
import { Code } from '@bufbuild/connect'
import { data as rrData } from 'react-router';
import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'
import { validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getUserSession } from '~/lib/kratos.server'

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const country = form.get('country') as string
  const phone = form.get('phone') as string

  await validateCSRFToken(request, form)

  // If the phone is missing we assume the user has a session and we can get it from there.
  const validation = form.has('phone')

  const data = {
    phone: '',
    success: false,
    errors: {
      form: '',
      country: '',
      phone: ''
    }
  }
  const mapping = {
    phone: 'To'
  }

  if (validation) {
    try {
      data.phone = parsePhoneNumberWithError(
        phone,
        country as CountryCode
      ).number
    } catch (err) {
      switch ((err as ParseError).message) {
        case 'NOT_A_NUMBER':
          data.errors.phone = 'Phone number is invalid.'
          return error(request, data)
        case 'INVALID_COUNTRY':
          data.errors.phone = 'Country is invalid.'
          return error(request, data)
        case 'TOO_SHORT':
          data.errors.phone = 'Phone number is too short.'
          return error(request, data)
        case 'TOO_LONG':
          data.errors.phone = 'Phone number is too long.'
          return error(request, data)
        default:
          throw err
      }
    }
  } else {
    data.phone = await getUserSession(request).then(
      (v) => v.identity.traits.phone
    )
  }

  let response = await grpc.sendPhoneVerification(request, {
    to: data.phone
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error(data, mapping)
    } else return response.error(data, mapping, { action: 'Contact support' })
  }

  data.success = true
  return rrData(data)
}
