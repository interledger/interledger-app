import { useState } from 'react'

interface Card {
  id: string
  nameOnCard: string
  maskedPan: string
  expiryDate: string
  status: number
  lockLevel: number
}

interface UseCardActionsOptions {
  initialShowSensitiveData?: boolean
  initialIsFrozen?: boolean
}

interface UseCardActionsReturn {
  showSensitiveData: boolean
  isFrozen: boolean
  toggleSensitiveData: (card: Card, onSuccess?: () => void) => void
  toggleFreeze: (card: Card, onSuccess?: () => void) => void
}

export const useCardActions = (
  options: UseCardActionsOptions = {}
): UseCardActionsReturn => {
  const { initialShowSensitiveData = false, initialIsFrozen = false } = options

  const [showSensitiveData, setShowSensitiveData] = useState(
    initialShowSensitiveData
  )
  const [isFrozen, setIsFrozen] = useState(initialIsFrozen)

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
