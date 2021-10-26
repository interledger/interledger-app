import type { NextPage } from 'next'
import { Routes, Dashboard } from 'components'
import { GetServerSideProps, InferGetServerSidePropsType } from 'next'
import { getSession } from 'lib/kratos'
import { RecoveryPasswordForm } from './RecoveryPasswordForm'

const OverviewPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ session }) => {
  return (
    <Dashboard route={Routes.profile}>
      <span className='text-4xl font-display font-medium'>Overview</span>
      <br />
      <span className='text-2xl font-normal font-display'>Personal info</span>
      <br />
      <span className='uppercase text-xs font-normal font-medium font-sans'>
        password
      </span>
      <span className='text-base font-normal font-sans'>
        &#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;
      </span>
      <br />
      <span className='uppercase text-xs font-normal font-medium font-sans'>
        email
      </span>
      <span className='text-base font-normal font-sans'>
        {session?.identity.traits.email}
      </span>
      <RecoveryPasswordForm />
    </Dashboard>
  )
}

export default OverviewPage

export const getServerSideProps: GetServerSideProps = getSession
