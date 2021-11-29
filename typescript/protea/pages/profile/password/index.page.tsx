import {
  GetServerSideProps,
  GetServerSidePropsResult,
  InferGetServerSidePropsType,
  NextPage
} from 'next'
import { Logo, Router, Routes } from 'components'
import { PasswordForm } from './PasswordForm'
import React from 'react'
import { getSessionOrRedirect } from 'lib/kratos'

const PasswordPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <main className='flex flex-col items-start justify-center max-w-sm mx-auto h-screen px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='text-4xl font-display font-medium text-strong mt-6 mb-1 leading-normal'>
        Set new password
      </h1>
      <p className='text-medium mb-10'>
        You’ve successfully recovered your account. Set a new password to
        continue.
      </p>
      <PasswordForm />
    </main>
  )
}

export default PasswordPage

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<any>> => {
  const session = await getSessionOrRedirect(context, true)
  if (session && 'redirect' in session) {
    return session
  }

  return {
    props: {}
  }
}
