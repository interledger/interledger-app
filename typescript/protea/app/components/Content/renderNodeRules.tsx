import { isCode, isLink } from 'datocms-structured-text-utils'
import { renderNodeRule } from 'react-datocms'
import { AnchorRouter, Router } from '~/components'
import { sanitizeCMSLinks } from './sanitize'

/**
 * Implements all the custom node types
 * https://github.com/datocms/structured-text/tree/main/packages/utils#typescript-type-guards
 * TODO: add custom mark rules and fix how <Prose> works
 */

export const renderCodeNodeRule = renderNodeRule(isCode, ({ node, key }) => {
  return <div key={key} dangerouslySetInnerHTML={{ __html: node.code }} />
})

export const renderLinkNodeRule = renderNodeRule(
  isLink,
  ({ node, children }) => {
    const target = node.meta?.find((meta) => meta.id === 'target')?.value
    const { toUrl, internal } = sanitizeCMSLinks(node.url as string)

    if (internal) {
      return (
        <Router key={node.url} to={toUrl} className={'text-primary'}>
          {children}
        </Router>
      )
    }

    return (
      <AnchorRouter
        key={node.url}
        to={toUrl}
        className={'text-primary'}
        target={target}
      >
        {children}
      </AnchorRouter>
    )
  }
)
