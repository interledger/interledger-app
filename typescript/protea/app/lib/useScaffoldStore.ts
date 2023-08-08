import { create } from 'zustand'

export type SnackbarType = {
  id?: string
  message: string
  canShow?: boolean
  fromServer?: boolean
  action?: string
  icon?: string
}

interface ScaffoldState {
  loading: boolean
  snackbars: SnackbarType[]
}

interface ScaffoldActions {
  setLoading: (loading: boolean) => void
  pushSnackbar: (snackbar: SnackbarType) => void
  shiftSnackbar: () => void
  reset: () => void
}

const scaffoldInitialState = {
  loading: false,
  snackbars: []
}

export const useScaffoldStore = create<ScaffoldState & ScaffoldActions>()(
  (set) => ({
    ...scaffoldInitialState,
    setLoading: (loading) => set((state) => ({ loading })),
    pushSnackbar: (snackbar) =>
      set((state) => {
        if (state.snackbars.findIndex((s) => s.id == snackbar.id) == -1) {
          snackbar.canShow = true
          state.snackbars.push(snackbar)
        }
        return { snackbars: state.snackbars }
      }),
    shiftSnackbar: () =>
      set((state) => {
        const visible = state.snackbars.findIndex((s) => s.canShow)
        if (visible != -1) {
          // We found the currently visible snackbar!
          if (state.snackbars[visible].fromServer) {
            // If it's from the server, we need to set canShow to false
            // The root loader doesn't revalidate on client side navigation
            state.snackbars[visible].canShow = false
          } else {
            // If it's not from the server, we can just remove it
            state.snackbars.splice(visible, 1)
          }
        }
        return {
          snackbars: state.snackbars
        }
      }),
    reset: () => set((state) => ({ ...scaffoldInitialState }))
  })
)
