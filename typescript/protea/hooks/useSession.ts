import { useContext } from 'react'
import { Session } from '@ory/kratos-client'
import { AuthContext } from 'contexts/auth'

export const useSession = (): Session | undefined => {
  const { session } = useContext(AuthContext)

  // TODO: make sure session has not expired

  return session
}
