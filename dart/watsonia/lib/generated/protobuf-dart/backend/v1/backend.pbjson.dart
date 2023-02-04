///
//  Generated code. Do not modify.
//  source: backend/v1/backend.proto
//
// @dart = 2.12
// ignore_for_file: annotate_overrides,camel_case_types,constant_identifier_names,deprecated_member_use_from_same_package,directives_ordering,library_prefixes,non_constant_identifier_names,prefer_final_fields,return_of_invalid_type,unnecessary_const,unnecessary_import,unnecessary_this,unused_import,unused_shown_name

import 'dart:core' as $core;
import 'dart:convert' as $convert;
import 'dart:typed_data' as $typed_data;
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
@$core.Deprecated('Use transactionDescriptor instead')
const Transaction$json = const {
  '1': 'Transaction',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    const {'1': 'amount', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'amount'},
    const {'1': 'source', '3': 4, '4': 1, '5': 9, '10': 'source'},
    const {'1': 'destination', '3': 5, '4': 1, '5': 9, '10': 'destination'},
    const {'1': 'timestamp', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
    const {'1': 'state', '3': 7, '4': 1, '5': 9, '10': 'state'},
    const {'1': 'transfers', '3': 8, '4': 3, '5': 11, '6': '.backend.v1.Transfer', '10': 'transfers'},
    const {'1': 'foreignId', '3': 9, '4': 1, '5': 9, '10': 'foreignId'},
  ],
};

/// Descriptor for `Transaction`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transactionDescriptor = $convert.base64Decode('CgtUcmFuc2FjdGlvbhIOCgJpZBgBIAEoCVICaWQSEgoEdHlwZRgCIAEoCVIEdHlwZRIqCgZhbW91bnQYAyABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIGYW1vdW50EhYKBnNvdXJjZRgEIAEoCVIGc291cmNlEiAKC2Rlc3RpbmF0aW9uGAUgASgJUgtkZXN0aW5hdGlvbhI4Cgl0aW1lc3RhbXAYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl0aW1lc3RhbXASFAoFc3RhdGUYByABKAlSBXN0YXRlEjIKCXRyYW5zZmVycxgIIAMoCzIULmJhY2tlbmQudjEuVHJhbnNmZXJSCXRyYW5zZmVycxIcCglmb3JlaWduSWQYCSABKAlSCWZvcmVpZ25JZA==');
@$core.Deprecated('Use listTransactionsResponseDescriptor instead')
const ListTransactionsResponse$json = const {
  '1': 'ListTransactionsResponse',
  '2': const [
    const {'1': 'transactions', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Transaction', '10': 'transactions'},
    const {'1': 'page', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.PaginationResponse', '10': 'page'},
  ],
};

/// Descriptor for `ListTransactionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTransactionsResponseDescriptor = $convert.base64Decode('ChhMaXN0VHJhbnNhY3Rpb25zUmVzcG9uc2USOwoMdHJhbnNhY3Rpb25zGAEgAygLMhcuYmFja2VuZC52MS5UcmFuc2FjdGlvblIMdHJhbnNhY3Rpb25zEjIKBHBhZ2UYAiABKAsyHi5iYWNrZW5kLnYxLlBhZ2luYXRpb25SZXNwb25zZVIEcGFnZQ==');
@$core.Deprecated('Use amountDescriptor instead')
const Amount$json = const {
  '1': 'Amount',
  '2': const [
    const {'1': 'amount', '3': 1, '4': 1, '5': 4, '10': 'amount'},
    const {'1': 'asset', '3': 2, '4': 1, '5': 9, '10': 'asset'},
    const {'1': 'assetScale', '3': 3, '4': 1, '5': 5, '10': 'assetScale'},
  ],
};

/// Descriptor for `Amount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List amountDescriptor = $convert.base64Decode('CgZBbW91bnQSFgoGYW1vdW50GAEgASgEUgZhbW91bnQSFAoFYXNzZXQYAiABKAlSBWFzc2V0Eh4KCmFzc2V0U2NhbGUYAyABKAVSCmFzc2V0U2NhbGU=');
@$core.Deprecated('Use lookupOutgoingPaymentRequestDescriptor instead')
const LookupOutgoingPaymentRequest$json = const {
  '1': 'LookupOutgoingPaymentRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `LookupOutgoingPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List lookupOutgoingPaymentRequestDescriptor = $convert.base64Decode('ChxMb29rdXBPdXRnb2luZ1BheW1lbnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');
@$core.Deprecated('Use outgoingPaymentDescriptor instead')
const OutgoingPayment$json = const {
  '1': 'OutgoingPayment',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'paymentPointer', '3': 2, '4': 1, '5': 9, '10': 'paymentPointer'},
    const {'1': 'toPaymentPointer', '3': 11, '4': 1, '5': 9, '10': 'toPaymentPointer'},
    const {'1': 'failed', '3': 3, '4': 1, '5': 8, '10': 'failed'},
    const {'1': 'receiver', '3': 4, '4': 1, '5': 9, '10': 'receiver'},
    const {'1': 'sendAmount', '3': 5, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'sendAmount'},
    const {'1': 'receiveAmount', '3': 6, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'receiveAmount'},
    const {'1': 'sentAmount', '3': 7, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'sentAmount'},
    const {'1': 'description', '3': 8, '4': 1, '5': 9, '10': 'description'},
    const {'1': 'createdAt', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'createdAt'},
    const {'1': 'updatedAt', '3': 10, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'updatedAt'},
  ],
};

/// Descriptor for `OutgoingPayment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List outgoingPaymentDescriptor = $convert.base64Decode('Cg9PdXRnb2luZ1BheW1lbnQSDgoCaWQYASABKAlSAmlkEiYKDnBheW1lbnRQb2ludGVyGAIgASgJUg5wYXltZW50UG9pbnRlchIqChB0b1BheW1lbnRQb2ludGVyGAsgASgJUhB0b1BheW1lbnRQb2ludGVyEhYKBmZhaWxlZBgDIAEoCFIGZmFpbGVkEhoKCHJlY2VpdmVyGAQgASgJUghyZWNlaXZlchIyCgpzZW5kQW1vdW50GAUgASgLMhIuYmFja2VuZC52MS5BbW91bnRSCnNlbmRBbW91bnQSOAoNcmVjZWl2ZUFtb3VudBgGIAEoCzISLmJhY2tlbmQudjEuQW1vdW50Ug1yZWNlaXZlQW1vdW50EjIKCnNlbnRBbW91bnQYByABKAsyEi5iYWNrZW5kLnYxLkFtb3VudFIKc2VudEFtb3VudBIgCgtkZXNjcmlwdGlvbhgIIAEoCVILZGVzY3JpcHRpb24SOAoJY3JlYXRlZEF0GAkgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJY3JlYXRlZEF0EjgKCXVwZGF0ZWRBdBgKIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXVwZGF0ZWRBdA==');
@$core.Deprecated('Use preCheckOutgoingPaymentRequestDescriptor instead')
const PreCheckOutgoingPaymentRequest$json = const {
  '1': 'PreCheckOutgoingPaymentRequest',
  '2': const [
    const {'1': 'quoteID', '3': 1, '4': 1, '5': 9, '10': 'quoteID'},
  ],
};

/// Descriptor for `PreCheckOutgoingPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List preCheckOutgoingPaymentRequestDescriptor = $convert.base64Decode('Ch5QcmVDaGVja091dGdvaW5nUGF5bWVudFJlcXVlc3QSGAoHcXVvdGVJRBgBIAEoCVIHcXVvdGVJRA==');
@$core.Deprecated('Use preCheckOutgoingPaymentResponseDescriptor instead')
const PreCheckOutgoingPaymentResponse$json = const {
  '1': 'PreCheckOutgoingPaymentResponse',
  '2': const [
    const {'1': 'exceedsLimits', '3': 1, '4': 1, '5': 8, '10': 'exceedsLimits'},
    const {'1': 'limitType', '3': 2, '4': 1, '5': 9, '10': 'limitType'},
    const {'1': 'insufficientBalance', '3': 3, '4': 1, '5': 8, '10': 'insufficientBalance'},
  ],
};

/// Descriptor for `PreCheckOutgoingPaymentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List preCheckOutgoingPaymentResponseDescriptor = $convert.base64Decode('Ch9QcmVDaGVja091dGdvaW5nUGF5bWVudFJlc3BvbnNlEiQKDWV4Y2VlZHNMaW1pdHMYASABKAhSDWV4Y2VlZHNMaW1pdHMSHAoJbGltaXRUeXBlGAIgASgJUglsaW1pdFR5cGUSMAoTaW5zdWZmaWNpZW50QmFsYW5jZRgDIAEoCFITaW5zdWZmaWNpZW50QmFsYW5jZQ==');
@$core.Deprecated('Use createOutgoingPaymentRequestDescriptor instead')
const CreateOutgoingPaymentRequest$json = const {
  '1': 'CreateOutgoingPaymentRequest',
  '2': const [
    const {'1': 'quoteID', '3': 1, '4': 1, '5': 9, '10': 'quoteID'},
    const {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    const {'1': 'externalRef', '3': 3, '4': 1, '5': 9, '10': 'externalRef'},
    const {'1': 'ipAddress', '3': 4, '4': 1, '5': 9, '10': 'ipAddress'},
    const {'1': 'idempotencyKey', '3': 5, '4': 1, '5': 9, '10': 'idempotencyKey'},
  ],
};

/// Descriptor for `CreateOutgoingPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createOutgoingPaymentRequestDescriptor = $convert.base64Decode('ChxDcmVhdGVPdXRnb2luZ1BheW1lbnRSZXF1ZXN0EhgKB3F1b3RlSUQYASABKAlSB3F1b3RlSUQSIAoLZGVzY3JpcHRpb24YAiABKAlSC2Rlc2NyaXB0aW9uEiAKC2V4dGVybmFsUmVmGAMgASgJUgtleHRlcm5hbFJlZhIcCglpcEFkZHJlc3MYBCABKAlSCWlwQWRkcmVzcxImCg5pZGVtcG90ZW5jeUtleRgFIAEoCVIOaWRlbXBvdGVuY3lLZXk=');
@$core.Deprecated('Use lookupIncomingPaymentRequestDescriptor instead')
const LookupIncomingPaymentRequest$json = const {
  '1': 'LookupIncomingPaymentRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `LookupIncomingPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List lookupIncomingPaymentRequestDescriptor = $convert.base64Decode('ChxMb29rdXBJbmNvbWluZ1BheW1lbnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');
@$core.Deprecated('Use createIncomingPaymentRequestDescriptor instead')
const CreateIncomingPaymentRequest$json = const {
  '1': 'CreateIncomingPaymentRequest',
  '2': const [
    const {'1': 'paymentPointer', '3': 1, '4': 1, '5': 9, '10': 'paymentPointer'},
    const {'1': 'amount', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Amount', '9': 0, '10': 'amount', '17': true},
    const {'1': 'reference', '3': 3, '4': 1, '5': 9, '10': 'reference'},
    const {'1': 'expiresAt', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'expiresAt'},
  ],
  '8': const [
    const {'1': '_amount'},
  ],
};

/// Descriptor for `CreateIncomingPaymentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createIncomingPaymentRequestDescriptor = $convert.base64Decode('ChxDcmVhdGVJbmNvbWluZ1BheW1lbnRSZXF1ZXN0EiYKDnBheW1lbnRQb2ludGVyGAEgASgJUg5wYXltZW50UG9pbnRlchIvCgZhbW91bnQYAiABKAsyEi5iYWNrZW5kLnYxLkFtb3VudEgAUgZhbW91bnSIAQESHAoJcmVmZXJlbmNlGAMgASgJUglyZWZlcmVuY2USOAoJZXhwaXJlc0F0GAQgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJZXhwaXJlc0F0QgkKB19hbW91bnQ=');
@$core.Deprecated('Use incomingPaymentDescriptor instead')
const IncomingPayment$json = const {
  '1': 'IncomingPayment',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'paymentPointer', '3': 2, '4': 1, '5': 9, '10': 'paymentPointer'},
    const {'1': 'fromPaymentPointer', '3': 10, '4': 1, '5': 9, '10': 'fromPaymentPointer'},
    const {'1': 'incomingAmount', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '9': 0, '10': 'incomingAmount', '17': true},
    const {'1': 'receivedAmount', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '9': 1, '10': 'receivedAmount', '17': true},
    const {'1': 'completed', '3': 5, '4': 1, '5': 8, '10': 'completed'},
    const {'1': 'externalRef', '3': 6, '4': 1, '5': 9, '10': 'externalRef'},
    const {'1': 'expiresAt', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'expiresAt'},
    const {'1': 'createdAt', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'createdAt'},
    const {'1': 'updatedAt', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'updatedAt'},
  ],
  '8': const [
    const {'1': '_incomingAmount'},
    const {'1': '_receivedAmount'},
  ],
};

/// Descriptor for `IncomingPayment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List incomingPaymentDescriptor = $convert.base64Decode('Cg9JbmNvbWluZ1BheW1lbnQSDgoCaWQYASABKAlSAmlkEiYKDnBheW1lbnRQb2ludGVyGAIgASgJUg5wYXltZW50UG9pbnRlchIuChJmcm9tUGF5bWVudFBvaW50ZXIYCiABKAlSEmZyb21QYXltZW50UG9pbnRlchI/Cg5pbmNvbWluZ0Ftb3VudBgDIAEoCzISLmJhY2tlbmQudjEuQW1vdW50SABSDmluY29taW5nQW1vdW50iAEBEj8KDnJlY2VpdmVkQW1vdW50GAQgASgLMhIuYmFja2VuZC52MS5BbW91bnRIAVIOcmVjZWl2ZWRBbW91bnSIAQESHAoJY29tcGxldGVkGAUgASgIUgljb21wbGV0ZWQSIAoLZXh0ZXJuYWxSZWYYBiABKAlSC2V4dGVybmFsUmVmEjgKCWV4cGlyZXNBdBgHIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWV4cGlyZXNBdBI4CgljcmVhdGVkQXQYCCABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgljcmVhdGVkQXQSOAoJdXBkYXRlZEF0GAkgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJdXBkYXRlZEF0QhEKD19pbmNvbWluZ0Ftb3VudEIRCg9fcmVjZWl2ZWRBbW91bnQ=');
@$core.Deprecated('Use lookupQuoteRequestDescriptor instead')
const LookupQuoteRequest$json = const {
  '1': 'LookupQuoteRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `LookupQuoteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List lookupQuoteRequestDescriptor = $convert.base64Decode('ChJMb29rdXBRdW90ZVJlcXVlc3QSDgoCaWQYASABKAlSAmlk');
@$core.Deprecated('Use createQuoteRequestDescriptor instead')
const CreateQuoteRequest$json = const {
  '1': 'CreateQuoteRequest',
  '2': const [
    const {'1': 'sendPaymentPointer', '3': 1, '4': 1, '5': 9, '10': 'sendPaymentPointer'},
    const {'1': 'receivePaymentPointer', '3': 2, '4': 1, '5': 9, '10': 'receivePaymentPointer'},
    const {'1': 'amount', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'amount'},
    const {'1': 'expiresAt', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'expiresAt'},
    const {'1': 'description', '3': 5, '4': 1, '5': 9, '10': 'description'},
    const {'1': 'sendLinkedAccount', '3': 6, '4': 1, '5': 9, '9': 0, '10': 'sendLinkedAccount', '17': true},
  ],
  '8': const [
    const {'1': '_sendLinkedAccount'},
  ],
};

/// Descriptor for `CreateQuoteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createQuoteRequestDescriptor = $convert.base64Decode('ChJDcmVhdGVRdW90ZVJlcXVlc3QSLgoSc2VuZFBheW1lbnRQb2ludGVyGAEgASgJUhJzZW5kUGF5bWVudFBvaW50ZXISNAoVcmVjZWl2ZVBheW1lbnRQb2ludGVyGAIgASgJUhVyZWNlaXZlUGF5bWVudFBvaW50ZXISKgoGYW1vdW50GAMgASgLMhIuYmFja2VuZC52MS5BbW91bnRSBmFtb3VudBI4CglleHBpcmVzQXQYBCABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUglleHBpcmVzQXQSIAoLZGVzY3JpcHRpb24YBSABKAlSC2Rlc2NyaXB0aW9uEjEKEXNlbmRMaW5rZWRBY2NvdW50GAYgASgJSABSEXNlbmRMaW5rZWRBY2NvdW50iAEBQhQKEl9zZW5kTGlua2VkQWNjb3VudA==');
@$core.Deprecated('Use quoteDescriptor instead')
const Quote$json = const {
  '1': 'Quote',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'paymentPointer', '3': 2, '4': 1, '5': 9, '10': 'paymentPointer'},
    const {'1': 'receiver', '3': 3, '4': 1, '5': 9, '10': 'receiver'},
    const {'1': 'sendAmount', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'sendAmount'},
    const {'1': 'receiveAmount', '3': 5, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'receiveAmount'},
    const {'1': 'expiresAt', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'expiresAt'},
    const {'1': 'createdAt', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'createdAt'},
  ],
};

/// Descriptor for `Quote`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List quoteDescriptor = $convert.base64Decode('CgVRdW90ZRIOCgJpZBgBIAEoCVICaWQSJgoOcGF5bWVudFBvaW50ZXIYAiABKAlSDnBheW1lbnRQb2ludGVyEhoKCHJlY2VpdmVyGAMgASgJUghyZWNlaXZlchIyCgpzZW5kQW1vdW50GAQgASgLMhIuYmFja2VuZC52MS5BbW91bnRSCnNlbmRBbW91bnQSOAoNcmVjZWl2ZUFtb3VudBgFIAEoCzISLmJhY2tlbmQudjEuQW1vdW50Ug1yZWNlaXZlQW1vdW50EjgKCWV4cGlyZXNBdBgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWV4cGlyZXNBdBI4CgljcmVhdGVkQXQYByABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgljcmVhdGVkQXQ=');
@$core.Deprecated('Use paymentPointerExistsRequestDescriptor instead')
const PaymentPointerExistsRequest$json = const {
  '1': 'PaymentPointerExistsRequest',
  '2': const [
    const {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `PaymentPointerExistsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paymentPointerExistsRequestDescriptor = $convert.base64Decode('ChtQYXltZW50UG9pbnRlckV4aXN0c1JlcXVlc3QSEAoDdXJsGAEgASgJUgN1cmw=');
@$core.Deprecated('Use paymentPointerExistsResponseDescriptor instead')
const PaymentPointerExistsResponse$json = const {
  '1': 'PaymentPointerExistsResponse',
  '2': const [
    const {'1': 'exists', '3': 1, '4': 1, '5': 8, '10': 'exists'},
  ],
};

/// Descriptor for `PaymentPointerExistsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paymentPointerExistsResponseDescriptor = $convert.base64Decode('ChxQYXltZW50UG9pbnRlckV4aXN0c1Jlc3BvbnNlEhYKBmV4aXN0cxgBIAEoCFIGZXhpc3Rz');
@$core.Deprecated('Use getPaymentPointerRequestDescriptor instead')
const GetPaymentPointerRequest$json = const {
  '1': 'GetPaymentPointerRequest',
  '2': const [
    const {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `GetPaymentPointerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPaymentPointerRequestDescriptor = $convert.base64Decode('ChhHZXRQYXltZW50UG9pbnRlclJlcXVlc3QSEAoDdXJsGAEgASgJUgN1cmw=');
@$core.Deprecated('Use listWalletPaymentPointersResponseDescriptor instead')
const ListWalletPaymentPointersResponse$json = const {
  '1': 'ListWalletPaymentPointersResponse',
  '2': const [
    const {'1': 'pointers', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.PaymentPointer', '10': 'pointers'},
  ],
};

/// Descriptor for `ListWalletPaymentPointersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWalletPaymentPointersResponseDescriptor = $convert.base64Decode('CiFMaXN0V2FsbGV0UGF5bWVudFBvaW50ZXJzUmVzcG9uc2USNgoIcG9pbnRlcnMYASADKAsyGi5iYWNrZW5kLnYxLlBheW1lbnRQb2ludGVyUghwb2ludGVycw==');
@$core.Deprecated('Use createPaymentPointerRequestDescriptor instead')
const CreatePaymentPointerRequest$json = const {
  '1': 'CreatePaymentPointerRequest',
  '2': const [
    const {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
    const {'1': 'asset', '3': 2, '4': 1, '5': 9, '10': 'asset'},
    const {'1': 'assetScale', '3': 3, '4': 1, '5': 5, '10': 'assetScale'},
    const {'1': 'alias', '3': 4, '4': 1, '5': 9, '10': 'alias'},
  ],
};

/// Descriptor for `CreatePaymentPointerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createPaymentPointerRequestDescriptor = $convert.base64Decode('ChtDcmVhdGVQYXltZW50UG9pbnRlclJlcXVlc3QSEAoDdXJsGAEgASgJUgN1cmwSFAoFYXNzZXQYAiABKAlSBWFzc2V0Eh4KCmFzc2V0U2NhbGUYAyABKAVSCmFzc2V0U2NhbGUSFAoFYWxpYXMYBCABKAlSBWFsaWFz');
@$core.Deprecated('Use paymentPointerDescriptor instead')
const PaymentPointer$json = const {
  '1': 'PaymentPointer',
  '2': const [
    const {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
    const {'1': 'asset', '3': 2, '4': 1, '5': 9, '10': 'asset'},
    const {'1': 'assetScale', '3': 3, '4': 1, '5': 5, '10': 'assetScale'},
    const {'1': 'alias', '3': 4, '4': 1, '5': 9, '10': 'alias'},
    const {'1': 'walletID', '3': 5, '4': 1, '5': 9, '10': 'walletID'},
    const {'1': 'formatted', '3': 6, '4': 1, '5': 9, '10': 'formatted'},
    const {'1': 'legalName', '3': 7, '4': 1, '5': 9, '10': 'legalName'},
  ],
};

/// Descriptor for `PaymentPointer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paymentPointerDescriptor = $convert.base64Decode('Cg5QYXltZW50UG9pbnRlchIQCgN1cmwYASABKAlSA3VybBIUCgVhc3NldBgCIAEoCVIFYXNzZXQSHgoKYXNzZXRTY2FsZRgDIAEoBVIKYXNzZXRTY2FsZRIUCgVhbGlhcxgEIAEoCVIFYWxpYXMSGgoId2FsbGV0SUQYBSABKAlSCHdhbGxldElEEhwKCWZvcm1hdHRlZBgGIAEoCVIJZm9ybWF0dGVkEhwKCWxlZ2FsTmFtZRgHIAEoCVIJbGVnYWxOYW1l');
@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = const {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor = $convert.base64Decode('CgVFbXB0eQ==');
@$core.Deprecated('Use transferDescriptor instead')
const Transfer$json = const {
  '1': 'Transfer',
  '2': const [
    const {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    const {'1': 'state', '3': 2, '4': 1, '5': 9, '10': 'state'},
    const {'1': 'timestamp', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
    const {'1': 'amount', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.Amount', '10': 'amount'},
    const {'1': 'foreignId', '3': 5, '4': 1, '5': 9, '10': 'foreignId'},
    const {'1': 'linkedAccountId', '3': 6, '4': 1, '5': 9, '10': 'linkedAccountId'},
  ],
};

/// Descriptor for `Transfer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transferDescriptor = $convert.base64Decode('CghUcmFuc2ZlchISCgR0eXBlGAEgASgJUgR0eXBlEhQKBXN0YXRlGAIgASgJUgVzdGF0ZRI4Cgl0aW1lc3RhbXAYAyABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl0aW1lc3RhbXASKgoGYW1vdW50GAQgASgLMhIuYmFja2VuZC52MS5BbW91bnRSBmFtb3VudBIcCglmb3JlaWduSWQYBSABKAlSCWZvcmVpZ25JZBIoCg9saW5rZWRBY2NvdW50SWQYBiABKAlSD2xpbmtlZEFjY291bnRJZA==');
@$core.Deprecated('Use listStatementsResponseDescriptor instead')
const ListStatementsResponse$json = const {
  '1': 'ListStatementsResponse',
  '2': const [
    const {'1': 'periods', '3': 1, '4': 3, '5': 9, '10': 'periods'},
    const {'1': 'page', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.PaginationResponse', '10': 'page'},
  ],
};

/// Descriptor for `ListStatementsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listStatementsResponseDescriptor = $convert.base64Decode('ChZMaXN0U3RhdGVtZW50c1Jlc3BvbnNlEhgKB3BlcmlvZHMYASADKAlSB3BlcmlvZHMSMgoEcGFnZRgCIAEoCzIeLmJhY2tlbmQudjEuUGFnaW5hdGlvblJlc3BvbnNlUgRwYWdl');
@$core.Deprecated('Use getStatementPDFRequestDescriptor instead')
const GetStatementPDFRequest$json = const {
  '1': 'GetStatementPDFRequest',
  '2': const [
    const {'1': 'period', '3': 1, '4': 1, '5': 9, '10': 'period'},
  ],
};

/// Descriptor for `GetStatementPDFRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getStatementPDFRequestDescriptor = $convert.base64Decode('ChZHZXRTdGF0ZW1lbnRQREZSZXF1ZXN0EhYKBnBlcmlvZBgBIAEoCVIGcGVyaW9k');
@$core.Deprecated('Use statementPDFDescriptor instead')
const StatementPDF$json = const {
  '1': 'StatementPDF',
  '2': const [
    const {'1': 'chunks', '3': 1, '4': 1, '5': 12, '10': 'chunks'},
  ],
};

/// Descriptor for `StatementPDF`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List statementPDFDescriptor = $convert.base64Decode('CgxTdGF0ZW1lbnRQREYSFgoGY2h1bmtzGAEgASgMUgZjaHVua3M=');
@$core.Deprecated('Use createSupportTicketRequestDescriptor instead')
const CreateSupportTicketRequest$json = const {
  '1': 'CreateSupportTicketRequest',
  '2': const [
    const {'1': 'description', '3': 1, '4': 1, '5': 9, '10': 'description'},
    const {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    const {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    const {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
  ],
};

/// Descriptor for `CreateSupportTicketRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSupportTicketRequestDescriptor = $convert.base64Decode('ChpDcmVhdGVTdXBwb3J0VGlja2V0UmVxdWVzdBIgCgtkZXNjcmlwdGlvbhgBIAEoCVILZGVzY3JpcHRpb24SHAoJZmlyc3ROYW1lGAIgASgJUglmaXJzdE5hbWUSGgoIbGFzdE5hbWUYAyABKAlSCGxhc3ROYW1lEhQKBWVtYWlsGAQgASgJUgVlbWFpbA==');
@$core.Deprecated('Use updateIndividualKYCRequestDescriptor instead')
const UpdateIndividualKYCRequest$json = const {
  '1': 'UpdateIndividualKYCRequest',
  '2': const [
    const {'1': 'firstName', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'firstName', '17': true},
    const {'1': 'lastName', '3': 2, '4': 1, '5': 9, '9': 1, '10': 'lastName', '17': true},
    const {'1': 'countryCode', '3': 3, '4': 1, '5': 9, '9': 2, '10': 'countryCode', '17': true},
    const {'1': 'gender', '3': 4, '4': 1, '5': 5, '9': 3, '10': 'gender', '17': true},
    const {'1': 'dateOfBirth', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '9': 4, '10': 'dateOfBirth', '17': true},
    const {'1': 'address', '3': 6, '4': 1, '5': 11, '6': '.backend.v1.Address', '9': 5, '10': 'address', '17': true},
    const {'1': 'ipAddress', '3': 7, '4': 1, '5': 9, '10': 'ipAddress'},
  ],
  '8': const [
    const {'1': '_firstName'},
    const {'1': '_lastName'},
    const {'1': '_countryCode'},
    const {'1': '_gender'},
    const {'1': '_dateOfBirth'},
    const {'1': '_address'},
  ],
};

/// Descriptor for `UpdateIndividualKYCRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateIndividualKYCRequestDescriptor = $convert.base64Decode('ChpVcGRhdGVJbmRpdmlkdWFsS1lDUmVxdWVzdBIhCglmaXJzdE5hbWUYASABKAlIAFIJZmlyc3ROYW1liAEBEh8KCGxhc3ROYW1lGAIgASgJSAFSCGxhc3ROYW1liAEBEiUKC2NvdW50cnlDb2RlGAMgASgJSAJSC2NvdW50cnlDb2RliAEBEhsKBmdlbmRlchgEIAEoBUgDUgZnZW5kZXKIAQESQQoLZGF0ZU9mQmlydGgYBSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wSARSC2RhdGVPZkJpcnRoiAEBEjIKB2FkZHJlc3MYBiABKAsyEy5iYWNrZW5kLnYxLkFkZHJlc3NIBVIHYWRkcmVzc4gBARIcCglpcEFkZHJlc3MYByABKAlSCWlwQWRkcmVzc0IMCgpfZmlyc3ROYW1lQgsKCV9sYXN0TmFtZUIOCgxfY291bnRyeUNvZGVCCQoHX2dlbmRlckIOCgxfZGF0ZU9mQmlydGhCCgoIX2FkZHJlc3M=');
@$core.Deprecated('Use addressDescriptor instead')
const Address$json = const {
  '1': 'Address',
  '2': const [
    const {'1': 'line1', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'line1', '17': true},
    const {'1': 'line2', '3': 2, '4': 1, '5': 9, '9': 1, '10': 'line2', '17': true},
    const {'1': 'building', '3': 3, '4': 1, '5': 9, '9': 2, '10': 'building', '17': true},
    const {'1': 'apartment', '3': 4, '4': 1, '5': 9, '9': 3, '10': 'apartment', '17': true},
    const {'1': 'city', '3': 5, '4': 1, '5': 9, '9': 4, '10': 'city', '17': true},
    const {'1': 'state', '3': 6, '4': 1, '5': 9, '9': 5, '10': 'state', '17': true},
    const {'1': 'zipCode', '3': 7, '4': 1, '5': 9, '9': 6, '10': 'zipCode', '17': true},
    const {'1': 'countryCode', '3': 8, '4': 1, '5': 9, '9': 7, '10': 'countryCode', '17': true},
    const {'1': 'placeID', '3': 9, '4': 1, '5': 9, '9': 8, '10': 'placeID', '17': true},
    const {'1': 'formattedAddress', '3': 10, '4': 1, '5': 9, '9': 9, '10': 'formattedAddress', '17': true},
  ],
  '8': const [
    const {'1': '_line1'},
    const {'1': '_line2'},
    const {'1': '_building'},
    const {'1': '_apartment'},
    const {'1': '_city'},
    const {'1': '_state'},
    const {'1': '_zipCode'},
    const {'1': '_countryCode'},
    const {'1': '_placeID'},
    const {'1': '_formattedAddress'},
  ],
};

/// Descriptor for `Address`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addressDescriptor = $convert.base64Decode('CgdBZGRyZXNzEhkKBWxpbmUxGAEgASgJSABSBWxpbmUxiAEBEhkKBWxpbmUyGAIgASgJSAFSBWxpbmUyiAEBEh8KCGJ1aWxkaW5nGAMgASgJSAJSCGJ1aWxkaW5niAEBEiEKCWFwYXJ0bWVudBgEIAEoCUgDUglhcGFydG1lbnSIAQESFwoEY2l0eRgFIAEoCUgEUgRjaXR5iAEBEhkKBXN0YXRlGAYgASgJSAVSBXN0YXRliAEBEh0KB3ppcENvZGUYByABKAlIBlIHemlwQ29kZYgBARIlCgtjb3VudHJ5Q29kZRgIIAEoCUgHUgtjb3VudHJ5Q29kZYgBARIdCgdwbGFjZUlEGAkgASgJSAhSB3BsYWNlSUSIAQESLwoQZm9ybWF0dGVkQWRkcmVzcxgKIAEoCUgJUhBmb3JtYXR0ZWRBZGRyZXNziAEBQggKBl9saW5lMUIICgZfbGluZTJCCwoJX2J1aWxkaW5nQgwKCl9hcGFydG1lbnRCBwoFX2NpdHlCCAoGX3N0YXRlQgoKCF96aXBDb2RlQg4KDF9jb3VudHJ5Q29kZUIKCghfcGxhY2VJREITChFfZm9ybWF0dGVkQWRkcmVzcw==');
@$core.Deprecated('Use isUSPSAddressResponseDescriptor instead')
const IsUSPSAddressResponse$json = const {
  '1': 'IsUSPSAddressResponse',
  '2': const [
    const {'1': 'valid', '3': 1, '4': 1, '5': 8, '10': 'valid'},
  ],
};

/// Descriptor for `IsUSPSAddressResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isUSPSAddressResponseDescriptor = $convert.base64Decode('ChVJc1VTUFNBZGRyZXNzUmVzcG9uc2USFAoFdmFsaWQYASABKAhSBXZhbGlk');
@$core.Deprecated('Use getBankAccountWidgetRequestDescriptor instead')
const GetBankAccountWidgetRequest$json = const {
  '1': 'GetBankAccountWidgetRequest',
};

/// Descriptor for `GetBankAccountWidgetRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBankAccountWidgetRequestDescriptor = $convert.base64Decode('ChtHZXRCYW5rQWNjb3VudFdpZGdldFJlcXVlc3Q=');
@$core.Deprecated('Use getBankAccountWidgetResponseDescriptor instead')
const GetBankAccountWidgetResponse$json = const {
  '1': 'GetBankAccountWidgetResponse',
  '2': const [
    const {'1': 'url', '3': 1, '4': 1, '5': 9, '10': 'url'},
  ],
};

/// Descriptor for `GetBankAccountWidgetResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getBankAccountWidgetResponseDescriptor = $convert.base64Decode('ChxHZXRCYW5rQWNjb3VudFdpZGdldFJlc3BvbnNlEhAKA3VybBgBIAEoCVIDdXJs');
@$core.Deprecated('Use addBankAccountRequestDescriptor instead')
const AddBankAccountRequest$json = const {
  '1': 'AddBankAccountRequest',
  '2': const [
    const {'1': 'userGuid', '3': 1, '4': 1, '5': 9, '10': 'userGuid'},
    const {'1': 'memberGuid', '3': 2, '4': 1, '5': 9, '10': 'memberGuid'},
  ],
};

/// Descriptor for `AddBankAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addBankAccountRequestDescriptor = $convert.base64Decode('ChVBZGRCYW5rQWNjb3VudFJlcXVlc3QSGgoIdXNlckd1aWQYASABKAlSCHVzZXJHdWlkEh4KCm1lbWJlckd1aWQYAiABKAlSCm1lbWJlckd1aWQ=');
@$core.Deprecated('Use addBankAccountResponseDescriptor instead')
const AddBankAccountResponse$json = const {
  '1': 'AddBankAccountResponse',
  '2': const [
    const {'1': 'fundingsourceId', '3': 1, '4': 1, '5': 9, '10': 'fundingsourceId'},
  ],
};

/// Descriptor for `AddBankAccountResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addBankAccountResponseDescriptor = $convert.base64Decode('ChZBZGRCYW5rQWNjb3VudFJlc3BvbnNlEigKD2Z1bmRpbmdzb3VyY2VJZBgBIAEoCVIPZnVuZGluZ3NvdXJjZUlk');
@$core.Deprecated('Use linkedAccountDescriptor instead')
const LinkedAccount$json = const {
  '1': 'LinkedAccount',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    const {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    const {'1': 'mask', '3': 4, '4': 1, '5': 9, '10': 'mask'},
  ],
};

/// Descriptor for `LinkedAccount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkedAccountDescriptor = $convert.base64Decode('Cg1MaW5rZWRBY2NvdW50Eg4KAmlkGAEgASgJUgJpZBISCgR0eXBlGAIgASgJUgR0eXBlEhIKBG5hbWUYAyABKAlSBG5hbWUSEgoEbWFzaxgEIAEoCVIEbWFzaw==');
@$core.Deprecated('Use getSignupRequestDescriptor instead')
const GetSignupRequest$json = const {
  '1': 'GetSignupRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSignupRequestDescriptor = $convert.base64Decode('ChBHZXRTaWdudXBSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');
@$core.Deprecated('Use setSignupUserDataRequestDescriptor instead')
const SetSignupUserDataRequest$json = const {
  '1': 'SetSignupUserDataRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '9': 0, '10': 'id', '17': true},
    const {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    const {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    const {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
    const {'1': 'countryCode', '3': 5, '4': 1, '5': 9, '10': 'countryCode'},
  ],
  '8': const [
    const {'1': '_id'},
  ],
};

/// Descriptor for `SetSignupUserDataRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupUserDataRequestDescriptor = $convert.base64Decode('ChhTZXRTaWdudXBVc2VyRGF0YVJlcXVlc3QSEwoCaWQYASABKAlIAFICaWSIAQESHAoJZmlyc3ROYW1lGAIgASgJUglmaXJzdE5hbWUSGgoIbGFzdE5hbWUYAyABKAlSCGxhc3ROYW1lEhQKBWVtYWlsGAQgASgJUgVlbWFpbBIgCgtjb3VudHJ5Q29kZRgFIAEoCVILY291bnRyeUNvZGVCBQoDX2lk');
@$core.Deprecated('Use setSignupUserDataResponseDescriptor instead')
const SetSignupUserDataResponse$json = const {
  '1': 'SetSignupUserDataResponse',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `SetSignupUserDataResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupUserDataResponseDescriptor = $convert.base64Decode('ChlTZXRTaWdudXBVc2VyRGF0YVJlc3BvbnNlEg4KAmlkGAEgASgJUgJpZA==');
@$core.Deprecated('Use setSignupMobileNumberRequestDescriptor instead')
const SetSignupMobileNumberRequest$json = const {
  '1': 'SetSignupMobileNumberRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'mobile', '3': 2, '4': 1, '5': 9, '10': 'mobile'},
    const {'1': 'otp', '3': 3, '4': 1, '5': 9, '10': 'otp'},
  ],
};

/// Descriptor for `SetSignupMobileNumberRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupMobileNumberRequestDescriptor = $convert.base64Decode('ChxTZXRTaWdudXBNb2JpbGVOdW1iZXJSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIWCgZtb2JpbGUYAiABKAlSBm1vYmlsZRIQCgNvdHAYAyABKAlSA290cA==');
@$core.Deprecated('Use signupDescriptor instead')
const Signup$json = const {
  '1': 'Signup',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    const {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    const {'1': 'email', '3': 4, '4': 1, '5': 9, '10': 'email'},
    const {'1': 'countryCode', '3': 5, '4': 1, '5': 9, '10': 'countryCode'},
    const {'1': 'mobileNumber', '3': 6, '4': 1, '5': 9, '10': 'mobileNumber'},
    const {'1': 'userId', '3': 7, '4': 1, '5': 9, '10': 'userId'},
    const {'1': 'completed', '3': 8, '4': 1, '5': 8, '10': 'completed'},
  ],
};

/// Descriptor for `Signup`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signupDescriptor = $convert.base64Decode('CgZTaWdudXASDgoCaWQYASABKAlSAmlkEhwKCWZpcnN0TmFtZRgCIAEoCVIJZmlyc3ROYW1lEhoKCGxhc3ROYW1lGAMgASgJUghsYXN0TmFtZRIUCgVlbWFpbBgEIAEoCVIFZW1haWwSIAoLY291bnRyeUNvZGUYBSABKAlSC2NvdW50cnlDb2RlEiIKDG1vYmlsZU51bWJlchgGIAEoCVIMbW9iaWxlTnVtYmVyEhYKBnVzZXJJZBgHIAEoCVIGdXNlcklkEhwKCWNvbXBsZXRlZBgIIAEoCFIJY29tcGxldGVk');
@$core.Deprecated('Use completeSignupRequestDescriptor instead')
const CompleteSignupRequest$json = const {
  '1': 'CompleteSignupRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'userId', '3': 2, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `CompleteSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeSignupRequestDescriptor = $convert.base64Decode('ChVDb21wbGV0ZVNpZ251cFJlcXVlc3QSDgoCaWQYASABKAlSAmlkEhYKBnVzZXJJZBgCIAEoCVIGdXNlcklk');
@$core.Deprecated('Use createUserDefaultWalletRequestDescriptor instead')
const CreateUserDefaultWalletRequest$json = const {
  '1': 'CreateUserDefaultWalletRequest',
  '2': const [
    const {'1': 'userID', '3': 1, '4': 1, '5': 9, '10': 'userID'},
  ],
};

/// Descriptor for `CreateUserDefaultWalletRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createUserDefaultWalletRequestDescriptor = $convert.base64Decode('Ch5DcmVhdGVVc2VyRGVmYXVsdFdhbGxldFJlcXVlc3QSFgoGdXNlcklEGAEgASgJUgZ1c2VySUQ=');
@$core.Deprecated('Use sendPhoneVerificationRequestDescriptor instead')
const SendPhoneVerificationRequest$json = const {
  '1': 'SendPhoneVerificationRequest',
  '2': const [
    const {'1': 'to', '3': 1, '4': 1, '5': 9, '10': 'to'},
  ],
};

/// Descriptor for `SendPhoneVerificationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendPhoneVerificationRequestDescriptor = $convert.base64Decode('ChxTZW5kUGhvbmVWZXJpZmljYXRpb25SZXF1ZXN0Eg4KAnRvGAEgASgJUgJ0bw==');
@$core.Deprecated('Use getAgreementRequestDescriptor instead')
const GetAgreementRequest$json = const {
  '1': 'GetAgreementRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetAgreementRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgreementRequestDescriptor = $convert.base64Decode('ChNHZXRBZ3JlZW1lbnRSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');
@$core.Deprecated('Use agreementDescriptor instead')
const Agreement$json = const {
  '1': 'Agreement',
  '2': const [
    const {'1': 'content', '3': 1, '4': 1, '5': 9, '10': 'content'},
  ],
};

/// Descriptor for `Agreement`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agreementDescriptor = $convert.base64Decode('CglBZ3JlZW1lbnQSGAoHY29udGVudBgBIAEoCVIHY29udGVudA==');
@$core.Deprecated('Use signAgreementsRequestDescriptor instead')
const SignAgreementsRequest$json = const {
  '1': 'SignAgreementsRequest',
  '2': const [
    const {'1': 'agreementIds', '3': 1, '4': 3, '5': 9, '10': 'agreementIds'},
    const {'1': 'userId', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    const {'1': 'ipAddress', '3': 3, '4': 1, '5': 9, '10': 'ipAddress'},
  ],
};

/// Descriptor for `SignAgreementsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signAgreementsRequestDescriptor = $convert.base64Decode('ChVTaWduQWdyZWVtZW50c1JlcXVlc3QSIgoMYWdyZWVtZW50SWRzGAEgAygJUgxhZ3JlZW1lbnRJZHMSFgoGdXNlcklkGAIgASgJUgZ1c2VySWQSHAoJaXBBZGRyZXNzGAMgASgJUglpcEFkZHJlc3M=');
@$core.Deprecated('Use signAgreementsResponseDescriptor instead')
const SignAgreementsResponse$json = const {
  '1': 'SignAgreementsResponse',
  '2': const [
    const {'1': 'signed', '3': 1, '4': 1, '5': 8, '10': 'signed'},
  ],
};

/// Descriptor for `SignAgreementsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signAgreementsResponseDescriptor = $convert.base64Decode('ChZTaWduQWdyZWVtZW50c1Jlc3BvbnNlEhYKBnNpZ25lZBgBIAEoCFIGc2lnbmVk');
@$core.Deprecated('Use joinWaitlistRequestDescriptor instead')
const JoinWaitlistRequest$json = const {
  '1': 'JoinWaitlistRequest',
  '2': const [
    const {'1': 'email', '3': 1, '4': 1, '5': 9, '10': 'email'},
    const {'1': 'country_code', '3': 2, '4': 1, '5': 9, '10': 'countryCode'},
    const {'1': 'full_name', '3': 3, '4': 1, '5': 9, '10': 'fullName'},
    const {'1': 'beta_opt_in', '3': 4, '4': 1, '5': 8, '10': 'betaOptIn'},
    const {'1': 'mug_id', '3': 5, '4': 1, '5': 9, '9': 0, '10': 'mugId', '17': true},
  ],
  '8': const [
    const {'1': '_mug_id'},
  ],
};

/// Descriptor for `JoinWaitlistRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitlistRequestDescriptor = $convert.base64Decode('ChNKb2luV2FpdGxpc3RSZXF1ZXN0EhQKBWVtYWlsGAEgASgJUgVlbWFpbBIhCgxjb3VudHJ5X2NvZGUYAiABKAlSC2NvdW50cnlDb2RlEhsKCWZ1bGxfbmFtZRgDIAEoCVIIZnVsbE5hbWUSHgoLYmV0YV9vcHRfaW4YBCABKAhSCWJldGFPcHRJbhIaCgZtdWdfaWQYBSABKAlIAFIFbXVnSWSIAQFCCQoHX211Z19pZA==');
@$core.Deprecated('Use joinWaitlistResponseDescriptor instead')
const JoinWaitlistResponse$json = const {
  '1': 'JoinWaitlistResponse',
};

/// Descriptor for `JoinWaitlistResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitlistResponseDescriptor = $convert.base64Decode('ChRKb2luV2FpdGxpc3RSZXNwb25zZQ==');
@$core.Deprecated('Use isMugAvailableRequestDescriptor instead')
const IsMugAvailableRequest$json = const {
  '1': 'IsMugAvailableRequest',
  '2': const [
    const {'1': 'mug_id', '3': 1, '4': 1, '5': 9, '10': 'mugId'},
  ],
};

/// Descriptor for `IsMugAvailableRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isMugAvailableRequestDescriptor = $convert.base64Decode('ChVJc011Z0F2YWlsYWJsZVJlcXVlc3QSFQoGbXVnX2lkGAEgASgJUgVtdWdJZA==');
@$core.Deprecated('Use isMugAvailableResponseDescriptor instead')
const IsMugAvailableResponse$json = const {
  '1': 'IsMugAvailableResponse',
  '2': const [
    const {'1': 'available', '3': 1, '4': 1, '5': 8, '10': 'available'},
  ],
};

/// Descriptor for `IsMugAvailableResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List isMugAvailableResponseDescriptor = $convert.base64Decode('ChZJc011Z0F2YWlsYWJsZVJlc3BvbnNlEhwKCWF2YWlsYWJsZRgBIAEoCFIJYXZhaWxhYmxl');
@$core.Deprecated('Use getLinkedAccountsResponseDescriptor instead')
const GetLinkedAccountsResponse$json = const {
  '1': 'GetLinkedAccountsResponse',
  '2': const [
    const {'1': 'linkedAccounts', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.LinkedAccount', '10': 'linkedAccounts'},
  ],
};

/// Descriptor for `GetLinkedAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinkedAccountsResponseDescriptor = $convert.base64Decode('ChlHZXRMaW5rZWRBY2NvdW50c1Jlc3BvbnNlEkEKDmxpbmtlZEFjY291bnRzGAEgAygLMhkuYmFja2VuZC52MS5MaW5rZWRBY2NvdW50Ug5saW5rZWRBY2NvdW50cw==');
@$core.Deprecated('Use getLinkedAccountRequestDescriptor instead')
const GetLinkedAccountRequest$json = const {
  '1': 'GetLinkedAccountRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetLinkedAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinkedAccountRequestDescriptor = $convert.base64Decode('ChdHZXRMaW5rZWRBY2NvdW50UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');
@$core.Deprecated('Use deleteLinkedAccountRequestDescriptor instead')
const DeleteLinkedAccountRequest$json = const {
  '1': 'DeleteLinkedAccountRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteLinkedAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteLinkedAccountRequestDescriptor = $convert.base64Decode('ChpEZWxldGVMaW5rZWRBY2NvdW50UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');
@$core.Deprecated('Use countryDescriptor instead')
const Country$json = const {
  '1': 'Country',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `Country`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List countryDescriptor = $convert.base64Decode('CgdDb3VudHJ5Eg4KAmlkGAEgASgJUgJpZBISCgRuYW1lGAIgASgJUgRuYW1l');
@$core.Deprecated('Use getCountriesResponseDescriptor instead')
const GetCountriesResponse$json = const {
  '1': 'GetCountriesResponse',
  '2': const [
    const {'1': 'countries', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Country', '10': 'countries'},
  ],
};

/// Descriptor for `GetCountriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCountriesResponseDescriptor = $convert.base64Decode('ChRHZXRDb3VudHJpZXNSZXNwb25zZRIxCgljb3VudHJpZXMYASADKAsyEy5iYWNrZW5kLnYxLkNvdW50cnlSCWNvdW50cmllcw==');
@$core.Deprecated('Use machnetWidgetTokenDescriptor instead')
const MachnetWidgetToken$json = const {
  '1': 'MachnetWidgetToken',
  '2': const [
    const {'1': 'value', '3': 1, '4': 1, '5': 9, '10': 'value'},
    const {'1': 'expiresInMinutes', '3': 2, '4': 1, '5': 3, '10': 'expiresInMinutes'},
    const {'1': 'userId', '3': 3, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `MachnetWidgetToken`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List machnetWidgetTokenDescriptor = $convert.base64Decode('ChJNYWNobmV0V2lkZ2V0VG9rZW4SFAoFdmFsdWUYASABKAlSBXZhbHVlEioKEGV4cGlyZXNJbk1pbnV0ZXMYAiABKANSEGV4cGlyZXNJbk1pbnV0ZXMSFgoGdXNlcklkGAMgASgJUgZ1c2VySWQ=');
@$core.Deprecated('Use branchDescriptor instead')
const Branch$json = const {
  '1': 'Branch',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    const {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `Branch`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List branchDescriptor = $convert.base64Decode('CgZCcmFuY2gSDgoCaWQYASABKA1SAmlkEhIKBG5hbWUYAiABKAlSBG5hbWU=');
@$core.Deprecated('Use bankDescriptor instead')
const Bank$json = const {
  '1': 'Bank',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    const {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    const {'1': 'branches', '3': 3, '4': 3, '5': 11, '6': '.backend.v1.Branch', '10': 'branches'},
  ],
};

/// Descriptor for `Bank`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List bankDescriptor = $convert.base64Decode('CgRCYW5rEg4KAmlkGAEgASgNUgJpZBISCgRuYW1lGAIgASgJUgRuYW1lEi4KCGJyYW5jaGVzGAMgAygLMhIuYmFja2VuZC52MS5CcmFuY2hSCGJyYW5jaGVz');
@$core.Deprecated('Use listBanksResponseDescriptor instead')
const ListBanksResponse$json = const {
  '1': 'ListBanksResponse',
  '2': const [
    const {'1': 'banks', '3': 1, '4': 3, '5': 11, '6': '.backend.v1.Bank', '10': 'banks'},
  ],
};

/// Descriptor for `ListBanksResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listBanksResponseDescriptor = $convert.base64Decode('ChFMaXN0QmFua3NSZXNwb25zZRImCgViYW5rcxgBIAMoCzIQLmJhY2tlbmQudjEuQmFua1IFYmFua3M=');
@$core.Deprecated('Use createReceiveBankAccountRequestDescriptor instead')
const CreateReceiveBankAccountRequest$json = const {
  '1': 'CreateReceiveBankAccountRequest',
  '2': const [
    const {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    const {'1': 'bank_id', '3': 2, '4': 1, '5': 13, '10': 'bankId'},
    const {'1': 'branch_id', '3': 3, '4': 1, '5': 13, '10': 'branchId'},
    const {'1': 'account_type', '3': 4, '4': 1, '5': 9, '10': 'accountType'},
    const {'1': 'account_number', '3': 5, '4': 1, '5': 9, '10': 'accountNumber'},
    const {'1': 'otp', '3': 6, '4': 1, '5': 9, '10': 'otp'},
  ],
};

/// Descriptor for `CreateReceiveBankAccountRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createReceiveBankAccountRequestDescriptor = $convert.base64Decode('Ch9DcmVhdGVSZWNlaXZlQmFua0FjY291bnRSZXF1ZXN0EhIKBG5hbWUYASABKAlSBG5hbWUSFwoHYmFua19pZBgCIAEoDVIGYmFua0lkEhsKCWJyYW5jaF9pZBgDIAEoDVIIYnJhbmNoSWQSIQoMYWNjb3VudF90eXBlGAQgASgJUgthY2NvdW50VHlwZRIlCg5hY2NvdW50X251bWJlchgFIAEoCVINYWNjb3VudE51bWJlchIQCgNvdHAYBiABKAlSA290cA==');
@$core.Deprecated('Use canSignupRequestDescriptor instead')
const CanSignupRequest$json = const {
  '1': 'CanSignupRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `CanSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List canSignupRequestDescriptor = $convert.base64Decode('ChBDYW5TaWdudXBSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');
@$core.Deprecated('Use canSignupResponseDescriptor instead')
const CanSignupResponse$json = const {
  '1': 'CanSignupResponse',
  '2': const [
    const {'1': 'canSignup', '3': 1, '4': 1, '5': 8, '10': 'canSignup'},
  ],
};

/// Descriptor for `CanSignupResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List canSignupResponseDescriptor = $convert.base64Decode('ChFDYW5TaWdudXBSZXNwb25zZRIcCgljYW5TaWdudXAYASABKAhSCWNhblNpZ251cA==');
@$core.Deprecated('Use setSignupCompleteRequestDescriptor instead')
const SetSignupCompleteRequest$json = const {
  '1': 'SetSignupCompleteRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    const {'1': 'userId', '3': 2, '4': 1, '5': 9, '10': 'userId'},
  ],
};

/// Descriptor for `SetSignupCompleteRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setSignupCompleteRequestDescriptor = $convert.base64Decode('ChhTZXRTaWdudXBDb21wbGV0ZVJlcXVlc3QSDgoCaWQYASABKAlSAmlkEhYKBnVzZXJJZBgCIAEoCVIGdXNlcklk');
@$core.Deprecated('Use hasSendUserResponseDescriptor instead')
const HasSendUserResponse$json = const {
  '1': 'HasSendUserResponse',
  '2': const [
    const {'1': 'hasSendUser', '3': 1, '4': 1, '5': 8, '10': 'hasSendUser'},
  ],
};

/// Descriptor for `HasSendUserResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List hasSendUserResponseDescriptor = $convert.base64Decode('ChNIYXNTZW5kVXNlclJlc3BvbnNlEiAKC2hhc1NlbmRVc2VyGAEgASgIUgtoYXNTZW5kVXNlcg==');
@$core.Deprecated('Use kYCStatusResponseDescriptor instead')
const KYCStatusResponse$json = const {
  '1': 'KYCStatusResponse',
  '2': const [
    const {'1': 'hasSendUser', '3': 1, '4': 1, '5': 8, '10': 'hasSendUser'},
    const {'1': 'kycStatus', '3': 2, '4': 1, '5': 5, '10': 'kycStatus'},
    const {'1': 'failedFields', '3': 3, '4': 3, '5': 9, '10': 'failedFields'},
  ],
};

/// Descriptor for `KYCStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List kYCStatusResponseDescriptor = $convert.base64Decode('ChFLWUNTdGF0dXNSZXNwb25zZRIgCgtoYXNTZW5kVXNlchgBIAEoCFILaGFzU2VuZFVzZXISHAoJa3ljU3RhdHVzGAIgASgFUglreWNTdGF0dXMSIgoMZmFpbGVkRmllbGRzGAMgAygJUgxmYWlsZWRGaWVsZHM=');
@$core.Deprecated('Use createWalletRequestDescriptor instead')
const CreateWalletRequest$json = const {
  '1': 'CreateWalletRequest',
  '2': const [
    const {'1': 'nickname', '3': 1, '4': 1, '5': 9, '10': 'nickname'},
  ],
};

/// Descriptor for `CreateWalletRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWalletRequestDescriptor = $convert.base64Decode('ChNDcmVhdGVXYWxsZXRSZXF1ZXN0EhoKCG5pY2tuYW1lGAEgASgJUghuaWNrbmFtZQ==');
@$core.Deprecated('Use walletBalanceDescriptor instead')
const WalletBalance$json = const {
  '1': 'WalletBalance',
  '2': const [
    const {'1': 'balance', '3': 1, '4': 1, '5': 4, '10': 'balance'},
    const {'1': 'available', '3': 2, '4': 1, '5': 4, '10': 'available'},
  ],
};

/// Descriptor for `WalletBalance`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletBalanceDescriptor = $convert.base64Decode('Cg1XYWxsZXRCYWxhbmNlEhgKB2JhbGFuY2UYASABKARSB2JhbGFuY2USHAoJYXZhaWxhYmxlGAIgASgEUglhdmFpbGFibGU=');
@$core.Deprecated('Use withdrawFromMachnetWalletRequestDescriptor instead')
const WithdrawFromMachnetWalletRequest$json = const {
  '1': 'WithdrawFromMachnetWalletRequest',
  '2': const [
    const {'1': 'toLinkedAccountId', '3': 1, '4': 1, '5': 9, '10': 'toLinkedAccountId'},
    const {'1': 'amount', '3': 2, '4': 1, '5': 4, '10': 'amount'},
    const {'1': 'ipAddress', '3': 3, '4': 1, '5': 9, '10': 'ipAddress'},
    const {'1': 'idempotencyKey', '3': 4, '4': 1, '5': 9, '10': 'idempotencyKey'},
  ],
};

/// Descriptor for `WithdrawFromMachnetWalletRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List withdrawFromMachnetWalletRequestDescriptor = $convert.base64Decode('CiBXaXRoZHJhd0Zyb21NYWNobmV0V2FsbGV0UmVxdWVzdBIsChF0b0xpbmtlZEFjY291bnRJZBgBIAEoCVIRdG9MaW5rZWRBY2NvdW50SWQSFgoGYW1vdW50GAIgASgEUgZhbW91bnQSHAoJaXBBZGRyZXNzGAMgASgJUglpcEFkZHJlc3MSJgoOaWRlbXBvdGVuY3lLZXkYBCABKAlSDmlkZW1wb3RlbmN5S2V5');
@$core.Deprecated('Use checkMachnetTXLimitRequestDescriptor instead')
const CheckMachnetTXLimitRequest$json = const {
  '1': 'CheckMachnetTXLimitRequest',
  '2': const [
    const {'1': 'amount', '3': 1, '4': 1, '5': 4, '10': 'amount'},
    const {'1': 'currency', '3': 2, '4': 1, '5': 9, '10': 'currency'},
  ],
};

/// Descriptor for `CheckMachnetTXLimitRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List checkMachnetTXLimitRequestDescriptor = $convert.base64Decode('ChpDaGVja01hY2huZXRUWExpbWl0UmVxdWVzdBIWCgZhbW91bnQYASABKARSBmFtb3VudBIaCghjdXJyZW5jeRgCIAEoCVIIY3VycmVuY3k=');
@$core.Deprecated('Use checkMachnetTXLimitResponseDescriptor instead')
const CheckMachnetTXLimitResponse$json = const {
  '1': 'CheckMachnetTXLimitResponse',
  '2': const [
    const {'1': 'exceedsLimits', '3': 1, '4': 1, '5': 8, '10': 'exceedsLimits'},
    const {'1': 'limitType', '3': 2, '4': 1, '5': 9, '10': 'limitType'},
  ],
};

/// Descriptor for `CheckMachnetTXLimitResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List checkMachnetTXLimitResponseDescriptor = $convert.base64Decode('ChtDaGVja01hY2huZXRUWExpbWl0UmVzcG9uc2USJAoNZXhjZWVkc0xpbWl0cxgBIAEoCFINZXhjZWVkc0xpbWl0cxIcCglsaW1pdFR5cGUYAiABKAlSCWxpbWl0VHlwZQ==');
@$core.Deprecated('Use startMachnetWalletTopupRequestDescriptor instead')
const StartMachnetWalletTopupRequest$json = const {
  '1': 'StartMachnetWalletTopupRequest',
  '2': const [
    const {'1': 'fromLinkedAccountId', '3': 1, '4': 1, '5': 9, '10': 'fromLinkedAccountId'},
    const {'1': 'amount', '3': 2, '4': 1, '5': 4, '10': 'amount'},
    const {'1': 'ipAddress', '3': 3, '4': 1, '5': 9, '10': 'ipAddress'},
    const {'1': 'currency', '3': 4, '4': 1, '5': 9, '10': 'currency'},
    const {'1': 'idempotencyKey', '3': 5, '4': 1, '5': 9, '10': 'idempotencyKey'},
  ],
};

/// Descriptor for `StartMachnetWalletTopupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startMachnetWalletTopupRequestDescriptor = $convert.base64Decode('Ch5TdGFydE1hY2huZXRXYWxsZXRUb3B1cFJlcXVlc3QSMAoTZnJvbUxpbmtlZEFjY291bnRJZBgBIAEoCVITZnJvbUxpbmtlZEFjY291bnRJZBIWCgZhbW91bnQYAiABKARSBmFtb3VudBIcCglpcEFkZHJlc3MYAyABKAlSCWlwQWRkcmVzcxIaCghjdXJyZW5jeRgEIAEoCVIIY3VycmVuY3kSJgoOaWRlbXBvdGVuY3lLZXkYBSABKAlSDmlkZW1wb3RlbmN5S2V5');
@$core.Deprecated('Use lookupTransactionRequestDescriptor instead')
const LookupTransactionRequest$json = const {
  '1': 'LookupTransactionRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `LookupTransactionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List lookupTransactionRequestDescriptor = $convert.base64Decode('ChhMb29rdXBUcmFuc2FjdGlvblJlcXVlc3QSDgoCaWQYASABKAlSAmlk');
@$core.Deprecated('Use getCurrentWalletResponseDescriptor instead')
const GetCurrentWalletResponse$json = const {
  '1': 'GetCurrentWalletResponse',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetCurrentWalletResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getCurrentWalletResponseDescriptor = $convert.base64Decode('ChhHZXRDdXJyZW50V2FsbGV0UmVzcG9uc2USDgoCaWQYASABKAlSAmlk');
@$core.Deprecated('Use getUserLimitsResponseDescriptor instead')
const GetUserLimitsResponse$json = const {
  '1': 'GetUserLimitsResponse',
  '2': const [
    const {'1': 'FundWallet', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.Limit', '10': 'FundWallet'},
    const {'1': 'Withdrawal', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.Limit', '10': 'Withdrawal'},
    const {'1': 'Transfer', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.Limit', '10': 'Transfer'},
  ],
};

/// Descriptor for `GetUserLimitsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getUserLimitsResponseDescriptor = $convert.base64Decode('ChVHZXRVc2VyTGltaXRzUmVzcG9uc2USMQoKRnVuZFdhbGxldBgBIAEoCzIRLmJhY2tlbmQudjEuTGltaXRSCkZ1bmRXYWxsZXQSMQoKV2l0aGRyYXdhbBgCIAEoCzIRLmJhY2tlbmQudjEuTGltaXRSCldpdGhkcmF3YWwSLQoIVHJhbnNmZXIYAyABKAsyES5iYWNrZW5kLnYxLkxpbWl0UghUcmFuc2Zlcg==');
@$core.Deprecated('Use limitDescriptor instead')
const Limit$json = const {
  '1': 'Limit',
  '2': const [
    const {'1': 'Annual', '3': 1, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'Annual'},
    const {'1': 'Daily', '3': 2, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'Daily'},
    const {'1': 'Monthly', '3': 3, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'Monthly'},
    const {'1': 'WalletHold', '3': 4, '4': 1, '5': 11, '6': '.backend.v1.LimitAmount', '10': 'WalletHold'},
  ],
};

/// Descriptor for `Limit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List limitDescriptor = $convert.base64Decode('CgVMaW1pdBIvCgZBbm51YWwYASABKAsyFy5iYWNrZW5kLnYxLkxpbWl0QW1vdW50UgZBbm51YWwSLQoFRGFpbHkYAiABKAsyFy5iYWNrZW5kLnYxLkxpbWl0QW1vdW50UgVEYWlseRIxCgdNb250aGx5GAMgASgLMhcuYmFja2VuZC52MS5MaW1pdEFtb3VudFIHTW9udGhseRI3CgpXYWxsZXRIb2xkGAQgASgLMhcuYmFja2VuZC52MS5MaW1pdEFtb3VudFIKV2FsbGV0SG9sZA==');
@$core.Deprecated('Use limitAmountDescriptor instead')
const LimitAmount$json = const {
  '1': 'LimitAmount',
  '2': const [
    const {'1': 'remaining', '3': 1, '4': 1, '5': 9, '10': 'remaining'},
    const {'1': 'total', '3': 2, '4': 1, '5': 9, '10': 'total'},
    const {'1': 'percentage', '3': 3, '4': 1, '5': 5, '10': 'percentage'},
  ],
};

/// Descriptor for `LimitAmount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List limitAmountDescriptor = $convert.base64Decode('CgtMaW1pdEFtb3VudBIcCglyZW1haW5pbmcYASABKAlSCXJlbWFpbmluZxIUCgV0b3RhbBgCIAEoCVIFdG90YWwSHgoKcGVyY2VudGFnZRgDIAEoBVIKcGVyY2VudGFnZQ==');
