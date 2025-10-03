// API endpoint configurations
export const GATEHUB_API_ENDPOINTS = {
  cards: {
    sensitiveData: () => `/cards/v1/proxy/clientDevice/cardData`,
  },

  tokens: {
    cardData: '/cards/v1/token/card-data',
  },
} as const
