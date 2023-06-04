import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, Layouts, Router, Shape } from '~/components'
import { flowType, requireFlow } from '~/lib/flows.server'
import { requireNoUserSession } from '~/lib/kratos.server'
import { canSignup } from '~/lib/signupCheck.server'

export async function loader({ request }: LoaderArgs) {
  await canSignup(request)
  await requireNoUserSession(request)
  await requireFlow(request, flowType.Signup)
  return json({})
}

export const handle = {
  title: 'Sign up',
  layout: Layouts.Focus
}

export const meta: MetaFunction = () => {
  return {
    title: 'Sign up'
  }
}

export default function Page() {
  return (
    <>
      <Form id='signup' action='/signup' method='post' className='hidden' />
      <Card>
        <span>Here's what we will need to create your account:</span>
        <div className='mt-6 flex items-start'>
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-tl-full'}
            color={'bg-yellow-300'}
          />
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-bl-full'}
            color={'bg-slate-500'}
          />
          <div className='ml-5'>
            <h3 className='mb-1 font-medium text-strong'>Profile details</h3>
            <p className='text-xs text-medium'>
              Your legal name, email address and country of residence.
            </p>
          </div>
        </div>
        <div className='mt-10 flex items-start'>
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-bl-full'}
            color={'bg-rose-400'}
          />
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-full'}
            color={'bg-lime-500'}
          />
          <div className='ml-5'>
            <h3 className='mb-1 font-medium text-strong'>
              Mobile phone number
            </h3>
            <p className='text-xs text-medium'>
              A mobile phone number we can verify.
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
            radius={'rounded-bl-full'}
            color={'bg-slate-300'}
          />
          <div className='ml-5'>
            <h3 className='mb-1 font-medium text-strong'>Password</h3>
            <p className='text-xs text-medium'>
              Create a password we can verify.
            </p>
          </div>
        </div>
        <div className='mt-10 flex items-start'>
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-full'}
            color={'bg-rose-500'}
          />
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-tr-full'}
            color={'bg-lime-300'}
          />
          <div className='ml-5'>
            <h3 className='mb-1 font-medium text-strong'>Payment pointer</h3>
            <p className='text-xs text-medium'>
              Create a unique payment pointer.
            </p>
          </div>
        </div>
      </Card>
      <Button form='signup' type='submit'>
        Let's get started
      </Button>
      <div className='flex justify-center'>
        <p className='text-sm'>
          Have an account?{' '}
          <Router className='text-primary' to={route('/login')}>
            Log in
          </Router>
        </p>
      </div>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  await requireFlow(request, flowType.Signup)
  return redirect(route('/signup/about'))
}
