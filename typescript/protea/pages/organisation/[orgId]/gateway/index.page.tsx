import type {
  GetServerSidePropsResult,
  NextPage,
  GetServerSideProps,
  InferGetServerSidePropsType
} from 'next'
import { Routes, Dashboard } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'
import { Session } from '@ory/kratos-client'

const GatewayPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({}) => {
  return (
    <Dashboard orgId={orgId} route={Routes.organisationGateway}>
      Hello there
    </Dashboard>
  )
}

export default GatewayPage

type GatewayPageProps = {
  session: Session
}

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<GatewayPageProps>> => {
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
