import clsx from 'clsx'
import type { FC, ReactNode } from 'react'
import { forwardRef } from 'react'

type CardProps = {
  children?: ReactNode
  className?: string
}

const CardRoot = forwardRef<any, CardProps>(({ children, className }, ref) => {
  return (
    <div
      ref={ref}
      className={clsx(
        'flex w-full flex-col rounded-2xl bg-container-strong p-4',
        className
      )}
    >
      {children}
    </div>
  )
})

CardRoot.displayName = 'Card'

type TitleProps = {
  children?: ReactNode
  className?: string
}

const Title = forwardRef<any, TitleProps>(({ children, className }, ref) => {
  return (
    <h2
      ref={ref}
      className={clsx('text-lg font-medium text-strong', className)}
    >
      {children}
    </h2>
  )
})

Title.displayName = 'Title'

type CardItemProps = {
  children?: ReactNode
  className?: string
  variant?: 'row' | 'col'
}

const Item: FC<CardItemProps> = ({ children, className, variant = 'row' }) => {
  const variantStyle =
    variant == 'row' ? 'justify-between' : 'flex-col space-y-1'
  return (
    <div className={clsx('flex w-full', variantStyle, className)}>
      {children}
    </div>
  )
}

export const Card = Object.assign(CardRoot, { Item, Title })
