import { useMemo, useState } from 'react'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  OutlineButton,
  RadioGroup
} from '~/components'
import { OrderCardStep, useOrderCardStore } from '~/lib/useOrderCardStore'
import { SelectableAddress, toSelectableAddresses } from './utils'

export const DeliveryAddresses = () => {
  const [setStep, newAddress, existingAddresses, setSelectedAddress] =
    useOrderCardStore((state) => [
      state.setStep,
      state.newAddress,
      state.addresses,
      state.setSelectedAddress
    ])
  const displayableAddresses = useMemo(
    () => existingAddresses.map(toSelectableAddresses),
    [existingAddresses]
  )

  const [currentAddress, setCurrentAddress] = useState<SelectableAddress>(
    displayableAddresses[0]
  )

  return (
    <Card>
      <CardContent className='space-y-4'>
        <CardHeader>
          <CardTitle>Pick a Delivery Address</CardTitle>
        </CardHeader>

        <RadioGroup
          id='address'
          value={currentAddress}
          options={displayableAddresses}
          onChange={(value) => setCurrentAddress(value)}
        />

        {!newAddress ? (
          <Button onClick={() => setStep(OrderCardStep.CREATE_ADDRESS)}>
            Create new address
          </Button>
        ) : (
          <OutlineButton onClick={() => setStep(OrderCardStep.CREATE_ADDRESS)}>
            Edit new address
          </OutlineButton>
        )}

        <Button
          onClick={() => {
            setSelectedAddress(
              existingAddresses.find((a) => a.id === currentAddress.id)!
            )
            setStep(OrderCardStep.CONFIRMATION)
          }}
        >
          Confirm
        </Button>
      </CardContent>
    </Card>
  )
}
