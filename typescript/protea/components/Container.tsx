import { FC } from 'react'

type ContainerProps = {
  className?: string
}

export const Container: FC<ContainerProps> = ({ children, className }) => {
  return (
    <div
      className={`flex flex-col max-w-full sm:max-w-[1080px] mx-auto min-h-screen ${className}`}
    >
      {children}
    </div>
  )
}
