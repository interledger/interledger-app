import type { PlainMessage } from '@bufbuild/protobuf'
import type { LoaderFunctionArgs } from 'react-router'
import { data } from 'react-router'
import { getFeatures } from '~/data/wallet.server'
import { envValue } from '~/env.server'
import type {
  Features,
  Payment,
  PublicWalletInfo,
  SearchResult
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

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

  const results: PlainMessage<SearchResult>[] = []
  const address: PlainMessage<SearchResult> | null = null
  const publicWalletInfo: PublicWalletInfo | null = null
  const phoneMask: string = ''
  let features: Features | null = null
  const payment: PlainMessage<Payment> | null = null

  // used only on route load, params change and form submission
  // TODO should figure out if we need these based on the status of the payment
  if (url.search == '') {
    features = await getFeatures(request)
  }

  return jsonWithCSRF(request, {
    features,
    results,
    address,
    phoneMask,
    publicWalletInfo,
    fynbosEnv: envValue('FYNBOS_ENV'),
    payment
  })
}
