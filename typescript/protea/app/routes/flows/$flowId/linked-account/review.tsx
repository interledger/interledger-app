import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, TextField } from '~/components'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import { grpcClient, GrpcError, isGrpcError, StatusError } from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  const cookie = request.headers.get('cookie') || ''
  let accountDetailsRpc = await grpcClient.getBankAccountDetails(
    {
      fundingsourceId: flow?.data.fundingsourceId,
    }, 
    {
      meta: {
        cookies: cookie
      }
    }
  ).then((v) => v)
    .catch(StatusError)
  if (isGrpcError(accountDetailsRpc)) {
    throw accountDetailsRpc
  }

  let otpRpc = await grpcClient.sendOTP(
    {},
    {
      meta: { cookies: cookie } 
    }
  ).then((v) => v)
    .catch(StatusError)
  if (isGrpcError(otpRpc)) {
    throw otpRpc
  }

  return json({
    flow,
    accountNumber: accountDetailsRpc.response.mask,
    institution: accountDetailsRpc.response.institution,
    type: accountDetailsRpc.response.type,
  })
}

export default function Page() {
  const { flow, accountNumber, institution, type } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  return (
    <>
      <Form
        id='linked-account-review'
        action={`/flows/${flow.id}/linked-account/review`}
        method='post'
        className='hidden'
      />

      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Account Type</span>
        <span>{type}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Institution</span>
        <span>{institution}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Account number</span>
        <span>*******{accountNumber}</span>
      </div>

      <TextField
        id='nickname'
        form='linked-account-review'
        label='Nickname'
        name='nickname'
        defaultValue={undefined}
        type='text'
        className='col-span-full flex flex-col selection:bg-primary/50 sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.nickname) || undefined}
        aria-describedby={actionData?.errors.nickname ? 'nickname-error' : undefined}
        required
        errorMessage={actionData?.errors.nickname || undefined}
      />

      <TextField
        id='otp'
        form='linked-account-review'
        label='OTP'
        name='otp'
        defaultValue={undefined}
        type='text'
        className='col-span-full flex flex-col selection:bg-primary/50 sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.otp) || undefined}
        aria-describedby={actionData?.errors.otp ? 'otp-error' : undefined}
        required
        errorMessage={actionData?.errors.otp || undefined}
      />

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='linked-account-review' type='submit'>
          Confirm
        </Button>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const flow = await getCurrentFlow(request, params)
  const form = await request.formData()
  const nickname = form.get("nickname") as string
  const otp = form.get("otp") as string
  const cookie = request.headers.get('cookie') || ''

  let rpc = await grpcClient.continueAddingBankAccount(
    {
      fundingsourceId: flow?.data.fundingsourceId,
      nickName: nickname,
      otp: otp,
    },
    {
      meta: {
        cookies: cookie
      }
    }
  ).then(v => v)
    .catch(StatusError)

  const fieldErrors = {
    otp: '',
    nickname: '',
  }

  if (isGrpcError(rpc)) {
    if (rpc.code == 3) {
      for (let violation of (rpc as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw rpc
  }
  
  const headers = await updateFlow(request, null, true)
  return redirect(
    route('/confirmation/:flowId/linked-account', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'Otp' | 'Nickname'
function mapper(
  field: fieldErrorsMap
): 'otp' | 'nickname' | null {
  switch (field) {
    case 'Otp':
      return 'otp'
    case 'Nickname':
      return 'nickname'
    default:
      return null
  }
}
