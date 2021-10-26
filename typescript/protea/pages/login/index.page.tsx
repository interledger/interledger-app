import { NextPage, InferGetServerSidePropsType, GetServerSideProps } from 'next'
import { Logo, Router, Routes } from 'components'
import { LoginForm } from './LoginForm'
import React from 'react'
import { checkSession } from 'lib/kratos'

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

export const getServerSideProps: GetServerSideProps = checkSession
