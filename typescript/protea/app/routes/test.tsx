import type { ApplicationProps } from '~/components'
import { Card, Fab, Layouts, WalletGrid } from '~/components'

export const handle: ApplicationProps = {
  title: 'Connections',
  layout: Layouts.Marketing,
  scaffold: {
    header: {
      title: 'Connections',
      // TODO Figure out a better way to do this
      actions: [{ type: 'search' }]
    },
    fab: Fab.Pay
  }
}

export default function Page() {
  return (
    <WalletGrid>
      <Card className='col-span-full h-screen'>
        <p className='text-medium'>Add and manage your connections.</p>
      </Card>
    </WalletGrid>
  )
}
