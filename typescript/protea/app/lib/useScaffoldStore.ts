import { create } from 'zustand'

export type SnackbarAction =
  | 'Contact support'
  | 'View cards'
  | 'Update mobile number'

export type SnackbarType = {
  id: string
  message: string
  canShow?: boolean
  fromServer?: boolean
  action?: SnackbarAction
  icon?: string
}

interface ScaffoldState {
  loading: boolean
  commandPalletOpen: boolean
  snackbars: SnackbarType[]
}

interface ScaffoldActions {
  setLoading: (loading: boolean) => void
  setCommandPalletOpen: (open: boolean) => void
  pushSnackbar: (snackbar: SnackbarType) => void
  shiftSnackbar: () => void
  reset: () => void
}

const scaffoldInitialState = {
  loading: false,
  commandPalletOpen: false,
  snackbars: []
}

export const useScaffoldStore = create<ScaffoldState & ScaffoldActions>()(
  (set) => ({
    ...scaffoldInitialState,
    setLoading: (loading) => set((state) => ({ loading })),
    setCommandPalletOpen: (commandPalletOpen) =>
      set((state) => ({ commandPalletOpen })),
    pushSnackbar: (snackbar) =>
      set((state) => {
        if (state.snackbars.findIndex((s) => s.id == snackbar.id) == -1) {
          let newSnackbars = [...state.snackbars]
          snackbar.canShow = true
          newSnackbars.push(snackbar)
          return { snackbars: newSnackbars }
        }
        return { snackbars: state.snackbars }
      }),
    shiftSnackbar: () =>
      set((state) => {
        const visibleIndex = state.snackbars.findIndex((s) => s.canShow)
        if (visibleIndex != -1) {
          // We found the currently visible snackbar!
          let newSnackbars = [...state.snackbars]
          if (state.snackbars[visibleIndex].fromServer) {
            // If it's from the server, we need to set canShow to false
            // Because the root loader doesn't revalidate on client side navigation
            newSnackbars[visibleIndex].canShow = false
          } else {
            // If it's not from the server, we can just remove it
            newSnackbars.splice(visibleIndex, 1)
          }
          return { snackbars: newSnackbars }
        }
        return {
          snackbars: state.snackbars
        }
      }),
    reset: () => set((state) => ({ ...scaffoldInitialState }))
  })
)
