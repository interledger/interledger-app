import type { ComponentProps } from 'react'

interface MasterCardLogoProps extends ComponentProps<'div'> {
  size?: 'sm' | 'md'
}

export const MasterCardLogo = ({
  size = 'md',
  className = '',
  ...props
}: MasterCardLogoProps) => {
  const sizeClasses = size === 'sm' ? 'h-6 w-6' : 'h-8 w-8'
  const overlapClass = size === 'sm' ? '-ml-2' : '-ml-3'

  return (
    <div className={`flex items-center ${className}`} {...props}>
      <div className={`${sizeClasses} rounded-full bg-red-500`}></div>
      <div
        className={`${overlapClass} ${sizeClasses} rounded-full bg-orange-400`}
      ></div>
    </div>
  )
}
