//
//  Generated code. Do not modify.
//  source: pacioli/v1/pacioli.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor = $convert.base64Decode(
    'CgVFbXB0eQ==');

@$core.Deprecated('Use ledgerDescriptor instead')
const Ledger$json = {
  '1': 'Ledger',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'asset', '3': 3, '4': 1, '5': 9, '10': 'asset'},
    {'1': 'scale', '3': 4, '4': 1, '5': 13, '10': 'scale'},
  ],
};

/// Descriptor for `Ledger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ledgerDescriptor = $convert.base64Decode(
    'CgZMZWRnZXISDgoCaWQYASABKA1SAmlkEhIKBG5hbWUYAiABKAlSBG5hbWUSFAoFYXNzZXQYAy'
    'ABKAlSBWFzc2V0EhQKBXNjYWxlGAQgASgNUgVzY2FsZQ==');

@$core.Deprecated('Use configureLedgersRequestDescriptor instead')
const ConfigureLedgersRequest$json = {
  '1': 'ConfigureLedgersRequest',
  '2': [
    {'1': 'args', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Ledger', '10': 'args'},
  ],
};

/// Descriptor for `ConfigureLedgersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureLedgersRequestDescriptor = $convert.base64Decode(
    'ChdDb25maWd1cmVMZWRnZXJzUmVxdWVzdBImCgRhcmdzGAEgAygLMhIucGFjaW9saS52MS5MZW'
    'RnZXJSBGFyZ3M=');

@$core.Deprecated('Use configureLedgersResponseDescriptor instead')
const ConfigureLedgersResponse$json = {
  '1': 'ConfigureLedgersResponse',
  '2': [
    {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `ConfigureLedgersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureLedgersResponseDescriptor = $convert.base64Decode(
    'ChhDb25maWd1cmVMZWRnZXJzUmVzcG9uc2USLgoGZXJyb3JzGAEgAygLMhYucGFjaW9saS52MS'
    '5FdmVudEVycm9yUgZlcnJvcnM=');

@$core.Deprecated('Use getLedgersRequestDescriptor instead')
const GetLedgersRequest$json = {
  '1': 'GetLedgersRequest',
  '2': [
    {'1': 'ids', '3': 1, '4': 3, '5': 13, '10': 'ids'},
  ],
};

/// Descriptor for `GetLedgersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLedgersRequestDescriptor = $convert.base64Decode(
    'ChFHZXRMZWRnZXJzUmVxdWVzdBIQCgNpZHMYASADKA1SA2lkcw==');

@$core.Deprecated('Use getLedgersResponseDescriptor instead')
const GetLedgersResponse$json = {
  '1': 'GetLedgersResponse',
  '2': [
    {'1': 'ledgers', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Ledger', '10': 'ledgers'},
  ],
};

/// Descriptor for `GetLedgersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLedgersResponseDescriptor = $convert.base64Decode(
    'ChJHZXRMZWRnZXJzUmVzcG9uc2USLAoHbGVkZ2VycxgBIAMoCzISLnBhY2lvbGkudjEuTGVkZ2'
    'VyUgdsZWRnZXJz');

@$core.Deprecated('Use accountDescriptor instead')
const Account$json = {
  '1': 'Account',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'ledgerId', '3': 2, '4': 1, '5': 13, '10': 'ledgerId'},
    {'1': 'code', '3': 3, '4': 1, '5': 13, '10': 'code'},
    {'1': 'debitsReserved', '3': 4, '4': 1, '5': 4, '10': 'debitsReserved'},
    {'1': 'debitsAccepted', '3': 5, '4': 1, '5': 4, '10': 'debitsAccepted'},
    {'1': 'creditsReserved', '3': 6, '4': 1, '5': 4, '10': 'creditsReserved'},
    {'1': 'creditsAccepted', '3': 7, '4': 1, '5': 4, '10': 'creditsAccepted'},
    {'1': 'debitsMustNotExceedCredits', '3': 8, '4': 1, '5': 8, '10': 'debitsMustNotExceedCredits'},
    {'1': 'creditsMustNotExceedDebits', '3': 9, '4': 1, '5': 8, '10': 'creditsMustNotExceedDebits'},
  ],
};

/// Descriptor for `Account`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountDescriptor = $convert.base64Decode(
    'CgdBY2NvdW50Eg4KAmlkGAEgASgJUgJpZBIaCghsZWRnZXJJZBgCIAEoDVIIbGVkZ2VySWQSEg'
    'oEY29kZRgDIAEoDVIEY29kZRImCg5kZWJpdHNSZXNlcnZlZBgEIAEoBFIOZGViaXRzUmVzZXJ2'
    'ZWQSJgoOZGViaXRzQWNjZXB0ZWQYBSABKARSDmRlYml0c0FjY2VwdGVkEigKD2NyZWRpdHNSZX'
    'NlcnZlZBgGIAEoBFIPY3JlZGl0c1Jlc2VydmVkEigKD2NyZWRpdHNBY2NlcHRlZBgHIAEoBFIP'
    'Y3JlZGl0c0FjY2VwdGVkEj4KGmRlYml0c011c3ROb3RFeGNlZWRDcmVkaXRzGAggASgIUhpkZW'
    'JpdHNNdXN0Tm90RXhjZWVkQ3JlZGl0cxI+ChpjcmVkaXRzTXVzdE5vdEV4Y2VlZERlYml0cxgJ'
    'IAEoCFIaY3JlZGl0c011c3ROb3RFeGNlZWREZWJpdHM=');

@$core.Deprecated('Use configureAccountsArgsDescriptor instead')
const ConfigureAccountsArgs$json = {
  '1': 'ConfigureAccountsArgs',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'ledgerId', '3': 2, '4': 1, '5': 13, '10': 'ledgerId'},
    {'1': 'code', '3': 3, '4': 1, '5': 13, '10': 'code'},
    {'1': 'debitsMustNotExceedCredits', '3': 4, '4': 1, '5': 8, '10': 'debitsMustNotExceedCredits'},
    {'1': 'creditsMustNotExceedDebits', '3': 5, '4': 1, '5': 8, '10': 'creditsMustNotExceedDebits'},
  ],
};

/// Descriptor for `ConfigureAccountsArgs`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureAccountsArgsDescriptor = $convert.base64Decode(
    'ChVDb25maWd1cmVBY2NvdW50c0FyZ3MSDgoCaWQYASABKAlSAmlkEhoKCGxlZGdlcklkGAIgAS'
    'gNUghsZWRnZXJJZBISCgRjb2RlGAMgASgNUgRjb2RlEj4KGmRlYml0c011c3ROb3RFeGNlZWRD'
    'cmVkaXRzGAQgASgIUhpkZWJpdHNNdXN0Tm90RXhjZWVkQ3JlZGl0cxI+ChpjcmVkaXRzTXVzdE'
    '5vdEV4Y2VlZERlYml0cxgFIAEoCFIaY3JlZGl0c011c3ROb3RFeGNlZWREZWJpdHM=');

@$core.Deprecated('Use configureAccountsRequestDescriptor instead')
const ConfigureAccountsRequest$json = {
  '1': 'ConfigureAccountsRequest',
  '2': [
    {'1': 'args', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.ConfigureAccountsArgs', '10': 'args'},
  ],
};

/// Descriptor for `ConfigureAccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureAccountsRequestDescriptor = $convert.base64Decode(
    'ChhDb25maWd1cmVBY2NvdW50c1JlcXVlc3QSNQoEYXJncxgBIAMoCzIhLnBhY2lvbGkudjEuQ2'
    '9uZmlndXJlQWNjb3VudHNBcmdzUgRhcmdz');

@$core.Deprecated('Use configureAccountsResponseDescriptor instead')
const ConfigureAccountsResponse$json = {
  '1': 'ConfigureAccountsResponse',
  '2': [
    {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `ConfigureAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureAccountsResponseDescriptor = $convert.base64Decode(
    'ChlDb25maWd1cmVBY2NvdW50c1Jlc3BvbnNlEi4KBmVycm9ycxgBIAMoCzIWLnBhY2lvbGkudj'
    'EuRXZlbnRFcnJvclIGZXJyb3Jz');

@$core.Deprecated('Use getAccountsRequestDescriptor instead')
const GetAccountsRequest$json = {
  '1': 'GetAccountsRequest',
  '2': [
    {'1': 'ids', '3': 1, '4': 3, '5': 9, '10': 'ids'},
  ],
};

/// Descriptor for `GetAccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAccountsRequestDescriptor = $convert.base64Decode(
    'ChJHZXRBY2NvdW50c1JlcXVlc3QSEAoDaWRzGAEgAygJUgNpZHM=');

@$core.Deprecated('Use getAccountsResponseDescriptor instead')
const GetAccountsResponse$json = {
  '1': 'GetAccountsResponse',
  '2': [
    {'1': 'accounts', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Account', '10': 'accounts'},
  ],
};

/// Descriptor for `GetAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAccountsResponseDescriptor = $convert.base64Decode(
    'ChNHZXRBY2NvdW50c1Jlc3BvbnNlEi8KCGFjY291bnRzGAEgAygLMhMucGFjaW9saS52MS5BY2'
    'NvdW50UghhY2NvdW50cw==');

@$core.Deprecated('Use transferDescriptor instead')
const Transfer$json = {
  '1': 'Transfer',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'debitAccountId', '3': 2, '4': 1, '5': 9, '10': 'debitAccountId'},
    {'1': 'creditAccountId', '3': 3, '4': 1, '5': 9, '10': 'creditAccountId'},
    {'1': 'amount', '3': 4, '4': 1, '5': 4, '10': 'amount'},
    {'1': 'code', '3': 5, '4': 1, '5': 13, '10': 'code'},
    {'1': 'timeout', '3': 7, '4': 1, '5': 4, '10': 'timeout'},
    {'1': 'ledger', '3': 8, '4': 1, '5': 13, '10': 'ledger'},
    {'1': 'pendingId', '3': 9, '4': 1, '5': 9, '10': 'pendingId'},
    {'1': 'pending', '3': 10, '4': 1, '5': 8, '10': 'pending'},
  ],
};

/// Descriptor for `Transfer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transferDescriptor = $convert.base64Decode(
    'CghUcmFuc2ZlchIOCgJpZBgBIAEoCVICaWQSJgoOZGViaXRBY2NvdW50SWQYAiABKAlSDmRlYm'
    'l0QWNjb3VudElkEigKD2NyZWRpdEFjY291bnRJZBgDIAEoCVIPY3JlZGl0QWNjb3VudElkEhYK'
    'BmFtb3VudBgEIAEoBFIGYW1vdW50EhIKBGNvZGUYBSABKA1SBGNvZGUSGAoHdGltZW91dBgHIA'
    'EoBFIHdGltZW91dBIWCgZsZWRnZXIYCCABKA1SBmxlZGdlchIcCglwZW5kaW5nSWQYCSABKAlS'
    'CXBlbmRpbmdJZBIYCgdwZW5kaW5nGAogASgIUgdwZW5kaW5n');

@$core.Deprecated('Use createTransfersRequestDescriptor instead')
const CreateTransfersRequest$json = {
  '1': 'CreateTransfersRequest',
  '2': [
    {'1': 'transfers', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Transfer', '10': 'transfers'},
  ],
};

/// Descriptor for `CreateTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTransfersRequestDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVUcmFuc2ZlcnNSZXF1ZXN0EjIKCXRyYW5zZmVycxgBIAMoCzIULnBhY2lvbGkudj'
    'EuVHJhbnNmZXJSCXRyYW5zZmVycw==');

@$core.Deprecated('Use eventErrorDescriptor instead')
const EventError$json = {
  '1': 'EventError',
  '2': [
    {'1': 'index', '3': 1, '4': 1, '5': 13, '10': 'index'},
    {'1': 'code', '3': 2, '4': 1, '5': 13, '10': 'code'},
  ],
};

/// Descriptor for `EventError`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List eventErrorDescriptor = $convert.base64Decode(
    'CgpFdmVudEVycm9yEhQKBWluZGV4GAEgASgNUgVpbmRleBISCgRjb2RlGAIgASgNUgRjb2Rl');

@$core.Deprecated('Use createTransfersResponseDescriptor instead')
const CreateTransfersResponse$json = {
  '1': 'CreateTransfersResponse',
  '2': [
    {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `CreateTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTransfersResponseDescriptor = $convert.base64Decode(
    'ChdDcmVhdGVUcmFuc2ZlcnNSZXNwb25zZRIuCgZlcnJvcnMYASADKAsyFi5wYWNpb2xpLnYxLk'
    'V2ZW50RXJyb3JSBmVycm9ycw==');

@$core.Deprecated('Use getTransfersRequestDescriptor instead')
const GetTransfersRequest$json = {
  '1': 'GetTransfersRequest',
  '2': [
    {'1': 'ids', '3': 1, '4': 3, '5': 9, '10': 'ids'},
  ],
};

/// Descriptor for `GetTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTransfersRequestDescriptor = $convert.base64Decode(
    'ChNHZXRUcmFuc2ZlcnNSZXF1ZXN0EhAKA2lkcxgBIAMoCVIDaWRz');

@$core.Deprecated('Use getTransfersResponseDescriptor instead')
const GetTransfersResponse$json = {
  '1': 'GetTransfersResponse',
  '2': [
    {'1': 'transfers', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Transfer', '10': 'transfers'},
  ],
};

/// Descriptor for `GetTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTransfersResponseDescriptor = $convert.base64Decode(
    'ChRHZXRUcmFuc2ZlcnNSZXNwb25zZRIyCgl0cmFuc2ZlcnMYASADKAsyFC5wYWNpb2xpLnYxLl'
    'RyYW5zZmVyUgl0cmFuc2ZlcnM=');

@$core.Deprecated('Use postTransfersRequestDescriptor instead')
const PostTransfersRequest$json = {
  '1': 'PostTransfersRequest',
  '2': [
    {'1': 'transferIds', '3': 1, '4': 3, '5': 9, '10': 'transferIds'},
  ],
};

/// Descriptor for `PostTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List postTransfersRequestDescriptor = $convert.base64Decode(
    'ChRQb3N0VHJhbnNmZXJzUmVxdWVzdBIgCgt0cmFuc2ZlcklkcxgBIAMoCVILdHJhbnNmZXJJZH'
    'M=');

@$core.Deprecated('Use postTransfersResponseDescriptor instead')
const PostTransfersResponse$json = {
  '1': 'PostTransfersResponse',
  '2': [
    {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `PostTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List postTransfersResponseDescriptor = $convert.base64Decode(
    'ChVQb3N0VHJhbnNmZXJzUmVzcG9uc2USLgoGZXJyb3JzGAEgAygLMhYucGFjaW9saS52MS5Fdm'
    'VudEVycm9yUgZlcnJvcnM=');

@$core.Deprecated('Use voidTransfersRequestDescriptor instead')
const VoidTransfersRequest$json = {
  '1': 'VoidTransfersRequest',
  '2': [
    {'1': 'transferIds', '3': 1, '4': 3, '5': 9, '10': 'transferIds'},
  ],
};

/// Descriptor for `VoidTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidTransfersRequestDescriptor = $convert.base64Decode(
    'ChRWb2lkVHJhbnNmZXJzUmVxdWVzdBIgCgt0cmFuc2ZlcklkcxgBIAMoCVILdHJhbnNmZXJJZH'
    'M=');

@$core.Deprecated('Use voidTransfersResponseDescriptor instead')
const VoidTransfersResponse$json = {
  '1': 'VoidTransfersResponse',
  '2': [
    {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `VoidTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidTransfersResponseDescriptor = $convert.base64Decode(
    'ChVWb2lkVHJhbnNmZXJzUmVzcG9uc2USLgoGZXJyb3JzGAEgAygLMhYucGFjaW9saS52MS5Fdm'
    'VudEVycm9yUgZlcnJvcnM=');

