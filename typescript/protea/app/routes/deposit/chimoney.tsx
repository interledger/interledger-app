import type { ActionFunctionArgs, LoaderFunctionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useNavigation, useSubmit } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import { Button, Card } from '~/components'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { PaySelect } from '../pay_.$paymentId/PaySelect'
import { stringToBigInt } from './fynbos'

export async function chimoneyDepositLoader({ request }: LoaderFunctionArgs) {
  return jsonWithCSRF(request, { provider: 'chimoney' })
}

export function ChimoneyDepositPage() {
  const [amount, setAmount] = useState<string>('')
  const actionData = useActionData<typeof chimoneyAmountAction>()
  const navigation = useNavigation()
  const submit = useSubmit()
  const _onChangeDepositAmount = useCallback<
    ChangeEventHandler<HTMLInputElement>
  >((event) => {
    setAmount(event.target.value)
  }, [])
  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    if (navigation.state == 'submitting') {
      setLoading(true)
    } else if (navigation.state == 'loading' || navigation.state == 'idle') {
      setLoading(false)
    }
    // This ensures that loading is false when this route is unmounted.
    return () => setLoading(false)
  }, [navigation.formMethod, navigation.state, setLoading])

  useEffect(() => {
    const onSuccessfullPayment = (e: MessageEvent) => {
      if (!e.data || !e.data.status) return

      if (
        e.data.status === 'success' &&
        typeof e.data.issueID === 'string' &&
        e.data.issueID.length > 0
      ) {
        submit(
          {
            formName: 'chimoney-successfull-deposit',
            issueId: e.data.issueID
          },
          {
            action: route('/deposit'),
            method: 'POST'
          }
        )
      }
    }

    window.addEventListener('message', onSuccessfullPayment)

    return () => {
      window.removeEventListener('message', onSuccessfullPayment)
    }
  }, [submit])

  return (
    <>
      {actionData?.chimoneyWidget && (
        <>
          <iframe
            title='Deposit'
            src={actionData.chimoneyWidget}
            sandbox='allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts'
            scrolling='yes'
            frameBorder='0'
            className='h-[850px] overflow-scroll'
          />
        </>
      )}

      {!actionData?.chimoneyWidget && (
        <>
          <Form
            id='chimoney-amount'
            action={route('/deposit')}
            method='post'
            className='hidden'
          />
          <input
            form='chimoney-amount'
            name='formName'
            value='chimoney-amount'
            type='hidden'
          />
          <Card>
            <PaySelect
              id='depositAmount'
              label='Deposit amount'
              name='depositAmount'
              form='chimoney-amount'
              value={amount}
              onChange={_onChangeDepositAmount}
              onChangeLinkedAccount={() => {}}
              linkedAccountOptions={[]}
              placeholder='0'
              prefixIcon={<div className={`flag:CA`} />}
              type='number'
              min='0'
              step='0.01'
            />
          </Card>
          <Button
            type='submit'
            form='chimoney-amount'
            disabled={navigation.state === 'submitting'}
          >
            Continue
          </Button>
        </>
      )}
    </>
  )
}

export async function chimoneyAmountAction({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const depositAmount = String(form.get('depositAmount') || '')

  const response = await grpc.getChimoneyDepositLink(request, {
    amount: stringToBigInt(depositAmount),
    asset: 'CAD',
    assetScale: 2
  })
  if (isConnectError(response)) throw response.error

  return json({
    chimoneyWidget: response.link
  })
}

export async function chimoneySuccessfullDepositAction({
  request
}: ActionFunctionArgs) {
  const form = await request.formData()
  const issueId = String(form.get('issueId') || '')

  const response = await grpc.createChimoneyDeposit(request, { issueId })
  if (isConnectError(response)) throw response.error

  return redirect(route('/'))
}
