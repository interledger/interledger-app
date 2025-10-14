import { useFetcher } from '@remix-run/react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { TotpChallengePopup } from '~/components/TotpChallengePopup'

interface TotpChallengeData {
  flowId: string
  csrfToken: string
  error?: string
}

interface UseTotpChallengeOptions {
  onSuccess?: () => void
  onError?: (error: string) => void
}

/**
 * Hook for retrieving TOTP challenge popup
 * Returns:
 * - popTotp: function to trigger the TOTP challenge popup
 * - TotpPopup: component to render the popup
 * - isLoading: boolean indicating if challenge is being initialized
 * - isOpen: boolean indicating if popup is currently open
 */
export function useTotpChallenge(options?: UseTotpChallengeOptions) {
  const [isOpen, setIsOpen] = useState(false)
  const initChallengeFetcher = useFetcher<TotpChallengeData>()

  const popTotp = useCallback(() => {
    initChallengeFetcher.load('/api/totp-challenge-init')
    setIsOpen(true)
  }, [])

  const closePopup = useCallback(() => {
    setIsOpen(false)
  }, [])

  const handleSuccess = useCallback(() => {
    options?.onSuccess?.()
    setIsOpen(false)
  }, [options])

  const handleError = useCallback(
    (error: string) => {
      options?.onError?.(error)
    },
    [options]
  )

  useEffect(() => {
    if (initChallengeFetcher.data?.error) {
      // TODO: log challenge init error
      setIsOpen(false)
      handleError(initChallengeFetcher.data.error)
    }
  }, [initChallengeFetcher.data, handleError])

  const TotpPopup = useMemo(() => {
    return () => {
      if (!initChallengeFetcher.data || initChallengeFetcher.data.error) {
        return null
      }

      return (
        <TotpChallengePopup
          flowId={initChallengeFetcher.data.flowId}
          csrfToken={initChallengeFetcher.data.csrfToken}
          onClose={closePopup}
          onSuccess={handleSuccess}
          onError={handleError}
        />
      )
    }
  }, [
    initChallengeFetcher.data,
    closePopup,
    handleSuccess,
    handleError
  ])

  return {
    popTotp,
    TotpPopup,
    isLoading: initChallengeFetcher.state === 'loading',
    isOpen
  }
}
