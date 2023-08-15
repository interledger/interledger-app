import { create } from 'zustand'

export enum ConnectDomainStep {
  LANDING,
  NAME
}

interface ConnectDomainState {
  step: ConnectDomainStep
  id: string
  domainName: string
}

interface ConnectDomainActions {
  setStep: (step: ConnectDomainStep) => void
  stepBack: () => void
  reset: () => void
}

const connectDomainInitialState = {
  step: ConnectDomainStep.LANDING,
  id: '',
  domainName: ''
}

export const useConnectDomainStore = create<
  ConnectDomainState & ConnectDomainActions
>((set) => ({
  ...connectDomainInitialState,
  setStep: (step) => set((state) => ({ step: step })),
  stepBack: () =>
    set((state) => {
      switch (state.step) {
        case ConnectDomainStep.NAME:
          return { step: ConnectDomainStep.LANDING }
        default:
          return { ...connectDomainInitialState }
      }
    }),

  reset: () => set((state) => ({ ...connectDomainInitialState }))
}))
