import { useEffect, useState } from 'react'
import { CardStatus } from '~/generated/connect/backend/v1/backend_pb'
import { decryptWithPrivateKey } from '~/lib/crypto'
import { useCardsStore } from '~/lib/gatehub/hooks/gatehub.stores'
import { useCardProcessorApi } from '~/lib/gatehub/hooks/useCardProcessorApiClient'
import {
  CardProcessorPinResponse,
  CardProcessorSensitiveDataResponse,
  StorableCard
} from '~/lib/gatehub/types'
import { useKeyGeneration } from '~/lib/useKeyGeneration'
import { ActionStatus, useActionExecute } from './useActionExecute'

interface UseCardActionsReturn {
  showSensitiveData: boolean
  showPin: boolean
  isLocked: boolean
  isBlocked: boolean
  sensitiveData: CardProcessorSensitiveDataResponse
  pin: string
  actionStatus: ActionStatus
  toggleSensitiveDataOff: (onSuccess?: () => void) => void
  toggleSensitiveDataOn: (onSuccess?: () => void) => void
  toggleLock: (onSuccess?: () => void) => void
  toggleUnlock: (onSuccess?: () => void) => void
  toggleViewPin: () => void
}

const getDefaultSensitiveData = (
  card: StorableCard
): CardProcessorSensitiveDataResponse => {
  return {
    Pan: card.maskedPan,
    ExpiryDate: card.expiryDate,
    Cvc2: '***'
  }
}

const isLocked = (card: StorableCard): boolean => {
  return (
    card.lockLevel !== 'CARD_LOCK_LEVEL_NONE' &&
    card.lockLevel !== 'CARD_LOCK_LEVEL_UNKNOWN'
  )
}

const isBlocked = (card: StorableCard): boolean => {
  return (
    card.status === CardStatus.BLOCKED ||
    card.status === CardStatus.TEMPORARY_BLOCKED
  )
}

export const useCardActions = (card: StorableCard): UseCardActionsReturn => {
  const { updateActiveCard } = useCardsStore()
  const { actionStatus, executeAction, resetStatus } = useActionExecute()
  const { keyPair } = useKeyGeneration()
  const { cardProcessorClient } = useCardProcessorApi()

  const [showSensitiveData, setShowSensitiveData] = useState(false)
  const [showPin, setShowPin] = useState(false)
  const [isLockedState, setIsLockedState] = useState(isLocked(card))
  const [isBlockedState, setIsBlockedState] = useState(isBlocked(card))
  const [sensitiveData, setSensitiveData] = useState(
    getDefaultSensitiveData(card)
  )
  const [pin, setPin] = useState('****')

  // Reset when switching cards
  useEffect(() => {
    setShowSensitiveData(false)
    setShowPin(false)
    setIsLockedState(isLocked(card))
    setIsBlockedState(isBlocked(card))
    setSensitiveData(getDefaultSensitiveData(card))
    setPin('****')
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
        setIsLockedState(true)
        updateActiveCard({ ...card, lockLevel: 'CARD_LOCK_LEVEL_CLIENT' })
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
        setIsLockedState(false)
        updateActiveCard({ ...card, lockLevel: 'CARD_LOCK_LEVEL_NONE' })
      },
      onError: (error) => {
        console.error('Failed to unfreeze card:', error)
      }
    })
  }

  const toggleViewPin = async () => {
    if (showPin) {
      // Hide PIN
      executeAction({
        execute: async () => {
          await new Promise((resolve) => setTimeout(resolve, 500))
        },
        onSuccess: () => {
          setShowPin(false)
          setPin('****')
        },
        onError: (error) => {
          console.error('Failed to hide PIN:', error)
        }
      })
    } else {
      // Show PIN
      executeAction({
        execute: async () => {
          console.log('🔑 Getting PIN change token')

          // 1. Request JWT token for PIN change
          const jwtTokenResponse =
            await cardProcessorClient.tokens.getPinShowToken()
          const jwtToken = jwtTokenResponse.data.token
          const href = jwtTokenResponse.data.links[0].href
          const method = jwtTokenResponse.data.links[0].method
          console.log('🔑 PIN JWT token', jwtToken)
          console.log('🔑 Method', method)
          console.log('🔑 Rel', jwtTokenResponse.data.links[0].rel)
          console.log('🔑 Href', href)

          // 2. Request encrypted PIN with JWT token
          const encryptedPinData = await cardProcessorClient.cards.getPin({
            jwtToken,
            cardProcessorUrl: href,
            httpMethod: method
          })
          console.log('🔑 Encrypted PIN data', encryptedPinData)

          // 3. Decrypt PIN using private key
          console.log(
            '🔑 Decrypting PIN using private key',
            keyPair?.privateKey
          )
          const decryptedPinData =
            await decryptWithPrivateKey<CardProcessorPinResponse>(
              keyPair!.privateKey,
              encryptedPinData.data.cypher
            )
          console.log('🔑 Decrypted PIN data', decryptedPinData)

          setPin(decryptedPinData.pin)
        },
        onSuccess: () => {
          setShowPin(true)
          console.log('✅ toggleViewPin success')
        },
        onError: (error) => {
          console.error('❌ Failed to view PIN:', error)
          setShowPin(true)
          setPin('****')
        }
      })
    }
  }

  return {
    showSensitiveData,
    showPin,
    isLocked: isLockedState,
    isBlocked: isBlockedState,
    sensitiveData,
    pin,
    actionStatus,
    toggleSensitiveDataOff,
    toggleSensitiveDataOn,
    toggleLock,
    toggleUnlock,
    toggleViewPin
  }
}
