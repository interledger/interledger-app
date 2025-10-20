import { useFetcher } from '@remix-run/react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { route } from 'routes-gen'
import {
  CardStatus,
  CardTokenType
} from '~/generated/connect/backend/v1/backend_pb'
import { useCardProcessorApi } from '~/lib/cards/hooks/useCardProcessorApiClient'
import {
  CardProcessorSensitiveDataResponse,
  HttpMethod,
  StorableCard
} from '~/lib/cards/types'
import { decryptWithPrivateKey } from '~/lib/crypto'
import { useKeyGeneration } from '~/lib/useKeyGeneration'
import { Operation, OperationResponse } from '~/routes/api_.cardOperation'
import { GetGatehubTokenResponse } from '~/routes/api_.getGatehubToken'
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

const isCardLocked = (card: StorableCard): boolean => {
  return (
    card.lockLevel !== 'CARD_LOCK_LEVEL_NONE' &&
    card.lockLevel !== 'CARD_LOCK_LEVEL_UNKNOWN'
  )
}

const isCardBlocked = (card: StorableCard): boolean => {
  return (
    card.status === CardStatus.BLOCKED ||
    card.status === CardStatus.TEMPORARY_BLOCKED
  )
}

export const useCardActions = (card: StorableCard) => {
  const fetcher = useFetcher<GetGatehubTokenResponse & OperationResponse>()
  const { actionStatus, executeAction, resetStatus, setActionStatus } =
    useActionExecute()
  const { keyPair } = useKeyGeneration()
  const { cardProcessorClient } = useCardProcessorApi()

  const [isSensitiveDataVisible, setIsSensitiveDataVisible] = useState(false)
  const [isPinVisible, setIsPinVisible] = useState(false)
  const [sensitiveData, setSensitiveData] = useState(
    getDefaultSensitiveData(card)
  )
  const [pin, setPin] = useState('****')

  // Reset when switching cards
  useEffect(() => {
    setIsSensitiveDataVisible(false)
    setSensitiveData(getDefaultSensitiveData(card))
    setIsPinVisible(false)
    setPin('****')
    resetStatus()
  }, [card])

  // Server Action listener
  useEffect(() => {
    if (fetcher.data?.errors || fetcher.data?.success == false) {
      setActionStatus('error')
      return
    }

    if (fetcher.data?.tokenType !== undefined) {
      const { tokenType, token, links } = fetcher.data

      switch (tokenType) {
        case CardTokenType.CARD_DATA:
          onSensitiveDataToken({ token, links })
          break
        case CardTokenType.PIN:
          onGetPinToken({ token, links })
          break
        default:
          setActionStatus('error')
          break
      }
    }
  }, [fetcher.data, setActionStatus])

  /**
   * Action listeners
   */
  const onSensitiveDataToken = async ({ token: jwtToken, links }: any) => {
    executeAction({
      execute: async () => {
        const hrefs = links[0].href
        const method = links[0].method

        const encryptedCardData =
          await cardProcessorClient.cards.getSensitiveData({
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
      }
    })
  }

  const onGetPinToken = async ({ token: jwtToken, links }: any) => {
    executeAction({
      execute: async () => {
        const href = links[0].href
        const method = links[0].method

        const encryptedPinData = await cardProcessorClient.cards.getPin({
          jwtToken,
          cardProcessorUrl: href,
          httpMethod: method as unknown as HttpMethod
        })
        const decryptedPinData = await decryptWithPrivateKey<string>(
          keyPair!.privateKey,
          encryptedPinData.cypher
        )

        setPin(decryptedPinData)
      },
      onSuccess: () => {
        setIsPinVisible(true)
      },
      onError: (error) => {
        setIsPinVisible(true)
        setPin('****')
      }
    })
  }

  /**
   * Toggles
   */
  const triggerTokenOperation = useCallback(
    async (tokenType: CardTokenType) => {
      setActionStatus('loading')

      if (!keyPair?.publicKey) {
        setActionStatus('error')
        return
      }

      // trigger the server action to retrieve the JWT token
      const formData = new FormData()
      formData.append('cardId', card.id)
      formData.append('tokenType', tokenType.toString())
      formData.append('publicKey', keyPair?.publicKey)
      fetcher.submit(formData, {
        method: 'post',
        action: route('/api/getGatehubToken')
      })
    },
    [card.id, keyPair?.publicKey]
  )

  const toggleSensitiveData = useCallback(() => {
    if (!isSensitiveDataVisible) {
      triggerTokenOperation(CardTokenType.CARD_DATA)
    } else {
      executeAction({
        execute: async () => {},
        onSuccess: () => {
          setIsSensitiveDataVisible(false)
          setSensitiveData(getDefaultSensitiveData(card))
        }
      })
    }
  }, [isSensitiveDataVisible, triggerTokenOperation, executeAction])

  const toggleViewPin = useCallback(() => {
    if (!isPinVisible) {
      triggerTokenOperation(CardTokenType.PIN)
    } else {
      executeAction({
        execute: async () => {},
        onSuccess: () => {
          setIsPinVisible(false)
          setPin('****')
        }
      })
    }
  }, [isPinVisible, triggerTokenOperation, executeAction])

  const triggerOperation = useCallback(
    async (operation: Operation) => {
      setActionStatus('loading')

      // trigger the server action to retrieve the JWT token
      const formData = new FormData()
      formData.append('cardId', card.id)
      formData.append('operation', operation)
      fetcher.submit(formData, {
        method: 'post',
        action: route('/api/cardOperation')
      })
    },
    [card.id, setActionStatus, fetcher]
  )

  const toggleLock = useCallback(
    () => triggerOperation('freeze'),
    [triggerOperation]
  )
  const toggleUnlock = useCallback(
    () => triggerOperation('unfreeze'),
    [triggerOperation]
  )
  const toggleBlock = useCallback(
    () => triggerOperation('block'),
    [triggerOperation]
  )
  const toggleTerminate = useCallback(
    () => triggerOperation('terminate'),
    [triggerOperation]
  )
  const isLocked = useMemo(() => isCardLocked(card), [card])
  const isBlocked = useMemo(() => isCardBlocked(card), [card])

  return {
    isSensitiveDataVisible,
    isPinVisible,
    isLocked,
    isBlocked,
    sensitiveData,
    pin,
    actionStatus,
    toggleSensitiveData,
    toggleLock,
    toggleUnlock,
    toggleBlock,
    toggleTerminate,
    toggleViewPin
  }
}
