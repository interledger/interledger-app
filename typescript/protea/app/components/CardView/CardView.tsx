import clsx from 'clsx'
import { motion } from 'framer-motion'
import { useState } from 'react'
import { Icon } from '~/components/Icon'
import { CardViewBack } from './CardViewBack'
import { CardViewFront } from './CardViewFront'
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
  const isBlocked = card.status !== 1 || card.lockLevel !== 0

  const {
    showSensitiveData,
    isFrozen,
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
              : '[transform:rotateY(0deg)]'
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
              nameOnCard={card.nameOnCard}
              cardNumber={sensitiveData.cardNumber}
              expiryDate={sensitiveData.expiryDate}
              isBlocked={isBlocked}
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
              fullCardNumber={sensitiveData.cardNumber}
              expiryDate={sensitiveData.expiryDate}
              cvv={sensitiveData.cvv}
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
          onClick={handleToggleSensitiveData}
        >
          <Icon>{showSensitiveData ? 'visibility_off' : 'visibility'}</Icon>
          <span>{showSensitiveData ? 'Hide' : 'View'}</span>
        </button>
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-red-500 px-4 py-2 text-white transition-colors hover:bg-red-600'
          onClick={isFrozen ? handleToggleUnfreeze : handleToggleFreeze}
        >
          <Icon>ac_unit</Icon>
          <span>{isFrozen ? 'Unfreeze' : 'Freeze'}</span>
        </button>
      </div>

      {/* Status popup - following SnackbarStage conventions */}
      {(actionStatus === 'success' || actionStatus === 'error') && (
        <div className='fixed bottom-4 left-0 z-50 mx-auto flex w-full justify-center px-4'>
          <motion.div
            key={actionStatus}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            initial={{ opacity: 0, scale: 0.5, y: 8 }}
            exit={{
              opacity: 0,
              scale: 0.5,
              y: 8,
              transition: {
                duration: 0.2
              }
            }}
            transition={{
              type: 'spring',
              stiffness: 400,
              damping: 20,
              duration: 0.3
            }}
            className={clsx(
              'flex w-full items-center space-x-3 overflow-hidden rounded-xl px-4 py-3 text-left shadow-lg sm:max-w-xs',
              actionStatus === 'success'
                ? 'bg-green-500 text-white'
                : 'bg-red-500 text-white'
            )}
          >
            <Icon>{actionStatus === 'success' ? 'check_circle' : 'error'}</Icon>
            <span className='text-sm'>
              {actionStatus === 'success'
                ? 'Card operation succeeded'
                : 'Card operation failed'}
            </span>
          </motion.div>
        </div>
      )}
    </div>
  )
}
