import type { PlainMessage } from '@bufbuild/protobuf'
import { create } from 'zustand'
import { type SelectOptions } from '~/components'
import type {
  CardApplicationProduct,
  CustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'

export enum AddCardStep {
  CARD_TYPE,
  DELIVERY,
  CONFIRMATION
}

interface AddCardState {
  step: AddCardStep
  type: CardType
  products: PlainMessage<CardApplicationProduct>[]
  address: CustomerDeliveryAddress | null
  productCode: string | null
}

const initialState = {
  step: AddCardStep.CARD_TYPE,
  type: CardType.Physical,
  products: [],
  productCode: null,
  address: null
} satisfies AddCardState

interface AddCardActions {
  setStep: (step: AddCardStep) => void
  stepBack: () => void
  setCardType: (type: AddCardState['type']) => void
  setProductCode: (productCode: string) => void
  setProducts: (products: PlainMessage<CardApplicationProduct>[]) => void
  setAddress: (address: CustomerDeliveryAddress | null) => void
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

export const useAddCardStore = create<AddCardState & AddCardActions>((set) => ({
  ...initialState,
  setStep: (step) => set(() => ({ step: step })),
  setCardType: (type) => set(() => ({ type: type })),
  setAddress: (address) => set(() => ({ address: address })),
  setProducts: (products) => set(() => ({ products: products })),
  setProductCode: (productCode) => set(() => ({ productCode: productCode })),
  stepBack: () =>
    set((state) => {
      switch (state.step) {
        case AddCardStep.DELIVERY:
          return { step: AddCardStep.CARD_TYPE }
        case AddCardStep.CONFIRMATION:
          return { step: AddCardStep.CONFIRMATION }
        default:
          return { ...initialState }
      }
    }),
  reset: () => set(() => ({ ...initialState }))
}))
