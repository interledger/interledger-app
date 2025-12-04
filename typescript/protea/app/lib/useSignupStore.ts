import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import { create } from 'zustand'
import type { Country } from '~/generated/connect/backend/v1/backend_pb'

export enum SignupStep {
  LANDING,
  ABOUT,
  PASSWORD
}

interface SignupState {
  step: SignupStep
  id: string
  firstName: string
  lastName: string
  email: string
  country: PlainMessage<Country> | null
  countries: PlainMessage<Country>[]
}

interface SignupActions {
  setStep: (step: SignupStep) => void
  stepBack: () => void
  setCountry: (country: PlainMessage<Country>) => void
  setCountries: (countries: PlainMessage<Country>[]) => void
  setDetails: (
    id: string,
    firstName: string,
    lastName: string,
    email: string
  ) => void
  reset: () => void
}

const signupInitialState = {
  step: SignupStep.LANDING,
  id: '',
  firstName: '',
  lastName: '',
  email: '',
  country: null,
  countries: []
}

export const useSignupStore = create<SignupState & SignupActions>((set) => ({
  ...signupInitialState,
  setStep: (step) => set((state) => ({ step: step })),
  stepBack: () =>
    set((state) => {
      switch (state.step) {
        case SignupStep.PASSWORD:
          return { step: SignupStep.ABOUT }
        case SignupStep.ABOUT:
          return { step: SignupStep.LANDING }
        default:
          return { ...signupInitialState }
      }
    }),
  setCountry: (country) => set((state) => ({ country })),
  setCountries: (countries) => set((state) => ({ countries })),
  setDetails: (id, firstName, lastName, email) =>
    set((state) => ({ id, firstName, lastName, email })),
  reset: () => set((state) => ({ ...signupInitialState }))
}))
