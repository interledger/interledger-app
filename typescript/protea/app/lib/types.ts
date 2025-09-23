export type IframeMessageType = 'WithdrawalCompleted' | 'StripeDepositCompleted';
export interface IframeMessage {
  type: IframeMessageType;
  uuid: string;
}