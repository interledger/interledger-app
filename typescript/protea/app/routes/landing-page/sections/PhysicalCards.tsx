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
// we just position them at their exact relative offsets on X and Y, with rotate: 0 to respect their internal matrix transforms!
// Offsets are derived precisely from the Absolute Bounding Boxes of the Figma node
const leftCardVariants = {
  hidden: { opacity: 0, y: "calc(-50% - 65.5px + 150px)", x: "calc(-50% - 157px)", rotate: 0, scale: 0.8 },
  visible: { 
    opacity: 1, 
    y: "calc(-50% - 65.5px)", 
    x: "calc(-50% - 157px)",
    rotate: 0,
    scale: 1,
    transition: { type: "spring", stiffness: 100, damping: 20 }
  }
}

const rightCardVariants = {
  hidden: { opacity: 0, y: "calc(-50% - 65.5px + 150px)", x: "calc(-50% + 101px)", rotate: 0, scale: 0.8 },
  visible: { 
    opacity: 1, 
    y: "calc(-50% - 65.5px)", 
    x: "calc(-50% + 101px)",
    rotate: 0,
    scale: 1,
    transition: { type: "spring", stiffness: 100, damping: 20 }
  }
}

const centerCardVariants = {
  hidden: { opacity: 0, y: "calc(-50% + 100px)", x: "-50%", rotate: 0, scale: 0.9 },
  visible: { 
    opacity: 1, 
    y: "-50%", 
    x: "-50%",
    rotate: 0,
    scale: 1,
    transition: { type: "spring", stiffness: 120, damping: 20 }
  }
}

export function PhysicalCards() {
  return (
    <section className="section-physical-cards content-section">
      <motion.div 
        initial={{ opacity: 0, y: 40 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      >
        <h2 className="text-h1 cards-heading">Virtual and physical cards</h2>
      </motion.div>

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
