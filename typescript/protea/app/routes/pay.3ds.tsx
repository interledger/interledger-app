import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useActionData, useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect, useRef } from 'react'
import { route } from 'routes-gen'
import { Layouts, LoadingShapes } from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { useScript } from '~/lib/useScript'

export async function loader({ request }: LoaderArgs) {
  await getUserSession(request)
  const flow = await requireFlow(request, flowType.Pay)

  let threeDSInit = await grpcClient
    .init3DS(
      {
        idempotencyKey: flow.data.idempotencyKey,
        quoteID: flow.data.quoteID
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie'))
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(threeDSInit)) throw json({}, httpMapping(threeDSInit.code))

  return json({
    flow,
    jwt: threeDSInit.response.jwt,
    threeDsId: threeDSInit.response.id,
    songbirdURL: threeDSInit.response.songbirdURL
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
  const {
    flow,
    jwt: initJWT,
    threeDsId,
    songbirdURL
  } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const submit = useSubmit()
  const state = useScript(songbirdURL)
  let cardinalRef = useRef<any>(null)

  useEffect(() => {
    if (typeof window !== 'undefined' && state == 'ready' && cardinalRef.current === null) {
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
        formData.append('idempotencyKey', flow?.data?.idempotencyKey)

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
              formData.append('idempotencyKey', flow?.data?.idempotencyKey)
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

  useEffect(() => {
    if (typeof window !== 'undefined' && cardinalRef.current !== null) {
      cardinalRef.current.continue(
        'cca',
        {
          AcsUrl: actionData?.challengeURL,
          Payload: actionData?.payload
        },
        {
          OrderDetails: {
            TransactionId: actionData?.processorTransactionID
          }
        }
      )
    }
  }, [actionData])

  return (
    <>
      <LoadingShapes />
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const formName = await form.get('name')
  const idempotencyKey = await form.get('idempotencyKey')
  const threeDsId = await form.get('threeDsId')

  if (formName === 'lookup') {
    let lookup3DS = await grpcClient
      .lookup3DS(
        {
          idempotencyKey: String(idempotencyKey),
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
    if (isGrpcError(lookup3DS)) {
      throw json({}, httpMapping(lookup3DS.code))
    }

    if (lookup3DS.response.challengeURL !== '') {
      return json(
        {
          challengeURL: lookup3DS.response.challengeURL,
          payload: lookup3DS.response.payload,
          processorTransactionID: lookup3DS.response.processorTransactionID
        },
        200
      )
    }
  }

  if (formName === 'authenticate') {
    const jwt = await form.get('jwt')
    let auth3DS = await grpcClient
      .authenticate3DS(
        {
          idempotencyKey: String(idempotencyKey),
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
    if (isGrpcError(auth3DS)) {
      throw json({}, httpMapping(auth3DS.code))
    }
  }

  const flow = await requireFlow(request, flowType.Pay)
  const clientIpAddress = getClientIP(request)
  let payment = await openPaymentsClient
    .createOutgoingPayment(
      {
        idempotencyKey: flow.data.idempotencyKey || '',
        quoteID: flow.data.quoteID,
        description: flow.data.note,
        externalRef: '',
        ipAddress: clientIpAddress,
        threeDSID: String(threeDsId)
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(payment)) throw json({}, httpMapping(payment.code))

  await exitFlow(request, flowType.Pay)
  return redirect(route('/'))
}
