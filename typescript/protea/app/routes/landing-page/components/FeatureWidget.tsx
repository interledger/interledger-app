import { motion } from "framer-motion"

interface FeatureWidgetProps {
  avatar: string        // initials (e.g. "MH")
  avatarColor?: string  // background hex
  name: string
  amount: string
  amountColor?: string  // defaults to green
  note?: string
  timestamp?: string
}

/**
 * Transaction card widget used in feature sections.
 * Matches the design reference card: avatar | name + amount | note.
 */
export function FeatureWidget({
  avatar,
  avatarColor = "#e87a7a",
  name,
  amount,
  amountColor = "var(--color-interactive-primary, #22c55e)",
  note,
  timestamp,
}: FeatureWidgetProps) {
  return (
    <motion.div 
      className="feature-widget"
      initial={{ opacity: 0, scale: 0.8 }}
      animate={{ 
        opacity: 1, 
        scale: 1,
        y: [0, 20],
      }}
      transition={{
        // Enter animation: Physics Spring as specified
        opacity: { 
          type: "spring", 
          stiffness: 400, 
          damping: 58, 
          mass: 1,
          delay: 0.8 
        },
        scale: { 
          type: "spring", 
          stiffness: 400, 
          damping: 58, 
          mass: 1,
          delay: 0.8 
        },
        // Loop animation: Time Ease as specified
        y: {
          duration: 2,
          ease: [0.54, 0.01, 0.51, 1],
          repeat: Infinity,
          repeatType: "mirror"
        }
      }}
    >
      <div className="feature-widget__avatar" style={{ background: avatarColor }}>
        {avatar}
      </div>
      <div className="feature-widget__body">
        <div className="feature-widget__row">
          <span className="feature-widget__name">{name}</span>
          <span className="feature-widget__amount" style={{ color: amountColor }}>
            {amount}
          </span>
        </div>
        {note && <p className="feature-widget__note">{note}</p>}
        {timestamp && <p className="feature-widget__timestamp">{timestamp}</p>}
      </div>
    </motion.div>
  )
}
