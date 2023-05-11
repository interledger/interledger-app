import type { FC, ReactNode } from 'react'

type ContainerProps = {
  className?: string
  children?: ReactNode
}

export const Container: FC<ContainerProps> = ({ children, className }) => {
  return (
    <div
      className={`mx-auto flex min-h-screen max-w-full flex-col sm:max-w-[1080px] ${className}`}
    >
      {children}
    </div>
  )
}
