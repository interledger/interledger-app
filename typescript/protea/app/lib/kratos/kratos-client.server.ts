/**
 * Ory Kratos SDK Client
 *
 * This module provides a centralized Kratos SDK client for server-side use.
 * It wraps the official @ory/client SDK for use with Remix loaders/actions.
 */
import {
  Configuration,
  FrontendApi,
} from '@ory/client'

export const KRATOS_SESSION_COOKIE = 'ory_kratos_session'

export const CLEAR_SESSION_COOKIE_HEADER = `${KRATOS_SESSION_COOKIE}=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax`

export const KRATOS_URL = process.env.KRATOS_URL
const KRATOS_PUBLIC_URL = KRATOS_URL
if (!KRATOS_PUBLIC_URL) {
  throw new Error('KRATOS_URL environment variable is not set')
}

const publicConfig = new Configuration({
  basePath: KRATOS_PUBLIC_URL,
  baseOptions: {
    withCredentials: true
  }
})
export const kratosPublic = new FrontendApi(publicConfig)
if (!kratosPublic) {
  throw new Error('Failed to initialize Kratos client.')
}
