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
      <div className='mb-10 pt-4'>
        <span className='font-display text-4xl font-medium'>Profile</span>
      </div>
      <div className='flex flex-col border-2 border-base text-medium'>
        <span className='p-6 font-display text-2xl font-normal'>
          Personal info
        </span>
        <div className='grid grid-cols-12 items-center gap-4 py-5 px-6'>
          <div className='col-span-3'>
            <span className='font-sans text-xs font-medium uppercase text-medium'>
              email
            </span>
          </div>
          <div className='col-span-8'>
            <span className='font-sans text-base font-normal'>
              {session?.identity.traits.email}
            </span>
          </div>
        </div>
        <div className='border-b border-base' />
        <Router href={Routes.profilePassword}>
          <div className='grid grid-cols-12 items-center gap-4 py-5 px-6'>
            <div className='col-span-3'>
              <span className='font-sans text-xs font-medium uppercase text-medium'>
                password
              </span>
            </div>
            <div className='col-span-8'>
              <span className='font-sans text-base font-normal'>
                &#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;
              </span>
            </div>
            <div className='col-span-1 justify-self-end'>
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
