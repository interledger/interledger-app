import { data, type ActionFunctionArgs } from 'react-router';
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export type PaymentConfirmationResponse = {
  success: boolean
  result?: 'confirmed' | 'declined'
  errors?: any
}

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData()
  const transactionId = formData.get('transactionId') as string
  const confirmed = formData.get('confirmed') as 'true' | 'false'

  if (!transactionId || !confirmed) {
    return data<PaymentConfirmationResponse>(
      {
        success: false,
        errors: {
          message: 'invalid data'
        }
      },

      { status: 400 }
    )
  }

  if (confirmed !== 'true' && confirmed !== 'false') {
    return data<PaymentConfirmationResponse>(
      {
        success: false,
        errors: {
          message: 'invalid data'
        }
      },
      { status: 400 }
    )
  }

  const response = await grpc.threeDSPaymentConfirmation(request, {
    transactionId: transactionId,
    confirmed: confirmed === 'true' ? true : false
  })

  if (isConnectError(response)) {
    return data<PaymentConfirmationResponse>(
      {
        success: false,
        errors: {
          message: 'payment confirmation call failed'
        }
      },
      { status: 400 }
    )
  }

  return data<PaymentConfirmationResponse>({
    success: true,
    result: confirmed === 'true' ? 'confirmed' : 'declined'
  })
}
