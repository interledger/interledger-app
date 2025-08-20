import { useEffect, useState } from 'react'

interface Card {
  id: string
  nameOnCard: string
  maskedPan: string
  expiryDate: string
  status: number
  lockLevel: number
}

interface SensitiveData {
  cardNumber: string
  expiryDate: string | null
  cvv: string | null
}

type ActionStatus = 'idle' | 'loading' | 'success' | 'error'

interface UseCardActionsReturn {
  showSensitiveData: boolean
  isFrozen: boolean
  sensitiveData: SensitiveData
  actionStatus: ActionStatus
  toggleSensitiveDataOff: (onSuccess?: () => void) => void
  toggleSensitiveDataOn: (onSuccess?: () => void) => void
  toggleFreeze: (onSuccess?: () => void) => void
  toggleUnfreeze: (onSuccess?: () => void) => void
}

const DELAY_AFTER_ACTION = 3000

const getDefaultSensitiveData = (card: Card): SensitiveData => {
  return {
    cardNumber: card.maskedPan,
    expiryDate: null,
    cvv: null
  }
}

export const useCardActions = (card: Card): UseCardActionsReturn => {
  const [showSensitiveData, setShowSensitiveData] = useState(false)
  const [isFrozen, setIsFrozen] = useState(false)
  const [actionStatus, setActionStatus] = useState<ActionStatus>('idle')

  const [sensitiveData, setSensitiveData] = useState(
    getDefaultSensitiveData(card)
  )

  // Reset state when card changes
  useEffect(() => {
    setShowSensitiveData(false)
    setIsFrozen(false)
    setActionStatus('idle')
    setSensitiveData(getDefaultSensitiveData(card))
  }, [card])

  const resetStatus = (delay: number) => {
    setTimeout(() => setActionStatus('idle'), delay)
  }

  const executeAction = async (
    execute: () => void,
    onSuccess?: () => void,
    onError?: (error: any) => void
  ) => {
    try {
      setActionStatus('loading')

      execute()
      onSuccess?.()

      setActionStatus('success')
      resetStatus(DELAY_AFTER_ACTION)
    } catch (error) {
      onError?.(error)

      setActionStatus('error')
      resetStatus(DELAY_AFTER_ACTION)
    }
  }

  const toggleSensitiveDataOff = async (onSuccess?: () => void) => {
    try {
      setActionStatus('loading')

      // Simulate API call delay
      await new Promise((resolve) => setTimeout(resolve, 500))

      setShowSensitiveData(false)
      setSensitiveData(getDefaultSensitiveData(card))
      setActionStatus('success')

      if (onSuccess) {
        onSuccess()
      }

      resetStatus(DELAY_AFTER_ACTION)
    } catch (error) {
      console.error('Failed to turn off sensitive data visibility:', error)
      setActionStatus('error')

      resetStatus(DELAY_AFTER_ACTION)
    }
  }

  const toggleSensitiveDataOn = async (onSuccess?: () => void) => {
    try {
      setActionStatus('loading')

      // Mock fetching sensitive data for the given card
      await new Promise((resolve) => setTimeout(resolve, 1000))
      const fullCardNumber = card.maskedPan
        .replace('****', '5287')
        .replace('****', '0012')
        .replace('****', '3456')
      setSensitiveData((prev) => ({
        ...prev,
        cardNumber: fullCardNumber,
        expiryDate: card.expiryDate,
        cvv: '123'
      }))

      setShowSensitiveData(true)
      setActionStatus('success')

      if (onSuccess) {
        onSuccess()
      }

      resetStatus(DELAY_AFTER_ACTION)
    } catch (error) {
      console.error('Failed to turn on sensitive data visibility:', error)
      setActionStatus('error')

      resetStatus(DELAY_AFTER_ACTION)
    }
  }

  const toggleFreeze = async (onSuccess?: () => void) => {
    try {
      setActionStatus('loading')

      // Simulate API call delay
      await new Promise((resolve) => setTimeout(resolve, 800))

      setIsFrozen(true)
      setActionStatus('success')

      if (onSuccess) {
        onSuccess()
      }

      resetStatus(DELAY_AFTER_ACTION)
    } catch (error) {
      console.error('Failed to freeze card:', error)
      setActionStatus('error')

      resetStatus(DELAY_AFTER_ACTION)
    }
  }

  const toggleUnfreeze = async (onSuccess?: () => void) => {
    try {
      setActionStatus('loading')

      // Simulate API call delay
      await new Promise((resolve) => setTimeout(resolve, 800))

      setIsFrozen(false)
      setActionStatus('success')

      if (onSuccess) {
        onSuccess()
      }

      resetStatus(DELAY_AFTER_ACTION)
    } catch (error) {
      console.error('Failed to unfreeze card:', error)
      setActionStatus('error')

      resetStatus(DELAY_AFTER_ACTION)
    }
  }

  return {
    showSensitiveData,
    isFrozen,
    sensitiveData,
    actionStatus,
    toggleSensitiveDataOff,
    toggleSensitiveDataOn,
    toggleFreeze,
    toggleUnfreeze
  }
}
