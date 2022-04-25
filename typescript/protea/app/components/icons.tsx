import type { FC } from 'react'

type IconProps = {
  className?: string
}

export const Icon: FC<IconProps> = ({ className, children }) => {
  className = [className, 'h-6 w-6 material-symbols-outlined fill-current']
    .filter(Boolean)
    .join(' ')
  return <span className={className}>{children}</span>
}
