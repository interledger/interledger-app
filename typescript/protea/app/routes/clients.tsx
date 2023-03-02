import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button, Card, Shape, TextArea } from '~/components'
import type { JWK } from '~/generated/protobuf-ts/backend/v1/backend'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  if (process.env.FYNBOS_ENV !== 'local' && process.env.FYNBOS_ENV !== 'dev') {
    return json({}, 404)
  }

  return json({})
}

export default function Page() {
  useLoaderData<typeof loader>()

  return (
    <Card>
      <div className='flex justify-between'>
        <h1 className='font-display text-2xl font-medium'>Address details</h1>
        <div className='hidden sm:flex'>
          <Shape
            width={'w-8'}
            radius={'rounded-full'}
            color={'bg-yellow-300'}
          />
          <Shape
            width={'w-8'}
            radius={'rounded-bl-full'}
            color={'bg-slate-300'}
          />
        </div>
      </div>
      <p className='mt-6 text-medium'>Please provide your address.</p>

      <Form
        id='client-jwk'
        action='/clients'
        method='post'
        className='hidden'
      />
      <TextArea
        id='key'
        form='client-jwk'
        label='Ed25519 Public key (JWK)'
        name='key'
        defaultValue=''
        className='mt-6'
      />

      <Button className='mt-12' form='client-jwk' type='submit'>
        Submit
      </Button>
    </Card>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const key = form.get('key') as string

  let jwk: JWK
  try {
    jwk = JSON.parse(key)
  } catch (e) {
    console.log(e)
    return json({}, 500)
  }
  console.log(key)
  const response = await grpcClient
    .createClientPublicKey(jwk, {
      meta: {
        cookies: String(request.headers.get('cookie')) || ''
      }
    })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return json({}, 201)
}
