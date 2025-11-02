import { useFetcher } from '@remix-run/react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { route } from 'routes-gen'
import {
  CardLockLevel,
  CardStatus,
  CardTokenType
} from '~/generated/connect/backend/v1/backend_pb'
import { cardProcessorClient } from '~/lib/cards/CardProcessorApiClient'
import type {
  CardProcessorSensitiveDataResponse,
  HttpMethod,
  StorableCard
} from '~/lib/cards/types'
import { useKeyGeneration } from '~/lib/cards/useKeyGeneration'
import { usePinChangePopup } from '~/lib/cards/usePinChangePopup'
import { decryptWithPrivateKey, encryptWithToken } from '~/lib/crypto'
import { useTotpChallenge } from '~/lib/useTotpChallenge'
import type { Operation, OperationResponse } from '~/routes/api_.cardOperation'
import type { GetCardTokenResponse } from '~/routes/api_.getCardToken'
import { useActionExecute } from './useActionExecute'

const getDefaultSensitiveData = (
  card: StorableCard
): CardProcessorSensitiveDataResponse => {
  return {
    Pan: card.maskedPan,
    ExpiryDate: card.expiryDate,
    Cvc2: '***'
  }
}

const isBlockedByAdmin = (card: StorableCard): boolean => {
  return card.lockLevel === CardLockLevel.ADMIN
}

const isCardFrozen = (card: StorableCard): boolean => {
  return card.lockLevel === CardLockLevel.CLIENT
}

const isCardBlocked = (card: StorableCard): boolean => {
  return (
    card.status === CardStatus.BLOCKED ||
    card.status === CardStatus.TEMPORARY_BLOCKED
  )
}

export const useCardActions = (card: StorableCard) => {
  const fetcher = useFetcher<GetCardTokenResponse & OperationResponse>()
  const {
    actionStatus,
    executeAction,
    resetStatus,
    setActionStatus,
    errorStatus,
    successStatus
  } = useActionExecute()
  const { keyPair } = useKeyGeneration()
  const { withTotpChallenge } = useTotpChallenge()
  const { withPinChangePopup, PinChangePopup } = usePinChangePopup()

  const [isSensitiveDataVisible, setIsSensitiveDataVisible] = useState(false)
  const [sensitiveData, setSensitiveData] = useState(
    getDefaultSensitiveData(card)
  )
  const [isPinVisible, setIsPinVisible] = useState(false)
  const [pinDisplayed, setPinDisplayed] = useState('****')

  const [showBack, setShowBack] = useState<boolean>(false)
  const newPinRef = useRef<string | null>(null)
  const timeoutRef = useRef<number | null>(null)

  const resetSensitiveData = useCallback(() => {
    setIsSensitiveDataVisible(false)
    setSensitiveData(getDefaultSensitiveData(card))
  }, [card, setSensitiveData, setIsSensitiveDataVisible])

  // Reset when switching cards
  useEffect(() => {
    resetSensitiveData()
    setIsPinVisible(false)
    setPinDisplayed('****')
    setShowBack(false)
  }, [
    card,
    setPinDisplayed,
    setIsSensitiveDataVisible,
    setIsPinVisible,
    setSensitiveData,
    resetSensitiveData
  ])

  useEffect(() => {
    // Reset only on id change, because operations will trigger a loader reset
    resetStatus()
  }, [card.id, resetStatus])

  /**
   * Token listeners
   */
  const onSensitiveDataToken = useCallback(
    async ({ token: jwtToken, links }: any) => {
      executeAction({
        execute: async () => {
          const hrefs = links[0].href
          const method = links[0].method

          const encryptedCardData =
            await cardProcessorClient.card.getSensitiveData({
              jwtToken,
              cardProcessorUrl: hrefs,
              httpMethod: method
            })

          const decryptedCardData =
            await decryptWithPrivateKey<CardProcessorSensitiveDataResponse>(
              keyPair!.privateKey,
              encryptedCardData.cypher
            )

          setSensitiveData(decryptedCardData)
        },
        onSuccess: () => {
          setIsSensitiveDataVisible(true)
          setShowBack(true)
        }
      })
    },
    [executeAction, keyPair]
  )

  const onGetPinToken = useCallback(
    async ({ token: jwtToken, links }: any) => {
      executeAction({
        execute: async () => {
          const href = links[0].href
          const method = links[0].method

          const encryptedPinData = await cardProcessorClient.card.getPin({
            jwtToken,
            cardProcessorUrl: href,
            httpMethod: method as unknown as HttpMethod
          })
          const decryptedPinData = await decryptWithPrivateKey<string>(
            keyPair!.privateKey,
            encryptedPinData.cypher
          )

          setPinDisplayed(decryptedPinData)
        },
        onSuccess: () => {
          setIsPinVisible(true)
        }
      })
    },
    [executeAction, keyPair]
  )

  const onChangePinToken = useCallback(
    async ({ token: jwtToken, links }: any) => {
      executeAction({
        execute: async () => {
          const newPin = newPinRef.current
          if (!newPin) {
            throw new Error('New PIN is required')
          }

          const href = links[0].href
          const method = links[0].method

          const newPinCypher = await encryptWithToken(jwtToken, newPin)

          const response = await cardProcessorClient.card.changePin({
            newPinCypher,
            jwtToken,
            cardProcessorUrl: href,
            httpMethod: method as unknown as HttpMethod
          })

          if (!response.ok) {
            throw new Error('Failed to change PIN')
          }
        }
      })
    },
    [executeAction]
  )

  /**
   * Server Action listener
   */
  useEffect(() => {
    if (fetcher.data?.errors || fetcher.data?.success == false) {
      errorStatus()
      return
    }

    if (fetcher.data?.success) {
      successStatus()
      return
    }

    // Make direct calls to the card processor after token retrieval
    if (fetcher.data?.tokenType !== undefined) {
      const { tokenType, token, links } = fetcher.data

      switch (tokenType) {
        case CardTokenType.CARD_DATA:
          onSensitiveDataToken({ token, links })
          break
        case CardTokenType.PIN:
          onGetPinToken({ token, links })
          break
        case CardTokenType.PIN_CHANGE:
          onChangePinToken({ token, links })
          break
        default:
          errorStatus()
          break
      }
    }
  }, [
    errorStatus,
    fetcher.data,
    onChangePinToken,
    onGetPinToken,
    onSensitiveDataToken,
    setActionStatus,
    successStatus
  ])

  /**
   * Toggles
   */
  const triggerTokenOperation = useCallback(
    (tokenType: CardTokenType) => {
      withTotpChallenge(() => {
        setActionStatus('loading')

        if (!keyPair?.publicKey) {
          errorStatus()
          return
        }

        const formData = new FormData()
        formData.append('cardId', card.id)
        formData.append('tokenType', tokenType.toString())
        formData.append('publicKey', keyPair?.publicKey)
        fetcher.submit(formData, {
          method: 'post',
          action: route('/api/getCardToken')
        })
      })
    },
    [
      card.id,
      errorStatus,
      fetcher,
      keyPair?.publicKey,
      setActionStatus,
      withTotpChallenge
    ]
  )

  const toggleSensitiveDataOn = useCallback(() => {
    triggerTokenOperation(CardTokenType.CARD_DATA)
  }, [triggerTokenOperation])

  const toggleViewPin = useCallback(() => {
    if (!isPinVisible) {
      triggerTokenOperation(CardTokenType.PIN)
    } else {
      executeAction({
        execute: async () => {
          setIsPinVisible(false)
          setPinDisplayed('****')
        }
      })
    }
  }, [isPinVisible, triggerTokenOperation, executeAction])

  const toggleChangePin = useCallback(() => {
    withPinChangePopup((newPin: string) => {
      newPinRef.current = newPin
      triggerTokenOperation(CardTokenType.PIN_CHANGE)
    })
  }, [withPinChangePopup, triggerTokenOperation])

  const triggerOperation = useCallback(
    (operation: Operation) => {
      withTotpChallenge(() => {
        setActionStatus('loading')

        const formData = new FormData()
        formData.append('cardId', card.id)
        formData.append('operation', operation)
        fetcher.submit(formData, {
          method: 'post',
          action: route('/api/cardOperation')
        })
      })
    },
    [card.id, setActionStatus, fetcher, withTotpChallenge]
  )
  const toggleFreeze = useCallback(
    () => triggerOperation('freeze'),
    [triggerOperation]
  )
  const toggleUnfreeze = useCallback(
    () => triggerOperation('unfreeze'),
    [triggerOperation]
  )
  const toggleBlock = useCallback(
    () => triggerOperation('block'),
    [triggerOperation]
  )

  const flip = () => {
    if (showBack && isSensitiveDataVisible) {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }

      // Using window just for the types it collides with the NodeJS ones.
      timeoutRef.current = window.setTimeout(() => {
        // Turning to the front resets sensitive data
        resetSensitiveData()
        timeoutRef.current = null
      }, 350)
    }
    setShowBack(!showBack)
  }

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const isFrozen = useMemo(() => isCardFrozen(card), [card])
  const isBlocked = useMemo(() => isCardBlocked(card), [card])
  const isAdminBlocked = useMemo(() => isBlockedByAdmin(card), [card])

  return {
    flip,
    showBack,
    setShowBack,
    isSensitiveDataVisible,
    isPinVisible,
    isFrozen,
    isBlocked,
    isBlockedByAdmin: isAdminBlocked,
    sensitiveData,
    pin: pinDisplayed,
    actionStatus,
    toggleSensitiveDataOn,
    toggleFreeze,
    toggleUnfreeze,
    toggleBlock,
    toggleViewPin,
    toggleChangePin,
    PinChangePopup
  }
}
