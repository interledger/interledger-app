import { forwardRef } from 'react'
import { CardIcon, Icon } from './'
import clsx from 'clsx'

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
                    <p className='text-sm text-medium text-justify p-4'>{text || ""}</p>
                </div>
            </span>
        )
    }
)

WarningBanner.displayName = 'WarningBanner'
