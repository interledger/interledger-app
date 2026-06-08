import { forwardRef } from 'react'
import type { ReactNode } from 'react'
import { Link } from 'react-router'
import clsx from 'clsx'

/**
 * TODO: Router refactor:
 * - Router - this is a headless routing element for accessibility.
 * - TextRouter - a router to be used for inline text that styles the text blue..
 * - ButtonRouter - a routing element that is styled like a button.
 * - IconRouter - a router that houses an Icon - should just take text of the icon as children.
 * - AnchorRouter - a headless <a> for routing externally, should target_blank as default?
 * - ActionButton -
 * - FloatingActionButton
 */

/**
 * Potentially type to as a route instead of string so it can't be used without typing
 */

type RouterProps = {
  className?: string
  to: string
  children?: ReactNode
  //Can only disable Router.Button
  disabled?: boolean
}

/**
 * Exposes the remix Link as a styled version.
 *
 * @param children The children of the Link.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
export const Router = forwardRef<any, RouterProps>(
  ({ className, children, to, ...rest }, ref) => {
    return (
      <Link
        ref={ref}
        to={to}
        className={clsx(
          'focus-visible:outline-2 focus-visible:outline-focus',
          className
        )}
        {...rest}
      >
        {children}
      </Link>
    )
  }
)

Router.displayName = 'Router'

type ButtonRouterProps = {
  className?: string
  to: string
  children?: ReactNode
  shrink?: boolean
}

/**
 * Exposes a headless button that will route to `to`,
 * and can be wrapped around a styled button for accessibility.
 *
 * @param children The children of the button.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
export const ButtonRouter = forwardRef<any, ButtonRouterProps>(
  ({ className, children, to, shrink, ...rest }, ref) => {
    return (
      <Link
        ref={ref}
        to={to}
        className={clsx(
          'flex h-12 w-full items-center justify-center rounded-full border border-transparent bg-primary px-10 font-display font-medium text-white hover:bg-blue-400 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-blue-500',
          shrink ? 'sm:max-w-fit' : '',
          className
        )}
        {...rest}
      >
        {children}
      </Link>
    )
  }
)

ButtonRouter.displayName = 'ButtonRouter'

/**
 * Exposes a headless anchor tag that will route to `to`,
 * and can be wrapped around a styled button for accessibility.
 *
 * @param children The children of the button.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
export const AnchorRouter = forwardRef<any, RouterProps>(
  ({ className, children, to, ...rest }, ref) => {
    return (
      <a
        ref={ref}
        href={to}
        target='_blank'
        rel='noreferrer'
        className={clsx(
          'focus-visible:outline-2 focus-visible:outline-focus',
          className
        )}
        {...rest}
      >
        {children}
      </a>
    )
  }
)

AnchorRouter.displayName = 'AnchorRouter'
