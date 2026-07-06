import { useInView } from 'framer-motion'
import { useRef } from 'react'
import { Link } from 'react-router'
import appStoreBadgeIcon from '../assets/app-store-badge-icon.svg'
import googlePlayBadgeIcon from '../assets/google-play-badge-icon.svg'
import { PageSection } from '../components/PageSection'

export function OtherFeatures() {
  const ref = useRef<HTMLDivElement>(null)
  const isInView = useInView(ref, { once: false, amount: 0.3 })

  return (
    <PageSection>
      <div
        ref={ref}
        className={`other-features-banner ${isInView ? 'is-visible' : ''}`}
      >
        <header className='other-features-header'>
          <div className='other-features-left'>
            <h2 className='text-h2 other-features-headline-stack'>
              <span className='other-features-headline'>Open standards.</span>
              <span className='other-features-headline'>
                Solid foundations.
              </span>
              <span className='other-features-headline'>
                Intentionally simple.
              </span>
            </h2>
          </div>
          <div className='other-features-right'>
            <p className='other-features-description'>
              This wallet is designed for a world where money moves across
              borders, systems, and contexts.
            </p>
            <p className='other-features-description'>
              It works wherever the web works, without assuming who you are,
              where you live, or how you participate.
            </p>
          </div>
        </header>

        <div className='other-features-cta-block'>
          <Link to='/signup' className='other-features-btn'>
            Get the Interledger Wallet
            <svg
              width='20'
              height='20'
              viewBox='0 0 24 24'
              fill='none'
              stroke='currentColor'
              strokeWidth='2'
              strokeLinecap='round'
              strokeLinejoin='round'
            >
              <path d='M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4' />
              <polyline points='7 10 12 15 17 10' />
              <line x1='12' y1='15' x2='12' y2='3' />
            </svg>
          </Link>

          <div className='other-features-badges'>
            {/* TODO: point to the App Store listing when available. */}
            <Link
              to='/signup'
              className='app-badge app-badge--ios'
              aria-label='Download on the App Store'
            >
              <span className='app-badge__icon' aria-hidden='true'>
                <img
                  src={appStoreBadgeIcon}
                  alt=''
                  width='24'
                  height='24'
                  loading='lazy'
                  decoding='async'
                />
              </span>
              <div className='app-badge__text'>
                <span className='app-badge__sub'>Download on the</span>
                <span className='app-badge__main'>App Store</span>
              </div>
            </Link>
            {/* TODO: point to the Google Play listing when available. */}
            <Link
              to='/signup'
              className='app-badge app-badge--android'
              aria-label='Get it on Google Play'
            >
              <span className='app-badge__icon' aria-hidden='true'>
                <img
                  src={googlePlayBadgeIcon}
                  alt=''
                  width='24'
                  height='24'
                  loading='lazy'
                  decoding='async'
                />
              </span>
              <div className='app-badge__text'>
                <span className='app-badge__sub'>GET IT ON</span>
                <span className='app-badge__main'>Google Play</span>
              </div>
            </Link>
          </div>

          <p className='text-h4 browser-availability'>
            Also available in browsers
          </p>
        </div>
      </div>
    </PageSection>
  )
}
