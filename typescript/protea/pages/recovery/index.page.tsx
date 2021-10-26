import { GetServerSideProps, InferGetServerSidePropsType, NextPage } from 'next'
import { RecoveryForm } from './RecoveryForm'
import { Logo, Router, Routes } from 'components'
import React from 'react'
import { checkSession } from 'lib/kratos'

const RecoveryPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <main className='flex flex-col items-start justify-center max-w-sm mx-auto h-screen px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='text-4xl font-display font-medium text-strong mt-6 mb-1 leading-normal'>
        Recover your account
      </h1>
      <p className='text-medium mb-10'>
        We’ll send you an email to change your password.
      </p>
      <RecoveryForm />
    </main>
  )
}

export default RecoveryPage

export const getServerSideProps: GetServerSideProps = checkSession
