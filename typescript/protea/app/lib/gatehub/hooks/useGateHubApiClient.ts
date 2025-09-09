import { useMemo } from 'react'
import { createGateHubApiClient } from '~/lib/gatehub/GateHubApiClient'
import { useGateHubStore } from '~/lib/gatehub/hooks/useGateHubStore'
import { useKeyGeneration } from '~/lib/useKeyGeneration'

// Hook for making GateHub API requests using Zustand store
export const useGateHubApi = (baseUrl?: string) => {
  const { keyPair } = useKeyGeneration()
  const { card, user, auth, token } = useGateHubStore()

  const apiClient = useMemo(() => {
    const client = createGateHubApiClient({
      baseUrl: baseUrl || '',
      cardAppId: auth.cardAppId,
      managedUserUuid: user.managedUserUuid,
      sessionToken: token || undefined
    })
    return client
  }, [baseUrl, auth.cardAppId, user.managedUserUuid, token])

  const api = {
    // Card operations
    cards: {
      list: () => {
        const { customerId } = user
        if (!customerId)
          throw new Error('Customer ID required for listing cards')
        return apiClient.listCards(customerId)
      },

      getDetails: (cardId?: string) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.getCardDetails(cardIdToUse)
      },

      getTransactions: (
        cardId?: string,
        pageSize?: number,
        pageNumber?: number
      ) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.getCardTransactions(cardIdToUse, pageSize, pageNumber)
      },

      lock: (cardId?: string, reasonCode?: string, note?: string) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.lockCard(cardIdToUse, reasonCode, note)
      },

      unlock: (cardId?: string, note?: string) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.unlockCard(cardIdToUse, note)
      },

      close: (cardId?: string, reasonCode?: string) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.closeCard(cardIdToUse, reasonCode)
      },

      getLimits: (cardId?: string) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.getCardLimits(cardIdToUse)
      },

      setLimits: (cardId?: string, limits?: any) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.setCardLimits(cardIdToUse, limits)
      }
    },

    // Address operations
    addresses: {
      list: () => apiClient.listAddresses(),
      create: (addressData: any) => apiClient.createAddress(addressData),
      delete: (addressId: string) => apiClient.deleteAddress(addressId)
    },

    // Token operations (for secure operations)
    tokens: {
      getCardDataToken: (publicKey?: string) => {
        const { activeCardId } = card
        const keyToUse = publicKey || keyPair?.publicKey
        if (!activeCardId || !keyToUse)
          throw new Error('Card ID and public key required')
        return apiClient.getCardDataToken(activeCardId, keyToUse)
      },

      getPinToken: (publicKey?: string) => {
        const { activeCardId } = card
        const keyToUse = publicKey || keyPair?.publicKey
        if (!activeCardId || !keyToUse)
          throw new Error('Card ID and public key required')
        return apiClient.getPinToken(activeCardId, keyToUse)
      },

      getPinChangeToken: (cardId?: string) => {
        const cardIdToUse = cardId || card.activeCardId
        if (!cardIdToUse) throw new Error('Card ID required')
        return apiClient.getPinChangeToken(cardIdToUse)
      },

      getAppleProvisioningToken: (publicKey?: string) => {
        const { activeCardId } = card
        const keyToUse = publicKey || keyPair?.publicKey
        if (!activeCardId || !keyToUse)
          throw new Error('Card ID and public key required')
        return apiClient.getAppleProvisioningToken(activeCardId, keyToUse)
      },

      getGoogleProvisioningToken: (publicKey?: string) => {
        const { activeCardId } = card
        const keyToUse = publicKey || keyPair?.publicKey
        if (!activeCardId || !keyToUse)
          throw new Error('Card ID and public key required')
        return apiClient.getGoogleProvisioningToken(activeCardId, keyToUse)
      }
    }
  }

  return {
    gatehubClient: api
  }
}
