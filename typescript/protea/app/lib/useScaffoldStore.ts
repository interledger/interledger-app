import { create } from 'zustand'

interface ScaffoldState {
  loading: boolean
}

interface ScaffoldActions {
  setLoading: (loading: boolean) => void
  reset: () => void
}

const scaffoldInitialState = {
  loading: false
}

export const useScaffoldStore = create<ScaffoldState & ScaffoldActions>(
  (set) => ({
    ...scaffoldInitialState,
    setLoading: (loading) => set((state) => ({ loading })),
    reset: () => set((state) => ({ ...scaffoldInitialState }))
  })
)
