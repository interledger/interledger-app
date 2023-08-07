import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  Icon,
  Layouts,
  Snackbar,
  WalletShapes
} from '~/components'
import { Label } from '~/components/Label'
import { getUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)

  return json({
    traits: session.identity.traits
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/settings'),
      title: 'Contact information',
      actions: <WalletShapes />
    },
    isNested: true
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Contact information'
  }
}

export default function Page() {
  const { traits } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent className='flex flex-col space-y-4'>
          <div className='flex w-full flex-col space-y-1'>
            <Label>Email address</Label>
            <div className='mt-1 flex w-full justify-between p-3'>
              <div className='flex space-x-3'>
                <Icon>mail</Icon>
                <span>{traits.email}</span>
              </div>
            </div>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <Label>Mobile phone number</Label>
            <div className='mt-1 flex w-full justify-between p-3'>
              <div className='flex space-x-3'>
                <Icon>call</Icon>
                <span>{traits.phone}</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
