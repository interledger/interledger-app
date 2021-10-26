import type { NextPage } from 'next'
import { Routes, Dashboard } from 'components'
import { GetServerSideProps, InferGetServerSidePropsType } from 'next'
import { getSession } from 'lib/kratos'

const GatewayPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ orgId, session }) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationGateway}>
      Hello there
    </Dashboard>
  )
}

export default GatewayPage

export const getServerSideProps: GetServerSideProps = getSession
