import clsx from 'clsx'
import { useState } from 'react'
import { Icon } from '~/components/Icon'
import { StorableCard } from '~/lib/gatehub/types'
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
          onClick={toggleSensitiveData}
        >
          <Icon>{isSensitiveDataVisible ? 'visibility_off' : 'visibility'}</Icon>
          <span>{isSensitiveDataVisible ? 'Hide' : 'View'}</span>
        </button>
        {/* Toggle freeze */}
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-red-500 px-4 py-2 text-white transition-colors hover:bg-red-600'
          onClick={isLocked ? toggleUnlock : toggleLock
          }
        >
          <Icon>lock</Icon>
          <span>{isLocked ? 'Unlock' : 'Lock'}</span>
        </button>

        {/* Toggle view pin code */}
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-purple-500 px-4 py-2 text-white transition-colors hover:bg-purple-600'
          onClick={toggleViewPin}
        >
          <Icon>{isPinVisible ? 'visibility_off' : 'visibility'}</Icon>
          <span>{isPinVisible ? 'Hide' : 'View'} PIN</span>
        </button>

        {/* Toggle block card */}
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-purple-500 px-4 py-2 text-white transition-colors hover:bg-purple-600'
          onClick={toggleBlock}
        >
          <Icon>block</Icon>
          <span>Block</span>
        </button>

        {/* Toggle block card */}
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-purple-500 px-4 py-2 text-white transition-colors hover:bg-purple-600'
          onClick={toggleTerminate}
        >
          <Icon>delete</Icon>
          <span>Terminate</span>
        </button>
      </div>

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
