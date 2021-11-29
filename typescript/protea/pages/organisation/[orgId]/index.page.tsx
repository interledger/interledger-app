import { Routes, Dashboard } from 'components'
import {
  GetServerSideProps,
  InferGetServerSidePropsType,
  GetServerSidePropsResult,
  NextPage
} from 'next'
import { getSessionOrRedirect } from 'lib/kratos'
import { Session } from '@ory/kratos-client'

const OverviewPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({}) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationOverview}>
      <span className='text-4xl font-medium'>
        Overview {session?.identity.traits.email}
      </span>
    </Dashboard>
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
      session,
    }
  }
}
