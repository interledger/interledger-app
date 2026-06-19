import { useInView } from 'framer-motion'
import { useRef } from 'react'
import handHoldingPhone from '../assets/send-receive-section/hand-hodling-phone.svg'
import privateIcon from '../assets/send-receive-section/icon-private.svg'
import secureIcon from '../assets/send-receive-section/icon-secure.svg'
import sendReceiveIcon from '../assets/send-receive-section/icon-sendreceive.svg'
import { PageSection } from '../components/PageSection'

function StatusItem({ icon, text }: { icon: React.ReactNode; text: string }) {
  return (
    <li className='status-item'>
      <div className='status-icon-container'>{icon}</div>
      <span className='text-h5 status-text'>{text}</span>
      <div className='status-check-container'>
        <div className='checkmark-css' />
      </div>
    </li>
  )
}

function HandAnimation() {
  const ref = useRef<HTMLDivElement>(null)
  const isInView = useInView(ref, { once: false, margin: '-100px' })

  return (
    <div
      ref={ref}
      className={`hand-visual-container ${isInView ? 'is-visible' : ''}`}
    >
      <div className='hand-phone-frame'>
        <img
          src={handHoldingPhone}
          alt='Hand holding phone'
          className='hand-svg-asset'
          loading='lazy'
          decoding='async'
        />

        <div className='currency-circle circle-dollar'>
          <span className='currency-char'>$</span>
        </div>

        <div className='currency-circle circle-pound'>
          <span className='currency-char'>£</span>
        </div>

        <div className='currency-circle circle-euro'>
          <span className='currency-char'>€</span>
        </div>
      </div>
    </div>
  )
}

export function SendReceive() {
  const ref = useRef<HTMLUListElement>(null)
  const isInView = useInView(ref, { once: true, margin: '-50px' })
  return (
    <PageSection className='send-receive-section'>
      <div className='send-receive-container'>
        <header className='send-receive-header'>
          <div className='text-h2 send-receive-title section-heading'>
            Today, the wallet lets you send and receive money
          </div>
          <p className='text-h4 send-receive-description'>
            Starting with the fundamentals keeps the wallet reliable, while
            allowing it to scale responsibly. Our aim is to remain true to the
            protocol it supports over the long run.
          </p>
        </header>

        <div className='send-receive-content'>
          <ul
            ref={ref}
            className={`send-receive-list ${isInView ? 'is-visible' : ''}`}
          >
            <StatusItem
              icon={
                <img src={secureIcon} alt='' loading='lazy' decoding='async' />
              }
              text='Secure by design'
            />
            <StatusItem
              icon={
                <img
                  src={sendReceiveIcon}
                  alt=''
                  loading='lazy'
                  decoding='async'
                />
              }
              text='Send & receive money'
            />
            <StatusItem
              icon={
                <img src={privateIcon} alt='' loading='lazy' decoding='async' />
              }
              text='Private by default'
            />
          </ul>

          <div className='send-receive-visual'>
            <HandAnimation />
          </div>
        </div>
      </div>
    </PageSection>
  )
}
