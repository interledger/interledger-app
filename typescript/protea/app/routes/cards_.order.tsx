// Placeholder page

import { route } from 'routes-gen'
import { Layouts, type ApplicationProps } from '~/components'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Order Card', back: route('/cards') }
  }
}

export default function Page() {
  return <div>Placeholder</div>
}
