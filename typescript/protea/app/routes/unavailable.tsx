import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardIcon,
  GridColumn,
  Icon,
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

export const meta = mergeMeta(() => [
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
        <Card className='flex !flex-row'>
          <CardIcon className='h-16 my-auto'>
            <Icon className='text-red-600'>warning</Icon>
          </CardIcon>
          <CardContent className='text-lg ml-2'>
            The application is not yet available in your location, but do not worry we are working tirelessly to solve it as fast as possible! <br />
            We will notify you by email once the application becomes fully functional in your region.
          </CardContent>
        </Card>
      </GridColumn>
    </WalletGrid>
  )
}
