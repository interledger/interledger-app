import {
  NextPage,
  InferGetServerSidePropsType,
  GetServerSideProps,
  GetServerSidePropsResult
} from 'next'
import { Logo, Router, Routes } from 'components'
import { LoginForm } from './LoginForm'
import React from 'react'
import { getSessionOrRedirect } from 'lib/kratos'

const LoginPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <main className='flex flex-col items-start justify-center max-w-sm mx-auto h-screen px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='text-4xl font-display font-medium text-strong mt-6 mb-1 leading-normal'>
        Sign in to your account
      </h1>
      <p className='text-medium mb-10'>
        Or{' '}
        <Router href={Routes.signup}>
          <span className='text-primary'>create a new account.</span>
        </Router>
      </p>
      <LoginForm />
    </main>
  )
}

export default LoginPage

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
