import { data, redirect } from 'react-router';
import { Form, useActionData, useNavigation, useSubmit } from 'react-router';
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { href } from 'react-router'
import { Button, Card } from '~/components'

import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { PaySelect } from '../pay_.$paymentId/PaySelect'
import { stringToBigInt } from './fynbos'


export function ChimoneyDepositPage() {
  const [amount, setAmount] = useState<string>('')
  const actionData = useActionData<any>()
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
            action: href('/deposit'),
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
            action={href('/deposit')}
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
              onChangeLinkedAccount={() => { }}
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

