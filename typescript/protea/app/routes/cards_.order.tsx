import { proto3 } from '@bufbuild/protobuf'
import {
  json,
  redirect,
  type ActionFunctionArgs,
  type LoaderFunctionArgs
} from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import {
  CardProcessingPlaceholder,
  Layouts,
  type ApplicationProps
} from '~/components'
import { ConfirmCard } from '~/components/OrderCardSteps/ConfirmCard'
import { CreateAddress } from '~/components/OrderCardSteps/CreateAddress'
import { DeliveryAddresses } from '~/components/OrderCardSteps/DeliveryAddresses'
import { ProductsSelect } from '~/components/OrderCardSteps/ProductsSelect'
import { createNewAddress } from '~/components/OrderCardSteps/utils'
import { getFeatures } from '~/data/wallet.server'
import {
  CardType,
  CustomerDeliveryAddressType,
  type CardApplicationProduct,
  type CustomerDeliveryAddress,
  type OrderCardRequest
} from '~/generated/connect/backend/v1/backend_pb'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { StorableNewAddress } from '~/lib/useOrderCardStore'
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
  const { products, addresses, countries, isWaitingForCreation } =
    useLoaderData<typeof loader>()
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
  let form = await request.formData()
  const cardProductCode = form.get('cardProductCode') as string
  const type = Number(form.get('type')) as CardType
  const deliveryAddressId = form.get('deliveryAddressId') as string
  const newDeliveryAddressJson = form.get('newDeliveryAddress') as string | null

  const deliveryAddress = createDeliveryAddress(
    type,
    deliveryAddressId,
    newDeliveryAddressJson
  )

  if ('error' in deliveryAddress) {
    return error(request, {
      errors: { form: deliveryAddress.error }
    })
  }

  const orderCardResponse = await grpc.orderCard(request, {
    cardProductCode: cardProductCode,
    type,
    deliveryAddress
  })

  if (isConnectError(orderCardResponse)) {
    return orderCardResponse.error({
      errors: { form: 'Card order failed, please try again.' }
    })
  }

  return redirectWithSnackbar(request, route('/cards'), {
    message:
      'Your card in the making! We’ll notify you as soon as it’s ready to go.',
    icon: 'close'
  })
}

const createDeliveryAddress = (
  type: CardType,
  deliveryAddressId: string,
  newDeliveryAddressJson: string | null
): OrderCardRequest['deliveryAddress'] | { error: string } => {
  if (type === CardType.VIRTUAL) {
    return {
      case: undefined,
      value: undefined
    }
  }

  if (!deliveryAddressId && !newDeliveryAddressJson) {
    return { error: 'No delivery address specified.' }
  }

  if (deliveryAddressId) {
    return {
      case: 'deliveryAddressId',
      value: deliveryAddressId
    }
  }

  if (newDeliveryAddressJson) {
    try {
      const newDeliveryAddress = JSON.parse(
        newDeliveryAddressJson
      ) as StorableNewAddress

      return {
        case: 'newDeliveryAddress',
        value: createNewAddress({
          details: {
            type: proto3.getEnumType(CustomerDeliveryAddressType).values[
              newDeliveryAddress.details!.type
            ] as unknown as CustomerDeliveryAddressType,
            country: newDeliveryAddress.details!.countryCode,
            line1: newDeliveryAddress.details!.line1,
            line2: newDeliveryAddress.details!.line2,
            line3: newDeliveryAddress.details!.line3,
            postOffice: newDeliveryAddress.details!.postOffice,
            city: newDeliveryAddress.details!.city,
            zipCode: newDeliveryAddress.details!.zipCode
          },
          reason: newDeliveryAddress.reason
        })
      }
    } catch (error) {
      return { error: 'Invalid delivery address. Please redo your new address.' }
    }
  }

  return { error: 'Invalid delivery address.' }
}
