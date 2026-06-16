import { useInView } from 'framer-motion'
import { useRef } from 'react'
import centerCardUrl from '../assets/cards-section/center.svg'
import leftCardUrl from '../assets/cards-section/left.svg'
import rightCardUrl from '../assets/cards-section/right.svg'

export function PhysicalCards() {
  const ref = useRef<HTMLDivElement>(null)
  const isInView = useInView(ref, { once: false, amount: 0.3 })

  return (
    <section className='section-physical-cards content-section'>
      <div className='cards-heading-container'>
        <h2 className='text-h2 cards-heading section-heading'>
          Virtual and physical cards
        </h2>
      </div>

      <div
        ref={ref}
        className={`cards-fan-container anim-cards-fan ${
          isInView ? 'is-visible' : ''
        }`}
      >
        <div className='cards-cluster'>
          {/* Left Card */}
          <div className='physical-card-layer'>
            <img
              src={leftCardUrl}
              alt='Left Debit Card'
              className='physical-card-img physical-card-img--left'
              loading='lazy'
              decoding='async'
            />
          </div>

          {/* Right Card */}
          <div className='physical-card-layer'>
            <img
              src={rightCardUrl}
              alt='Right Debit Card'
              className='physical-card-img physical-card-img--right'
              loading='lazy'
              decoding='async'
            />
          </div>

          {/* Center Card */}
          <div className='physical-card-layer'>
            <img
              src={centerCardUrl}
              alt='Center Debit Card'
              className='physical-card-img physical-card-img--center'
              loading='lazy'
              decoding='async'
            />
          </div>
        </div>
      </div>
    </section>
  )
}
