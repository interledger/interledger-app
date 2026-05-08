import { getFeatures, getKycStatus } from '~/data/wallet.server'
import type {
  Features,
  Payment,
  PublicWalletInfo,
  SearchResult
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { KycStatus } from '~/lib/types'
import type { PlainMessage } from '@bufbuild/protobuf'
import { data, href, redirect } from 'react-router';
import type { LoaderFunctionArgs } from 'react-router';
import { grpc } from '~/lib/grpc.server';
import { isConnectError } from '~/lib/error.server';
import { envValue } from '~/env.server';

export async function searchLoader(
  { request }: LoaderFunctionArgs,
  term: string
) {
  const response = await grpc.searchWallets(request, { term })

  if (isConnectError(response)) throw response.errorResponse

  const results: PlainMessage<SearchResult>[] = response.results
  return data({ results })
}

export async function payLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)

  let results: PlainMessage<SearchResult>[] = []
  let address: PlainMessage<SearchResult> | null = null
  let publicWalletInfo: PublicWalletInfo | null = null
  let phoneMask: string = ''
  let features: Features | null = null
  let payment: PlainMessage<Payment> | null = null

  // used only on route load, params change and form submission
  // TODO should figure out if we need these based on the status of the payment
  if (url.search == '') {
    const { kycStatus } = await getKycStatus(request)
    if (kycStatus != KycStatus.Approved)
      return redirect(href('/personal-details'))

    features = await getFeatures(request)
  }

  return jsonWithCSRF(request, {
    features,
    results,
    address,
    phoneMask,
    publicWalletInfo,
    fynbosEnv: envValue("FYNBOS_ENV"),
    payment
  })
}