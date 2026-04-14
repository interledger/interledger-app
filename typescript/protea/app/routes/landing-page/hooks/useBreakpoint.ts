import { useState, useEffect } from "react"

export type Breakpoint = "mobile" | "tablet" | "desktop"

// Matches Figma breakpoints from layout-and-heights.md:
// mobile = 360px, tablet = 900px, desktop = 1440px
const TABLET_MIN = 900
const DESKTOP_MIN = 1280

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
    function handleResize() {
      setBreakpoint(getBreakpoint(window.innerWidth))
    }
    window.addEventListener("resize", handleResize)
    return () => window.removeEventListener("resize", handleResize)
  }, [])

  return breakpoint
}
