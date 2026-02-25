import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import logger from '~/lib/logger.server'
import { Code } from '@bufbuild/connect'
import type { SuccessfulNativeRegistration } from '@ory/client'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { getCookie, withCookie, buildHeadersWithCookies } from '~/lib/kratos/cookie.util'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.util'
import { handleFlowError, mapFlowToFieldErrors } from '~/lib/kratos/error.server'
import { requireNoUserSession } from '~/lib/kratos/session.util.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import { About, isEUCountry } from '~/routes/signup/About'
import { Landing } from '~/routes/signup/Landing'
import { Password } from '~/routes/signup/Password'
import { Phone } from '~/routes/signup/Phone'
import styles from '~/styles/flags.css'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)

  await requireNoUserSession(request)

  let response = await grpc.getCountries(request, {})

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
    throw err
  }

  return jsonWithCSRF(
    request,
    {
      countries,
      kratosFlowId: flowResponse.data.id,
      kratosCsrfToken: getCsrfTokenFromFlow(flowResponse.data),
      fynbosEnv: process.env.FYNBOS_ENV
    },
    {
      headers: buildHeadersWithCookies(flowResponse)
    }
  )
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Sign up', back: 'signup' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Sign up'
  }
])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const { countries } = useLoaderData<typeof loader>()
  const [step, setCountries, reset] = useSignupStore((state) => [
    state.step,
    state.setCountries,
    state.reset
  ])

  useEffect(() => {
    // This ensures that the state is only cleared when this route is unmounted.
    return () => {
      reset()
    }
  }, [reset])

  useEffect(() => {
    setCountries(countries)
  }, [countries, setCountries])

  return (
    <>
      {step === SignupStep.LANDING && <Landing />}
      {step === SignupStep.ABOUT && <About />}
      {step === SignupStep.PHONE && <Phone />}
      {step === SignupStep.PASSWORD && <Password />}
    </>
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

  throw json(
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

  return json({
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
    }, 'Failed to set mobile number')

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

  logger.info({ id, flow: 'signup' }, 'Mobile number set successfully for signup')
  return json({ id, phone, errors: data.errors })
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
  logger.info({
    flowId: kratosFlowId,
    countryCode: kratosRequestPayload.traits.countryCode,
    hasTraits: !!kratosRequestPayload.traits,
    hasPassword: !!password,
    hasCsrfToken: !!kratosCsrfToken,
    flow: 'signup'
  }, 'Sending registration request to Kratos')

  let response
  try {
    response = await kratosPublic.updateRegistrationFlow(
      {
        flow: kratosFlowId,
        updateRegistrationFlowBody: kratosRequestPayload
      },
      withCookie(cookie)
    )
    logger.info({
      status: response.status,
      statusText: response.statusText,
      headers: response.headers?.['set-cookie'],
      flow: 'signup'
    }, 'Kratos registration response')
  } catch (err: any) {
    logger.error({
      status: err.response?.status,
      statusText: err.response?.statusText,
      flow: 'signup'
    }, 'Kratos registration error')
    const flowData = err.response?.data
    const errs = mapFlowToFieldErrors(flowData, errors)
    if ((errs as any)[KratosErrorTraits.PHONE]) {
      errors.phone = KratosErrorMessages[KratosErrorTraits.PHONE]
      return error(request, { errors })
    }
    return error(request, { errors: errs })
  }

  const successData = response.data as SuccessfulNativeRegistration
  logger.info({
    identityId: successData.identity.id,
    signupId: id,
    flow: 'signup'
  }, 'Kratos registration successful')

  // Mark signup complete
  // TODO: also handle via kratos webhook, add retry here and error handling
  logger.info({ id, userId: successData.identity.id, flow: 'signup' }, 'Completing signup in backend')
  await grpc.completeSignup(request, {
    id,
    userId: successData.identity.id
  })

  return redirectWithSnackbar(
    request,
    route('/wallet-address'),
    {
      message: 'Your account was created successfully.',
      icon: 'close'
    },
    {
      headers: buildHeadersWithCookies(response)
    }
  )
}