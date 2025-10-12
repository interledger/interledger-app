import { create } from 'zustand'
import { type SelectOptions } from '~/components'
import type {
  CardApplicationProduct,
  CustomerDeliveryAddress,
  NewCustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'

export enum OrderCardStep {
  CARD_TYPE,
  DELIVERY,
  CREATE_ADDRESS,
  CONFIRMATION
}
const ID_NEW_ADDRESS = 'new'

export type StorableNewAddress = NewCustomerDeliveryAddress & {
  id: typeof ID_NEW_ADDRESS
}
export type DeliveryAddress = CustomerDeliveryAddress | StorableNewAddress
export type Country = { id: string; name: string }

interface OrderCardState {
  step: OrderCardStep
  type: CardType
  productCode: string | null
  products: CardApplicationProduct[]
  addresses: DeliveryAddress[]
  newAddress: NewCustomerDeliveryAddress | null
  selectedAddress: DeliveryAddress | null
  countries: Country[]
}

const initialState = {
  step: OrderCardStep.CARD_TYPE,
  type: CardType.PHYSICAL,
  productCode: null,
  products: [],
  addresses: [],
  newAddress: null,
  selectedAddress: null,
  countries: []
} satisfies OrderCardState

interface OrderCardActions {
  setAddresses: (addresses: CustomerDeliveryAddress[]) => void
  setNewAddress: (address: NewCustomerDeliveryAddress) => void
  setSelectedAddress: (address: DeliveryAddress) => void
  isNewAddressSelected: () => boolean

  setProducts: (products: CardApplicationProduct[]) => void
  setProductCode: (productCode: string) => void
  setCardType: (type: OrderCardState['type']) => void

  setCountries: (countries: Country[]) => void

  setStep: (step: OrderCardStep) => void
  stepBack: () => void
  reset: () => void
}

export const cardTypes: SelectOptions[] = [
  {
    id: 'physical',
    name: 'Physical'
  },
  {
    id: 'virtual',
    name: 'Virtual'
  }
]

export const useOrderCardStore = create<OrderCardState & OrderCardActions>(
  (set, get) => ({
    ...initialState,

    setAddresses: (addresses) => set(() => ({ addresses: addresses })),
    setNewAddress: (address) => {
      const newStorableAddress = {
        ...address,
        id: ID_NEW_ADDRESS
      } as StorableNewAddress
      const existingAddresses = get().addresses.filter(
        (a) => a.id !== ID_NEW_ADDRESS
      )

      set(() => ({
        newAddress: address,
        addresses: [newStorableAddress, ...existingAddresses]
      }))
    },
    setSelectedAddress: (address) => set(() => ({ selectedAddress: address })),
    isNewAddressSelected: () => get().selectedAddress?.id === ID_NEW_ADDRESS,

    setProducts: (products) => set(() => ({ products: products })),
    setProductCode: (productCode) => set(() => ({ productCode: productCode })),
    setCardType: (type) => set(() => ({ type: type })),

    setCountries: (countries) => set(() => ({ countries: countries })),

    setStep: (step) => set(() => ({ step: step })),
    stepBack: () =>
      set((state) => {
        switch (state.step) {
          case OrderCardStep.DELIVERY:
            return { step: OrderCardStep.CARD_TYPE }
          case OrderCardStep.CONFIRMATION:
            return { step: OrderCardStep.DELIVERY }
          case OrderCardStep.CREATE_ADDRESS:
            return { step: OrderCardStep.DELIVERY }
          default:
            return { ...initialState }
        }
      }),
    reset: () => set(() => ({ ...initialState }))
  })
)
