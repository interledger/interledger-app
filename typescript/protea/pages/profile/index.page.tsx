import type {
  GetServerSidePropsResult,
  NextPage,
  GetServerSideProps,
  InferGetServerSidePropsType
} from 'next'
import { Routes, Dashboard, redirect, Router, NextIcon } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'
import { Session } from '@ory/kratos-client'
import { getOrgsForDashboardOrRedirect, OrgsForDashboard } from 'lib/dashboard'

const ProfilePage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ session, orgsForDashboard, preview }) => {
  return (
    <Dashboard
      preview={preview}
      route={Routes.profile}
      orgsForDashboard={orgsForDashboard}
    >
      <div className='pt-4 mb-10'>
        <span className='text-4xl font-display font-medium'>Profile</span>
      </div>
      <div className='flex flex-col text-medium border-2 border-base'>
        <span className='text-2xl font-normal font-display p-6'>
          Personal info
        </span>
        <div className='grid grid-cols-12 gap-4 items-center py-5 px-6'>
          <div className='col-span-3'>
            <span className='uppercase text-xs text-medium font-medium font-sans'>
              email
            </span>
          </div>
          <div className='col-span-8'>
            <span className='text-base font-normal font-sans'>
              {session?.identity.traits.email}
            </span>
          </div>
        </div>
        <div className='border-b border-base' />
        <Router href={Routes.profilePassword}>
          <div className='grid grid-cols-12 gap-4 items-center py-5 px-6'>
            <div className='col-span-3'>
              <span className='uppercase text-xs text-medium font-medium font-sans'>
                password
              </span>
            </div>
            <div className='col-span-8'>
              <span className='text-base font-normal font-sans'>
                &#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;
              </span>
            </div>
            <div className='justify-self-end col-span-1'>
              <span>
                <NextIcon />
              </span>
            </div>
          </div>
        </Router>
        <div className='flex items-center justify-end bg-base py-5 px-6'>
          <Router href={Routes.logout}>
            <span className='text-primary'>Logout</span>
          </Router>
        </div>
      </div>
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
  const session = await getSessionOrRedirect(context, true)
  if ('redirect' in session) {
    return session
  }

  const { flow: flowId } = context.query
  if (flowId) return redirect(`${Routes.profilePassword}?flow=${flowId}`)

  const orgsForDashboard = await getOrgsForDashboardOrRedirect(context)
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
