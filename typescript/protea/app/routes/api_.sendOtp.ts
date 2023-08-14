// The field names given by the backend for field violations
import type { ActionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'
import { validateCSRFToken } from '~/lib/csrf.server'
import { error } from '~/lib/error.server'
import { getUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import { StatusError, grpcClient, isGrpcError } from '~/lib/proto.server'

type fieldErrorsMap = 'To'

function mapper(field: fieldErrorsMap): 'phone' | null {
  switch (field) {
    case 'To':
      return 'phone'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const country = form.get('country') as string
  const phone = form.get('phone') as string

  await validateCSRFToken(request, form)

  // If the phone is missing we assume the user has a session and we can get it from there.
  const validation = form.has('phone')

  const fieldErrors = {
    form: '',
    country: '',
    phone: ''
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
          fieldErrors.phone = 'Phone number is invalid.'
          return error(request, { errors: { ...fieldErrors } })
        case 'INVALID_COUNTRY':
          fieldErrors.phone = 'Country is invalid.'
          return error(request, { errors: { ...fieldErrors } })
        case 'TOO_SHORT':
          fieldErrors.phone = 'Phone number is too short.'
          return error(request, { errors: { ...fieldErrors } })
        case 'TOO_LONG':
          fieldErrors.phone = 'Phone number is too long.'
          return error(request, { errors: { ...fieldErrors } })
        default:
          throw err
      }
    }
  } else {
    phoneNumber = await getUserSession(request).then(
      (v) => v.identity.traits.phone
    )
  }

  let response = await grpcClient
    .sendPhoneVerification({
      to: phoneNumber
    })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    if (response.code == 3) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return error(request, { errors: { ...fieldErrors } })
    } else
      return error(
        request,
        { errors: { ...fieldErrors } },
        { action: 'Contact support' }
      )
  }

  return json({ phone: phoneNumber, success: true })
}
