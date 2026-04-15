import { usePhoneCarousel } from "../context/PhoneCarouselContext"
import type { CarouselScreen } from "../context/PhoneCarouselContext"
import { motion } from "framer-motion"
import { Glow, type GlowScrollTransform } from "./Glow"

import startScreen from "../assets/start-screen.svg"
import screen2 from "../assets/screen2.png"
import screen3 from "../assets/screen3.png"
import screen4 from "../assets/screen4.png"
import screen5 from "../assets/screen5.png"

const SCREEN_ASSETS: Record<CarouselScreen, string> = {
  1: startScreen,
  2: screen2,
  3: screen3,
  4: screen4,
  5: screen5,
}

const GLOW_STATES: Record<CarouselScreen, GlowScrollTransform> = {
  1: { scale: 1.2, rotate: 0,   y: "-50%", opacity: 1   },
  2: { scale: 1,   rotate: 180, y: "-50%", opacity: 1   },
  3: { scale: 1,   rotate: 270, y: "-50%", opacity: 0.5 },
  4: { scale: 1,   rotate: 360, y: "-50%", opacity: 1   },
  5: { scale: 1,   rotate: 450, y: "-50%", opacity: 1   },
}

/**
 * Phone mockup that displays the active carousel screen.
 *
 * Dimensions (from Figma):
 *   Mobile viewport:  170 × 354px
 *   Desktop viewport: 195 × 410px (rendered at 200 × 416px frame)
 *
 * Screen 1 is loaded eagerly (LCP-critical — visible on page load).
 * Screens 2–5 are lazy-loaded (only visible after significant scroll).
 * See design/concerns/05-phone-screen-images.md
 */
export function PhoneFrame() {
  const { activeScreen } = usePhoneCarousel()

  return (
    <div className="hero-phone-container">
      <Glow x="50%" y="50%" scrollTransform={GLOW_STATES[activeScreen]} />
      <div className="phone-frame" aria-label={`App screen ${activeScreen}`}>
        <div className="phone-dynamic-island" />
        <div className="phone-screen-viewport">
          <motion.div
            style={{
              display: "flex",
              width: "500%",
              height: "100%",
            }}
            initial={false}
            animate={{ x: `-${(activeScreen - 1) * 20}%` }}
            transition={{ type: "spring", stiffness: 300, damping: 30 }}
          >
            {Object.entries(SCREEN_ASSETS).map(([screen, src]) => {
              const s = Number(screen) as CarouselScreen
              return (
                <img
                  key={s}
                  src={src}
                  alt=""
                  className="phone-screen"
                  loading={s === 1 ? "eager" : "lazy"}
                  style={{ width: "20%", height: "100%", objectFit: "cover" }}
                />
              )
            })}
          </motion.div>
        </div>
      </div>
    </div>
  )
}
