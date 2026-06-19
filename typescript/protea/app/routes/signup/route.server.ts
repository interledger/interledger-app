import { Code } from '@bufbuild/connect'
import type { SuccessfulNativeRegistration } from '@ory/client'
import { data, href, redirect } from 'react-router'
import { envValue } from '~/env.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { ErrorDescriptions } from '~/lib/error.constants'
import { error, isConnectError, isOtpValidationError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import {
  buildHeadersWithCookies,
  getCookie,
  withCookie
} from '~/lib/kratos/cookie.server'
import {
  handleFlowError,
  mapFlowToFieldErrors
} from '~/lib/kratos/error.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { requireNoUserSession } from '~/lib/kratos/session.server'
import logger from '~/lib/logger.server'
import { parseUserPhone } from '~/lib/phone.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { isEUCountry } from '~/routes/signup/About'
import type { Route } from './+types/route'

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)

  await requireNoUserSession(request)

  const response = await grpc.getCountries(request, {})

  if (isConnectError(response)) throw response.errorResponse

  const countries = response.countries

  // Init kratos flow once per signup flow
  const cookie = getCookie(request)
  let flowResponse
  try {
    flowResponse = await kratosPublic.createBrowserRegistrationFlow(
      { returnTo: url.searchParams.get('returnTo') ?? undefined },
      withCookie(cookie)
    )
  } catch (err: any) {
    handleFlowError(err, 'signup')
    logger.error(
      { error: err, route: 'signup' },
      'Failed to initialize signup flow'
    )
    throw new Error('Failed to initialize signup flow')
  }

  return jsonWithCSRF(
    request,
    {
      countries,
      kratosFlowId: flowResponse.data.id,
      kratosCsrfToken: getCsrfTokenFromFlow(flowResponse.data),
      fynbosEnv: envValue('FYNBOS_ENV')
    },
    {
      headers: buildHeadersWithCookies(flowResponse)
    }
  )
}

export async function detailsAction({ request }: Route.ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const actionData = {
    id: '',
    firstName: '',
    lastName: '',
    email: '',
    phone: '',
    errors: {
      form: '',
      firstName: '',
      lastName: '',
      country: '',
      email: '',
      phone: ''
    }
  }
  const mapping = { country: 'CountryCode', phone: 'MobileNumber' }

  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const country = form.get('country') as string
  const email = form.get('email') as string
  const phone = form.get('phone') as string

  if (
    !(
      country == 'CA' ||
      country == 'US' ||
      country == 'ZA' ||
      isEUCountry(country)
    )
  ) {
    return redirect(
      `/waitlist?country=${country}&email=${email}&fullName=${firstName} ${lastName}`
    )
  }

  const parsedPhone = parseUserPhone(phone, country)
  if (!parsedPhone.success) {
    actionData.errors.phone = parsedPhone.error
    return error(request, actionData)
  }

  const signupResponse = await grpc.setSignupUserData(request, {
    firstName,
    lastName,
    countryCode: country,
    email,
    mobile: parsedPhone.phone
  })

  if (isConnectError(signupResponse)) {
    if (signupResponse.code == Code.InvalidArgument) {
      return signupResponse.error(actionData, mapping)
    } else
      return signupResponse.error(actionData, mapping, {
        action: 'Contact support'
      })
  }

  return data({
    id: signupResponse.id,
    firstName,
    lastName,
    email,
    phone: parsedPhone.phone,
    errors: actionData.errors
  })
}

export async function otpAction({ request }: Route.ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const actionData = {
    id: '',
    phone: '',
    errors: {
      form: '',
      otp: '',
      phone: ''
    }
  }
  const mapping = {
    phone: 'MobileNumber'
  }

  const id = form.get('id') as string
  const otp = form.get('otp') as string
  const phone = form.get('phone') as string

  logger.info(
    {
      id,
      phone,
      hasOtp: !!otp,
      flow: 'signup'
    },
    'Setting mobile number for signup'
  )

  const response = await grpc.setSignupMobileNumber(request, {
    id,
    mobile: phone,
    otp: otp
  })

  if (isConnectError(response)) {
    logger.error(
      {
        code: response.code,
        hasPhone: !!phone,
        flow: 'signup'
      },
      'Failed to set mobile number'
    )

    if (isOtpValidationError(response)) {
      actionData.errors.otp = ErrorDescriptions.INVALID_OTP
      return error(request, actionData)
    } else if (response.code == Code.InvalidArgument) {
      actionData.errors.phone = 'Mobile phone number is invalid.'
      return error(request, actionData)
    } else if (response.code == Code.AlreadyExists) {
      actionData.errors.phone = 'Mobile phone number is already registered.'
      return error(request, actionData)
    } else {
      return response.error(actionData, mapping, { action: 'Contact support' })
    }
  }

  logger.info(
    { id, flow: 'signup' },
    'Mobile number set successfully for signup'
  )
  return data({ id, phone, errors: actionData.errors })
}

const KratosErrorTraits = {
  PHONE: 'traits.phone'
}
const KratosErrorMessages = {
  [KratosErrorTraits.PHONE]: 'Mobile phone number is invalid.'
}

export async function passwordAction({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const id = form.get('id') as string
  const kratosFlowId = form.get('kratosFlowId') as string
  const kratosCsrfToken = form.get('csrf_token') as string
  const password = form.get('password') as string
  const confirmPassword = form.get('confirm-password') as string
  const serviceAgreement = form.get('service-agreement') as string
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const country = form.get('country') as string
  const email = form.get('email') as string
  const phone = form.get('phone') as string
  const cookie = getCookie(request)

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    serviceAgreement: '',
    password: '',
    phone: ''
  }

  if (serviceAgreement == null) {
    errors.serviceAgreement = 'You are required to agree to continue.'
    return error(request, { errors })
  }

  if (password !== confirmPassword) {
    errors.password = 'Passwords do not match.'
    return error(request, { errors })
  }

  const kratosRequestPayload = {
    method: 'password' as const,
    traits: {
      email,
      phone,
      firstName,
      lastName,
      countryCode: country
    },
    password,
    csrf_token: kratosCsrfToken
  }
  logger.info(
    {
      flowId: kratosFlowId,
      countryCode: kratosRequestPayload.traits.countryCode,
      phone: kratosRequestPayload.traits.phone ? 'present' : 'not present',
      hasTraits: !!kratosRequestPayload.traits,
      hasPassword: !!password,
      hasCsrfToken: !!kratosCsrfToken,
      flow: 'signup'
    },
    'Sending registration request to Kratos'
  )

  let kratosResponse
  try {
    kratosResponse = await kratosPublic.updateRegistrationFlow(
      {
        flow: kratosFlowId,
        updateRegistrationFlowBody: kratosRequestPayload
      },
      withCookie(cookie)
    )
    logger.info(
      {
        status: kratosResponse.status,
        statusText: kratosResponse.statusText,
        headers: kratosResponse.headers?.['set-cookie'],
        flow: 'signup'
      },
      'Kratos registration response'
    )
  } catch (err: any) {
    logger.error(
      {
        status: err.response?.status,
        statusText: err.response?.statusText,
        flow: 'signup'
      },
      'Kratos registration error'
    )
    const flowData = err.response?.data
    const errs = mapFlowToFieldErrors(flowData, errors)
    if ((errs as any)[KratosErrorTraits.PHONE]) {
      errors.phone = KratosErrorMessages[KratosErrorTraits.PHONE]
      return error(request, { errors })
    }
    return error(request, { errors: errs })
  }

  const successData = kratosResponse.data as SuccessfulNativeRegistration
  logger.info(
    {
      identityId: successData.identity.id,
      signupId: id,
      flow: 'signup'
    },
    'Kratos registration successful'
  )

  logger.info(
    { id, userId: successData.identity.id, flow: 'signup' },
    'Completing signup in backend'
  )
  await grpc.completeSignup(request, {
    id,
    userId: successData.identity.id
  })

  return redirectWithSnackbar(
    request,
    href('/verify'),
    {
      message: 'Your account was created successfully.',
      icon: 'close'
    },
    {
      headers: buildHeadersWithCookies(kratosResponse)
    }
  )
}
