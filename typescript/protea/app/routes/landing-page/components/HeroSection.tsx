import { useRef, useState, useEffect } from "react"
import { motion, useScroll, useTransform } from "framer-motion"
import { Glow } from "./Glow"
import type { GlowScrollTransform } from "./Glow"
import { PhoneFrame } from "./PhoneFrame"
import { ScrollStep } from "./ScrollStep"
import { usePhoneCarousel } from "../context/PhoneCarouselContext"
import type { CarouselScreen } from "../context/PhoneCarouselContext"
import { Feature1 } from "../sections/Feature1"
import { Feature2 } from "../sections/Feature2"
import { Feature3 } from "../sections/Feature3"
import { Feature4 } from "../sections/Feature4"

// Per-screen glow transform — sourced from Figma Scroll Transform annotation.
// Glow lives at sticky-viewport level (not inside phone container) so it persists across all screens.
// y shifts −250px when entering feature screens; rotation increments 90° per section.
const GLOW_STATES: Record<CarouselScreen, GlowScrollTransform> = {
  1: { scale: 1.2, rotate: 0,   y: 0,    opacity: 1   },
  2: { scale: 1,   rotate: 180, y: -250, opacity: 1   },
  3: { scale: 1,   rotate: 270, y: -250, opacity: 0.5 },
  4: { scale: 1,   rotate: 360, y: -250, opacity: 1   },
  5: { scale: 1,   rotate: 450, y: -250, opacity: 1   },
}

/**
 * Animated Hero Section — scrollytelling container.
 *
 * Structure:
 *   <section class="animated-hero">             ← full scroll height (sum of step heights)
 *     <div class="sticky-viewport">             ← pinned to viewport (100vh)
 *       <div class="hero-content">              ← headline, subhead, CTA
 *         <div class="hero-punch-scroll">       ← JS scroll-parallax wrapper
 *           <div class="hero-punch anim-enter"> ← CSS enter animation wrapper
 *       <div class="hero-phone-container">      ← PhoneFrame lives here
 *       <Glow />                                ← positioned absolutely
 *       <div class="feature-content-slot">      ← features animate in/out here
 *     </div>
 *     <div class="scroll-step scroll-step--hero">       ← hero scroll spacer
 *     <div class="scroll-step scroll-step--feature">    ← feature 1 spacer (task 07)
 *     <div class="scroll-step scroll-step--feature">    ← feature 2 spacer (task 08)
 *     ...
 *   </section>
 *
 * Punch text animations:
 *   Enter effect  — CSS .anim-enter + .is-visible toggled after first paint (opacity 0→1, translateY 24px→0)
 *   Scroll parallax — useMotionValueEvent drifts text up -40px over first 30% of hero scroll
 */
export function HeroSection() {
  const sectionRef = useRef<HTMLElement>(null)
  const [isVisible, setIsVisible] = useState(false)
  const { activeScreen } = usePhoneCarousel()

  // Scope scroll progress to the full animated-hero section
  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ["start start", "end end"],
  })

  // Trigger enter animation after first paint
  useEffect(() => {
    const raf = requestAnimationFrame(() => setIsVisible(true))
    return () => cancelAnimationFrame(raf)
  }, [])

  // Scroll parallax: drift down (-40px) and fade opacity natively
  const y = useTransform(scrollYProgress, [0, 0.2], [0, -40])
  const opacity = useTransform(scrollYProgress, [0, 0.125], [1, 0])

  return (
    <section ref={sectionRef} className="animated-hero" data-screen={activeScreen}>
      <div className="sticky-viewport">
        {/* Hero text — headline, subhead, CTA */}
        <div className="hero-content">
          <motion.div style={{ y, opacity }} className="hero-punch-scroll">
            <div
              className={`hero-punch anim-enter${isVisible ? " is-visible" : ""}`}
              data-anim="punch-text"
            >
              <h1 className="text-h1 hero-headline">A wallet for what&apos;s next</h1>
              <p className="text-body-lg hero-subhead">
                Built for interoperability, inclusion,<br />and the long run.
              </p>
            </div>
          </motion.div>
        </div>

        {/* Glow — ambient backdrop behind phone + feature text; persists across all screens */}
        <Glow x="50%" y={300} scrollTransform={GLOW_STATES[activeScreen]} />

        <PhoneFrame />

        {/* Feature content slot — all feature panels live here as absolute overlays */}
        <div className="feature-content-slot">
          <Feature1 />
          <Feature2 />
          <Feature3 />
          <Feature4 />
        </div>
      </div>

      {/* Scroll spacers — hero dwell + Feature 1 dwell */}
      <ScrollStep screen={1} className="scroll-step scroll-step--hero" />
      <ScrollStep screen={2} className="scroll-step scroll-step--feature" />

      {/* Feature 2, 3, 4 scroll zones (single 320px dwell each) */}
      <ScrollStep screen={3} className="scroll-step scroll-step--feature" />
      <ScrollStep screen={4} className="scroll-step scroll-step--feature" />
      <ScrollStep screen={5} className="scroll-step scroll-step--feature" />
    </section>
  )
}
