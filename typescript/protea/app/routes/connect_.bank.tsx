import * as widgetSdk from '@mxenabled/web-widget-sdk'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { useLoaderData, useRevalidator, useSubmit } from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Dialog, Layouts, Shape } from '~/components'
import { getFeatures } from '~/data/wallet.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export async function loader({ request }: LoaderArgs) {
  // TODO Add colorScheme option once theme is in the users session
  const features = await getFeatures(request)

  if (!features.banksEnabled) {
    throw redirect(route('/'))
  }

  let response = await grpc.getMXWidget(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return jsonWithCSRF(request, { url: response.url })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: 'Connect bank account'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Connect bank account'
  }
}

export default function Page() {
  const submit = useSubmit()
  const { url, csrfToken } = useLoaderData<typeof loader>()
  let widgetRef = useRef<any>(null)
  const { revalidate } = useRevalidator()

  const [showDialog, setShowDialog] = useState<boolean>(false)

  useEffect(() => {
    if (showDialog && widgetRef.current === null) {
      widgetRef.current = new widgetSdk.ConnectWidget({
        container: '#widget',
        url,
        onMemberConnected: (event) => {
          let formData = new FormData()
          formData.append('userGuid', event.user_guid)
          formData.append('memberGuid', event.member_guid)
          formData.append('sessionGuid', event.session_guid)
          formData.append('csrfToken', csrfToken)
          submit(formData, {
            action: '/connect/bank',
            method: 'post'
          })
        }
      })
    }

    return () => {
      if (widgetRef.current) widgetRef.current.unmount()
      widgetRef.current = null
      revalidate()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showDialog])

  return (
    <>
      <Card>
        <CardContent>
          <p>Connect a bank account to easily send and receive payments.</p>
          <div className='mt-6 flex items-start'>
            <Shape
              width='w-8'
              flex='flex-none'
              radius='rounded-tr-full rounded-bl-full'
              color='bg-indigo-400'
            />
            <Shape
              width='w-8'
              flex='flex-none'
              radius='rounded-full rounded-tl-none'
              color='bg-yellow-300'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Bank details</h3>
              <p className='text-medium'>
                We will retrieve your bank information with your permission via
                a secure connection.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
      <Button type='button' onClick={() => setShowDialog(true)}>
        Let's go
      </Button>
      <Dialog grow unmount={false} open={showDialog} setOpen={setShowDialog}>
        <div id='widget' />
      </Dialog>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    form: ''
  }

  let response = await grpc.createMXBankAccounts(request, {
    memberGuid: form.get('memberGuid') as string,
    sessionGuid: form.get('sessionGuid') as string,
    userGuid: form.get('userGuid') as string
  })

  if (isConnectError(response)) {
    errors.form = 'Failed to connect bank account.'
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirect(route('/accounts'))
}
