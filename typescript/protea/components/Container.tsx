import { FC } from 'react'

type ContainerProps = {
  className?: string
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
