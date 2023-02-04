//
//  Generated code. Do not modify.
//  source: backend/v1/backend.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use paginationRequestDescriptor instead')
const PaginationRequest$json = {
  '1': 'PaginationRequest',
  '2': [
    {'1': 'pageSize', '3': 1, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'pageToken', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'pageToken', '17': true},
  ],
  '8': [
    {'1': '_pageToken'},
  ],
};

/// Descriptor for `PaginationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paginationRequestDescriptor = $convert.base64Decode(
    'ChFQYWdpbmF0aW9uUmVxdWVzdBIaCghwYWdlU2l6ZRgBIAEoBVIIcGFnZVNpemUSIQoJcGFnZV'
    'Rva2VuGAIgASgJSABSCXBhZ2VUb2tlbogBAUIMCgpfcGFnZVRva2Vu');

@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor = $convert.base64Decode(
    'CgVFbXB0eQ==');

@$core.Deprecated('Use getLinkedAccountsForPaymentRequestDescriptor instead')
const GetLinkedAccountsForPaymentRequest$json = {
  '1': 'GetLinkedAccountsForPaymentRequest',
  '2': [
    {'1': 'paymentId', '3': 1, '4': 1, '5': 9, '10': 'paymentId'},
  ],
};

/// Descriptor for `GetLinkedAccountsForPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinkedAccountsForPaymentRequestDescriptor = $convert.base64Decode(
    'CiJHZXRMaW5rZWRBY2NvdW50c0ZvclBheW1lbnRSZXF1ZXN0EhwKCXBheW1lbnRJZBgBIAEoCV'
    'IJcGF5bWVudElk');

@$core.Deprecated('Use getLinkedAccountsForPaymentResponseDescriptor instead')
const GetLinkedAccountsForPaymentResponse$json = {
  '1': 'GetLinkedAccountsForPaymentResponse',
  '2': [
    {'1': 'linkedAccounts', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.LinkedAccountForPayment', '10': 'linkedAccounts'},
  ],
};

/// Descriptor for `GetLinkedAccountsForPaymentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinkedAccountsForPaymentResponseDescriptor = $convert.base64Decode(
    'CiNHZXRMaW5rZWRBY2NvdW50c0ZvclBheW1lbnRSZXNwb25zZRJLCg5saW5rZWRBY2NvdW50cx'
    'gBIAMoCzIjLmJhY2tlbmQudjEuTGlua2VkQWNjb3VudEZvclBheW1lbnRSDmxpbmtlZEFjY291'
    'bnRz');

@$core.Deprecated('Use linkedAccountForPaymentDescriptor instead')
const LinkedAccountForPayment$json = {
  '1': 'LinkedAccountForPayment',
  '2': [
    {'1': 'details', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.LinkedAccount', '10': 'details'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
  ],
};

/// Descriptor for `LinkedAccountForPayment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkedAccountForPaymentDescriptor = $convert.base64Decode(
    'ChdMaW5rZWRBY2NvdW50Rm9yUGF5bWVudBIzCgdkZXRhaWxzGAEgASgLMhkuYmFja2VuZC52MS'
    '5MaW5rZWRBY2NvdW50UgdkZXRhaWxzEhgKB2VuYWJsZWQYAiABKAhSB2VuYWJsZWQ=');

@$core.Deprecated('Use getXagoDepositDetailsRequestDescriptor instead')
const GetXagoDepositDetailsRequest$json = {
  '1': 'GetXagoDepositDetailsRequest',
  '2': [
    {'1': 'linkedAccount', '3': 1, '4': 1, '5': 9, '10': 'linkedAccount'},
  ],
};

/// Descriptor for `GetXagoDepositDetailsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getXagoDepositDetailsRequestDescriptor = $convert.base64Decode(
    'ChxHZXRYYWdvRGVwb3NpdERldGFpbHNSZXF1ZXN0EiQKDWxpbmtlZEFjY291bnQYASABKAlSDW'
    'xpbmtlZEFjY291bnQ=');

@$core.Deprecated('Use getXagoDepositDetailsResponseDescriptor instead')
const GetXagoDepositDetailsResponse$json = {
  '1': 'GetXagoDepositDetailsResponse',
  '2': [
    {'1': 'details', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.XagoDepositDetails', '10': 'details'},
  ],
};

/// Descriptor for `GetXagoDepositDetailsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getXagoDepositDetailsResponseDescriptor = $convert.base64Decode(
    'Ch1HZXRYYWdvRGVwb3NpdERldGFpbHNSZXNwb25zZRI4CgdkZXRhaWxzGAEgAygLMh4uYmFja2'
    'VuZC52MS5YYWdvRGVwb3NpdERldGFpbHNSB2RldGFpbHM=');

@$core.Deprecated('Use xagoDepositDetailsDescriptor instead')
const XagoDepositDetails$json = {
  '1': 'XagoDepositDetails',
  '2': [
    {'1': 'currency', '3': 1, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'accountNumber', '3': 2, '4': 1, '5': 9, '10': 'accountNumber'},
    {'1': 'branchCode', '3': 3, '4': 1, '5': 9, '10': 'branchCode'},
    {'1': 'bankName', '3': 4, '4': 1, '5': 9, '10': 'bankName'},
    {'1': 'depositReference', '3': 5, '4': 1, '5': 9, '10': 'depositReference'},
  ],
};

/// Descriptor for `XagoDepositDetails`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List xagoDepositDetailsDescriptor = $convert.base64Decode(
    'ChJYYWdvRGVwb3NpdERldGFpbHMSGgoIY3VycmVuY3kYASABKAlSCGN1cnJlbmN5EiQKDWFjY2'
    '91bnROdW1iZXIYAiABKAlSDWFjY291bnROdW1iZXISHgoKYnJhbmNoQ29kZRgDIAEoCVIKYnJh'
    'bmNoQ29kZRIaCghiYW5rTmFtZRgEIAEoCVIIYmFua05hbWUSKgoQZGVwb3NpdFJlZmVyZW5jZR'
    'gFIAEoCVIQZGVwb3NpdFJlZmVyZW5jZQ==');

@$core.Deprecated('Use getXagoBalanceRequestDescriptor instead')
const GetXagoBalanceRequest$json = {
  '1': 'GetXagoBalanceRequest',
  '2': [
    {'1': 'linkedAccount', '3': 1, '4': 1, '5': 9, '10': 'linkedAccount'},
  ],
};

/// Descriptor for `GetXagoBalanceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getXagoBalanceRequestDescriptor = $convert.base64Decode(
    'ChVHZXRYYWdvQmFsYW5jZVJlcXVlc3QSJAoNbGlua2VkQWNjb3VudBgBIAEoCVINbGlua2VkQW'
    'Njb3VudA==');

@$core.Deprecated('Use getXagoBalanceResponseDescriptor instead')
const GetXagoBalanceResponse$json = {
  '1': 'GetXagoBalanceResponse',
  '2': [
    {'1': 'balances', '3': 2, '4': 3, '5': 11, '6': '.backend.v1.XagoBalance', '10': 'balances'},
  ],
};

/// Descriptor for `GetXagoBalanceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getXagoBalanceResponseDescriptor = $convert.base64Decode(
    'ChZHZXRYYWdvQmFsYW5jZVJlc3BvbnNlEjMKCGJhbGFuY2VzGAIgAygLMhcuYmFja2VuZC52MS'
    '5YYWdvQmFsYW5jZVIIYmFsYW5jZXM=');

@$core.Deprecated('Use xagoBalanceDescriptor instead')
const XagoBalance$json = {
  '1': 'XagoBalance',
  '2': [
    {'1': 'balance', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'balance'},
    {'1': 'available', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'available'},
    {'1': 'currency', '3': 3, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'linkedAccount', '3': 4, '4': 1, '5': 9, '10': 'linkedAccount'},
    {'1': 'formattedBalance', '3': 5, '4': 1, '5': 9, '10': 'formattedBalance'},
    {'1': 'formattedAvailableBalance', '3': 6, '4': 1, '5': 9, '10': 'formattedAvailableBalance'},
  ],
};

/// Descriptor for `XagoBalance`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List xagoBalanceDescriptor = $convert.base64Decode(
    'CgtYYWdvQmFsYW5jZRIsCgdiYWxhbmNlGAEgASgLMhIuYmFja2VuZC52MS5BbW91bnRSB2JhbG'
    'FuY2USMAoJYXZhaWxhYmxlGAIgASgLMhIuYmFja2VuZC52MS5BbW91bnRSCWF2YWlsYWJsZRIa'
    'CghjdXJyZW5jeRgDIAEoCVIIY3VycmVuY3kSJAoNbGlua2VkQWNjb3VudBgEIAEoCVINbGlua2'
    'VkQWNjb3VudBIqChBmb3JtYXR0ZWRCYWxhbmNlGAUgASgJUhBmb3JtYXR0ZWRCYWxhbmNlEjwK'
    'GWZvcm1hdHRlZEF2YWlsYWJsZUJhbGFuY2UYBiABKAlSGWZvcm1hdHRlZEF2YWlsYWJsZUJhbG'
    'FuY2U=');

@$core.Deprecated('Use withdrawXagoBalanceRequestDescriptor instead')
const WithdrawXagoBalanceRequest$json = {
  '1': 'WithdrawXagoBalanceRequest',
  '2': [
    {'1': 'fromLinkedAccount', '3': 1, '4': 1, '5': 9, '10': 'fromLinkedAccount'},
    {'1': 'toLinkedAccount', '3': 2, '4': 1, '5': 9, '10': 'toLinkedAccount'},
    {'1': 'amount', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'amount'},
    {'1': 'note', '3': 4, '4': 1, '5': 9, '10': 'note'},
  ],
};

/// Descriptor for `WithdrawXagoBalanceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List withdrawXagoBalanceRequestDescriptor = $convert.base64Decode(
    'ChpXaXRoZHJhd1hhZ29CYWxhbmNlUmVxdWVzdBIsChFmcm9tTGlua2VkQWNjb3VudBgBIAEoCV'
    'IRZnJvbUxpbmtlZEFjY291bnQSKAoPdG9MaW5rZWRBY2NvdW50GAIgASgJUg90b0xpbmtlZEFj'
    'Y291bnQSKgoGYW1vdW50GAMgASgLMhIuYmFja2VuZC52MS5BbW91bnRSBmFtb3VudBISCgRub3'
    'RlGAQgASgJUgRub3Rl');

@$core.Deprecated('Use addXagoBalanceAccountRequestDescriptor instead')
const AddXagoBalanceAccountRequest$json = {
  '1': 'AddXagoBalanceAccountRequest',
  '2': [
    {'1': 'currencyCode', '3': 1, '4': 1, '5': 9, '10': 'currencyCode'},
    {'1': 'nickname', '3': 2, '4': 1, '5': 9, '10': 'nickname'},
    {'1': 'title', '3': 3, '4': 1, '5': 9, '10': 'title'},
  ],
};

/// Descriptor for `AddXagoBalanceAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addXagoBalanceAccountRequestDescriptor = $convert.base64Decode(
    'ChxBZGRYYWdvQmFsYW5jZUFjY291bnRSZXF1ZXN0EiIKDGN1cnJlbmN5Q29kZRgBIAEoCVIMY3'
    'VycmVuY3lDb2RlEhoKCG5pY2tuYW1lGAIgASgJUghuaWNrbmFtZRIUCgV0aXRsZRgDIAEoCVIF'
    'dGl0bGU=');

@$core.Deprecated('Use addXagoBankAccountRequestDescriptor instead')
const AddXagoBankAccountRequest$json = {
  '1': 'AddXagoBankAccountRequest',
  '2': [
    {'1': 'accountNumber', '3': 1, '4': 1, '5': 9, '10': 'accountNumber'},
    {'1': 'branchCode', '3': 2, '4': 1, '5': 9, '10': 'branchCode'},
    {'1': 'bankName', '3': 3, '4': 1, '5': 9, '10': 'bankName'},
    {'1': 'iban', '3': 4, '4': 1, '5': 9, '10': 'iban'},
    {'1': 'bic', '3': 5, '4': 1, '5': 9, '10': 'bic'},
  ],
};

/// Descriptor for `AddXagoBankAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addXagoBankAccountRequestDescriptor = $convert.base64Decode(
    'ChlBZGRYYWdvQmFua0FjY291bnRSZXF1ZXN0EiQKDWFjY291bnROdW1iZXIYASABKAlSDWFjY2'
    '91bnROdW1iZXISHgoKYnJhbmNoQ29kZRgCIAEoCVIKYnJhbmNoQ29kZRIaCghiYW5rTmFtZRgD'
    'IAEoCVIIYmFua05hbWUSEgoEaWJhbhgEIAEoCVIEaWJhbhIQCgNiaWMYBSABKAlSA2JpYw==');

@$core.Deprecated('Use setDefaultSendLinkedAccountRequestDescriptor instead')
const SetDefaultSendLinkedAccountRequest$json = {
  '1': 'SetDefaultSendLinkedAccountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `SetDefaultSendLinkedAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setDefaultSendLinkedAccountRequestDescriptor = $convert.base64Decode(
    'CiJTZXREZWZhdWx0U2VuZExpbmtlZEFjY291bnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use setDefaultReceiveLinkedAccountRequestDescriptor instead')
const SetDefaultReceiveLinkedAccountRequest$json = {
  '1': 'SetDefaultReceiveLinkedAccountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `SetDefaultReceiveLinkedAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setDefaultReceiveLinkedAccountRequestDescriptor = $convert.base64Decode(
    'CiVTZXREZWZhdWx0UmVjZWl2ZUxpbmtlZEFjY291bnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA'
    '==');

@$core.Deprecated('Use slackCallbackRequestDescriptor instead')
const SlackCallbackRequest$json = {
  '1': 'SlackCallbackRequest',
  '2': [
    {'1': 'state', '3': 1, '4': 1, '5': 9, '10': 'state'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
  ],
};

/// Descriptor for `SlackCallbackRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List slackCallbackRequestDescriptor = $convert.base64Decode(
    'ChRTbGFja0NhbGxiYWNrUmVxdWVzdBIUCgVzdGF0ZRgBIAEoCVIFc3RhdGUSEgoEY29kZRgCIA'
    'EoCVIEY29kZQ==');

@$core.Deprecated('Use slackCallbackResponseDescriptor instead')
const SlackCallbackResponse$json = {
  '1': 'SlackCallbackResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `SlackCallbackResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List slackCallbackResponseDescriptor = $convert.base64Decode(
    'ChVTbGFja0NhbGxiYWNrUmVzcG9uc2USDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use createSlackAuthURLResponseDescriptor instead')
const CreateSlackAuthURLResponse$json = {
  '1': 'CreateSlackAuthURLResponse',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `CreateSlackAuthURLResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSlackAuthURLResponseDescriptor = $convert.base64Decode(
    'ChpDcmVhdGVTbGFja0F1dGhVUkxSZXNwb25zZRIQCgN1cmwYASABKAlSA3VybA==');

@$core.Deprecated('Use amountDescriptor instead')
const Amount$json = {
  '1': 'Amount',
  '2': [
    {'1': 'amount', '3': 1, '4': 1, '5': 4, '10': 'amount'},
    {'1': 'asset', '3': 2, '4': 1, '5': 9, '10': 'asset'},
    {'1': 'assetScale', '3': 3, '4': 1, '5': 5, '10': 'assetScale'},
    {'1': 'country', '3': 4, '4': 1, '5': 9, '10': 'country'},
  ],
};

/// Descriptor for `Amount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List amountDescriptor = $convert.base64Decode(
    'CgZBbW91bnQSFgoGYW1vdW50GAEgASgEUgZhbW91bnQSFAoFYXNzZXQYAiABKAlSBWFzc2V0Eh'
    '4KCmFzc2V0U2NhbGUYAyABKAVSCmFzc2V0U2NhbGUSGAoHY291bnRyeRgEIAEoCVIHY291bnRy'
    'eQ==');

@$core.Deprecated('Use transactionDescriptor instead')
const Transaction$json = {
  '1': 'Transaction',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'amount', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'amount'},
    {'1': 'source', '3': 4, '4': 1, '5': 9, '10': 'source'},
    {'1': 'destination', '3': 5, '4': 1, '5': 9, '10': 'destination'},
    {'1': 'timestamp', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
    {'1': 'state', '3': 7, '4': 1, '5': 9, '10': 'state'},
    {'1': 'foreignId', '3': 9, '4': 1, '5': 9, '10': 'foreignId'},
    {'1': 'receiverAccountId', '3': 10, '4': 1, '5': 9, '10': 'receiverAccountId'},
    {'1': 'senderAccountId', '3': 11, '4': 1, '5': 9, '10': 'senderAccountId'},
    {'1': 'title', '3': 12, '4': 1, '5': 9, '10': 'title'},
    {'1': 'formattedAmount', '3': 13, '4': 1, '5': 9, '10': 'formattedAmount'},
    {'1': 'formattedTime', '3': 14, '4': 1, '5': 9, '10': 'formattedTime'},
    {'1': 'formattedDate', '3': 15, '4': 1, '5': 9, '10': 'formattedDate'},
    {'1': 'subtotal', '3': 16, '4': 1, '5': 9, '10': 'subtotal'},
    {'1': 'fees', '3': 17, '4': 1, '5': 9, '10': 'fees'},
    {'1': 'accountTitle', '3': 18, '4': 1, '5': 9, '10': 'accountTitle'},
    {'1': 'reference', '3': 19, '4': 1, '5': 9, '10': 'reference'},
    {'1': 'destinationIdentity', '3': 20, '4': 1, '5': 9, '10': 'destinationIdentity'},
    {'1': 'destinationIdentityType', '3': 21, '4': 1, '5': 9, '10': 'destinationIdentityType'},
    {'1': 'refundState', '3': 22, '4': 1, '5': 5, '10': 'refundState'},
    {'1': 'paymentProtectionAmount', '3': 23, '4': 1, '5': 9, '10': 'paymentProtectionAmount'},
    {'1': 'hasPaymentProtection', '3': 24, '4': 1, '5': 8, '10': 'hasPaymentProtection'},
  ],
  '9': [
    {'1': 8, '2': 9},
  ],
};

/// Descriptor for `Transaction`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transactionDescriptor = $convert.base64Decode(
    'CgtUcmFuc2FjdGlvbhIOCgJpZBgBIAEoCVICaWQSEgoEdHlwZRgCIAEoCVIEdHlwZRIqCgZhbW'
    '91bnQYAyABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIGYW1vdW50EhYKBnNvdXJjZRgEIAEoCVIG'
    'c291cmNlEiAKC2Rlc3RpbmF0aW9uGAUgASgJUgtkZXN0aW5hdGlvbhI4Cgl0aW1lc3RhbXAYBi'
    'ABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl0aW1lc3RhbXASFAoFc3RhdGUYByAB'
    'KAlSBXN0YXRlEhwKCWZvcmVpZ25JZBgJIAEoCVIJZm9yZWlnbklkEiwKEXJlY2VpdmVyQWNjb3'
    'VudElkGAogASgJUhFyZWNlaXZlckFjY291bnRJZBIoCg9zZW5kZXJBY2NvdW50SWQYCyABKAlS'
    'D3NlbmRlckFjY291bnRJZBIUCgV0aXRsZRgMIAEoCVIFdGl0bGUSKAoPZm9ybWF0dGVkQW1vdW'
    '50GA0gASgJUg9mb3JtYXR0ZWRBbW91bnQSJAoNZm9ybWF0dGVkVGltZRgOIAEoCVINZm9ybWF0'
    'dGVkVGltZRIkCg1mb3JtYXR0ZWREYXRlGA8gASgJUg1mb3JtYXR0ZWREYXRlEhoKCHN1YnRvdG'
    'FsGBAgASgJUghzdWJ0b3RhbBISCgRmZWVzGBEgASgJUgRmZWVzEiIKDGFjY291bnRUaXRsZRgS'
    'IAEoCVIMYWNjb3VudFRpdGxlEhwKCXJlZmVyZW5jZRgTIAEoCVIJcmVmZXJlbmNlEjAKE2Rlc3'
    'RpbmF0aW9uSWRlbnRpdHkYFCABKAlSE2Rlc3RpbmF0aW9uSWRlbnRpdHkSOAoXZGVzdGluYXRp'
    'b25JZGVudGl0eVR5cGUYFSABKAlSF2Rlc3RpbmF0aW9uSWRlbnRpdHlUeXBlEiAKC3JlZnVuZF'
    'N0YXRlGBYgASgFUgtyZWZ1bmRTdGF0ZRI4ChdwYXltZW50UHJvdGVjdGlvbkFtb3VudBgXIAEo'
    'CVIXcGF5bWVudFByb3RlY3Rpb25BbW91bnQSMgoUaGFzUGF5bWVudFByb3RlY3Rpb24YGCABKA'
    'hSFGhhc1BheW1lbnRQcm90ZWN0aW9uSgQICBAJ');

@$core.Deprecated('Use listTransactionsResponseDescriptor instead')
const ListTransactionsResponse$json = {
  '1': 'ListTransactionsResponse',
  '2': [
    {'1': 'transactions', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Transaction', '10': 'transactions'},
    {'1': 'nextPageToken', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListTransactionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTransactionsResponseDescriptor = $convert.base64Decode(
    'ChhMaXN0VHJhbnNhY3Rpb25zUmVzcG9uc2USOwoMdHJhbnNhY3Rpb25zGAEgAygLMhcuYmFja2'
    'VuZC52MS5UcmFuc2FjdGlvblIMdHJhbnNhY3Rpb25zEiQKDW5leHRQYWdlVG9rZW4YAiABKAlS'
    'DW5leHRQYWdlVG9rZW4=');

@$core.Deprecated('Use confirmPaymentRequestDescriptor instead')
const ConfirmPaymentRequest$json = {
  '1': 'ConfirmPaymentRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `ConfirmPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List confirmPaymentRequestDescriptor = $convert.base64Decode(
    'ChVDb25maXJtUGF5bWVudFJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getPaymentRequestDescriptor instead')
const GetPaymentRequest$json = {
  '1': 'GetPaymentRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPaymentRequestDescriptor = $convert.base64Decode(
    'ChFHZXRQYXltZW50UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use updatePaymentRequestDescriptor instead')
const UpdatePaymentRequest$json = {
  '1': 'UpdatePaymentRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'senderAmount', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Amount', '9': 0, '10': 'senderAmount', '17': true},
    {'1': 'receiverAmount', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '9': 1, '10': 'receiverAmount', '17': true},
    {'1': 'note', '3': 4, '4': 1, '5': 9, '9': 2, '10': 'note', '17': true},
    {'1': 'senderAccount', '3': 5, '4': 1, '5': 9, '9': 3, '10': 'senderAccount', '17': true},
    {'1': 'receiverAccount', '3': 6, '4': 1, '5': 9, '9': 4, '10': 'receiverAccount', '17': true},
    {'1': 'receiverIdentity', '3': 7, '4': 1, '5': 9, '9': 5, '10': 'receiverIdentity', '17': true},
    {'1': 'receiverIdentityType', '3': 8, '4': 1, '5': 5, '9': 6, '10': 'receiverIdentityType', '17': true},
    {'1': 'threeDSID', '3': 9, '4': 1, '5': 9, '9': 7, '10': 'threeDSID', '17': true},
    {'1': 'otp', '3': 10, '4': 1, '5': 9, '9': 8, '10': 'otp', '17': true},
    {'1': 'ipAddress', '3': 11, '4': 1, '5': 9, '9': 9, '10': 'ipAddress', '17': true},
    {'1': 'addPaymentProtection', '3': 12, '4': 1, '5': 8, '9': 10, '10': 'addPaymentProtection', '17': true},
  ],
  '8': [
    {'1': '_senderAmount'},
    {'1': '_receiverAmount'},
    {'1': '_note'},
    {'1': '_senderAccount'},
    {'1': '_receiverAccount'},
    {'1': '_receiverIdentity'},
    {'1': '_receiverIdentityType'},
    {'1': '_threeDSID'},
    {'1': '_otp'},
    {'1': '_ipAddress'},
    {'1': '_addPaymentProtection'},
  ],
};

/// Descriptor for `UpdatePaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updatePaymentRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVQYXltZW50UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSOwoMc2VuZGVyQW1vdW50GA'
    'IgASgLMhIuYmFja2VuZC52MS5BbW91bnRIAFIMc2VuZGVyQW1vdW50iAEBEj8KDnJlY2VpdmVy'
    'QW1vdW50GAMgASgLMhIuYmFja2VuZC52MS5BbW91bnRIAVIOcmVjZWl2ZXJBbW91bnSIAQESFw'
    'oEbm90ZRgEIAEoCUgCUgRub3RliAEBEikKDXNlbmRlckFjY291bnQYBSABKAlIA1INc2VuZGVy'
    'QWNjb3VudIgBARItCg9yZWNlaXZlckFjY291bnQYBiABKAlIBFIPcmVjZWl2ZXJBY2NvdW50iA'
    'EBEi8KEHJlY2VpdmVySWRlbnRpdHkYByABKAlIBVIQcmVjZWl2ZXJJZGVudGl0eYgBARI3ChRy'
    'ZWNlaXZlcklkZW50aXR5VHlwZRgIIAEoBUgGUhRyZWNlaXZlcklkZW50aXR5VHlwZYgBARIhCg'
    'l0aHJlZURTSUQYCSABKAlIB1IJdGhyZWVEU0lEiAEBEhUKA290cBgKIAEoCUgIUgNvdHCIAQES'
    'IQoJaXBBZGRyZXNzGAsgASgJSAlSCWlwQWRkcmVzc4gBARI3ChRhZGRQYXltZW50UHJvdGVjdG'
    'lvbhgMIAEoCEgKUhRhZGRQYXltZW50UHJvdGVjdGlvbogBAUIPCg1fc2VuZGVyQW1vdW50QhEK'
    'D19yZWNlaXZlckFtb3VudEIHCgVfbm90ZUIQCg5fc2VuZGVyQWNjb3VudEISChBfcmVjZWl2ZX'
    'JBY2NvdW50QhMKEV9yZWNlaXZlcklkZW50aXR5QhcKFV9yZWNlaXZlcklkZW50aXR5VHlwZUIM'
    'CgpfdGhyZWVEU0lEQgYKBF9vdHBCDAoKX2lwQWRkcmVzc0IXChVfYWRkUGF5bWVudFByb3RlY3'
    'Rpb24=');

@$core.Deprecated('Use paymentDescriptor instead')
const Payment$json = {
  '1': 'Payment',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'publicID', '3': 2, '4': 1, '5': 9, '10': 'publicID'},
    {'1': 'state', '3': 3, '4': 1, '5': 5, '10': 'state'},
    {'1': 'receiverWalletUrl', '3': 4, '4': 1, '5': 9, '10': 'receiverWalletUrl'},
    {'1': 'receiverIdentity', '3': 5, '4': 1, '5': 9, '10': 'receiverIdentity'},
    {'1': 'receiverIdentityType', '3': 6, '4': 1, '5': 5, '10': 'receiverIdentityType'},
    {'1': 'senderAmount', '3': 7, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'senderAmount'},
    {'1': 'senderAccount', '3': 8, '4': 1, '5': 9, '10': 'senderAccount'},
    {'1': 'note', '3': 9, '4': 1, '5': 9, '10': 'note'},
    {'1': 'requiredActions', '3': 10, '4': 3, '5': 5, '10': 'requiredActions'},
    {'1': 'hasPaymentProtection', '3': 11, '4': 1, '5': 8, '10': 'hasPaymentProtection'},
    {'1': 'paymentProtectionAmount', '3': 12, '4': 1, '5': 9, '10': 'paymentProtectionAmount'},
    {'1': 'fxRate', '3': 13, '4': 1, '5': 9, '10': 'fxRate'},
    {'1': 'receiverAmount', '3': 14, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'receiverAmount'},
    {'1': 'totalSendAmount', '3': 15, '4': 1, '5': 9, '10': 'totalSendAmount'},
    {'1': 'receiverLinkedAccountCountryCode', '3': 16, '4': 1, '5': 9, '10': 'receiverLinkedAccountCountryCode'},
    {'1': 'formattedFees', '3': 17, '4': 1, '5': 9, '10': 'formattedFees'},
    {'1': 'receiverAccount', '3': 18, '4': 1, '5': 9, '10': 'receiverAccount'},
  ],
};

/// Descriptor for `Payment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paymentDescriptor = $convert.base64Decode(
    'CgdQYXltZW50Eg4KAmlkGAEgASgJUgJpZBIaCghwdWJsaWNJRBgCIAEoCVIIcHVibGljSUQSFA'
    'oFc3RhdGUYAyABKAVSBXN0YXRlEiwKEXJlY2VpdmVyV2FsbGV0VXJsGAQgASgJUhFyZWNlaXZl'
    'cldhbGxldFVybBIqChByZWNlaXZlcklkZW50aXR5GAUgASgJUhByZWNlaXZlcklkZW50aXR5Ej'
    'IKFHJlY2VpdmVySWRlbnRpdHlUeXBlGAYgASgFUhRyZWNlaXZlcklkZW50aXR5VHlwZRI2Cgxz'
    'ZW5kZXJBbW91bnQYByABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIMc2VuZGVyQW1vdW50EiQKDX'
    'NlbmRlckFjY291bnQYCCABKAlSDXNlbmRlckFjY291bnQSEgoEbm90ZRgJIAEoCVIEbm90ZRIo'
    'Cg9yZXF1aXJlZEFjdGlvbnMYCiADKAVSD3JlcXVpcmVkQWN0aW9ucxIyChRoYXNQYXltZW50UH'
    'JvdGVjdGlvbhgLIAEoCFIUaGFzUGF5bWVudFByb3RlY3Rpb24SOAoXcGF5bWVudFByb3RlY3Rp'
    'b25BbW91bnQYDCABKAlSF3BheW1lbnRQcm90ZWN0aW9uQW1vdW50EhYKBmZ4UmF0ZRgNIAEoCV'
    'IGZnhSYXRlEjoKDnJlY2VpdmVyQW1vdW50GA4gASgLMhIuYmFja2VuZC52MS5BbW91bnRSDnJl'
    'Y2VpdmVyQW1vdW50EigKD3RvdGFsU2VuZEFtb3VudBgPIAEoCVIPdG90YWxTZW5kQW1vdW50Ek'
    'oKIHJlY2VpdmVyTGlua2VkQWNjb3VudENvdW50cnlDb2RlGBAgASgJUiByZWNlaXZlckxpbmtl'
    'ZEFjY291bnRDb3VudHJ5Q29kZRIkCg1mb3JtYXR0ZWRGZWVzGBEgASgJUg1mb3JtYXR0ZWRGZW'
    'VzEigKD3JlY2VpdmVyQWNjb3VudBgSIAEoCVIPcmVjZWl2ZXJBY2NvdW50');

@$core.Deprecated('Use createPaymentRequestDescriptor instead')
const CreatePaymentRequest$json = {
  '1': 'CreatePaymentRequest',
  '2': [
    {'1': 'senderAmount', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'senderAmount'},
    {'1': 'receiverAmount', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'receiverAmount'},
    {'1': 'receiverIdentity', '3': 3, '4': 1, '5': 9, '10': 'receiverIdentity'},
    {'1': 'receiverIdentityType', '3': 4, '4': 1, '5': 5, '10': 'receiverIdentityType'},
    {'1': 'senderAccount', '3': 5, '4': 1, '5': 9, '9': 0, '10': 'senderAccount', '17': true},
    {'1': 'receiverAccount', '3': 6, '4': 1, '5': 9, '9': 1, '10': 'receiverAccount', '17': true},
    {'1': 'note', '3': 7, '4': 1, '5': 9, '9': 2, '10': 'note', '17': true},
    {'1': 'ipAddress', '3': 8, '4': 1, '5': 9, '9': 3, '10': 'ipAddress', '17': true},
    {'1': 'addPaymentProtection', '3': 9, '4': 1, '5': 8, '9': 4, '10': 'addPaymentProtection', '17': true},
  ],
  '8': [
    {'1': '_senderAccount'},
    {'1': '_receiverAccount'},
    {'1': '_note'},
    {'1': '_ipAddress'},
    {'1': '_addPaymentProtection'},
  ],
};

/// Descriptor for `CreatePaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createPaymentRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVQYXltZW50UmVxdWVzdBI2CgxzZW5kZXJBbW91bnQYASABKAsyEi5iYWNrZW5kLn'
    'YxLkFtb3VudFIMc2VuZGVyQW1vdW50EjoKDnJlY2VpdmVyQW1vdW50GAIgASgLMhIuYmFja2Vu'
    'ZC52MS5BbW91bnRSDnJlY2VpdmVyQW1vdW50EioKEHJlY2VpdmVySWRlbnRpdHkYAyABKAlSEH'
    'JlY2VpdmVySWRlbnRpdHkSMgoUcmVjZWl2ZXJJZGVudGl0eVR5cGUYBCABKAVSFHJlY2VpdmVy'
    'SWRlbnRpdHlUeXBlEikKDXNlbmRlckFjY291bnQYBSABKAlIAFINc2VuZGVyQWNjb3VudIgBAR'
    'ItCg9yZWNlaXZlckFjY291bnQYBiABKAlIAVIPcmVjZWl2ZXJBY2NvdW50iAEBEhcKBG5vdGUY'
    'ByABKAlIAlIEbm90ZYgBARIhCglpcEFkZHJlc3MYCCABKAlIA1IJaXBBZGRyZXNziAEBEjcKFG'
    'FkZFBheW1lbnRQcm90ZWN0aW9uGAkgASgISARSFGFkZFBheW1lbnRQcm90ZWN0aW9uiAEBQhAK'
    'Dl9zZW5kZXJBY2NvdW50QhIKEF9yZWNlaXZlckFjY291bnRCBwoFX25vdGVCDAoKX2lwQWRkcm'
    'Vzc0IXChVfYWRkUGF5bWVudFByb3RlY3Rpb24=');

@$core.Deprecated('Use getCardDetailsRequestDescriptor instead')
const GetCardDetailsRequest$json = {
  '1': 'GetCardDetailsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetCardDetailsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCardDetailsRequestDescriptor = $convert.base64Decode(
    'ChVHZXRDYXJkRGV0YWlsc1JlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use cardDetailsDescriptor instead')
const CardDetails$json = {
  '1': 'CardDetails',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'network', '3': 2, '4': 1, '5': 9, '10': 'network'},
    {'1': 'bin', '3': 3, '4': 1, '5': 9, '10': 'bin'},
    {'1': 'last4', '3': 4, '4': 1, '5': 9, '10': 'last4'},
    {'1': 'type', '3': 5, '4': 1, '5': 9, '10': 'type'},
    {'1': 'expiration', '3': 6, '4': 1, '5': 9, '10': 'expiration'},
    {'1': 'nickname', '3': 7, '4': 1, '5': 9, '10': 'nickname'},
    {'1': 'state', '3': 8, '4': 1, '5': 9, '10': 'state'},
    {'1': 'canSend', '3': 9, '4': 1, '5': 8, '10': 'canSend'},
    {'1': 'canReceive', '3': 10, '4': 1, '5': 8, '10': 'canReceive'},
  ],
};

/// Descriptor for `CardDetails`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cardDetailsDescriptor = $convert.base64Decode(
    'CgtDYXJkRGV0YWlscxIOCgJpZBgBIAEoCVICaWQSGAoHbmV0d29yaxgCIAEoCVIHbmV0d29yax'
    'IQCgNiaW4YAyABKAlSA2JpbhIUCgVsYXN0NBgEIAEoCVIFbGFzdDQSEgoEdHlwZRgFIAEoCVIE'
    'dHlwZRIeCgpleHBpcmF0aW9uGAYgASgJUgpleHBpcmF0aW9uEhoKCG5pY2tuYW1lGAcgASgJUg'
    'huaWNrbmFtZRIUCgVzdGF0ZRgIIAEoCVIFc3RhdGUSGAoHY2FuU2VuZBgJIAEoCFIHY2FuU2Vu'
    'ZBIeCgpjYW5SZWNlaXZlGAogASgIUgpjYW5SZWNlaXZl');

@$core.Deprecated('Use searchWalletsRequestDescriptor instead')
const SearchWalletsRequest$json = {
  '1': 'SearchWalletsRequest',
  '2': [
    {'1': 'term', '3': 1, '4': 1, '5': 9, '10': 'term'},
  ],
};

/// Descriptor for `SearchWalletsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchWalletsRequestDescriptor = $convert.base64Decode(
    'ChRTZWFyY2hXYWxsZXRzUmVxdWVzdBISCgR0ZXJtGAEgASgJUgR0ZXJt');

@$core.Deprecated('Use searchWalletsResponseDescriptor instead')
const SearchWalletsResponse$json = {
  '1': 'SearchWalletsResponse',
  '2': [
    {'1': 'results', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.SearchResult', '10': 'results'},
  ],
};

/// Descriptor for `SearchWalletsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchWalletsResponseDescriptor = $convert.base64Decode(
    'ChVTZWFyY2hXYWxsZXRzUmVzcG9uc2USMgoHcmVzdWx0cxgBIAMoCzIYLmJhY2tlbmQudjEuU2'
    'VhcmNoUmVzdWx0UgdyZXN1bHRz');

@$core.Deprecated('Use searchResultDescriptor instead')
const SearchResult$json = {
  '1': 'SearchResult',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'walletUrl', '3': 5, '4': 1, '5': 9, '10': 'walletUrl'},
    {
      '1': 'canSend',
      '3': 4,
      '4': 1,
      '5': 8,
      '8': {'3': true},
      '10': 'canSend',
    },
    {'1': 'identifier', '3': 2, '4': 1, '5': 9, '10': 'identifier'},
    {'1': 'identifierType', '3': 3, '4': 1, '5': 9, '10': 'identifierType'},
    {'1': 'subResults', '3': 6, '4': 3, '5': 11, '6': '.backend.v1.SearchResult', '10': 'subResults'},
  ],
};

/// Descriptor for `SearchResult`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchResultDescriptor = $convert.base64Decode(
    'CgxTZWFyY2hSZXN1bHQSGgoId2FsbGV0SUQYASABKAlSCHdhbGxldElEEhwKCXdhbGxldFVybB'
    'gFIAEoCVIJd2FsbGV0VXJsEhwKB2NhblNlbmQYBCABKAhCAhgBUgdjYW5TZW5kEh4KCmlkZW50'
    'aWZpZXIYAiABKAlSCmlkZW50aWZpZXISJgoOaWRlbnRpZmllclR5cGUYAyABKAlSDmlkZW50aW'
    'ZpZXJUeXBlEjgKCnN1YlJlc3VsdHMYBiADKAsyGC5iYWNrZW5kLnYxLlNlYXJjaFJlc3VsdFIK'
    'c3ViUmVzdWx0cw==');

@$core.Deprecated('Use getPublicWalletInfoRequestDescriptor instead')
const GetPublicWalletInfoRequest$json = {
  '1': 'GetPublicWalletInfoRequest',
  '2': [
    {'1': 'walletAddress', '3': 1, '4': 1, '5': 9, '10': 'walletAddress'},
  ],
};

/// Descriptor for `GetPublicWalletInfoRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPublicWalletInfoRequestDescriptor = $convert.base64Decode(
    'ChpHZXRQdWJsaWNXYWxsZXRJbmZvUmVxdWVzdBIkCg13YWxsZXRBZGRyZXNzGAEgASgJUg13YW'
    'xsZXRBZGRyZXNz');

@$core.Deprecated('Use publicWalletInfoDescriptor instead')
const PublicWalletInfo$json = {
  '1': 'PublicWalletInfo',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'address', '3': 2, '4': 1, '5': 9, '10': 'address'},
    {'1': 'shortAddress', '3': 3, '4': 1, '5': 9, '10': 'shortAddress'},
    {'1': 'publicName', '3': 4, '4': 1, '5': 9, '10': 'publicName'},
    {'1': 'identities', '3': 5, '4': 3, '5': 11, '6': '.backend.v1.Identity', '10': 'identities'},
    {'1': 'canReceive', '3': 6, '4': 1, '5': 8, '10': 'canReceive'},
  ],
};

/// Descriptor for `PublicWalletInfo`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List publicWalletInfoDescriptor = $convert.base64Decode(
    'ChBQdWJsaWNXYWxsZXRJbmZvEhoKCHdhbGxldElEGAEgASgJUgh3YWxsZXRJRBIYCgdhZGRyZX'
    'NzGAIgASgJUgdhZGRyZXNzEiIKDHNob3J0QWRkcmVzcxgDIAEoCVIMc2hvcnRBZGRyZXNzEh4K'
    'CnB1YmxpY05hbWUYBCABKAlSCnB1YmxpY05hbWUSNAoKaWRlbnRpdGllcxgFIAMoCzIULmJhY2'
    'tlbmQudjEuSWRlbnRpdHlSCmlkZW50aXRpZXMSHgoKY2FuUmVjZWl2ZRgGIAEoCFIKY2FuUmVj'
    'ZWl2ZQ==');

@$core.Deprecated('Use walletInfoDescriptor instead')
const WalletInfo$json = {
  '1': 'WalletInfo',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'url', '3': 2, '4': 1, '5': 9, '10': 'url'},
    {'1': 'formattedURL', '3': 3, '4': 1, '5': 9, '10': 'formattedURL'},
    {'1': 'hasCard', '3': 4, '4': 1, '5': 8, '10': 'hasCard'},
    {'1': 'hasBank', '3': 5, '4': 1, '5': 8, '10': 'hasBank'},
    {'1': 'hasIdentities', '3': 6, '4': 1, '5': 8, '10': 'hasIdentities'},
    {'1': 'hasTransacted', '3': 7, '4': 1, '5': 8, '10': 'hasTransacted'},
    {'1': 'hasWalletAddress', '3': 8, '4': 1, '5': 8, '10': 'hasWalletAddress'},
    {'1': 'hasBalances', '3': 9, '4': 1, '5': 8, '10': 'hasBalances'},
    {'1': 'exceededLimits', '3': 10, '4': 1, '5': 8, '10': 'exceededLimits'},
  ],
};

/// Descriptor for `WalletInfo`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletInfoDescriptor = $convert.base64Decode(
    'CgpXYWxsZXRJbmZvEhoKCHdhbGxldElEGAEgASgJUgh3YWxsZXRJRBIQCgN1cmwYAiABKAlSA3'
    'VybBIiCgxmb3JtYXR0ZWRVUkwYAyABKAlSDGZvcm1hdHRlZFVSTBIYCgdoYXNDYXJkGAQgASgI'
    'UgdoYXNDYXJkEhgKB2hhc0JhbmsYBSABKAhSB2hhc0JhbmsSJAoNaGFzSWRlbnRpdGllcxgGIA'
    'EoCFINaGFzSWRlbnRpdGllcxIkCg1oYXNUcmFuc2FjdGVkGAcgASgIUg1oYXNUcmFuc2FjdGVk'
    'EioKEGhhc1dhbGxldEFkZHJlc3MYCCABKAhSEGhhc1dhbGxldEFkZHJlc3MSIAoLaGFzQmFsYW'
    '5jZXMYCSABKAhSC2hhc0JhbGFuY2VzEiYKDmV4Y2VlZGVkTGltaXRzGAogASgIUg5leGNlZWRl'
    'ZExpbWl0cw==');

@$core.Deprecated('Use featuresDescriptor instead')
const Features$json = {
  '1': 'Features',
  '2': [
    {'1': 'sendEnabled', '3': 1, '4': 1, '5': 8, '10': 'sendEnabled'},
    {'1': 'receiveEnabled', '3': 2, '4': 1, '5': 8, '10': 'receiveEnabled'},
    {'1': 'linkedAccountsEnabled', '3': 3, '4': 1, '5': 8, '10': 'linkedAccountsEnabled'},
    {'1': 'cardsEnabled', '3': 4, '4': 1, '5': 8, '10': 'cardsEnabled'},
    {'1': 'banksEnabled', '3': 5, '4': 1, '5': 8, '10': 'banksEnabled'},
    {'1': 'identitiesEnabled', '3': 6, '4': 1, '5': 8, '10': 'identitiesEnabled'},
    {'1': 'twitterEnabled', '3': 7, '4': 1, '5': 8, '10': 'twitterEnabled'},
    {'1': 'addCardsEnabled', '3': 8, '4': 1, '5': 8, '10': 'addCardsEnabled'},
  ],
};

/// Descriptor for `Features`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List featuresDescriptor = $convert.base64Decode(
    'CghGZWF0dXJlcxIgCgtzZW5kRW5hYmxlZBgBIAEoCFILc2VuZEVuYWJsZWQSJgoOcmVjZWl2ZU'
    'VuYWJsZWQYAiABKAhSDnJlY2VpdmVFbmFibGVkEjQKFWxpbmtlZEFjY291bnRzRW5hYmxlZBgD'
    'IAEoCFIVbGlua2VkQWNjb3VudHNFbmFibGVkEiIKDGNhcmRzRW5hYmxlZBgEIAEoCFIMY2FyZH'
    'NFbmFibGVkEiIKDGJhbmtzRW5hYmxlZBgFIAEoCFIMYmFua3NFbmFibGVkEiwKEWlkZW50aXRp'
    'ZXNFbmFibGVkGAYgASgIUhFpZGVudGl0aWVzRW5hYmxlZBImCg50d2l0dGVyRW5hYmxlZBgHIA'
    'EoCFIOdHdpdHRlckVuYWJsZWQSKAoPYWRkQ2FyZHNFbmFibGVkGAggASgIUg9hZGRDYXJkc0Vu'
    'YWJsZWQ=');

@$core.Deprecated('Use createCardRequestDescriptor instead')
const CreateCardRequest$json = {
  '1': 'CreateCardRequest',
  '2': [
    {'1': 'tokenID', '3': 1, '4': 1, '5': 9, '10': 'tokenID'},
  ],
};

/// Descriptor for `CreateCardRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createCardRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVDYXJkUmVxdWVzdBIYCgd0b2tlbklEGAEgASgJUgd0b2tlbklE');

@$core.Deprecated('Use initQuote3DSRequestDescriptor instead')
const InitQuote3DSRequest$json = {
  '1': 'InitQuote3DSRequest',
  '2': [
    {'1': 'quoteID', '3': 1, '4': 1, '5': 9, '10': 'quoteID'},
  ],
};

/// Descriptor for `InitQuote3DSRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List initQuote3DSRequestDescriptor = $convert.base64Decode(
    'ChNJbml0UXVvdGUzRFNSZXF1ZXN0EhgKB3F1b3RlSUQYASABKAlSB3F1b3RlSUQ=');

@$core.Deprecated('Use init3DSRequestDescriptor instead')
const Init3DSRequest$json = {
  '1': 'Init3DSRequest',
  '2': [
    {'1': 'paymentID', '3': 1, '4': 1, '5': 9, '10': 'paymentID'},
  ],
};

/// Descriptor for `Init3DSRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List init3DSRequestDescriptor = $convert.base64Decode(
    'Cg5Jbml0M0RTUmVxdWVzdBIcCglwYXltZW50SUQYASABKAlSCXBheW1lbnRJRA==');

@$core.Deprecated('Use init3DSResponseDescriptor instead')
const Init3DSResponse$json = {
  '1': 'Init3DSResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'jwt', '3': 2, '4': 1, '5': 9, '10': 'jwt'},
    {'1': 'deviceCollectionURL', '3': 3, '4': 1, '5': 9, '10': 'deviceCollectionURL'},
    {'1': 'songbirdURL', '3': 4, '4': 1, '5': 9, '10': 'songbirdURL'},
  ],
};

/// Descriptor for `Init3DSResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List init3DSResponseDescriptor = $convert.base64Decode(
    'Cg9Jbml0M0RTUmVzcG9uc2USDgoCaWQYASABKAlSAmlkEhAKA2p3dBgCIAEoCVIDand0EjAKE2'
    'RldmljZUNvbGxlY3Rpb25VUkwYAyABKAlSE2RldmljZUNvbGxlY3Rpb25VUkwSIAoLc29uZ2Jp'
    'cmRVUkwYBCABKAlSC3NvbmdiaXJkVVJM');

@$core.Deprecated('Use lookup3DSResponseDescriptor instead')
const Lookup3DSResponse$json = {
  '1': 'Lookup3DSResponse',
  '2': [
    {'1': 'processorTransactionID', '3': 1, '4': 1, '5': 9, '10': 'processorTransactionID'},
    {'1': 'challengeURL', '3': 2, '4': 1, '5': 9, '10': 'challengeURL'},
    {'1': 'payload', '3': 3, '4': 1, '5': 9, '10': 'payload'},
  ],
};

/// Descriptor for `Lookup3DSResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List lookup3DSResponseDescriptor = $convert.base64Decode(
    'ChFMb29rdXAzRFNSZXNwb25zZRI2ChZwcm9jZXNzb3JUcmFuc2FjdGlvbklEGAEgASgJUhZwcm'
    '9jZXNzb3JUcmFuc2FjdGlvbklEEiIKDGNoYWxsZW5nZVVSTBgCIAEoCVIMY2hhbGxlbmdlVVJM'
    'EhgKB3BheWxvYWQYAyABKAlSB3BheWxvYWQ=');

@$core.Deprecated('Use lookup3DSRequestDescriptor instead')
const Lookup3DSRequest$json = {
  '1': 'Lookup3DSRequest',
  '2': [
    {'1': 'threeDSID', '3': 1, '4': 1, '5': 9, '10': 'threeDSID'},
    {'1': 'javascriptEnabled', '3': 2, '4': 1, '5': 8, '10': 'javascriptEnabled'},
    {'1': 'userAgent', '3': 3, '4': 1, '5': 9, '10': 'userAgent'},
    {'1': 'header', '3': 4, '4': 1, '5': 9, '10': 'header'},
    {'1': 'javaEnabled', '3': 5, '4': 1, '5': 8, '10': 'javaEnabled'},
    {'1': 'language', '3': 6, '4': 1, '5': 9, '10': 'language'},
    {'1': 'colorDepth', '3': 7, '4': 1, '5': 9, '10': 'colorDepth'},
    {'1': 'screenHeight', '3': 8, '4': 1, '5': 9, '10': 'screenHeight'},
    {'1': 'screenWidth', '3': 9, '4': 1, '5': 9, '10': 'screenWidth'},
    {'1': 'timezone', '3': 10, '4': 1, '5': 9, '10': 'timezone'},
    {'1': 'ipAddress', '3': 12, '4': 1, '5': 9, '10': 'ipAddress'},
  ],
};

/// Descriptor for `Lookup3DSRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List lookup3DSRequestDescriptor = $convert.base64Decode(
    'ChBMb29rdXAzRFNSZXF1ZXN0EhwKCXRocmVlRFNJRBgBIAEoCVIJdGhyZWVEU0lEEiwKEWphdm'
    'FzY3JpcHRFbmFibGVkGAIgASgIUhFqYXZhc2NyaXB0RW5hYmxlZBIcCgl1c2VyQWdlbnQYAyAB'
    'KAlSCXVzZXJBZ2VudBIWCgZoZWFkZXIYBCABKAlSBmhlYWRlchIgCgtqYXZhRW5hYmxlZBgFIA'
    'EoCFILamF2YUVuYWJsZWQSGgoIbGFuZ3VhZ2UYBiABKAlSCGxhbmd1YWdlEh4KCmNvbG9yRGVw'
    'dGgYByABKAlSCmNvbG9yRGVwdGgSIgoMc2NyZWVuSGVpZ2h0GAggASgJUgxzY3JlZW5IZWlnaH'
    'QSIAoLc2NyZWVuV2lkdGgYCSABKAlSC3NjcmVlbldpZHRoEhoKCHRpbWV6b25lGAogASgJUgh0'
    'aW1lem9uZRIcCglpcEFkZHJlc3MYDCABKAlSCWlwQWRkcmVzcw==');

@$core.Deprecated('Use authenticate3DSRequestDescriptor instead')
const Authenticate3DSRequest$json = {
  '1': 'Authenticate3DSRequest',
  '2': [
    {'1': 'threeDSID', '3': 1, '4': 1, '5': 9, '10': 'threeDSID'},
    {'1': 'jwt', '3': 2, '4': 1, '5': 9, '10': 'jwt'},
  ],
};

/// Descriptor for `Authenticate3DSRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List authenticate3DSRequestDescriptor = $convert.base64Decode(
    'ChZBdXRoZW50aWNhdGUzRFNSZXF1ZXN0EhwKCXRocmVlRFNJRBgBIAEoCVIJdGhyZWVEU0lEEh'
    'AKA2p3dBgCIAEoCVIDand0');

@$core.Deprecated('Use authenticate3DSResponseDescriptor instead')
const Authenticate3DSResponse$json = {
  '1': 'Authenticate3DSResponse',
  '2': [
    {'1': 'status', '3': 1, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `Authenticate3DSResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List authenticate3DSResponseDescriptor = $convert.base64Decode(
    'ChdBdXRoZW50aWNhdGUzRFNSZXNwb25zZRIWCgZzdGF0dXMYASABKAlSBnN0YXR1cw==');

@$core.Deprecated('Use createMXBankAccountsRequestDescriptor instead')
const CreateMXBankAccountsRequest$json = {
  '1': 'CreateMXBankAccountsRequest',
  '2': [
    {'1': 'sessionGuid', '3': 1, '4': 1, '5': 9, '10': 'sessionGuid'},
    {'1': 'userGuid', '3': 2, '4': 1, '5': 9, '10': 'userGuid'},
    {'1': 'memberGuid', '3': 3, '4': 1, '5': 9, '10': 'memberGuid'},
  ],
};

/// Descriptor for `CreateMXBankAccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMXBankAccountsRequestDescriptor = $convert.base64Decode(
    'ChtDcmVhdGVNWEJhbmtBY2NvdW50c1JlcXVlc3QSIAoLc2Vzc2lvbkd1aWQYASABKAlSC3Nlc3'
    'Npb25HdWlkEhoKCHVzZXJHdWlkGAIgASgJUgh1c2VyR3VpZBIeCgptZW1iZXJHdWlkGAMgASgJ'
    'UgptZW1iZXJHdWlk');

@$core.Deprecated('Use createMXBankAccountsResponseDescriptor instead')
const CreateMXBankAccountsResponse$json = {
  '1': 'CreateMXBankAccountsResponse',
  '2': [
    {'1': 'linkedAccounts', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.LinkedAccount', '10': 'linkedAccounts'},
  ],
};

/// Descriptor for `CreateMXBankAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMXBankAccountsResponseDescriptor = $convert.base64Decode(
    'ChxDcmVhdGVNWEJhbmtBY2NvdW50c1Jlc3BvbnNlEkEKDmxpbmtlZEFjY291bnRzGAEgAygLMh'
    'kuYmFja2VuZC52MS5MaW5rZWRBY2NvdW50Ug5saW5rZWRBY2NvdW50cw==');

@$core.Deprecated('Use mXWidgetResponseDescriptor instead')
const MXWidgetResponse$json = {
  '1': 'MXWidgetResponse',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `MXWidgetResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mXWidgetResponseDescriptor = $convert.base64Decode(
    'ChBNWFdpZGdldFJlc3BvbnNlEhAKA3VybBgBIAEoCVIDdXJs');

@$core.Deprecated('Use connectionLimitsDescriptor instead')
const ConnectionLimits$json = {
  '1': 'ConnectionLimits',
  '2': [
    {'1': 'daily', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'daily'},
    {'1': 'monthly', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'monthly'},
    {'1': 'overall', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'overall'},
  ],
};

/// Descriptor for `ConnectionLimits`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List connectionLimitsDescriptor = $convert.base64Decode(
    'ChBDb25uZWN0aW9uTGltaXRzEigKBWRhaWx5GAEgASgLMhIuYmFja2VuZC52MS5BbW91bnRSBW'
    'RhaWx5EiwKB21vbnRobHkYAiABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIHbW9udGhseRIsCgdv'
    'dmVyYWxsGAMgASgLMhIuYmFja2VuZC52MS5BbW91bnRSB292ZXJhbGw=');

@$core.Deprecated('Use connectionDescriptor instead')
const Connection$json = {
  '1': 'Connection',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'applicationName', '3': 2, '4': 1, '5': 9, '10': 'applicationName'},
    {'1': 'publicKeyFingerprint', '3': 3, '4': 1, '5': 9, '10': 'publicKeyFingerprint'},
    {'1': 'createdAt', '3': 4, '4': 1, '5': 9, '10': 'createdAt'},
    {'1': 'lastUsedAt', '3': 5, '4': 1, '5': 9, '10': 'lastUsedAt'},
  ],
};

/// Descriptor for `Connection`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List connectionDescriptor = $convert.base64Decode(
    'CgpDb25uZWN0aW9uEg4KAmlkGAEgASgJUgJpZBIoCg9hcHBsaWNhdGlvbk5hbWUYAiABKAlSD2'
    'FwcGxpY2F0aW9uTmFtZRIyChRwdWJsaWNLZXlGaW5nZXJwcmludBgDIAEoCVIUcHVibGljS2V5'
    'RmluZ2VycHJpbnQSHAoJY3JlYXRlZEF0GAQgASgJUgljcmVhdGVkQXQSHgoKbGFzdFVzZWRBdB'
    'gFIAEoCVIKbGFzdFVzZWRBdA==');

@$core.Deprecated('Use createConnectionRequestDescriptor instead')
const CreateConnectionRequest$json = {
  '1': 'CreateConnectionRequest',
  '2': [
    {'1': 'applicationName', '3': 1, '4': 1, '5': 9, '10': 'applicationName'},
    {'1': 'publicKey', '3': 2, '4': 1, '5': 9, '10': 'publicKey'},
    {'1': 'dailyLimit', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'dailyLimit'},
    {'1': 'monthlyLimit', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'monthlyLimit'},
    {'1': 'overallLimit', '3': 5, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'overallLimit'},
  ],
};

/// Descriptor for `CreateConnectionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createConnectionRequestDescriptor = $convert.base64Decode(
    'ChdDcmVhdGVDb25uZWN0aW9uUmVxdWVzdBIoCg9hcHBsaWNhdGlvbk5hbWUYASABKAlSD2FwcG'
    'xpY2F0aW9uTmFtZRIcCglwdWJsaWNLZXkYAiABKAlSCXB1YmxpY0tleRIyCgpkYWlseUxpbWl0'
    'GAMgASgLMhIuYmFja2VuZC52MS5BbW91bnRSCmRhaWx5TGltaXQSNgoMbW9udGhseUxpbWl0GA'
    'QgASgLMhIuYmFja2VuZC52MS5BbW91bnRSDG1vbnRobHlMaW1pdBI2CgxvdmVyYWxsTGltaXQY'
    'BSABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIMb3ZlcmFsbExpbWl0');

@$core.Deprecated('Use getConnectionRequestDescriptor instead')
const GetConnectionRequest$json = {
  '1': 'GetConnectionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetConnectionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getConnectionRequestDescriptor = $convert.base64Decode(
    'ChRHZXRDb25uZWN0aW9uUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use getConnectionLimitsRequestDescriptor instead')
const GetConnectionLimitsRequest$json = {
  '1': 'GetConnectionLimitsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetConnectionLimitsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getConnectionLimitsRequestDescriptor = $convert.base64Decode(
    'ChpHZXRDb25uZWN0aW9uTGltaXRzUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use deleteConnectionRequestDescriptor instead')
const DeleteConnectionRequest$json = {
  '1': 'DeleteConnectionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteConnectionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteConnectionRequestDescriptor = $convert.base64Decode(
    'ChdEZWxldGVDb25uZWN0aW9uUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use listConnectionsResponseDescriptor instead')
const ListConnectionsResponse$json = {
  '1': 'ListConnectionsResponse',
  '2': [
    {'1': 'keys', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Connection', '10': 'keys'},
  ],
};

/// Descriptor for `ListConnectionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listConnectionsResponseDescriptor = $convert.base64Decode(
    'ChdMaXN0Q29ubmVjdGlvbnNSZXNwb25zZRIqCgRrZXlzGAEgAygLMhYuYmFja2VuZC52MS5Db2'
    '5uZWN0aW9uUgRrZXlz');

@$core.Deprecated('Use updateConnectionLimitsRequestDescriptor instead')
const UpdateConnectionLimitsRequest$json = {
  '1': 'UpdateConnectionLimitsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'daily', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'daily'},
    {'1': 'monthly', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'monthly'},
    {'1': 'overall', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'overall'},
  ],
};

/// Descriptor for `UpdateConnectionLimitsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateConnectionLimitsRequestDescriptor = $convert.base64Decode(
    'Ch1VcGRhdGVDb25uZWN0aW9uTGltaXRzUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSKAoFZGFpbH'
    'kYAiABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIFZGFpbHkSLAoHbW9udGhseRgDIAEoCzISLmJh'
    'Y2tlbmQudjEuQW1vdW50Ugdtb250aGx5EiwKB292ZXJhbGwYBCABKAsyEi5iYWNrZW5kLnYxLk'
    'Ftb3VudFIHb3ZlcmFsbA==');

@$core.Deprecated('Use transferDescriptor instead')
const Transfer$json = {
  '1': 'Transfer',
  '2': [
    {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    {'1': 'state', '3': 2, '4': 1, '5': 9, '10': 'state'},
    {'1': 'timestamp', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
    {'1': 'amount', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'amount'},
    {'1': 'foreignId', '3': 5, '4': 1, '5': 9, '10': 'foreignId'},
    {'1': 'linkedAccountId', '3': 6, '4': 1, '5': 9, '10': 'linkedAccountId'},
  ],
};

/// Descriptor for `Transfer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transferDescriptor = $convert.base64Decode(
    'CghUcmFuc2ZlchISCgR0eXBlGAEgASgJUgR0eXBlEhQKBXN0YXRlGAIgASgJUgVzdGF0ZRI4Cg'
    'l0aW1lc3RhbXAYAyABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl0aW1lc3RhbXAS'
    'KgoGYW1vdW50GAQgASgLMhIuYmFja2VuZC52MS5BbW91bnRSBmFtb3VudBIcCglmb3JlaWduSW'
    'QYBSABKAlSCWZvcmVpZ25JZBIoCg9saW5rZWRBY2NvdW50SWQYBiABKAlSD2xpbmtlZEFjY291'
    'bnRJZA==');

@$core.Deprecated('Use listStatementsResponseDescriptor instead')
const ListStatementsResponse$json = {
  '1': 'ListStatementsResponse',
  '2': [
    {'1': 'periods', '3': 1, '4': 3, '5': 9, '10': 'periods'},
    {'1': 'nextPageToken', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListStatementsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listStatementsResponseDescriptor = $convert.base64Decode(
    'ChZMaXN0U3RhdGVtZW50c1Jlc3BvbnNlEhgKB3BlcmlvZHMYASADKAlSB3BlcmlvZHMSJAoNbm'
    'V4dFBhZ2VUb2tlbhgCIAEoCVINbmV4dFBhZ2VUb2tlbg==');

@$core.Deprecated('Use createSupportTicketRequestDescriptor instead')
const CreateSupportTicketRequest$json = {
  '1': 'CreateSupportTicketRequest',
  '2': [
    {'1': 'description', '3': 1, '4': 1, '5': 9, '10': 'description'},
    {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
  ],
};

/// Descriptor for `CreateSupportTicketRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSupportTicketRequestDescriptor = $convert.base64Decode(
    'ChpDcmVhdGVTdXBwb3J0VGlja2V0UmVxdWVzdBIgCgtkZXNjcmlwdGlvbhgBIAEoCVILZGVzY3'
    'JpcHRpb24SHAoJZmlyc3ROYW1lGAIgASgJUglmaXJzdE5hbWUSGgoIbGFzdE5hbWUYAyABKAlS'
    'CGxhc3ROYW1lEhQKBWVtYWlsGAQgASgJUgVlbWFpbA==');

@$core.Deprecated('Use individualKYCResponseDescriptor instead')
const IndividualKYCResponse$json = {
  '1': 'IndividualKYCResponse',
  '2': [
    {'1': 'firstName', '3': 1, '4': 1, '5': 9, '10': 'firstName'},
    {'1': 'lastName', '3': 2, '4': 1, '5': 9, '10': 'lastName'},
    {'1': 'countryCode', '3': 3, '4': 1, '5': 9, '10': 'countryCode'},
    {'1': 'gender', '3': 4, '4': 1, '5': 5, '10': 'gender'},
    {'1': 'dateOfBirth', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'dateOfBirth'},
    {'1': 'address', '3': 6, '4': 1, '5': 11, '6': '.backend.v1.Address', '10': 'address'},
    {'1': 'placeOfBirth', '3': 7, '4': 1, '5': 9, '10': 'placeOfBirth'},
    {'1': 'nationality', '3': 8, '4': 1, '5': 9, '10': 'nationality'},
  ],
};

/// Descriptor for `IndividualKYCResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List individualKYCResponseDescriptor = $convert.base64Decode(
    'ChVJbmRpdmlkdWFsS1lDUmVzcG9uc2USHAoJZmlyc3ROYW1lGAEgASgJUglmaXJzdE5hbWUSGg'
    'oIbGFzdE5hbWUYAiABKAlSCGxhc3ROYW1lEiAKC2NvdW50cnlDb2RlGAMgASgJUgtjb3VudHJ5'
    'Q29kZRIWCgZnZW5kZXIYBCABKAVSBmdlbmRlchI8CgtkYXRlT2ZCaXJ0aBgFIAEoCzIaLmdvb2'
    'dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSC2RhdGVPZkJpcnRoEi0KB2FkZHJlc3MYBiABKAsyEy5i'
    'YWNrZW5kLnYxLkFkZHJlc3NSB2FkZHJlc3MSIgoMcGxhY2VPZkJpcnRoGAcgASgJUgxwbGFjZU'
    '9mQmlydGgSIAoLbmF0aW9uYWxpdHkYCCABKAlSC25hdGlvbmFsaXR5');

@$core.Deprecated('Use updateIndividualKYCRequestDescriptor instead')
const UpdateIndividualKYCRequest$json = {
  '1': 'UpdateIndividualKYCRequest',
  '2': [
    {'1': 'firstName', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'firstName', '17': true},
    {'1': 'lastName', '3': 2, '4': 1, '5': 9, '9': 1, '10': 'lastName', '17': true},
    {'1': 'countryCode', '3': 3, '4': 1, '5': 9, '9': 2, '10': 'countryCode', '17': true},
    {'1': 'gender', '3': 4, '4': 1, '5': 5, '9': 3, '10': 'gender', '17': true},
    {'1': 'dateOfBirth', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '9': 4, '10': 'dateOfBirth', '17': true},
    {'1': 'address', '3': 6, '4': 1, '5': 11, '6': '.backend.v1.Address', '9': 5, '10': 'address', '17': true},
    {'1': 'ipAddress', '3': 7, '4': 1, '5': 9, '10': 'ipAddress'},
    {'1': 'placeOfBirth', '3': 8, '4': 1, '5': 9, '9': 6, '10': 'placeOfBirth', '17': true},
    {'1': 'nationality', '3': 9, '4': 1, '5': 9, '9': 7, '10': 'nationality', '17': true},
  ],
  '8': [
    {'1': '_firstName'},
    {'1': '_lastName'},
    {'1': '_countryCode'},
    {'1': '_gender'},
    {'1': '_dateOfBirth'},
    {'1': '_address'},
    {'1': '_placeOfBirth'},
    {'1': '_nationality'},
  ],
};

/// Descriptor for `UpdateIndividualKYCRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateIndividualKYCRequestDescriptor = $convert.base64Decode(
    'ChpVcGRhdGVJbmRpdmlkdWFsS1lDUmVxdWVzdBIhCglmaXJzdE5hbWUYASABKAlIAFIJZmlyc3'
    'ROYW1liAEBEh8KCGxhc3ROYW1lGAIgASgJSAFSCGxhc3ROYW1liAEBEiUKC2NvdW50cnlDb2Rl'
    'GAMgASgJSAJSC2NvdW50cnlDb2RliAEBEhsKBmdlbmRlchgEIAEoBUgDUgZnZW5kZXKIAQESQQ'
    'oLZGF0ZU9mQmlydGgYBSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wSARSC2RhdGVP'
    'ZkJpcnRoiAEBEjIKB2FkZHJlc3MYBiABKAsyEy5iYWNrZW5kLnYxLkFkZHJlc3NIBVIHYWRkcm'
    'Vzc4gBARIcCglpcEFkZHJlc3MYByABKAlSCWlwQWRkcmVzcxInCgxwbGFjZU9mQmlydGgYCCAB'
    'KAlIBlIMcGxhY2VPZkJpcnRoiAEBEiUKC25hdGlvbmFsaXR5GAkgASgJSAdSC25hdGlvbmFsaX'
    'R5iAEBQgwKCl9maXJzdE5hbWVCCwoJX2xhc3ROYW1lQg4KDF9jb3VudHJ5Q29kZUIJCgdfZ2Vu'
    'ZGVyQg4KDF9kYXRlT2ZCaXJ0aEIKCghfYWRkcmVzc0IPCg1fcGxhY2VPZkJpcnRoQg4KDF9uYX'
    'Rpb25hbGl0eQ==');

@$core.Deprecated('Use addressDescriptor instead')
const Address$json = {
  '1': 'Address',
  '2': [
    {'1': 'line1', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'line1', '17': true},
    {'1': 'line2', '3': 2, '4': 1, '5': 9, '9': 1, '10': 'line2', '17': true},
    {'1': 'building', '3': 3, '4': 1, '5': 9, '9': 2, '10': 'building', '17': true},
    {'1': 'apartment', '3': 4, '4': 1, '5': 9, '9': 3, '10': 'apartment', '17': true},
    {'1': 'city', '3': 5, '4': 1, '5': 9, '9': 4, '10': 'city', '17': true},
    {'1': 'state', '3': 6, '4': 1, '5': 9, '9': 5, '10': 'state', '17': true},
    {'1': 'zipCode', '3': 7, '4': 1, '5': 9, '9': 6, '10': 'zipCode', '17': true},
    {'1': 'countryCode', '3': 8, '4': 1, '5': 9, '9': 7, '10': 'countryCode', '17': true},
    {'1': 'placeID', '3': 9, '4': 1, '5': 9, '9': 8, '10': 'placeID', '17': true},
    {'1': 'formattedAddress', '3': 10, '4': 1, '5': 9, '9': 9, '10': 'formattedAddress', '17': true},
  ],
  '8': [
    {'1': '_line1'},
    {'1': '_line2'},
    {'1': '_building'},
    {'1': '_apartment'},
    {'1': '_city'},
    {'1': '_state'},
    {'1': '_zipCode'},
    {'1': '_countryCode'},
    {'1': '_placeID'},
    {'1': '_formattedAddress'},
  ],
};

/// Descriptor for `Address`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addressDescriptor = $convert.base64Decode(
    'CgdBZGRyZXNzEhkKBWxpbmUxGAEgASgJSABSBWxpbmUxiAEBEhkKBWxpbmUyGAIgASgJSAFSBW'
    'xpbmUyiAEBEh8KCGJ1aWxkaW5nGAMgASgJSAJSCGJ1aWxkaW5niAEBEiEKCWFwYXJ0bWVudBgE'
    'IAEoCUgDUglhcGFydG1lbnSIAQESFwoEY2l0eRgFIAEoCUgEUgRjaXR5iAEBEhkKBXN0YXRlGA'
    'YgASgJSAVSBXN0YXRliAEBEh0KB3ppcENvZGUYByABKAlIBlIHemlwQ29kZYgBARIlCgtjb3Vu'
    'dHJ5Q29kZRgIIAEoCUgHUgtjb3VudHJ5Q29kZYgBARIdCgdwbGFjZUlEGAkgASgJSAhSB3BsYW'
    'NlSUSIAQESLwoQZm9ybWF0dGVkQWRkcmVzcxgKIAEoCUgJUhBmb3JtYXR0ZWRBZGRyZXNziAEB'
    'QggKBl9saW5lMUIICgZfbGluZTJCCwoJX2J1aWxkaW5nQgwKCl9hcGFydG1lbnRCBwoFX2NpdH'
    'lCCAoGX3N0YXRlQgoKCF96aXBDb2RlQg4KDF9jb3VudHJ5Q29kZUIKCghfcGxhY2VJREITChFf'
    'Zm9ybWF0dGVkQWRkcmVzcw==');

@$core.Deprecated('Use isUSPSAddressResponseDescriptor instead')
const IsUSPSAddressResponse$json = {
  '1': 'IsUSPSAddressResponse',
  '2': [
    {'1': 'valid', '3': 1, '4': 1, '5': 8, '10': 'valid'},
  ],
};

/// Descriptor for `IsUSPSAddressResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isUSPSAddressResponseDescriptor = $convert.base64Decode(
    'ChVJc1VTUFNBZGRyZXNzUmVzcG9uc2USFAoFdmFsaWQYASABKAhSBXZhbGlk');

@$core.Deprecated('Use getBankAccountWidgetRequestDescriptor instead')
const GetBankAccountWidgetRequest$json = {
  '1': 'GetBankAccountWidgetRequest',
};

/// Descriptor for `GetBankAccountWidgetRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBankAccountWidgetRequestDescriptor = $convert.base64Decode(
    'ChtHZXRCYW5rQWNjb3VudFdpZGdldFJlcXVlc3Q=');

@$core.Deprecated('Use getBankAccountWidgetResponseDescriptor instead')
const GetBankAccountWidgetResponse$json = {
  '1': 'GetBankAccountWidgetResponse',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `GetBankAccountWidgetResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBankAccountWidgetResponseDescriptor = $convert.base64Decode(
    'ChxHZXRCYW5rQWNjb3VudFdpZGdldFJlc3BvbnNlEhAKA3VybBgBIAEoCVIDdXJs');

@$core.Deprecated('Use addBankAccountRequestDescriptor instead')
const AddBankAccountRequest$json = {
  '1': 'AddBankAccountRequest',
  '2': [
    {'1': 'userGuid', '3': 1, '4': 1, '5': 9, '10': 'userGuid'},
    {'1': 'memberGuid', '3': 2, '4': 1, '5': 9, '10': 'memberGuid'},
  ],
};

/// Descriptor for `AddBankAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addBankAccountRequestDescriptor = $convert.base64Decode(
    'ChVBZGRCYW5rQWNjb3VudFJlcXVlc3QSGgoIdXNlckd1aWQYASABKAlSCHVzZXJHdWlkEh4KCm'
    '1lbWJlckd1aWQYAiABKAlSCm1lbWJlckd1aWQ=');

@$core.Deprecated('Use addBankAccountResponseDescriptor instead')
const AddBankAccountResponse$json = {
  '1': 'AddBankAccountResponse',
  '2': [
    {'1': 'fundingsourceId', '3': 1, '4': 1, '5': 9, '10': 'fundingsourceId'},
  ],
};

/// Descriptor for `AddBankAccountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addBankAccountResponseDescriptor = $convert.base64Decode(
    'ChZBZGRCYW5rQWNjb3VudFJlc3BvbnNlEigKD2Z1bmRpbmdzb3VyY2VJZBgBIAEoCVIPZnVuZG'
    'luZ3NvdXJjZUlk');

@$core.Deprecated('Use linkedAccountDescriptor instead')
const LinkedAccount$json = {
  '1': 'LinkedAccount',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'mask', '3': 4, '4': 1, '5': 9, '10': 'mask'},
    {'1': 'nickname', '3': 5, '4': 1, '5': 9, '10': 'nickname'},
    {'1': 'canSend', '3': 6, '4': 1, '5': 8, '10': 'canSend'},
    {'1': 'canReceive', '3': 7, '4': 1, '5': 8, '10': 'canReceive'},
    {'1': 'title', '3': 8, '4': 1, '5': 9, '10': 'title'},
    {'1': 'sendCurrencyCode', '3': 9, '4': 1, '5': 9, '10': 'sendCurrencyCode'},
    {'1': 'sendCurrencyCountryCode', '3': 10, '4': 1, '5': 9, '10': 'sendCurrencyCountryCode'},
    {'1': 'receiveCurrencyCode', '3': 11, '4': 1, '5': 9, '10': 'receiveCurrencyCode'},
    {'1': 'receiveCurrencyCountryCode', '3': 12, '4': 1, '5': 9, '10': 'receiveCurrencyCountryCode'},
    {'1': 'defaultSend', '3': 13, '4': 1, '5': 8, '10': 'defaultSend'},
    {'1': 'defaultReceive', '3': 14, '4': 1, '5': 8, '10': 'defaultReceive'},
    {'1': 'state', '3': 15, '4': 1, '5': 9, '10': 'state'},
  ],
};

/// Descriptor for `LinkedAccount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkedAccountDescriptor = $convert.base64Decode(
    'Cg1MaW5rZWRBY2NvdW50Eg4KAmlkGAEgASgJUgJpZBISCgR0eXBlGAIgASgJUgR0eXBlEhIKBG'
    '5hbWUYAyABKAlSBG5hbWUSEgoEbWFzaxgEIAEoCVIEbWFzaxIaCghuaWNrbmFtZRgFIAEoCVII'
    'bmlja25hbWUSGAoHY2FuU2VuZBgGIAEoCFIHY2FuU2VuZBIeCgpjYW5SZWNlaXZlGAcgASgIUg'
    'pjYW5SZWNlaXZlEhQKBXRpdGxlGAggASgJUgV0aXRsZRIqChBzZW5kQ3VycmVuY3lDb2RlGAkg'
    'ASgJUhBzZW5kQ3VycmVuY3lDb2RlEjgKF3NlbmRDdXJyZW5jeUNvdW50cnlDb2RlGAogASgJUh'
    'dzZW5kQ3VycmVuY3lDb3VudHJ5Q29kZRIwChNyZWNlaXZlQ3VycmVuY3lDb2RlGAsgASgJUhNy'
    'ZWNlaXZlQ3VycmVuY3lDb2RlEj4KGnJlY2VpdmVDdXJyZW5jeUNvdW50cnlDb2RlGAwgASgJUh'
    'pyZWNlaXZlQ3VycmVuY3lDb3VudHJ5Q29kZRIgCgtkZWZhdWx0U2VuZBgNIAEoCFILZGVmYXVs'
    'dFNlbmQSJgoOZGVmYXVsdFJlY2VpdmUYDiABKAhSDmRlZmF1bHRSZWNlaXZlEhQKBXN0YXRlGA'
    '8gASgJUgVzdGF0ZQ==');

@$core.Deprecated('Use getSignupRequestDescriptor instead')
const GetSignupRequest$json = {
  '1': 'GetSignupRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSignupRequestDescriptor = $convert.base64Decode(
    'ChBHZXRTaWdudXBSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use setSignupUserDataRequestDescriptor instead')
const SetSignupUserDataRequest$json = {
  '1': 'SetSignupUserDataRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'id', '17': true},
    {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
    {'1': 'countryCode', '3': 5, '4': 1, '5': 9, '10': 'countryCode'},
  ],
  '8': [
    {'1': '_id'},
  ],
};

/// Descriptor for `SetSignupUserDataRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupUserDataRequestDescriptor = $convert.base64Decode(
    'ChhTZXRTaWdudXBVc2VyRGF0YVJlcXVlc3QSEwoCaWQYASABKAlIAFICaWSIAQESHAoJZmlyc3'
    'ROYW1lGAIgASgJUglmaXJzdE5hbWUSGgoIbGFzdE5hbWUYAyABKAlSCGxhc3ROYW1lEhQKBWVt'
    'YWlsGAQgASgJUgVlbWFpbBIgCgtjb3VudHJ5Q29kZRgFIAEoCVILY291bnRyeUNvZGVCBQoDX2'
    'lk');

@$core.Deprecated('Use setSignupUserDataResponseDescriptor instead')
const SetSignupUserDataResponse$json = {
  '1': 'SetSignupUserDataResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `SetSignupUserDataResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupUserDataResponseDescriptor = $convert.base64Decode(
    'ChlTZXRTaWdudXBVc2VyRGF0YVJlc3BvbnNlEg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use setSignupMobileNumberRequestDescriptor instead')
const SetSignupMobileNumberRequest$json = {
  '1': 'SetSignupMobileNumberRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'mobile', '3': 2, '4': 1, '5': 9, '10': 'mobile'},
    {'1': 'otp', '3': 3, '4': 1, '5': 9, '10': 'otp'},
  ],
};

/// Descriptor for `SetSignupMobileNumberRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupMobileNumberRequestDescriptor = $convert.base64Decode(
    'ChxTZXRTaWdudXBNb2JpbGVOdW1iZXJSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIWCgZtb2JpbG'
    'UYAiABKAlSBm1vYmlsZRIQCgNvdHAYAyABKAlSA290cA==');

@$core.Deprecated('Use signupDescriptor instead')
const Signup$json = {
  '1': 'Signup',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
    {'1': 'countryCode', '3': 5, '4': 1, '5': 9, '10': 'countryCode'},
    {'1': 'mobileNumber', '3': 6, '4': 1, '5': 9, '10': 'mobileNumber'},
    {'1': 'userId', '3': 7, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'completed', '3': 8, '4': 1, '5': 8, '10': 'completed'},
  ],
};

/// Descriptor for `Signup`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signupDescriptor = $convert.base64Decode(
    'CgZTaWdudXASDgoCaWQYASABKAlSAmlkEhwKCWZpcnN0TmFtZRgCIAEoCVIJZmlyc3ROYW1lEh'
    'oKCGxhc3ROYW1lGAMgASgJUghsYXN0TmFtZRIUCgVlbWFpbBgEIAEoCVIFZW1haWwSIAoLY291'
    'bnRyeUNvZGUYBSABKAlSC2NvdW50cnlDb2RlEiIKDG1vYmlsZU51bWJlchgGIAEoCVIMbW9iaW'
    'xlTnVtYmVyEhYKBnVzZXJJZBgHIAEoCVIGdXNlcklkEhwKCWNvbXBsZXRlZBgIIAEoCFIJY29t'
    'cGxldGVk');

@$core.Deprecated('Use completeSignupRequestDescriptor instead')
const CompleteSignupRequest$json = {
  '1': 'CompleteSignupRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'userId', '3': 2, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `CompleteSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeSignupRequestDescriptor = $convert.base64Decode(
    'ChVDb21wbGV0ZVNpZ251cFJlcXVlc3QSDgoCaWQYASABKAlSAmlkEhYKBnVzZXJJZBgCIAEoCV'
    'IGdXNlcklk');

@$core.Deprecated('Use createUserDefaultWalletRequestDescriptor instead')
const CreateUserDefaultWalletRequest$json = {
  '1': 'CreateUserDefaultWalletRequest',
  '2': [
    {'1': 'userID', '3': 1, '4': 1, '5': 9, '10': 'userID'},
  ],
};

/// Descriptor for `CreateUserDefaultWalletRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createUserDefaultWalletRequestDescriptor = $convert.base64Decode(
    'Ch5DcmVhdGVVc2VyRGVmYXVsdFdhbGxldFJlcXVlc3QSFgoGdXNlcklEGAEgASgJUgZ1c2VySU'
    'Q=');

@$core.Deprecated('Use sendPhoneVerificationRequestDescriptor instead')
const SendPhoneVerificationRequest$json = {
  '1': 'SendPhoneVerificationRequest',
  '2': [
    {'1': 'to', '3': 1, '4': 1, '5': 9, '10': 'to'},
  ],
};

/// Descriptor for `SendPhoneVerificationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendPhoneVerificationRequestDescriptor = $convert.base64Decode(
    'ChxTZW5kUGhvbmVWZXJpZmljYXRpb25SZXF1ZXN0Eg4KAnRvGAEgASgJUgJ0bw==');

@$core.Deprecated('Use checkPhoneVerificationRequestDescriptor instead')
const CheckPhoneVerificationRequest$json = {
  '1': 'CheckPhoneVerificationRequest',
  '2': [
    {'1': 'to', '3': 1, '4': 1, '5': 9, '10': 'to'},
    {'1': 'otp', '3': 2, '4': 1, '5': 9, '10': 'otp'},
  ],
};

/// Descriptor for `CheckPhoneVerificationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List checkPhoneVerificationRequestDescriptor = $convert.base64Decode(
    'Ch1DaGVja1Bob25lVmVyaWZpY2F0aW9uUmVxdWVzdBIOCgJ0bxgBIAEoCVICdG8SEAoDb3RwGA'
    'IgASgJUgNvdHA=');

@$core.Deprecated('Use getAgreementRequestDescriptor instead')
const GetAgreementRequest$json = {
  '1': 'GetAgreementRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetAgreementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgreementRequestDescriptor = $convert.base64Decode(
    'ChNHZXRBZ3JlZW1lbnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use agreementDescriptor instead')
const Agreement$json = {
  '1': 'Agreement',
  '2': [
    {'1': 'content', '3': 1, '4': 1, '5': 9, '10': 'content'},
  ],
};

/// Descriptor for `Agreement`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agreementDescriptor = $convert.base64Decode(
    'CglBZ3JlZW1lbnQSGAoHY29udGVudBgBIAEoCVIHY29udGVudA==');

@$core.Deprecated('Use signAgreementsRequestDescriptor instead')
const SignAgreementsRequest$json = {
  '1': 'SignAgreementsRequest',
  '2': [
    {'1': 'agreementIds', '3': 1, '4': 3, '5': 9, '10': 'agreementIds'},
    {'1': 'userId', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'ipAddress', '3': 3, '4': 1, '5': 9, '10': 'ipAddress'},
  ],
};

/// Descriptor for `SignAgreementsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signAgreementsRequestDescriptor = $convert.base64Decode(
    'ChVTaWduQWdyZWVtZW50c1JlcXVlc3QSIgoMYWdyZWVtZW50SWRzGAEgAygJUgxhZ3JlZW1lbn'
    'RJZHMSFgoGdXNlcklkGAIgASgJUgZ1c2VySWQSHAoJaXBBZGRyZXNzGAMgASgJUglpcEFkZHJl'
    'c3M=');

@$core.Deprecated('Use signAgreementsResponseDescriptor instead')
const SignAgreementsResponse$json = {
  '1': 'SignAgreementsResponse',
  '2': [
    {'1': 'signed', '3': 1, '4': 1, '5': 8, '10': 'signed'},
  ],
};

/// Descriptor for `SignAgreementsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signAgreementsResponseDescriptor = $convert.base64Decode(
    'ChZTaWduQWdyZWVtZW50c1Jlc3BvbnNlEhYKBnNpZ25lZBgBIAEoCFIGc2lnbmVk');

@$core.Deprecated('Use joinWaitlistRequestDescriptor instead')
const JoinWaitlistRequest$json = {
  '1': 'JoinWaitlistRequest',
  '2': [
    {'1': 'email', '3': 1, '4': 1, '5': 9, '10': 'email'},
    {'1': 'country_code', '3': 2, '4': 1, '5': 9, '10': 'countryCode'},
    {'1': 'full_name', '3': 3, '4': 1, '5': 9, '10': 'fullName'},
    {'1': 'beta_opt_in', '3': 4, '4': 1, '5': 8, '10': 'betaOptIn'},
    {'1': 'mug_id', '3': 5, '4': 1, '5': 9, '9': 0, '10': 'mugId', '17': true},
  ],
  '8': [
    {'1': '_mug_id'},
  ],
};

/// Descriptor for `JoinWaitlistRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitlistRequestDescriptor = $convert.base64Decode(
    'ChNKb2luV2FpdGxpc3RSZXF1ZXN0EhQKBWVtYWlsGAEgASgJUgVlbWFpbBIhCgxjb3VudHJ5X2'
    'NvZGUYAiABKAlSC2NvdW50cnlDb2RlEhsKCWZ1bGxfbmFtZRgDIAEoCVIIZnVsbE5hbWUSHgoL'
    'YmV0YV9vcHRfaW4YBCABKAhSCWJldGFPcHRJbhIaCgZtdWdfaWQYBSABKAlIAFIFbXVnSWSIAQ'
    'FCCQoHX211Z19pZA==');

@$core.Deprecated('Use joinWaitlistResponseDescriptor instead')
const JoinWaitlistResponse$json = {
  '1': 'JoinWaitlistResponse',
};

/// Descriptor for `JoinWaitlistResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitlistResponseDescriptor = $convert.base64Decode(
    'ChRKb2luV2FpdGxpc3RSZXNwb25zZQ==');

@$core.Deprecated('Use isMugAvailableRequestDescriptor instead')
const IsMugAvailableRequest$json = {
  '1': 'IsMugAvailableRequest',
  '2': [
    {'1': 'mug_id', '3': 1, '4': 1, '5': 9, '10': 'mugId'},
  ],
};

/// Descriptor for `IsMugAvailableRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isMugAvailableRequestDescriptor = $convert.base64Decode(
    'ChVJc011Z0F2YWlsYWJsZVJlcXVlc3QSFQoGbXVnX2lkGAEgASgJUgVtdWdJZA==');

@$core.Deprecated('Use isMugAvailableResponseDescriptor instead')
const IsMugAvailableResponse$json = {
  '1': 'IsMugAvailableResponse',
  '2': [
    {'1': 'available', '3': 1, '4': 1, '5': 8, '10': 'available'},
  ],
};

/// Descriptor for `IsMugAvailableResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isMugAvailableResponseDescriptor = $convert.base64Decode(
    'ChZJc011Z0F2YWlsYWJsZVJlc3BvbnNlEhwKCWF2YWlsYWJsZRgBIAEoCFIJYXZhaWxhYmxl');

@$core.Deprecated('Use getLinkedAccountsResponseDescriptor instead')
const GetLinkedAccountsResponse$json = {
  '1': 'GetLinkedAccountsResponse',
  '2': [
    {'1': 'linkedAccounts', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.LinkedAccount', '10': 'linkedAccounts'},
  ],
};

/// Descriptor for `GetLinkedAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinkedAccountsResponseDescriptor = $convert.base64Decode(
    'ChlHZXRMaW5rZWRBY2NvdW50c1Jlc3BvbnNlEkEKDmxpbmtlZEFjY291bnRzGAEgAygLMhkuYm'
    'Fja2VuZC52MS5MaW5rZWRBY2NvdW50Ug5saW5rZWRBY2NvdW50cw==');

@$core.Deprecated('Use getLinkedAccountRequestDescriptor instead')
const GetLinkedAccountRequest$json = {
  '1': 'GetLinkedAccountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetLinkedAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinkedAccountRequestDescriptor = $convert.base64Decode(
    'ChdHZXRMaW5rZWRBY2NvdW50UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use setNicknameLinkedAccountRequestDescriptor instead')
const SetNicknameLinkedAccountRequest$json = {
  '1': 'SetNicknameLinkedAccountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'nickname', '3': 2, '4': 1, '5': 9, '10': 'nickname'},
  ],
};

/// Descriptor for `SetNicknameLinkedAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setNicknameLinkedAccountRequestDescriptor = $convert.base64Decode(
    'Ch9TZXROaWNrbmFtZUxpbmtlZEFjY291bnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIaCghuaW'
    'NrbmFtZRgCIAEoCVIIbmlja25hbWU=');

@$core.Deprecated('Use deleteLinkedAccountRequestDescriptor instead')
const DeleteLinkedAccountRequest$json = {
  '1': 'DeleteLinkedAccountRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteLinkedAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteLinkedAccountRequestDescriptor = $convert.base64Decode(
    'ChpEZWxldGVMaW5rZWRBY2NvdW50UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use countryDescriptor instead')
const Country$json = {
  '1': 'Country',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `Country`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List countryDescriptor = $convert.base64Decode(
    'CgdDb3VudHJ5Eg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1l');

@$core.Deprecated('Use getCountriesResponseDescriptor instead')
const GetCountriesResponse$json = {
  '1': 'GetCountriesResponse',
  '2': [
    {'1': 'countries', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Country', '10': 'countries'},
  ],
};

/// Descriptor for `GetCountriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCountriesResponseDescriptor = $convert.base64Decode(
    'ChRHZXRDb3VudHJpZXNSZXNwb25zZRIxCgljb3VudHJpZXMYASADKAsyEy5iYWNrZW5kLnYxLk'
    'NvdW50cnlSCWNvdW50cmllcw==');

@$core.Deprecated('Use canSignupRequestDescriptor instead')
const CanSignupRequest$json = {
  '1': 'CanSignupRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `CanSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List canSignupRequestDescriptor = $convert.base64Decode(
    'ChBDYW5TaWdudXBSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use canSignupResponseDescriptor instead')
const CanSignupResponse$json = {
  '1': 'CanSignupResponse',
  '2': [
    {'1': 'canSignup', '3': 1, '4': 1, '5': 8, '10': 'canSignup'},
  ],
};

/// Descriptor for `CanSignupResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List canSignupResponseDescriptor = $convert.base64Decode(
    'ChFDYW5TaWdudXBSZXNwb25zZRIcCgljYW5TaWdudXAYASABKAhSCWNhblNpZ251cA==');

@$core.Deprecated('Use setSignupCompleteRequestDescriptor instead')
const SetSignupCompleteRequest$json = {
  '1': 'SetSignupCompleteRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'userId', '3': 2, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `SetSignupCompleteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupCompleteRequestDescriptor = $convert.base64Decode(
    'ChhTZXRTaWdudXBDb21wbGV0ZVJlcXVlc3QSDgoCaWQYASABKAlSAmlkEhYKBnVzZXJJZBgCIA'
    'EoCVIGdXNlcklk');

@$core.Deprecated('Use lookupTransactionRequestDescriptor instead')
const LookupTransactionRequest$json = {
  '1': 'LookupTransactionRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `LookupTransactionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List lookupTransactionRequestDescriptor = $convert.base64Decode(
    'ChhMb29rdXBUcmFuc2FjdGlvblJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getCurrentWalletResponseDescriptor instead')
const GetCurrentWalletResponse$json = {
  '1': 'GetCurrentWalletResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetCurrentWalletResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCurrentWalletResponseDescriptor = $convert.base64Decode(
    'ChhHZXRDdXJyZW50V2FsbGV0UmVzcG9uc2USDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use limitDescriptor instead')
const Limit$json = {
  '1': 'Limit',
  '2': [
    {'1': 'Annual', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'Annual'},
    {'1': 'Daily', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'Daily'},
    {'1': 'Monthly', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'Monthly'},
    {'1': 'WalletHold', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'WalletHold'},
  ],
};

/// Descriptor for `Limit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List limitDescriptor = $convert.base64Decode(
    'CgVMaW1pdBIvCgZBbm51YWwYASABKAsyFy5iYWNrZW5kLnYxLkxpbWl0QW1vdW50UgZBbm51YW'
    'wSLQoFRGFpbHkYAiABKAsyFy5iYWNrZW5kLnYxLkxpbWl0QW1vdW50UgVEYWlseRIxCgdNb250'
    'aGx5GAMgASgLMhcuYmFja2VuZC52MS5MaW1pdEFtb3VudFIHTW9udGhseRI3CgpXYWxsZXRIb2'
    'xkGAQgASgLMhcuYmFja2VuZC52MS5MaW1pdEFtb3VudFIKV2FsbGV0SG9sZA==');

@$core.Deprecated('Use limitAmountDescriptor instead')
const LimitAmount$json = {
  '1': 'LimitAmount',
  '2': [
    {'1': 'remaining', '3': 1, '4': 1, '5': 9, '10': 'remaining'},
    {'1': 'total', '3': 2, '4': 1, '5': 9, '10': 'total'},
    {'1': 'percentage', '3': 3, '4': 1, '5': 5, '10': 'percentage'},
  ],
};

/// Descriptor for `LimitAmount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List limitAmountDescriptor = $convert.base64Decode(
    'CgtMaW1pdEFtb3VudBIcCglyZW1haW5pbmcYASABKAlSCXJlbWFpbmluZxIUCgV0b3RhbBgCIA'
    'EoCVIFdG90YWwSHgoKcGVyY2VudGFnZRgDIAEoBVIKcGVyY2VudGFnZQ==');

@$core.Deprecated('Use walletAddressExistsRequestDescriptor instead')
const WalletAddressExistsRequest$json = {
  '1': 'WalletAddressExistsRequest',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `WalletAddressExistsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletAddressExistsRequestDescriptor = $convert.base64Decode(
    'ChpXYWxsZXRBZGRyZXNzRXhpc3RzUmVxdWVzdBIQCgN1cmwYASABKAlSA3VybA==');

@$core.Deprecated('Use walletAddressExistsResponseDescriptor instead')
const WalletAddressExistsResponse$json = {
  '1': 'WalletAddressExistsResponse',
  '2': [
    {'1': 'exists', '3': 1, '4': 1, '5': 8, '10': 'exists'},
  ],
};

/// Descriptor for `WalletAddressExistsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletAddressExistsResponseDescriptor = $convert.base64Decode(
    'ChtXYWxsZXRBZGRyZXNzRXhpc3RzUmVzcG9uc2USFgoGZXhpc3RzGAEgASgIUgZleGlzdHM=');

@$core.Deprecated('Use createWalletAddressRequestDescriptor instead')
const CreateWalletAddressRequest$json = {
  '1': 'CreateWalletAddressRequest',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
    {'1': 'asset', '3': 2, '4': 1, '5': 9, '10': 'asset'},
    {'1': 'assetScale', '3': 3, '4': 1, '5': 5, '10': 'assetScale'},
    {'1': 'alias', '3': 4, '4': 1, '5': 9, '10': 'alias'},
  ],
};

/// Descriptor for `CreateWalletAddressRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWalletAddressRequestDescriptor = $convert.base64Decode(
    'ChpDcmVhdGVXYWxsZXRBZGRyZXNzUmVxdWVzdBIQCgN1cmwYASABKAlSA3VybBIUCgVhc3NldB'
    'gCIAEoCVIFYXNzZXQSHgoKYXNzZXRTY2FsZRgDIAEoBVIKYXNzZXRTY2FsZRIUCgVhbGlhcxgE'
    'IAEoCVIFYWxpYXM=');

@$core.Deprecated('Use setWalletNameRequestDescriptor instead')
const SetWalletNameRequest$json = {
  '1': 'SetWalletNameRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `SetWalletNameRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setWalletNameRequestDescriptor = $convert.base64Decode(
    'ChRTZXRXYWxsZXROYW1lUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use getPublicWalletDetailsRequestDescriptor instead')
const GetPublicWalletDetailsRequest$json = {
  '1': 'GetPublicWalletDetailsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetPublicWalletDetailsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPublicWalletDetailsRequestDescriptor = $convert.base64Decode(
    'Ch1HZXRQdWJsaWNXYWxsZXREZXRhaWxzUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use getPublicWalletDetailsResponseDescriptor instead')
const GetPublicWalletDetailsResponse$json = {
  '1': 'GetPublicWalletDetailsResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'publicName', '3': 2, '4': 1, '5': 9, '10': 'publicName'},
  ],
};

/// Descriptor for `GetPublicWalletDetailsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPublicWalletDetailsResponseDescriptor = $convert.base64Decode(
    'Ch5HZXRQdWJsaWNXYWxsZXREZXRhaWxzUmVzcG9uc2USDgoCaWQYASABKAlSAmlkEh4KCnB1Ym'
    'xpY05hbWUYAiABKAlSCnB1YmxpY05hbWU=');

@$core.Deprecated('Use listLimitsResponseDescriptor instead')
const ListLimitsResponse$json = {
  '1': 'ListLimitsResponse',
  '2': [
    {'1': 'limits', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.ConfiguredLimit', '10': 'limits'},
  ],
};

/// Descriptor for `ListLimitsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLimitsResponseDescriptor = $convert.base64Decode(
    'ChJMaXN0TGltaXRzUmVzcG9uc2USMwoGbGltaXRzGAEgAygLMhsuYmFja2VuZC52MS5Db25maW'
    'd1cmVkTGltaXRSBmxpbWl0cw==');

@$core.Deprecated('Use configuredLimitDescriptor instead')
const ConfiguredLimit$json = {
  '1': 'ConfiguredLimit',
  '2': [
    {'1': 'foreignId', '3': 1, '4': 1, '5': 9, '10': 'foreignId'},
    {'1': 'foreignDisplay', '3': 2, '4': 1, '5': 9, '10': 'foreignDisplay'},
    {'1': 'foreignType', '3': 3, '4': 1, '5': 9, '10': 'foreignType'},
    {'1': 'daily', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'daily'},
    {'1': 'monthly', '3': 5, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'monthly'},
    {'1': 'overall', '3': 6, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'overall'},
  ],
};

/// Descriptor for `ConfiguredLimit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configuredLimitDescriptor = $convert.base64Decode(
    'Cg9Db25maWd1cmVkTGltaXQSHAoJZm9yZWlnbklkGAEgASgJUglmb3JlaWduSWQSJgoOZm9yZW'
    'lnbkRpc3BsYXkYAiABKAlSDmZvcmVpZ25EaXNwbGF5EiAKC2ZvcmVpZ25UeXBlGAMgASgJUgtm'
    'b3JlaWduVHlwZRIoCgVkYWlseRgEIAEoCzISLmJhY2tlbmQudjEuQW1vdW50UgVkYWlseRIsCg'
    'dtb250aGx5GAUgASgLMhIuYmFja2VuZC52MS5BbW91bnRSB21vbnRobHkSLAoHb3ZlcmFsbBgG'
    'IAEoCzISLmJhY2tlbmQudjEuQW1vdW50UgdvdmVyYWxs');

@$core.Deprecated('Use updateClientLimitsRequestDescriptor instead')
const UpdateClientLimitsRequest$json = {
  '1': 'UpdateClientLimitsRequest',
  '2': [
    {'1': 'clientUrl', '3': 1, '4': 1, '5': 9, '10': 'clientUrl'},
    {'1': 'daily', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'daily'},
    {'1': 'monthly', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'monthly'},
    {'1': 'overall', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'overall'},
  ],
};

/// Descriptor for `UpdateClientLimitsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateClientLimitsRequestDescriptor = $convert.base64Decode(
    'ChlVcGRhdGVDbGllbnRMaW1pdHNSZXF1ZXN0EhwKCWNsaWVudFVybBgBIAEoCVIJY2xpZW50VX'
    'JsEigKBWRhaWx5GAIgASgLMhIuYmFja2VuZC52MS5BbW91bnRSBWRhaWx5EiwKB21vbnRobHkY'
    'AyABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIHbW9udGhseRIsCgdvdmVyYWxsGAQgASgLMhIuYm'
    'Fja2VuZC52MS5BbW91bnRSB292ZXJhbGw=');

@$core.Deprecated('Use contactDescriptor instead')
const Contact$json = {
  '1': 'Contact',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'payment_pointer', '3': 2, '4': 1, '5': 9, '10': 'paymentPointer'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'wallet_id', '3': 4, '4': 1, '5': 9, '10': 'walletId'},
  ],
};

/// Descriptor for `Contact`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List contactDescriptor = $convert.base64Decode(
    'CgdDb250YWN0Eg4KAmlkGAEgASgJUgJpZBInCg9wYXltZW50X3BvaW50ZXIYAiABKAlSDnBheW'
    '1lbnRQb2ludGVyEhIKBG5hbWUYAyABKAlSBG5hbWUSGwoJd2FsbGV0X2lkGAQgASgJUgh3YWxs'
    'ZXRJZA==');

@$core.Deprecated('Use listContactsRequestDescriptor instead')
const ListContactsRequest$json = {
  '1': 'ListContactsRequest',
  '2': [
    {'1': 'page_size', '3': 1, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'page_token', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'pageToken', '17': true},
    {'1': 'order_by', '3': 3, '4': 1, '5': 9, '9': 1, '10': 'orderBy', '17': true},
  ],
  '8': [
    {'1': '_page_token'},
    {'1': '_order_by'},
  ],
};

/// Descriptor for `ListContactsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listContactsRequestDescriptor = $convert.base64Decode(
    'ChNMaXN0Q29udGFjdHNSZXF1ZXN0EhsKCXBhZ2Vfc2l6ZRgBIAEoBVIIcGFnZVNpemUSIgoKcG'
    'FnZV90b2tlbhgCIAEoCUgAUglwYWdlVG9rZW6IAQESHgoIb3JkZXJfYnkYAyABKAlIAVIHb3Jk'
    'ZXJCeYgBAUINCgtfcGFnZV90b2tlbkILCglfb3JkZXJfYnk=');

@$core.Deprecated('Use listContactsResponseDescriptor instead')
const ListContactsResponse$json = {
  '1': 'ListContactsResponse',
  '2': [
    {'1': 'contacts', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Contact', '10': 'contacts'},
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListContactsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listContactsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0Q29udGFjdHNSZXNwb25zZRIvCghjb250YWN0cxgBIAMoCzITLmJhY2tlbmQudjEuQ2'
    '9udGFjdFIIY29udGFjdHMSJgoPbmV4dF9wYWdlX3Rva2VuGAIgASgJUg1uZXh0UGFnZVRva2Vu');

@$core.Deprecated('Use createContactRequestDescriptor instead')
const CreateContactRequest$json = {
  '1': 'CreateContactRequest',
  '2': [
    {'1': 'payment_pointer', '3': 1, '4': 1, '5': 9, '10': 'paymentPointer'},
  ],
};

/// Descriptor for `CreateContactRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createContactRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVDb250YWN0UmVxdWVzdBInCg9wYXltZW50X3BvaW50ZXIYASABKAlSDnBheW1lbn'
    'RQb2ludGVy');

@$core.Deprecated('Use listIdentitiesResponseDescriptor instead')
const ListIdentitiesResponse$json = {
  '1': 'ListIdentitiesResponse',
  '2': [
    {'1': 'identities', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Identity', '10': 'identities'},
  ],
};

/// Descriptor for `ListIdentitiesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listIdentitiesResponseDescriptor = $convert.base64Decode(
    'ChZMaXN0SWRlbnRpdGllc1Jlc3BvbnNlEjQKCmlkZW50aXRpZXMYASADKAsyFC5iYWNrZW5kLn'
    'YxLklkZW50aXR5UgppZGVudGl0aWVz');

@$core.Deprecated('Use identityDescriptor instead')
const Identity$json = {
  '1': 'Identity',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'wallet', '3': 2, '4': 1, '5': 9, '10': 'wallet'},
    {'1': 'platform', '3': 3, '4': 1, '5': 9, '10': 'platform'},
    {'1': 'identifier', '3': 4, '4': 1, '5': 9, '10': 'identifier'},
    {'1': 'state', '3': 5, '4': 1, '5': 9, '10': 'state'},
    {'1': 'key_id', '3': 6, '4': 1, '5': 9, '10': 'keyId'},
    {'1': 'signature', '3': 7, '4': 1, '5': 9, '10': 'signature'},
    {'1': 'signature_hash', '3': 8, '4': 1, '5': 9, '10': 'signatureHash'},
    {'1': 'proof', '3': 9, '4': 1, '5': 9, '10': 'proof'},
    {'1': 'ctime', '3': 10, '4': 1, '5': 9, '10': 'ctime'},
    {'1': 'public', '3': 11, '4': 1, '5': 8, '10': 'public'},
    {'1': 'verified_at', '3': 12, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'verifiedAt'},
    {'1': 'wallet_id', '3': 13, '4': 1, '5': 9, '10': 'walletId'},
    {'1': 'txt_record', '3': 14, '4': 1, '5': 9, '9': 0, '10': 'txtRecord', '17': true},
  ],
  '8': [
    {'1': '_txt_record'},
  ],
};

/// Descriptor for `Identity`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List identityDescriptor = $convert.base64Decode(
    'CghJZGVudGl0eRIOCgJpZBgBIAEoCVICaWQSFgoGd2FsbGV0GAIgASgJUgZ3YWxsZXQSGgoIcG'
    'xhdGZvcm0YAyABKAlSCHBsYXRmb3JtEh4KCmlkZW50aWZpZXIYBCABKAlSCmlkZW50aWZpZXIS'
    'FAoFc3RhdGUYBSABKAlSBXN0YXRlEhUKBmtleV9pZBgGIAEoCVIFa2V5SWQSHAoJc2lnbmF0dX'
    'JlGAcgASgJUglzaWduYXR1cmUSJQoOc2lnbmF0dXJlX2hhc2gYCCABKAlSDXNpZ25hdHVyZUhh'
    'c2gSFAoFcHJvb2YYCSABKAlSBXByb29mEhQKBWN0aW1lGAogASgJUgVjdGltZRIWCgZwdWJsaW'
    'MYCyABKAhSBnB1YmxpYxI7Cgt2ZXJpZmllZF9hdBgMIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5U'
    'aW1lc3RhbXBSCnZlcmlmaWVkQXQSGwoJd2FsbGV0X2lkGA0gASgJUgh3YWxsZXRJZBIiCgp0eH'
    'RfcmVjb3JkGA4gASgJSABSCXR4dFJlY29yZIgBAUINCgtfdHh0X3JlY29yZA==');

@$core.Deprecated('Use identityVerificationInstructionsDescriptor instead')
const IdentityVerificationInstructions$json = {
  '1': 'IdentityVerificationInstructions',
  '2': [
    {'1': 'identity_id', '3': 1, '4': 1, '5': 9, '10': 'identityId'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
    {'1': 'instructions', '3': 3, '4': 1, '5': 9, '10': 'instructions'},
  ],
};

/// Descriptor for `IdentityVerificationInstructions`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List identityVerificationInstructionsDescriptor = $convert.base64Decode(
    'CiBJZGVudGl0eVZlcmlmaWNhdGlvbkluc3RydWN0aW9ucxIfCgtpZGVudGl0eV9pZBgBIAEoCV'
    'IKaWRlbnRpdHlJZBISCgRjb2RlGAIgASgJUgRjb2RlEiIKDGluc3RydWN0aW9ucxgDIAEoCVIM'
    'aW5zdHJ1Y3Rpb25z');

@$core.Deprecated('Use deleteIdentityRequestDescriptor instead')
const DeleteIdentityRequest$json = {
  '1': 'DeleteIdentityRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteIdentityRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteIdentityRequestDescriptor = $convert.base64Decode(
    'ChVEZWxldGVJZGVudGl0eVJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use setIdentityPublicRequestDescriptor instead')
const SetIdentityPublicRequest$json = {
  '1': 'SetIdentityPublicRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'public', '3': 2, '4': 1, '5': 8, '10': 'public'},
  ],
};

/// Descriptor for `SetIdentityPublicRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setIdentityPublicRequestDescriptor = $convert.base64Decode(
    'ChhTZXRJZGVudGl0eVB1YmxpY1JlcXVlc3QSDgoCaWQYASABKAlSAmlkEhYKBnB1YmxpYxgCIA'
    'EoCFIGcHVibGlj');

@$core.Deprecated('Use listPublicIdentitiesRequestDescriptor instead')
const ListPublicIdentitiesRequest$json = {
  '1': 'ListPublicIdentitiesRequest',
  '2': [
    {'1': 'wallet_id', '3': 1, '4': 1, '5': 9, '10': 'walletId'},
  ],
};

/// Descriptor for `ListPublicIdentitiesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listPublicIdentitiesRequestDescriptor = $convert.base64Decode(
    'ChtMaXN0UHVibGljSWRlbnRpdGllc1JlcXVlc3QSGwoJd2FsbGV0X2lkGAEgASgJUgh3YWxsZX'
    'RJZA==');

@$core.Deprecated('Use kYCStatusResponseDescriptor instead')
const KYCStatusResponse$json = {
  '1': 'KYCStatusResponse',
  '2': [
    {'1': 'kyc_status', '3': 1, '4': 1, '5': 5, '10': 'kycStatus'},
  ],
};

/// Descriptor for `KYCStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List kYCStatusResponseDescriptor = $convert.base64Decode(
    'ChFLWUNTdGF0dXNSZXNwb25zZRIdCgpreWNfc3RhdHVzGAEgASgFUglreWNTdGF0dXM=');

@$core.Deprecated('Use kYCPersonaInquiryRequestDescriptor instead')
const KYCPersonaInquiryRequest$json = {
  '1': 'KYCPersonaInquiryRequest',
  '2': [
    {'1': 'idempotency_key', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'idempotencyKey', '17': true},
  ],
  '8': [
    {'1': '_idempotency_key'},
  ],
};

/// Descriptor for `KYCPersonaInquiryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List kYCPersonaInquiryRequestDescriptor = $convert.base64Decode(
    'ChhLWUNQZXJzb25hSW5xdWlyeVJlcXVlc3QSLAoPaWRlbXBvdGVuY3lfa2V5GAEgASgJSABSDm'
    'lkZW1wb3RlbmN5S2V5iAEBQhIKEF9pZGVtcG90ZW5jeV9rZXk=');

@$core.Deprecated('Use kYCPersonaInquiryResponseDescriptor instead')
const KYCPersonaInquiryResponse$json = {
  '1': 'KYCPersonaInquiryResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'session_token',
      '3': 2,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '9': 0,
      '10': 'sessionToken',
      '17': true,
    },
  ],
  '8': [
    {'1': '_session_token'},
  ],
};

/// Descriptor for `KYCPersonaInquiryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List kYCPersonaInquiryResponseDescriptor = $convert.base64Decode(
    'ChlLWUNQZXJzb25hSW5xdWlyeVJlc3BvbnNlEg4KAmlkGAEgASgJUgJpZBIsCg1zZXNzaW9uX3'
    'Rva2VuGAIgASgJQgIYAUgAUgxzZXNzaW9uVG9rZW6IAQFCEAoOX3Nlc3Npb25fdG9rZW4=');

@$core.Deprecated('Use createTwitterAuthURLResponseDescriptor instead')
const CreateTwitterAuthURLResponse$json = {
  '1': 'CreateTwitterAuthURLResponse',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `CreateTwitterAuthURLResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTwitterAuthURLResponseDescriptor = $convert.base64Decode(
    'ChxDcmVhdGVUd2l0dGVyQXV0aFVSTFJlc3BvbnNlEhAKA3VybBgBIAEoCVIDdXJs');

@$core.Deprecated('Use twitterCallbackRequestDescriptor instead')
const TwitterCallbackRequest$json = {
  '1': 'TwitterCallbackRequest',
  '2': [
    {'1': 'state', '3': 1, '4': 1, '5': 9, '10': 'state'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
  ],
};

/// Descriptor for `TwitterCallbackRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List twitterCallbackRequestDescriptor = $convert.base64Decode(
    'ChZUd2l0dGVyQ2FsbGJhY2tSZXF1ZXN0EhQKBXN0YXRlGAEgASgJUgVzdGF0ZRISCgRjb2RlGA'
    'IgASgJUgRjb2Rl');

@$core.Deprecated('Use twitterCallbackResponseDescriptor instead')
const TwitterCallbackResponse$json = {
  '1': 'TwitterCallbackResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `TwitterCallbackResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List twitterCallbackResponseDescriptor = $convert.base64Decode(
    'ChdUd2l0dGVyQ2FsbGJhY2tSZXNwb25zZRIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use discordCallbackRequestDescriptor instead')
const DiscordCallbackRequest$json = {
  '1': 'DiscordCallbackRequest',
  '2': [
    {'1': 'state', '3': 1, '4': 1, '5': 9, '10': 'state'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
  ],
};

/// Descriptor for `DiscordCallbackRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List discordCallbackRequestDescriptor = $convert.base64Decode(
    'ChZEaXNjb3JkQ2FsbGJhY2tSZXF1ZXN0EhQKBXN0YXRlGAEgASgJUgVzdGF0ZRISCgRjb2RlGA'
    'IgASgJUgRjb2Rl');

@$core.Deprecated('Use discordCallbackResponseDescriptor instead')
const DiscordCallbackResponse$json = {
  '1': 'DiscordCallbackResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DiscordCallbackResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List discordCallbackResponseDescriptor = $convert.base64Decode(
    'ChdEaXNjb3JkQ2FsbGJhY2tSZXNwb25zZRIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use createDiscordAuthURLResponseDescriptor instead')
const CreateDiscordAuthURLResponse$json = {
  '1': 'CreateDiscordAuthURLResponse',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `CreateDiscordAuthURLResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDiscordAuthURLResponseDescriptor = $convert.base64Decode(
    'ChxDcmVhdGVEaXNjb3JkQXV0aFVSTFJlc3BvbnNlEhAKA3VybBgBIAEoCVIDdXJs');

@$core.Deprecated('Use getIdentityRequestDescriptor instead')
const GetIdentityRequest$json = {
  '1': 'GetIdentityRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetIdentityRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getIdentityRequestDescriptor = $convert.base64Decode(
    'ChJHZXRJZGVudGl0eVJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getIdentityResponseDescriptor instead')
const GetIdentityResponse$json = {
  '1': 'GetIdentityResponse',
  '2': [
    {'1': 'identity', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.Identity', '10': 'identity'},
  ],
};

/// Descriptor for `GetIdentityResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getIdentityResponseDescriptor = $convert.base64Decode(
    'ChNHZXRJZGVudGl0eVJlc3BvbnNlEjAKCGlkZW50aXR5GAEgASgLMhQuYmFja2VuZC52MS5JZG'
    'VudGl0eVIIaWRlbnRpdHk=');

@$core.Deprecated('Use getIdentityBySignatureHashRequestDescriptor instead')
const GetIdentityBySignatureHashRequest$json = {
  '1': 'GetIdentityBySignatureHashRequest',
  '2': [
    {'1': 'signature_hash', '3': 1, '4': 1, '5': 9, '10': 'signatureHash'},
  ],
};

/// Descriptor for `GetIdentityBySignatureHashRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getIdentityBySignatureHashRequestDescriptor = $convert.base64Decode(
    'CiFHZXRJZGVudGl0eUJ5U2lnbmF0dXJlSGFzaFJlcXVlc3QSJQoOc2lnbmF0dXJlX2hhc2gYAS'
    'ABKAlSDXNpZ25hdHVyZUhhc2g=');

@$core.Deprecated('Use getPaymentAddressRequestDescriptor instead')
const GetPaymentAddressRequest$json = {
  '1': 'GetPaymentAddressRequest',
  '2': [
    {'1': 'address', '3': 1, '4': 1, '5': 9, '10': 'address'},
  ],
};

/// Descriptor for `GetPaymentAddressRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPaymentAddressRequestDescriptor = $convert.base64Decode(
    'ChhHZXRQYXltZW50QWRkcmVzc1JlcXVlc3QSGAoHYWRkcmVzcxgBIAEoCVIHYWRkcmVzcw==');

@$core.Deprecated('Use getPaymentAddressResponseDescriptor instead')
const GetPaymentAddressResponse$json = {
  '1': 'GetPaymentAddressResponse',
  '2': [
    {'1': 'wallet_url', '3': 1, '4': 1, '5': 9, '10': 'walletUrl'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'handle', '3': 3, '4': 1, '5': 9, '10': 'handle'},
    {'1': 'canSendToAddress', '3': 4, '4': 1, '5': 8, '10': 'canSendToAddress'},
  ],
};

/// Descriptor for `GetPaymentAddressResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPaymentAddressResponseDescriptor = $convert.base64Decode(
    'ChlHZXRQYXltZW50QWRkcmVzc1Jlc3BvbnNlEh0KCndhbGxldF91cmwYASABKAlSCXdhbGxldF'
    'VybBISCgR0eXBlGAIgASgJUgR0eXBlEhYKBmhhbmRsZRgDIAEoCVIGaGFuZGxlEioKEGNhblNl'
    'bmRUb0FkZHJlc3MYBCABKAhSEGNhblNlbmRUb0FkZHJlc3M=');

@$core.Deprecated('Use createDomainIdentityRequestDescriptor instead')
const CreateDomainIdentityRequest$json = {
  '1': 'CreateDomainIdentityRequest',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `CreateDomainIdentityRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDomainIdentityRequestDescriptor = $convert.base64Decode(
    'ChtDcmVhdGVEb21haW5JZGVudGl0eVJlcXVlc3QSEAoDdXJsGAEgASgJUgN1cmw=');

@$core.Deprecated('Use createDomainIdentityResponseDescriptor instead')
const CreateDomainIdentityResponse$json = {
  '1': 'CreateDomainIdentityResponse',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `CreateDomainIdentityResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createDomainIdentityResponseDescriptor = $convert.base64Decode(
    'ChxDcmVhdGVEb21haW5JZGVudGl0eVJlc3BvbnNlEg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use verifyIdentityRequestDescriptor instead')
const VerifyIdentityRequest$json = {
  '1': 'VerifyIdentityRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `VerifyIdentityRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List verifyIdentityRequestDescriptor = $convert.base64Decode(
    'ChVWZXJpZnlJZGVudGl0eVJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use submitFormRequestDescriptor instead')
const SubmitFormRequest$json = {
  '1': 'SubmitFormRequest',
  '2': [
    {'1': 'form_id', '3': 1, '4': 1, '5': 9, '10': 'formId'},
    {'1': 'data', '3': 2, '4': 1, '5': 9, '10': 'data'},
  ],
};

/// Descriptor for `SubmitFormRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List submitFormRequestDescriptor = $convert.base64Decode(
    'ChFTdWJtaXRGb3JtUmVxdWVzdBIXCgdmb3JtX2lkGAEgASgJUgZmb3JtSWQSEgoEZGF0YRgCIA'
    'EoCVIEZGF0YQ==');

