import type { ReactNode } from 'react'
import { createContext, useContext, useState } from 'react'

export type CarouselScreen = 1 | 2 | 3 | 4 | 5

interface PhoneCarouselContextValue {
  activeScreen: CarouselScreen
  setActiveScreen: (screen: CarouselScreen) => void
}

export const PhoneCarouselContext = createContext<PhoneCarouselContextValue>({
  activeScreen: 1,
  setActiveScreen: () => {}
})

/**
 * Provides the active phone carousel screen (1–5) to all descendant components.
 *
 * Usage:
 *   const { activeScreen, setActiveScreen } = usePhoneCarousel()
 */
export function PhoneCarouselProvider({ children }: { children: ReactNode }) {
  const [activeScreen, setActiveScreen] = useState<CarouselScreen>(1)

  return (
    <PhoneCarouselContext.Provider value={{ activeScreen, setActiveScreen }}>
      {children}
    </PhoneCarouselContext.Provider>
  )
}

export function usePhoneCarousel() {
  return useContext(PhoneCarouselContext)
}
