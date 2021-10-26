import type { NextPage } from 'next'
import { Routes, Dashboard } from 'components'
import { GetServerSideProps, InferGetServerSidePropsType } from 'next'
import { getSession } from 'lib/kratos'

const OverviewPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ orgId, session }) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationOverview}>
      <span className='text-4xl font-medium'>
        Overview {session?.identity.traits.email}
      </span>
    </Dashboard>
  )
}

export default OverviewPage

export const getServerSideProps: GetServerSideProps = getSession
