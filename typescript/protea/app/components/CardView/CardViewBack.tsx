import type { ComponentProps } from 'react'
import { Icon } from '~/components/Icon'
import { CardViewContainer } from './CardViewContainer'
import { MasterCardLogo } from './MasterCardLogo'

interface CardViewBackProps extends ComponentProps<'div'> {
  fullCardNumber: string
  expiryDate: string | null
  cvv: string | null
  className?: string
}

export const CardViewBack = ({
  fullCardNumber,
  expiryDate,
  cvv,
  className,
  ...props
}: CardViewBackProps) => {
  // Format card number with spaces
  const formattedCardNumber = fullCardNumber.replace(/(\d{4})(?=\d)/g, '$1 ')
  // Render censored versions if props are null
  const displayExpiryDate = expiryDate || '**/**'
  const displayCvv = cvv || '***'

  return (
    <CardViewContainer className={className} {...props}>
      <div className='flex h-full flex-col'>
        {/* interledger.org text */}
        <div className='mt-2 text-left'>
          <p className='text-xs text-black'>interledger.org</p>
        </div>
        {/* Dark teal stripe */}
        <div
          className='-mx-5 mt-1 h-8'
          style={{ backgroundColor: '#035d5e' }}
        />

        {/* Card details */}
        <div className='mt-auto space-y-6'>
          <div>
            {/* <p className='text-xs font-medium leading-3 text-black/70'>
                Card Number
              </p> */}
            <div className='flex items-center gap-x-3'>
              <p className='font-mono text-lg text-black'>
                {formattedCardNumber}
              </p>
              <button className='h-4 w-4 p-0 text-black/50 hover:text-black'>
                <Icon>content_copy</Icon>
              </button>
            </div>
          </div>
          <div className='flex gap-x-6'>
            <div>
              <p className='text-xs font-medium leading-3 text-black/70'>
                Expiry
              </p>
              <p className='font-mono text-sm text-black'>
                {displayExpiryDate}
              </p>
            </div>
            <div>
              <p className='text-xs font-medium leading-3 text-black/70'>CVV</p>
              <p className='font-mono text-sm text-black'>{displayCvv}</p>
            </div>
            <button className='-ml-3 mt-2.5 h-4 w-4 p-0 text-black/50 hover:text-black'>
              <Icon>content_copy</Icon>
            </button>
            {/* Mastercard logo */}
            <div className='ml-auto'>
              <MasterCardLogo />
            </div>
          </div>
        </div>
      </div>
    </CardViewContainer>
  )
}
