import { NavLink } from 'react-router';
import type { NavLinkProps } from 'react-router';
import clsx from 'clsx'
import type { ButtonHTMLAttributes, HTMLAttributes, RefAttributes } from 'react'
import { forwardRef, useEffect, useState } from 'react'
import { Icon } from '~/components/Icon'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx(
          'flex w-full flex-col rounded-[1.25rem] bg-container-strong p-2',
          className
        )}
        {...props}
      />
    )
  }
)
Card.displayName = 'Card'

const CardHeader = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={clsx('flex w-full justify-between px-2 pb-0 pt-2', className)}
      {...props}
    />
  )
)
CardHeader.displayName = 'CardHeader'

const CardTitle = forwardRef<
  HTMLHeadingElement,
  HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => {
  return (
    <h2
      ref={ref}
      className={clsx('text-lg font-medium text-strong', className)}
      {...props}
    />
  )
})
CardTitle.displayName = 'CardTitle'

const CardIcon = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={clsx(
        'flex items-center justify-between rounded-full bg-nav p-5 text-medium',
        className
      )}
      {...props}
    />
  )
)
CardIcon.displayName = 'CardIcon'

const CardContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={clsx('p-2', className)} {...props} />
  )
)
CardContent.displayName = 'CardContent'

const CardLink = forwardRef<
  any,
  NavLinkProps & RefAttributes<HTMLAnchorElement>
>(({ children, className, ...props }, ref) => {
  return (
    <NavLink
      ref={ref}
      className={({ isActive }) =>
        clsx(
          'my-1 flex rounded-xl p-3 first:mt-0 last-of-type:mb-0 focus-visible:outline-2 focus-visible:outline-focus',
          isActive ? 'bg-nav-hover' : 'hover:bg-nav',
          className
        )
      }
      {...props}
    >
      {children}
    </NavLink>
  )
})
CardLink.displayName = 'CardLink'

interface CardButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  noHover?: boolean
}

const CardButton = forwardRef<HTMLButtonElement, CardButtonProps>(
  ({ className, noHover, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={clsx(
          noHover && 'bg-nav',
          'my-1 flex rounded-xl p-3 first:mt-0 last-of-type:mb-0 hover:bg-nav focus-visible:outline-2 focus-visible:outline-focus active:bg-nav-hover',
          className
        )}
        {...props}
      />
    )
  }
)
CardButton.displayName = 'CardButton'

interface CardCopyProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  shareData: ShareData
  copyContent: string
  success: string
  copyError: string
  shareError: string
}

const CardCopy = forwardRef<HTMLButtonElement, CardCopyProps>(
  (
    {
      className,
      shareData,
      copyContent,
      success,
      copyError,
      shareError,
      ...props
    },
    ref
  ) => {
    const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

    const [features, setFeatures] = useState({
      clipboard: false,
      share: false
    })

    useEffect(() => {
      if (
        typeof navigator !== 'undefined' &&
        typeof navigator.clipboard !== 'undefined'
      ) {
        setFeatures((prev) => ({ ...prev, clipboard: true }))
      }
    }, [])

    return (
      <button
        ref={ref}
        onClick={() => {
          if (features.share) {
            navigator.share(shareData).then(
              () => {},
              () => {
                pushSnackbar({
                  id: `share-fail-${shareData.text}`,
                  message: shareError,
                  icon: 'close',
                  canShow: true
                })
              }
            )
          } else if (features.clipboard) {
            navigator.clipboard.writeText(copyContent).then(
              () => {
                pushSnackbar({
                  id: `clipboard-success-${copyContent}`,
                  message: success,
                  icon: 'close',
                  canShow: true
                })
              },
              () => {
                pushSnackbar({
                  id: `clipboard-fail-${copyContent}`,
                  message: copyError,
                  icon: 'close',
                  canShow: true
                })
              }
            )
          }
        }}
        className={clsx(
          'my-1 flex items-center justify-between rounded-xl bg-nav p-3 first:mt-0 last-of-type:mb-0 focus-visible:outline-2 focus-visible:outline-focus active:bg-nav-hover',
          className
        )}
        {...props}
      >
        <span className='truncate text-left font-medium text-medium'>
          {props.children}
        </span>
        <Icon className='text-medium'>
          {features.share ? 'share' : features.clipboard ? 'content_copy' : ''}
        </Icon>
      </button>
    )
  }
)
CardCopy.displayName = 'CardCopy'

export {
  Card,
  CardButton,
  CardContent,
  CardCopy,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle
}
