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
      if (isOpen && initChallengeFetcher.state === 'loading') {
        return (
          <div className='fixed inset-0 z-50 flex items-center justify-center'>
            <div
              className='absolute inset-0 bg-black/50 backdrop-blur-sm'
              aria-hidden='true'
            />
            <div className='relative z-10 rounded-lg bg-white p-8 shadow-xl dark:bg-gray-800'>
              <div className='flex flex-col items-center space-y-4'>
                <div className='h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600' />
                <p className='text-sm text-gray-600 dark:text-gray-300'>
                  Initializing ...
                </p>
              </div>
            </div>
          </div>
        )
      }

      if (!initChallengeFetcher.data || initChallengeFetcher.data.error) {
        return null
      }

      return (
        <TotpChallengePopup
          flowId={initChallengeFetcher.data.flowId}
          csrfToken={initChallengeFetcher.data.csrfToken}
          isOpen={isOpen}
          onClose={closePopup}
          onSuccess={handleSuccess}
          onError={handleError}
        />
      )
    }
  }, [
    initChallengeFetcher.data,
    initChallengeFetcher.state,
    isOpen,
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
