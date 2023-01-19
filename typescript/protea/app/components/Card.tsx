import type { FC, ReactNode } from 'react'
import clsx from 'clsx'

type CardProps = {
  children?: ReactNode
  className?: string
}

const CardRoot: FC<CardProps> = ({ children, className }) => {
  return (
    <div
      className={clsx(
        'flex w-full flex-col rounded-2xl bg-page p-4 pb-6',
        className
      )}
    >
      {children}
    </div>
  )
}

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

export const Card = Object.assign(CardRoot, { Item })
