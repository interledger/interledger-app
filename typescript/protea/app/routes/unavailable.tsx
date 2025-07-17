import type { MetaFunction } from '@remix-run/node'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  GridColumn,
  Layouts,
  WalletGrid
} from '~/components'
import { mergeMeta } from '~/lib/meta'


export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Notice'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Not available'
  }
])

export default function Page() {

  return (
    <WalletGrid>
      <GridColumn
        className='col-span-full'
      >
        <Card>
          <CardContent>
            The app is not yet available in your location, but do not worry we are working tirelessly to solve it asap! <br />
            We will notify you by email once the app becomes fully functional in your region.
          </CardContent>
        </Card>
      </GridColumn>
    </WalletGrid>
  )
}
