import { useGateHubStore } from './useGateHubStore'

export const useCardsStore = () => {
  return useGateHubStore((state) => ({ ...state.card, ...state.actions.card }))
}

export const useTokenStore = () => {
  return useGateHubStore((state) => ({
    token: state.token,
    ...state.actions.token
  }))
}
