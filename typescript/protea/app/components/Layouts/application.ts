import type { RouteMatch } from '@remix-run/react'
import type { ReactNode } from 'react'

export type ApplicationProps = {
  title?: string | ((match: RouteMatch) => string) // TODO deprecate once all routes are updated with scaffold
  layout: Layouts | ((match: RouteMatch) => Layouts)
  scaffold?: ScaffoldProps
}

export enum Layouts {
  Focus = 'Focus',
  Docs = 'Docs',
  Wallet = 'Wallet',
  Marketing = 'Marketing'
}

export enum Fab {
  Pay = 'Pay',
  Identity = 'Identity',
  Account = 'Account'
}

export type ScaffoldHeaderActions = {
  type: 'search' | 'chip' | 'shapes'
  content?: (match: RouteMatch) => ReactNode
}

//github.com/datocms/react-datocms/blob/master/docs/structured-text.md#override-default-rendering-of-nodes
/**
 * ScaffoldProps
 * @property header - Scaffold header props
 * @property header.back - Scaffold Back button route
 * @property header.title - Scaffold header title
 * @property header.actions - Scaffold header actions
 * @property hasFab - Scaffold has floating action button
 * @property hasNavDrawer - Scaffold has navigation drawer
 * @property isNested - Is the current route nested in a parent route?
 */
export type ScaffoldProps = {
  header: {
    // Back should check the history stack, and if the previous route is the same as the specified route, it should pop the history stack
    back?: string | ((match: RouteMatch) => string)
    title?: string | ((match: RouteMatch) => string)
    actions?: ScaffoldHeaderActions[] // TODO: use a better type here, this is too generic
  }
  fab?: Fab
  isNested?: boolean
}
