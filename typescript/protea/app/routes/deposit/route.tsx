import type { MetaFunction } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Alert, AlertBody, Icon, Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Deposit',
      back: '/'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Deposit'
  }
])

export default function Page() {
  return (
    <Alert role='status'>
      <Icon>error</Icon>
      <AlertBody>Deposits are temporarily unavailable.</AlertBody>
    </Alert>
  )
}
