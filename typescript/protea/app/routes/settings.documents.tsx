import { href, Navigate } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardHeader,
  CardLink,
  CardTitle,
  Icon,
  Layouts
} from '~/components'
import { mergeMeta } from '~/lib/meta'
import { useSettingsContext } from './settings'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Documents'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [{ title: 'Documents' }])

export default function Page() {
  const { isDocumentsEnabled } = useSettingsContext()

  if (!isDocumentsEnabled) return <Navigate to={href('/settings')} replace />

  return (
    <Card>
      <CardHeader>
        <CardTitle>Confirmations</CardTitle>
      </CardHeader>
      <CardLink
        end
        preventScrollReset
        reloadDocument
        target='_blank'
        to={href('/api/statements/accountConfirmation')}
      >
        <div className='mr-auto flex space-x-3'>
          <Icon>description</Icon>
          <span>Account confirmation</span>
        </div>
        <Icon>open_in_new</Icon>
      </CardLink>
    </Card>
  )
}
