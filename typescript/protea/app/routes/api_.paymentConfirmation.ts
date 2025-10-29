import { json, type ActionFunctionArgs } from '@remix-run/node'

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
    return json<PaymentConfirmationResponse>(
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
    return json<PaymentConfirmationResponse>(
      {
        success: false,
        errors: {
          message: 'invalid data'
        }
      },
      { status: 400 }
    )
  }

  // const response = await grpc.threeDSPaymentConfirmation(request, {
  //   transactionId: transactionId,
  //   confirmed: confirmed === 'true' ? true : false
  // })
  //
  // if (isConnectError(response)) {
  //   return json<PaymentConfirmationResponse>(
  //     {
  //       success: false,
  //       errors: {
  //         message: 'payment confirmation call failed'
  //       }
  //     },
  //     { status: 400 }
  //   )
  // }

  return json<PaymentConfirmationResponse>({
    success: true,
    result: confirmed === 'true' ? 'confirmed' : 'declined'
  })
}
