export type IframeMessageType = 'WithdrawalCompleted' | 'DepositCompleted' | 'StripeDepositCompleted';
export interface IframeMessage {
  type: IframeMessageType;
  uuid: string;
}