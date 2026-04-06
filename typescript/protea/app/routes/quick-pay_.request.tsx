import type { Route } from './+types/quick-pay_.request'
import { data } from 'react-router'
import { useLoaderData } from 'react-router'
import type { MetaFunction } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  console.log({ url })
  const features = null

  return data({
    features
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Interledger Pay'
  }
])

export default function Page() {
  const { features } = useLoaderData<typeof loader>()

  return (
    <>
      <p>{features}</p>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const type = form.get('type') as string

  console.log(type)

  return null
}
