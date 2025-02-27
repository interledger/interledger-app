import { useCallback } from 'react'
import { route } from 'routes-gen'
import { Button, Card, CardContent, Router, Shape } from '~/components'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'

export function Landing() {
  const setStep = useSignupStore((state) => state.setStep)

  const _onClick = useCallback(() => {
    console.log('clicked')
    setStep(SignupStep.ABOUT)
  }, [setStep])

  return (
    <>
      <Card>
        <CardContent>
          <span>Sign up for an account by completing these steps:</span>
          <div className='mt-6 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full rounded-bl-none'
              color='bg-yellow-300'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full rounded-tr-none'
              color='bg-indigo-400'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Profile details</h3>
              <p className='text-xs text-medium'>
                Submit your legal name, email, and country of residence.
              </p>
            </div>
          </div>
          <div className='mt-10 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-tr-full rounded-bl-full'
              color='bg-rose-400'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full'
              color='bg-lime-400'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>
                Mobile phone number
              </h3>
              <p className='text-xs text-medium'>
                Provide a mobile phone number we can verify.
              </p>
            </div>
          </div>
          <div className='mt-10 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-r-full'
              color='bg-sky-300'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-bl-full'
              color='bg-slate-300'
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
              width='w-8'
              radius='rounded-full'
              color='bg-rose-500'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full rounded-bl-none'
              color='bg-lime-300'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Wallet address</h3>
              <p className='text-xs text-medium'>
                Create a unique wallet address.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
      <Button type='button' onClick={_onClick}>
        Let's get started
      </Button>
      <div className='flex justify-center'>
        <p className='text-sm font-medium'>
          Have an account?{' '}
          <Router className='font-medium text-primary' to={route('/login')}>
            Log in
          </Router>
        </p>
      </div>
    </>
  )
}
