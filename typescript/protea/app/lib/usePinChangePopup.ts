import { useCallback } from 'react'
import { PinChangePopup } from '~/components/CardView/PinChangePopup'
import { usePinChangePopupStore } from './usePinChangePopupStore'

/**
 * Hook to trigger PIN change popup
 * Usage:
 *   const { withPinChangePopup, PinChangePopup } = usePinChangePopup()
 *   // Render <PinChangePopup /> in your component
 *   // Call withPinChangePopup((newPin: string) => doSomethingWithPin(newPin))
 */
export function usePinChangePopup() {
  const { openPopup } = usePinChangePopupStore()

  const withPinChangePopup = useCallback(
    (callback: (pin: string) => void) => {
      openPopup(callback)
    },
    [openPopup]
  )

  return {
    withPinChangePopup,
    PinChangePopup
  }
}
