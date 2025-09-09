import type { ComponentProps } from 'react'
import { CardViewContainer } from './CardViewContainer'

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
            {/* GateHub SVG Logo */}
            <img
              src='/gatehub.svg'
              alt='GateHub logo'
              className='h-8 max-w-[120px] object-contain opacity-90 brightness-0 invert filter'
              onError={(e) => {
                const target = e.target as HTMLImageElement
                target.style.display = 'none'
                const fallback = target.nextElementSibling as HTMLElement
                if (fallback) fallback.classList.remove('hidden')
              }}
            />
          </div>

          <div className='flex items-center space-x-3'>
            <span className='text-xs uppercase tracking-wider text-white/50'>
              debit
            </span>
            {/* Mastercard logo */}
            <div className='flex items-center'>
              <div className='h-6 w-6 rounded-full bg-red-500 opacity-80'></div>
              <div className='-ml-2 h-6 w-6 rounded-full bg-yellow-500 opacity-80'></div>
            </div>
          </div>
        </div>

        {/* Card number */}
        <div className='mt-8'>
          <div className='font-mono text-xl tracking-wider'>
            {displayCardNumber}
          </div>
        </div>

        {/* Footer with name and expiry */}
        <div className='mt-auto flex items-end justify-between'>
          <div className='flex flex-col'>
            <div className='text-xs uppercase tracking-wide text-white/50'>
              Card Holder
            </div>
            <span className='text-sm font-medium uppercase'>{nameOnCard}</span>
          </div>
          <div className='flex flex-col text-right'>
            <div className='text-xs uppercase tracking-wide text-white/50'>
              Expires
            </div>
            <span className='font-mono text-sm'>{displayExpiryDate}</span>
          </div>
        </div>
      </div>
    </CardViewContainer>
  )
}
