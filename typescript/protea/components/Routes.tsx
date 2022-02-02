import { ParsedUrlQueryInput } from 'querystring'
import { FC } from 'react'
import Link from 'next/link'
import { Redirect } from 'next'

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
  waitlist = '/waitlist',
  verify = '/verify',
  verifyDetails = '/verify/details',
  logout = '/logout',
  recovery = '/recovery',

  settings = '/settings',
  settingsPassword = '/settings/password',

  //
  // Wallet
  //

  walletHome = '/home',
  deposit = '/deposit',
  withdraw = '/withdraw',
  activity = '/activity',
  activityFilter = '/activity/filter',
  activityTransaction = '/activity/transaction/[id]',
  transact = '/transact',
  transactReceive = '/transact/receive',
  transactPreview = '/transact/preview',
  connect = '/connect',

  //
  // Organisation
  //
  organisation = '/organisation',
  organisationOverview = '/organisation/[orgId]',
  organisationIntegration = '/organisation/[orgId]/integration',
  organisationSettings = '/organisation/[orgId]/settings',
  organisationGateway = '/organisation/[orgId]/gateway',

  organisationWallet = '/organisation/[orgId]/wallet',
  organisationWalletAccounts = '/organisation/[orgId]/wallet/accounts',
  organisationWalletTransactions = '/organisation/[orgId]/wallet/transactions',
  organisationWalletRisk = '/organisation/[orgId]/wallet/risk',
  organisationWalletOperations = '/organisation/[orgId]/wallet/operations',

  //
  // External
  //
  interledger = 'https://interledger.org',
  openPayments = 'https://openpayments.dev',
  email = 'mailto:hello@fynbos.dev',
  twitter = 'https://twitter.com/'
}

export type RouterProps = {
  className?: string
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
export const Router: FC<RouterProps> = ({
  className,
  children,
  href,
  ...rest
}) => {
  return (
    <Link href={href}>
      <a
        className={`focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-1 ${className}`}
        {...rest}
      >
        {children}
      </a>
    </Link>
  )
}

/**
 * Redirects to the given route. To be used in getServerSideProps.
 * @param destination The destination of the redirect.
 * @returns A redirect to the destination.
 */
export const redirect = (destination: string): { redirect: Redirect } => {
  return {
    redirect: {
      destination,
      permanent: false
    }
  }
}
