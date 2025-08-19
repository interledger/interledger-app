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
import { AddCardStep, useAddCardStore } from '~/lib/useAddCardStore'
import { toSelectableAddresses } from './utils'

export const DeliveryAddresses = () => {
  const [setStep, newAddress, existingAddresses, setSelectedAddress] =
    useAddCardStore((state) => [
      state.setStep,
      state.newAddress,
      state.addresses,
      state.setSelectedAddress
    ])
  const displayableAddresses = useMemo(
    () => existingAddresses.map(toSelectableAddresses),
    [existingAddresses]
  )

  const [currentAddress, setCurrentAddress] = useState(displayableAddresses[0])

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
          onChange={setCurrentAddress}
        />
        {!newAddress ? (
          <Button onClick={() => setStep(AddCardStep.CREATE_ADDRESS)}>
            Create new address
          </Button>
        ) : (
          <OutlineButton onClick={() => setStep(AddCardStep.CREATE_ADDRESS)}>
            Edit new address
          </OutlineButton>
        )}
        <Button
          onClick={() => {
            setSelectedAddress(
              existingAddresses.find((a) => a.id === currentAddress.id)!
            )
            setStep(AddCardStep.CONFIRMATION)
          }}
        >
          Confirm
        </Button>
      </CardContent>
    </Card>
  )
}
