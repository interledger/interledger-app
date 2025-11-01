import { motion } from 'framer-motion'
import type { ReactNode } from 'react'

interface FadeProps {
  nonce: string
  children: ReactNode
  className?: string
}

export const Fade = ({ nonce, className = '', children }: FadeProps) => {
  return (
    <motion.div
      key={'card-' + nonce}
      animate={{ opacity: 1, scale: 1 }}
      initial={{ opacity: 0, scale: 0.5 }}
      exit={{
        opacity: 0,
        scale: 0.5,
        transition: {
          duration: 0.2
        }
      }}
      transition={{
        type: 'spring',
        stiffness: 400,
        damping: 20
      }}
      className={className}
    >
      {children}
    </motion.div>
  )
}
