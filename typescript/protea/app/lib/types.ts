export type IframeMessageType = 'WithdrawalCompleted' | 'StripeDepositCompleted';
export interface IframeMessage {
  type: IframeMessageType;
  uuid: string;
}

export type Errors = {
  errors?: Array<string | null | undefined> | null
}

export type FormattedAmount = {
  amount: number
  amountWithCurrency: string
  symbol: string
}

export interface Amount {
  value: string
  assetCode: string
  assetScale: number
}

export interface WalletAddressType {
  id: string
  assetScale: number
  assetCode: string
  authServer: string
  resourceServer: string
}

export type FormatAmountArgs = Amount & {
  value: string
}

export type QuickPaySession = {
  validWalletAddress?: any
  receiverAddress?: any
  quote?: any
  grants?: any
}

export type ActionData = {
  errors?: {
    walletAddress?: Errors
    receiverAddress?: Errors
    note?: Errors
    actionError?: Errors
  }
}