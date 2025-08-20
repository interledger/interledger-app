import clsx from 'clsx'
import type { ComponentProps } from 'react'
import { useState } from 'react'
import { Icon } from '~/components/Icon'

export type CardViewContainerProps = ComponentProps<'div'>

const CardViewContainer = ({
  children,
  className,
  ...props
}: CardViewContainerProps) => {
  return (
    <div
      className={clsx(
        'relative h-52 w-80 overflow-hidden rounded-xl bg-gradient-to-br from-slate-800 to-slate-900 px-5 py-4 font-sans text-white shadow-lg',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}

interface CardViewFrontProps extends ComponentProps<'div'> {
  nameOnCard: string
  maskedPan: string
  expiryDate: string
  isBlocked?: boolean
  showSensitiveData?: boolean
}

export const CardViewFront = ({
  nameOnCard,
  maskedPan,
  expiryDate,
  isBlocked = false,
  showSensitiveData = false,
  className,
  ...props
}: CardViewFrontProps) => {
  // Extract last 4 digits from maskedPan or hide completely
  const last4Digits = maskedPan.slice(-4)
  const displayCardNumber = showSensitiveData
    ? `•••• •••• •••• ${last4Digits}`
    : '•••• •••• •••• ••••'

  return (
    <CardViewContainer className={className} {...props}>
      <div
        className={clsx(
          'flex h-full flex-col',
          isBlocked ? 'pointer-events-none select-none blur-sm' : ''
        )}
      >
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
            <span className='font-mono text-sm'>
              {showSensitiveData ? expiryDate : '••/••'}
            </span>
          </div>
        </div>
      </div>

      {/* Blocked overlay */}
      {isBlocked && (
        <div className='absolute inset-0 z-10 flex items-center justify-center bg-black/30'>
          <div className='rounded-lg bg-red-500 px-4 py-2 font-semibold text-white'>
            CARD BLOCKED
          </div>
        </div>
      )}
    </CardViewContainer>
  )
}

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

interface CardViewProps {
  card: {
    id: string
    nameOnCard: string
    maskedPan: string
    expiryDate: string
    status: number
    lockLevel: number
  }
}

export const CardView = ({ card }: CardViewProps) => {
  const [showBack, setShowBack] = useState(false)
  const [showSensitiveData, setShowSensitiveData] = useState(false)
  const [isFrozen, setIsFrozen] = useState(false)
  const isBlocked = card.status !== 1 || card.lockLevel !== 0

  // Mock data for card back view - in real implementation this would come from secure API
  const mockFullCardNumber = card.maskedPan
    .replace('****', '5287')
    .replace('****', '0012')
    .replace('****', '3456')
  const mockCvv = '123'

  return (
    <div className='flex flex-col items-center space-y-6 p-6'>
      {/* Card flip container */}
      <div className='relative h-52 w-80' style={{ perspective: '1000px' }}>
        <div
          className={clsx(
            'relative h-full w-full transition-transform duration-700 ease-in-out',
            showBack
              ? '[transform:rotateY(180deg)]'
              : '[transform:rotateY(0deg)]'
          )}
          style={{ transformStyle: 'preserve-3d' }}
        >
          {/* Front of card */}
          <div
            className='absolute inset-0 h-full w-full'
            style={{ backfaceVisibility: 'hidden' }}
          >
            <CardViewFront
              nameOnCard={card.nameOnCard}
              maskedPan={card.maskedPan}
              expiryDate={card.expiryDate}
              isBlocked={isBlocked}
              showSensitiveData={showSensitiveData}
            />
          </div>

          {/* Back of card */}
          <div
            className='absolute inset-0 h-full w-full'
            style={{
              backfaceVisibility: 'hidden',
              transform: 'rotateY(180deg)'
            }}
          >
            <CardViewBack
              fullCardNumber={mockFullCardNumber}
              expiryDate={card.expiryDate}
              cvv={mockCvv}
              showSensitiveData={showSensitiveData}
            />
          </div>
        </div>
      </div>

      {/* Card actions */}
      <div className='flex space-x-4'>
        <button
          className='flex w-24 items-center justify-center space-x-2 rounded-lg bg-blue-500 px-4 py-2 text-white transition-colors hover:bg-blue-600'
          onClick={() => setShowBack(!showBack)}
        >
          <Icon>{showBack ? 'flip_to_front' : 'flip_to_back'}</Icon>
          <span>Flip</span>
        </button>
        <button
          className='flex w-24 items-center justify-center space-x-2 rounded-lg bg-green-500 px-4 py-2 text-white transition-colors hover:bg-green-600'
          onClick={() => setShowSensitiveData(!showSensitiveData)}
        >
          <Icon>{showSensitiveData ? 'visibility_off' : 'visibility'}</Icon>
          <span>{showSensitiveData ? 'Hide' : 'View'}</span>
        </button>
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-red-500 px-4 py-2 text-white transition-colors hover:bg-red-600'
          onClick={() => setIsFrozen(!isFrozen)}
        >
          <Icon>ac_unit</Icon>
          <span>{isFrozen ? 'Unfreeze' : 'Freeze'}</span>
        </button>
      </div>
    </div>
  )
}
