import { forwardRef } from 'react'
import type { ReactNode } from 'react'
import { Link, useNavigate } from '@remix-run/react'

/**
 * Ensure all routes are prefixed with `/` to prevent relative routing.
 */

type RouterProps = {
  className?: string
  to: string
  children?: ReactNode
}

/**
 * Exposes the remix Link as a styled version.
 *
 * @param children The children of the Link.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
const RouterRoot = forwardRef<any, RouterProps>(
  ({ className, children, to, ...rest }, ref) => {
    return (
      <Link
        ref={ref}
        to={to}
        className={`focus-visible:outline-2 focus-visible:outline-focus ${className}`}
        {...rest}
      >
        {children}
      </Link>
    )
  }
)

RouterRoot.displayName = 'Link Router'

/**
 * Exposes a headless button that will route to `to`,
 * and can be wrapped around a styled button for accessibility.
 *
 * @param children The children of the button.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
const Button = forwardRef<any, RouterProps>(
  ({ className, children, to, ...rest }, ref) => {
    const navigate = useNavigate()
    return (
      <button
        ref={ref}
        onClick={() => navigate(to)}
        className={`focus-visible:outline-2 focus-visible:outline-focus ${className}`}
        {...rest}
      >
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button Router'

/**
 * Exposes a headless anchor tag that will route to `to`,
 * and can be wrapped around a styled button for accessibility.
 *
 * @param children The children of the button.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
const a = forwardRef<any, RouterProps>(
  ({ className, children, to, ...rest }, ref) => {
    return (
      <a
        ref={ref}
        href={to}
        className={`focus-visible:outline-2 focus-visible:outline-focus ${className}`}
        {...rest}
      >
        {children}
      </a>
    )
  }
)

a.displayName = 'Anchor Router'

export const Router = Object.assign(RouterRoot, { Button, a })
