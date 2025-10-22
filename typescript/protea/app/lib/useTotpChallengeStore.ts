import { create } from 'zustand'

interface TotpChallengeData {
  flowId: string
  csrfToken: string
  error?: string
}

interface TotpChallengeState {
  isOpen: boolean
  pendingCallback: (() => void) | null
  challengeData: TotpChallengeData | null
}

interface TotpChallengeActions {
  openChallenge: (data: TotpChallengeData, callback: () => void) => void
  closeChallenge: () => void
  handleSuccess: () => void
  handleError: (error: string) => void
  reset: () => void
}

const totpChallengeInitialState: TotpChallengeState = {
  isOpen: false,
  pendingCallback: null,
  challengeData: null
}

export const useTotpChallengeStore = create<
  TotpChallengeState & TotpChallengeActions
>()((set, get) => ({
  ...totpChallengeInitialState,
  openChallenge: (data, callback) =>
    set(() => ({
      isOpen: true,
      challengeData: data,
      pendingCallback: callback
    })),
  closeChallenge: () =>
    set(() => ({
      isOpen: false,
      pendingCallback: null
    })),
  handleSuccess: () => {
    const { pendingCallback } = get()
    if (pendingCallback) {
      pendingCallback()
    }
    set(() => ({
      isOpen: false,
      pendingCallback: null
    }))
  },
  handleError: (error) => {
    console.error('TOTP challenge error:', error)
    // Keep popup open for retry
  },
  reset: () => set(() => ({ ...totpChallengeInitialState }))
}))
