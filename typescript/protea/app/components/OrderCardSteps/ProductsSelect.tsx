import clsx from 'clsx'
import { useEffect } from 'react'
import { Button, Card, CardContent, CardHeader, CardTitle } from '~/components'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'
import { OrderCardStep, useOrderCardStore } from '~/lib/useOrderCardStore'

export const ProductsSelect = () => {
  const [productCode, setProductCode, setType, products, setStep] =
    useOrderCardStore((state) => [
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
                setType(CardType.PHYSICAL)
                setStep(OrderCardStep.DELIVERY)
              }}
            >
              Physical
            </Button>
            <Button
              onClick={() => {
                setType(CardType.VIRTUAL)
                setStep(OrderCardStep.CONFIRMATION)
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
