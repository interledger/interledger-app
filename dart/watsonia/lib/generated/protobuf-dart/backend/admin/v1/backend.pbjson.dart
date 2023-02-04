///
//  Generated code. Do not modify.
//  source: backend/admin/v1/backend.proto
//
// @dart = 2.12
// ignore_for_file: annotate_overrides,camel_case_types,constant_identifier_names,deprecated_member_use_from_same_package,directives_ordering,library_prefixes,non_constant_identifier_names,prefer_final_fields,return_of_invalid_type,unnecessary_const,unnecessary_import,unnecessary_this,unused_import,unused_shown_name

import 'dart:core' as $core;
import 'dart:convert' as $convert;
import 'dart:typed_data' as $typed_data;
@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = const {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor = $convert.base64Decode('CgVFbXB0eQ==');
@$core.Deprecated('Use emailWalletStatementRequestDescriptor instead')
const EmailWalletStatementRequest$json = const {
  '1': 'EmailWalletStatementRequest',
  '2': const [
    const {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    const {'1': 'period', '3': 2, '4': 1, '5': 9, '10': 'period'},
  ],
};

/// Descriptor for `EmailWalletStatementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emailWalletStatementRequestDescriptor = $convert.base64Decode('ChtFbWFpbFdhbGxldFN0YXRlbWVudFJlcXVlc3QSGgoId2FsbGV0SUQYASABKAlSCHdhbGxldElEEhYKBnBlcmlvZBgCIAEoCVIGcGVyaW9k');
@$core.Deprecated('Use listWalletTransactionsRequestDescriptor instead')
const ListWalletTransactionsRequest$json = const {
  '1': 'ListWalletTransactionsRequest',
  '2': const [
    const {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    const {'1': 'page', '3': 2, '4': 1, '5': 11, '6': '.backend.admin.v1.PaginationRequest', '10': 'page'},
  ],
};

/// Descriptor for `ListWalletTransactionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWalletTransactionsRequestDescriptor = $convert.base64Decode('Ch1MaXN0V2FsbGV0VHJhbnNhY3Rpb25zUmVxdWVzdBIaCgh3YWxsZXRJRBgBIAEoCVIId2FsbGV0SUQSNwoEcGFnZRgCIAEoCzIjLmJhY2tlbmQuYWRtaW4udjEuUGFnaW5hdGlvblJlcXVlc3RSBHBhZ2U=');
@$core.Deprecated('Use listWalletTransactionsResponseDescriptor instead')
const ListWalletTransactionsResponse$json = const {
  '1': 'ListWalletTransactionsResponse',
  '2': const [
    const {'1': 'transactions', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.Transaction', '10': 'transactions'},
    const {'1': 'page', '3': 2, '4': 1, '5': 11, '6': '.backend.admin.v1.PaginationResponse', '10': 'page'},
  ],
};

/// Descriptor for `ListWalletTransactionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWalletTransactionsResponseDescriptor = $convert.base64Decode('Ch5MaXN0V2FsbGV0VHJhbnNhY3Rpb25zUmVzcG9uc2USQQoMdHJhbnNhY3Rpb25zGAEgAygLMh0uYmFja2VuZC5hZG1pbi52MS5UcmFuc2FjdGlvblIMdHJhbnNhY3Rpb25zEjgKBHBhZ2UYAiABKAsyJC5iYWNrZW5kLmFkbWluLnYxLlBhZ2luYXRpb25SZXNwb25zZVIEcGFnZQ==');
@$core.Deprecated('Use transactionDescriptor instead')
const Transaction$json = const {
  '1': 'Transaction',
  '2': const [
    const {'1': 'walletID', '3': 8, '4': 1, '5': 9, '10': 'walletID'},
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    const {'1': 'asset', '3': 7, '4': 1, '5': 9, '10': 'asset'},
    const {'1': 'amount', '3': 3, '4': 1, '5': 1, '10': 'amount'},
    const {'1': 'source', '3': 4, '4': 1, '5': 9, '10': 'source'},
    const {'1': 'destination', '3': 5, '4': 1, '5': 9, '10': 'destination'},
    const {'1': 'timestamp', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
  ],
};

/// Descriptor for `Transaction`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transactionDescriptor = $convert.base64Decode('CgtUcmFuc2FjdGlvbhIaCgh3YWxsZXRJRBgIIAEoCVIId2FsbGV0SUQSDgoCaWQYASABKAlSAmlkEhIKBHR5cGUYAiABKAlSBHR5cGUSFAoFYXNzZXQYByABKAlSBWFzc2V0EhYKBmFtb3VudBgDIAEoAVIGYW1vdW50EhYKBnNvdXJjZRgEIAEoCVIGc291cmNlEiAKC2Rlc3RpbmF0aW9uGAUgASgJUgtkZXN0aW5hdGlvbhI4Cgl0aW1lc3RhbXAYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl0aW1lc3RhbXA=');
@$core.Deprecated('Use getUserTransactionsRequestDescriptor instead')
const GetUserTransactionsRequest$json = const {
  '1': 'GetUserTransactionsRequest',
  '2': const [
    const {'1': 'userID', '3': 1, '4': 1, '5': 9, '10': 'userID'},
  ],
};

/// Descriptor for `GetUserTransactionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getUserTransactionsRequestDescriptor = $convert.base64Decode('ChpHZXRVc2VyVHJhbnNhY3Rpb25zUmVxdWVzdBIWCgZ1c2VySUQYASABKAlSBnVzZXJJRA==');
@$core.Deprecated('Use getWalletDetailsRequestDescriptor instead')
const GetWalletDetailsRequest$json = const {
  '1': 'GetWalletDetailsRequest',
  '2': const [
    const {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
  ],
};

/// Descriptor for `GetWalletDetailsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWalletDetailsRequestDescriptor = $convert.base64Decode('ChdHZXRXYWxsZXREZXRhaWxzUmVxdWVzdBIaCgh3YWxsZXRJRBgBIAEoCVIId2FsbGV0SUQ=');
@$core.Deprecated('Use walletDetailsDescriptor instead')
const WalletDetails$json = const {
  '1': 'WalletDetails',
  '2': const [
    const {'1': 'users', '3': 8, '4': 3, '5': 11, '6': '.backend.admin.v1.User', '10': 'users'},
    const {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    const {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    const {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    const {'1': 'countryCode', '3': 4, '4': 1, '5': 9, '10': 'countryCode'},
    const {'1': 'gender', '3': 5, '4': 1, '5': 5, '10': 'gender'},
    const {'1': 'dateOfBirth', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'dateOfBirth'},
    const {'1': 'address', '3': 7, '4': 1, '5': 9, '10': 'address'},
  ],
};

/// Descriptor for `WalletDetails`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletDetailsDescriptor = $convert.base64Decode('Cg1XYWxsZXREZXRhaWxzEiwKBXVzZXJzGAggAygLMhYuYmFja2VuZC5hZG1pbi52MS5Vc2VyUgV1c2VycxIaCgh3YWxsZXRJRBgBIAEoCVIId2FsbGV0SUQSHAoJZmlyc3ROYW1lGAIgASgJUglmaXJzdE5hbWUSGgoIbGFzdE5hbWUYAyABKAlSCGxhc3ROYW1lEiAKC2NvdW50cnlDb2RlGAQgASgJUgtjb3VudHJ5Q29kZRIWCgZnZW5kZXIYBSABKAVSBmdlbmRlchI8CgtkYXRlT2ZCaXJ0aBgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSC2RhdGVPZkJpcnRoEhgKB2FkZHJlc3MYByABKAlSB2FkZHJlc3M=');
@$core.Deprecated('Use paginationRequestDescriptor instead')
const PaginationRequest$json = const {
  '1': 'PaginationRequest',
  '2': const [
    const {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    const {'1': 'pageSize', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `PaginationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paginationRequestDescriptor = $convert.base64Decode('ChFQYWdpbmF0aW9uUmVxdWVzdBISCgRwYWdlGAEgASgFUgRwYWdlEhoKCHBhZ2VTaXplGAIgASgFUghwYWdlU2l6ZQ==');
@$core.Deprecated('Use paginationResponseDescriptor instead')
const PaginationResponse$json = const {
  '1': 'PaginationResponse',
  '2': const [
    const {'1': 'page', '3': 1, '4': 1, '5': 5, '10': 'page'},
    const {'1': 'pageSize', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    const {'1': 'hasNextPage', '3': 3, '4': 1, '5': 8, '10': 'hasNextPage'},
  ],
};

/// Descriptor for `PaginationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paginationResponseDescriptor = $convert.base64Decode('ChJQYWdpbmF0aW9uUmVzcG9uc2USEgoEcGFnZRgBIAEoBVIEcGFnZRIaCghwYWdlU2l6ZRgCIAEoBVIIcGFnZVNpemUSIAoLaGFzTmV4dFBhZ2UYAyABKAhSC2hhc05leHRQYWdl');
@$core.Deprecated('Use walletDescriptor instead')
const Wallet$json = const {
  '1': 'Wallet',
  '2': const [
    const {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    const {'1': 'walletName', '3': 2, '4': 1, '5': 9, '10': 'walletName'},
    const {'1': 'users', '3': 3, '4': 3, '5': 11, '6': '.backend.admin.v1.User', '10': 'users'},
  ],
};

/// Descriptor for `Wallet`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletDescriptor = $convert.base64Decode('CgZXYWxsZXQSGgoId2FsbGV0SUQYASABKAlSCHdhbGxldElEEh4KCndhbGxldE5hbWUYAiABKAlSCndhbGxldE5hbWUSLAoFdXNlcnMYAyADKAsyFi5iYWNrZW5kLmFkbWluLnYxLlVzZXJSBXVzZXJz');
@$core.Deprecated('Use listWalletsResponseDescriptor instead')
const ListWalletsResponse$json = const {
  '1': 'ListWalletsResponse',
  '2': const [
    const {'1': 'wallets', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.Wallet', '10': 'wallets'},
    const {'1': 'page', '3': 2, '4': 1, '5': 11, '6': '.backend.admin.v1.PaginationResponse', '10': 'page'},
  ],
};

/// Descriptor for `ListWalletsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWalletsResponseDescriptor = $convert.base64Decode('ChNMaXN0V2FsbGV0c1Jlc3BvbnNlEjIKB3dhbGxldHMYASADKAsyGC5iYWNrZW5kLmFkbWluLnYxLldhbGxldFIHd2FsbGV0cxI4CgRwYWdlGAIgASgLMiQuYmFja2VuZC5hZG1pbi52MS5QYWdpbmF0aW9uUmVzcG9uc2VSBHBhZ2U=');
@$core.Deprecated('Use userDescriptor instead')
const User$json = const {
  '1': 'User',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'email', '3': 2, '4': 1, '5': 9, '10': 'email'},
    const {'1': 'phoneNumber', '3': 3, '4': 1, '5': 9, '10': 'phoneNumber'},
  ],
};

/// Descriptor for `User`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userDescriptor = $convert.base64Decode('CgRVc2VyEg4KAmlkGAEgASgJUgJpZBIUCgVlbWFpbBgCIAEoCVIFZW1haWwSIAoLcGhvbmVOdW1iZXIYAyABKAlSC3Bob25lTnVtYmVy');
@$core.Deprecated('Use allowWaitlistSignupRequestDescriptor instead')
const AllowWaitlistSignupRequest$json = const {
  '1': 'AllowWaitlistSignupRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `AllowWaitlistSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allowWaitlistSignupRequestDescriptor = $convert.base64Decode('ChpBbGxvd1dhaXRsaXN0U2lnbnVwUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');
@$core.Deprecated('Use listWaitlistSignupsResponseDescriptor instead')
const ListWaitlistSignupsResponse$json = const {
  '1': 'ListWaitlistSignupsResponse',
  '2': const [
    const {'1': 'signups', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.WaitlistSignup', '10': 'signups'},
  ],
};

/// Descriptor for `ListWaitlistSignupsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWaitlistSignupsResponseDescriptor = $convert.base64Decode('ChtMaXN0V2FpdGxpc3RTaWdudXBzUmVzcG9uc2USOgoHc2lnbnVwcxgBIAMoCzIgLmJhY2tlbmQuYWRtaW4udjEuV2FpdGxpc3RTaWdudXBSB3NpZ251cHM=');
@$core.Deprecated('Use waitlistSignupDescriptor instead')
const WaitlistSignup$json = const {
  '1': 'WaitlistSignup',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    const {'1': 'email', '3': 3, '4': 1, '5': 9, '10': 'email'},
    const {'1': 'beta_opt_in', '3': 4, '4': 1, '5': 8, '10': 'betaOptIn'},
    const {'1': 'can_signup', '3': 5, '4': 1, '5': 8, '10': 'canSignup'},
    const {'1': 'mug_id', '3': 6, '4': 1, '5': 9, '10': 'mugId'},
    const {'1': 'country_code', '3': 7, '4': 1, '5': 9, '10': 'countryCode'},
  ],
};

/// Descriptor for `WaitlistSignup`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitlistSignupDescriptor = $convert.base64Decode('Cg5XYWl0bGlzdFNpZ251cBIOCgJpZBgBIAEoCVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZRIUCgVlbWFpbBgDIAEoCVIFZW1haWwSHgoLYmV0YV9vcHRfaW4YBCABKAhSCWJldGFPcHRJbhIdCgpjYW5fc2lnbnVwGAUgASgIUgljYW5TaWdudXASFQoGbXVnX2lkGAYgASgJUgVtdWdJZBIhCgxjb3VudHJ5X2NvZGUYByABKAlSC2NvdW50cnlDb2Rl');
