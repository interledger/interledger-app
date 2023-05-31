// The field names given by the backend for field violations
import type { ActionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { CountryCode, ParseError } from 'libphonenumber-js'
import { parsePhoneNumberWithError } from 'libphonenumber-js'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'

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

  const fieldErrors = {
    form: '',
    country: '',
    phone: ''
  }

  let phoneNumber = ''
  try {
    phoneNumber = parsePhoneNumberWithError(
      phone,
      country as CountryCode
    ).number
  } catch (error) {
    switch ((error as ParseError).message) {
      case 'NOT_A_NUMBER':
        return json(
          { errors: { phone: 'Phone number is invalid.' } },
          { status: 400 }
        )
      case 'INVALID_COUNTRY':
        return json(
          { errors: { phone: 'Country is invalid.' } },
          { status: 400 }
        )
      case 'TOO_SHORT':
        return json(
          { errors: { phone: 'Phone number is too short.' } },
          { status: 400 }
        )
      case 'TOO_LONG':
        return json(
          { errors: { phone: 'Phone number is too long.' } },
          { status: 400 }
        )
      default:
        throw error
    }
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
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  return json({ phone: phoneNumber, success: true })
}
