import {
  GetServerSideProps,
  GetServerSidePropsResult,
  InferGetServerSidePropsType,
  NextPage
} from 'next'
import React, { FC } from 'react'
import { Router, Routes, Logo } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'
import { Session } from '@ory/kratos-client'

const OverviewPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({}) => {
  return (
    <div className='relative overflow-hidden w-full'>
      <Container className='overflow-x-hidden'>
        <main className='flex flex-grow flex-col px-4 sm:p-8 justify-center items-start'>
          <p className='text-9xl font-medium font-display text-primary'>404</p>
          <div className='sm:mt-12'>
            <div>
              <h1 className='text-4xl font-medium font-display text-medium'>
                Page not found
              </h1>
              <p className='mt-2 text-weak font-sans'>
                Please check the URL in the address bar and try again.
              </p>
            </div>
            <div className='mt-10'>
              <Router href={Routes.home}>
                <span className='text-primary'>Go back home</span>
              </Router>
            </div>
          </div>
        </main>
        <LeavesDecor />
      </Container>
    </div>
  )
}

export default OverviewPage

type OverviewPageProps = {
  session: Session
}

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<OverviewPageProps>> => {
  const session = await getSessionOrRedirect(context, true)
  if ('redirect' in session) {
    return session
  }

  return {
    props: {
      session: session,
    }
  }
}
