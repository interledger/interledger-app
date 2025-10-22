import { useFetcher } from '@remix-run/react'
import { useCallback, useEffect, useRef } from 'react'
import { useTotpChallengeStore } from '../useTotpChallengeStore'

interface TotpChallengeData {
  flowId: string
  csrfToken: string
  error?: string
}

/**
 * Hook to trigger TOTP challenge
 * Usage:
 *   const { withTotpChallenge } = useTotpChallengeAction()
 *   withTotpChallenge(() => doSomething())
 */
export function useTotpChallengeAction() {
  const initChallengeFetcher = useFetcher<TotpChallengeData>()
  const { openChallenge, handleError } = useTotpChallengeStore()
  const pendingCallbackRef = useRef<(() => void) | null>(null)

  // When challenge data arrives, open the popup
  useEffect(() => {
    if (initChallengeFetcher.data) {
      if (initChallengeFetcher.data?.error) {
        handleError(initChallengeFetcher.data.error)
        return
      }

      if (pendingCallbackRef.current) {
        openChallenge(initChallengeFetcher.data, pendingCallbackRef.current)
        pendingCallbackRef.current = null
      }
    }
  }, [initChallengeFetcher.data, openChallenge, handleError])

  const withTotpChallenge = useCallback(
    (callback: () => void) => {
      pendingCallbackRef.current = callback
      initChallengeFetcher.submit({}, { method: 'post', action: '/api/totp-challenge-init' })
    },
    [initChallengeFetcher]
  )

  return {
    withTotpChallenge,
    isLoading: initChallengeFetcher.state === 'loading'
  }
}
