import { useEffect, useState } from 'react'
import {
  CardLockLevel,
  CardStatus
} from '~/generated/connect/backend/v1/backend_pb'
import { ActionStatus, useActionExecute } from './useActionExecute'

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

interface UseCardActionsReturn {
  showSensitiveData: boolean
  isFrozen: boolean
  isBlocked: boolean
  sensitiveData: SensitiveData
  actionStatus: ActionStatus
  toggleSensitiveDataOff: (onSuccess?: () => void) => void
  toggleSensitiveDataOn: (onSuccess?: () => void) => void
  toggleFreeze: (onSuccess?: () => void) => void
  toggleUnfreeze: (onSuccess?: () => void) => void
}

const getDefaultSensitiveData = (card: Card): SensitiveData => {
  return {
    cardNumber: card.maskedPan,
    expiryDate: null,
    cvv: null
  }
}

const isFrozen = (card: Card): boolean => {
  return card.lockLevel === CardLockLevel.ADMIN
}

const isBlocked = (card: Card): boolean => {
  return (
    card.status === CardStatus.BLOCKED ||
    card.status === CardStatus.TEMPORARY_BLOCKED
  )
}

export const useCardActions = (card: Card): UseCardActionsReturn => {
  const [showSensitiveData, setShowSensitiveData] = useState(false)
  const [isFrozenState, setIsFrozenState] = useState(isFrozen(card))
  const [isBlockedState, setIsBlockedState] = useState(isBlocked(card))
  const { actionStatus, executeAction, resetStatus } = useActionExecute()

  const [sensitiveData, setSensitiveData] = useState(
    getDefaultSensitiveData(card)
  )

  // Reset when switching cards
  useEffect(() => {
    setShowSensitiveData(false)
    setIsFrozenState(isFrozen(card))
    setIsBlockedState(isBlocked(card))
    setSensitiveData(getDefaultSensitiveData(card))
    resetStatus()
  }, [card])

  const toggleSensitiveDataOff = async (onSuccess?: () => void) =>
    executeAction({
      execute: async () => {
        // Simulate API call delay
        await new Promise((resolve) => setTimeout(resolve, 500))

        setShowSensitiveData(false)
        setSensitiveData(getDefaultSensitiveData(card))
      },
      onSuccess,
      onError: (error) => {
        console.error('Failed to turn off sensitive data visibility:', error)
      }
    })

  const toggleSensitiveDataOn = async (onSuccess?: () => void) => {
    executeAction({
      execute: async () => {
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
      },
      onSuccess,
      onError: (error) => {
        console.error('Failed to turn on sensitive data visibility:', error)
      }
    })
  }

  const toggleFreeze = async (onSuccess?: () => void) => {
    executeAction({
      execute: async () => {
        // Simulate API call delay
        await new Promise((resolve) => setTimeout(resolve, 800))

        setIsFrozenState(true)
      },
      onSuccess,
      onError: (error) => {
        console.error('Failed to freeze card:', error)
      }
    })
  }

  const toggleUnfreeze = async (onSuccess?: () => void) => {
    executeAction({
      execute: async () => {
        // Simulate API call delay
        await new Promise((resolve) => setTimeout(resolve, 800))

        setIsFrozenState(false)
      },
      onSuccess,
      onError: (error) => {
        console.error('Failed to unfreeze card:', error)
      }
    })
  }

  return {
    showSensitiveData,
    isFrozen: isFrozenState,
    isBlocked: isBlockedState,
    sensitiveData,
    actionStatus,
    toggleSensitiveDataOff,
    toggleSensitiveDataOn,
    toggleFreeze,
    toggleUnfreeze
  }
}
