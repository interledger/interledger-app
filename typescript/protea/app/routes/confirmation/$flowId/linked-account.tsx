import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button } from '~/components'
import { exitFlow, getCurrentFlow } from '~/lib/flows.server'

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
        id='linked-account-confirmation'
        action={`/confirmation/${flow.id}/linked-account`}
        method='post'
        className='hidden'
      />
      <div className='col-span-full flex flex-col pb-8 pt-4 text-strong sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-4xl font-medium'>
          Payment method added
        </span>
      </div>
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
        <Button form='linked-account-confirmation' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const headers = await exitFlow(request)
  return redirect(route('/settings'), { headers })
}
