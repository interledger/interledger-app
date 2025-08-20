import type { ComponentProps } from 'react'
import { Icon } from '~/components/Icon'
import { CardViewContainer } from './CardViewContainer'

interface CardViewBackProps extends ComponentProps<'div'> {
  fullCardNumber: string
  expiryDate: string
  cvv: string
  className?: string
  showSensitiveData?: boolean
}

export const CardViewBack = ({
  fullCardNumber,
  expiryDate,
  cvv,
  className,
  showSensitiveData = false,
  ...props
}: CardViewBackProps) => {
  // Format card number with spaces or hide it
  const formattedCardNumber = showSensitiveData
    ? fullCardNumber.replace(/(\d{4})(?=\d)/g, '$1 ')
    : '•••• •••• •••• ••••'

  return (
    <CardViewContainer className={className} {...props}>
      <div className='flex h-full flex-col'>
        {/* Black stripe */}
        <div className='-mx-5 mt-3 h-12 bg-black' />

        {/* Card details */}
        <div className='mt-auto space-y-6'>
          <div>
            <p className='text-xs font-medium leading-3 text-white/50'>
              Card Number
            </p>
            <div className='flex items-center gap-x-3'>
              <p className='font-mono text-lg'>{formattedCardNumber}</p>
              <button className='h-4 w-4 p-0 text-white/50 hover:text-white'>
                <Icon>content_copy</Icon>
              </button>
            </div>
          </div>
          <div className='flex gap-x-6'>
            <div>
              <p className='text-xs font-medium leading-3 text-white/50'>
                Expiry
              </p>
              <p className='font-mono text-sm'>
                {showSensitiveData ? expiryDate : '••/••'}
              </p>
            </div>
            <div>
              <p className='text-xs font-medium leading-3 text-white/50'>CVV</p>
              <p className='font-mono text-sm'>
                {showSensitiveData ? cvv : '•••'}
              </p>
            </div>
            <button className='-ml-3 mt-2.5 h-4 w-4 p-0 text-white/50 hover:text-white'>
              <Icon>content_copy</Icon>
            </button>
            {/* Mastercard logo */}
            <div className='ml-auto flex items-center'>
              <div className='h-6 w-6 rounded-full bg-red-500 opacity-80'></div>
              <div className='-ml-2 h-6 w-6 rounded-full bg-yellow-500 opacity-80'></div>
            </div>
          </div>
        </div>
      </div>
    </CardViewContainer>
  )
}
