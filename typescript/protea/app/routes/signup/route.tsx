import type { MetaFunction } from 'react-router';
import { useLoaderData } from 'react-router';
import { useEffect } from 'react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import { About } from '~/routes/signup/About'
import { Landing } from '~/routes/signup/Landing'
import { Password } from '~/routes/signup/Password'
import { Phone } from '~/routes/signup/Phone'
import styles from '~/styles/flags.css?url'
import type { loader } from './route.server'

export { loader, action } from './route.server'

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
