import { motion } from 'framer-motion'
import { useState } from 'react'

interface FeatureWidgetProps {
  avatar: string
  avatarColor?: string
  name: string
  amount: string
  note?: string
  timestamp?: string
}

export function FeatureWidget({
  avatar,
  avatarColor = '#e87a7a',
  name,
  amount,
  note,
  timestamp
}: FeatureWidgetProps) {
  const [isVisible, setIsVisible] = useState(false)

  return (
    <div style={{ width: '100%' }}>
      <div className={`feature-widget-bob ${isVisible ? 'is-visible' : ''}`}>
        <motion.div
          onViewportEnter={() => setIsVisible(true)}
          onViewportLeave={() => setIsVisible(false)}
          viewport={{ once: false, amount: 0.5 }}
        >
          <div className='feature-widget'>
            <div
              className='feature-widget__avatar'
              style={{ background: avatarColor }}
            >
              {avatar}
            </div>
            <div className='feature-widget__body'>
              <div className='feature-widget__row'>
                <span className='feature-widget__name'>{name}</span>
                <span className='feature-widget__amount'>{amount}</span>
              </div>
              <div className='feature-widget__row'>
                {note && <p className='feature-widget__note'>{note}</p>}
                {timestamp && (
                  <p className='feature-widget__timestamp'>{timestamp}</p>
                )}
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  )
}
