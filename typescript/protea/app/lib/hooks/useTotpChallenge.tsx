import { useFetcher } from '@remix-run/react'
import { useEffect, useState } from 'react'
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
 * Hook for inline TOTP challenge popup (no redirects)
 * Returns:
 * - popTotp: function to trigger the TOTP challenge popup
 * - TotpPopup: component to render the popup
 */
export function useTotpChallenge(options?: UseTotpChallengeOptions) {
  const [isOpen, setIsOpen] = useState(false)
  const initFetcher = useFetcher<TotpChallengeData>()

  // Trigger TOTP challenge popup
  const popTotp = () => {
    // Initialize the flow
    console.log('🍀 (useTotpChallenge) Triggering TOTP challenge popup')
    initFetcher.load('/api/totp-challenge-init')
    setIsOpen(true)
  }

  // Close popup
  const closePopup = () => {
    setIsOpen(false)
  }

  // Handle success
  const handleSuccess = () => {
    console.log('✅ TOTP Challenge: Session upgraded to AAL2')
    options?.onSuccess?.()
    setIsOpen(false)
  }

  // Handle error
  const handleError = (error: string) => {
    console.error('❌ TOTP Challenge failed:', error)
    options?.onError?.(error)
  }

  // Handle init errors (e.g., TOTP not configured)
  useEffect(() => {
    if (initFetcher.data?.error) {
      console.error(
        'Failed to initialize TOTP challenge:',
        initFetcher.data.error
      )
      setIsOpen(false)
      options?.onError?.(initFetcher.data.error)
    }
  }, [initFetcher.data, options])

  // Popup component
  const TotpPopup = () => {
    // Don't render until we have flow data
    if (!initFetcher.data || initFetcher.data.error) {
      return null
    }

    return (
      <TotpChallengePopup
        flowId={initFetcher.data.flowId}
        csrfToken={initFetcher.data.csrfToken}
        isOpen={isOpen}
        onClose={closePopup}
        onSuccess={handleSuccess}
        onError={handleError}
      />
    )
  }

  return {
    popTotp,
    TotpPopup,
    isLoading: initFetcher.state === 'loading',
    isOpen
  }
}
