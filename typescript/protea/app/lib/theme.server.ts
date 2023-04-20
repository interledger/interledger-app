import { getSession, commitSession } from '~/session.server'

export type ThemeType = 'light' | 'dark' | 'system'

/**
 * Allows saving the theme to session storage.
 * @param request
 * @param theme The theme to set.
 * @returns Promise<string> The Set-Cookie header to be used in the HTTP response.
 */
export async function setTheme(
  request: Request,
  theme: ThemeType
): Promise<string> {
  const session = await getSession(request.headers.get('Cookie'))
  session.set('theme', theme)
  return commitSession(session)
}

/**
 * Fetches the current theme.
 * @param request Request
 * @returns Promise<ThemeType>
 */
export async function getTheme(request: Request): Promise<ThemeType> {
  const session = await getSession(request.headers.get('Cookie'))
  return session.has('theme') ? session.get('theme') : 'system'
}
