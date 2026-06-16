import { useEffect } from 'react'
import { data, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import { About } from '~/routes/signup/About'
import { Landing } from '~/routes/signup/Landing'
import { Password } from '~/routes/signup/Password'
import styles from '~/styles/flags.css?url'
import type { Route } from './+types/route'
export { loader } from '~/routes/signup/route.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Sign up', back: 'signup' }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Sign up'
  }
])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const { countries } = useLoaderData()
  const [step, setCountries, reset] = useSignupStore((state) => [
    state.step,
    state.setCountries,
    state.reset
  ])

  useEffect(() => {
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
      {step === SignupStep.PASSWORD && <Password />}
    </>
  )
}

export async function action(args: Route.ActionArgs) {
  const { detailsAction, otpAction, passwordAction } = await import(
    '~/routes/signup/route.server'
  )
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

  throw data(
    { title: "Submitted a form that doesn't exist" },
    {
      status: 400
    }
  )
}
