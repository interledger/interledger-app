import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button } from '~/components'
import { updateFlow, getCurrentFlow } from '~/lib/flows.server'
import { apolloClient } from '~/lib/apollo.server'
import type {
  LinkUsdBankAccountMutation,
  LinkUsdBankAccountMutationVariables,
  VerifyUsdBankAccountMutation,
  VerifyUsdBankAccountMutationVariables
} from '~/generated/types'
import {
  LinkUsdBankAccountDocument,
  VerifyUsdBankAccountDocument
} from '~/generated/types'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  return json({
    flow
  })
}

export default function Page() {
  const { flow } = useLoaderData<typeof loader>()
  const { accountNumber, institution, name, routingNumber, type } = flow?.data
  return (
    <>
      <Form
        id='linked-account-review'
        action={`/flows/${flow.id}/linked-account/review`}
        method='post'
        className='hidden'
      />

      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-sans text-sm font-medium'>Account Type</span>
        <span className='font-sans text-base font-normal'>{type}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-sans text-sm font-medium'>Institution</span>
        <span className='font-sans text-base font-normal'>{institution}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-sans text-sm font-medium'>Account number</span>
        <span className='font-sans text-base font-normal'>{accountNumber}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-sans text-sm font-medium'>Routing number</span>
        <span className='font-sans text-base font-normal'>{routingNumber}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-sans text-sm font-medium'>Nickname</span>
        <span className='font-sans text-base font-normal'>{name}</span>
      </div>

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
  const { accountNumber, institution, name, routingNumber, type } = flow?.data
  const cookie = request.headers.get('cookie')
  const linkUsdBankAccountMutationVariables = {
    input: {
      accountNumber: accountNumber,
      institution: institution,
      name: name,
      routingNumber: routingNumber,
      type: type
    }
  }

  const linkRes = await apolloClient.mutate<
    LinkUsdBankAccountMutation,
    LinkUsdBankAccountMutationVariables
  >({
    mutation: LinkUsdBankAccountDocument,
    variables: linkUsdBankAccountMutationVariables,
    context: {
      headers: {
        cookie: cookie
      }
    }
  })
  if (
    linkRes.data?.linkUsdBankAccount.success &&
    linkRes.data?.linkUsdBankAccount.fundingSource != null
  ) {
    const verifyUsdBankAccountMutationVariables = {
      input: {
        FundingSourceId: linkRes.data?.linkUsdBankAccount?.fundingSource.id
      }
    }

    const res = await apolloClient.mutate<
      VerifyUsdBankAccountMutation,
      VerifyUsdBankAccountMutationVariables
    >({
      mutation: VerifyUsdBankAccountDocument,
      variables: verifyUsdBankAccountMutationVariables,
      context: {
        headers: {
          cookie: cookie
        }
      }
    })
    const headers = await updateFlow(request, null, true)
    if (res.data?.verifyUsdBankAccount.success)
      return redirect(
        route('/confirmation/:flowId/linked-account', {
          flowId: flow?.id as string
        }),
        { headers }
      )
  }
}
