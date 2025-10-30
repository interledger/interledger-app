import {
  CustomerDeliveryAddressType,
  NewCustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import type { Country, DeliveryAddress } from '~/lib/useOrderCardStore'
import type { SelectOptions } from '../Select'
import type { AddressFormData} from './CreateAddress';
import { addressTypeOptions } from './CreateAddress'

export type SelectableAddress = {
  id: string
  name: string
  icon: string
  label?: string
}

export function toSelectableAddresses(
  address: DeliveryAddress
): SelectableAddress {
  return {
    id: address.id,
    name: `${address.details?.line1} ${address.details?.city} (${address.details?.countryCode})`,
    icon: getAddressIcon(address.details?.type || CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_OTHER),
    label: address.id === 'new' ? 'New' : undefined
  }
}

export function getAddressIcon(type: CustomerDeliveryAddressType) {
  switch (type) {
    case CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE:
      return 'home'
    default:
      return 'location_on'
  }
}

export const getAddressTypeValue = (
  currentType: CustomerDeliveryAddressType
) => {
  switch (currentType) {
    case CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_WORK:
      return addressTypeOptions[1]
    default:
      return addressTypeOptions[0]
  }
}

export const getCountryOptions = (countries: Country[]): SelectOptions[] => {
  return countries.map((country) => ({
    id: country.id,
    name: country.name
  }))
}

export const getCountryValue = (countryCode: string, countries: Country[]) => {
  const country = countries.find((c) => c.id === countryCode)
  return country
}

export const createNewAddress = (
  data: AddressFormData
): NewCustomerDeliveryAddress => {
  return new NewCustomerDeliveryAddress({
    details: {
      type: data.details.type,
      countryCode: data.details.country.toUpperCase(),
      line1: data.details.line1,
      line2: data.details.line2 || undefined,
      line3: data.details.line3 || undefined,
      postOffice: data.details.postOffice || undefined,
      city: data.details.city,
      zipCode: data.details.zipCode
    },
    reason: data.reason
  })
}
