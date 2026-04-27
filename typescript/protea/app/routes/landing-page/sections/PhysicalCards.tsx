import { motion } from "framer-motion"
import leftCardUrl from "../assets/cards-section/left.svg"
import centerCardUrl from "../assets/cards-section/center.svg"
import rightCardUrl from "../assets/cards-section/right.svg"

// Parent stagger control
const fanVariants = {
  hidden: {},
  visible: {
    transition: {
      staggerChildren: 0.15,
      delayChildren: 0.1, // Wait briefly before starting fan
    }
  }
}

// Child entry properties
// Since the SVGs already strictly have their offsets and rotations baked deeply into their bounding boxes in Figma,
// we just position them at their exact relative offsets on X and Y.
// To satisfy "I want only a rotation", we keep X and Y static across hidden and visible states.
// We apply an initial CSS rotation to counter-act the baked-in tilt, giving a pure rotation-only drop-in effect.
const leftCardVariants = {
  hidden: { rotate: 20, scale: 1, x: "calc(-50% - 129px)", y: "calc(-50% - 65.5px)" },
  visible: { 
    rotate: 0,
    scale: 1,
    x: "calc(-50% - 129px)", 
    y: "calc(-50% - 65.5px)",
    transition: { duration: 0.6, ease: "easeOut" }
  }
}

const rightCardVariants = {
  hidden: { rotate: -20, scale: 1, x: "calc(-50% + 129px)", y: "calc(-50% - 65.5px)" },
  visible: { 
    rotate: 0,
    scale: 1,
    x: "calc(-50% + 129px)", 
    y: "calc(-50% - 65.5px)",
    transition: { duration: 0.6, ease: "easeOut" }
  }
}

const centerCardVariants = {
  hidden: { rotate: 0, scale: 1, x: "-50%", y: "-50%" },
  visible: { 
    rotate: 0,
    scale: 1,
    x: "-50%", 
    y: "-50%",
    transition: { duration: 0.6, ease: "easeOut" }
  }
}

export function PhysicalCards() {
  return (
    <section className="section-physical-cards content-section">
      <div className="cards-heading-container">
        <h2 className="text-h1 cards-heading">Virtual and physical cards</h2>
      </div>

      <motion.div 
        className="cards-fan-container anim-cards-fan"
        variants={fanVariants}
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true, amount: 0.3 }}
      >
        <div className="cards-cluster">
          {/* Left Card */}
          <motion.div variants={leftCardVariants} className="physical-card-layer">
            <img src={leftCardUrl} alt="Left Debit Card" className="physical-card-img physical-card-img--left" />
          </motion.div>
          
          {/* Right Card */}
          <motion.div variants={rightCardVariants} className="physical-card-layer">
            <img src={rightCardUrl} alt="Right Debit Card" className="physical-card-img physical-card-img--right" />
          </motion.div>

          {/* Center Card */}
          <motion.div variants={centerCardVariants} className="physical-card-layer">
            <img src={centerCardUrl} alt="Center Debit Card" className="physical-card-img physical-card-img--center" />
          </motion.div>
        </div>
      </motion.div>
    </section>
  )
}
