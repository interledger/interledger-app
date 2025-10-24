import {
  json,
  redirect,
  type ActionFunctionArgs,
  type LoaderFunctionArgs
} from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { CardProcessingPlaceholder, Layouts, type ApplicationProps } from '~/components'
import { ConfirmCard } from '~/components/OrderCardSteps/ConfirmCard'
import { CreateAddress } from '~/components/OrderCardSteps/CreateAddress'
import { DeliveryAddresses } from '~/components/OrderCardSteps/DeliveryAddresses'
import { ProductsSelect } from '~/components/OrderCardSteps/ProductsSelect'
import { getFeatures } from '~/data/wallet.server'
import {
  CardType,
  type CardApplicationProduct,
  type CustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { OrderCardStep, useOrderCardStore } from '~/lib/useOrderCardStore'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Order card', back: 'cards' }
  }
}

export async function loader({ request }: LoaderFunctionArgs) {
  const features = await getFeatures(request)
  if (!features.manageWalletCardsEnabled) {
    throw redirect('/')
  }

  const res = await grpc.getCardOrderOptions(request, {})
  if (isConnectError(res)) {
    throw res.errorResponse
  }

  return json(res)
}

export default function Page() {
  const { products, addresses, countries, isWaitingForCreation } = useLoaderData<typeof loader>()
  const [step, setProducts, setAddresses, reset, setCountries] =
    useOrderCardStore((state) => [
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

  if (isWaitingForCreation) {
    return <CardProcessingPlaceholder />
  }

  return (
    <>
      {step === OrderCardStep.CARD_TYPE && <ProductsSelect />}
      {step === OrderCardStep.DELIVERY && <DeliveryAddresses />}
      {step === OrderCardStep.CREATE_ADDRESS && <CreateAddress />}
      {step === OrderCardStep.CONFIRMATION && <ConfirmCard />}
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  await grpc.orderCard(request, {
    cardProductCode: 'PMDSGWEEA',
    type: CardType.VIRTUAL
  })

  return null
}
