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
  request?: any
}

export type ActionData = {
  errors?: {
    walletAddress?: Errors
    receiverAddress?: Errors
    senderAddress?: Errors
    note?: Errors
    actionError?: Errors
  }
}
export enum KycStatus {
  Unknown = 0,
  Pending = 1,
  DocumentsRequired = 2,
  Approved = 3,
  Denied = 4,
  InReview = 5,
  Level1 = 6,
  Level2 = 7
}

export enum PaymentRequiredAction {
  Unknown,
  ThreeDS,
  SenderIdentifier,
  SenderAccount,
  ReceiverIdentifier,
  SenderAmount,
  ReceiverAmount,
  OTP,
  IPAddress
}

export enum PaymentIdentityType {
  Unknown,
  Twitter,
  WalletID,
  WalletURL,
  Slack,
  Discord,
  Sentinel // End of range value must be last, no need to public
}
