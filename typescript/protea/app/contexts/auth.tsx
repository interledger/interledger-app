import { createContext } from 'react'
import { Session } from '@ory/kratos-client'

interface IAuthContext {
  session: Session | undefined
}

export const AuthContext = createContext<IAuthContext>({
  session: undefined
})
