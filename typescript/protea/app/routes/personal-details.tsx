import type { LoaderArgs, ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { requireUserSession } from '~/lib/kratos.server'
import { flowType, requireFlow } from '~/lib/flows.server'
import { Button, Layouts, Shape } from '~/components'
import { route } from 'routes-gen'
import { Form } from '@remix-run/react'

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  await requireFlow(request, flowType.PersonalDetails)
  return json({})
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>
        Activate payment pointer
      </h1>
      <span>
        Here’s what we will need in order to activate your payment pointer:
      </span>
      <div className='mt-6 flex items-start'>
        <Shape
          flex='flex-none'
          width={'w-8'}
          radius={'rounded-br-full'}
          color={'bg-rose-300'}
        />
        <Shape
          flex='flex-none'
          width={'w-8'}
          radius={'rounded-full'}
          color={'bg-lime-500'}
        />
        <div className='ml-5'>
          <h3 className='mb-1 font-medium text-strong'>Personal details</h3>
          <p className='text-xs text-medium'>
            Confirmation of first and last name, your date of birth and gender.
          </p>
        </div>
      </div>
      <div className='mt-10 flex items-start'>
        <Shape
          flex='flex-none'
          width={'w-8'}
          radius={'rounded-tl-full'}
          color={'bg-slate-300'}
        />
        <Shape
          flex='flex-none'
          width={'w-8'}
          radius={'rounded-full'}
          color={'bg-yellow-300'}
        />
        <div className='ml-5'>
          <h3 className='mb-1 font-medium text-strong'>Address details</h3>
          <p className='text-xs text-medium'>Your physical address details.</p>
        </div>
      </div>

      <Form
        id='personal-details'
        action='/personal-details'
        method='post'
        className='hidden'
      />
      <div className='mt-12'>
        <Button form='personal-details' type='submit'>
          Continue
        </Button>
      </div>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  await requireFlow(request, flowType.PersonalDetails)
  return redirect(route('/personal-details/about'))
}
