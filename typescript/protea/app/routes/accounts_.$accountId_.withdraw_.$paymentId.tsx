import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData, useParams } from '@remix-run/react'
import { DateTime } from 'luxon'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Icon, Layouts } from '~/components'
import { Label } from '~/components/Label'
import { getKycStatus } from '~/data/wallet.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { KycStatus } from '~/routes/_index/route'
import { PaymentRequiredAction } from './pay_.$paymentId/route'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const { kycStatus } = await getKycStatus(request)
  if (kycStatus != KycStatus.Approved)
    return redirect(route('/personal-details'))

  const payment = await grpc.getPayment(request, { id: params.paymentId })

  if (isConnectError(payment)) throw payment.errorResponse

  // This payment is already confirmed
  if (payment.state > 1)
    throw redirect(
      route('/accounts/:accountId/withdraw', {
        accountId: params.accountId as string
      })
    )

  const linkedAccountsResponse = await grpc.getLinkedAccounts(request, {})
  if (isConnectError(linkedAccountsResponse)) throw linkedAccountsResponse.error
  const linkedAccount = linkedAccountsResponse.linkedAccounts.find(
    (account) => account.id == payment.receiverAccount
  )

  // TODO We should probably rather fail this silently and just not show the account name - it shouldn't ever fail but shit happens
  if (!linkedAccount) throw json({}, { status: 404 })

  return jsonWithCSRF(request, {
    nickname: linkedAccount.nickname,
    fynbosEnv: process.env.FYNBOS_ENV,
    payment,
    requiresOTP: payment.requiredActions.includes(PaymentRequiredAction.OTP)
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Confirm withdraw', back: '/accounts' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Confirm withdraw'
  }
])

export default function Page() {
  const { payment, nickname, csrfToken } = useLoaderData<typeof loader>()
  const params = useParams()

  return (
    <>
      <Form
        id='withdraw-confirm'
        action={route('/accounts/:accountId/withdraw/:paymentId', {
          accountId: params.accountId as string,
          paymentId: payment.id
        })}
        method='post'
        className='hidden'
      />
      <input
        form='withdraw-confirm'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <Label className='mt-2'>Withdraw to</Label>
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <Icon>account_balance</Icon>
          <span>{nickname}</span>
        </div>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Withdraw from</span>
            <span className='text-medium'>{nickname}</span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Payment date</span>
            <span className='text-medium'>
              {DateTime.now().toFormat('dd MMMM yyyy')}
            </span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Fees</span>
            <span className='text-medium'>{payment.formattedFees}</span>
          </div>
          <div className='mt-4 flex w-full justify-between font-medium'>
            <span className='text-medium'>You will receive</span>
            <span className='text-error'>{payment.totalSendAmount}</span>
          </div>
        </CardContent>
      </Card>
      {payment.note && (
        <Card>
          <CardContent>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Payment note</span>
              <span className='text-medium'>{payment.note}</span>
            </div>
          </CardContent>
        </Card>
      )}
      <Button
        form='withdraw-confirm'
        name='formName'
        value='confirmPayment'
        type='submit'
      >
        Confirm withdraw
      </Button>
      <div className='flex w-full justify-center'>
        <span className='text-sm text-medium'>
          A withdrawal could take up to 3 business days.
        </span>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    form: ''
  }

  const clientIpAddress = getClientIP(request)

  let response = await grpc.updatePayment(request, {
    id: params.paymentId,
    ipAddress: clientIpAddress
  })
  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  response = await grpc.confirmPayment(request, {
    id: params.paymentId
  })
  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, route('/'), {
    message: 'Withdraw created successfully.',
    icon: 'close'
  })
}
