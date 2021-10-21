import { ParsedUrlQueryInput } from 'querystring'
import { FC } from 'react'
import Link from 'next/link'

/**
 * Predefined routes that allow for consistent routing.
 *
 * Ensure all routes are prefixed with `/` to prevent relative routing.
 */
export enum Routes {
  //
  // Marketing pages
  //
  home = '/',
  blog = '/blog',
  blogPost = '/blog/[slug]',

  //
  // User
  //
  login = '/login',
  signup = '/signup',
  verify = '/verify',
  signout = '/signout',
  profile = '/profile',

  //
  // Organisation
  //
  organisation = '/organisation',
  organisationOverview = '/organisation/[id]',
  organisationIntegration = '/organisation/[id]/integration',
  organisationSettings = '/organisation/[id]/settings',
  organisationGateway = '/organisation/[id]/gateway',
  organisationWallet = '/organisation/[id]/wallet',

  //
  // External
  //
  interledger = 'https://interledger.org',
  openPayments = 'https://openpayments.dev',
  email = 'mailto:hello@fynbos.dev',
  twitter = 'https://twitter.com/[handle]'
}

export type RouterProps = {
  href:
    | Routes
    | {
        pathname: Routes
        query?: string | null | ParsedUrlQueryInput
      }
}

/**
 * Router replaces the next/link with a Link that only accepts predefined Routes.
 * Wraps and passes props to an anchor tag for a11y.
 *
 * @param children The children of the Link.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
export const Router: FC<RouterProps> = ({ children, href, ...rest }) => {
  return (
    <Link href={href}>
      <a {...rest}>{children}</a>
    </Link>
  )
}
