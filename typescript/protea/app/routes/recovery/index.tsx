import {
  GetServerSideProps,
  GetServerSidePropsResult,
  InferGetServerSidePropsType,
  NextPage
} from 'next'
import { RecoveryForm } from './RecoveryForm'
import { Logo, Router, Routes } from 'components'
import React from 'react'
import { getSessionOrRedirect } from 'lib/kratos'

const RecoveryPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        Recover your account
      </h1>
      <p className='mb-10 text-medium'>
        We’ll send you an email to change your password.
      </p>
      <RecoveryForm />
    </main>
  )
}

export default RecoveryPage

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
