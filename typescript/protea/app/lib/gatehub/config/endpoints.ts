// API endpoint configurations
export const GATEHUB_API_ENDPOINTS = {
  // Card endpoints
  cards: {
    list: (customerId: string) => `/cards/v1/customers/${customerId}/cards`,
    details: (cardId: string) => `/cards/v1/cards/${cardId}`,
    transactions: (cardId: string) => `/cards/v1/cards/${cardId}/transactions`,
    lock: (cardId: string) => `/cards/v1/cards/${cardId}/lock`,
    unlock: (cardId: string) => `/cards/v1/cards/${cardId}/unlock`,
    block: (cardId: string) => `/cards/v1/cards/${cardId}/block`,
    close: (cardId: string) => `/cards/v1/cards/${cardId}/card`,
    limits: {
      get: (cardId: string) => `/cards/v1/cards/${cardId}/limits`,
      create: (cardId: string) => `/cards/v1/cards/${cardId}/limits`
    },
    plastic: (cardId: string) => `/cards/v1/cards/${cardId}/plastic`
  },

  // Address endpoints
  addresses: {
    list: '/cards/v2/addresses',
    create: '/cards/v2/addresses',
    delete: (addressId: string) => `/cards/v1/address/${addressId}`
  },

  // Token endpoints (for secure operations)
  tokens: {
    cardData: '/cards/v1/token/card-data',
    pin: '/cards/v1/token/pin',
    pinChange: '/cards/v1/token/pin-change',
    appleProvisioning: '/cards/v1/token/apple-provisioning',
    googleProvisioning: '/cards/v1/token/google-provisioning',
    pai: '/cards/v1/token/pai'
  },

  // Customer endpoints
  customers: {
    create: '/cards/v1/customers/managed',
    details: (customerId: string) => `/cards/v1/customers/${customerId}`,
    addresses: {
      list: (customerId: string) =>
        `/cards/v1/customers/${customerId}/addresses`,
      create: (customerId: string) =>
        `/cards/v1/customers/${customerId}/addresses`
    },
    cards: (customerId: string) => `/cards/v1/customers/${customerId}/cards`
  },

  // Transaction endpoints
  transactions: {
    details: (transactionId: string) =>
      `/cards/v1/transactions/${transactionId}`
  },

  // Admin endpoints (card applications and products)
  admin: {
    applications: {
      list: '/cards/v1/card-applications',
      create: '/cards/v1/card-applications',
      details: (applicationId: string) =>
        `/cards/v1/card-applications/${applicationId}`,
      limits: (applicationId: string) =>
        `/cards/v1/card-applications/${applicationId}/limits`,
      products: (applicationId: string) =>
        `/cards/v1/card-applications/${applicationId}/card-products`
    },
    products: {
      list: '/cards/v1/card-products',
      create: '/cards/v1/card-products',
      details: (productId: string) => `/cards/v1/card-products/${productId}`,
      limits: (productId: string) =>
        `/cards/v1/card-products/${productId}/limits`
    },
    users: {
      customerCards: (userUuid: string) =>
        `/cards/v1/user/${userUuid}/customer-cards`
    }
  },

  // 3DS endpoints
  threeDS: {
    confirmations: '/cards/v1/3ds',
    deviceConfirmation: '/cards/v1/3ds'
  },

  // Clearing endpoints (admin only)
  clearing: {
    upload: '/cards/v1/clearing/file',
    process: '/cards/v1/clearing/process',
    authorizations: '/cards/v1/clearing/authorizations',
    transactions: '/cards/v1/clearing/transactions',
    files: '/cards/v1/clearing/files'
  }
} as const
