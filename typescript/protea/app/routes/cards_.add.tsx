import { json, redirect, type LoaderFunctionArgs } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { Layouts, type ApplicationProps } from '~/components'
import { ConfirmCard } from '~/components/AddCardSteps/ConfirmCard'
import { CreateAddress } from '~/components/AddCardSteps/CreateAddress'
import { DeliveryAddresses } from '~/components/AddCardSteps/DeliveryAddresses'
import { ProductsSelect } from '~/components/AddCardSteps/ProductsSelect'
import { getFeatures } from '~/data/wallet.server'
import {
  CardApplicationProduct,
  CustomerDeliveryAddress
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

  const [deliveryAddresses, products, countries] = await Promise.all([
    grpc.getCustomerDeliveryAddresses(request, {}),
    grpc.getCardApplicationProducts(request, {}),
    grpc.getCountries(request, {})
  ])

  if (isConnectError(deliveryAddresses)) {
    console.error('Error getting delivery addresses: ', deliveryAddresses)
    throw deliveryAddresses.errorResponse
  }
  if (isConnectError(products)) {
    console.error('Error getting products: ', products)
    throw products.errorResponse
  }
  if (isConnectError(countries)) {
    console.error('Error getting countries: ', countries)
    throw countries.errorResponse
  }

  return json({
    products: products.products,
    addresses: deliveryAddresses.deliveryAddresses,
    countries: countries.countries
  })
}

export default function Page() {
  const { products, addresses, countries } = useLoaderData<typeof loader>()
  const [step, setProducts, setAddresses, reset, setCountries] =
    useAddCardStore((state) => [
      state.step,
      state.setProducts,
      state.setAddresses,
      state.reset,
      state.setCountries
    ])

  useEffect(() => {
    return () => reset()
  }, [reset])

  useEffect(() => {
    setProducts(products as CardApplicationProduct[])
  }, [products, setProducts])

  useEffect(() => {
    setAddresses(addresses as CustomerDeliveryAddress[])
  }, [addresses, setAddresses])

  useEffect(() => {
    setCountries(countries)
  }, [countries, setCountries])

  return (
    <>
      {step === AddCardStep.CARD_TYPE && <ProductsSelect />}
      {step === AddCardStep.DELIVERY && <DeliveryAddresses />}
      {step === AddCardStep.CREATE_ADDRESS && <CreateAddress />}
      {step === AddCardStep.CONFIRMATION && <ConfirmCard />}
    </>
  )
}
