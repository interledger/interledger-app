import { useFetcher } from '@remix-run/react'

import { memo, useEffect } from 'react'
import { route } from 'routes-gen'
import type { PendingThreeDSConfirmation } from '~/generated/connect/backend/v1/backend_pb'
import { usePendingConfirmations } from '~/lib/usePendingConfirmations'
import type { GetPendingConfirmationsResponse } from '~/routes/api_.getPendingConfirmations'

export const PendingConfirmationsLoader = memo(
  ({ walletId }: { walletId?: string }) => {
    const confirmationsFetcher = useFetcher<GetPendingConfirmationsResponse>()

    const { initializePendingConfirmations, clearTimeouts, hasFetched } =
      usePendingConfirmations()

    // Fetch confirmations on mount when walletId is defined
    useEffect(() => {
      if (walletId) {
        confirmationsFetcher.load(route('/api/getPendingConfirmations'))
      } else {
        clearTimeouts()
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [walletId])

    // Initialize store when data is fetched
    useEffect(() => {
      // No need to reset the confirmations if they are fetched
      if (confirmationsFetcher.data?.confirmations && !hasFetched) {
        initializePendingConfirmations(
          confirmationsFetcher.data
            ?.confirmations as PendingThreeDSConfirmation[] // I hate JsonifyObject...
        )
      }
    }, [
      confirmationsFetcher.data?.confirmations,
      initializePendingConfirmations,
      hasFetched
    ])

    return null
  }
)

PendingConfirmationsLoader.displayName = 'PendingConfirmationsLoader'
