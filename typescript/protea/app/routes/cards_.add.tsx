import {
  json,
  redirect,
  type LoaderFunctionArgs
} from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
import { useEffect } from 'react'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Icon,
  Layouts,
  RadioGroup,
  type ApplicationProps
} from '~/components'
import { CreateAddress } from '~/components/AddCardSteps/CreateAddress'
import { Label } from '~/components/Label'
import { getFeatures } from '~/data/wallet.server'
import {
  CardType,
  CustomerDeliveryAddress,
  CustomerDeliveryAddressType
} from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { AddCardStep, useAddCardStore } from '~/lib/useAddCardStore'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Add card', back: 'cards' }
  }
}

export async function loader({ request }: LoaderFunctionArgs) {
  const features = await getFeatures(request)
  if (!features.manageWalletCardsEnabled) {
    throw redirect('/')
  }

  const [deliveryAddresses, products] = await Promise.all([
    grpc.getCustomerDeliveryAddresses(request, {}),
    grpc.getCardApplicationProducts(request, {})
  ])

  if (isConnectError(deliveryAddresses)) throw deliveryAddresses.errorResponse
  if (isConnectError(products)) throw products.errorResponse

  return json({
    products: products.products,
    addresses: deliveryAddresses.deliveryAddresses
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
                setType(CardType.PHYSICAL)
                setStep(AddCardStep.DELIVERY)
              }}
            >
              Physical
            </Button>
            <Button
              onClick={() => {
                setType(CardType.VIRTUAL)
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

function ConfirmCard() {
  const [address, productCode, products, type] = useAddCardStore((state) => [
    state.address,
    state.productCode,
    state.products,
    state.type
  ])
  const pickedProduct = products.find((p) => p.code === productCode)
  return (
    <Card>
      <CardContent className='space-y-4'>
        <CardHeader>
          <CardTitle>Confirm Card</CardTitle>
        </CardHeader>
        <Label>{CardType[type]}</Label>
        <div className='flex items-center justify-center gap-x-4'>
          <img
            className='w-48'
            src={pickedProduct && `/cards/${pickedProduct.code}.png`}
            alt={pickedProduct?.name}
          />
        </div>
        {type === CardType.PHYSICAL && (
          <>
            <Label>Delivery Address</Label>

            <div className='mt-1 flex w-full justify-between p-3'>
              <div className='flex space-x-3'>
                <Icon>location_on</Icon>
                <div className='flex flex-col'>
                  <span>{address?.details?.line1}</span>
                  {address?.details?.line2 && (
                    <span>{address.details.line2}</span>
                  )}
                  {address?.details?.line3 && (
                    <span>{address.details.line3}</span>
                  )}
                  <span>{address?.details?.zipCode}</span>
                  <span>{address?.details?.city}</span>
                  <span>{address?.details?.countryCode}</span>
                  {address?.details?.postOffice && (
                    <span>{address?.details?.postOffice}</span>
                  )}
                </div>
              </div>
            </div>
          </>
        )}
        <Button>Confirm</Button>
      </CardContent>
    </Card>
  )
}

function getAddressIcon(type: CustomerDeliveryAddressType) {
  switch (type) {
    case CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE:
      return 'home'
    default:
      return 'location_on'
  }
}

function DeliveryAddresses({
  addresses
}: {
  addresses: CustomerDeliveryAddress[]
}) {
  console.log('addresses in delivery addresses', addresses)
  const pickedAddresses = addresses.map((a) => ({
    id: a.id,
    name: `${a.details?.line1} ${a.details?.city} (${a.details?.countryCode})`,
    icon: getAddressIcon(a.details?.type!)
  }))
  const [setStep, setAddress] = useAddCardStore((state) => [
    state.setStep,
    state.setAddress
  ])

  return (
    <Card>
      <CardContent className='space-y-4'>
        <CardHeader>
          <CardTitle>Pick a Delivery Address</CardTitle>
        </CardHeader>
        <RadioGroup
          id='address'
          value={pickedAddresses[0]}
          options={pickedAddresses}
          // TODO: Fix
          // @ts-expect-error: fix
          onChange={(v) => setAddress(pickedAddresses.id)}
        />
        <Button onClick={() => setStep(AddCardStep.CREATE_ADDRESS)}>
          Create new address
        </Button>
        <Button
          onClick={() => {
            setStep(AddCardStep.CONFIRMATION)
            setAddress(addresses[0] as any)
          }}
        >
          Confirm
        </Button>
      </CardContent>
    </Card>
  )
}

export default function Page() {
  const { products, addresses } = useLoaderData<typeof loader>()
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
      {step === AddCardStep.DELIVERY && (
        // @ts-expect-error
        <DeliveryAddresses addresses={addresses} />
      )}
      {step === AddCardStep.CREATE_ADDRESS && <CreateAddress />}
      {step === AddCardStep.CONFIRMATION && <ConfirmCard />}
    </>
  )
}
