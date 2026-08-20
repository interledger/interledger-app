import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Alert, AlertBody, Icon, Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Sign up', back: href('/') }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Sign up'
  }
])

export default function Page() {
  return (
    <Alert role='status'>
      <Icon>error</Icon>
      <AlertBody>
        New account registrations are temporarily unavailable.
      </AlertBody>
    </Alert>
  )
}
