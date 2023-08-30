import { useCallback } from 'react'
import { Button, Card, CardContent, Shape } from '~/components'
import {
  ConnectDomainStep,
  useConnectDomainStore
} from '~/lib/useConnectDomainStore'

export function Landing() {
  const setStep = useConnectDomainStore((state) => state.setStep)

  const _onClick = useCallback(() => {
    setStep(ConnectDomainStep.NAME)
  }, [setStep])

  return (
    <>
      <Card>
        <CardContent>
          <span>To connect a domain, please follow these simple steps:</span>
          <div className='mt-6 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-tl-full rounded-br-full'
              color='bg-lime-400'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full'
              color='bg-orange-400'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Domain name</h3>
              <p className='text-xs text-medium'>
                Provide the domain name you would like to connect.
              </p>
            </div>
          </div>
          <div className='mt-10 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full rounded-tl-none'
              color='bg-sky-300'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full rounded-br-none'
              color='bg-yellow-300'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>DNS TXT record</h3>
              <p className='text-xs text-medium'>
                Add a DNS TXT record with hostname and code provided.
              </p>
            </div>
          </div>
          <div className='mt-10 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-t-full'
              color='bg-rose-300'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-tr-full rounded-bl-full'
              color='bg-purple-400'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Propagation</h3>
              <p className='text-xs text-medium'>
                Wait until the DNS configuration changes.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
      <Button type='button' onClick={_onClick}>
        Continue
      </Button>
    </>
  )
}
