import { Form } from '@remix-run/react'
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
import { useOrderCardStore } from '~/lib/useOrderCardStore'

export const ConfirmCard = () => {
  const [address, productCode, products, type] = useOrderCardStore((state) => [
    state.selectedAddress,
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

        {type === CardType.PHYSICAL && (
          <div className='mt-1'>
            <Label className='flex items-center gap-x-2'>
              <Icon>location_on</Icon>
              Delivery Address:
            </Label>

            <div className='flex w-full justify-between p-3'>
              <div className='flex space-x-3'>
                <div className='flex flex-col'>
                  <span>
                    <i>Primary address:</i> {address?.details?.line1}
                  </span>
                  {address?.details?.line2 && (
                    <span>
                      <i>Secondary address:</i> {address.details.line2}
                    </span>
                  )}
                  {address?.details?.line3 && (
                    <span>
                      <i>Third address:</i> {address.details.line3}
                    </span>
                  )}
                  <span>
                    <i>Zip code:</i> {address?.details?.zipCode}
                  </span>
                  <span>
                    <i>City:</i> {address?.details?.city}
                  </span>
                  <span>
                    <i>Country:</i> {address?.details?.countryCode}
                  </span>
                  {address?.details?.postOffice && (
                    <span>{address?.details?.postOffice}</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        <Label className='flex items-center gap-x-2'>
          <Icon>credit_card</Icon>
          Card type: {CardType[type]}
        </Label>
        <div className='flex items-center justify-center gap-x-4'>
          <img
            className='w-48'
            src={pickedProduct && `/cards/${pickedProduct.code}.png`}
            alt={pickedProduct?.name}
          />
        </div>
        {/* TODO remove: added for tests */}
        <Form
          id='confirm-card'
          action={`/cards/order`}
          method='post'
          className='hidden'
        />
        <Button form='confirm-card' type='submit'>
          Confirm
        </Button>
      </CardContent>
    </Card>
  )
}
