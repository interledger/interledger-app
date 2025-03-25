// Placeholder page

import { json, redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { Layouts, type ApplicationProps } from '~/components'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Add card', back: route('/cards') }
  }
}

export async function loader() {
  if (process.env.FYNBOS_ENV !== 'local') {
    throw redirect('/')
  }
  return json({})
}

export default function Page() {
  return <div>Placeholder</div>
}
