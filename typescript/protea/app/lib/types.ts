export type IframeMessageType = 'WithdrawalCompleted' | 'StripeDepositCompleted'

export interface IframeMessage {
  type: IframeMessageType
  uuid: string
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
