import type { LoaderArgs, V2_MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardContent, Icon, Layouts } from '~/components'
import { Label } from '~/components/Label'
import { getUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  const len = session.identity.traits.phone.length
  const phoneMask = session.identity.traits.phone
    .substring(len - 4, len)
    .padStart(len, '*')

  return json({
    phoneMask,
    email: session.identity.traits.email
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/settings'),
      title: 'Contact information'
    },
    isNested: true
  }
}

export const meta: V2_MetaFunction = mergeMeta(() => [
  {
    title: 'Contact information'
  }
])

export default function Page() {
  const { phoneMask, email } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent className='flex flex-col space-y-4'>
          <div className='flex w-full flex-col space-y-1'>
            <Label>Email address</Label>
            <div className='mt-1 flex w-full justify-between p-3'>
              <div className='flex space-x-3'>
                <Icon>mail</Icon>
                <span>{email}</span>
              </div>
            </div>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <Label>Mobile phone number</Label>
            <div className='mt-1 flex w-full justify-between p-3'>
              <div className='flex space-x-3'>
                <Icon>call</Icon>
                <span>{phoneMask}</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
