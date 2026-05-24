import { href, matchPath, redirect } from 'react-router'
import type { RouteParams } from 'routes-gen'
import { getKycStatus } from '~/data/wallet.server'
import { redirectWithSnackbar } from './snackbar.server'
import { KycStatus } from './types'

const KYC_APPROVED_PATHS = [
  '/pay',
  '/pay/:paymentId',
  '/deposit',
  '/deposit/:paymentId',
  '/withdraw',
  '/withdraw/:paymentId'
] as const satisfies readonly (keyof RouteParams)[]

function requiresApprovedKyc(pathname: string): boolean {
  return KYC_APPROVED_PATHS.some(
    (pattern) => matchPath(pattern, pathname) != null
  )
}

export async function kycApprovedGuard(
  pathname: string,
  request: Request
): Promise<void> {
  if (!requiresApprovedKyc(pathname)) return

  const { kycStatus } = await getKycStatus(request)
  if (kycStatus === KycStatus.Approved) return

  if (kycStatus === KycStatus.Pending || kycStatus === KycStatus.InReview) {
    throw await redirectWithSnackbar(request, href('/'), {
      message:
        'Your identity is being verified.\nThis functionality will be available once approved.'
    })
  }

  throw redirect(href('/personal-details'))
}
