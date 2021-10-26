import type { NextPage } from 'next'
import { Routes, Dashboard } from 'components'
import { GetServerSideProps, InferGetServerSidePropsType } from 'next'
import { getSession } from 'lib/kratos'

const SettingsPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ orgId, session }) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationSettings}>
      Hello there
    </Dashboard>
  )
}

export default SettingsPage

export const getServerSideProps: GetServerSideProps = getSession
