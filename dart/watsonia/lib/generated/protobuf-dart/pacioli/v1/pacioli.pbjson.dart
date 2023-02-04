///
//  Generated code. Do not modify.
//  source: pacioli/v1/pacioli.proto
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
@$core.Deprecated('Use ledgerDescriptor instead')
const Ledger$json = const {
  '1': 'Ledger',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    const {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    const {'1': 'asset', '3': 3, '4': 1, '5': 9, '10': 'asset'},
    const {'1': 'scale', '3': 4, '4': 1, '5': 13, '10': 'scale'},
  ],
};

/// Descriptor for `Ledger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ledgerDescriptor = $convert.base64Decode('CgZMZWRnZXISDgoCaWQYASABKA1SAmlkEhIKBG5hbWUYAiABKAlSBG5hbWUSFAoFYXNzZXQYAyABKAlSBWFzc2V0EhQKBXNjYWxlGAQgASgNUgVzY2FsZQ==');
@$core.Deprecated('Use configureLedgersRequestDescriptor instead')
const ConfigureLedgersRequest$json = const {
  '1': 'ConfigureLedgersRequest',
  '2': const [
    const {'1': 'args', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Ledger', '10': 'args'},
  ],
};

/// Descriptor for `ConfigureLedgersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureLedgersRequestDescriptor = $convert.base64Decode('ChdDb25maWd1cmVMZWRnZXJzUmVxdWVzdBImCgRhcmdzGAEgAygLMhIucGFjaW9saS52MS5MZWRnZXJSBGFyZ3M=');
@$core.Deprecated('Use configureLedgersResponseDescriptor instead')
const ConfigureLedgersResponse$json = const {
  '1': 'ConfigureLedgersResponse',
  '2': const [
    const {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `ConfigureLedgersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureLedgersResponseDescriptor = $convert.base64Decode('ChhDb25maWd1cmVMZWRnZXJzUmVzcG9uc2USLgoGZXJyb3JzGAEgAygLMhYucGFjaW9saS52MS5FdmVudEVycm9yUgZlcnJvcnM=');
@$core.Deprecated('Use getLedgersRequestDescriptor instead')
const GetLedgersRequest$json = const {
  '1': 'GetLedgersRequest',
  '2': const [
    const {'1': 'ids', '3': 1, '4': 3, '5': 13, '10': 'ids'},
  ],
};

/// Descriptor for `GetLedgersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLedgersRequestDescriptor = $convert.base64Decode('ChFHZXRMZWRnZXJzUmVxdWVzdBIQCgNpZHMYASADKA1SA2lkcw==');
@$core.Deprecated('Use getLedgersResponseDescriptor instead')
const GetLedgersResponse$json = const {
  '1': 'GetLedgersResponse',
  '2': const [
    const {'1': 'ledgers', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Ledger', '10': 'ledgers'},
  ],
};

/// Descriptor for `GetLedgersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLedgersResponseDescriptor = $convert.base64Decode('ChJHZXRMZWRnZXJzUmVzcG9uc2USLAoHbGVkZ2VycxgBIAMoCzISLnBhY2lvbGkudjEuTGVkZ2VyUgdsZWRnZXJz');
@$core.Deprecated('Use accountDescriptor instead')
const Account$json = const {
  '1': 'Account',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'ledgerId', '3': 2, '4': 1, '5': 13, '10': 'ledgerId'},
    const {'1': 'code', '3': 3, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'debitsReserved', '3': 4, '4': 1, '5': 4, '10': 'debitsReserved'},
    const {'1': 'debitsAccepted', '3': 5, '4': 1, '5': 4, '10': 'debitsAccepted'},
    const {'1': 'creditsReserved', '3': 6, '4': 1, '5': 4, '10': 'creditsReserved'},
    const {'1': 'creditsAccepted', '3': 7, '4': 1, '5': 4, '10': 'creditsAccepted'},
    const {'1': 'flags', '3': 8, '4': 1, '5': 11, '6': '.pacioli.v1.AccountFlags', '10': 'flags'},
  ],
};

/// Descriptor for `Account`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountDescriptor = $convert.base64Decode('CgdBY2NvdW50Eg4KAmlkGAEgASgJUgJpZBIaCghsZWRnZXJJZBgCIAEoDVIIbGVkZ2VySWQSEgoEY29kZRgDIAEoDVIEY29kZRImCg5kZWJpdHNSZXNlcnZlZBgEIAEoBFIOZGViaXRzUmVzZXJ2ZWQSJgoOZGViaXRzQWNjZXB0ZWQYBSABKARSDmRlYml0c0FjY2VwdGVkEigKD2NyZWRpdHNSZXNlcnZlZBgGIAEoBFIPY3JlZGl0c1Jlc2VydmVkEigKD2NyZWRpdHNBY2NlcHRlZBgHIAEoBFIPY3JlZGl0c0FjY2VwdGVkEi4KBWZsYWdzGAggASgLMhgucGFjaW9saS52MS5BY2NvdW50RmxhZ3NSBWZsYWdz');
@$core.Deprecated('Use accountFlagsDescriptor instead')
const AccountFlags$json = const {
  '1': 'AccountFlags',
  '2': const [
    const {'1': 'linked', '3': 1, '4': 1, '5': 8, '10': 'linked'},
    const {'1': 'debitsMustNotExceedCredits', '3': 2, '4': 1, '5': 8, '10': 'debitsMustNotExceedCredits'},
    const {'1': 'creditsMustNotExceedDebits', '3': 3, '4': 1, '5': 8, '10': 'creditsMustNotExceedDebits'},
  ],
};

/// Descriptor for `AccountFlags`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accountFlagsDescriptor = $convert.base64Decode('CgxBY2NvdW50RmxhZ3MSFgoGbGlua2VkGAEgASgIUgZsaW5rZWQSPgoaZGViaXRzTXVzdE5vdEV4Y2VlZENyZWRpdHMYAiABKAhSGmRlYml0c011c3ROb3RFeGNlZWRDcmVkaXRzEj4KGmNyZWRpdHNNdXN0Tm90RXhjZWVkRGViaXRzGAMgASgIUhpjcmVkaXRzTXVzdE5vdEV4Y2VlZERlYml0cw==');
@$core.Deprecated('Use configureAccountsArgsDescriptor instead')
const ConfigureAccountsArgs$json = const {
  '1': 'ConfigureAccountsArgs',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'ledgerId', '3': 2, '4': 1, '5': 13, '10': 'ledgerId'},
    const {'1': 'code', '3': 3, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'flags', '3': 4, '4': 1, '5': 11, '6': '.pacioli.v1.AccountFlags', '10': 'flags'},
  ],
};

/// Descriptor for `ConfigureAccountsArgs`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureAccountsArgsDescriptor = $convert.base64Decode('ChVDb25maWd1cmVBY2NvdW50c0FyZ3MSDgoCaWQYASABKAlSAmlkEhoKCGxlZGdlcklkGAIgASgNUghsZWRnZXJJZBISCgRjb2RlGAMgASgNUgRjb2RlEi4KBWZsYWdzGAQgASgLMhgucGFjaW9saS52MS5BY2NvdW50RmxhZ3NSBWZsYWdz');
@$core.Deprecated('Use configureAccountsRequestDescriptor instead')
const ConfigureAccountsRequest$json = const {
  '1': 'ConfigureAccountsRequest',
  '2': const [
    const {'1': 'args', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.ConfigureAccountsArgs', '10': 'args'},
  ],
};

/// Descriptor for `ConfigureAccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureAccountsRequestDescriptor = $convert.base64Decode('ChhDb25maWd1cmVBY2NvdW50c1JlcXVlc3QSNQoEYXJncxgBIAMoCzIhLnBhY2lvbGkudjEuQ29uZmlndXJlQWNjb3VudHNBcmdzUgRhcmdz');
@$core.Deprecated('Use configureAccountsResponseDescriptor instead')
const ConfigureAccountsResponse$json = const {
  '1': 'ConfigureAccountsResponse',
  '2': const [
    const {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `ConfigureAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureAccountsResponseDescriptor = $convert.base64Decode('ChlDb25maWd1cmVBY2NvdW50c1Jlc3BvbnNlEi4KBmVycm9ycxgBIAMoCzIWLnBhY2lvbGkudjEuRXZlbnRFcnJvclIGZXJyb3Jz');
@$core.Deprecated('Use getAccountsRequestDescriptor instead')
const GetAccountsRequest$json = const {
  '1': 'GetAccountsRequest',
  '2': const [
    const {'1': 'ids', '3': 1, '4': 3, '5': 9, '10': 'ids'},
  ],
};

/// Descriptor for `GetAccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAccountsRequestDescriptor = $convert.base64Decode('ChJHZXRBY2NvdW50c1JlcXVlc3QSEAoDaWRzGAEgAygJUgNpZHM=');
@$core.Deprecated('Use getAccountsResponseDescriptor instead')
const GetAccountsResponse$json = const {
  '1': 'GetAccountsResponse',
  '2': const [
    const {'1': 'accounts', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Account', '10': 'accounts'},
  ],
};

/// Descriptor for `GetAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAccountsResponseDescriptor = $convert.base64Decode('ChNHZXRBY2NvdW50c1Jlc3BvbnNlEi8KCGFjY291bnRzGAEgAygLMhMucGFjaW9saS52MS5BY2NvdW50UghhY2NvdW50cw==');
@$core.Deprecated('Use transferDescriptor instead')
const Transfer$json = const {
  '1': 'Transfer',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'debitAccountId', '3': 2, '4': 1, '5': 9, '10': 'debitAccountId'},
    const {'1': 'creditAccountId', '3': 3, '4': 1, '5': 9, '10': 'creditAccountId'},
    const {'1': 'amount', '3': 4, '4': 1, '5': 4, '10': 'amount'},
    const {'1': 'code', '3': 5, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'flags', '3': 6, '4': 1, '5': 11, '6': '.pacioli.v1.TransferFlags', '10': 'flags'},
    const {'1': 'timeout', '3': 7, '4': 1, '5': 4, '10': 'timeout'},
    const {'1': 'ledger', '3': 8, '4': 1, '5': 13, '10': 'ledger'},
    const {'1': 'pendingId', '3': 9, '4': 1, '5': 9, '10': 'pendingId'},
  ],
};

/// Descriptor for `Transfer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transferDescriptor = $convert.base64Decode('CghUcmFuc2ZlchIOCgJpZBgBIAEoCVICaWQSJgoOZGViaXRBY2NvdW50SWQYAiABKAlSDmRlYml0QWNjb3VudElkEigKD2NyZWRpdEFjY291bnRJZBgDIAEoCVIPY3JlZGl0QWNjb3VudElkEhYKBmFtb3VudBgEIAEoBFIGYW1vdW50EhIKBGNvZGUYBSABKA1SBGNvZGUSLwoFZmxhZ3MYBiABKAsyGS5wYWNpb2xpLnYxLlRyYW5zZmVyRmxhZ3NSBWZsYWdzEhgKB3RpbWVvdXQYByABKARSB3RpbWVvdXQSFgoGbGVkZ2VyGAggASgNUgZsZWRnZXISHAoJcGVuZGluZ0lkGAkgASgJUglwZW5kaW5nSWQ=');
@$core.Deprecated('Use transferFlagsDescriptor instead')
const TransferFlags$json = const {
  '1': 'TransferFlags',
  '2': const [
    const {'1': 'linked', '3': 1, '4': 1, '5': 8, '10': 'linked'},
    const {'1': 'pending', '3': 2, '4': 1, '5': 8, '10': 'pending'},
    const {'1': 'postPending', '3': 3, '4': 1, '5': 8, '10': 'postPending'},
    const {'1': 'voidPending', '3': 4, '4': 1, '5': 8, '10': 'voidPending'},
  ],
};

/// Descriptor for `TransferFlags`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transferFlagsDescriptor = $convert.base64Decode('Cg1UcmFuc2ZlckZsYWdzEhYKBmxpbmtlZBgBIAEoCFIGbGlua2VkEhgKB3BlbmRpbmcYAiABKAhSB3BlbmRpbmcSIAoLcG9zdFBlbmRpbmcYAyABKAhSC3Bvc3RQZW5kaW5nEiAKC3ZvaWRQZW5kaW5nGAQgASgIUgt2b2lkUGVuZGluZw==');
@$core.Deprecated('Use createTransfersRequestDescriptor instead')
const CreateTransfersRequest$json = const {
  '1': 'CreateTransfersRequest',
  '2': const [
    const {'1': 'transfers', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Transfer', '10': 'transfers'},
  ],
};

/// Descriptor for `CreateTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTransfersRequestDescriptor = $convert.base64Decode('ChZDcmVhdGVUcmFuc2ZlcnNSZXF1ZXN0EjIKCXRyYW5zZmVycxgBIAMoCzIULnBhY2lvbGkudjEuVHJhbnNmZXJSCXRyYW5zZmVycw==');
@$core.Deprecated('Use eventErrorDescriptor instead')
const EventError$json = const {
  '1': 'EventError',
  '2': const [
    const {'1': 'index', '3': 1, '4': 1, '5': 13, '10': 'index'},
    const {'1': 'code', '3': 2, '4': 1, '5': 13, '10': 'code'},
  ],
};

/// Descriptor for `EventError`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List eventErrorDescriptor = $convert.base64Decode('CgpFdmVudEVycm9yEhQKBWluZGV4GAEgASgNUgVpbmRleBISCgRjb2RlGAIgASgNUgRjb2Rl');
@$core.Deprecated('Use createTransfersResponseDescriptor instead')
const CreateTransfersResponse$json = const {
  '1': 'CreateTransfersResponse',
  '2': const [
    const {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `CreateTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTransfersResponseDescriptor = $convert.base64Decode('ChdDcmVhdGVUcmFuc2ZlcnNSZXNwb25zZRIuCgZlcnJvcnMYASADKAsyFi5wYWNpb2xpLnYxLkV2ZW50RXJyb3JSBmVycm9ycw==');
@$core.Deprecated('Use getTransfersRequestDescriptor instead')
const GetTransfersRequest$json = const {
  '1': 'GetTransfersRequest',
  '2': const [
    const {'1': 'ids', '3': 1, '4': 3, '5': 9, '10': 'ids'},
  ],
};

/// Descriptor for `GetTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTransfersRequestDescriptor = $convert.base64Decode('ChNHZXRUcmFuc2ZlcnNSZXF1ZXN0EhAKA2lkcxgBIAMoCVIDaWRz');
@$core.Deprecated('Use getTransfersResponseDescriptor instead')
const GetTransfersResponse$json = const {
  '1': 'GetTransfersResponse',
  '2': const [
    const {'1': 'transfers', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.Transfer', '10': 'transfers'},
  ],
};

/// Descriptor for `GetTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTransfersResponseDescriptor = $convert.base64Decode('ChRHZXRUcmFuc2ZlcnNSZXNwb25zZRIyCgl0cmFuc2ZlcnMYASADKAsyFC5wYWNpb2xpLnYxLlRyYW5zZmVyUgl0cmFuc2ZlcnM=');
@$core.Deprecated('Use postTransfersRequestDescriptor instead')
const PostTransfersRequest$json = const {
  '1': 'PostTransfersRequest',
  '2': const [
    const {'1': 'transferIds', '3': 1, '4': 3, '5': 9, '10': 'transferIds'},
  ],
};

/// Descriptor for `PostTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List postTransfersRequestDescriptor = $convert.base64Decode('ChRQb3N0VHJhbnNmZXJzUmVxdWVzdBIgCgt0cmFuc2ZlcklkcxgBIAMoCVILdHJhbnNmZXJJZHM=');
@$core.Deprecated('Use postTransfersResponseDescriptor instead')
const PostTransfersResponse$json = const {
  '1': 'PostTransfersResponse',
  '2': const [
    const {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `PostTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List postTransfersResponseDescriptor = $convert.base64Decode('ChVQb3N0VHJhbnNmZXJzUmVzcG9uc2USLgoGZXJyb3JzGAEgAygLMhYucGFjaW9saS52MS5FdmVudEVycm9yUgZlcnJvcnM=');
@$core.Deprecated('Use voidTransfersRequestDescriptor instead')
const VoidTransfersRequest$json = const {
  '1': 'VoidTransfersRequest',
  '2': const [
    const {'1': 'transferIds', '3': 1, '4': 3, '5': 9, '10': 'transferIds'},
  ],
};

/// Descriptor for `VoidTransfersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidTransfersRequestDescriptor = $convert.base64Decode('ChRWb2lkVHJhbnNmZXJzUmVxdWVzdBIgCgt0cmFuc2ZlcklkcxgBIAMoCVILdHJhbnNmZXJJZHM=');
@$core.Deprecated('Use voidTransfersResponseDescriptor instead')
const VoidTransfersResponse$json = const {
  '1': 'VoidTransfersResponse',
  '2': const [
    const {'1': 'errors', '3': 1, '4': 3, '5': 11, '6': '.pacioli.v1.EventError', '10': 'errors'},
  ],
};

/// Descriptor for `VoidTransfersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List voidTransfersResponseDescriptor = $convert.base64Decode('ChVWb2lkVHJhbnNmZXJzUmVzcG9uc2USLgoGZXJyb3JzGAEgAygLMhYucGFjaW9saS52MS5FdmVudEVycm9yUgZlcnJvcnM=');
