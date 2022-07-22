import { useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Router, Icon, RadioGroup } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'
import { apolloClient } from '~/lib/apollo.server'
import type {
  FlowsDepositPaymentMethodQuery,
  FlowsDepositPaymentMethodQueryVariables
} from '~/generated/types'
import { FlowsDepositPaymentMethodDocument } from '~/generated/types'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  // TODO fetch current payment methods
  const cookie = String(request.headers.get('cookie'))

  const res = await apolloClient.query<
    FlowsDepositPaymentMethodQuery,
    FlowsDepositPaymentMethodQueryVariables
  >({
    query: FlowsDepositPaymentMethodDocument,
    context: {
      headers: {
        cookie: cookie
      }
    }
  })
  const paymentMethods = res.data.fundingSources.map((fs) => ({
    id: fs?.id,
    name: fs?.name,
    description: fs?.mask,
    icon: 'account_balance' // TODO: get actual icon from fundingsource subtype
  }))

  return json({
    paymentMethods,
    flow
  })
}

export default function Page() {
  const { paymentMethods, flow } = useLoaderData<typeof loader>()

  const [selected, setSelected] = useState(paymentMethods[0])

  return (
    <>
      <Form
        id='payment-method'
        action={`/flows/${flow.id}/deposit/payment-method`}
        method='post'
        className='hidden'
      />
      {paymentMethods.length == 0 && (
        <div className='col-span-full flex items-center justify-between space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Icon>tips_and_updates</Icon>
          <span className='font-sans text-sm font-normal'>
            You need to add a payment method before you can deposit money.
          </span>
        </div>
      )}
      {paymentMethods.length > 0 && (
        <>
          <RadioGroup
            className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'
            id='radio'
            label='Payment method'
            value={selected}
            onChange={setSelected}
            options={paymentMethods}
          />
          <input
            form='payment-method'
            value={String(selected.id)}
            name='id'
            type='hidden'
          />
          <input
            form='payment-method'
            value={String(selected.description)}
            name='mask'
            type='hidden'
          />
        </>
      )}
      <Router
        to={route('/flows/:flowId/payment-method/type', {
          flowId: 'init'
        })}
        className='col-span-full mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <span className='font-sans text-base font-normal'>
          New payment method
        </span>
        <Icon>navigate_next</Icon>
      </Router>
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button
          form='payment-method'
          disabled={paymentMethods.length == 0}
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
  const paymentMethodId = form.get('id')
  const paymentMethodMask = form.get('mask')
  await stepFlow(request, {
    paymentMethodId,
    paymentMethodMask
  })
}
