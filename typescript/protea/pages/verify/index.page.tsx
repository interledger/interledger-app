import {
  GetServerSideProps,
  GetServerSidePropsResult,
  InferGetServerSidePropsType,
  NextPage
} from 'next'
import { Logo, Router, Routes } from 'components'
import { VerifyForm } from './VerifyForm'
import React from 'react'
import { kratos } from 'lib/kratos'
import { AxiosError } from 'axios'

const VerifyPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ created_at, value }) => {
  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        Verify your email
      </h1>
      <p className='mb-10 text-medium'>
        We’ve sent you a verification email to {value}.
      </p>
      <VerifyForm email={value} createdAt={created_at} />
    </main>
  )
}

export default VerifyPage

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<any>> => {
  // Check if the user has a session already.
  const cookie = context.req?.headers.cookie
  try {
    const session = await kratos
      .toSession(undefined, cookie)
      .then((res) => res.data)
    // TODO: Check the status of the session and pass as prop to render form or not. enum: ["pending","sent","completed"]

    if (!session.identity.verifiable_addresses)
      return {
        redirect: {
          destination: Routes.signup,
          permanent: false
        }
      }
    // TODO look into allowing multiple addresses if necessary.
    if (session.identity.verifiable_addresses[0].verified)
      return {
        redirect: {
          destination: Routes.profile,
          permanent: false
        }
      }

    console.log(session.identity.verifiable_addresses[0])
    return {
      props: {
        ...session.identity.verifiable_addresses[0]
      }
    }
  } catch (error) {
    console.log('error')
    switch ((error as AxiosError)?.response?.status) {
      case 403:
      // This is a legacy error code thrown. See code 422 for
      // more details.
      case 422:
        // This status code is returned when we are trying to
        // validate a session which has not yet completed
        // it's second factor
        // redirect(Routes.login + '?aal=aal2', appContext)
        return {
          redirect: {
            destination: Routes.login + '?aal=aal2',
            permanent: false
          }
        }
    }
    return {
      redirect: {
        destination: Routes.signup,
        permanent: false
      }
    }
  }
}
