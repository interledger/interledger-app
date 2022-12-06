import type { FC } from 'react'
import clsx from 'clsx'

type IconProps = {
  className?: string
}

export const Icon: FC<IconProps> = ({ className, children }) => {
  return (
    <span
      className={clsx(
        'inline-block h-6 w-6 select-none whitespace-nowrap fill-current text-center font-icon text-2xl font-normal normal-case not-italic leading-6 tracking-normal antialiased',
        className
      )}
    >
      {children}
    </span>
  )
}
