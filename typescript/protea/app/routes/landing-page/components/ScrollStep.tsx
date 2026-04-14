import { useRef, useEffect } from "react"
import { useScroll, useMotionValueEvent } from "framer-motion"
import type { MotionValue } from "framer-motion"
import type { CarouselScreen } from "../context/PhoneCarouselContext"
import { usePhoneCarousel } from "../context/PhoneCarouselContext"

interface ScrollStepProps {
  /** Which phone carousel screen to activate when this step is in view */
  screen: CarouselScreen
  /** Called on every scroll frame with this step's progress (0→1). */
  onProgress?: (progress: MotionValue<number>) => void
  /** CSS class for the step spacer (controls responsive height) */
  className?: string
  children?: React.ReactNode
}

/**
 * Invisible scroll spacer that drives one "step" of the scrollytelling sequence.
 *
 * Layout pattern:
 *   <section>
 *     <div class="sticky-viewport">  ← pinned content (phone, text, glow)
 *     <ScrollStep screen={1} />       ← creates scroll height, tracks progress
 *     <ScrollStep screen={2} />
 *     ...
 *   </section>
 *
 * Each ScrollStep has its own `useScroll({ target })`, so `scrollYProgress` is
 * 0→1 relative to THIS step's height — not the entire section. This makes it
 * responsive by construction: step heights are CSS-controlled (media queries),
 * and progress values stay valid across all breakpoints.
 *
 * When the step's progress enters the active range (0.2–0.8), it sets itself
 * as the active carousel screen via PhoneCarouselContext.
 */
export function ScrollStep({ screen, onProgress, className, children }: ScrollStepProps) {
  const ref = useRef<HTMLDivElement>(null)
  const { setActiveScreen } = usePhoneCarousel()

  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  })

  // Expose the MotionValue to parent so it can derive transforms
  useEffect(() => {
    onProgress?.(scrollYProgress)
  }, [scrollYProgress, onProgress])

  // Advance carousel when this step is the dominant one in the viewport
  useMotionValueEvent(scrollYProgress, "change", (v) => {
    if (v > 0.2 && v < 0.8) {
      setActiveScreen(screen)
    }
  })

  return (
    <div ref={ref} className={className} aria-hidden="true">
      {children}
    </div>
  )
}
