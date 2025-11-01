import clsx from 'clsx'
import { AnimatePresence } from 'framer-motion'
import type { ComponentProps } from 'react'
import { Fade } from '../Animations/Fade'
import { MasterCardLogo } from '../Logos'
import { CardViewContainer } from './CardViewContainer'

interface CardViewFrontProps extends ComponentProps<'div'> {
  nameOnCard: string
}

export const CardViewFront = ({
  nameOnCard,
  className,
  ...props
}: CardViewFrontProps) => {
  return (
    <CardViewContainer className={clsx(className, 'px-5 pt-4')} {...props}>
      <div className='flex h-full flex-col'>
        {/* Header with logos */}
        <div className='flex items-center justify-between text-sm'>
          <div className='flex items-center'>
            {/* Interledger Foundation SVG Logo */}
            <img
              src='/interledger-card-logo.svg'
              alt='Interledger Logo'
              className='h-8 max-w-[120px] object-contain opacity-90'
              onError={(e) => {
                const target = e.target as HTMLImageElement
                target.style.display = 'none'
                const fallback = target.nextElementSibling as HTMLElement
                if (fallback) fallback.classList.remove('hidden')
              }}
            />
          </div>

          <div className='mt-1 flex items-center justify-end'>
            <span className='text-xs tracking-wider text-charcoal'>debit</span>
          </div>
        </div>

        {/* EMV Chip */}
        <div className='ml-3.5 mt-4'>
          <div className='w-fit'>
            <img
              src='/emv_chip.svg'
              alt='EMV chip'
              className='h-9 w-12 object-contain'
            />
          </div>
        </div>

        <div className='-mr-4 mt-auto flex items-center justify-between'>
          <AnimatePresence mode='wait'>
            <Fade className='mt-2' nonce={nameOnCard}>
              <span className='text-xs font-medium uppercase text-black'>
                {nameOnCard}
              </span>
            </Fade>
          </AnimatePresence>
          <MasterCardLogo />
        </div>
      </div>
    </CardViewContainer>
  )
}
