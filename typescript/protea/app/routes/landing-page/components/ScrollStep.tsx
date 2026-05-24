import type { CarouselScreen } from "../context/PhoneCarouselContext"

interface ScrollStepProps {
  /** Which phone carousel screen to activate when this step is in view */
  screen: CarouselScreen
  /** CSS class for the step spacer (controls responsive height) */
  className?: string
  children?: React.ReactNode
}

/**
 * Invisible scroll spacer that drives one "step" of the scrollytelling sequence.
 * This is a pure CSS spacer; scroll tracking is handled globally by HeroSection.
 */
export function ScrollStep({ className, children }: ScrollStepProps) {
  return (
    <div className={className} aria-hidden="true">
      {children}
    </div>
  )
}
