import type { NextPage } from 'next'
import { Routes, Dashboard } from 'components'
import { GetServerSideProps, InferGetServerSidePropsType } from 'next'
import { getSession } from 'lib/kratos'

const WalletPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ orgId, session }) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationWallet}>
      Hello there
    </Dashboard>
  )
}

export default WalletPage

export const getServerSideProps: GetServerSideProps = getSession
