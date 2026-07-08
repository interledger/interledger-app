import { motion } from 'framer-motion'
import type { ReactNode } from 'react'
import type { CarouselScreen } from '../context/PhoneCarouselContext'
import { usePhoneCarousel } from '../context/PhoneCarouselContext'

interface FeatureSectionProps {
  screen: CarouselScreen
  heading: string
  body?: string
  widget?: ReactNode
  visual?: ReactNode
  columnOrder?: 'text-left' | 'text-right'
}

const containerVariants = {
  hidden: {
    opacity: 0,
    transition: { type: 'spring', duration: 0.8, bounce: 0 }
  },
  visible: {
    opacity: 1,
    transition: {
      type: 'spring',
      duration: 1,
      bounce: 0,
      staggerChildren: 0.1
    }
  }
}

const childVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      type: 'spring',
      duration: 1,
      bounce: 0
    }
  }
}

/**
 * Absolute overlay panel inside .feature-content-slot.
 * Rendered for every feature; only the panel whose `screen` matches
 */
export function FeatureSection({
  screen,
  heading,
  body,
  widget,
  visual
}: FeatureSectionProps) {
  const { activeScreen } = usePhoneCarousel()
  const isActive = activeScreen === screen
  const isFirst = screen === 2

  const textCol = (
    <motion.div className='feature-col feature-col--left'>
      <motion.h2 variants={childVariants} className='text-h2 feature-heading'>
        {heading}
      </motion.h2>
    </motion.div>
  )

  const rightCol = (
    <motion.div className='feature-col feature-col--right'>
      {body && (
        <motion.p variants={childVariants} className='text-h4 feature-body'>
          {body}
        </motion.p>
      )}
      {widget && (
        <motion.div variants={childVariants} className='feature-col__widget'>
          {widget}
        </motion.div>
      )}
      {visual && <motion.div variants={childVariants}>{visual}</motion.div>}
    </motion.div>
  )

  return (
    <motion.div
      className={`feature-panel ${isActive ? 'is-visible' : ''} ${
        isFirst ? 'feature-panel--first' : ''
      }`}
      variants={containerVariants}
      initial='hidden'
      animate={isActive ? 'visible' : 'hidden'}
      style={{ pointerEvents: isActive ? 'auto' : 'none' }}
    >
      {textCol}
      <div className='feature-phone-spacer' />
      {rightCol}
    </motion.div>
  )
}
