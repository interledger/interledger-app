import clsx from 'clsx'
import { useState } from 'react'
import { Icon } from '~/components/Icon'
import { CardViewBack } from './CardViewBack'
import { CardViewFront } from './CardViewFront'
import { StatusPopup } from './StatusPopup'
import { useCardActions } from './useCardActions'

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

  const {
    showSensitiveData,
    isFrozen,
    isBlocked,
    sensitiveData,
    actionStatus,
    toggleSensitiveDataOff,
    toggleSensitiveDataOn,
    toggleFreeze,
    toggleUnfreeze
  } = useCardActions(card)

  const handleToggleSensitiveData = () => {
    if (showSensitiveData) {
      toggleSensitiveDataOff()
    } else {
      toggleSensitiveDataOn()
    }
  }

  const handleToggleFreeze = () => {
    toggleFreeze()
  }

  const handleToggleUnfreeze = () => {
    toggleUnfreeze()
  }

  return (
    <div className='flex flex-col items-center space-y-6 p-6'>
      {/* Card flip container */}
      <div className='relative h-52 w-80' style={{ perspective: '1000px' }}>
        <div
          className={clsx(
            'relative h-full w-full transition-transform duration-700 ease-in-out',
            showBack
              ? '[transform:rotateY(180deg)]'
              : '[transform:rotateY(0deg)]',
            isBlocked && 'blur-sm'
          )}
          style={{ transformStyle: 'preserve-3d' }}
        >
          {/* Loading overlay */}
          {actionStatus === 'loading' && (
            <div className='absolute inset-0 z-50 flex items-center justify-center rounded-xl bg-black/50 backdrop-blur-sm'>
              <div className='flex flex-col items-center space-y-3'>
                <div className='h-8 w-8 animate-spin rounded-full border-4 border-white/30 border-t-white'></div>
                <span className='text-sm font-medium text-white'>
                  Processing...
                </span>
              </div>
            </div>
          )}
          {/* Front of card */}
          <div
            className='absolute inset-0 h-full w-full'
            style={{ backfaceVisibility: 'hidden' }}
          >
            <CardViewFront 
            // todo: remove card number and expiry
              nameOnCard={card.nameOnCard}
              cardNumber={sensitiveData.Pan}
              expiryDate={sensitiveData.ExpiryDate.slice(0, 2) + '/' + sensitiveData.ExpiryDate.slice(2)}
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
              fullCardNumber={sensitiveData.Pan.replace(/\s+/g, '').replace(/(.{4})(?=.{4})/g, '$1 ')}
              expiryDate={sensitiveData.ExpiryDate.slice(0, 2) + '/' + sensitiveData.ExpiryDate.slice(2)}
              cvv={sensitiveData.Cvc2}
            />
          </div>
        </div>

        {/* Blocked overlay - outside flip container so text doesn't flip */}
        {isBlocked && (
          <div className='absolute inset-0 z-40 flex items-center justify-center rounded-xl bg-black/30'>
            <div className='rounded-lg bg-red-500 px-4 py-2 font-semibold text-white'>
              CARD BLOCKED
            </div>
          </div>
        )}
      </div>

      {/* Card actions */}
      <div className='flex space-x-4'>
        {/* Flip card */}
        <button
          className='flex w-24 items-center justify-center space-x-2 rounded-lg bg-blue-500 px-4 py-2 text-white transition-colors hover:bg-blue-600'
          onClick={() => setShowBack(!showBack)}
        >
          <Icon>{showBack ? 'flip_to_front' : 'flip_to_back'}</Icon>
          <span>Flip</span>
        </button>
        {/* Toggle sensitive data */}
        <button
          className='flex w-24 items-center justify-center space-x-2 rounded-lg bg-green-500 px-4 py-2 text-white transition-colors hover:bg-green-600'
          onClick={handleToggleSensitiveData}
        >
          <Icon>{showSensitiveData ? 'visibility_off' : 'visibility'}</Icon>
          <span>{showSensitiveData ? 'Hide' : 'View'}</span>
        </button>
        {/* Toggle freeze */}
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-red-500 px-4 py-2 text-white transition-colors hover:bg-red-600'
          onClick={isFrozen ? handleToggleUnfreeze : handleToggleFreeze}
        >
          <Icon>ac_unit</Icon>
          <span>{isFrozen ? 'Unfreeze' : 'Freeze'}</span>
        </button>
      </div>

      {/* Status popup */}
      {(actionStatus === 'success' || actionStatus === 'error') && (
        <StatusPopup
          type={actionStatus === 'success' ? 'success' : 'error'}
          message={
            actionStatus === 'success'
              ? 'Card operation succeeded'
              : 'Card operation failed'
          }
        />
      )}
    </div>
  )
}
