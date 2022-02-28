import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
};

export type Account = {
  __typename?: 'Account';
  balance: Scalars['String'];
  id: Scalars['ID'];
  recentTransactions: Array<Transaction>;
};

export type CreateIdentityInput = {
  country: Scalars['String'];
  firstName: Scalars['String'];
  lastName: Scalars['String'];
  mobileNumber: Scalars['String'];
};

export type CreateIdentityMutationResponse = MutationResponse & {
  __typename?: 'CreateIdentityMutationResponse';
  code: Scalars['String'];
  identity?: Maybe<Identity>;
  message: Scalars['String'];
  success: Scalars['Boolean'];
};

export type DepositInput = {
  amount: Scalars['String'];
  fundingSourceID: Scalars['ID'];
};

export type DepositMutationResponse = MutationResponse & {
  __typename?: 'DepositMutationResponse';
  code: Scalars['String'];
  message: Scalars['String'];
  success: Scalars['Boolean'];
  transaction?: Maybe<Transaction>;
};

export type FundingSource = {
  __typename?: 'FundingSource';
  id: Scalars['ID'];
  mask: Scalars['String'];
  name: Scalars['String'];
  subType: Scalars['String'];
  type: Scalars['String'];
  verificationStatus: Scalars['String'];
};

export type Identity = {
  __typename?: 'Identity';
  address: Array<Maybe<Scalars['String']>>;
  city: Scalars['String'];
  country: Scalars['String'];
  dateOfBirth: Scalars['String'];
  email: Scalars['String'];
  firstName: Scalars['String'];
  id: Scalars['ID'];
  lastName: Scalars['String'];
  mobileNumber: Scalars['String'];
  postalCode: Scalars['String'];
  state: Scalars['String'];
  taxIdNumber: Scalars['String'];
  verificationState: Scalars['String'];
};

export type LinkFundingSourceMutationResponse = MutationResponse & {
  __typename?: 'LinkFundingSourceMutationResponse';
  code: Scalars['String'];
  fundingSource?: Maybe<FundingSource>;
  message: Scalars['String'];
  success: Scalars['Boolean'];
};

export type LinkUsdBankAccountInput = {
  accountNumber: Scalars['String'];
  institution: Scalars['String'];
  name: Scalars['String'];
  routingNumber: Scalars['String'];
  type: Scalars['String'];
};

export type Mutation = {
  __typename?: 'Mutation';
  createIdentity: CreateIdentityMutationResponse;
  initiateDeposit: DepositMutationResponse;
  initiateOutgoingPayment: OutgoingPaymentMutationResponse;
  initiateWithdrawal: WithdrawalMutationResponse;
  linkUsdBankAccount: LinkFundingSourceMutationResponse;
  verify: VerifyMutationResponse;
  verifyUsdBankAccount: VerifyUsdBankAccountMutationResponse;
};


export type MutationCreateIdentityArgs = {
  input: CreateIdentityInput;
};


export type MutationInitiateDepositArgs = {
  input: DepositInput;
};


export type MutationInitiateOutgoingPaymentArgs = {
  input: OutgoingPaymentInput;
};


export type MutationInitiateWithdrawalArgs = {
  input: WithdrawalInput;
};


export type MutationLinkUsdBankAccountArgs = {
  input: LinkUsdBankAccountInput;
};


export type MutationVerifyArgs = {
  input: VerificationInput;
};


export type MutationVerifyUsdBankAccountArgs = {
  input: VerifyUsdBankAccountInput;
};

export type MutationResponse = {
  code: Scalars['String'];
  message: Scalars['String'];
  success: Scalars['Boolean'];
};

export type OutgoingPaymentInput = {
  amount: Scalars['String'];
  to: Scalars['String'];
};

export type OutgoingPaymentMutationResponse = MutationResponse & {
  __typename?: 'OutgoingPaymentMutationResponse';
  code: Scalars['String'];
  message: Scalars['String'];
  success: Scalars['Boolean'];
  transaction?: Maybe<Transaction>;
};

export type Query = {
  __typename?: 'Query';
  account?: Maybe<Account>;
  fundingSources: Array<Maybe<FundingSource>>;
  identity?: Maybe<Identity>;
};

export type Transaction = {
  __typename?: 'Transaction';
  amount: Scalars['String'];
  description: Scalars['String'];
  id: Scalars['ID'];
  status: Scalars['String'];
  timestamp: Scalars['String'];
  type: TransactionType;
};

export enum TransactionType {
  Deposit = 'DEPOSIT',
  Sent = 'SENT',
  Withdrawal = 'WITHDRAWAL'
}

export type VerificationInput = {
  Address: Array<Scalars['String']>;
  City: Scalars['String'];
  DateOfBirth: Scalars['String'];
  PostalCode: Scalars['String'];
  State: Scalars['String'];
  TaxIdNumber: Scalars['String'];
};

export type VerifyMutationResponse = MutationResponse & {
  __typename?: 'VerifyMutationResponse';
  code: Scalars['String'];
  identity?: Maybe<Identity>;
  message: Scalars['String'];
  success: Scalars['Boolean'];
};

export type VerifyUsdBankAccountInput = {
  FundingSourceId: Scalars['String'];
};

export type VerifyUsdBankAccountMutationResponse = MutationResponse & {
  __typename?: 'VerifyUsdBankAccountMutationResponse';
  code: Scalars['String'];
  fundingSource?: Maybe<FundingSource>;
  message: Scalars['String'];
  success: Scalars['Boolean'];
};

export type WithdrawalInput = {
  amount: Scalars['String'];
  fundingSourceID: Scalars['ID'];
};

export type WithdrawalMutationResponse = MutationResponse & {
  __typename?: 'WithdrawalMutationResponse';
  code: Scalars['String'];
  message: Scalars['String'];
  success: Scalars['Boolean'];
  transaction?: Maybe<Transaction>;
};

export type LinkUsdBankAccountMutationVariables = Exact<{
  input: LinkUsdBankAccountInput;
}>;


export type LinkUsdBankAccountMutation = { __typename?: 'Mutation', linkUsdBankAccount: { __typename?: 'LinkFundingSourceMutationResponse', code: string, success: boolean, message: string, fundingSource?: { __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined } };

export type GetFundingSourcesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetFundingSourcesQuery = { __typename?: 'Query', fundingSources: Array<{ __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined> };


export const LinkUsdBankAccountDocument = gql`
    mutation LinkUsdBankAccount($input: LinkUsdBankAccountInput!) {
  linkUsdBankAccount(input: $input) {
    code
    success
    message
    fundingSource {
      id
      name
      verificationStatus
      mask
      type
      subType
    }
  }
}
    `;
export type LinkUsdBankAccountMutationFn = Apollo.MutationFunction<LinkUsdBankAccountMutation, LinkUsdBankAccountMutationVariables>;
export type LinkUsdBankAccountMutationResult = Apollo.MutationResult<LinkUsdBankAccountMutation>;
export type LinkUsdBankAccountMutationOptions = Apollo.BaseMutationOptions<LinkUsdBankAccountMutation, LinkUsdBankAccountMutationVariables>;
export const GetFundingSourcesDocument = gql`
    query GetFundingSources {
  fundingSources {
    id
    name
    verificationStatus
    mask
    type
    subType
  }
}
    `;
export type GetFundingSourcesQueryResult = Apollo.QueryResult<GetFundingSourcesQuery, GetFundingSourcesQueryVariables>;