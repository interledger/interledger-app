import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import { create } from 'zustand'
import type { Country } from '~/generated/connect/backend/v1/backend_pb'

export enum SignupStep {
  LANDING,
  ABOUT,
  PHONE,
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
  phone: string
  isCompleted: boolean
}

interface SignupActions {
  setStep: (step: SignupStep) => void
  stepBack: () => void
  setFirstName: (firstName: string) => void
  setLastName: (lastName: string) => void
  setEmail: (email: string) => void
  setCountry: (country: PlainMessage<Country>) => void
  setCountries: (countries: PlainMessage<Country>[]) => void
  setDetails: (
    id: string,
    firstName: string,
    lastName: string,
    email: string
  ) => void
  setPhone: (phone: string) => void
  reset: () => void
}

type StateType = SignupState & SignupActions

const signupInitialState = {
  isCompleted: false,
  step: SignupStep.LANDING,
  id: '',
  firstName: '',
  lastName: '',
  email: '',
  country: null,
  countries: [],
  phone: ''
}

export const useSignupStore = create<StateType>((set) => ({
  ...signupInitialState,
  setStep: (step) => set((state) => ({ step: step })),
  stepBack: () =>
    set((state) => {
      switch (state.step) {
        case SignupStep.PASSWORD:
          return { step: SignupStep.PHONE }
        case SignupStep.PHONE:
          return { step: SignupStep.ABOUT }
        case SignupStep.ABOUT:
          return { step: SignupStep.LANDING }
        default:
          return { ...signupInitialState }
      }
    }),
  setFirstName: (firstName) =>
    set((state) => ({
      isCompleted: !!(state.lastName && state.email && state.country && firstName),
      firstName
    })),
  setLastName: (lastName) =>
    set((state) => ({
      isCompleted: !!(state.firstName && state.email && state.country && lastName),
      lastName
    })),
  setEmail: (email) =>
    set((state) => ({
      isCompleted: !!(state.firstName && state.lastName && state.country && email),
      email
    })),
  setCountry: (country) =>
    set((state) => ({
      isCompleted: !!(state.firstName && state.lastName, state.email && country),
      country
    })),
  setCountries: (countries) => set((state) => ({ countries })),
  setDetails: (id, firstName, lastName, email) =>
    set((state) => ({ id, firstName, lastName, email })),
  setPhone: (phone) => set((state) => ({ phone })),
  reset: () => set((state) => ({ ...signupInitialState }))
}))
