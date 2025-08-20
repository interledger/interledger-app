import { useEffect, useState } from 'react'

interface Card {
  id: string
  nameOnCard: string
  maskedPan: string
  expiryDate: string
  status: number
  lockLevel: number
}

interface UseCardActionsReturn {
  showSensitiveData: boolean
  isFrozen: boolean
  toggleSensitiveData: (card: Card, onSuccess?: () => void) => void
  toggleFreeze: (card: Card, onSuccess?: () => void) => void
}

export const useCardActions = (cardId: string): UseCardActionsReturn => {
  const [showSensitiveData, setShowSensitiveData] = useState(false)
  const [isFrozen, setIsFrozen] = useState(false)

  // Reset state when card changes
  useEffect(() => {
    setShowSensitiveData(false)
    setIsFrozen(false)
  }, [cardId])

  const toggleSensitiveData = (card: Card, onSuccess?: () => void) => {
    try {
      console.log(`Toggling sensitive data visibility for card ${card.id}`)

      setShowSensitiveData((prev) => !prev)

      if (onSuccess) {
        onSuccess()
      }
    } catch (error) {
      console.error('Failed to toggle sensitive data visibility:', error)
    }
  }

  const toggleFreeze = (card: Card, onSuccess?: () => void) => {
    try {
      console.log(`${isFrozen ? 'Unfreezing' : 'Freezing'} card ${card.id}`)

      setIsFrozen((prev) => !prev)

      if (onSuccess) {
        onSuccess()
      }
    } catch (error) {
      console.error('Failed to toggle card freeze status:', error)
    }
  }

  return {
    showSensitiveData,
    isFrozen,
    toggleSensitiveData,
    toggleFreeze
  }
}
