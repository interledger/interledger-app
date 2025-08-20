import clsx from 'clsx'
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

  // Use the custom hook for card actions
  const { showSensitiveData, isFrozen, toggleSensitiveData, toggleFreeze } =
    useCardActions()

  // Mock data for card back view - in real implementation this would come from secure API
  const mockFullCardNumber = card.maskedPan
    .replace('****', '5287')
    .replace('****', '0012')
    .replace('****', '3456')
  const mockCvv = '123'

  const handleToggleSensitiveData = () => {
    toggleSensitiveData(card, () => {
      console.log('Sensitive data visibility toggled successfully')
    })
  }

  const handleToggleFreeze = () => {
    toggleFreeze(card, () => {
      console.log('Card freeze status toggled successfully')
    })
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
              fullCardNumber={mockFullCardNumber}
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
          onClick={handleToggleSensitiveData}
        >
          <Icon>{showSensitiveData ? 'visibility_off' : 'visibility'}</Icon>
          <span>{showSensitiveData ? 'Hide' : 'View'}</span>
        </button>
        <button
          className='flex w-32 items-center justify-center space-x-2 rounded-lg bg-red-500 px-4 py-2 text-white transition-colors hover:bg-red-600'
          onClick={handleToggleFreeze}
        >
          <Icon>ac_unit</Icon>
          <span>{isFrozen ? 'Unfreeze' : 'Freeze'}</span>
        </button>
      </div>
    </div>
  )
}
