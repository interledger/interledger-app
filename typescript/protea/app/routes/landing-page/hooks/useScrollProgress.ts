import { useRef } from "react"
import { useScroll } from "framer-motion"
import type { MotionValue } from "framer-motion"

interface UseScrollProgressReturn {
  ref: React.RefObject<HTMLDivElement>
  scrollYProgress: MotionValue<number>
}

/**
 * Scopes a Framer Motion scroll tracker to a specific container element.
 * scrollYProgress is 0 when the element's top hits the viewport top,
 * and 1 when the element's bottom hits the viewport bottom.
 *
 * Always prefer this over global useScroll() — see design/concerns/02-usescroll-scope.md
 */
export function useScrollProgress(
  offset: ["start start", "end end"] | ["start start", "end start"] = ["start start", "end end"],
): UseScrollProgressReturn {
  const ref = useRef<HTMLDivElement>(null)
  const { scrollYProgress } = useScroll({ target: ref, offset })
  return { ref, scrollYProgress }
}
