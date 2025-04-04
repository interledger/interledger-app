// Placeholder page

import { json, redirect, type LoaderFunctionArgs } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
import { useEffect } from 'react'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Layouts,
  type ApplicationProps
} from '~/components'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'
import { grpc } from '~/lib/grpc.server'
import { AddCardStep, useAddCardStore } from '~/lib/useAddCardStore'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Add card', back: 'cards' }
  }
}

export async function loader({ request }: LoaderFunctionArgs) {
  if (process.env.FYNBOS_ENV !== 'local') {
    throw redirect('/')
  }

  const response = await grpc.getCustomerDeliveryAddresses(request, {})

  console.log(response)

  return json({
    products: [{ code: 'test', name: 'testing' }]
  })
}

function Product() {
  const [productCode, setProductCode, setType, products, setStep] =
    useAddCardStore((state) => [
      state.productCode,
      state.setProductCode,
      state.setCardType,
      state.products,
      state.setStep
    ])

  useEffect(() => {
    if (!productCode && products.length === 1) {
      setProductCode(products[0].code)
    }
  }, [productCode, products, setProductCode])

  return (
    <Card>
      <CardContent className='space-y-4'>
        <CardHeader>
          <CardTitle>Pick your card</CardTitle>
        </CardHeader>
        <div className='flex items-center justify-center gap-x-4'>
          {products.map((p) => (
            <button
              key={p.code}
              className={clsx(
                'overflow-hidden rounded-md',
                productCode === p.code ? 'outline outline-blue-500' : ''
              )}
              onClick={() => setProductCode(p.code)}
            >
              <img className='w-48' src={`/cards/${p.code}.png`} alt={p.name} />
            </button>
          ))}
        </div>
        {productCode && (
          <div className='flex w-full flex-col justify-between gap-y-4'>
            <Button
              onClick={() => {
                setType(CardType.Physical)
                setStep(AddCardStep.DELIVERY)
              }}
            >
              Physical
            </Button>
            <Button
              onClick={() => {
                setType(CardType.Virtual)
                setStep(AddCardStep.CONFIRMATION)
              }}
            >
              Virtual
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default function Page() {
  const { products } = useLoaderData<typeof loader>()
  const [step, setProducts, reset] = useAddCardStore((state) => [
    state.step,
    state.setProducts,
    state.reset
  ])

  useEffect(() => {
    return () => reset()
  }, [reset])

  useEffect(() => {
    setProducts(products)
  }, [products, setProducts])

  return (
    <>
      {step === AddCardStep.CARD_TYPE && <Product />}
      {step === AddCardStep.DELIVERY && <>delivery</>}
      {step === AddCardStep.CONFIRMATION && <>confirmation</>}
    </>
  )
}
