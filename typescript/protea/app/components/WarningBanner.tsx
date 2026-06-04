import clsx from 'clsx'
import { forwardRef } from 'react'
import { CardIcon, Icon } from './'

interface WarningBannerProps {
  className?: string
  text?: string
}

export const WarningBanner = forwardRef<any, WarningBannerProps>(
  ({ className, text }, ref) => {
    return (
      <span className={clsx('flex items-center', className)}>
        <CardIcon>
          <Icon className='text-red-600'>warning</Icon>
        </CardIcon>
        <div className='flex flex-col items-start'>
          <p className='p-4 text-justify text-sm text-medium'>{text || ''}</p>
        </div>
      </span>
    )
  }
)

WarningBanner.displayName = 'WarningBanner'
