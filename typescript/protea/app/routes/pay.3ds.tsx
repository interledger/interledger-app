import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect, useRef } from 'react'
import { route } from 'routes-gen'
import { Layouts } from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { useScript } from '~/lib/useScript'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)
  let params = Object.fromEntries(url.searchParams.entries())

  await getUserSession(request)
  const flow = await requireFlow(request, flowType.Pay)
  return json({
    flow,
    jwt: params['jwt'],
    threeDsId: params['id']
  })
}

export const handle = {
  title: '3DS Verification',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: '3DS Verification'
  }
}

export default function Page() {
  const { flow, jwt: initJWT, threeDsId } = useLoaderData<typeof loader>()
  const submit = useSubmit()

  const state = useScript(
    'https://songbirdstag.cardinalcommerce.com/edge/v1/songbird.js'
  )
  let cardinalRef = useRef<any>(null)

  useEffect(() => {
    if (typeof window !== 'undefined' && state == 'ready') {
      cardinalRef.current = (window as any).Cardinal
      cardinalRef.current.configure({
        logging: {
          level: 'on'
        }
      })
      cardinalRef.current.setup('init', {
        jwt: initJWT
      })
      cardinalRef.current.on('payments.setupComplete', (data: any) => {
        console.log('payments.setupComplete', data)
        let formData = new FormData()
        formData.append('name', 'lookup')
        formData.append('threeDsId', threeDsId)
        formData.append('outgoingPaymentId', flow?.data?.idempotencyKey)

        submit(formData, {
          action: '/pay/3ds',
          method: 'post'
        })
      })

      cardinalRef.current.on(
        'payments.validated',
        (data: { ActionCode: any }, jwt: string) => {
          switch (data.ActionCode) {
            case 'SUCCESS':
              // Handle successful transaction, send JWT to backend to verify
              console.log('payments.validated SUCCESS', data)
              let formData = new FormData()
              formData.append('name', 'authenticate')
              formData.append('threeDsId', threeDsId)
              formData.append('outgoingPaymentId', flow?.data?.idempotencyKey)
              formData.append('jwt', jwt)

              submit(formData, {
                action: '/pay/3ds',
                method: 'post'
              })
              break

            case 'NOACTION':
              // Handle no actionable outcome
              console.log('payments.validated NOACTION', data)
              break

            case 'FAILURE':
              // Handle failed transaction attempt
              console.log('payments.validated FAILURE', data)
              break

            case 'ERROR':
              // Handle service level error`
              console.log('payments.validated ERROR', data)
              break
          }
        }
      )
    }
  }, [initJWT, state, threeDsId, flow?.data?.idempotencyKey, submit])

  return <></>
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const formName = await form.get('name')
  const outgoingPaymentID = await form.get('outgoingPaymentId')
  const threeDsId = await form.get('threeDsId')

  if (formName === 'lookup') {
    let rpc = await grpcClient
      .lookup3DS(
        {
          outgoingPaymentID: String(outgoingPaymentID),
          threeDSID: String(threeDsId)
        },
        {
          meta: {
            cookies: String(request.headers.get('cookie'))
          }
        }
      )
      .then((v) => v)
      .catch((e) => {
        return StatusError(e)
      })
    if (isGrpcError(rpc)) {
      throw json({}, httpMapping(rpc.code))
    }

    return
  }

  const jwt = await form.get('jwt')
  let rpc = await grpcClient
    .authenticate3DS(
      {
        outgoingPaymentID: String(outgoingPaymentID),
        threeDSID: String(threeDsId),
        jwt: String(jwt)
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie'))
        }
      }
    )
    .then((v) => v)
    .catch((e) => {
      return StatusError(e)
    })
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  await exitFlow(request, flowType.Pay)
  return redirect(route('/'))
}
