import type {
  GetServerSidePropsResult,
  NextPage,
  GetServerSideProps,
  InferGetServerSidePropsType
} from 'next'
import { Routes, Dashboard } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'
import { Session } from '@ory/kratos-client'

const SettingsPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({}) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationSettings}>
      Hello there
    </Dashboard>
  )
}

export default SettingsPage

type SettingsPageProps = {
  session: Session
}

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<SettingsPageProps>> => {
  const session = await getSessionOrRedirect(context, true)
  if ('redirect' in session) {
    return session
  }

  return {
    props: {
      session,
    }
  }
}
