import type { ReactNode } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { usePhoneCarousel } from "../context/PhoneCarouselContext"
import type { CarouselScreen } from "../context/PhoneCarouselContext"

interface FeatureSectionProps {
  /** Show this panel when carousel activeScreen matches */
  screen: CarouselScreen
  heading: string
  body?: string
  widget?: ReactNode
  visual?: ReactNode
  columnOrder?: "text-left" | "text-right"
}


const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      type: "spring",
      duration: 1,
      bounce: 0,
      staggerChildren: 0.1
    }
  },
  exit: {
    opacity: 0,
    transition: {
      type: "spring",
      duration: 0.8,
      bounce: 0
    }
  }
}

const childVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      type: "spring",
      duration: 1,
      bounce: 0
    }
  }
}


/**
 * Absolute overlay panel inside .feature-content-slot.
 * Rendered for every feature; only the panel whose `screen` matches
 * PhoneCarouselContext.activeScreen is interactive + visible (via .is-visible).
 *
 * Layout: [text col] [phone spacer] [visual col]
 * The phone spacer creates a transparent gap that mirrors the phone width,
 * so the two columns naturally flank the sticky phone frame.
 */
export function FeatureSection({
  screen,
  heading,
  body,
  widget,
  visual,
}: FeatureSectionProps) {
  const { activeScreen } = usePhoneCarousel()

  const textCol = (
    <motion.div className="feature-col feature-col--left">
      <motion.h2 variants={childVariants} className="text-h2 feature-heading">{heading}</motion.h2>
    </motion.div>
  )

  const rightCol = (
    <div className="feature-col feature-col--right">
      {body && <motion.p variants={childVariants} className="text-body-lg feature-body">{body}</motion.p>}
      {widget && <motion.div variants={childVariants}>{widget}</motion.div>}
      {visual && <motion.div variants={childVariants}>{visual}</motion.div>}
    </div>
  )

  return (
    <AnimatePresence>
      {activeScreen === screen && (
        <motion.div
          className="feature-panel is-visible"
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          exit="exit"
        >
          <>
            {textCol}
            <div className="feature-phone-spacer" />
            {rightCol}
          </>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
