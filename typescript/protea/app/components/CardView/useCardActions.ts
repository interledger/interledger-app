import { useEffect, useState } from 'react'
import { CardStatus } from '~/generated/connect/backend/v1/backend_pb'
import { decryptWithPrivateKey } from '~/lib/crypto'
import { useCardProcessorApi } from '~/lib/gatehub/hooks/useCardProcessorApiClient'
import {
  StorableCard as Card,
  CardProcessorSensitiveDataResponse
} from '~/lib/gatehub/types'
import { useKeyGeneration } from '~/lib/useKeyGeneration'
import { ActionStatus, useActionExecute } from './useActionExecute'

// export interface Card {
//   id: string
//   nameOnCard: string
//   maskedPan: string
//   expiryDate: string
//   status: number
//   lockLevel: number
// }

interface UseCardActionsReturn {
  showSensitiveData: boolean
  isLocked: boolean
  isBlocked: boolean
  sensitiveData: CardProcessorSensitiveDataResponse
  actionStatus: ActionStatus
  toggleSensitiveDataOff: (onSuccess?: () => void) => void
  toggleSensitiveDataOn: (onSuccess?: () => void) => void
  toggleLock: (onSuccess?: () => void) => void
  toggleUnlock: (onSuccess?: () => void) => void
  toggleViewPin: () => void
}

const getDefaultSensitiveData = (
  card: Card
): CardProcessorSensitiveDataResponse => {
  return {
    Pan: card.maskedPan,
    ExpiryDate: card.expiryDate,
    Cvc2: '***'
  }
}

const isLocked = (card: Card): boolean => {
  return card.lockLevel !== 'CARD_LOCK_LEVEL_UNKNOWN'
}

const isBlocked = (card: Card): boolean => {
  return (
    card.status === CardStatus.BLOCKED ||
    card.status === CardStatus.TEMPORARY_BLOCKED
  )
}

export const useCardActions = (card: Card): UseCardActionsReturn => {
  const [showSensitiveData, setShowSensitiveData] = useState(false)
  const [isFrozenState, setIsFrozenState] = useState(isLocked(card))
  const [isBlockedState, setIsBlockedState] = useState(isBlocked(card))
  const { actionStatus, executeAction, resetStatus } = useActionExecute()
  const [sensitiveData, setSensitiveData] = useState(
    getDefaultSensitiveData(card)
  )
  const { keyPair } = useKeyGeneration()
  const { cardProcessorClient } = useCardProcessorApi()

  // Reset when switching cards
  useEffect(() => {
    setShowSensitiveData(false)
    setIsFrozenState(isLocked(card))
    setIsBlockedState(isBlocked(card))
    setSensitiveData(getDefaultSensitiveData(card))
    resetStatus()
  }, [card])

  const toggleSensitiveDataOff = async (onSuccess?: () => void) =>
    executeAction({
      execute: async () => {
        // Simulate API call delay
        await new Promise((resolve) => setTimeout(resolve, 500))
      },
      onSuccess: () => {
        onSuccess?.()
        setShowSensitiveData(false)
        setSensitiveData(getDefaultSensitiveData(card))
      },
      onError: (error) => {
        console.error('Failed to turn off sensitive data visibility:', error)
      }
    })

  const toggleSensitiveDataOn = async (onSuccess?: () => void) => {
    executeAction({
      execute: async () => {
        // 0. THE JWT RETRIEVAL SHOULD HAPPEN ON THE BACKEND

        // 1 Get generated keys
        console.log('🔑 Getting generated keys')
        console.log(keyPair)

        // 2 Request JWT token
        const jwtTokenResponse =
          await cardProcessorClient.tokens.getCardDataToken(keyPair?.publicKey)
        const jwtToken = jwtTokenResponse.data.token
        const hrefs = jwtTokenResponse.data.links[0].href
        const method = jwtTokenResponse.data.links[0].method
        console.log('🔑 JWT token', jwtToken)
        console.log('🔑 Method', method)
        console.log('🔑 Href', hrefs)

        // 3 Request card data with JWT token
        const encryptedCardData =
          await cardProcessorClient.cards.getSensitiveData({
            jwtToken,
            // cardProcessorUrl: GATEHUB_API_ENDPOINTS.cards.sensitiveData(),
            cardProcessorUrl: hrefs,
            httpMethod: method
          })
        console.log('🔑 Encrypted card data', encryptedCardData)

        // 4 Decrypt card data using public key
        console.log(
          '🔑 Decrypting card data using private key',
          keyPair?.privateKey
        )
        const decryptedCardData =
          await decryptWithPrivateKey<CardProcessorSensitiveDataResponse>(
            keyPair!.privateKey,
            encryptedCardData.data.cypher
          )
        console.log('🔑 Decrypted card data', decryptedCardData)

        setSensitiveData(decryptedCardData)
      },
      onSuccess: () => {
        onSuccess?.()
        setShowSensitiveData(true)
        console.log('✅ toggleSensitiveDataOn success')
      },
      onError: (error) => {
        console.error('❌ Failed to turn on sensitive data visibility:', error)
      }
    })
  }

  const toggleLock = async (onSuccess?: () => void) => {
    executeAction({
      execute: async () => {
        await cardProcessorClient.cards.lockCard(card.id)
      },
      onSuccess: () => {
        onSuccess?.()
        setIsFrozenState(true)
      },
      onError: (error) => {
        console.error('Failed to freeze card:', error)
      }
    })
  }

  const toggleUnlock = async (onSuccess?: () => void) => {
    executeAction({
      execute: async () => {
        await cardProcessorClient.cards.unlockCard(card.id)
      },
      onSuccess: () => {
        onSuccess?.()
        setIsFrozenState(false)
      },
      onError: (error) => {
        console.error('Failed to unfreeze card:', error)
      }
    })
  }

  return {
    showSensitiveData,
    isLocked: isFrozenState,
    isBlocked: isBlockedState,
    sensitiveData,
    actionStatus,
    toggleSensitiveDataOff,
    toggleSensitiveDataOn,
    toggleLock,
    toggleUnlock,
    toggleViewPin: () => {
      console.log('🔑 Viewing PIN')
    }
  }
}
