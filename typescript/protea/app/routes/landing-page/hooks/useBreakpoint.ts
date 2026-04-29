import { useState, useEffect } from "react"

export type Breakpoint = "mobile" | "tablet" | "desktop"

// Matches Figma breakpoints: mobile < 900px, tablet < 1440px, desktop >= 1440px
const TABLET_MIN = 900
const DESKTOP_MIN = 1440

function getBreakpoint(width: number): Breakpoint {
  if (width < TABLET_MIN) return "mobile"
  if (width < DESKTOP_MIN) return "tablet"
  return "desktop"
}

/**
 * Returns the current breakpoint based on window width.
 * SSR-safe: defaults to "desktop" on the server.
 * Only use for JS-only breakpoint logic — prefer CSS media queries for styling.
 */
export function useBreakpoint(): Breakpoint {
  const [breakpoint, setBreakpoint] = useState<Breakpoint>(() => {
    if (typeof window === "undefined") return "desktop"
    return getBreakpoint(window.innerWidth)
  })

  useEffect(() => {
    const tabletMql = window.matchMedia(`(min-width: ${TABLET_MIN}px)`)
    const desktopMql = window.matchMedia(`(min-width: ${DESKTOP_MIN}px)`)

    function update() {
      setBreakpoint(getBreakpoint(window.innerWidth))
    }

    tabletMql.addEventListener("change", update)
    desktopMql.addEventListener("change", update)

    return () => {
      tabletMql.removeEventListener("change", update)
      desktopMql.removeEventListener("change", update)
    }
  }, [])

  return breakpoint
}
