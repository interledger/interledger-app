import {
  CustomerDeliveryAddress,
  CustomerDeliveryAddressType,
  NewCustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import { DeliveryAddress } from '~/lib/useAddCardStore'

export function toSelectableAddresses(address: DeliveryAddress) {
  return {
    id: address.id,
    name: `${address.details?.line1} ${address.details?.city} (${address.details?.countryCode})`,
    icon: getAddressIcon(address.details?.type!)
  }
}

export function toSelectableNewAddress(address: NewCustomerDeliveryAddress) {
  return {
    id: 'new',
    name: `${address.details?.line1} ${address.details?.city} (${address.details?.countryCode})`,
    icon: getAddressIcon(address.details?.type!)
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
