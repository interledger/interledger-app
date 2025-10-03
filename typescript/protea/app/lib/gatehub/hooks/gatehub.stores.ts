import { useGateHubStore } from './useGateHubStore'

export const useCardsStore = () => {
  return useGateHubStore((state) => ({ ...state.card, ...state.actions.card }))
}
