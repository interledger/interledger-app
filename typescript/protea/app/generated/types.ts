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

export type Country = {
  __typename?: 'Country';
  id: Scalars['String'];
  name: Scalars['String'];
};

export type CreateAccountInput = {
  country: Scalars['String'];
  firstName: Scalars['String'];
  lastName: Scalars['String'];
  mobileNumber: Scalars['String'];
};

export type CreateAccountMutationResponse = MutationResponse & {
  __typename?: 'CreateAccountMutationResponse';
  account?: Maybe<Account>;
  code: Scalars['String'];
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
  country: Scalars['String'];
  email: Scalars['String'];
  firstName: Scalars['String'];
  id: Scalars['ID'];
  lastName: Scalars['String'];
  mobileNumber: Scalars['String'];
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
  createAccount: CreateAccountMutationResponse;
  initiateDeposit: DepositMutationResponse;
  initiateOutgoingPayment: OutgoingPaymentMutationResponse;
  initiateWithdrawal: WithdrawalMutationResponse;
  linkUsdBankAccount: LinkFundingSourceMutationResponse;
  onboardAccount: CreateAccountMutationResponse;
  verifyAccount: VerifyAccountMutationResponse;
  verifyUsdBankAccount: VerifyUsdBankAccountMutationResponse;
};


export type MutationCreateAccountArgs = {
  input: CreateAccountInput;
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


export type MutationVerifyAccountArgs = {
  input: VerifyAccountInput;
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
  countries: Array<Country>;
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

export type VerifyAccountInput = {
  Address: Array<Scalars['String']>;
  City: Scalars['String'];
  DateOfBirth: Scalars['String'];
  PostalCode: Scalars['String'];
  State: Scalars['String'];
  TaxIdNumber: Scalars['String'];
};

export type VerifyAccountMutationResponse = MutationResponse & {
  __typename?: 'VerifyAccountMutationResponse';
  account?: Maybe<Account>;
  code: Scalars['String'];
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

export type VerifyUsdBankAccountMutationVariables = Exact<{
  input: VerifyUsdBankAccountInput;
}>;


export type VerifyUsdBankAccountMutation = { __typename?: 'Mutation', verifyUsdBankAccount: { __typename?: 'VerifyUsdBankAccountMutationResponse', code: string, success: boolean, message: string, fundingSource?: { __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined } };

export type GetFundingSourcesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetFundingSourcesQuery = { __typename?: 'Query', fundingSources: Array<{ __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined> };

export type InitiateDepositMutationVariables = Exact<{
  input: DepositInput;
}>;


export type InitiateDepositMutation = { __typename?: 'Mutation', initiateDeposit: { __typename?: 'DepositMutationResponse', code: string, success: boolean, message: string, transaction?: { __typename?: 'Transaction', id: string, type: TransactionType, description: string, amount: string, timestamp: string, status: string } | null | undefined } };

export type InitiateWithdrawalMutationVariables = Exact<{
  input: WithdrawalInput;
}>;


export type InitiateWithdrawalMutation = { __typename?: 'Mutation', initiateWithdrawal: { __typename?: 'WithdrawalMutationResponse', code: string, success: boolean, message: string, transaction?: { __typename?: 'Transaction', id: string, type: TransactionType, description: string, amount: string, timestamp: string, status: string } | null | undefined } };

export type GetCountriesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetCountriesQuery = { __typename?: 'Query', countries: Array<{ __typename?: 'Country', id: string, name: string }> };

export type OnboardAccountMutationVariables = Exact<{ [key: string]: never; }>;


export type OnboardAccountMutation = { __typename?: 'Mutation', onboardAccount: { __typename?: 'CreateAccountMutationResponse', code: string, success: boolean, message: string, account?: { __typename?: 'Account', id: string } | null | undefined } };


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
export const VerifyUsdBankAccountDocument = gql`
    mutation VerifyUsdBankAccount($input: VerifyUsdBankAccountInput!) {
  verifyUsdBankAccount(input: $input) {
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
export type VerifyUsdBankAccountMutationFn = Apollo.MutationFunction<VerifyUsdBankAccountMutation, VerifyUsdBankAccountMutationVariables>;
export type VerifyUsdBankAccountMutationResult = Apollo.MutationResult<VerifyUsdBankAccountMutation>;
export type VerifyUsdBankAccountMutationOptions = Apollo.BaseMutationOptions<VerifyUsdBankAccountMutation, VerifyUsdBankAccountMutationVariables>;
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
export const InitiateDepositDocument = gql`
    mutation InitiateDeposit($input: DepositInput!) {
  initiateDeposit(input: $input) {
    code
    success
    message
    transaction {
      id
      type
      description
      amount
      timestamp
      status
    }
  }
}
    `;
export type InitiateDepositMutationFn = Apollo.MutationFunction<InitiateDepositMutation, InitiateDepositMutationVariables>;
export type InitiateDepositMutationResult = Apollo.MutationResult<InitiateDepositMutation>;
export type InitiateDepositMutationOptions = Apollo.BaseMutationOptions<InitiateDepositMutation, InitiateDepositMutationVariables>;
export const InitiateWithdrawalDocument = gql`
    mutation InitiateWithdrawal($input: WithdrawalInput!) {
  initiateWithdrawal(input: $input) {
    code
    success
    message
    transaction {
      id
      type
      description
      amount
      timestamp
      status
    }
  }
}
    `;
export type InitiateWithdrawalMutationFn = Apollo.MutationFunction<InitiateWithdrawalMutation, InitiateWithdrawalMutationVariables>;
export type InitiateWithdrawalMutationResult = Apollo.MutationResult<InitiateWithdrawalMutation>;
export type InitiateWithdrawalMutationOptions = Apollo.BaseMutationOptions<InitiateWithdrawalMutation, InitiateWithdrawalMutationVariables>;
export const GetCountriesDocument = gql`
    query GetCountries {
  countries {
    id
    name
  }
}
    `;
export type GetCountriesQueryResult = Apollo.QueryResult<GetCountriesQuery, GetCountriesQueryVariables>;
export const OnboardAccountDocument = gql`
    mutation OnboardAccount {
  onboardAccount {
    code
    success
    message
    account {
      id
    }
  }
}
    `;
export type OnboardAccountMutationFn = Apollo.MutationFunction<OnboardAccountMutation, OnboardAccountMutationVariables>;
export type OnboardAccountMutationResult = Apollo.MutationResult<OnboardAccountMutation>;
export type OnboardAccountMutationOptions = Apollo.BaseMutationOptions<OnboardAccountMutation, OnboardAccountMutationVariables>;