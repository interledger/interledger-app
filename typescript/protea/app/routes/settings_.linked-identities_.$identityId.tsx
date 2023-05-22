import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData, useParams } from '@remix-run/react'
import { Button, Card, Layouts, TextField } from '~/components'
import { route } from 'routes-gen'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import { getIdentity, getLinkedAccount } from '~/lib/wallet.server'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'

export async function loader({ request, params }: LoaderArgs) {
  // get linked identity
  const identity = await getIdentity(request, params.identityId as string)

  return json({
    identity: identity
  })
}

export const handle = {
  title: 'Linked identity name',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings | Edit public name'
  }
}

export default function Page() {
  const { identity } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='verify-linked-identity'
        action={`/settings/linked-identities/${identity.id}`}
        method='post'
        className='hidden'
      />
      <Card>
        <div className='flex flex-col space-y-6'>
          <h1 className='font-display text-2xl font-medium'>
            {identity.identifier}
          </h1>
        </div>
        <div>{identity.id}</div>
        <div>{identity.state}</div>
      </Card>
      <Button form='verify-linked-identity' type='submit'>
        Verify
      </Button>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const cookie = String(request.headers.get('cookie'))
  const form = await request.formData()

  const identityId = params.identityId as string
  console.log(identityId)

  // Call verify identity flow.

  await flashSnackbar(request, {
    message: 'Linked identity verification in progress',
    icon: 'close'
  })

  return redirect(route('/settings/linked-identities'))
}
