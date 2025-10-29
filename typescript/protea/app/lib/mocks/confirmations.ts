/**
 * 3DS Pending Confirmation types and mock data
 * Based on GateHub Cards API: https://api.gatehub.net/cards/openapi#tag/3ds
 */

import type { PendingConfirmation } from '../usePendingConfirmations'

/**
 * 3DS Pending Confirmation structure from GateHub
 */

/**
 * Mock pending 3DS confirmations for testing and development
 */
export const mockPendingConfirmations: PendingConfirmation[] = [
  {
    transactionId: '62469058-d962-4f7c-a9c0-8c2a1b6efaa3',
    merchantName: 'Shop One',
    purchaseAmount: '99.98',
    purchaseCurrency: 'EUR',
    purchaseDate: `${Date.now() - 60000}`, // 1 minute ago
    timeout: '300'
  },
  {
    transactionId: 'a7f82c14-b3e5-4a1d-9f8b-3c4d5e6f7a8b',
    merchantName: 'Shop Two',
    purchaseAmount: '229.98',
    purchaseCurrency: 'EUR',
    purchaseDate: `${Date.now() - 120000}`, // 2 minutes ago
    timeout: '300'
  },
  {
    transactionId: '9b3e4c5d-6a7f-4e8b-a9c0-1d2e3f4a5b6c',
    merchantName: 'Shop Three',
    purchaseAmount: '30000',
    purchaseCurrency: 'EUR',
    purchaseDate: `${Date.now() - 180000}`, // 3 minutes ago
    timeout: '3000'
  }
]
