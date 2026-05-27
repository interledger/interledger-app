import { PendingGrant, IncomingPayment, Quote, type WalletAddress } from '@interledger/open-payments'

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

export type FormatAmountArgs = Amount & {
  value: string
}
export type QuoteResponse = Quote & {
  incomingPaymentGrantToken: string
}

export type QuickPaySession = {
  senderAddress?: WalletAddress
  receiverAddress?: WalletAddress
  quote?: QuoteResponse
  grants?: Record<string, PendingGrant>
  request?: IncomingPayment
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
  Sentinel // End of range value must be last, no need to public
}
