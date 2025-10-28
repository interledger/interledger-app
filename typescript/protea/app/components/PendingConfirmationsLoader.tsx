import { useFetcher } from '@remix-run/react'
import React, { useEffect } from 'react'
import { usePendingConfirmations } from '~/lib/usePendingConfirmations'
import type { GetPendingConfirmationsResponse } from '~/routes/api_.getPendingConfimations'

const PendingConfirmationsLoader = ({ walletId }: { walletId?: string }) => {
  const confirmationsFetcher = useFetcher<GetPendingConfirmationsResponse>()
  const { initializePendingConfirmations, clearTimeouts } = usePendingConfirmations()

  // Fetch confirmations on mount when walletId is defined
  useEffect(() => {
    if (walletId) {
      confirmationsFetcher.load('/api/getPendingConfimations')
      console.log('[PendingConfirmationsLoader] 🐳 started fetching')
    } else {
      clearTimeouts()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [walletId])

  // Initialize store when data is fetched
  useEffect(() => {
    if (confirmationsFetcher.data?.confirmations) {
      console.log('[PendingConfirmationsLoader] 🐳 got confirmations !')
      initializePendingConfirmations(confirmationsFetcher.data.confirmations)
    }
  }, [confirmationsFetcher.data, initializePendingConfirmations])

  return null
}

export default React.memo(PendingConfirmationsLoader)
