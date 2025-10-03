import { useMemo } from 'react'
import { createCardProcessorApiClient } from '~/lib/gatehub/CardProcessorApiClient'
import { useGateHubStore } from '~/lib/gatehub/hooks/useGateHubStore'
import { useKeyGeneration } from '~/lib/useKeyGeneration'
import { HttpMethod } from '../types'

export const useCardProcessorApi = (baseUrl?: string) => {
  const { keyPair } = useKeyGeneration()
  const { card, user, auth } = useGateHubStore()

  const apiClient = useMemo(() => {
    // todo: update config when dependencies change instead of recreating the client
    const client = createCardProcessorApiClient({
      baseUrl,
      appId: auth.appId,
      cardAppId: auth.cardAppId,
      managedUserUuid: user.managedUserUuid,
    })
    return client
  }, [baseUrl, auth.cardAppId, user.managedUserUuid])

  const api = {
    // Card operations
    cards: {
      getSensitiveData: ({
        jwtToken,
        cardProcessorUrl,
        httpMethod
      }: {
        jwtToken: string
        cardProcessorUrl: string
        httpMethod: HttpMethod
      }) => {
        if (!jwtToken) throw new Error('JWT token required')
        return apiClient.getCardSensitiveData({ jwtToken, cardProcessorUrl, httpMethod })
      }
    },

    // Token operations (for secure operations)
    tokens: {
      getCardDataToken: (publicKey?: string) => {
        const { activeCardId } = card
        const keyToUse = publicKey || keyPair?.publicKey
        if (!activeCardId || !keyToUse)
          throw new Error('Card ID and public key required')
        return apiClient.getCardDataToken(activeCardId, keyToUse)
      }
    }
  }

  return {
    gatehubClient: api
  }
}
