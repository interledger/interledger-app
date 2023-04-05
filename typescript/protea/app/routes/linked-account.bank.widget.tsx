import * as widgetSdk from '@mxenabled/web-widget-sdk'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSubmit, useRevalidator } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import { Card, Layouts } from '~/components'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { Code } from "~/generated/protobuf-ts/google/rpc/code";

export async function loader({ request }: LoaderArgs) {
  let url
  let rpc = await grpcClient
    .getMXWidget(
      {},
      {
        meta: { cookies: String(request.headers.get('cookie')) }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    if (rpc.code === Code.INTERNAL || Code.DEADLINE_EXCEEDED) url = ''
    else throw json({}, httpMapping(rpc.code))
  } else url = rpc.response.url

  return json({ url })
}

export const handle = {
  title: 'Bank details',
  layout: Layouts.FocusLayout
}

export default function Page() {
  const submit = useSubmit()
  const { state ,revalidate } = useRevalidator()
  const { url } = useLoaderData<typeof loader>()

  useEffect(() => {
    console.log('url', url)
    let widget: widgetSdk.ConnectWidget

    if (!url && state !== 'loading') revalidate()
    else {
      widget = new widgetSdk.ConnectWidget({
        container: '#widget',
        style: {
          width: '100%',
          border: "none",
          minHeight: '600px',
        },
        url,
        onLoad: (payload) => {
          const widget = document.getElementById('widget')
          const iframe = widget?.getElementsByTagName('iframe')[0]
          console.log('scrollHeight', iframe?.contentWindow)
          // iframe?.setAttribute('style', `width: 100%; border: none; height: ${iframe.scrollHeight}px;`)
          console.log('actual body height', payload)
        },
        // onPing: (event) => {
        //   console.log('ping', event)
        //   const widget = document.getElementById('widget')
        //   const iframe = widget?.getElementsByTagName('iframe')[0]
        //   iframe?.setAttribute('style', `width: 100%; border: none; height: ${iframe.scrollHeight}px;`)
        // },
        // onMessage: (event) => {
        //   console.log('message', event.toString())
        // },
        // onStepChange: (event) => {
        //   console.log('step', event)
        //   const widget = document.getElementById('widget')
        //   const iframe = widget?.getElementsByTagName('iframe')[0]
        //   console.log('scrollHeight', iframe?.scrollHeight)
        //   console.log('clientHeight', iframe?.clientHeight)
        //   console.log('actual body height', iframe?.contentWindow)
        //   iframe?.setAttribute('style', `width: 100%; border: none; height: ${iframe.scrollHeight+10}px;`)
        // },
        onMemberConnected: (event) => {
          let formData = new FormData()
          formData.append('userGuid', event.user_guid)
          formData.append('memberGuid', event.member_guid)
          formData.append('sessionGuid', event.session_guid)
          submit(formData, {
            action: '/linked-account/bank/widget',
            method: 'post'
          })
        }
      })
    }

    return () => {
      widget.unmount()
    }
  }, [revalidate, state, submit, url])

  useEffect(() => {
    // Get div element by id widget and fetch the iframe element inside it and set the height of the iframe to scrollheight of iframe
    const widget = document.getElementById('widget')
    const iframe = widget?.getElementsByTagName('iframe')[0]
    iframe?.setAttribute('style', `height: ${iframe.scrollHeight}px`)
  }, [])

  return (
    <div className='overflow-hidden flex w-full flex-col rounded-2xl bg-page'>
      <div id='widget' />
     </div>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  let rpc = await grpcClient
    .createMXBankAccounts(
      {
        memberGuid: form.get('memberGuid') as string,
        sessionGuid: form.get('sessionGuid') as string,
        userGuid: form.get('userGuid') as string
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie'))
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json(null, httpMapping(rpc.code))
  }

  return redirect(route('/settings/linked-accounts'))
}
