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

export type Mutation = {
  __typename?: 'Mutation';
  createOrganisation: OrganisationMutationResponse;
};


export type MutationCreateOrganisationArgs = {
  name: Scalars['String'];
};

export type MutationResponse = {
  code: Scalars['String'];
  message: Scalars['String'];
  success: Scalars['Boolean'];
};

export type Organisation = {
  __typename?: 'Organisation';
  id: Scalars['ID'];
  name: Scalars['String'];
  verified: Scalars['Boolean'];
};

export type OrganisationMutationResponse = {
  __typename?: 'OrganisationMutationResponse';
  code: Scalars['String'];
  message: Scalars['String'];
  organisation?: Maybe<Organisation>;
  success: Scalars['Boolean'];
};

export type Query = {
  __typename?: 'Query';
  organisation?: Maybe<Organisation>;
  organisations: Array<Maybe<Organisation>>;
};


export type QueryOrganisationArgs = {
  id: Scalars['ID'];
};

export type CreateOrganisationMutationVariables = Exact<{
  name: Scalars['String'];
}>;


export type CreateOrganisationMutation = { __typename?: 'Mutation', createOrganisation: { __typename?: 'OrganisationMutationResponse', code: string, success: boolean, message: string, organisation?: { __typename?: 'Organisation', id: string, name: string } | null | undefined } };

export type GetOrganisationsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetOrganisationsQuery = { __typename?: 'Query', organisations: Array<{ __typename?: 'Organisation', id: string, name: string } | null | undefined> };


export const CreateOrganisationDocument = gql`
    mutation CreateOrganisation($name: String!) {
  createOrganisation(name: $name) {
    code
    success
    message
    organisation {
      id
      name
    }
  }
}
    `;
export type CreateOrganisationMutationFn = Apollo.MutationFunction<CreateOrganisationMutation, CreateOrganisationMutationVariables>;
export type CreateOrganisationMutationResult = Apollo.MutationResult<CreateOrganisationMutation>;
export type CreateOrganisationMutationOptions = Apollo.BaseMutationOptions<CreateOrganisationMutation, CreateOrganisationMutationVariables>;
export const GetOrganisationsDocument = gql`
    query GetOrganisations {
  organisations {
    id
    name
  }
}
    `;
export type GetOrganisationsQueryResult = Apollo.QueryResult<GetOrganisationsQuery, GetOrganisationsQueryVariables>;