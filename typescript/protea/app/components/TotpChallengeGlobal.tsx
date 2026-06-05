import { useTotpChallengeStore } from '~/lib/useTotpChallengeStore'
import { TotpChallengePopup } from './TotpChallengePopup'

/**
 * Global TOTP Challenge Popup
 * Renders when store.isOpen is true
 */
export function TotpChallengeGlobal() {
  const { isOpen, challengeData, closeChallenge, handleSuccess, handleError } =
    useTotpChallengeStore()

  if (!isOpen || !challengeData || challengeData.error) {
    return null
  }

  return (
    <TotpChallengePopup
      flowId={challengeData.flowId}
      csrfToken={challengeData.csrfToken}
      onClose={closeChallenge}
      onSuccess={handleSuccess}
      onError={handleError}
    />
  )
}
