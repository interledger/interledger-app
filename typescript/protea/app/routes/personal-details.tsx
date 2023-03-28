import type { LoaderArgs, ActionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { flowType, requireFlow } from '~/lib/flows.server'
import { Button, Card, FocusGrid, Layouts, Shape } from '~/components'
import { route } from 'routes-gen'
import { Form } from '@remix-run/react'

export async function loader({ request }: LoaderArgs) {
  await requireFlow(request, flowType.PersonalDetails)
  return json({})
}

export const handle = {
  title: 'Activate payment pointer',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Activate payment pointer'
  }
}

export default function Page() {
  return (
    <FocusGrid>
      <Card>
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
              Confirmation of first and last name, your date of birth and
              gender.
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
            <p className='text-xs text-medium'>
              Your physical address details.
            </p>
          </div>
        </div>
        <div className='mt-10 flex items-start'>
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-tl-full'}
            color={'bg-yellow-300'}
          />
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-tr-full'}
            color={'bg-rose-400'}
          />
          <div className='ml-5'>
            <h3 className='mb-1 font-medium text-strong'>Privacy and terms</h3>
            <p className='text-xs text-medium'>
              Please read and agree to the Machnet privacy policy and terms of
              service.
            </p>
          </div>
        </div>

        <Form
          id='personal-details'
          action='/personal-details'
          method='post'
          className='hidden'
        />
      </Card>
      <Button form='personal-details' type='submit'>
        Continue
      </Button>
    </FocusGrid>
  )
}

export async function action({ request }: ActionArgs) {
  await requireFlow(request, flowType.PersonalDetails)
  return redirect(route('/personal-details/about'))
}
