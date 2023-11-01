import { Code } from '@bufbuild/connect'
import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
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
import type { FormattedLinkedAccount } from '~/data/wallet.server'
import {
  getFeatures,
  getKycStatus,
  getLinkedAccounts
} from '~/data/wallet.server'
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
import { PaymentIdentityType } from '~/lib/types/payment'
import { KycStatus } from '~/routes/_index/route'
import { Search } from '~/routes/pay/Search'

export async function loader(args: LoaderFunctionArgs) {
  const url = new URL(args.request.url)

  const term = url.searchParams.get('term')
  if (term) {
    return searchLoader(args, term)
  }

  return payLoader(args)
}

export async function searchLoader(
  { request }: LoaderFunctionArgs,
  term: string
) {
  const response = await grpc.searchWallets(request, { term })

  if (isConnectError(response)) throw response.errorResponse

  const results: PlainMessage<SearchResult>[] = response.results
  return json({ results })
}

export async function payLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)

  let results: PlainMessage<SearchResult>[] = []
  let address: PlainMessage<SearchResult> | null = null
  let sendAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PublicWalletInfo | null = null
  let phoneMask: string = ''
  let features: Features | null = null
  let payment: PlainMessage<Payment> | null = null

  // used only on route load, params change and form submission
  // TODO should figure out if we need these based on the status of the payment
  if (url.search == '') {
    const { kycStatus } = await getKycStatus(request)
    if (kycStatus != KycStatus.Approved)
      return redirect(route('/personal-details'))

    features = await getFeatures(request)

    const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)
    sendAccounts = [...cardAccounts, ...bankAccounts].filter(
      (acc) => acc.canSend
    )
  }

  return jsonWithCSRF(request, {
    features,
    results,
    address,
    sendAccounts,
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
  const { features, sendAccounts, fynbosEnv } =
    useLoaderData<typeof payLoader>()

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
                  to={route('/accounts')}
                >
                  Go to accounts page
                </Router>
              </div>
            </div>
          </CardContent>
        </Card>
      </>
    )

  if (sendAccounts.length === 0)
    return (
      <Card>
        <CardContent>
          <div className='flex items-start space-x-4'>
            <CardIcon>
              <Icon>credit_card</Icon>
            </CardIcon>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                To send a payment, first connect a card.
              </p>
              <Router
                prefetch='render'
                className='text-sm font-medium text-primary'
                to={route('/accounts')}
              >
                Go to accounts page
              </Router>
            </div>
          </div>
        </CardContent>
      </Card>
    )

  return <Search fynbosEnv={fynbosEnv} />
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const type = form.get('type') as string
  const receiverIdentifier = String(form.get('receiverIdentifier') || '')
  const receiverIdentifierType = Number(
    form.get('receiverIdentifierType') || PaymentIdentityType.Unknown
  )

  const errors = {
    form: '',
    amount: '',
    address: '',
    linkedAccount: '',
    note: ''
  }

  const clientIpAddress = getClientIP(request)

  let payment = await grpc.createPayment(request, {
    receiverIdentity: receiverIdentifier,
    receiverIdentityType: receiverIdentifierType,
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

  return redirect(route('/pay/:paymentId', { paymentId: payment.id }))
}
