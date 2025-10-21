import clsx from 'clsx'
import { useState } from 'react'
import { StorableCard } from '~/lib/cards/types'
import { CardActions } from './CardActions'
import { CardViewBack } from './CardViewBack'
import { CardViewFront } from './CardViewFront'
import { StatusPopup } from './StatusPopup'
import { useCardActions } from './useCardActions'

export const CardView = ({ card }: { card: StorableCard }) => {
  const [showBack, setShowBack] = useState(false)

  const {
    isSensitiveDataVisible,
    isPinVisible,
    isLocked,
    sensitiveData,
    pin,
    actionStatus,
    toggleSensitiveData,
    toggleLock,
    toggleBlock,
    toggleTerminate,
    toggleUnlock,
    toggleViewPin
  } = useCardActions(card)

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
            isLocked && 'blur-sm'
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
            <CardViewFront nameOnCard={card.nameOnCard} />
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
              fullCardNumber={sensitiveData.Pan.replace(/\s+/g, '').replace(
                /(.{4})(?=.{4})/g,
                '$1 '
              )}
              expiryDate={
                sensitiveData.ExpiryDate.slice(0, 2) +
                '/' +
                sensitiveData.ExpiryDate.slice(2)
              }
              cvv={sensitiveData.Cvc2}
            />
          </div>
        </div>

        {/* Locked overlay */}
        {isLocked && (
          <div className='absolute inset-0 z-40 flex items-center justify-center rounded-xl bg-black/30'>
            <div className='rounded-lg bg-red-500 px-4 py-2 font-semibold text-white'>
              CARD LOCKED
            </div>
          </div>
        )}
      </div>

      {/* Card actions */}
      <CardActions
        showBack={showBack}
        setShowBack={setShowBack}
        isSensitiveDataVisible={isSensitiveDataVisible}
        isPinVisible={isPinVisible}
        isLocked={isLocked}
        toggleSensitiveData={toggleSensitiveData}
        toggleLock={toggleLock}
        toggleUnlock={toggleUnlock}
        toggleViewPin={toggleViewPin}
        toggleBlock={toggleBlock}
        toggleTerminate={toggleTerminate}
      />

      {/* PIN Display - Testing only */}
      {isPinVisible && (
        <div className='rounded-lg border-2 border-purple-500 bg-purple-50 px-8 py-4'>
          <div className='text-center'>
            <div className='mb-2 text-sm font-semibold uppercase tracking-wide text-purple-700'>
              Card PIN
            </div>
            <div className='text-3xl font-bold tracking-widest text-purple-900'>
              {pin}
            </div>
          </div>
        </div>
      )}

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
