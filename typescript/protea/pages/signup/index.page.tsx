import { GetServerSideProps, InferGetServerSidePropsType, NextPage } from 'next'
import { Logo, Router, Routes } from 'components'
import { SignupForm } from './SignupForm'
import React from 'react'
import { checkSession } from 'lib/kratos'

const SignupPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <main className='flex flex-col items-start justify-center max-w-sm mx-auto h-screen px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='text-4xl font-display font-medium text-strong mt-6 mb-1 leading-normal'>
        Create a new account
      </h1>
      <p className='text-medium mb-10'>
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

export const getServerSideProps: GetServerSideProps = checkSession
