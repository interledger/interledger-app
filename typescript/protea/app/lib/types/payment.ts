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
