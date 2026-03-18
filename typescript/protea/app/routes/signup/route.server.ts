import type { ActionFunctionArgs, LoaderFunctionArgs } from 'react-router';
import { data as rrData, redirect } from 'react-router';

import { Code } from '@bufbuild/connect'
import type { SuccessfulSelfServiceRegistrationWithoutBrowser } from '@ory/kratos-client'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { trimHeaders } from '~/lib/headers.server'
import logger from '~/lib/logger.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireNoUserSession
} from '~/lib/kratos.server'
import { fromKratosResponse, sendBffError } from '~/lib/error-mapper.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { isEUCountry } from '~/routes/signup/About'
import { href } from 'react-router'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)

  await requireNoUserSession(request)

  let response = await grpc.getCountries(request, {})

  if (isConnectError(response)) throw response.errorResponse

  const countries = response.countries

  // Init kratos flow once per signup flow
  const flowRes = await fetch(
    `${KRATOS_URL}/self-service/registration/browser?${url.searchParams}`,
    { headers: { Accept: 'application/json' } }
  )
  const kratosFlow = await flowRes.json()
  if (flowRes.status >= 400) handleFlowError(kratosFlow, 'signup')

  return jsonWithCSRF(
    request,
    {
      countries,
      kratosFlowId: kratosFlow.id,
      kratosCsrfToken: getCsrfTokenFromFlow(kratosFlow),
      fynbosEnv: process.env.FYNBOS_ENV
    },
    {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    }
  )
}

export async function action(args: ActionFunctionArgs) {
  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string

  if (formName === 'details') {
    return detailsAction(args)
  }

  if (formName === 'otp') {
    return otpAction(args)
  }

  if (formName === 'password') {
    return passwordAction(args)
  }

  throw rrData(
    { title: "Submitted a form that doesn't exist" },
    {
      status: 400
    }
  )
}

export async function detailsAction({ request }: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const data = {
    id: '',
    firstName: '',
    lastName: '',
    email: '',
    errors: {
      form: '',
      firstName: '',
      lastName: '',
      country: '',
      email: ''
    }
  }
  const mapping = { country: 'CountryOfResidence' }

  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const country = form.get('country') as string
  const email = form.get('email') as string

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

  let response = await grpc.setSignupUserData(request, {
    firstName,
    lastName,
    countryCode: country,
    email
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error(data, mapping)
    } else return response.error(data, mapping, { action: 'Contact support' })
  }

  return rrData({
    id: response.id,
    firstName,
    lastName,
    email,
    errors: data.errors
  })
}

export async function otpAction({ request }: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const data = {
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

  logger.info({
    id,
    phone,
    hasOtp: !!otp,
    flow: 'signup'
  }, 'Setting mobile number for signup')

  const response = await grpc.setSignupMobileNumber(request, {
    id,
    mobile: phone,
    otp: otp
  })

  if (isConnectError(response)) {
    logger.error({
      code: response.code,
      hasPhone: !!phone,
      flow: 'signup'
    }, '[SIGNUP] Failed to set mobile number')

    if (response.code == Code.InvalidArgument) {
      data.errors.phone = 'Mobile phone number is invalid.'
      return response.error(data, mapping)
    } else if (response.code == Code.AlreadyExists) {
      data.errors.phone = 'Mobile phone number is already registered.'
      return response.error(data)
    } else {
      return response.error(data, mapping, { action: 'Contact support' })
    }
  }

  logger.info({ id, flow: 'signup' }, '[SIGNUP] Mobile number set successfully for signup')

  return rrData({ id, phone, errors: data.errors })
}

const KratosErrorTraits = {
  PHONE: 'traits.phone'
}
const KratosErrorMessages = {
  [KratosErrorTraits.PHONE]: 'Mobile phone number is invalid.'
}

export async function passwordAction({ request }: ActionFunctionArgs) {
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
    method: 'password',
    traits: {
      email: email,
      phone: phone,
      firstName: firstName,
      lastName: lastName,
      countryCode: country
    },
    password: password,
    csrf_token: kratosCsrfToken
  }

  logger.info({
    url: `${KRATOS_URL}/self-service/registration?flow=${kratosFlowId}`,
    flowId: kratosFlowId,
    countryCode: kratosRequestPayload.traits.countryCode,
    hasTraits: !!kratosRequestPayload.traits,
    hasPassword: !!password,
    hasCsrfToken: !!kratosCsrfToken,
    flow: 'signup'
  }, '[SIGNUP] Sending registration request to Kratos')

  const response = await fetch(
    `${KRATOS_URL}/self-service/registration?flow=${kratosFlowId}`,
    {
      method: 'POST',
      body: JSON.stringify(kratosRequestPayload),
      headers: {
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )
  logger.info({
    status: response.status,
    statusText: response.statusText,
    headers: Object.fromEntries(trimHeaders(response.headers, ['set-cookie']).entries()),
    flow: 'signup'
  }, '[SIGNUP] Kratos registration response')

  if (response.status >= 400) {
    const responseText = await response.clone().text()
    logger.error({
      status: response.status,
      statusText: response.statusText,
      flowId: kratosFlowId,
      responseBody: responseText,
      requestTraits: kratosRequestPayload?.traits
        ? Object.keys(kratosRequestPayload.traits)
        : undefined,
      flow: 'signup'
    }, '[SIGNUP] Kratos registration failed')

    try {
      const responseJson = JSON.parse(responseText)
      logger.error({
        ui: responseJson.ui,
        messages: responseJson.ui?.messages,
        nodes: responseJson.ui?.nodes,
        flow: 'signup'
      }, '[SIGNUP] Kratos error details')
    } catch (e) {
      logger.error({ flow: 'signup' }, '[SIGNUP] Could not parse Kratos error response as JSON')
    }
    const bffError = await fromKratosResponse(response)
    if (bffError.formErrors?.[KratosErrorTraits.PHONE]) {
      delete bffError.formErrors[KratosErrorTraits.PHONE]
      bffError.formErrors.phone = KratosErrorMessages[KratosErrorTraits.PHONE]
    }
    return sendBffError(bffError)
  }

  const data = await response.json()
  // The SuccessfulSelfServiceRegistrationWithoutBrowser is correct here. The OpenAPI spec for kratos
  // has some weird naming for types....
  const successData = data as SuccessfulSelfServiceRegistrationWithoutBrowser

  logger.info({
    identityId: successData.identity.id,
    signupId: id,
    flow: 'signup'
  }, '[SIGNUP] Kratos registration successful')

  // Mark signup complete
  // TODO: also handle via kratos webhook, add retry here and error handling
  logger.info({ id, userId: successData.identity.id, flow: 'signup' }, '[SIGNUP] Completing signup in backend')
  await grpc.completeSignup(request, {
    id,
    userId: successData.identity.id
  })

  return redirectWithSnackbar(
    request,
    href('/wallet-address'),
    {
      message: 'Your account was created successfully.',
      icon: 'close'
    },
    {
      headers: trimHeaders(response.headers, ['set-cookie'])
    }
  )
}
