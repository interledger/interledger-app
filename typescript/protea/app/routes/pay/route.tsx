import { Code } from '@bufbuild/connect'
import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type { ActionFunctionArgs, LoaderFunctionArgs, MetaFunction } from 'react-router';
import { data, redirect } from 'react-router';
import { useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  Card,
  CardContent,
  CardIcon,
  Icon,
  Layouts,
  Router
} from '~/components'
import { CommandActions } from '~/components/Scaffold/CommandActions'
import { getFeatures, getKycStatus } from '~/data/wallet.server'
import type {
  Features,
  Payment,
  PublicWalletInfo,
  SearchResult
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { mergeMeta } from '~/lib/meta'
import { KycStatus } from '~/lib/types'
import { PaymentIdentityType } from '~/lib/types'

export async function loader(args: LoaderFunctionArgs) {
  const url = new URL(args.request.url)

  const term = url.searchParams.get('term')
  if (term) {
    return searchLoader(args, term)
  }

  return payLoader(args)
}

async function searchLoader(
  { request }: LoaderFunctionArgs,
  term: string
) {
  const response = await grpc.searchWallets(request, { term })

  if (isConnectError(response)) throw response.errorResponse

  const results: PlainMessage<SearchResult>[] = response.results
  return data({ results })
}

async function payLoader({ request }: LoaderFunctionArgs) {
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
    fynbosEnv: process.env.FYNBOS_ENV,
    payment
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Pay search', back: 'pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Pay search'
  }
])

export default function Page() {
  const { features } = useLoaderData<typeof payLoader>()

  if (features && !features.sendEnabled)
    return (
      <>
        <Alert>
          <Icon>error</Icon>
          <AlertBody>
            Making payments in your state is currently unavailable. We're
            working to unlock all regions and will notify you when accessible.
            Thank you for your patience.
          </AlertBody>
        </Alert>
        <Card>
          <CardContent>
            <div className='flex items-start space-x-4'>
              <CardIcon>
                <Icon>credit_card</Icon>
              </CardIcon>
              <div className='flex flex-col space-y-4'>
                <p className='text-sm text-medium'>
                  Connect a card to receive payments.
                </p>
                <Router
                  prefetch='render'
                  className='text-sm font-medium text-primary'
                  to={href('/accounts')}
                >
                  Go to accounts page
                </Router>
              </div>
            </div>
          </CardContent>
        </Card>
      </>
    )

  return <CommandActions />
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const type = form.get('type') as string
  const walletUrl = form.get('walletUrl') as string

  const errors = {
    form: '',
    amount: '',
    address: '',
    linkedAccount: '',
    note: ''
  }

  const clientIpAddress = getClientIP(request)

  let payment = await grpc.createPayment(request, {
    receiverIdentity: walletUrl,
    receiverIdentityType: PaymentIdentityType.WalletURL,
    ipAddress: clientIpAddress
  })
  if (isConnectError(payment)) {
    if (payment.code == Code.InvalidArgument) {
      return payment.error({ errors, payment: null, type })
    }
    return payment.error(
      { errors, payment: null, type },
      {},
      { action: 'Contact support' }
    )
  }

  return redirect(href('/pay/:paymentId', { paymentId: payment.id }))
}
