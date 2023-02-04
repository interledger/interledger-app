//
//  Generated code. Do not modify.
//  source: backend/admin/v1/backend.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use setWalletCountryRequestDescriptor instead')
const SetWalletCountryRequest$json = {
  '1': 'SetWalletCountryRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'countryCode', '3': 2, '4': 1, '5': 9, '10': 'countryCode'},
  ],
};

/// Descriptor for `SetWalletCountryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setWalletCountryRequestDescriptor = $convert.base64Decode(
    'ChdTZXRXYWxsZXRDb3VudHJ5UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSIAoLY291bnRyeUNvZG'
    'UYAiABKAlSC2NvdW50cnlDb2Rl');

@$core.Deprecated('Use countryDescriptor instead')
const Country$json = {
  '1': 'Country',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'numeric', '3': 3, '4': 1, '5': 9, '10': 'numeric'},
    {'1': 'supported', '3': 4, '4': 1, '5': 8, '10': 'supported'},
  ],
};

/// Descriptor for `Country`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List countryDescriptor = $convert.base64Decode(
    'CgdDb3VudHJ5EhIKBGNvZGUYASABKAlSBGNvZGUSEgoEbmFtZRgCIAEoCVIEbmFtZRIYCgdudW'
    '1lcmljGAMgASgJUgdudW1lcmljEhwKCXN1cHBvcnRlZBgEIAEoCFIJc3VwcG9ydGVk');

@$core.Deprecated('Use listCountriesResponseDescriptor instead')
const ListCountriesResponse$json = {
  '1': 'ListCountriesResponse',
  '2': [
    {'1': 'countries', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.Country', '10': 'countries'},
  ],
};

/// Descriptor for `ListCountriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listCountriesResponseDescriptor = $convert.base64Decode(
    'ChVMaXN0Q291bnRyaWVzUmVzcG9uc2USNwoJY291bnRyaWVzGAEgAygLMhkuYmFja2VuZC5hZG'
    '1pbi52MS5Db3VudHJ5Ugljb3VudHJpZXM=');

@$core.Deprecated('Use listPaymentsAwaitingSignalResponseDescriptor instead')
const ListPaymentsAwaitingSignalResponse$json = {
  '1': 'ListPaymentsAwaitingSignalResponse',
  '2': [
    {'1': 'payments', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.Payment', '10': 'payments'},
  ],
};

/// Descriptor for `ListPaymentsAwaitingSignalResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listPaymentsAwaitingSignalResponseDescriptor = $convert.base64Decode(
    'CiJMaXN0UGF5bWVudHNBd2FpdGluZ1NpZ25hbFJlc3BvbnNlEjUKCHBheW1lbnRzGAEgAygLMh'
    'kuYmFja2VuZC5hZG1pbi52MS5QYXltZW50UghwYXltZW50cw==');

@$core.Deprecated('Use paymentDescriptor instead')
const Payment$json = {
  '1': 'Payment',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'publicID', '3': 2, '4': 1, '5': 9, '10': 'publicID'},
    {'1': 'state', '3': 3, '4': 1, '5': 9, '10': 'state'},
    {'1': 'receiverWalletUrl', '3': 4, '4': 1, '5': 9, '10': 'receiverWalletUrl'},
    {'1': 'receiverIdentity', '3': 5, '4': 1, '5': 9, '10': 'receiverIdentity'},
    {'1': 'receiverIdentityType', '3': 6, '4': 1, '5': 9, '10': 'receiverIdentityType'},
    {'1': 'senderAmount', '3': 7, '4': 1, '5': 9, '10': 'senderAmount'},
    {'1': 'senderAccount', '3': 8, '4': 1, '5': 9, '10': 'senderAccount'},
    {'1': 'note', '3': 9, '4': 1, '5': 9, '10': 'note'},
    {'1': 'requiredActions', '3': 10, '4': 3, '5': 9, '10': 'requiredActions'},
    {'1': 'senderWalletUrl', '3': 11, '4': 1, '5': 9, '10': 'senderWalletUrl'},
    {'1': 'updatedAt', '3': 12, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'updatedAt'},
  ],
};

/// Descriptor for `Payment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List paymentDescriptor = $convert.base64Decode(
    'CgdQYXltZW50Eg4KAmlkGAEgASgJUgJpZBIaCghwdWJsaWNJRBgCIAEoCVIIcHVibGljSUQSFA'
    'oFc3RhdGUYAyABKAlSBXN0YXRlEiwKEXJlY2VpdmVyV2FsbGV0VXJsGAQgASgJUhFyZWNlaXZl'
    'cldhbGxldFVybBIqChByZWNlaXZlcklkZW50aXR5GAUgASgJUhByZWNlaXZlcklkZW50aXR5Ej'
    'IKFHJlY2VpdmVySWRlbnRpdHlUeXBlGAYgASgJUhRyZWNlaXZlcklkZW50aXR5VHlwZRIiCgxz'
    'ZW5kZXJBbW91bnQYByABKAlSDHNlbmRlckFtb3VudBIkCg1zZW5kZXJBY2NvdW50GAggASgJUg'
    '1zZW5kZXJBY2NvdW50EhIKBG5vdGUYCSABKAlSBG5vdGUSKAoPcmVxdWlyZWRBY3Rpb25zGAog'
    'AygJUg9yZXF1aXJlZEFjdGlvbnMSKAoPc2VuZGVyV2FsbGV0VXJsGAsgASgJUg9zZW5kZXJXYW'
    'xsZXRVcmwSOAoJdXBkYXRlZEF0GAwgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJ'
    'dXBkYXRlZEF0');

@$core.Deprecated('Use listExternalApiCallsRequestDescriptor instead')
const ListExternalApiCallsRequest$json = {
  '1': 'ListExternalApiCallsRequest',
  '2': [
    {'1': 'paymentId', '3': 1, '4': 1, '5': 9, '10': 'paymentId'},
  ],
};

/// Descriptor for `ListExternalApiCallsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listExternalApiCallsRequestDescriptor = $convert.base64Decode(
    'ChtMaXN0RXh0ZXJuYWxBcGlDYWxsc1JlcXVlc3QSHAoJcGF5bWVudElkGAEgASgJUglwYXltZW'
    '50SWQ=');

@$core.Deprecated('Use externalApiCallDescriptor instead')
const ExternalApiCall$json = {
  '1': 'ExternalApiCall',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'provider', '3': 2, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'context', '3': 3, '4': 1, '5': 9, '10': 'context'},
    {'1': 'method', '3': 4, '4': 1, '5': 9, '10': 'method'},
    {'1': 'requestBody', '3': 5, '4': 1, '5': 9, '10': 'requestBody'},
    {'1': 'requestPath', '3': 6, '4': 1, '5': 9, '10': 'requestPath'},
    {'1': 'responseBody', '3': 7, '4': 1, '5': 9, '10': 'responseBody'},
    {'1': 'responseStatus', '3': 8, '4': 1, '5': 9, '10': 'responseStatus'},
  ],
};

/// Descriptor for `ExternalApiCall`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List externalApiCallDescriptor = $convert.base64Decode(
    'Cg9FeHRlcm5hbEFwaUNhbGwSDgoCaWQYASABKAlSAmlkEhoKCHByb3ZpZGVyGAIgASgJUghwcm'
    '92aWRlchIYCgdjb250ZXh0GAMgASgJUgdjb250ZXh0EhYKBm1ldGhvZBgEIAEoCVIGbWV0aG9k'
    'EiAKC3JlcXVlc3RCb2R5GAUgASgJUgtyZXF1ZXN0Qm9keRIgCgtyZXF1ZXN0UGF0aBgGIAEoCV'
    'ILcmVxdWVzdFBhdGgSIgoMcmVzcG9uc2VCb2R5GAcgASgJUgxyZXNwb25zZUJvZHkSJgoOcmVz'
    'cG9uc2VTdGF0dXMYCCABKAlSDnJlc3BvbnNlU3RhdHVz');

@$core.Deprecated('Use listExternalApiCallsResponseDescriptor instead')
const ListExternalApiCallsResponse$json = {
  '1': 'ListExternalApiCallsResponse',
  '2': [
    {'1': 'list', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.ExternalApiCall', '10': 'list'},
  ],
};

/// Descriptor for `ListExternalApiCallsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listExternalApiCallsResponseDescriptor = $convert.base64Decode(
    'ChxMaXN0RXh0ZXJuYWxBcGlDYWxsc1Jlc3BvbnNlEjUKBGxpc3QYASADKAsyIS5iYWNrZW5kLm'
    'FkbWluLnYxLkV4dGVybmFsQXBpQ2FsbFIEbGlzdA==');

@$core.Deprecated('Use emptyDescriptor instead')
const Empty$json = {
  '1': 'Empty',
};

/// Descriptor for `Empty`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List emptyDescriptor = $convert.base64Decode(
    'CgVFbXB0eQ==');

@$core.Deprecated('Use linkedAccountReviewDescriptor instead')
const LinkedAccountReview$json = {
  '1': 'LinkedAccountReview',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'state', '3': 2, '4': 1, '5': 9, '10': 'state'},
    {'1': 'newState', '3': 3, '4': 1, '5': 9, '10': 'newState'},
    {'1': 'linkedAccountID', '3': 4, '4': 1, '5': 9, '10': 'linkedAccountID'},
    {'1': 'reviewedBy', '3': 5, '4': 1, '5': 9, '10': 'reviewedBy'},
    {'1': 'reason', '3': 6, '4': 1, '5': 9, '10': 'reason'},
    {'1': 'walletID', '3': 9, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'walletName', '3': 10, '4': 1, '5': 9, '10': 'walletName'},
    {'1': 'mask', '3': 11, '4': 1, '5': 9, '10': 'mask'},
    {'1': 'createdAt', '3': 7, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'createdAt'},
    {'1': 'completedAt', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'completedAt'},
  ],
};

/// Descriptor for `LinkedAccountReview`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkedAccountReviewDescriptor = $convert.base64Decode(
    'ChNMaW5rZWRBY2NvdW50UmV2aWV3Eg4KAmlkGAEgASgJUgJpZBIUCgVzdGF0ZRgCIAEoCVIFc3'
    'RhdGUSGgoIbmV3U3RhdGUYAyABKAlSCG5ld1N0YXRlEigKD2xpbmtlZEFjY291bnRJRBgEIAEo'
    'CVIPbGlua2VkQWNjb3VudElEEh4KCnJldmlld2VkQnkYBSABKAlSCnJldmlld2VkQnkSFgoGcm'
    'Vhc29uGAYgASgJUgZyZWFzb24SGgoId2FsbGV0SUQYCSABKAlSCHdhbGxldElEEh4KCndhbGxl'
    'dE5hbWUYCiABKAlSCndhbGxldE5hbWUSEgoEbWFzaxgLIAEoCVIEbWFzaxI4CgljcmVhdGVkQX'
    'QYByABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgljcmVhdGVkQXQSPAoLY29tcGxl'
    'dGVkQXQYCCABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgtjb21wbGV0ZWRBdA==');

@$core.Deprecated('Use linkedAccountReviewsDescriptor instead')
const LinkedAccountReviews$json = {
  '1': 'LinkedAccountReviews',
  '2': [
    {'1': 'reviews', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.LinkedAccountReview', '10': 'reviews'},
  ],
};

/// Descriptor for `LinkedAccountReviews`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkedAccountReviewsDescriptor = $convert.base64Decode(
    'ChRMaW5rZWRBY2NvdW50UmV2aWV3cxI/CgdyZXZpZXdzGAEgAygLMiUuYmFja2VuZC5hZG1pbi'
    '52MS5MaW5rZWRBY2NvdW50UmV2aWV3UgdyZXZpZXdz');

@$core.Deprecated('Use getLinkedAccountReviewRequestDescriptor instead')
const GetLinkedAccountReviewRequest$json = {
  '1': 'GetLinkedAccountReviewRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetLinkedAccountReviewRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinkedAccountReviewRequestDescriptor = $convert.base64Decode(
    'Ch1HZXRMaW5rZWRBY2NvdW50UmV2aWV3UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use completeLinkedAccountReviewRequestDescriptor instead')
const CompleteLinkedAccountReviewRequest$json = {
  '1': 'CompleteLinkedAccountReviewRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'reason', '3': 2, '4': 1, '5': 9, '10': 'reason'},
    {'1': 'newState', '3': 3, '4': 1, '5': 9, '10': 'newState'},
  ],
};

/// Descriptor for `CompleteLinkedAccountReviewRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeLinkedAccountReviewRequestDescriptor = $convert.base64Decode(
    'CiJDb21wbGV0ZUxpbmtlZEFjY291bnRSZXZpZXdSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZBIWCg'
    'ZyZWFzb24YAiABKAlSBnJlYXNvbhIaCghuZXdTdGF0ZRgDIAEoCVIIbmV3U3RhdGU=');

@$core.Deprecated('Use getWalletFeaturesRequestDescriptor instead')
const GetWalletFeaturesRequest$json = {
  '1': 'GetWalletFeaturesRequest',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
  ],
};

/// Descriptor for `GetWalletFeaturesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWalletFeaturesRequestDescriptor = $convert.base64Decode(
    'ChhHZXRXYWxsZXRGZWF0dXJlc1JlcXVlc3QSGgoId2FsbGV0SUQYASABKAlSCHdhbGxldElE');

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
    {'1': 'walletID', '3': 8, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'addCardsEnabled', '3': 9, '4': 1, '5': 8, '10': 'addCardsEnabled'},
    {'1': 'zarBalanceEnabled', '3': 10, '4': 1, '5': 8, '10': 'zarBalanceEnabled'},
  ],
};

/// Descriptor for `Features`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List featuresDescriptor = $convert.base64Decode(
    'CghGZWF0dXJlcxIgCgtzZW5kRW5hYmxlZBgBIAEoCFILc2VuZEVuYWJsZWQSJgoOcmVjZWl2ZU'
    'VuYWJsZWQYAiABKAhSDnJlY2VpdmVFbmFibGVkEjQKFWxpbmtlZEFjY291bnRzRW5hYmxlZBgD'
    'IAEoCFIVbGlua2VkQWNjb3VudHNFbmFibGVkEiIKDGNhcmRzRW5hYmxlZBgEIAEoCFIMY2FyZH'
    'NFbmFibGVkEiIKDGJhbmtzRW5hYmxlZBgFIAEoCFIMYmFua3NFbmFibGVkEiwKEWlkZW50aXRp'
    'ZXNFbmFibGVkGAYgASgIUhFpZGVudGl0aWVzRW5hYmxlZBImCg50d2l0dGVyRW5hYmxlZBgHIA'
    'EoCFIOdHdpdHRlckVuYWJsZWQSGgoId2FsbGV0SUQYCCABKAlSCHdhbGxldElEEigKD2FkZENh'
    'cmRzRW5hYmxlZBgJIAEoCFIPYWRkQ2FyZHNFbmFibGVkEiwKEXphckJhbGFuY2VFbmFibGVkGA'
    'ogASgIUhF6YXJCYWxhbmNlRW5hYmxlZA==');

@$core.Deprecated('Use listAuditRequestDescriptor instead')
const ListAuditRequest$json = {
  '1': 'ListAuditRequest',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
  ],
};

/// Descriptor for `ListAuditRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAuditRequestDescriptor = $convert.base64Decode(
    'ChBMaXN0QXVkaXRSZXF1ZXN0EhoKCHdhbGxldElEGAEgASgJUgh3YWxsZXRJRA==');

@$core.Deprecated('Use listAuditResponseDescriptor instead')
const ListAuditResponse$json = {
  '1': 'ListAuditResponse',
  '2': [
    {'1': 'operations', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.AuditOperation', '10': 'operations'},
  ],
};

/// Descriptor for `ListAuditResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAuditResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0QXVkaXRSZXNwb25zZRJACgpvcGVyYXRpb25zGAEgAygLMiAuYmFja2VuZC5hZG1pbi'
    '52MS5BdWRpdE9wZXJhdGlvblIKb3BlcmF0aW9ucw==');

@$core.Deprecated('Use auditOperationDescriptor instead')
const AuditOperation$json = {
  '1': 'AuditOperation',
  '2': [
    {'1': 'adminUser', '3': 1, '4': 1, '5': 9, '10': 'adminUser'},
    {'1': 'walletID', '3': 2, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'operation', '3': 3, '4': 1, '5': 9, '10': 'operation'},
    {'1': 'parameters', '3': 4, '4': 1, '5': 9, '10': 'parameters'},
    {'1': 'timestamp', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
  ],
};

/// Descriptor for `AuditOperation`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List auditOperationDescriptor = $convert.base64Decode(
    'Cg5BdWRpdE9wZXJhdGlvbhIcCglhZG1pblVzZXIYASABKAlSCWFkbWluVXNlchIaCgh3YWxsZX'
    'RJRBgCIAEoCVIId2FsbGV0SUQSHAoJb3BlcmF0aW9uGAMgASgJUglvcGVyYXRpb24SHgoKcGFy'
    'YW1ldGVycxgEIAEoCVIKcGFyYW1ldGVycxI4Cgl0aW1lc3RhbXAYBSABKAsyGi5nb29nbGUucH'
    'JvdG9idWYuVGltZXN0YW1wUgl0aW1lc3RhbXA=');

@$core.Deprecated('Use listLinkedAccountsRequestDescriptor instead')
const ListLinkedAccountsRequest$json = {
  '1': 'ListLinkedAccountsRequest',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
  ],
};

/// Descriptor for `ListLinkedAccountsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinkedAccountsRequestDescriptor = $convert.base64Decode(
    'ChlMaXN0TGlua2VkQWNjb3VudHNSZXF1ZXN0EhoKCHdhbGxldElEGAEgASgJUgh3YWxsZXRJRA'
    '==');

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

@$core.Deprecated('Use listLinkedAccountsResponseDescriptor instead')
const ListLinkedAccountsResponse$json = {
  '1': 'ListLinkedAccountsResponse',
  '2': [
    {'1': 'accounts', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.LinkedAccount', '10': 'accounts'},
  ],
};

/// Descriptor for `ListLinkedAccountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinkedAccountsResponseDescriptor = $convert.base64Decode(
    'ChpMaXN0TGlua2VkQWNjb3VudHNSZXNwb25zZRI7CghhY2NvdW50cxgBIAMoCzIfLmJhY2tlbm'
    'QuYWRtaW4udjEuTGlua2VkQWNjb3VudFIIYWNjb3VudHM=');

@$core.Deprecated('Use linkedAccountDescriptor instead')
const LinkedAccount$json = {
  '1': 'LinkedAccount',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'walletID', '3': 2, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'nickname', '3': 4, '4': 1, '5': 9, '10': 'nickname'},
    {'1': 'mask', '3': 5, '4': 1, '5': 9, '10': 'mask'},
    {'1': 'provider', '3': 6, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'providerID', '3': 7, '4': 1, '5': 9, '10': 'providerID'},
    {'1': 'type', '3': 8, '4': 1, '5': 9, '10': 'type'},
    {'1': 'state', '3': 9, '4': 1, '5': 9, '10': 'state'},
    {'1': 'canSend', '3': 10, '4': 1, '5': 9, '10': 'canSend'},
    {'1': 'canReceive', '3': 11, '4': 1, '5': 9, '10': 'canReceive'},
    {'1': 'sendCurrencyCode', '3': 12, '4': 1, '5': 9, '10': 'sendCurrencyCode'},
    {'1': 'sendCurrencyCountryCode', '3': 13, '4': 1, '5': 9, '10': 'sendCurrencyCountryCode'},
    {'1': 'receiveCurrencyCode', '3': 14, '4': 1, '5': 9, '10': 'receiveCurrencyCode'},
    {'1': 'receiveCurrencyCountryCode', '3': 15, '4': 1, '5': 9, '10': 'receiveCurrencyCountryCode'},
    {'1': 'deletedAt', '3': 16, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'deletedAt'},
    {'1': 'defaultSend', '3': 17, '4': 1, '5': 8, '10': 'defaultSend'},
    {'1': 'defaultReceive', '3': 18, '4': 1, '5': 8, '10': 'defaultReceive'},
  ],
};

/// Descriptor for `LinkedAccount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linkedAccountDescriptor = $convert.base64Decode(
    'Cg1MaW5rZWRBY2NvdW50Eg4KAmlkGAEgASgJUgJpZBIaCgh3YWxsZXRJRBgCIAEoCVIId2FsbG'
    'V0SUQSEgoEbmFtZRgDIAEoCVIEbmFtZRIaCghuaWNrbmFtZRgEIAEoCVIIbmlja25hbWUSEgoE'
    'bWFzaxgFIAEoCVIEbWFzaxIaCghwcm92aWRlchgGIAEoCVIIcHJvdmlkZXISHgoKcHJvdmlkZX'
    'JJRBgHIAEoCVIKcHJvdmlkZXJJRBISCgR0eXBlGAggASgJUgR0eXBlEhQKBXN0YXRlGAkgASgJ'
    'UgVzdGF0ZRIYCgdjYW5TZW5kGAogASgJUgdjYW5TZW5kEh4KCmNhblJlY2VpdmUYCyABKAlSCm'
    'NhblJlY2VpdmUSKgoQc2VuZEN1cnJlbmN5Q29kZRgMIAEoCVIQc2VuZEN1cnJlbmN5Q29kZRI4'
    'ChdzZW5kQ3VycmVuY3lDb3VudHJ5Q29kZRgNIAEoCVIXc2VuZEN1cnJlbmN5Q291bnRyeUNvZG'
    'USMAoTcmVjZWl2ZUN1cnJlbmN5Q29kZRgOIAEoCVITcmVjZWl2ZUN1cnJlbmN5Q29kZRI+Chpy'
    'ZWNlaXZlQ3VycmVuY3lDb3VudHJ5Q29kZRgPIAEoCVIacmVjZWl2ZUN1cnJlbmN5Q291bnRyeU'
    'NvZGUSOAoJZGVsZXRlZEF0GBAgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJZGVs'
    'ZXRlZEF0EiAKC2RlZmF1bHRTZW5kGBEgASgIUgtkZWZhdWx0U2VuZBImCg5kZWZhdWx0UmVjZW'
    'l2ZRgSIAEoCFIOZGVmYXVsdFJlY2VpdmU=');

@$core.Deprecated('Use getTransactionDetailsRequestDescriptor instead')
const GetTransactionDetailsRequest$json = {
  '1': 'GetTransactionDetailsRequest',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'transactionID', '3': 2, '4': 1, '5': 9, '10': 'transactionID'},
  ],
};

/// Descriptor for `GetTransactionDetailsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTransactionDetailsRequestDescriptor = $convert.base64Decode(
    'ChxHZXRUcmFuc2FjdGlvbkRldGFpbHNSZXF1ZXN0EhoKCHdhbGxldElEGAEgASgJUgh3YWxsZX'
    'RJRBIkCg10cmFuc2FjdGlvbklEGAIgASgJUg10cmFuc2FjdGlvbklE');

@$core.Deprecated('Use getTransactionDetailsResponseDescriptor instead')
const GetTransactionDetailsResponse$json = {
  '1': 'GetTransactionDetailsResponse',
  '2': [
    {'1': 'transaction', '3': 1, '4': 1, '5': 11, '6': '.backend.admin.v1.Transaction', '10': 'transaction'},
    {'1': 'transfers', '3': 2, '4': 3, '5': 11, '6': '.backend.admin.v1.Transfer', '10': 'transfers'},
  ],
};

/// Descriptor for `GetTransactionDetailsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTransactionDetailsResponseDescriptor = $convert.base64Decode(
    'Ch1HZXRUcmFuc2FjdGlvbkRldGFpbHNSZXNwb25zZRI/Cgt0cmFuc2FjdGlvbhgBIAEoCzIdLm'
    'JhY2tlbmQuYWRtaW4udjEuVHJhbnNhY3Rpb25SC3RyYW5zYWN0aW9uEjgKCXRyYW5zZmVycxgC'
    'IAMoCzIaLmJhY2tlbmQuYWRtaW4udjEuVHJhbnNmZXJSCXRyYW5zZmVycw==');

@$core.Deprecated('Use transferDescriptor instead')
const Transfer$json = {
  '1': 'Transfer',
  '2': [
    {'1': 'ID', '3': 1, '4': 1, '5': 9, '10': 'ID'},
    {'1': 'linkedAccountID', '3': 2, '4': 1, '5': 9, '10': 'linkedAccountID'},
    {'1': 'linkedAccountProvider', '3': 3, '4': 1, '5': 9, '10': 'linkedAccountProvider'},
    {'1': 'linkedAccountType', '3': 4, '4': 1, '5': 9, '10': 'linkedAccountType'},
    {'1': 'foreignID', '3': 9, '4': 1, '5': 9, '10': 'foreignID'},
    {'1': 'amount', '3': 5, '4': 1, '5': 1, '10': 'amount'},
    {'1': 'currency', '3': 6, '4': 1, '5': 9, '10': 'currency'},
    {'1': 'state', '3': 7, '4': 1, '5': 9, '10': 'state'},
    {'1': 'timestamp', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
  ],
};

/// Descriptor for `Transfer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transferDescriptor = $convert.base64Decode(
    'CghUcmFuc2ZlchIOCgJJRBgBIAEoCVICSUQSKAoPbGlua2VkQWNjb3VudElEGAIgASgJUg9saW'
    '5rZWRBY2NvdW50SUQSNAoVbGlua2VkQWNjb3VudFByb3ZpZGVyGAMgASgJUhVsaW5rZWRBY2Nv'
    'dW50UHJvdmlkZXISLAoRbGlua2VkQWNjb3VudFR5cGUYBCABKAlSEWxpbmtlZEFjY291bnRUeX'
    'BlEhwKCWZvcmVpZ25JRBgJIAEoCVIJZm9yZWlnbklEEhYKBmFtb3VudBgFIAEoAVIGYW1vdW50'
    'EhoKCGN1cnJlbmN5GAYgASgJUghjdXJyZW5jeRIUCgVzdGF0ZRgHIAEoCVIFc3RhdGUSOAoJdG'
    'ltZXN0YW1wGAggASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJdGltZXN0YW1w');

@$core.Deprecated('Use listTransactionsRequestDescriptor instead')
const ListTransactionsRequest$json = {
  '1': 'ListTransactionsRequest',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'page', '3': 2, '4': 1, '5': 11, '6': '.backend.admin.v1.PaginationRequest', '10': 'page'},
  ],
};

/// Descriptor for `ListTransactionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTransactionsRequestDescriptor = $convert.base64Decode(
    'ChdMaXN0VHJhbnNhY3Rpb25zUmVxdWVzdBIaCgh3YWxsZXRJRBgBIAEoCVIId2FsbGV0SUQSNw'
    'oEcGFnZRgCIAEoCzIjLmJhY2tlbmQuYWRtaW4udjEuUGFnaW5hdGlvblJlcXVlc3RSBHBhZ2U=');

@$core.Deprecated('Use listTransactionsResponseDescriptor instead')
const ListTransactionsResponse$json = {
  '1': 'ListTransactionsResponse',
  '2': [
    {'1': 'transactions', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.Transaction', '10': 'transactions'},
    {'1': 'nextPageToken', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListTransactionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTransactionsResponseDescriptor = $convert.base64Decode(
    'ChhMaXN0VHJhbnNhY3Rpb25zUmVzcG9uc2USQQoMdHJhbnNhY3Rpb25zGAEgAygLMh0uYmFja2'
    'VuZC5hZG1pbi52MS5UcmFuc2FjdGlvblIMdHJhbnNhY3Rpb25zEiQKDW5leHRQYWdlVG9rZW4Y'
    'AiABKAlSDW5leHRQYWdlVG9rZW4=');

@$core.Deprecated('Use transactionDescriptor instead')
const Transaction$json = {
  '1': 'Transaction',
  '2': [
    {'1': 'walletID', '3': 8, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'asset', '3': 7, '4': 1, '5': 9, '10': 'asset'},
    {'1': 'amount', '3': 3, '4': 1, '5': 1, '10': 'amount'},
    {'1': 'source', '3': 4, '4': 1, '5': 9, '10': 'source'},
    {'1': 'destination', '3': 5, '4': 1, '5': 9, '10': 'destination'},
    {'1': 'timestamp', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
    {'1': 'paymentId', '3': 9, '4': 1, '5': 9, '10': 'paymentId'},
  ],
};

/// Descriptor for `Transaction`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List transactionDescriptor = $convert.base64Decode(
    'CgtUcmFuc2FjdGlvbhIaCgh3YWxsZXRJRBgIIAEoCVIId2FsbGV0SUQSDgoCaWQYASABKAlSAm'
    'lkEhIKBHR5cGUYAiABKAlSBHR5cGUSFAoFYXNzZXQYByABKAlSBWFzc2V0EhYKBmFtb3VudBgD'
    'IAEoAVIGYW1vdW50EhYKBnNvdXJjZRgEIAEoCVIGc291cmNlEiAKC2Rlc3RpbmF0aW9uGAUgAS'
    'gJUgtkZXN0aW5hdGlvbhI4Cgl0aW1lc3RhbXAYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVGlt'
    'ZXN0YW1wUgl0aW1lc3RhbXASHAoJcGF5bWVudElkGAkgASgJUglwYXltZW50SWQ=');

@$core.Deprecated('Use getUserTransactionsRequestDescriptor instead')
const GetUserTransactionsRequest$json = {
  '1': 'GetUserTransactionsRequest',
  '2': [
    {'1': 'userID', '3': 1, '4': 1, '5': 9, '10': 'userID'},
  ],
};

/// Descriptor for `GetUserTransactionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getUserTransactionsRequestDescriptor = $convert.base64Decode(
    'ChpHZXRVc2VyVHJhbnNhY3Rpb25zUmVxdWVzdBIWCgZ1c2VySUQYASABKAlSBnVzZXJJRA==');

@$core.Deprecated('Use getWalletDetailsRequestDescriptor instead')
const GetWalletDetailsRequest$json = {
  '1': 'GetWalletDetailsRequest',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
  ],
};

/// Descriptor for `GetWalletDetailsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWalletDetailsRequestDescriptor = $convert.base64Decode(
    'ChdHZXRXYWxsZXREZXRhaWxzUmVxdWVzdBIaCgh3YWxsZXRJRBgBIAEoCVIId2FsbGV0SUQ=');

@$core.Deprecated('Use walletDetailsDescriptor instead')
const WalletDetails$json = {
  '1': 'WalletDetails',
  '2': [
    {'1': 'users', '3': 8, '4': 3, '5': 11, '6': '.backend.admin.v1.User', '10': 'users'},
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'firstName', '3': 2, '4': 1, '5': 9, '10': 'firstName'},
    {'1': 'lastName', '3': 3, '4': 1, '5': 9, '10': 'lastName'},
    {'1': 'countryCode', '3': 4, '4': 1, '5': 9, '10': 'countryCode'},
    {'1': 'gender', '3': 5, '4': 1, '5': 5, '10': 'gender'},
    {'1': 'dateOfBirth', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'dateOfBirth'},
    {'1': 'address', '3': 7, '4': 1, '5': 9, '10': 'address'},
    {'1': 'kycStatus', '3': 9, '4': 1, '5': 9, '10': 'kycStatus'},
    {'1': 'placeOfBirth', '3': 10, '4': 1, '5': 9, '10': 'placeOfBirth'},
    {'1': 'nationality', '3': 11, '4': 1, '5': 9, '10': 'nationality'},
    {'1': 'walletName', '3': 12, '4': 1, '5': 9, '10': 'walletName'},
  ],
};

/// Descriptor for `WalletDetails`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletDetailsDescriptor = $convert.base64Decode(
    'Cg1XYWxsZXREZXRhaWxzEiwKBXVzZXJzGAggAygLMhYuYmFja2VuZC5hZG1pbi52MS5Vc2VyUg'
    'V1c2VycxIaCgh3YWxsZXRJRBgBIAEoCVIId2FsbGV0SUQSHAoJZmlyc3ROYW1lGAIgASgJUglm'
    'aXJzdE5hbWUSGgoIbGFzdE5hbWUYAyABKAlSCGxhc3ROYW1lEiAKC2NvdW50cnlDb2RlGAQgAS'
    'gJUgtjb3VudHJ5Q29kZRIWCgZnZW5kZXIYBSABKAVSBmdlbmRlchI8CgtkYXRlT2ZCaXJ0aBgG'
    'IAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSC2RhdGVPZkJpcnRoEhgKB2FkZHJlc3'
    'MYByABKAlSB2FkZHJlc3MSHAoJa3ljU3RhdHVzGAkgASgJUglreWNTdGF0dXMSIgoMcGxhY2VP'
    'ZkJpcnRoGAogASgJUgxwbGFjZU9mQmlydGgSIAoLbmF0aW9uYWxpdHkYCyABKAlSC25hdGlvbm'
    'FsaXR5Eh4KCndhbGxldE5hbWUYDCABKAlSCndhbGxldE5hbWU=');

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

@$core.Deprecated('Use walletDescriptor instead')
const Wallet$json = {
  '1': 'Wallet',
  '2': [
    {'1': 'walletID', '3': 1, '4': 1, '5': 9, '10': 'walletID'},
    {'1': 'walletName', '3': 2, '4': 1, '5': 9, '10': 'walletName'},
    {'1': 'users', '3': 3, '4': 3, '5': 11, '6': '.backend.admin.v1.User', '10': 'users'},
  ],
};

/// Descriptor for `Wallet`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List walletDescriptor = $convert.base64Decode(
    'CgZXYWxsZXQSGgoId2FsbGV0SUQYASABKAlSCHdhbGxldElEEh4KCndhbGxldE5hbWUYAiABKA'
    'lSCndhbGxldE5hbWUSLAoFdXNlcnMYAyADKAsyFi5iYWNrZW5kLmFkbWluLnYxLlVzZXJSBXVz'
    'ZXJz');

@$core.Deprecated('Use listWalletsResponseDescriptor instead')
const ListWalletsResponse$json = {
  '1': 'ListWalletsResponse',
  '2': [
    {'1': 'wallets', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.Wallet', '10': 'wallets'},
    {'1': 'nextPageToken', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListWalletsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWalletsResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0V2FsbGV0c1Jlc3BvbnNlEjIKB3dhbGxldHMYASADKAsyGC5iYWNrZW5kLmFkbWluLn'
    'YxLldhbGxldFIHd2FsbGV0cxIkCg1uZXh0UGFnZVRva2VuGAIgASgJUg1uZXh0UGFnZVRva2Vu');

@$core.Deprecated('Use userDescriptor instead')
const User$json = {
  '1': 'User',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'email', '3': 2, '4': 1, '5': 9, '10': 'email'},
    {'1': 'phoneNumber', '3': 3, '4': 1, '5': 9, '10': 'phoneNumber'},
  ],
};

/// Descriptor for `User`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userDescriptor = $convert.base64Decode(
    'CgRVc2VyEg4KAmlkGAEgASgJUgJpZBIUCgVlbWFpbBgCIAEoCVIFZW1haWwSIAoLcGhvbmVOdW'
    '1iZXIYAyABKAlSC3Bob25lTnVtYmVy');

@$core.Deprecated('Use allowWaitlistSignupRequestDescriptor instead')
const AllowWaitlistSignupRequest$json = {
  '1': 'AllowWaitlistSignupRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `AllowWaitlistSignupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allowWaitlistSignupRequestDescriptor = $convert.base64Decode(
    'ChpBbGxvd1dhaXRsaXN0U2lnbnVwUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use listWaitlistSignupsResponseDescriptor instead')
const ListWaitlistSignupsResponse$json = {
  '1': 'ListWaitlistSignupsResponse',
  '2': [
    {'1': 'signups', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.WaitlistSignup', '10': 'signups'},
  ],
};

/// Descriptor for `ListWaitlistSignupsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWaitlistSignupsResponseDescriptor = $convert.base64Decode(
    'ChtMaXN0V2FpdGxpc3RTaWdudXBzUmVzcG9uc2USOgoHc2lnbnVwcxgBIAMoCzIgLmJhY2tlbm'
    'QuYWRtaW4udjEuV2FpdGxpc3RTaWdudXBSB3NpZ251cHM=');

@$core.Deprecated('Use waitlistSignupDescriptor instead')
const WaitlistSignup$json = {
  '1': 'WaitlistSignup',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'email', '3': 3, '4': 1, '5': 9, '10': 'email'},
    {'1': 'beta_opt_in', '3': 4, '4': 1, '5': 8, '10': 'betaOptIn'},
    {'1': 'can_signup', '3': 5, '4': 1, '5': 8, '10': 'canSignup'},
    {'1': 'mug_id', '3': 6, '4': 1, '5': 9, '10': 'mugId'},
    {'1': 'country_code', '3': 7, '4': 1, '5': 9, '10': 'countryCode'},
  ],
};

/// Descriptor for `WaitlistSignup`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitlistSignupDescriptor = $convert.base64Decode(
    'Cg5XYWl0bGlzdFNpZ251cBIOCgJpZBgBIAEoCVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZRIUCg'
    'VlbWFpbBgDIAEoCVIFZW1haWwSHgoLYmV0YV9vcHRfaW4YBCABKAhSCWJldGFPcHRJbhIdCgpj'
    'YW5fc2lnbnVwGAUgASgIUgljYW5TaWdudXASFQoGbXVnX2lkGAYgASgJUgVtdWdJZBIhCgxjb3'
    'VudHJ5X2NvZGUYByABKAlSC2NvdW50cnlDb2Rl');

@$core.Deprecated('Use formSubmissionCountDescriptor instead')
const FormSubmissionCount$json = {
  '1': 'FormSubmissionCount',
  '2': [
    {'1': 'form_id', '3': 1, '4': 1, '5': 9, '10': 'formId'},
    {'1': 'submission_count', '3': 2, '4': 1, '5': 5, '10': 'submissionCount'},
  ],
};

/// Descriptor for `FormSubmissionCount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List formSubmissionCountDescriptor = $convert.base64Decode(
    'ChNGb3JtU3VibWlzc2lvbkNvdW50EhcKB2Zvcm1faWQYASABKAlSBmZvcm1JZBIpChBzdWJtaX'
    'NzaW9uX2NvdW50GAIgASgFUg9zdWJtaXNzaW9uQ291bnQ=');

@$core.Deprecated('Use listFormSubmissionCountsResponseDescriptor instead')
const ListFormSubmissionCountsResponse$json = {
  '1': 'ListFormSubmissionCountsResponse',
  '2': [
    {'1': 'form_submission_counts', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.FormSubmissionCount', '10': 'formSubmissionCounts'},
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListFormSubmissionCountsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listFormSubmissionCountsResponseDescriptor = $convert.base64Decode(
    'CiBMaXN0Rm9ybVN1Ym1pc3Npb25Db3VudHNSZXNwb25zZRJbChZmb3JtX3N1Ym1pc3Npb25fY2'
    '91bnRzGAEgAygLMiUuYmFja2VuZC5hZG1pbi52MS5Gb3JtU3VibWlzc2lvbkNvdW50UhRmb3Jt'
    'U3VibWlzc2lvbkNvdW50cxImCg9uZXh0X3BhZ2VfdG9rZW4YAiABKAlSDW5leHRQYWdlVG9rZW'
    '4=');

@$core.Deprecated('Use exportFormSubmissionsRequestDescriptor instead')
const ExportFormSubmissionsRequest$json = {
  '1': 'ExportFormSubmissionsRequest',
  '2': [
    {'1': 'form_id', '3': 1, '4': 1, '5': 9, '10': 'formId'},
  ],
};

/// Descriptor for `ExportFormSubmissionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List exportFormSubmissionsRequestDescriptor = $convert.base64Decode(
    'ChxFeHBvcnRGb3JtU3VibWlzc2lvbnNSZXF1ZXN0EhcKB2Zvcm1faWQYASABKAlSBmZvcm1JZA'
    '==');

@$core.Deprecated('Use exportFormSubmissionsResponseDescriptor instead')
const ExportFormSubmissionsResponse$json = {
  '1': 'ExportFormSubmissionsResponse',
  '2': [
    {'1': 'chunk', '3': 1, '4': 1, '5': 12, '10': 'chunk'},
  ],
};

/// Descriptor for `ExportFormSubmissionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List exportFormSubmissionsResponseDescriptor = $convert.base64Decode(
    'Ch1FeHBvcnRGb3JtU3VibWlzc2lvbnNSZXNwb25zZRIUCgVjaHVuaxgBIAEoDFIFY2h1bms=');

@$core.Deprecated('Use formSubmissionDescriptor instead')
const FormSubmission$json = {
  '1': 'FormSubmission',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'form_id', '3': 2, '4': 1, '5': 9, '10': 'formId'},
    {'1': 'timestamp', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
  ],
};

/// Descriptor for `FormSubmission`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List formSubmissionDescriptor = $convert.base64Decode(
    'Cg5Gb3JtU3VibWlzc2lvbhIOCgJpZBgBIAEoCVICaWQSFwoHZm9ybV9pZBgCIAEoCVIGZm9ybU'
    'lkEjgKCXRpbWVzdGFtcBgDIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXRpbWVz'
    'dGFtcA==');

@$core.Deprecated('Use listFormSubmissionsRequestDescriptor instead')
const ListFormSubmissionsRequest$json = {
  '1': 'ListFormSubmissionsRequest',
  '2': [
    {'1': 'form_id', '3': 1, '4': 1, '5': 9, '10': 'formId'},
  ],
};

/// Descriptor for `ListFormSubmissionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listFormSubmissionsRequestDescriptor = $convert.base64Decode(
    'ChpMaXN0Rm9ybVN1Ym1pc3Npb25zUmVxdWVzdBIXCgdmb3JtX2lkGAEgASgJUgZmb3JtSWQ=');

@$core.Deprecated('Use listFormSubmissionsResponseDescriptor instead')
const ListFormSubmissionsResponse$json = {
  '1': 'ListFormSubmissionsResponse',
  '2': [
    {'1': 'form_submissions', '3': 1, '4': 3, '5': 11, '6': '.backend.admin.v1.FormSubmission', '10': 'formSubmissions'},
  ],
};

/// Descriptor for `ListFormSubmissionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listFormSubmissionsResponseDescriptor = $convert.base64Decode(
    'ChtMaXN0Rm9ybVN1Ym1pc3Npb25zUmVzcG9uc2USSwoQZm9ybV9zdWJtaXNzaW9ucxgBIAMoCz'
    'IgLmJhY2tlbmQuYWRtaW4udjEuRm9ybVN1Ym1pc3Npb25SD2Zvcm1TdWJtaXNzaW9ucw==');

@$core.Deprecated('Use getFormSubmissionDetailsRequestDescriptor instead')
const GetFormSubmissionDetailsRequest$json = {
  '1': 'GetFormSubmissionDetailsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetFormSubmissionDetailsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getFormSubmissionDetailsRequestDescriptor = $convert.base64Decode(
    'Ch9HZXRGb3JtU3VibWlzc2lvbkRldGFpbHNSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use formSubmissionDetailsDescriptor instead')
const FormSubmissionDetails$json = {
  '1': 'FormSubmissionDetails',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'wallet_id', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'walletId', '17': true},
    {'1': 'form_id', '3': 3, '4': 1, '5': 9, '10': 'formId'},
    {'1': 'data', '3': 4, '4': 1, '5': 9, '10': 'data'},
    {'1': 'timestamp', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '10': 'timestamp'},
  ],
  '8': [
    {'1': '_wallet_id'},
  ],
};

/// Descriptor for `FormSubmissionDetails`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List formSubmissionDetailsDescriptor = $convert.base64Decode(
    'ChVGb3JtU3VibWlzc2lvbkRldGFpbHMSDgoCaWQYASABKAlSAmlkEiAKCXdhbGxldF9pZBgCIA'
    'EoCUgAUgh3YWxsZXRJZIgBARIXCgdmb3JtX2lkGAMgASgJUgZmb3JtSWQSEgoEZGF0YRgEIAEo'
    'CVIEZGF0YRI4Cgl0aW1lc3RhbXAYBSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUg'
    'l0aW1lc3RhbXBCDAoKX3dhbGxldF9pZA==');

@$core.Deprecated('Use setWalletXagoBalanceEnabledRequestDescriptor instead')
const SetWalletXagoBalanceEnabledRequest$json = {
  '1': 'SetWalletXagoBalanceEnabledRequest',
  '2': [
    {'1': 'wallet_id', '3': 1, '4': 1, '5': 9, '10': 'walletId'},
  ],
};

/// Descriptor for `SetWalletXagoBalanceEnabledRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setWalletXagoBalanceEnabledRequestDescriptor = $convert.base64Decode(
    'CiJTZXRXYWxsZXRYYWdvQmFsYW5jZUVuYWJsZWRSZXF1ZXN0EhsKCXdhbGxldF9pZBgBIAEoCV'
    'IId2FsbGV0SWQ=');

@$core.Deprecated('Use getWalletXagoBalanceRequestDescriptor instead')
const GetWalletXagoBalanceRequest$json = {
  '1': 'GetWalletXagoBalanceRequest',
  '2': [
    {'1': 'wallet_id', '3': 1, '4': 1, '5': 9, '10': 'walletId'},
  ],
};

/// Descriptor for `GetWalletXagoBalanceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWalletXagoBalanceRequestDescriptor = $convert.base64Decode(
    'ChtHZXRXYWxsZXRYYWdvQmFsYW5jZVJlcXVlc3QSGwoJd2FsbGV0X2lkGAEgASgJUgh3YWxsZX'
    'RJZA==');

@$core.Deprecated('Use getWalletXagoBalanceResponseDescriptor instead')
const GetWalletXagoBalanceResponse$json = {
  '1': 'GetWalletXagoBalanceResponse',
  '2': [
    {'1': 'balance', '3': 1, '4': 1, '5': 11, '6': '.backend.admin.v1.Amount', '10': 'balance'},
    {'1': 'available', '3': 2, '4': 1, '5': 11, '6': '.backend.admin.v1.Amount', '10': 'available'},
  ],
};

/// Descriptor for `GetWalletXagoBalanceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWalletXagoBalanceResponseDescriptor = $convert.base64Decode(
    'ChxHZXRXYWxsZXRYYWdvQmFsYW5jZVJlc3BvbnNlEjIKB2JhbGFuY2UYASABKAsyGC5iYWNrZW'
    '5kLmFkbWluLnYxLkFtb3VudFIHYmFsYW5jZRI2CglhdmFpbGFibGUYAiABKAsyGC5iYWNrZW5k'
    'LmFkbWluLnYxLkFtb3VudFIJYXZhaWxhYmxl');

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

