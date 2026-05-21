import { Code } from '@bufbuild/connect'
import { data as rrData } from 'react-router'
import { validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getSessionTraits, getUserSession } from '~/lib/kratos/session.server'
import { parseUserPhone } from '~/lib/phone.server'
import type { Route } from './+types/api_.sendOtp'

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
    const parsedPhone = parseUserPhone(phone, country)
    if (!parsedPhone.success) {
      data.errors.phone = parsedPhone.error
      return error(request, data)
    }

    data.phone = parsedPhone.phone
  } else {
    const session = await getUserSession(request)
    const sessionPhone = getSessionTraits(session).phone

    if (!sessionPhone) {
      data.errors.phone = 'Phone number is required.'
      return error(request, data)
    }

    data.phone = sessionPhone
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
