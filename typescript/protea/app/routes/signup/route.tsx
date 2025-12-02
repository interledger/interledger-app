import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'

import { Code } from '@bufbuild/connect'
import type { SuccessfulSelfServiceRegistrationWithoutBrowser } from '@ory/kratos-client'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { verifyRecaptcha } from '~/lib/recaptcha.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import { About, isEUCountry } from '~/routes/signup/About'
import { Landing } from '~/routes/signup/Landing'
import { Password } from '~/routes/signup/Password'
import { Phone } from '~/routes/signup/Phone'
import styles from '~/styles/flags.css'
import { ErrorDescriptions } from '~/lib/error.constants'

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

  const response = await grpc.setSignupMobileNumber(request, {
    id,
    mobile: phone,
    otp: otp
  })

  if (isConnectError(response)) {
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

  const recaptchaToken = form.get('recaptcha_token')
  if (!recaptchaToken) {
    errors.form = ErrorDescriptions.RECAPTCHA_VERIFICATION_FAILED
    return error(request, { errors })
  }

  const isValid = await verifyRecaptcha(recaptchaToken.toString())
  if (!isValid) {
    errors.form = ErrorDescriptions.RECAPTCHA_VERIFICATION_FAILED
    return error(request, { errors })
  }

  const response = await fetch(
    `${KRATOS_URL}/self-service/registration?flow=${kratosFlowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
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
      }),
      headers: {
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )
  if (response.status >= 400) {
    const errs = await kratosErrorMapping(response, errors)
    if ((errs as any)[KratosErrorTraits.PHONE]) {
      errors.phone = KratosErrorMessages[KratosErrorTraits.PHONE]
      return error(request, { errors })
    }
    return error(request, { errors: errs })
  }

  const data = await response.json()
  // The SuccessfulSelfServiceRegistrationWithoutBrowser is correct here. The OpenAPI spec for kratos
  // has some weird naming for types....
  const successData = data as SuccessfulSelfServiceRegistrationWithoutBrowser

  // Mark signup complete
  // TODO: also handle via kratos webhook, add retry here and error handling
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
      headers: trimHeaders(response.headers, ['set-cookie'])
    }
  )
}
