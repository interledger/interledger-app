import type { Route } from './+types/deposit_.$paymentId'
import { redirect } from 'react-router';
import { Form, useLoaderData } from 'react-router';
import { DateTime } from 'luxon'
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Icon, Layouts } from '~/components'
import { Label } from '~/components/Label'
import { getKycStatus } from '~/data/wallet.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { ErrorDescriptions } from '~/lib/error.constants'
import { isConnectError, isOtpValidationError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { usePTISdk } from '~/lib/usePTISdk'
import { KycStatus, PaymentRequiredAction } from '~/lib/types'

export async function loader({ request, params }: Route.LoaderArgs) {
  const { kycStatus } = await getKycStatus(request)
  if (kycStatus != KycStatus.Approved)
    return redirect(href('/personal-details'))

  const payment = await grpc.getPayment(request, { id: params.paymentId })

  if (isConnectError(payment)) throw payment.errorResponse

  // This payment is already confirmed
  if (payment.state > 1) throw redirect(href('/deposit'))

  const linkedAccountsResponse = await grpc.getLinkedAccounts(request, {})
  if (isConnectError(linkedAccountsResponse)) throw linkedAccountsResponse.error

  return jsonWithCSRF(request, {
    receiverAccountTitle: linkedAccountsResponse.linkedAccounts.find(
      (account) => account.id == payment.receiverAccount
    )?.title,
    senderAccountTitle: linkedAccountsResponse.linkedAccounts.find(
      (account) => account.id == payment.senderAccount
    )?.title,
    payment,
    requiresOTP: payment.requiredActions.includes(PaymentRequiredAction.OTP)
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Confirm deposit', back: '/' }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Confirm deposit'
  }
])

export default function Page() {
  const {
    payment,
    receiverAccountTitle,
    senderAccountTitle,
    csrfToken
  } = useLoaderData()

  usePTISdk(payment.id, payment.receiverAmount?.clientId ?? '')

  return (
    <>
      <Form
        id='deposit-confirm'
        action={href('/deposit/:paymentId', {
          paymentId: payment.id
        })}
        method='post'
        className='hidden'
      />
      <input
        form='deposit-confirm'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='deposit-confirm'
        value={payment.receiverAmount?.country}
        name='country'
        type='hidden'
      />
      <Card>
        <Label className='mt-2'>Deposit to</Label>
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <Icon>account_balance</Icon>
          <span>{receiverAccountTitle}</span>
        </div>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Deposit from</span>
            <span className='text-medium'>{senderAccountTitle}</span>
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
        form='deposit-confirm'
        name='formName'
        value='confirmPayment'
        type='submit'
      >
        Confirm deposit
      </Button>
    </>
  )
}

export async function action({ request, params }: Route.ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    otp: ''
  }

  const clientIpAddress = getClientIP(request)
  const country = form.get('country') as string
  if (country === 'US') {
    const response = await grpc.ptiCreateDeposit(request, {
      id: params.paymentId,
      ipAddress: clientIpAddress
    })
    if (isConnectError(response)) {
      return response.error({ errors }, {}, { action: 'Contact support' })
    }
    return redirectWithSnackbar(request, href('/'), {
      message: 'Deposit created successfully.',
      icon: 'close'
    })
  } else {
    let response = await grpc.updatePayment(request, {
      id: params.paymentId,
      ipAddress: clientIpAddress
    })
    if (isConnectError(response)) {
      if (isOtpValidationError(response)) {
        return response.error(
          { errors },
          {},
          { action: 'Contact support', message: ErrorDescriptions.INVALID_OTP }
        )
      }
      return response.error({ errors }, {}, { action: 'Contact support' })
    }

    response = await grpc.confirmPayment(request, {
      id: params.paymentId
    })
    if (isConnectError(response)) {
      return response.error({ errors }, {}, { action: 'Contact support' })
    }
  }

  return redirectWithSnackbar(request, href('/'), {
    message: 'Deposit created successfully.',
    icon: 'close'
  })
}
