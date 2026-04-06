import type { Route } from './+types/settings.profile-contact'
import { data } from 'react-router';
import { useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Card, CardLink, Icon, Layouts } from '~/components'
import { Label } from '~/components/Label'
import { getUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: Route.LoaderArgs) {
  const session = await getUserSession(request)
  const len = session.identity.traits.phone.length
  const phoneMask = session.identity.traits.phone
    .substring(len - 4, len)
    .padStart(len, '*')

  return data({
    phoneMask,
    email: session.identity.traits.email
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Contact information'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Contact information'
  }
])

export default function Page() {
  const { phoneMask, email } = useLoaderData()
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
        to={href('/otp/challenge')}
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
