import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { Card, Icon, Layouts, Snackbar } from '~/components'
import { getUserSession } from '~/lib/kratos.server'
import { getSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)

  const snackbar = await getSnackbar(request)

  return json({
    traits: session.identity.traits,
    snackbar
  })
}

export const handle = {
  title: 'Contact information',
  layout: Layouts.Focus
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings | Personal details'
  }
}

export default function Page() {
  const { traits, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)
  return (
    <>
      <Card>
        <h2 className='text-sm font-medium'>Email address</h2>
        <div className='mt-2 flex items-center justify-start rounded-xl bg-nav p-3 text-medium'>
          <div className='flex space-x-3'>
            <Icon>mail</Icon>
            <span>{traits.email}</span>
          </div>
        </div>
        <h2 className='mt-6 text-sm font-medium'>Mobile phone number</h2>
        <div className='mt-2 flex items-center justify-start rounded-xl bg-nav p-3 text-medium'>
          <div className='flex space-x-3'>
            <Icon>call</Icon>
            <span>{traits.phone}</span>
          </div>
        </div>
      </Card>
      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        icon={snackbar.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        offset
        onClose={() => setSnackbar(false)}
      />
    </>
  )
}
