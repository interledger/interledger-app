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
  toggleSensitiveDataOff: (onSuccess?: () => void) => void
  toggleSensitiveDataOn: (onSuccess?: () => void) => void
  toggleFreeze: (onSuccess?: () => void) => void
  toggleUnfreeze: (onSuccess?: () => void) => void
}

export const useCardActions = (card: Card): UseCardActionsReturn => {
  const [showSensitiveData, setShowSensitiveData] = useState(false)
  const [isFrozen, setIsFrozen] = useState(false)

  // Reset state when card changes
  useEffect(() => {
    setShowSensitiveData(false)
    setIsFrozen(false)
  }, [card])

  const toggleSensitiveDataOff = (onSuccess?: () => void) => {
    try {
      console.log(`Turning off sensitive data visibility for card ${card.id}`)

      setShowSensitiveData(false)

      if (onSuccess) {
        onSuccess()
      }
    } catch (error) {
      console.error('Failed to turn off sensitive data visibility:', error)
    }
  }

  const toggleSensitiveDataOn = (onSuccess?: () => void) => {
    try {
      console.log(`Turning on sensitive data visibility for card ${card.id}`)

      setShowSensitiveData(true)

      if (onSuccess) {
        onSuccess()
      }
    } catch (error) {
      console.error('Failed to turn on sensitive data visibility:', error)
    }
  }

  const toggleFreeze = (onSuccess?: () => void) => {
    try {
      console.log(`Freezing card ${card.id}`)

      if (onSuccess) {
        onSuccess()
      }
      setIsFrozen(true)

    } catch (error) {
      console.error('Failed to toggle card freeze status:', error)
    }
  }

  const toggleUnfreeze = (onSuccess?: () => void) => {
    try {
      console.log(`Unfreezing card ${card.id}`)
      
      if (onSuccess) {
        onSuccess()
      }

      setIsFrozen(false)
    } catch (error) {
      console.error('Failed to toggle card unfreeze status:', error)
    }
  }

  return {
    showSensitiveData,
    isFrozen,
    toggleSensitiveDataOff,
    toggleSensitiveDataOn,
    toggleFreeze,
    toggleUnfreeze
  }
}
