import { useFetcher } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { TotpChallengePopup } from '~/components/TotpChallengePopup'

interface TotpChallengeData {
  flowId: string
  csrfToken: string
  error?: string
}

interface TotpCheckData {
  enabled: boolean
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
 * - isTotpEnabled: boolean indicating if user has TOTP configured
 */
export function useTotpChallenge(options?: UseTotpChallengeOptions) {
  const [isOpen, setIsOpen] = useState(false)
  const isTotpEnabledFetcher = useFetcher<TotpCheckData>()
  const initChallengeFetcher = useFetcher<TotpChallengeData>()

  // Trigger TOTP challenge popup
  const popTotp = () => {
    console.log('🍀 (useTotpChallenge) Checking if TOTP is enabled...')
    // First, check if user has TOTP enabled
    isTotpEnabledFetcher.load('/api/check-totp-enabled')
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

  // Handle TOTP enabled check result
  useEffect(() => {
    if (isTotpEnabledFetcher.data?.enabled) {
      console.log('🍀 (useTotpChallenge) TOTP is enabled, initializing flow...')
      // User has TOTP enabled, proceed with flow initialization
      initChallengeFetcher.load('/api/totp-challenge-init')
      setIsOpen(true)
    } else {
      console.log(
        '🍀 (useTotpChallenge) TOTP is not enabled, aborting challenge'
      )
      // User doesn't have TOTP set up yet, do nothing (silently abort)
      // This is expected behavior, not an error
    }
  }, [isTotpEnabledFetcher.data])

  // Handle init errors (e.g., flow expired, network issues)
  useEffect(() => {
    if (initChallengeFetcher.data?.error) {
      console.error(
        'Failed to initialize TOTP challenge:',
        initChallengeFetcher.data.error
      )
      setIsOpen(false)
      options?.onError?.(initChallengeFetcher.data.error)
    }
  }, [initChallengeFetcher.data, options])

  // Popup component
  const TotpPopup = () => {
    // Don't render until we have flow data
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
  

  return {
    popTotp,
    TotpPopup,
    isLoading: initChallengeFetcher.state === 'loading',
    isOpen,
    isTotpEnabled: isTotpEnabledFetcher.data?.enabled ?? false
  }
}
