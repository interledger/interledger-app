// API endpoint configurations
export const GATEHUB_API_ENDPOINTS = {
  cards: {
    sensitiveData: () => `/cards/v1/proxy/clientDevice/cardData`,
    pin: () => `/cards/v1/proxy/clientDevice/pin`,
    lockCard: (cardId: string) => `/cards/v1/cards/${cardId}/lock`,
    unlockCard: (cardId: string) => `/cards/v1/cards/${cardId}/unlock`
  },

  tokens: {
    cardData: '/cards/v1/token/card-data',
    pinShow: '/cards/v1/token/pin',
    pinChange: '/cards/v1/token/pin-change'
  }
} as const
