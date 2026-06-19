import clsx from 'clsx'
import {
  AnchorRouter,
  ButtonRouter,
  OutlineButtonRouter,
  Router
} from '~/components'
import type { LinkRecord } from '~/generated/dato-cms-graphql'
import { sanitizeCMSLinks } from './sanitize'

type ContentRouterProps = {
  className?: string
  to: LinkRecord
  shrink?: boolean
  outline?: boolean
}

export function ContentRouter({
  to,
  shrink,
  outline,
  className
}: ContentRouterProps) {
  const target = to.target ? '_blank' : '_self'
  const { toUrl, internal } = sanitizeCMSLinks(to.url as string)
  if (to.button) {
    if (outline) {
      return (
        <OutlineButtonRouter
          to={toUrl}
          shrink={shrink}
          className={clsx('h-20 px-20', className)}
        >
          {to.displayText}
        </OutlineButtonRouter>
      )
    } else
      return (
        <ButtonRouter
          to={toUrl}
          shrink={shrink}
          className={clsx('h-20 px-20', className)}
        >
          {to.displayText}
        </ButtonRouter>
      )
  } else {
    if (internal) {
      return (
        <Router to={toUrl} className={className}>
          {to.displayText}
        </Router>
      )
    }
    return (
      <AnchorRouter to={toUrl} className={className} target={target}>
        {to.displayText}
      </AnchorRouter>
    )
  }
}
