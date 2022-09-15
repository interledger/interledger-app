import { useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Router, Icon, RadioGroup } from '~/components'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  // TODO fetch current payment methods
  const cookie = String(request.headers.get('cookie'))

  const response = await grpcClient
    .getFundingsources(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  const linkedAccounts = response.response.fundingsources.map((fs) => ({
    id: fs?.id,
    name: fs?.name,
    description: fs?.mask,
    icon: 'account_balance' // TODO: get actual icon from fundingsource subtype
  }))

  return json({
    linkedAccounts,
    flow
  })
}

export default function Page() {
  const { linkedAccounts, flow } = useLoaderData<typeof loader>()

  const [selected, setSelected] = useState(linkedAccounts[0])

  return (
    <>
      <Form
        id='linked-account'
        action={`/flows/${flow.id}/withdraw/linked-account`}
        method='post'
        className='hidden'
      />
      {linkedAccounts.length == 0 && (
        <div className='col-span-full flex items-center justify-between space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Icon>tips_and_updates</Icon>
          <span className='text-sm'>
            You need to add a payment method before you can withdraw money.
          </span>
        </div>
      )}
      {linkedAccounts.length > 0 && (
        <>
          <RadioGroup
            className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'
            id='radio'
            label='Payment method'
            value={selected}
            onChange={setSelected}
            options={linkedAccounts}
          />
          <input
            form='linked-account'
            value={String(selected.id)}
            name='id'
            type='hidden'
          />
          <input
            form='linked-account'
            value={String(selected.description)}
            name='mask'
            type='hidden'
          />
        </>
      )}
      <Router
        to={route('/flows/:flowId/linked-account/type', {
          flowId: 'init'
        })}
        className='col-span-full mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <span>New payment method</span>
        <Icon>navigate_next</Icon>
      </Router>
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button
          form='linked-account'
          disabled={linkedAccounts.length == 0}
          type='submit'
        >
          Continue
        </Button>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const linkedAccountId = form.get('id')
  const linkedAccountMask = form.get('mask')
  const flow = await getCurrentFlow(request, params)
  const headers = await updateFlow(request, {
    linkedAccountId,
    linkedAccountMask
  })
  return redirect(
    route('/flows/:flowId/withdraw/amount', { flowId: flow?.id as string }),
    { headers }
  )
}
