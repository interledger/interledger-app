import { motion } from "framer-motion"
import logoImg from "../assets/interledger-wallet-logo.png"

/**
 * Renders a single CSS-based debit card replica
 */
function Card({ className }: { className: string }) {
  return (
    <div className={`physical-card ${className}`}>
      <div className="card-top">
        <div className="card-chip" />
        <span className="card-debit">Debit</span>
      </div>
      
      <div className="card-logo">Interledger</div>
      
      <div className="card-bottom">
        <div className="mastercard-circles">
          <div className="mc-circle mc-circle--red" />
          <div className="mc-circle mc-circle--yellow" />
        </div>
      </div>
    </div>
  )
}

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
const cardVariants = {
  hidden: { opacity: 0, y: 120, scale: 0.9 },
  visible: { 
    opacity: 1, 
    y: 0, 
    scale: 1,
    transition: { type: "spring", stiffness: 120, damping: 20 }
  }
}

export function PhysicalCards() {
  return (
    <section className="section-physical-cards">
      <motion.div 
        initial={{ opacity: 0, y: 40 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.3 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      >
        <h2 className="text-h1 cards-heading">Virtual and physical cards</h2>
      </motion.div>

      {/* Wrapping the fan container allows the CSS to apply rotate based on desktop media queries immediately when visibly triggered. 
          Actually, we use CSS classes for the final rotated states, but motion will spring-animate the 'y' and 'scale' upwards smoothly. 
          To make CSS pick it up, we just apply 'is-visible' based on whileInView automatically by using whileInView="visible" AND having a class that reacts to it. But since we use variants, we will just use motion natively.
      */}
      <motion.div 
        className="cards-fan-container anim-cards-fan"
        variants={fanVariants}
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true, amount: 0.3 }}
      >
        {/* We keep the physical-card--X class for the rotation math, but the spring entry comes from motion's childVariants */}
        <motion.div variants={cardVariants} className="physical-card-wrapper">
          <Card className="physical-card--left" />
        </motion.div>
        
        <motion.div variants={cardVariants} className="physical-card-wrapper">
          <Card className="physical-card--right" />
        </motion.div>

        <motion.div variants={cardVariants} className="physical-card-wrapper">
          <Card className="physical-card--center" />
        </motion.div>
      </motion.div>
    </section>
  )
}
