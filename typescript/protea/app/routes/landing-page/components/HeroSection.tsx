import { useRef } from "react"
import { motion, useScroll, useTransform, useMotionValueEvent } from "framer-motion"

import { PhoneFrame } from "./PhoneFrame"
import { ScrollStep } from "./ScrollStep"
import { usePhoneCarousel } from "../context/PhoneCarouselContext"
import type { CarouselScreen } from "../context/PhoneCarouselContext"
import { FeatureSection } from "./FeatureSection"
import { FeatureWidget } from "./FeatureWidget"
import orbitSvg from "../assets/orbit.svg"

const FEATURES = [
  {
    screen: 2,
    heading: "Global by design. Inclusive by default.",
    body: "Designed to meet people where they are, how they are.",
    visual: (
      <FeatureWidget
        avatar="MH"
        avatarColor="#e87a7a"
        name="Mike H"
        amount="+ $348"
        note="tnx for the adventure"
        timestamp="28.10.2026 21:57"
      />
    ),
  },
  {
    screen: 3,
    heading: "Broad in reach. Borderless by design.",
    body: "Built broad, built borderless. Designed to work everywhere.",
    widget: <img src={orbitSvg} alt="Orbit graphic" loading="lazy" decoding="async" style={{ width: "80px", marginTop: "16px", display: "block", marginInline: "auto" }} />,
  },
  {
    screen: 4,
    heading: "One system. Many contexts.",
    body: "Ready to diverse environments and needs.",
    widget: <img src={orbitSvg} alt="Orbit graphic" loading="lazy" decoding="async" style={{ width: "80px", marginTop: "16px", display: "block", marginInline: "auto" }} />,
  },
  {
    screen: 5,
    heading: "Designed to adopt. Easy to adopt.",
    body: "We are building the wallet in the open and want to participate from the start.",
    visual: (
      <div style={{ marginTop: "16px", display: "flex", justifyContent: "center" }}>
        <a href="/signup" className="nav-cta" style={{ display: "inline-flex" }}>
          Sign up now <span style={{ marginLeft: "4px" }}>&rarr;</span>
        </a>
      </div>
    ),
  },
]

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
 */
export function HeroSection() {
  const sectionRef = useRef<HTMLElement>(null)
  const { activeScreen, setActiveScreen } = usePhoneCarousel()

  // Scope scroll progress to the full animated-hero section
  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ["start start", "end end"],
  })

  const thresholds = [0.20, 0.40, 0.60, 0.80] // screen 2..5
  useMotionValueEvent(scrollYProgress, "change", v => {
    const next = (1 + thresholds.filter(t => v >= t).length) as CarouselScreen
    if (next !== activeScreen) setActiveScreen(next)
  })

  const y = useTransform(scrollYProgress, [0, 0.2], [0, -40])
  const opacity = useTransform(scrollYProgress, [0, 0.125], [1, 0])

  return (

    <div className="page-content">
      <section ref={sectionRef} className="animated-hero" data-screen={activeScreen}>
        <div className="sticky-viewport">
          <div className="sticky-viewport-content">
            {/* Hero text — headline, subhead, CTA */}
            <div className="hero-content">
              <motion.div style={{ y, opacity }} className="hero-punch-scroll">
                <div
                  className="hero-punch anim-enter"
                  data-anim="punch-text"
                >
                  <h1 className="text-h1 hero-headline">A wallet for what&apos;s next</h1>
                  <p className="text-h3 hero-subhead">
                    Built for interoperability, inclusion,<br />and the long run.
                  </p>
                </div>
              </motion.div>
            </div>

            <PhoneFrame />

            {/* Feature content slot — all feature panels live here as absolute overlays */}
            <div className="feature-content-slot">
              {FEATURES.map(f => (
                <FeatureSection
                  key={f.screen}
                  screen={f.screen as CarouselScreen}
                  heading={f.heading}
                  body={f.body}
                  widget={f.widget}
                  visual={f.visual}
                />
              ))}
            </div>
          </div>
        </div>

        {/* Scroll spacers */}
        <ScrollStep screen={1} className="scroll-step scroll-step--hero" />
        <ScrollStep screen={2} className="scroll-step scroll-step--feature" />
        <ScrollStep screen={3} className="scroll-step scroll-step--feature" />
        <ScrollStep screen={4} className="scroll-step scroll-step--feature" />
        <ScrollStep screen={5} className="scroll-step scroll-step--feature" />
      </section>
    </div>
  )
}
