import { create } from 'zustand'
import { type SelectOptions } from '~/components'
import type {
  CardApplicationProduct,
  CustomerDeliveryAddress,
  NewCustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'

export enum AddCardStep {
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

interface AddCardState {
  step: AddCardStep
  type: CardType
  productCode: string | null
  products: CardApplicationProduct[]
  addresses: DeliveryAddress[]
  newAddress: NewCustomerDeliveryAddress | null
  selectedAddress: DeliveryAddress | null
  countries: Country[]
}

const initialState = {
  step: AddCardStep.CARD_TYPE,
  type: CardType.PHYSICAL,
  productCode: null,
  products: [],
  addresses: [],
  newAddress: null,
  selectedAddress: null,
  countries: []
} satisfies AddCardState

interface AddCardActions {
  setAddresses: (addresses: CustomerDeliveryAddress[]) => void
  setNewAddress: (address: NewCustomerDeliveryAddress) => void
  setSelectedAddress: (address: DeliveryAddress) => void
  isNewAddressSelected: () => boolean

  setProducts: (products: CardApplicationProduct[]) => void
  setProductCode: (productCode: string) => void
  setCardType: (type: AddCardState['type']) => void

  setCountries: (countries: Country[]) => void

  setStep: (step: AddCardStep) => void
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

export const useAddCardStore = create<AddCardState & AddCardActions>(
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
          case AddCardStep.DELIVERY:
            return { step: AddCardStep.CARD_TYPE }
          case AddCardStep.CONFIRMATION:
            return { step: AddCardStep.DELIVERY }
          case AddCardStep.CREATE_ADDRESS:
            return { step: AddCardStep.DELIVERY }
          default:
            return { ...initialState }
        }
      }),
    reset: () => set(() => ({ ...initialState }))
  })
)
