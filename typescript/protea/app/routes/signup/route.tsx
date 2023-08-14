import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'

import type { SuccessfulSelfServiceRegistrationWithoutBrowser } from '@ory/kratos-client'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error } from '~/lib/error.server'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { canSignup, setWaitlistSignupComplete } from '~/lib/signupCheck.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import { About } from '~/routes/signup/About'
import { Landing } from '~/routes/signup/Landing'
import { Password } from '~/routes/signup/Password'
import { Phone } from '~/routes/signup/Phone'
import styles from '~/styles/flags.css'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  await requireNoUserSession(request)
  await canSignup(request)

  let response = await grpcClient
    .getCountries({})
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }
  const countries = response.response.countries

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

export const meta: MetaFunction = () => {
  return {
    title: 'Sign up'
  }
}

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

// The field names given by the backend for field violations
type fieldErrorsMap =
  | 'FirstName'
  | 'LastName'
  | 'CountryOfResidence'
  | 'Email'
  | 'MobileNumber'
  | 'OTP'

function mapper(
  field: fieldErrorsMap
): 'firstName' | 'lastName' | 'country' | 'email' | 'phone' | 'otp' | null {
  switch (field) {
    case 'FirstName':
      return 'firstName'
    case 'LastName':
      return 'lastName'
    case 'CountryOfResidence':
      return 'country'
    case 'Email':
      return 'email'
    case 'MobileNumber':
      return 'phone'
    case 'OTP':
      return 'otp'
    default:
      return null
  }
}
export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const formName = (await form.get('formName')) as string

  await validateCSRFToken(request, form)

  const fieldErrors = {
    form: '',
    firstName: '',
    lastName: '',
    country: '',
    email: '',
    otp: '',
    phone: '',
    serviceAgreement: '',
    password: ''
  }

  if (formName === 'details') {
    const firstName = form.get('firstName') as string
    const lastName = form.get('lastName') as string
    const country = form.get('country') as string
    const email = form.get('email') as string

    if (!(country == 'US')) {
      return redirect(
        `/waitlist?country=${country}&email=${email}&fullName=${firstName} ${lastName}`
      )
    }

    let response = await grpcClient
      .setSignupUserData({
        firstName,
        lastName,
        countryCode: country,
        email
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

    return json({ id: response.response.id, firstName, lastName, email })
  }

  if (formName === 'otp') {
    const id = form.get('id') as string
    const otp = form.get('otp') as string
    const phone = form.get('phone') as string

    let response = await grpcClient
      .setSignupMobileNumber({
        id,
        mobile: phone,
        otp: otp
      })
      .then((v) => v)
      .catch(StatusError)

    if (isGrpcError(response)) {
      if (response.code == Code.INVALID_ARGUMENT) {
        for (let violation of (response as GrpcError).details[0]
          .fieldViolations) {
          const field = mapper(violation.field as fieldErrorsMap)
          if (field != null) fieldErrors[field] = violation.description
        }
        return error(request, { errors: { ...fieldErrors } })
      } else if (response.code == Code.ALREADY_EXISTS) {
        fieldErrors['phone'] = 'Mobile phone number is already registered.'
        return error(request, { errors: { ...fieldErrors } })
      } else
        return error(
          request,
          { errors: { ...fieldErrors } },
          { action: 'Contact support' }
        )
    }
    return json({ id, phone })
  }

  if (formName === 'password') {
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

    if (serviceAgreement == null) {
      fieldErrors.serviceAgreement = 'You are required to agree to continue.'
      return error(request, { errors: { ...fieldErrors } })
    }

    const res = await fetch(
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
    if (res.status >= 400) {
      const errs = await kratosErrorMapping(res, fieldErrors)
      return error(request, errs)
    }

    const data = await res.json()
    // The SuccessfulSelfServiceRegistrationWithoutBrowser is correct here. The OpenAPI spec for kratos
    // has some weird naming for types....
    const successData = data as SuccessfulSelfServiceRegistrationWithoutBrowser

    // Mark signup complete
    const userId = successData.identity.id
    // TODO: also handle via kratos webhook, add retry here and error handling
    await grpcClient.completeSignup({
      id,
      userId: userId
    })
    await setWaitlistSignupComplete(request, userId)

    return redirectWithSnackbar(
      request,
      route('/wallet-address'),
      {
        message: 'Your account was created successfully.',
        icon: 'close'
      },
      {
        headers: trimHeaders(res.headers, ['set-cookie'])
      }
    )
  }

  throw json(
    { title: "Submitted a form that doesn't exist" },
    {
      status: 400
    }
  )
}
