import type {
  GetServerSidePropsResult,
  NextPage,
  GetServerSideProps,
  InferGetServerSidePropsType
} from 'next'
import { getSession } from 'lib/kratos'
import { Session } from '@ory/kratos-client'
import { Dashboard, redirect, Routes } from 'components'
import { getOrgsForDashboard, OrgsForDashboard } from 'lib/dashboard'

const ProfilePage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ session, orgsForDashboard, preview }) => {
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

export default ProfilePage

type ProfilePageProps = {
  session: Session
  orgsForDashboard: OrgsForDashboard
  preview: boolean
}

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<ProfilePageProps>> => {
  const session = await getSession(context, true)
  if ('redirect' in session) {
    return session
  }

  const { flow: flowId } = context.query
  if (flowId) return redirect(`${Routes.profilePassword}?flow=${flowId}`)

  const orgsForDashboard = await getOrgsForDashboard(context)
  if ('redirect' in orgsForDashboard) {
    return orgsForDashboard
  }

  // Profile page can't be in preview as it's not under an org.
  const preview = false

  return {
    props: {
      session,
      orgsForDashboard,
      preview
    }
  }
}
