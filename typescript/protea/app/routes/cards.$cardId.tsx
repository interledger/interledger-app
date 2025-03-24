// Placeholder page

import { route } from 'routes-gen'
import { Layouts, type ApplicationProps } from '~/components'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/cards'),
      title: 'Cards'
    },
    isNested: true
  }
}

export default function Page() {
  return <div>Placeholder</div>
}
