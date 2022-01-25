import type {
  GetServerSidePropsResult,
  NextPage,
  GetServerSideProps,
  InferGetServerSidePropsType
} from 'next'
import { Routes, Dashboard } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'
import { Session } from '@ory/kratos-client'
import { OrgsForDashboard, getOrgsForDashboardOrRedirect } from 'lib/dashboard'

const WalletPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ orgsForDashboard, preview }) => {
  return (
    <Dashboard
      preview={preview}
      orgsForDashboard={orgsForDashboard}
      route={Routes.organisationWallet}
    >
      <div className='mb-10 pt-4'>
        <span className='font-display text-4xl font-medium'>Wallet</span>
      </div>
    </Dashboard>
  )
}

export default WalletPage

type WalletPageProps = {
  session: Session
  orgsForDashboard: OrgsForDashboard
  preview: boolean
}

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<WalletPageProps>> => {
  const session = await getSessionOrRedirect(context, true)
  if ('redirect' in session) {
    return session
  }

  const orgsForDashboard = await getOrgsForDashboardOrRedirect(context)
  if ('redirect' in orgsForDashboard) {
    return orgsForDashboard
  }

  /**
   * Preview should be enforced unless the org is verified.
   * If the org is verified, use the last state set by user.
   */
  let preview = true
  if (orgsForDashboard.currentOrg?.verified) {
    preview = context.preview || false // Preview is either true or undefined, there is no false
  }

  return {
    props: {
      session,
      orgsForDashboard,
      preview
    }
  }
}
