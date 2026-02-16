import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardLink, Icon, Layouts } from '~/components'
import { Label } from '~/components/Label'
import { getUserSession, getSessionTraits } from '~/lib/kratos/session.util'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const session = await getUserSession(request)
  const { phone, email } = getSessionTraits(session)

  const len = phone.length
  const phoneMask = phone
    .substring(len - 4, len)
    .padStart(len, '*')

  return json({
    phoneMask,
    email
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

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Contact information'
  }
])

export default function Page() {
  const { phoneMask, email } = useLoaderData<typeof loader>()
  return (
    <Card>
      <Label>Email address</Label>
      <div className='mt-1 flex w-full justify-between p-3'>
        <div className='flex space-x-3'>
          <Icon>mail</Icon>
          <span>{email}</span>
        </div>
      </div>
      <Label className='mt-4'>Mobile phone number</Label>
      <CardLink
        className='items-center justify-between'
        to={route('/otp/challenge')}
      >
        <div className='flex space-x-3'>
          <Icon>call</Icon>
          <span>{phoneMask}</span>
        </div>
        <Icon>navigate_next</Icon>
      </CardLink>
    </Card>
  )
}
