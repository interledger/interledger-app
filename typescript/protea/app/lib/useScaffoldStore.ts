import { create } from 'zustand'

export type SnackbarType = {
  id?: string
  message: string
  show?: boolean
  action?: string
  icon?: string
}

interface ScaffoldState {
  loading: boolean
  snackbar: SnackbarType
}

interface ScaffoldActions {
  setLoading: (loading: boolean) => void
  setSnackbar: (snackbar: SnackbarType) => void
  hideSnackbar: () => void
  reset: () => void
}

const scaffoldInitialState = {
  loading: false,
  snackbar: { id: '', message: '', show: false }
}

export const useScaffoldStore = create<ScaffoldState & ScaffoldActions>(
  (set) => ({
    ...scaffoldInitialState,
    setLoading: (loading) => set((state) => ({ loading })),
    setSnackbar: (snackbar) =>
      set((state) => {
        if (state.snackbar.id != snackbar.id) return { snackbar }
        return { snackbar: { ...snackbar, show: false } }
      }),
    hideSnackbar: () =>
      set((state) => ({ snackbar: { ...state.snackbar, show: false } })),
    reset: () => set((state) => ({ ...scaffoldInitialState }))
  })
)
