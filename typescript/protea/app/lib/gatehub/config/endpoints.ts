// API endpoint configurations
export const GATEHUB_API_ENDPOINTS = {
  cards: {
    sensitiveData: () => `/cards/v1/proxy/clientDevice/cardData`,
    lockCard: (cardId: string) => `/cards/v1/cards/${cardId}/lock`,
    unlockCard: (cardId: string) => `/cards/v1/cards/${cardId}/unlock`,
  },

  tokens: {
    cardData: '/cards/v1/token/card-data',
  },
} as const
