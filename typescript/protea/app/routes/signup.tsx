import type { LoaderArgs, ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { requireNoUserSession } from '~/lib/kratos.server'
import { flowType, getCurrentFlow, requireFlow } from '~/lib/flows.server'
import { Button, Router, Shape } from '~/components'
import { route } from 'routes-gen'
import { Form } from '@remix-run/react'

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  const headers = await requireFlow(request, flowType.Signup)
  return json({}, { headers })
}

export default function Page() {
  return (
    <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto rounded-2xl bg-page px-4 pb-16 pt-6 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:pt-12 xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2'>
        <h1 className='mb-4 font-display text-2xl font-medium'>Sign up</h1>
        <h1>Here's what we will need to create your account.</h1>
      </div>
      <div className='col-span-1 mt-4 flex items-start sm:col-start-2'>
        <Shape
          width={'w-8'}
          radius={'rounded-tl-full'}
          color={'bg-yellow-300'}
        />
        <Shape
          width={'w-8'}
          radius={'rounded-bl-full'}
          color={'bg-slate-500'}
        />
      </div>
      <div className='col-span-3 col-start-2 mt-4 font-medium sm:col-span-5 sm:col-start-3'>
        <h3 className='mb-1 font-medium'>Personal details</h3>
        <p className='text-sm text-medium'>
          Your name, email address and country of residence.
        </p>
      </div>
      <div className='col-span-1 mt-10 flex items-start sm:col-start-2'>
        <Shape
          width={'w-8'}
          radius={'rounded-bl-full'}
          color={'bg-slate-300'}
        />
        <Shape width={'w-8'} radius={'rounded-full'} color={'bg-lime-500'} />
      </div>
      <div className='col-span-3 col-start-2 mt-10 font-medium sm:col-span-5 sm:col-start-3'>
        <h3 className='mb-1 font-medium'>Phone number</h3>
        <p className='text-sm text-medium'>
          A mobile phone number we can verify.
        </p>
      </div>
      <div className='col-span-1 mt-10 flex items-start sm:col-start-2'>
        <Shape width={'w-8'} radius={'rounded-bl-full'} color={'bg-rose-400'} />
        <Shape
          width={'w-8'}
          radius={'rounded-tl-full'}
          color={'bg-yellow-300'}
        />
      </div>
      <div className='col-span-3 col-start-2 mt-10 font-medium sm:col-span-5 sm:col-start-3'>
        <h3 className='mb-1 font-medium'>Password</h3>
        <p className='text-sm text-medium'>Create a password we can verify.</p>
      </div>
      <Form id='signup' action={'/signup'} method='post' className='hidden' />
      <div className='col-span-full mt-10 sm:col-span-6 sm:col-start-2'>
        <Button form='signup' type='submit'>
          Let's get started
        </Button>
      </div>
      <div className='col-span-full mt-2 flex justify-center'>
        <p className='text-sm'>
          Have an account?{' '}
          <Router className='text-primary' to={route('/login')}>
            Log in
          </Router>
        </p>
      </div>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const flow = await getCurrentFlow(request, flowType.Signup)

  return redirect(
    route('/signup/:flowId/about', {
      flowId: flow.id
    })
  )
}
