import type { ComponentProps } from 'react'
import { CardViewContainer } from './CardViewContainer'
import { MasterCardLogo } from './MasterCardLogo'

interface CardViewFrontProps extends ComponentProps<'div'> {
  nameOnCard: string
  cardNumber: string
  expiryDate: string | null
}

export const CardViewFront = ({
  nameOnCard,
  cardNumber,
  expiryDate,
  className,
  ...props
}: CardViewFrontProps) => {
  // Format card number with spaces
  const displayCardNumber = cardNumber.replace(/(\d{4})(?=\d)/g, '$1 ')
  // Render censored expiry if null
  const displayExpiryDate = expiryDate || '••/••'

  return (
    <CardViewContainer className={className} {...props}>
      <div className='flex h-full flex-col'>
        {/* Header with logos */}
        <div className='flex items-center justify-between text-sm'>
          <div className='flex items-center'>
            {/* Interledger Foundation SVG Logo */}
            <img
              src='/interledger-foundation.svg'
              alt='Interledger Foundation logo'
              className='h-8 max-w-[120px] object-contain opacity-90'
              onError={(e) => {
                const target = e.target as HTMLImageElement
                target.style.display = 'none'
                const fallback = target.nextElementSibling as HTMLElement
                if (fallback) fallback.classList.remove('hidden')
              }}
            />
          </div>

          <div className='flex items-center justify-end'>
            <span className='text-xs font-light uppercase tracking-wider text-black/80'>
              debit
            </span>
          </div>
        </div>

        {/* EMV Chip */}
        <div className='mt-8'>
          <div className='w-fit'>
            <img
              src='/emv_chip.svg'
              alt='EMV chip'
              className='h-9 w-12 object-contain'
            />
          </div>
        </div>

        {/* Footer with name and mastercard logo */}
        <div className='mt-auto flex items-end justify-between'>
          <div className='flex flex-col'>
            {/* <div className='text-xs uppercase tracking-wide text-black/50'>
              CARDHOLDER NAME
            </div> */}
            <span className='text-sm font-medium uppercase text-black'>
              {nameOnCard}
            </span>
          </div>
          <MasterCardLogo />
        </div>
      </div>
    </CardViewContainer>
  )
}
