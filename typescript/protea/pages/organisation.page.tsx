import { NextPage } from 'next'
import { createLogoutHandler, useSession } from 'hooks'

const OrgDashboard: NextPage = () => {
  const session = useSession()
  const handleLogout = createLogoutHandler()

  return (
    <div>
      <div>Dashboard</div>

      <div>
        <code>Logged in with email: {session?.identity.traits.email}</code>
      </div>

      <button onClick={handleLogout} className='mt-4 border-4'>
        Logout
      </button>
    </div>
  )
}

export default OrgDashboard
