import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Icon
} from '~/components'
import { Label } from '~/components/Label'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'
import { useAddCardStore } from '~/lib/useAddCardStore'

export const ConfirmCard = () => {
  const [address, productCode, products, type] = useAddCardStore((state) => [
    state.selectedAddress,
    state.productCode,
    state.products,
    state.type
  ])
  console.log('address in confirm card', address)
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
