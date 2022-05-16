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

export type Deposit = {
  __typename?: 'Deposit';
  amount: Scalars['String'];
  id: Scalars['ID'];
  state: Scalars['String'];
  timestamp: Scalars['String'];
};

export type DepositInput = {
  amount: Scalars['String'];
  fundingSourceID: Scalars['ID'];
};

export type DepositMutationResponse = MutationResponse & {
  __typename?: 'DepositMutationResponse';
  code: Scalars['String'];
  deposit?: Maybe<Deposit>;
  message: Scalars['String'];
  success: Scalars['Boolean'];
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

export type InitiateOnboardingMutationResponse = MutationResponse & {
  __typename?: 'InitiateOnboardingMutationResponse';
  code: Scalars['String'];
  message: Scalars['String'];
  providerOnboarding: UnitProviderOnboarding;
  success: Scalars['Boolean'];
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
  initiateOnboarding: InitiateOnboardingMutationResponse;
  initiateOutgoingPayment: OutgoingPaymentMutationResponse;
  initiateWithdrawal: WithdrawalMutationResponse;
  linkUsdBankAccount: LinkFundingSourceMutationResponse;
  onboardAccount: OnboardingMutationResponse;
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

export type OnboardingMutationResponse = MutationResponse & {
  __typename?: 'OnboardingMutationResponse';
  account: Account;
  code: Scalars['String'];
  message: Scalars['String'];
  success: Scalars['Boolean'];
};

export type OutgoingPayment = {
  __typename?: 'OutgoingPayment';
  amount: Scalars['String'];
  destination: Scalars['String'];
  id: Scalars['ID'];
  state: Scalars['String'];
  timestamp: Scalars['String'];
};

export type OutgoingPaymentInput = {
  amount: Scalars['String'];
  to: Scalars['String'];
};

export type OutgoingPaymentMutationResponse = MutationResponse & {
  __typename?: 'OutgoingPaymentMutationResponse';
  code: Scalars['String'];
  message: Scalars['String'];
  outgoingPayment?: Maybe<OutgoingPayment>;
  success: Scalars['Boolean'];
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor: Scalars['String'];
  hasNextPage: Scalars['Boolean'];
  startCursor: Scalars['String'];
};

export type Pagination = {
  after: Scalars['String'];
  first: Scalars['Int'];
};

export type Query = {
  __typename?: 'Query';
  account?: Maybe<Account>;
  countries: Array<Country>;
  fundingSources: Array<Maybe<FundingSource>>;
  identity?: Maybe<Identity>;
  transaction: Transaction;
  transactions: TransactionsConnection;
};


export type QueryTransactionArgs = {
  id: Scalars['String'];
};


export type QueryTransactionsArgs = {
  input: Pagination;
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

export type TransactionEdge = {
  __typename?: 'TransactionEdge';
  cursor: Scalars['String'];
  node: Transaction;
};

export enum TransactionType {
  Deposit = 'DEPOSIT',
  Outgoingpayment = 'OUTGOINGPAYMENT',
  Sent = 'SENT',
  Withdrawal = 'WITHDRAWAL'
}

export type TransactionsConnection = {
  __typename?: 'TransactionsConnection';
  edges: Array<TransactionEdge>;
  pageInfo: PageInfo;
};

export type UnitProviderOnboarding = {
  __typename?: 'UnitProviderOnboarding';
  formUrl: Scalars['String'];
  id: Scalars['String'];
};

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

export type Withdrawal = {
  __typename?: 'Withdrawal';
  amount: Scalars['String'];
  id: Scalars['ID'];
  state: Scalars['String'];
  timestamp: Scalars['String'];
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
  withdrawal?: Maybe<Withdrawal>;
};

export type ActivityQueryVariables = Exact<{
  input: Pagination;
}>;


export type ActivityQuery = { __typename?: 'Query', transactions: { __typename?: 'TransactionsConnection', pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, startCursor: string, endCursor: string }, edges: Array<{ __typename?: 'TransactionEdge', cursor: string, node: { __typename?: 'Transaction', id: string, type: TransactionType, description: string, amount: string, timestamp: string, status: string } }> } };

export type ActivityTransactionQueryVariables = Exact<{
  id: Scalars['String'];
}>;


export type ActivityTransactionQuery = { __typename?: 'Query', transaction: { __typename?: 'Transaction', id: string, type: TransactionType, description: string, amount: string, timestamp: string, status: string } };

export type HomeQueryVariables = Exact<{ [key: string]: never; }>;


export type HomeQuery = { __typename?: 'Query', account?: { __typename?: 'Account', id: string, balance: string, recentTransactions: Array<{ __typename?: 'Transaction', id: string, type: TransactionType, description: string, amount: string, timestamp: string, status: string }> } | null | undefined };

export type SettingsPaymentMethodsQueryVariables = Exact<{ [key: string]: never; }>;


export type SettingsPaymentMethodsQuery = { __typename?: 'Query', fundingSources: Array<{ __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined> };

export type FlowsDepositPaymentMethodQueryVariables = Exact<{ [key: string]: never; }>;


export type FlowsDepositPaymentMethodQuery = { __typename?: 'Query', fundingSources: Array<{ __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined> };

export type InitiateDepositMutationVariables = Exact<{
  input: DepositInput;
}>;


export type InitiateDepositMutation = { __typename?: 'Mutation', initiateDeposit: { __typename?: 'DepositMutationResponse', code: string, success: boolean, message: string, deposit?: { __typename?: 'Deposit', id: string, state: string, amount: string, timestamp: string } | null | undefined } };

export type LinkUsdBankAccountMutationVariables = Exact<{
  input: LinkUsdBankAccountInput;
}>;


export type LinkUsdBankAccountMutation = { __typename?: 'Mutation', linkUsdBankAccount: { __typename?: 'LinkFundingSourceMutationResponse', code: string, success: boolean, message: string, fundingSource?: { __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined } };

export type VerifyUsdBankAccountMutationVariables = Exact<{
  input: VerifyUsdBankAccountInput;
}>;


export type VerifyUsdBankAccountMutation = { __typename?: 'Mutation', verifyUsdBankAccount: { __typename?: 'VerifyUsdBankAccountMutationResponse', code: string, success: boolean, message: string, fundingSource?: { __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined } };

export type InitiateOutgoingPaymentMutationVariables = Exact<{
  input: OutgoingPaymentInput;
}>;


export type InitiateOutgoingPaymentMutation = { __typename?: 'Mutation', initiateOutgoingPayment: { __typename?: 'OutgoingPaymentMutationResponse', code: string, success: boolean, message: string, outgoingPayment?: { __typename?: 'OutgoingPayment', id: string, destination: string, state: string, amount: string, timestamp: string } | null | undefined } };

export type FlowsWithdrawPaymentMethodQueryVariables = Exact<{ [key: string]: never; }>;


export type FlowsWithdrawPaymentMethodQuery = { __typename?: 'Query', fundingSources: Array<{ __typename?: 'FundingSource', id: string, name: string, verificationStatus: string, mask: string, type: string, subType: string } | null | undefined> };

export type InitiateWithdrawalMutationVariables = Exact<{
  input: WithdrawalInput;
}>;


export type InitiateWithdrawalMutation = { __typename?: 'Mutation', initiateWithdrawal: { __typename?: 'WithdrawalMutationResponse', code: string, success: boolean, message: string, withdrawal?: { __typename?: 'Withdrawal', id: string, timestamp: string, amount: string, state: string } | null | undefined } };

export type SignupQueryVariables = Exact<{ [key: string]: never; }>;


export type SignupQuery = { __typename?: 'Query', countries: Array<{ __typename?: 'Country', id: string, name: string }> };

export type InitiateOnboardingMutationVariables = Exact<{ [key: string]: never; }>;


export type InitiateOnboardingMutation = { __typename?: 'Mutation', initiateOnboarding: { __typename?: 'InitiateOnboardingMutationResponse', code: string, success: boolean, message: string, providerOnboarding: { __typename?: 'UnitProviderOnboarding', id: string, formUrl: string } } };


export const ActivityDocument = gql`
    query Activity($input: Pagination!) {
  transactions(input: $input) {
    pageInfo {
      hasNextPage
      startCursor
      endCursor
    }
    edges {
      node {
        id
        type
        description
        amount
        timestamp
        status
      }
      cursor
    }
  }
}
    `;
export type ActivityQueryResult = Apollo.QueryResult<ActivityQuery, ActivityQueryVariables>;
export const ActivityTransactionDocument = gql`
    query ActivityTransaction($id: String!) {
  transaction(id: $id) {
    id
    type
    description
    amount
    timestamp
    status
  }
}
    `;
export type ActivityTransactionQueryResult = Apollo.QueryResult<ActivityTransactionQuery, ActivityTransactionQueryVariables>;
export const HomeDocument = gql`
    query Home {
  account {
    id
    balance
    recentTransactions {
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
export type HomeQueryResult = Apollo.QueryResult<HomeQuery, HomeQueryVariables>;
export const SettingsPaymentMethodsDocument = gql`
    query SettingsPaymentMethods {
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
export type SettingsPaymentMethodsQueryResult = Apollo.QueryResult<SettingsPaymentMethodsQuery, SettingsPaymentMethodsQueryVariables>;
export const FlowsDepositPaymentMethodDocument = gql`
    query FlowsDepositPaymentMethod {
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
export type FlowsDepositPaymentMethodQueryResult = Apollo.QueryResult<FlowsDepositPaymentMethodQuery, FlowsDepositPaymentMethodQueryVariables>;
export const InitiateDepositDocument = gql`
    mutation InitiateDeposit($input: DepositInput!) {
  initiateDeposit(input: $input) {
    code
    success
    message
    deposit {
      id
      state
      amount
      timestamp
    }
  }
}
    `;
export type InitiateDepositMutationFn = Apollo.MutationFunction<InitiateDepositMutation, InitiateDepositMutationVariables>;
export type InitiateDepositMutationResult = Apollo.MutationResult<InitiateDepositMutation>;
export type InitiateDepositMutationOptions = Apollo.BaseMutationOptions<InitiateDepositMutation, InitiateDepositMutationVariables>;
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
export const InitiateOutgoingPaymentDocument = gql`
    mutation InitiateOutgoingPayment($input: OutgoingPaymentInput!) {
  initiateOutgoingPayment(input: $input) {
    code
    success
    message
    outgoingPayment {
      id
      destination
      state
      amount
      timestamp
    }
  }
}
    `;
export type InitiateOutgoingPaymentMutationFn = Apollo.MutationFunction<InitiateOutgoingPaymentMutation, InitiateOutgoingPaymentMutationVariables>;
export type InitiateOutgoingPaymentMutationResult = Apollo.MutationResult<InitiateOutgoingPaymentMutation>;
export type InitiateOutgoingPaymentMutationOptions = Apollo.BaseMutationOptions<InitiateOutgoingPaymentMutation, InitiateOutgoingPaymentMutationVariables>;
export const FlowsWithdrawPaymentMethodDocument = gql`
    query FlowsWithdrawPaymentMethod {
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
export type FlowsWithdrawPaymentMethodQueryResult = Apollo.QueryResult<FlowsWithdrawPaymentMethodQuery, FlowsWithdrawPaymentMethodQueryVariables>;
export const InitiateWithdrawalDocument = gql`
    mutation InitiateWithdrawal($input: WithdrawalInput!) {
  initiateWithdrawal(input: $input) {
    code
    success
    message
    withdrawal {
      id
      timestamp
      amount
      state
    }
  }
}
    `;
export type InitiateWithdrawalMutationFn = Apollo.MutationFunction<InitiateWithdrawalMutation, InitiateWithdrawalMutationVariables>;
export type InitiateWithdrawalMutationResult = Apollo.MutationResult<InitiateWithdrawalMutation>;
export type InitiateWithdrawalMutationOptions = Apollo.BaseMutationOptions<InitiateWithdrawalMutation, InitiateWithdrawalMutationVariables>;
export const SignupDocument = gql`
    query Signup {
  countries {
    id
    name
  }
}
    `;
export type SignupQueryResult = Apollo.QueryResult<SignupQuery, SignupQueryVariables>;
export const InitiateOnboardingDocument = gql`
    mutation InitiateOnboarding {
  initiateOnboarding {
    code
    success
    message
    providerOnboarding {
      id
      formUrl
    }
  }
}
    `;
export type InitiateOnboardingMutationFn = Apollo.MutationFunction<InitiateOnboardingMutation, InitiateOnboardingMutationVariables>;
export type InitiateOnboardingMutationResult = Apollo.MutationResult<InitiateOnboardingMutation>;
export type InitiateOnboardingMutationOptions = Apollo.BaseMutationOptions<InitiateOnboardingMutation, InitiateOnboardingMutationVariables>;