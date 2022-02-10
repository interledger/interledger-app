import {
  GetServerSideProps,
  GetServerSidePropsResult,
  InferGetServerSidePropsType,
  NextPage
} from 'next'
import { Logo, Router, Routes } from 'components'
import { SignupForm } from './SignupForm'
import React from 'react'
import { getSessionOrRedirect } from 'lib/kratos'

const SignupPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        Create a new account
      </h1>
      <p className='mb-10 text-medium'>
        Or{' '}
        <Router href={Routes.login}>
          <span className='text-primary'>sign in to an existing account.</span>
        </Router>
      </p>
      <SignupForm />
    </main>
  )
}

export default SignupPage

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<any>> => {
  const session = await getSessionOrRedirect(context, false)
  if (session && 'redirect' in session) {
    return session
  }

  return {
    props: {}
  }
}
