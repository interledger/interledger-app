import { Code } from '@bufbuild/connect'
import { href, redirect, type ActionFunctionArgs } from 'react-router'
import { validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

function stringToBigInt(amount: string) {
  if (amount == '') return BigInt(0)
  const dotIndex = amount.lastIndexOf('.')
  if (dotIndex > -1) {
    const amounts = amount.split('.')
    return BigInt(amounts[0] + amounts[1].slice(0, 2).padEnd(2, '0'))
  }
  return BigInt(parseFloat(amount) * 100)
}

export async function ilpDepositAction({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const depositAmount = String(form.get('depositAmount') || '')
  const toLinkedAccount = form.get('toLinkedAccount') as string
  const fromLinkedAccount = form.get('fromLinkedAccount') as string
  const provider = form.get('provider') as string
  const note = form.get('note') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    depositAmount: '',
    toLinkedAccount: '',
    note: ''
  }

  const cc = provider === 'pit' ? 'USD' : 'ZAR'
  const depositResponse = await grpc.depositBalance(request, {
    fromLinkedAccount: fromLinkedAccount,
    toLinkedAccount: toLinkedAccount,
    amount: {
      amount: stringToBigInt(depositAmount),
      asset: cc,
      assetScale: 2
    },
    note
  })
  if (isConnectError(depositResponse)) {
    if (depositResponse.code == Code.InvalidArgument) {
      return depositResponse.error({
        errors: {
          ...errors,
          depositAmount:
            depositResponse?.fieldViolations?.find(
              (v: { field: string }) => v.field === 'amount'
            )?.description ?? ''
        }
      })
    }
    if (
      depositResponse.code == Code.FailedPrecondition &&
      depositResponse.violations.findIndex(
        (violation) =>
          violation.type === 'Payment' &&
          violation.subject === 'insufficientFunds'
      ) > -1
    ) {
      return depositResponse.error({
        errors: {
          ...errors,
          depositAmount: 'You have insufficient funds available.'
        }
      })
    }
    errors.form = 'Failed to create deposital.'
    return depositResponse.error(
      { errors },
      {},
      {
        action: 'Contact support'
      }
    )
  }

  return redirect(
    href('/deposit/:paymentId', {
      paymentId: depositResponse.id
    })
  )
}

export async function xagoTestAccountDepositAction({
  request
}: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const response = await grpc.depositTestXago(request, {})

  if (isConnectError(response)) {
    return response.errorResponse
  }

  return redirectWithSnackbar(request, href('/'), {
    message: 'Test deposit added successfully.',
    icon: 'close'
  })
}
