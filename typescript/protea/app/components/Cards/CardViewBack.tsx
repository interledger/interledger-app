import clsx from 'clsx'
import type { ComponentProps } from 'react'
import { Icon } from '~/components/Icon'
import { formatPan } from '~/lib/cards/useCardsStore'
import { CardViewContainer } from './CardViewContainer'

interface CardViewBackProps extends ComponentProps<'div'> {
  fullCardNumber: string
  expiryDate: string | null
  cvv: string | null
  className?: string
}

export const CardViewBack = ({
  fullCardNumber,
  expiryDate,
  cvv,
  className,
  ...props
}: CardViewBackProps) => {
  const displayExpiryDate = expiryDate
    ? expiryDate.slice(0, 2) + '/' + expiryDate.slice(2)
    : '**/**'
  const displayCvv = cvv || '***'

  return (
    <CardViewContainer className={clsx(className, 'px-5 pt-2')} {...props}>
      <div className='flex h-full flex-col'>
        <p className='text-[6px] font-semibold uppercase leading-none text-black'>
          interledger.org
        </p>

        {/* Dark teal stripe */}
        <div className='-mx-5 mt-1 h-12 bg-teal-deep' />

        {/* Card details */}
        <div className='ml-2 mt-6 space-y-4'>
          <div>
            <div className='flex items-center gap-x-3'>
              <p className='tracking-wide text-charcoal'>
                {formatPan(fullCardNumber)}
              </p>
              <button
                className='h-6 w-6 p-0 text-black/50 hover:text-black'
                onClick={() => navigator.clipboard.writeText(fullCardNumber)}
              >
                <Icon>content_copy</Icon>
              </button>
            </div>
          </div>
          <div className='flex items-center text-charcoal'>
            <div className='flex items-baseline space-x-1'>
              <p className='text-xs uppercase'>Exp</p>
              <p className='leading-none'>{displayExpiryDate}</p>
            </div>
            <div className='ml-6 flex items-baseline space-x-1'>
              <p className='text-xs uppercase'>Cvv</p>
              <p className='leading-none'>{displayCvv}</p>
            </div>
            <button
              className='ml-3 self-center p-0 text-charcoal/50 hover:text-charcoal'
              onClick={() => cvv && navigator.clipboard.writeText(cvv)}
            >
              <Icon>content_copy</Icon>
            </button>
          </div>
        </div>
      </div>
    </CardViewContainer>
  )
}
