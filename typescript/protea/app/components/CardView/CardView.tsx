import clsx from 'clsx'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'
import type { SerializedCard } from '~/lib/cards/types'
import { usePinChangePopup } from '~/lib/usePinChangePopup'
import { GridColumn } from '../Grid'
import { CardActions } from './CardActions'
import type { SerializedTransaction } from './CardTransactions'
import { CardTransactions } from './CardTransactions'
import { CardViewBack } from './CardViewBack'
import { CardViewFront } from './CardViewFront'
import { StatusPopup } from './StatusPopup'
import { TimedPinPopup } from './TimedPinPopup'
import { useCardActions } from './useCardActions'

interface CardViewProps {
  card: SerializedCard
  transactions: SerializedTransaction[]
}

export const CardView = ({ card, transactions }: CardViewProps) => {
  const {
    flip,
    showBack,
    isSensitiveDataVisible,
    isPinVisible,
    isFrozen,
    isBlockedByAdmin,
    sensitiveData,
    pin,
    actionStatus,
    toggleSensitiveDataOn,
    toggleFreeze,
    toggleUnfreeze,
    toggleBlock,
    toggleViewPin,
    toggleChangePin
  } = useCardActions(card)

  const { PinChangePopup } = usePinChangePopup()

  if (isBlockedByAdmin) {
    return (
      <div className='flex flex-col items-center space-y-6 p-6'>
        <div className='relative h-52 w-80' style={{ perspective: '1000px' }}>
          <div className='relative h-full w-full'>
            {/* Grey card skeleton */}
            <div className='absolute inset-0 h-full w-full'>
              <CardViewFront
                nameOnCard={card.nameOnCard}
                className='opacity-50 grayscale'
              />
            </div>
          </div>

          {/* Blocked overlay */}
          <div className='absolute inset-0 z-40 flex items-center justify-center rounded-xl bg-black/30'>
            <div className='rounded-lg bg-red-500 px-4 py-2 font-semibold text-white'>
              Card is blocked by admin. <br /> Please contact support.
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <GridColumn>
      {/* Card flip container */}
      <div
        className='relative mx-auto h-52 w-[21rem]'
        style={{ perspective: '1000px' }}
      >
        <div
          className={clsx(
            'relative h-full w-full transition-transform duration-700 ease-in-out',
            showBack
              ? '[transform:rotateY(180deg)]'
              : '[transform:rotateY(0deg)]',
            isFrozen && 'blur-sm'
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
              fullCardNumber={sensitiveData.Pan}
              expiryDate={
                sensitiveData.ExpiryDate.slice(0, 2) +
                '/' +
                sensitiveData.ExpiryDate.slice(2)
              }
              cvv={sensitiveData.Cvc2}
            />
          </div>
        </div>

        {/* Frozen overlay */}
        {isFrozen && (
          <div className='absolute inset-0 z-40 flex items-center justify-center rounded-xl bg-black/30'>
            <div className='rounded-lg bg-red-500 px-4 py-2 font-semibold text-white'>
              Card was frozen
            </div>
          </div>
        )}
      </div>
      {/* Card actions */}
      <CardActions
        isPhysical={card.type === CardType.PHYSICAL}
        showBack={showBack}
        flip={flip}
        isSensitiveDataVisible={isSensitiveDataVisible}
        isPinVisible={isPinVisible}
        isFrozen={isFrozen}
        toggleSensitiveDataOn={toggleSensitiveDataOn}
        toggleFreeze={toggleFreeze}
        toggleUnfreeze={toggleUnfreeze}
        toggleViewPin={toggleViewPin}
        toggleBlock={toggleBlock}
        toggleChangePin={toggleChangePin}
      />

      {/* PIN Display Popup */}
      <TimedPinPopup
        pin={pin}
        isVisible={isPinVisible}
        onClose={toggleViewPin}
        duration={7}
      />

      {/* PIN Change Popup */}
      <PinChangePopup />

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
      <CardTransactions transactions={transactions} />
    </GridColumn>
  )
}
