import { useRef } from "react"
import { Link } from "react-router"
import { motion } from "framer-motion"

import { PhoneFrame } from "./PhoneFrame"
import { usePhoneCarousel } from "../context/PhoneCarouselContext"
import type { CarouselScreen } from "../context/PhoneCarouselContext"
import { FeatureSection } from "./FeatureSection"
import { FeatureWidget } from "./FeatureWidget"
import { useHeroSnap } from "../hooks/useHeroSnap"
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
      <div className="hero-signup-cta-wrap">
        <Link to="/signup" className="nav-cta nav-cta--hero">
          Sign up now <span aria-hidden="true" style={{ marginLeft: "4px" }}>&rarr;</span>
        </Link>
      </div>
    ),
  },
]

/**
 * Animated Hero Section — snap-per-event carousel.
 *
 * Hero is a single 100vh pinned viewport. Document scroll is locked while
 * hero is engaged; each wheel/touch/key gesture advances `activeScreen` by
 * exactly one (1..5). At edges (5 going down, 1 going up) the lock releases
 * and native scroll resumes into the rest of the page.
 */
export function HeroSection() {
  const sectionRef = useRef<HTMLElement>(null)
  const { activeScreen, setActiveScreen } = usePhoneCarousel()

  useHeroSnap({
    sectionRef,
    activeScreen,
    setActiveScreen,
    screenCount: 5,
  })

  const heroTextVisible = activeScreen === 1

  return (
    <div className="page-content">
      <section ref={sectionRef} className="animated-hero" data-screen={activeScreen}>
        <div className="sticky-viewport">
          <div className="sticky-viewport-content">
            {/* Hero text — headline, subhead, CTA */}
            <div className="hero-content">
              <motion.div
                className="hero-punch-scroll"
                animate={{
                  y: heroTextVisible ? 0 : -40,
                  opacity: heroTextVisible ? 1 : 0,
                }}
                transition={{ type: "spring", duration: 0.8, bounce: 0 }}
              >
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
      </section>
    </div>
  )
}
