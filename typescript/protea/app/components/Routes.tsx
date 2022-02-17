import React, { FC } from 'react'
import { Link } from 'remix'

/**
 * Predefined routes that allow for consistent routing.
 * Ensure all routes are prefixed with `/` to prevent relative routing.
 */

export type RouterProps = {
  className?: string
  to: string
}

/**
 * Exposes the remix Link as a styled version.
 *
 * @param children The children of the Link.
 * @param href - Routes or Routes object with params
 * @param rest The props passed through to the anchor tag.
 */
export const Router: FC<RouterProps> = ({
  className,
  children,
  to,
  ...rest
}) => {
  return (
    <Link
      to={to}
      className={`focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-1 ${className}`}
      {...rest}
    >
      {children}
    </Link>
  )
}
