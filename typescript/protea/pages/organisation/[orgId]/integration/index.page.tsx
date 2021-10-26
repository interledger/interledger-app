import type { NextPage } from 'next'
import { Routes, Dashboard } from 'components'
import { GetServerSideProps, InferGetServerSidePropsType } from 'next'
import { getSession } from 'lib/kratos'

const IntegrationPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ orgId, session }) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationIntegration}>
      Hello there
    </Dashboard>
  )
}

export default IntegrationPage

export const getServerSideProps: GetServerSideProps = getSession
