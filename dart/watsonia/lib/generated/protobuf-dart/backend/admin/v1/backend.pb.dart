//
//  Generated code. Do not modify.
//  source: backend/admin/v1/backend.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../../google/protobuf/timestamp.pb.dart' as $6;

class SetWalletCountryRequest extends $pb.GeneratedMessage {
  factory SetWalletCountryRequest({
    $core.String? id,
    $core.String? countryCode,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (countryCode != null) {
      $result.countryCode = countryCode;
    }
    return $result;
  }
  SetWalletCountryRequest._() : super();
  factory SetWalletCountryRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetWalletCountryRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetWalletCountryRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'countryCode', protoName: 'countryCode')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetWalletCountryRequest clone() => SetWalletCountryRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetWalletCountryRequest copyWith(void Function(SetWalletCountryRequest) updates) => super.copyWith((message) => updates(message as SetWalletCountryRequest)) as SetWalletCountryRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetWalletCountryRequest create() => SetWalletCountryRequest._();
  SetWalletCountryRequest createEmptyInstance() => create();
  static $pb.PbList<SetWalletCountryRequest> createRepeated() => $pb.PbList<SetWalletCountryRequest>();
  @$core.pragma('dart2js:noInline')
  static SetWalletCountryRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetWalletCountryRequest>(create);
  static SetWalletCountryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get countryCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set countryCode($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCountryCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCountryCode() => clearField(2);
}

class Country extends $pb.GeneratedMessage {
  factory Country({
    $core.String? code,
    $core.String? name,
    $core.String? numeric,
    $core.bool? supported,
  }) {
    final $result = create();
    if (code != null) {
      $result.code = code;
    }
    if (name != null) {
      $result.name = name;
    }
    if (numeric != null) {
      $result.numeric = numeric;
    }
    if (supported != null) {
      $result.supported = supported;
    }
    return $result;
  }
  Country._() : super();
  factory Country.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Country.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Country', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'numeric')
    ..aOB(4, _omitFieldNames ? '' : 'supported')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Country clone() => Country()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Country copyWith(void Function(Country) updates) => super.copyWith((message) => updates(message as Country)) as Country;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Country create() => Country._();
  Country createEmptyInstance() => create();
  static $pb.PbList<Country> createRepeated() => $pb.PbList<Country>();
  @$core.pragma('dart2js:noInline')
  static Country getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Country>(create);
  static Country? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get code => $_getSZ(0);
  @$pb.TagNumber(1)
  set code($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get numeric => $_getSZ(2);
  @$pb.TagNumber(3)
  set numeric($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasNumeric() => $_has(2);
  @$pb.TagNumber(3)
  void clearNumeric() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get supported => $_getBF(3);
  @$pb.TagNumber(4)
  set supported($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSupported() => $_has(3);
  @$pb.TagNumber(4)
  void clearSupported() => clearField(4);
}

class ListCountriesResponse extends $pb.GeneratedMessage {
  factory ListCountriesResponse({
    $core.Iterable<Country>? countries,
  }) {
    final $result = create();
    if (countries != null) {
      $result.countries.addAll(countries);
    }
    return $result;
  }
  ListCountriesResponse._() : super();
  factory ListCountriesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListCountriesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListCountriesResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<Country>(1, _omitFieldNames ? '' : 'countries', $pb.PbFieldType.PM, subBuilder: Country.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListCountriesResponse clone() => ListCountriesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListCountriesResponse copyWith(void Function(ListCountriesResponse) updates) => super.copyWith((message) => updates(message as ListCountriesResponse)) as ListCountriesResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListCountriesResponse create() => ListCountriesResponse._();
  ListCountriesResponse createEmptyInstance() => create();
  static $pb.PbList<ListCountriesResponse> createRepeated() => $pb.PbList<ListCountriesResponse>();
  @$core.pragma('dart2js:noInline')
  static ListCountriesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListCountriesResponse>(create);
  static ListCountriesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Country> get countries => $_getList(0);
}

class ListPaymentsAwaitingSignalResponse extends $pb.GeneratedMessage {
  factory ListPaymentsAwaitingSignalResponse({
    $core.Iterable<Payment>? payments,
  }) {
    final $result = create();
    if (payments != null) {
      $result.payments.addAll(payments);
    }
    return $result;
  }
  ListPaymentsAwaitingSignalResponse._() : super();
  factory ListPaymentsAwaitingSignalResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListPaymentsAwaitingSignalResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListPaymentsAwaitingSignalResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<Payment>(1, _omitFieldNames ? '' : 'payments', $pb.PbFieldType.PM, subBuilder: Payment.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListPaymentsAwaitingSignalResponse clone() => ListPaymentsAwaitingSignalResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListPaymentsAwaitingSignalResponse copyWith(void Function(ListPaymentsAwaitingSignalResponse) updates) => super.copyWith((message) => updates(message as ListPaymentsAwaitingSignalResponse)) as ListPaymentsAwaitingSignalResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListPaymentsAwaitingSignalResponse create() => ListPaymentsAwaitingSignalResponse._();
  ListPaymentsAwaitingSignalResponse createEmptyInstance() => create();
  static $pb.PbList<ListPaymentsAwaitingSignalResponse> createRepeated() => $pb.PbList<ListPaymentsAwaitingSignalResponse>();
  @$core.pragma('dart2js:noInline')
  static ListPaymentsAwaitingSignalResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListPaymentsAwaitingSignalResponse>(create);
  static ListPaymentsAwaitingSignalResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Payment> get payments => $_getList(0);
}

class Payment extends $pb.GeneratedMessage {
  factory Payment({
    $core.String? id,
    $core.String? publicID,
    $core.String? state,
    $core.String? receiverWalletUrl,
    $core.String? receiverIdentity,
    $core.String? receiverIdentityType,
    $core.String? senderAmount,
    $core.String? senderAccount,
    $core.String? note,
    $core.Iterable<$core.String>? requiredActions,
    $core.String? senderWalletUrl,
    $6.Timestamp? updatedAt,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (publicID != null) {
      $result.publicID = publicID;
    }
    if (state != null) {
      $result.state = state;
    }
    if (receiverWalletUrl != null) {
      $result.receiverWalletUrl = receiverWalletUrl;
    }
    if (receiverIdentity != null) {
      $result.receiverIdentity = receiverIdentity;
    }
    if (receiverIdentityType != null) {
      $result.receiverIdentityType = receiverIdentityType;
    }
    if (senderAmount != null) {
      $result.senderAmount = senderAmount;
    }
    if (senderAccount != null) {
      $result.senderAccount = senderAccount;
    }
    if (note != null) {
      $result.note = note;
    }
    if (requiredActions != null) {
      $result.requiredActions.addAll(requiredActions);
    }
    if (senderWalletUrl != null) {
      $result.senderWalletUrl = senderWalletUrl;
    }
    if (updatedAt != null) {
      $result.updatedAt = updatedAt;
    }
    return $result;
  }
  Payment._() : super();
  factory Payment.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Payment.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Payment', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'publicID', protoName: 'publicID')
    ..aOS(3, _omitFieldNames ? '' : 'state')
    ..aOS(4, _omitFieldNames ? '' : 'receiverWalletUrl', protoName: 'receiverWalletUrl')
    ..aOS(5, _omitFieldNames ? '' : 'receiverIdentity', protoName: 'receiverIdentity')
    ..aOS(6, _omitFieldNames ? '' : 'receiverIdentityType', protoName: 'receiverIdentityType')
    ..aOS(7, _omitFieldNames ? '' : 'senderAmount', protoName: 'senderAmount')
    ..aOS(8, _omitFieldNames ? '' : 'senderAccount', protoName: 'senderAccount')
    ..aOS(9, _omitFieldNames ? '' : 'note')
    ..pPS(10, _omitFieldNames ? '' : 'requiredActions', protoName: 'requiredActions')
    ..aOS(11, _omitFieldNames ? '' : 'senderWalletUrl', protoName: 'senderWalletUrl')
    ..aOM<$6.Timestamp>(12, _omitFieldNames ? '' : 'updatedAt', protoName: 'updatedAt', subBuilder: $6.Timestamp.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Payment clone() => Payment()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Payment copyWith(void Function(Payment) updates) => super.copyWith((message) => updates(message as Payment)) as Payment;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Payment create() => Payment._();
  Payment createEmptyInstance() => create();
  static $pb.PbList<Payment> createRepeated() => $pb.PbList<Payment>();
  @$core.pragma('dart2js:noInline')
  static Payment getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Payment>(create);
  static Payment? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get publicID => $_getSZ(1);
  @$pb.TagNumber(2)
  set publicID($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPublicID() => $_has(1);
  @$pb.TagNumber(2)
  void clearPublicID() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get state => $_getSZ(2);
  @$pb.TagNumber(3)
  set state($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasState() => $_has(2);
  @$pb.TagNumber(3)
  void clearState() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get receiverWalletUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set receiverWalletUrl($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasReceiverWalletUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearReceiverWalletUrl() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get receiverIdentity => $_getSZ(4);
  @$pb.TagNumber(5)
  set receiverIdentity($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasReceiverIdentity() => $_has(4);
  @$pb.TagNumber(5)
  void clearReceiverIdentity() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get receiverIdentityType => $_getSZ(5);
  @$pb.TagNumber(6)
  set receiverIdentityType($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasReceiverIdentityType() => $_has(5);
  @$pb.TagNumber(6)
  void clearReceiverIdentityType() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get senderAmount => $_getSZ(6);
  @$pb.TagNumber(7)
  set senderAmount($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasSenderAmount() => $_has(6);
  @$pb.TagNumber(7)
  void clearSenderAmount() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get senderAccount => $_getSZ(7);
  @$pb.TagNumber(8)
  set senderAccount($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasSenderAccount() => $_has(7);
  @$pb.TagNumber(8)
  void clearSenderAccount() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get note => $_getSZ(8);
  @$pb.TagNumber(9)
  set note($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasNote() => $_has(8);
  @$pb.TagNumber(9)
  void clearNote() => clearField(9);

  @$pb.TagNumber(10)
  $core.List<$core.String> get requiredActions => $_getList(9);

  @$pb.TagNumber(11)
  $core.String get senderWalletUrl => $_getSZ(10);
  @$pb.TagNumber(11)
  set senderWalletUrl($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasSenderWalletUrl() => $_has(10);
  @$pb.TagNumber(11)
  void clearSenderWalletUrl() => clearField(11);

  @$pb.TagNumber(12)
  $6.Timestamp get updatedAt => $_getN(11);
  @$pb.TagNumber(12)
  set updatedAt($6.Timestamp v) { setField(12, v); }
  @$pb.TagNumber(12)
  $core.bool hasUpdatedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearUpdatedAt() => clearField(12);
  @$pb.TagNumber(12)
  $6.Timestamp ensureUpdatedAt() => $_ensure(11);
}

class ListExternalApiCallsRequest extends $pb.GeneratedMessage {
  factory ListExternalApiCallsRequest({
    $core.String? paymentId,
  }) {
    final $result = create();
    if (paymentId != null) {
      $result.paymentId = paymentId;
    }
    return $result;
  }
  ListExternalApiCallsRequest._() : super();
  factory ListExternalApiCallsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListExternalApiCallsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListExternalApiCallsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'paymentId', protoName: 'paymentId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListExternalApiCallsRequest clone() => ListExternalApiCallsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListExternalApiCallsRequest copyWith(void Function(ListExternalApiCallsRequest) updates) => super.copyWith((message) => updates(message as ListExternalApiCallsRequest)) as ListExternalApiCallsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListExternalApiCallsRequest create() => ListExternalApiCallsRequest._();
  ListExternalApiCallsRequest createEmptyInstance() => create();
  static $pb.PbList<ListExternalApiCallsRequest> createRepeated() => $pb.PbList<ListExternalApiCallsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListExternalApiCallsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListExternalApiCallsRequest>(create);
  static ListExternalApiCallsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get paymentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set paymentId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPaymentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPaymentId() => clearField(1);
}

class ExternalApiCall extends $pb.GeneratedMessage {
  factory ExternalApiCall({
    $core.String? id,
    $core.String? provider,
    $core.String? context,
    $core.String? method,
    $core.String? requestBody,
    $core.String? requestPath,
    $core.String? responseBody,
    $core.String? responseStatus,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (provider != null) {
      $result.provider = provider;
    }
    if (context != null) {
      $result.context = context;
    }
    if (method != null) {
      $result.method = method;
    }
    if (requestBody != null) {
      $result.requestBody = requestBody;
    }
    if (requestPath != null) {
      $result.requestPath = requestPath;
    }
    if (responseBody != null) {
      $result.responseBody = responseBody;
    }
    if (responseStatus != null) {
      $result.responseStatus = responseStatus;
    }
    return $result;
  }
  ExternalApiCall._() : super();
  factory ExternalApiCall.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ExternalApiCall.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ExternalApiCall', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'provider')
    ..aOS(3, _omitFieldNames ? '' : 'context')
    ..aOS(4, _omitFieldNames ? '' : 'method')
    ..aOS(5, _omitFieldNames ? '' : 'requestBody', protoName: 'requestBody')
    ..aOS(6, _omitFieldNames ? '' : 'requestPath', protoName: 'requestPath')
    ..aOS(7, _omitFieldNames ? '' : 'responseBody', protoName: 'responseBody')
    ..aOS(8, _omitFieldNames ? '' : 'responseStatus', protoName: 'responseStatus')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ExternalApiCall clone() => ExternalApiCall()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ExternalApiCall copyWith(void Function(ExternalApiCall) updates) => super.copyWith((message) => updates(message as ExternalApiCall)) as ExternalApiCall;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ExternalApiCall create() => ExternalApiCall._();
  ExternalApiCall createEmptyInstance() => create();
  static $pb.PbList<ExternalApiCall> createRepeated() => $pb.PbList<ExternalApiCall>();
  @$core.pragma('dart2js:noInline')
  static ExternalApiCall getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ExternalApiCall>(create);
  static ExternalApiCall? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get provider => $_getSZ(1);
  @$pb.TagNumber(2)
  set provider($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasProvider() => $_has(1);
  @$pb.TagNumber(2)
  void clearProvider() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get context => $_getSZ(2);
  @$pb.TagNumber(3)
  set context($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasContext() => $_has(2);
  @$pb.TagNumber(3)
  void clearContext() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get method => $_getSZ(3);
  @$pb.TagNumber(4)
  set method($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasMethod() => $_has(3);
  @$pb.TagNumber(4)
  void clearMethod() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get requestBody => $_getSZ(4);
  @$pb.TagNumber(5)
  set requestBody($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasRequestBody() => $_has(4);
  @$pb.TagNumber(5)
  void clearRequestBody() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get requestPath => $_getSZ(5);
  @$pb.TagNumber(6)
  set requestPath($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasRequestPath() => $_has(5);
  @$pb.TagNumber(6)
  void clearRequestPath() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get responseBody => $_getSZ(6);
  @$pb.TagNumber(7)
  set responseBody($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasResponseBody() => $_has(6);
  @$pb.TagNumber(7)
  void clearResponseBody() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get responseStatus => $_getSZ(7);
  @$pb.TagNumber(8)
  set responseStatus($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasResponseStatus() => $_has(7);
  @$pb.TagNumber(8)
  void clearResponseStatus() => clearField(8);
}

class ListExternalApiCallsResponse extends $pb.GeneratedMessage {
  factory ListExternalApiCallsResponse({
    $core.Iterable<ExternalApiCall>? list,
  }) {
    final $result = create();
    if (list != null) {
      $result.list.addAll(list);
    }
    return $result;
  }
  ListExternalApiCallsResponse._() : super();
  factory ListExternalApiCallsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListExternalApiCallsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListExternalApiCallsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<ExternalApiCall>(1, _omitFieldNames ? '' : 'list', $pb.PbFieldType.PM, subBuilder: ExternalApiCall.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListExternalApiCallsResponse clone() => ListExternalApiCallsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListExternalApiCallsResponse copyWith(void Function(ListExternalApiCallsResponse) updates) => super.copyWith((message) => updates(message as ListExternalApiCallsResponse)) as ListExternalApiCallsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListExternalApiCallsResponse create() => ListExternalApiCallsResponse._();
  ListExternalApiCallsResponse createEmptyInstance() => create();
  static $pb.PbList<ListExternalApiCallsResponse> createRepeated() => $pb.PbList<ListExternalApiCallsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListExternalApiCallsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListExternalApiCallsResponse>(create);
  static ListExternalApiCallsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<ExternalApiCall> get list => $_getList(0);
}

class Empty extends $pb.GeneratedMessage {
  factory Empty() => create();
  Empty._() : super();
  factory Empty.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Empty.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Empty', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Empty clone() => Empty()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Empty copyWith(void Function(Empty) updates) => super.copyWith((message) => updates(message as Empty)) as Empty;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Empty create() => Empty._();
  Empty createEmptyInstance() => create();
  static $pb.PbList<Empty> createRepeated() => $pb.PbList<Empty>();
  @$core.pragma('dart2js:noInline')
  static Empty getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Empty>(create);
  static Empty? _defaultInstance;
}

class LinkedAccountReview extends $pb.GeneratedMessage {
  factory LinkedAccountReview({
    $core.String? id,
    $core.String? state,
    $core.String? newState,
    $core.String? linkedAccountID,
    $core.String? reviewedBy,
    $core.String? reason,
    $6.Timestamp? createdAt,
    $6.Timestamp? completedAt,
    $core.String? walletID,
    $core.String? walletName,
    $core.String? mask,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (state != null) {
      $result.state = state;
    }
    if (newState != null) {
      $result.newState = newState;
    }
    if (linkedAccountID != null) {
      $result.linkedAccountID = linkedAccountID;
    }
    if (reviewedBy != null) {
      $result.reviewedBy = reviewedBy;
    }
    if (reason != null) {
      $result.reason = reason;
    }
    if (createdAt != null) {
      $result.createdAt = createdAt;
    }
    if (completedAt != null) {
      $result.completedAt = completedAt;
    }
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (walletName != null) {
      $result.walletName = walletName;
    }
    if (mask != null) {
      $result.mask = mask;
    }
    return $result;
  }
  LinkedAccountReview._() : super();
  factory LinkedAccountReview.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LinkedAccountReview.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LinkedAccountReview', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'state')
    ..aOS(3, _omitFieldNames ? '' : 'newState', protoName: 'newState')
    ..aOS(4, _omitFieldNames ? '' : 'linkedAccountID', protoName: 'linkedAccountID')
    ..aOS(5, _omitFieldNames ? '' : 'reviewedBy', protoName: 'reviewedBy')
    ..aOS(6, _omitFieldNames ? '' : 'reason')
    ..aOM<$6.Timestamp>(7, _omitFieldNames ? '' : 'createdAt', protoName: 'createdAt', subBuilder: $6.Timestamp.create)
    ..aOM<$6.Timestamp>(8, _omitFieldNames ? '' : 'completedAt', protoName: 'completedAt', subBuilder: $6.Timestamp.create)
    ..aOS(9, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(10, _omitFieldNames ? '' : 'walletName', protoName: 'walletName')
    ..aOS(11, _omitFieldNames ? '' : 'mask')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LinkedAccountReview clone() => LinkedAccountReview()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LinkedAccountReview copyWith(void Function(LinkedAccountReview) updates) => super.copyWith((message) => updates(message as LinkedAccountReview)) as LinkedAccountReview;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LinkedAccountReview create() => LinkedAccountReview._();
  LinkedAccountReview createEmptyInstance() => create();
  static $pb.PbList<LinkedAccountReview> createRepeated() => $pb.PbList<LinkedAccountReview>();
  @$core.pragma('dart2js:noInline')
  static LinkedAccountReview getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LinkedAccountReview>(create);
  static LinkedAccountReview? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get state => $_getSZ(1);
  @$pb.TagNumber(2)
  set state($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasState() => $_has(1);
  @$pb.TagNumber(2)
  void clearState() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get newState => $_getSZ(2);
  @$pb.TagNumber(3)
  set newState($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasNewState() => $_has(2);
  @$pb.TagNumber(3)
  void clearNewState() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get linkedAccountID => $_getSZ(3);
  @$pb.TagNumber(4)
  set linkedAccountID($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasLinkedAccountID() => $_has(3);
  @$pb.TagNumber(4)
  void clearLinkedAccountID() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get reviewedBy => $_getSZ(4);
  @$pb.TagNumber(5)
  set reviewedBy($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasReviewedBy() => $_has(4);
  @$pb.TagNumber(5)
  void clearReviewedBy() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get reason => $_getSZ(5);
  @$pb.TagNumber(6)
  set reason($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasReason() => $_has(5);
  @$pb.TagNumber(6)
  void clearReason() => clearField(6);

  @$pb.TagNumber(7)
  $6.Timestamp get createdAt => $_getN(6);
  @$pb.TagNumber(7)
  set createdAt($6.Timestamp v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasCreatedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearCreatedAt() => clearField(7);
  @$pb.TagNumber(7)
  $6.Timestamp ensureCreatedAt() => $_ensure(6);

  @$pb.TagNumber(8)
  $6.Timestamp get completedAt => $_getN(7);
  @$pb.TagNumber(8)
  set completedAt($6.Timestamp v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasCompletedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearCompletedAt() => clearField(8);
  @$pb.TagNumber(8)
  $6.Timestamp ensureCompletedAt() => $_ensure(7);

  @$pb.TagNumber(9)
  $core.String get walletID => $_getSZ(8);
  @$pb.TagNumber(9)
  set walletID($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasWalletID() => $_has(8);
  @$pb.TagNumber(9)
  void clearWalletID() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get walletName => $_getSZ(9);
  @$pb.TagNumber(10)
  set walletName($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasWalletName() => $_has(9);
  @$pb.TagNumber(10)
  void clearWalletName() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get mask => $_getSZ(10);
  @$pb.TagNumber(11)
  set mask($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasMask() => $_has(10);
  @$pb.TagNumber(11)
  void clearMask() => clearField(11);
}

class LinkedAccountReviews extends $pb.GeneratedMessage {
  factory LinkedAccountReviews({
    $core.Iterable<LinkedAccountReview>? reviews,
  }) {
    final $result = create();
    if (reviews != null) {
      $result.reviews.addAll(reviews);
    }
    return $result;
  }
  LinkedAccountReviews._() : super();
  factory LinkedAccountReviews.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LinkedAccountReviews.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LinkedAccountReviews', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<LinkedAccountReview>(1, _omitFieldNames ? '' : 'reviews', $pb.PbFieldType.PM, subBuilder: LinkedAccountReview.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LinkedAccountReviews clone() => LinkedAccountReviews()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LinkedAccountReviews copyWith(void Function(LinkedAccountReviews) updates) => super.copyWith((message) => updates(message as LinkedAccountReviews)) as LinkedAccountReviews;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LinkedAccountReviews create() => LinkedAccountReviews._();
  LinkedAccountReviews createEmptyInstance() => create();
  static $pb.PbList<LinkedAccountReviews> createRepeated() => $pb.PbList<LinkedAccountReviews>();
  @$core.pragma('dart2js:noInline')
  static LinkedAccountReviews getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LinkedAccountReviews>(create);
  static LinkedAccountReviews? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<LinkedAccountReview> get reviews => $_getList(0);
}

class GetLinkedAccountReviewRequest extends $pb.GeneratedMessage {
  factory GetLinkedAccountReviewRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetLinkedAccountReviewRequest._() : super();
  factory GetLinkedAccountReviewRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLinkedAccountReviewRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLinkedAccountReviewRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLinkedAccountReviewRequest clone() => GetLinkedAccountReviewRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLinkedAccountReviewRequest copyWith(void Function(GetLinkedAccountReviewRequest) updates) => super.copyWith((message) => updates(message as GetLinkedAccountReviewRequest)) as GetLinkedAccountReviewRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountReviewRequest create() => GetLinkedAccountReviewRequest._();
  GetLinkedAccountReviewRequest createEmptyInstance() => create();
  static $pb.PbList<GetLinkedAccountReviewRequest> createRepeated() => $pb.PbList<GetLinkedAccountReviewRequest>();
  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountReviewRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetLinkedAccountReviewRequest>(create);
  static GetLinkedAccountReviewRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class CompleteLinkedAccountReviewRequest extends $pb.GeneratedMessage {
  factory CompleteLinkedAccountReviewRequest({
    $core.String? id,
    $core.String? reason,
    $core.String? newState,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (reason != null) {
      $result.reason = reason;
    }
    if (newState != null) {
      $result.newState = newState;
    }
    return $result;
  }
  CompleteLinkedAccountReviewRequest._() : super();
  factory CompleteLinkedAccountReviewRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CompleteLinkedAccountReviewRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CompleteLinkedAccountReviewRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'reason')
    ..aOS(3, _omitFieldNames ? '' : 'newState', protoName: 'newState')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CompleteLinkedAccountReviewRequest clone() => CompleteLinkedAccountReviewRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CompleteLinkedAccountReviewRequest copyWith(void Function(CompleteLinkedAccountReviewRequest) updates) => super.copyWith((message) => updates(message as CompleteLinkedAccountReviewRequest)) as CompleteLinkedAccountReviewRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CompleteLinkedAccountReviewRequest create() => CompleteLinkedAccountReviewRequest._();
  CompleteLinkedAccountReviewRequest createEmptyInstance() => create();
  static $pb.PbList<CompleteLinkedAccountReviewRequest> createRepeated() => $pb.PbList<CompleteLinkedAccountReviewRequest>();
  @$core.pragma('dart2js:noInline')
  static CompleteLinkedAccountReviewRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CompleteLinkedAccountReviewRequest>(create);
  static CompleteLinkedAccountReviewRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get reason => $_getSZ(1);
  @$pb.TagNumber(2)
  set reason($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasReason() => $_has(1);
  @$pb.TagNumber(2)
  void clearReason() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get newState => $_getSZ(2);
  @$pb.TagNumber(3)
  set newState($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasNewState() => $_has(2);
  @$pb.TagNumber(3)
  void clearNewState() => clearField(3);
}

class GetWalletFeaturesRequest extends $pb.GeneratedMessage {
  factory GetWalletFeaturesRequest({
    $core.String? walletID,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    return $result;
  }
  GetWalletFeaturesRequest._() : super();
  factory GetWalletFeaturesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetWalletFeaturesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetWalletFeaturesRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetWalletFeaturesRequest clone() => GetWalletFeaturesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetWalletFeaturesRequest copyWith(void Function(GetWalletFeaturesRequest) updates) => super.copyWith((message) => updates(message as GetWalletFeaturesRequest)) as GetWalletFeaturesRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetWalletFeaturesRequest create() => GetWalletFeaturesRequest._();
  GetWalletFeaturesRequest createEmptyInstance() => create();
  static $pb.PbList<GetWalletFeaturesRequest> createRepeated() => $pb.PbList<GetWalletFeaturesRequest>();
  @$core.pragma('dart2js:noInline')
  static GetWalletFeaturesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetWalletFeaturesRequest>(create);
  static GetWalletFeaturesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);
}

class Features extends $pb.GeneratedMessage {
  factory Features({
    $core.bool? sendEnabled,
    $core.bool? receiveEnabled,
    $core.bool? linkedAccountsEnabled,
    $core.bool? cardsEnabled,
    $core.bool? banksEnabled,
    $core.bool? identitiesEnabled,
    $core.bool? twitterEnabled,
    $core.String? walletID,
    $core.bool? addCardsEnabled,
    $core.bool? zarBalanceEnabled,
  }) {
    final $result = create();
    if (sendEnabled != null) {
      $result.sendEnabled = sendEnabled;
    }
    if (receiveEnabled != null) {
      $result.receiveEnabled = receiveEnabled;
    }
    if (linkedAccountsEnabled != null) {
      $result.linkedAccountsEnabled = linkedAccountsEnabled;
    }
    if (cardsEnabled != null) {
      $result.cardsEnabled = cardsEnabled;
    }
    if (banksEnabled != null) {
      $result.banksEnabled = banksEnabled;
    }
    if (identitiesEnabled != null) {
      $result.identitiesEnabled = identitiesEnabled;
    }
    if (twitterEnabled != null) {
      $result.twitterEnabled = twitterEnabled;
    }
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (addCardsEnabled != null) {
      $result.addCardsEnabled = addCardsEnabled;
    }
    if (zarBalanceEnabled != null) {
      $result.zarBalanceEnabled = zarBalanceEnabled;
    }
    return $result;
  }
  Features._() : super();
  factory Features.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Features.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Features', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'sendEnabled', protoName: 'sendEnabled')
    ..aOB(2, _omitFieldNames ? '' : 'receiveEnabled', protoName: 'receiveEnabled')
    ..aOB(3, _omitFieldNames ? '' : 'linkedAccountsEnabled', protoName: 'linkedAccountsEnabled')
    ..aOB(4, _omitFieldNames ? '' : 'cardsEnabled', protoName: 'cardsEnabled')
    ..aOB(5, _omitFieldNames ? '' : 'banksEnabled', protoName: 'banksEnabled')
    ..aOB(6, _omitFieldNames ? '' : 'identitiesEnabled', protoName: 'identitiesEnabled')
    ..aOB(7, _omitFieldNames ? '' : 'twitterEnabled', protoName: 'twitterEnabled')
    ..aOS(8, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOB(9, _omitFieldNames ? '' : 'addCardsEnabled', protoName: 'addCardsEnabled')
    ..aOB(10, _omitFieldNames ? '' : 'zarBalanceEnabled', protoName: 'zarBalanceEnabled')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Features clone() => Features()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Features copyWith(void Function(Features) updates) => super.copyWith((message) => updates(message as Features)) as Features;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Features create() => Features._();
  Features createEmptyInstance() => create();
  static $pb.PbList<Features> createRepeated() => $pb.PbList<Features>();
  @$core.pragma('dart2js:noInline')
  static Features getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Features>(create);
  static Features? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get sendEnabled => $_getBF(0);
  @$pb.TagNumber(1)
  set sendEnabled($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSendEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearSendEnabled() => clearField(1);

  @$pb.TagNumber(2)
  $core.bool get receiveEnabled => $_getBF(1);
  @$pb.TagNumber(2)
  set receiveEnabled($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasReceiveEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearReceiveEnabled() => clearField(2);

  @$pb.TagNumber(3)
  $core.bool get linkedAccountsEnabled => $_getBF(2);
  @$pb.TagNumber(3)
  set linkedAccountsEnabled($core.bool v) { $_setBool(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasLinkedAccountsEnabled() => $_has(2);
  @$pb.TagNumber(3)
  void clearLinkedAccountsEnabled() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get cardsEnabled => $_getBF(3);
  @$pb.TagNumber(4)
  set cardsEnabled($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCardsEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearCardsEnabled() => clearField(4);

  @$pb.TagNumber(5)
  $core.bool get banksEnabled => $_getBF(4);
  @$pb.TagNumber(5)
  set banksEnabled($core.bool v) { $_setBool(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasBanksEnabled() => $_has(4);
  @$pb.TagNumber(5)
  void clearBanksEnabled() => clearField(5);

  @$pb.TagNumber(6)
  $core.bool get identitiesEnabled => $_getBF(5);
  @$pb.TagNumber(6)
  set identitiesEnabled($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasIdentitiesEnabled() => $_has(5);
  @$pb.TagNumber(6)
  void clearIdentitiesEnabled() => clearField(6);

  @$pb.TagNumber(7)
  $core.bool get twitterEnabled => $_getBF(6);
  @$pb.TagNumber(7)
  set twitterEnabled($core.bool v) { $_setBool(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasTwitterEnabled() => $_has(6);
  @$pb.TagNumber(7)
  void clearTwitterEnabled() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get walletID => $_getSZ(7);
  @$pb.TagNumber(8)
  set walletID($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasWalletID() => $_has(7);
  @$pb.TagNumber(8)
  void clearWalletID() => clearField(8);

  @$pb.TagNumber(9)
  $core.bool get addCardsEnabled => $_getBF(8);
  @$pb.TagNumber(9)
  set addCardsEnabled($core.bool v) { $_setBool(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasAddCardsEnabled() => $_has(8);
  @$pb.TagNumber(9)
  void clearAddCardsEnabled() => clearField(9);

  @$pb.TagNumber(10)
  $core.bool get zarBalanceEnabled => $_getBF(9);
  @$pb.TagNumber(10)
  set zarBalanceEnabled($core.bool v) { $_setBool(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasZarBalanceEnabled() => $_has(9);
  @$pb.TagNumber(10)
  void clearZarBalanceEnabled() => clearField(10);
}

class ListAuditRequest extends $pb.GeneratedMessage {
  factory ListAuditRequest({
    $core.String? walletID,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    return $result;
  }
  ListAuditRequest._() : super();
  factory ListAuditRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListAuditRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListAuditRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListAuditRequest clone() => ListAuditRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListAuditRequest copyWith(void Function(ListAuditRequest) updates) => super.copyWith((message) => updates(message as ListAuditRequest)) as ListAuditRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListAuditRequest create() => ListAuditRequest._();
  ListAuditRequest createEmptyInstance() => create();
  static $pb.PbList<ListAuditRequest> createRepeated() => $pb.PbList<ListAuditRequest>();
  @$core.pragma('dart2js:noInline')
  static ListAuditRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListAuditRequest>(create);
  static ListAuditRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);
}

class ListAuditResponse extends $pb.GeneratedMessage {
  factory ListAuditResponse({
    $core.Iterable<AuditOperation>? operations,
  }) {
    final $result = create();
    if (operations != null) {
      $result.operations.addAll(operations);
    }
    return $result;
  }
  ListAuditResponse._() : super();
  factory ListAuditResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListAuditResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListAuditResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<AuditOperation>(1, _omitFieldNames ? '' : 'operations', $pb.PbFieldType.PM, subBuilder: AuditOperation.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListAuditResponse clone() => ListAuditResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListAuditResponse copyWith(void Function(ListAuditResponse) updates) => super.copyWith((message) => updates(message as ListAuditResponse)) as ListAuditResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListAuditResponse create() => ListAuditResponse._();
  ListAuditResponse createEmptyInstance() => create();
  static $pb.PbList<ListAuditResponse> createRepeated() => $pb.PbList<ListAuditResponse>();
  @$core.pragma('dart2js:noInline')
  static ListAuditResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListAuditResponse>(create);
  static ListAuditResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<AuditOperation> get operations => $_getList(0);
}

class AuditOperation extends $pb.GeneratedMessage {
  factory AuditOperation({
    $core.String? adminUser,
    $core.String? walletID,
    $core.String? operation,
    $core.String? parameters,
    $6.Timestamp? timestamp,
  }) {
    final $result = create();
    if (adminUser != null) {
      $result.adminUser = adminUser;
    }
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (operation != null) {
      $result.operation = operation;
    }
    if (parameters != null) {
      $result.parameters = parameters;
    }
    if (timestamp != null) {
      $result.timestamp = timestamp;
    }
    return $result;
  }
  AuditOperation._() : super();
  factory AuditOperation.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AuditOperation.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AuditOperation', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'adminUser', protoName: 'adminUser')
    ..aOS(2, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(3, _omitFieldNames ? '' : 'operation')
    ..aOS(4, _omitFieldNames ? '' : 'parameters')
    ..aOM<$6.Timestamp>(5, _omitFieldNames ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AuditOperation clone() => AuditOperation()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AuditOperation copyWith(void Function(AuditOperation) updates) => super.copyWith((message) => updates(message as AuditOperation)) as AuditOperation;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AuditOperation create() => AuditOperation._();
  AuditOperation createEmptyInstance() => create();
  static $pb.PbList<AuditOperation> createRepeated() => $pb.PbList<AuditOperation>();
  @$core.pragma('dart2js:noInline')
  static AuditOperation getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AuditOperation>(create);
  static AuditOperation? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get adminUser => $_getSZ(0);
  @$pb.TagNumber(1)
  set adminUser($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAdminUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearAdminUser() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get walletID => $_getSZ(1);
  @$pb.TagNumber(2)
  set walletID($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasWalletID() => $_has(1);
  @$pb.TagNumber(2)
  void clearWalletID() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get operation => $_getSZ(2);
  @$pb.TagNumber(3)
  set operation($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasOperation() => $_has(2);
  @$pb.TagNumber(3)
  void clearOperation() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get parameters => $_getSZ(3);
  @$pb.TagNumber(4)
  set parameters($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasParameters() => $_has(3);
  @$pb.TagNumber(4)
  void clearParameters() => clearField(4);

  @$pb.TagNumber(5)
  $6.Timestamp get timestamp => $_getN(4);
  @$pb.TagNumber(5)
  set timestamp($6.Timestamp v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasTimestamp() => $_has(4);
  @$pb.TagNumber(5)
  void clearTimestamp() => clearField(5);
  @$pb.TagNumber(5)
  $6.Timestamp ensureTimestamp() => $_ensure(4);
}

class ListLinkedAccountsRequest extends $pb.GeneratedMessage {
  factory ListLinkedAccountsRequest({
    $core.String? walletID,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    return $result;
  }
  ListLinkedAccountsRequest._() : super();
  factory ListLinkedAccountsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListLinkedAccountsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListLinkedAccountsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListLinkedAccountsRequest clone() => ListLinkedAccountsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListLinkedAccountsRequest copyWith(void Function(ListLinkedAccountsRequest) updates) => super.copyWith((message) => updates(message as ListLinkedAccountsRequest)) as ListLinkedAccountsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListLinkedAccountsRequest create() => ListLinkedAccountsRequest._();
  ListLinkedAccountsRequest createEmptyInstance() => create();
  static $pb.PbList<ListLinkedAccountsRequest> createRepeated() => $pb.PbList<ListLinkedAccountsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListLinkedAccountsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListLinkedAccountsRequest>(create);
  static ListLinkedAccountsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);
}

class GetLinkedAccountRequest extends $pb.GeneratedMessage {
  factory GetLinkedAccountRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetLinkedAccountRequest._() : super();
  factory GetLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLinkedAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLinkedAccountRequest clone() => GetLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLinkedAccountRequest copyWith(void Function(GetLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as GetLinkedAccountRequest)) as GetLinkedAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountRequest create() => GetLinkedAccountRequest._();
  GetLinkedAccountRequest createEmptyInstance() => create();
  static $pb.PbList<GetLinkedAccountRequest> createRepeated() => $pb.PbList<GetLinkedAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetLinkedAccountRequest>(create);
  static GetLinkedAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class ListLinkedAccountsResponse extends $pb.GeneratedMessage {
  factory ListLinkedAccountsResponse({
    $core.Iterable<LinkedAccount>? accounts,
  }) {
    final $result = create();
    if (accounts != null) {
      $result.accounts.addAll(accounts);
    }
    return $result;
  }
  ListLinkedAccountsResponse._() : super();
  factory ListLinkedAccountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListLinkedAccountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListLinkedAccountsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<LinkedAccount>(1, _omitFieldNames ? '' : 'accounts', $pb.PbFieldType.PM, subBuilder: LinkedAccount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListLinkedAccountsResponse clone() => ListLinkedAccountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListLinkedAccountsResponse copyWith(void Function(ListLinkedAccountsResponse) updates) => super.copyWith((message) => updates(message as ListLinkedAccountsResponse)) as ListLinkedAccountsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListLinkedAccountsResponse create() => ListLinkedAccountsResponse._();
  ListLinkedAccountsResponse createEmptyInstance() => create();
  static $pb.PbList<ListLinkedAccountsResponse> createRepeated() => $pb.PbList<ListLinkedAccountsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListLinkedAccountsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListLinkedAccountsResponse>(create);
  static ListLinkedAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<LinkedAccount> get accounts => $_getList(0);
}

class LinkedAccount extends $pb.GeneratedMessage {
  factory LinkedAccount({
    $core.String? id,
    $core.String? walletID,
    $core.String? name,
    $core.String? nickname,
    $core.String? mask,
    $core.String? provider,
    $core.String? providerID,
    $core.String? type,
    $core.String? state,
    $core.String? canSend,
    $core.String? canReceive,
    $core.String? sendCurrencyCode,
    $core.String? sendCurrencyCountryCode,
    $core.String? receiveCurrencyCode,
    $core.String? receiveCurrencyCountryCode,
    $6.Timestamp? deletedAt,
    $core.bool? defaultSend,
    $core.bool? defaultReceive,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (name != null) {
      $result.name = name;
    }
    if (nickname != null) {
      $result.nickname = nickname;
    }
    if (mask != null) {
      $result.mask = mask;
    }
    if (provider != null) {
      $result.provider = provider;
    }
    if (providerID != null) {
      $result.providerID = providerID;
    }
    if (type != null) {
      $result.type = type;
    }
    if (state != null) {
      $result.state = state;
    }
    if (canSend != null) {
      $result.canSend = canSend;
    }
    if (canReceive != null) {
      $result.canReceive = canReceive;
    }
    if (sendCurrencyCode != null) {
      $result.sendCurrencyCode = sendCurrencyCode;
    }
    if (sendCurrencyCountryCode != null) {
      $result.sendCurrencyCountryCode = sendCurrencyCountryCode;
    }
    if (receiveCurrencyCode != null) {
      $result.receiveCurrencyCode = receiveCurrencyCode;
    }
    if (receiveCurrencyCountryCode != null) {
      $result.receiveCurrencyCountryCode = receiveCurrencyCountryCode;
    }
    if (deletedAt != null) {
      $result.deletedAt = deletedAt;
    }
    if (defaultSend != null) {
      $result.defaultSend = defaultSend;
    }
    if (defaultReceive != null) {
      $result.defaultReceive = defaultReceive;
    }
    return $result;
  }
  LinkedAccount._() : super();
  factory LinkedAccount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LinkedAccount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LinkedAccount', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'nickname')
    ..aOS(5, _omitFieldNames ? '' : 'mask')
    ..aOS(6, _omitFieldNames ? '' : 'provider')
    ..aOS(7, _omitFieldNames ? '' : 'providerID', protoName: 'providerID')
    ..aOS(8, _omitFieldNames ? '' : 'type')
    ..aOS(9, _omitFieldNames ? '' : 'state')
    ..aOS(10, _omitFieldNames ? '' : 'canSend', protoName: 'canSend')
    ..aOS(11, _omitFieldNames ? '' : 'canReceive', protoName: 'canReceive')
    ..aOS(12, _omitFieldNames ? '' : 'sendCurrencyCode', protoName: 'sendCurrencyCode')
    ..aOS(13, _omitFieldNames ? '' : 'sendCurrencyCountryCode', protoName: 'sendCurrencyCountryCode')
    ..aOS(14, _omitFieldNames ? '' : 'receiveCurrencyCode', protoName: 'receiveCurrencyCode')
    ..aOS(15, _omitFieldNames ? '' : 'receiveCurrencyCountryCode', protoName: 'receiveCurrencyCountryCode')
    ..aOM<$6.Timestamp>(16, _omitFieldNames ? '' : 'deletedAt', protoName: 'deletedAt', subBuilder: $6.Timestamp.create)
    ..aOB(17, _omitFieldNames ? '' : 'defaultSend', protoName: 'defaultSend')
    ..aOB(18, _omitFieldNames ? '' : 'defaultReceive', protoName: 'defaultReceive')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LinkedAccount clone() => LinkedAccount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LinkedAccount copyWith(void Function(LinkedAccount) updates) => super.copyWith((message) => updates(message as LinkedAccount)) as LinkedAccount;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LinkedAccount create() => LinkedAccount._();
  LinkedAccount createEmptyInstance() => create();
  static $pb.PbList<LinkedAccount> createRepeated() => $pb.PbList<LinkedAccount>();
  @$core.pragma('dart2js:noInline')
  static LinkedAccount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LinkedAccount>(create);
  static LinkedAccount? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get walletID => $_getSZ(1);
  @$pb.TagNumber(2)
  set walletID($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasWalletID() => $_has(1);
  @$pb.TagNumber(2)
  void clearWalletID() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get nickname => $_getSZ(3);
  @$pb.TagNumber(4)
  set nickname($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasNickname() => $_has(3);
  @$pb.TagNumber(4)
  void clearNickname() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get mask => $_getSZ(4);
  @$pb.TagNumber(5)
  set mask($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasMask() => $_has(4);
  @$pb.TagNumber(5)
  void clearMask() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get provider => $_getSZ(5);
  @$pb.TagNumber(6)
  set provider($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasProvider() => $_has(5);
  @$pb.TagNumber(6)
  void clearProvider() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get providerID => $_getSZ(6);
  @$pb.TagNumber(7)
  set providerID($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasProviderID() => $_has(6);
  @$pb.TagNumber(7)
  void clearProviderID() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get type => $_getSZ(7);
  @$pb.TagNumber(8)
  set type($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasType() => $_has(7);
  @$pb.TagNumber(8)
  void clearType() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get state => $_getSZ(8);
  @$pb.TagNumber(9)
  set state($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasState() => $_has(8);
  @$pb.TagNumber(9)
  void clearState() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get canSend => $_getSZ(9);
  @$pb.TagNumber(10)
  set canSend($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasCanSend() => $_has(9);
  @$pb.TagNumber(10)
  void clearCanSend() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get canReceive => $_getSZ(10);
  @$pb.TagNumber(11)
  set canReceive($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasCanReceive() => $_has(10);
  @$pb.TagNumber(11)
  void clearCanReceive() => clearField(11);

  @$pb.TagNumber(12)
  $core.String get sendCurrencyCode => $_getSZ(11);
  @$pb.TagNumber(12)
  set sendCurrencyCode($core.String v) { $_setString(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasSendCurrencyCode() => $_has(11);
  @$pb.TagNumber(12)
  void clearSendCurrencyCode() => clearField(12);

  @$pb.TagNumber(13)
  $core.String get sendCurrencyCountryCode => $_getSZ(12);
  @$pb.TagNumber(13)
  set sendCurrencyCountryCode($core.String v) { $_setString(12, v); }
  @$pb.TagNumber(13)
  $core.bool hasSendCurrencyCountryCode() => $_has(12);
  @$pb.TagNumber(13)
  void clearSendCurrencyCountryCode() => clearField(13);

  @$pb.TagNumber(14)
  $core.String get receiveCurrencyCode => $_getSZ(13);
  @$pb.TagNumber(14)
  set receiveCurrencyCode($core.String v) { $_setString(13, v); }
  @$pb.TagNumber(14)
  $core.bool hasReceiveCurrencyCode() => $_has(13);
  @$pb.TagNumber(14)
  void clearReceiveCurrencyCode() => clearField(14);

  @$pb.TagNumber(15)
  $core.String get receiveCurrencyCountryCode => $_getSZ(14);
  @$pb.TagNumber(15)
  set receiveCurrencyCountryCode($core.String v) { $_setString(14, v); }
  @$pb.TagNumber(15)
  $core.bool hasReceiveCurrencyCountryCode() => $_has(14);
  @$pb.TagNumber(15)
  void clearReceiveCurrencyCountryCode() => clearField(15);

  @$pb.TagNumber(16)
  $6.Timestamp get deletedAt => $_getN(15);
  @$pb.TagNumber(16)
  set deletedAt($6.Timestamp v) { setField(16, v); }
  @$pb.TagNumber(16)
  $core.bool hasDeletedAt() => $_has(15);
  @$pb.TagNumber(16)
  void clearDeletedAt() => clearField(16);
  @$pb.TagNumber(16)
  $6.Timestamp ensureDeletedAt() => $_ensure(15);

  @$pb.TagNumber(17)
  $core.bool get defaultSend => $_getBF(16);
  @$pb.TagNumber(17)
  set defaultSend($core.bool v) { $_setBool(16, v); }
  @$pb.TagNumber(17)
  $core.bool hasDefaultSend() => $_has(16);
  @$pb.TagNumber(17)
  void clearDefaultSend() => clearField(17);

  @$pb.TagNumber(18)
  $core.bool get defaultReceive => $_getBF(17);
  @$pb.TagNumber(18)
  set defaultReceive($core.bool v) { $_setBool(17, v); }
  @$pb.TagNumber(18)
  $core.bool hasDefaultReceive() => $_has(17);
  @$pb.TagNumber(18)
  void clearDefaultReceive() => clearField(18);
}

class GetTransactionDetailsRequest extends $pb.GeneratedMessage {
  factory GetTransactionDetailsRequest({
    $core.String? walletID,
    $core.String? transactionID,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (transactionID != null) {
      $result.transactionID = transactionID;
    }
    return $result;
  }
  GetTransactionDetailsRequest._() : super();
  factory GetTransactionDetailsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetTransactionDetailsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetTransactionDetailsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(2, _omitFieldNames ? '' : 'transactionID', protoName: 'transactionID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetTransactionDetailsRequest clone() => GetTransactionDetailsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetTransactionDetailsRequest copyWith(void Function(GetTransactionDetailsRequest) updates) => super.copyWith((message) => updates(message as GetTransactionDetailsRequest)) as GetTransactionDetailsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTransactionDetailsRequest create() => GetTransactionDetailsRequest._();
  GetTransactionDetailsRequest createEmptyInstance() => create();
  static $pb.PbList<GetTransactionDetailsRequest> createRepeated() => $pb.PbList<GetTransactionDetailsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetTransactionDetailsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTransactionDetailsRequest>(create);
  static GetTransactionDetailsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get transactionID => $_getSZ(1);
  @$pb.TagNumber(2)
  set transactionID($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasTransactionID() => $_has(1);
  @$pb.TagNumber(2)
  void clearTransactionID() => clearField(2);
}

class GetTransactionDetailsResponse extends $pb.GeneratedMessage {
  factory GetTransactionDetailsResponse({
    Transaction? transaction,
    $core.Iterable<Transfer>? transfers,
  }) {
    final $result = create();
    if (transaction != null) {
      $result.transaction = transaction;
    }
    if (transfers != null) {
      $result.transfers.addAll(transfers);
    }
    return $result;
  }
  GetTransactionDetailsResponse._() : super();
  factory GetTransactionDetailsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetTransactionDetailsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetTransactionDetailsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOM<Transaction>(1, _omitFieldNames ? '' : 'transaction', subBuilder: Transaction.create)
    ..pc<Transfer>(2, _omitFieldNames ? '' : 'transfers', $pb.PbFieldType.PM, subBuilder: Transfer.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetTransactionDetailsResponse clone() => GetTransactionDetailsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetTransactionDetailsResponse copyWith(void Function(GetTransactionDetailsResponse) updates) => super.copyWith((message) => updates(message as GetTransactionDetailsResponse)) as GetTransactionDetailsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTransactionDetailsResponse create() => GetTransactionDetailsResponse._();
  GetTransactionDetailsResponse createEmptyInstance() => create();
  static $pb.PbList<GetTransactionDetailsResponse> createRepeated() => $pb.PbList<GetTransactionDetailsResponse>();
  @$core.pragma('dart2js:noInline')
  static GetTransactionDetailsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTransactionDetailsResponse>(create);
  static GetTransactionDetailsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Transaction get transaction => $_getN(0);
  @$pb.TagNumber(1)
  set transaction(Transaction v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasTransaction() => $_has(0);
  @$pb.TagNumber(1)
  void clearTransaction() => clearField(1);
  @$pb.TagNumber(1)
  Transaction ensureTransaction() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<Transfer> get transfers => $_getList(1);
}

class Transfer extends $pb.GeneratedMessage {
  factory Transfer({
    $core.String? iD,
    $core.String? linkedAccountID,
    $core.String? linkedAccountProvider,
    $core.String? linkedAccountType,
    $core.double? amount,
    $core.String? currency,
    $core.String? state,
    $6.Timestamp? timestamp,
    $core.String? foreignID,
  }) {
    final $result = create();
    if (iD != null) {
      $result.iD = iD;
    }
    if (linkedAccountID != null) {
      $result.linkedAccountID = linkedAccountID;
    }
    if (linkedAccountProvider != null) {
      $result.linkedAccountProvider = linkedAccountProvider;
    }
    if (linkedAccountType != null) {
      $result.linkedAccountType = linkedAccountType;
    }
    if (amount != null) {
      $result.amount = amount;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (state != null) {
      $result.state = state;
    }
    if (timestamp != null) {
      $result.timestamp = timestamp;
    }
    if (foreignID != null) {
      $result.foreignID = foreignID;
    }
    return $result;
  }
  Transfer._() : super();
  factory Transfer.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transfer.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Transfer', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'ID', protoName: 'ID')
    ..aOS(2, _omitFieldNames ? '' : 'linkedAccountID', protoName: 'linkedAccountID')
    ..aOS(3, _omitFieldNames ? '' : 'linkedAccountProvider', protoName: 'linkedAccountProvider')
    ..aOS(4, _omitFieldNames ? '' : 'linkedAccountType', protoName: 'linkedAccountType')
    ..a<$core.double>(5, _omitFieldNames ? '' : 'amount', $pb.PbFieldType.OD)
    ..aOS(6, _omitFieldNames ? '' : 'currency')
    ..aOS(7, _omitFieldNames ? '' : 'state')
    ..aOM<$6.Timestamp>(8, _omitFieldNames ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..aOS(9, _omitFieldNames ? '' : 'foreignID', protoName: 'foreignID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Transfer clone() => Transfer()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Transfer copyWith(void Function(Transfer) updates) => super.copyWith((message) => updates(message as Transfer)) as Transfer;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Transfer create() => Transfer._();
  Transfer createEmptyInstance() => create();
  static $pb.PbList<Transfer> createRepeated() => $pb.PbList<Transfer>();
  @$core.pragma('dart2js:noInline')
  static Transfer getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Transfer>(create);
  static Transfer? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get iD => $_getSZ(0);
  @$pb.TagNumber(1)
  set iD($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasID() => $_has(0);
  @$pb.TagNumber(1)
  void clearID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get linkedAccountID => $_getSZ(1);
  @$pb.TagNumber(2)
  set linkedAccountID($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLinkedAccountID() => $_has(1);
  @$pb.TagNumber(2)
  void clearLinkedAccountID() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get linkedAccountProvider => $_getSZ(2);
  @$pb.TagNumber(3)
  set linkedAccountProvider($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasLinkedAccountProvider() => $_has(2);
  @$pb.TagNumber(3)
  void clearLinkedAccountProvider() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get linkedAccountType => $_getSZ(3);
  @$pb.TagNumber(4)
  set linkedAccountType($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasLinkedAccountType() => $_has(3);
  @$pb.TagNumber(4)
  void clearLinkedAccountType() => clearField(4);

  @$pb.TagNumber(5)
  $core.double get amount => $_getN(4);
  @$pb.TagNumber(5)
  set amount($core.double v) { $_setDouble(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearAmount() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get currency => $_getSZ(5);
  @$pb.TagNumber(6)
  set currency($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasCurrency() => $_has(5);
  @$pb.TagNumber(6)
  void clearCurrency() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get state => $_getSZ(6);
  @$pb.TagNumber(7)
  set state($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasState() => $_has(6);
  @$pb.TagNumber(7)
  void clearState() => clearField(7);

  @$pb.TagNumber(8)
  $6.Timestamp get timestamp => $_getN(7);
  @$pb.TagNumber(8)
  set timestamp($6.Timestamp v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasTimestamp() => $_has(7);
  @$pb.TagNumber(8)
  void clearTimestamp() => clearField(8);
  @$pb.TagNumber(8)
  $6.Timestamp ensureTimestamp() => $_ensure(7);

  @$pb.TagNumber(9)
  $core.String get foreignID => $_getSZ(8);
  @$pb.TagNumber(9)
  set foreignID($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasForeignID() => $_has(8);
  @$pb.TagNumber(9)
  void clearForeignID() => clearField(9);
}

class ListTransactionsRequest extends $pb.GeneratedMessage {
  factory ListTransactionsRequest({
    $core.String? walletID,
    PaginationRequest? page,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (page != null) {
      $result.page = page;
    }
    return $result;
  }
  ListTransactionsRequest._() : super();
  factory ListTransactionsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListTransactionsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListTransactionsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOM<PaginationRequest>(2, _omitFieldNames ? '' : 'page', subBuilder: PaginationRequest.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListTransactionsRequest clone() => ListTransactionsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListTransactionsRequest copyWith(void Function(ListTransactionsRequest) updates) => super.copyWith((message) => updates(message as ListTransactionsRequest)) as ListTransactionsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListTransactionsRequest create() => ListTransactionsRequest._();
  ListTransactionsRequest createEmptyInstance() => create();
  static $pb.PbList<ListTransactionsRequest> createRepeated() => $pb.PbList<ListTransactionsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListTransactionsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListTransactionsRequest>(create);
  static ListTransactionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);

  @$pb.TagNumber(2)
  PaginationRequest get page => $_getN(1);
  @$pb.TagNumber(2)
  set page(PaginationRequest v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPage() => $_has(1);
  @$pb.TagNumber(2)
  void clearPage() => clearField(2);
  @$pb.TagNumber(2)
  PaginationRequest ensurePage() => $_ensure(1);
}

class ListTransactionsResponse extends $pb.GeneratedMessage {
  factory ListTransactionsResponse({
    $core.Iterable<Transaction>? transactions,
    $core.String? nextPageToken,
  }) {
    final $result = create();
    if (transactions != null) {
      $result.transactions.addAll(transactions);
    }
    if (nextPageToken != null) {
      $result.nextPageToken = nextPageToken;
    }
    return $result;
  }
  ListTransactionsResponse._() : super();
  factory ListTransactionsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListTransactionsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListTransactionsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<Transaction>(1, _omitFieldNames ? '' : 'transactions', $pb.PbFieldType.PM, subBuilder: Transaction.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken', protoName: 'nextPageToken')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListTransactionsResponse clone() => ListTransactionsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListTransactionsResponse copyWith(void Function(ListTransactionsResponse) updates) => super.copyWith((message) => updates(message as ListTransactionsResponse)) as ListTransactionsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListTransactionsResponse create() => ListTransactionsResponse._();
  ListTransactionsResponse createEmptyInstance() => create();
  static $pb.PbList<ListTransactionsResponse> createRepeated() => $pb.PbList<ListTransactionsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListTransactionsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListTransactionsResponse>(create);
  static ListTransactionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Transaction> get transactions => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => clearField(2);
}

class Transaction extends $pb.GeneratedMessage {
  factory Transaction({
    $core.String? id,
    $core.String? type,
    $core.double? amount,
    $core.String? source,
    $core.String? destination,
    $6.Timestamp? timestamp,
    $core.String? asset,
    $core.String? walletID,
    $core.String? paymentId,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (type != null) {
      $result.type = type;
    }
    if (amount != null) {
      $result.amount = amount;
    }
    if (source != null) {
      $result.source = source;
    }
    if (destination != null) {
      $result.destination = destination;
    }
    if (timestamp != null) {
      $result.timestamp = timestamp;
    }
    if (asset != null) {
      $result.asset = asset;
    }
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (paymentId != null) {
      $result.paymentId = paymentId;
    }
    return $result;
  }
  Transaction._() : super();
  factory Transaction.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transaction.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Transaction', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..a<$core.double>(3, _omitFieldNames ? '' : 'amount', $pb.PbFieldType.OD)
    ..aOS(4, _omitFieldNames ? '' : 'source')
    ..aOS(5, _omitFieldNames ? '' : 'destination')
    ..aOM<$6.Timestamp>(6, _omitFieldNames ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..aOS(7, _omitFieldNames ? '' : 'asset')
    ..aOS(8, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(9, _omitFieldNames ? '' : 'paymentId', protoName: 'paymentId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Transaction clone() => Transaction()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Transaction copyWith(void Function(Transaction) updates) => super.copyWith((message) => updates(message as Transaction)) as Transaction;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Transaction create() => Transaction._();
  Transaction createEmptyInstance() => create();
  static $pb.PbList<Transaction> createRepeated() => $pb.PbList<Transaction>();
  @$core.pragma('dart2js:noInline')
  static Transaction getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Transaction>(create);
  static Transaction? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => clearField(2);

  @$pb.TagNumber(3)
  $core.double get amount => $_getN(2);
  @$pb.TagNumber(3)
  set amount($core.double v) { $_setDouble(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearAmount() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get source => $_getSZ(3);
  @$pb.TagNumber(4)
  set source($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSource() => $_has(3);
  @$pb.TagNumber(4)
  void clearSource() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get destination => $_getSZ(4);
  @$pb.TagNumber(5)
  set destination($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDestination() => $_has(4);
  @$pb.TagNumber(5)
  void clearDestination() => clearField(5);

  @$pb.TagNumber(6)
  $6.Timestamp get timestamp => $_getN(5);
  @$pb.TagNumber(6)
  set timestamp($6.Timestamp v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasTimestamp() => $_has(5);
  @$pb.TagNumber(6)
  void clearTimestamp() => clearField(6);
  @$pb.TagNumber(6)
  $6.Timestamp ensureTimestamp() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.String get asset => $_getSZ(6);
  @$pb.TagNumber(7)
  set asset($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasAsset() => $_has(6);
  @$pb.TagNumber(7)
  void clearAsset() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get walletID => $_getSZ(7);
  @$pb.TagNumber(8)
  set walletID($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasWalletID() => $_has(7);
  @$pb.TagNumber(8)
  void clearWalletID() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get paymentId => $_getSZ(8);
  @$pb.TagNumber(9)
  set paymentId($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasPaymentId() => $_has(8);
  @$pb.TagNumber(9)
  void clearPaymentId() => clearField(9);
}

class GetUserTransactionsRequest extends $pb.GeneratedMessage {
  factory GetUserTransactionsRequest({
    $core.String? userID,
  }) {
    final $result = create();
    if (userID != null) {
      $result.userID = userID;
    }
    return $result;
  }
  GetUserTransactionsRequest._() : super();
  factory GetUserTransactionsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetUserTransactionsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetUserTransactionsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userID', protoName: 'userID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetUserTransactionsRequest clone() => GetUserTransactionsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetUserTransactionsRequest copyWith(void Function(GetUserTransactionsRequest) updates) => super.copyWith((message) => updates(message as GetUserTransactionsRequest)) as GetUserTransactionsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetUserTransactionsRequest create() => GetUserTransactionsRequest._();
  GetUserTransactionsRequest createEmptyInstance() => create();
  static $pb.PbList<GetUserTransactionsRequest> createRepeated() => $pb.PbList<GetUserTransactionsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetUserTransactionsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetUserTransactionsRequest>(create);
  static GetUserTransactionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userID => $_getSZ(0);
  @$pb.TagNumber(1)
  set userID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUserID() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserID() => clearField(1);
}

class GetWalletDetailsRequest extends $pb.GeneratedMessage {
  factory GetWalletDetailsRequest({
    $core.String? walletID,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    return $result;
  }
  GetWalletDetailsRequest._() : super();
  factory GetWalletDetailsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetWalletDetailsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetWalletDetailsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetWalletDetailsRequest clone() => GetWalletDetailsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetWalletDetailsRequest copyWith(void Function(GetWalletDetailsRequest) updates) => super.copyWith((message) => updates(message as GetWalletDetailsRequest)) as GetWalletDetailsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetWalletDetailsRequest create() => GetWalletDetailsRequest._();
  GetWalletDetailsRequest createEmptyInstance() => create();
  static $pb.PbList<GetWalletDetailsRequest> createRepeated() => $pb.PbList<GetWalletDetailsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetWalletDetailsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetWalletDetailsRequest>(create);
  static GetWalletDetailsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);
}

class WalletDetails extends $pb.GeneratedMessage {
  factory WalletDetails({
    $core.String? walletID,
    $core.String? firstName,
    $core.String? lastName,
    $core.String? countryCode,
    $core.int? gender,
    $6.Timestamp? dateOfBirth,
    $core.String? address,
    $core.Iterable<User>? users,
    $core.String? kycStatus,
    $core.String? placeOfBirth,
    $core.String? nationality,
    $core.String? walletName,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (firstName != null) {
      $result.firstName = firstName;
    }
    if (lastName != null) {
      $result.lastName = lastName;
    }
    if (countryCode != null) {
      $result.countryCode = countryCode;
    }
    if (gender != null) {
      $result.gender = gender;
    }
    if (dateOfBirth != null) {
      $result.dateOfBirth = dateOfBirth;
    }
    if (address != null) {
      $result.address = address;
    }
    if (users != null) {
      $result.users.addAll(users);
    }
    if (kycStatus != null) {
      $result.kycStatus = kycStatus;
    }
    if (placeOfBirth != null) {
      $result.placeOfBirth = placeOfBirth;
    }
    if (nationality != null) {
      $result.nationality = nationality;
    }
    if (walletName != null) {
      $result.walletName = walletName;
    }
    return $result;
  }
  WalletDetails._() : super();
  factory WalletDetails.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WalletDetails.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WalletDetails', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(2, _omitFieldNames ? '' : 'firstName', protoName: 'firstName')
    ..aOS(3, _omitFieldNames ? '' : 'lastName', protoName: 'lastName')
    ..aOS(4, _omitFieldNames ? '' : 'countryCode', protoName: 'countryCode')
    ..a<$core.int>(5, _omitFieldNames ? '' : 'gender', $pb.PbFieldType.O3)
    ..aOM<$6.Timestamp>(6, _omitFieldNames ? '' : 'dateOfBirth', protoName: 'dateOfBirth', subBuilder: $6.Timestamp.create)
    ..aOS(7, _omitFieldNames ? '' : 'address')
    ..pc<User>(8, _omitFieldNames ? '' : 'users', $pb.PbFieldType.PM, subBuilder: User.create)
    ..aOS(9, _omitFieldNames ? '' : 'kycStatus', protoName: 'kycStatus')
    ..aOS(10, _omitFieldNames ? '' : 'placeOfBirth', protoName: 'placeOfBirth')
    ..aOS(11, _omitFieldNames ? '' : 'nationality')
    ..aOS(12, _omitFieldNames ? '' : 'walletName', protoName: 'walletName')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WalletDetails clone() => WalletDetails()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WalletDetails copyWith(void Function(WalletDetails) updates) => super.copyWith((message) => updates(message as WalletDetails)) as WalletDetails;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WalletDetails create() => WalletDetails._();
  WalletDetails createEmptyInstance() => create();
  static $pb.PbList<WalletDetails> createRepeated() => $pb.PbList<WalletDetails>();
  @$core.pragma('dart2js:noInline')
  static WalletDetails getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WalletDetails>(create);
  static WalletDetails? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get firstName => $_getSZ(1);
  @$pb.TagNumber(2)
  set firstName($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasFirstName() => $_has(1);
  @$pb.TagNumber(2)
  void clearFirstName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get lastName => $_getSZ(2);
  @$pb.TagNumber(3)
  set lastName($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasLastName() => $_has(2);
  @$pb.TagNumber(3)
  void clearLastName() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get countryCode => $_getSZ(3);
  @$pb.TagNumber(4)
  set countryCode($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCountryCode() => $_has(3);
  @$pb.TagNumber(4)
  void clearCountryCode() => clearField(4);

  @$pb.TagNumber(5)
  $core.int get gender => $_getIZ(4);
  @$pb.TagNumber(5)
  set gender($core.int v) { $_setSignedInt32(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasGender() => $_has(4);
  @$pb.TagNumber(5)
  void clearGender() => clearField(5);

  @$pb.TagNumber(6)
  $6.Timestamp get dateOfBirth => $_getN(5);
  @$pb.TagNumber(6)
  set dateOfBirth($6.Timestamp v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasDateOfBirth() => $_has(5);
  @$pb.TagNumber(6)
  void clearDateOfBirth() => clearField(6);
  @$pb.TagNumber(6)
  $6.Timestamp ensureDateOfBirth() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.String get address => $_getSZ(6);
  @$pb.TagNumber(7)
  set address($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasAddress() => $_has(6);
  @$pb.TagNumber(7)
  void clearAddress() => clearField(7);

  @$pb.TagNumber(8)
  $core.List<User> get users => $_getList(7);

  @$pb.TagNumber(9)
  $core.String get kycStatus => $_getSZ(8);
  @$pb.TagNumber(9)
  set kycStatus($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasKycStatus() => $_has(8);
  @$pb.TagNumber(9)
  void clearKycStatus() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get placeOfBirth => $_getSZ(9);
  @$pb.TagNumber(10)
  set placeOfBirth($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasPlaceOfBirth() => $_has(9);
  @$pb.TagNumber(10)
  void clearPlaceOfBirth() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get nationality => $_getSZ(10);
  @$pb.TagNumber(11)
  set nationality($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasNationality() => $_has(10);
  @$pb.TagNumber(11)
  void clearNationality() => clearField(11);

  @$pb.TagNumber(12)
  $core.String get walletName => $_getSZ(11);
  @$pb.TagNumber(12)
  set walletName($core.String v) { $_setString(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasWalletName() => $_has(11);
  @$pb.TagNumber(12)
  void clearWalletName() => clearField(12);
}

class PaginationRequest extends $pb.GeneratedMessage {
  factory PaginationRequest({
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final $result = create();
    if (pageSize != null) {
      $result.pageSize = pageSize;
    }
    if (pageToken != null) {
      $result.pageToken = pageToken;
    }
    return $result;
  }
  PaginationRequest._() : super();
  factory PaginationRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PaginationRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PaginationRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3, protoName: 'pageSize')
    ..aOS(2, _omitFieldNames ? '' : 'pageToken', protoName: 'pageToken')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PaginationRequest clone() => PaginationRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PaginationRequest copyWith(void Function(PaginationRequest) updates) => super.copyWith((message) => updates(message as PaginationRequest)) as PaginationRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PaginationRequest create() => PaginationRequest._();
  PaginationRequest createEmptyInstance() => create();
  static $pb.PbList<PaginationRequest> createRepeated() => $pb.PbList<PaginationRequest>();
  @$core.pragma('dart2js:noInline')
  static PaginationRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PaginationRequest>(create);
  static PaginationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get pageSize => $_getIZ(0);
  @$pb.TagNumber(1)
  set pageSize($core.int v) { $_setSignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPageSize() => $_has(0);
  @$pb.TagNumber(1)
  void clearPageSize() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get pageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set pageToken($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageToken() => clearField(2);
}

class Wallet extends $pb.GeneratedMessage {
  factory Wallet({
    $core.String? walletID,
    $core.String? walletName,
    $core.Iterable<User>? users,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (walletName != null) {
      $result.walletName = walletName;
    }
    if (users != null) {
      $result.users.addAll(users);
    }
    return $result;
  }
  Wallet._() : super();
  factory Wallet.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Wallet.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Wallet', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(2, _omitFieldNames ? '' : 'walletName', protoName: 'walletName')
    ..pc<User>(3, _omitFieldNames ? '' : 'users', $pb.PbFieldType.PM, subBuilder: User.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Wallet clone() => Wallet()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Wallet copyWith(void Function(Wallet) updates) => super.copyWith((message) => updates(message as Wallet)) as Wallet;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Wallet create() => Wallet._();
  Wallet createEmptyInstance() => create();
  static $pb.PbList<Wallet> createRepeated() => $pb.PbList<Wallet>();
  @$core.pragma('dart2js:noInline')
  static Wallet getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Wallet>(create);
  static Wallet? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get walletName => $_getSZ(1);
  @$pb.TagNumber(2)
  set walletName($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasWalletName() => $_has(1);
  @$pb.TagNumber(2)
  void clearWalletName() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<User> get users => $_getList(2);
}

class ListWalletsResponse extends $pb.GeneratedMessage {
  factory ListWalletsResponse({
    $core.Iterable<Wallet>? wallets,
    $core.String? nextPageToken,
  }) {
    final $result = create();
    if (wallets != null) {
      $result.wallets.addAll(wallets);
    }
    if (nextPageToken != null) {
      $result.nextPageToken = nextPageToken;
    }
    return $result;
  }
  ListWalletsResponse._() : super();
  factory ListWalletsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListWalletsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListWalletsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<Wallet>(1, _omitFieldNames ? '' : 'wallets', $pb.PbFieldType.PM, subBuilder: Wallet.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken', protoName: 'nextPageToken')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListWalletsResponse clone() => ListWalletsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListWalletsResponse copyWith(void Function(ListWalletsResponse) updates) => super.copyWith((message) => updates(message as ListWalletsResponse)) as ListWalletsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListWalletsResponse create() => ListWalletsResponse._();
  ListWalletsResponse createEmptyInstance() => create();
  static $pb.PbList<ListWalletsResponse> createRepeated() => $pb.PbList<ListWalletsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListWalletsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListWalletsResponse>(create);
  static ListWalletsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Wallet> get wallets => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => clearField(2);
}

class User extends $pb.GeneratedMessage {
  factory User({
    $core.String? id,
    $core.String? email,
    $core.String? phoneNumber,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (email != null) {
      $result.email = email;
    }
    if (phoneNumber != null) {
      $result.phoneNumber = phoneNumber;
    }
    return $result;
  }
  User._() : super();
  factory User.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory User.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'User', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'email')
    ..aOS(3, _omitFieldNames ? '' : 'phoneNumber', protoName: 'phoneNumber')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  User clone() => User()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  User copyWith(void Function(User) updates) => super.copyWith((message) => updates(message as User)) as User;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static User create() => User._();
  User createEmptyInstance() => create();
  static $pb.PbList<User> createRepeated() => $pb.PbList<User>();
  @$core.pragma('dart2js:noInline')
  static User getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<User>(create);
  static User? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get email => $_getSZ(1);
  @$pb.TagNumber(2)
  set email($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasEmail() => $_has(1);
  @$pb.TagNumber(2)
  void clearEmail() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get phoneNumber => $_getSZ(2);
  @$pb.TagNumber(3)
  set phoneNumber($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPhoneNumber() => $_has(2);
  @$pb.TagNumber(3)
  void clearPhoneNumber() => clearField(3);
}

class AllowWaitlistSignupRequest extends $pb.GeneratedMessage {
  factory AllowWaitlistSignupRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  AllowWaitlistSignupRequest._() : super();
  factory AllowWaitlistSignupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AllowWaitlistSignupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AllowWaitlistSignupRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AllowWaitlistSignupRequest clone() => AllowWaitlistSignupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AllowWaitlistSignupRequest copyWith(void Function(AllowWaitlistSignupRequest) updates) => super.copyWith((message) => updates(message as AllowWaitlistSignupRequest)) as AllowWaitlistSignupRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllowWaitlistSignupRequest create() => AllowWaitlistSignupRequest._();
  AllowWaitlistSignupRequest createEmptyInstance() => create();
  static $pb.PbList<AllowWaitlistSignupRequest> createRepeated() => $pb.PbList<AllowWaitlistSignupRequest>();
  @$core.pragma('dart2js:noInline')
  static AllowWaitlistSignupRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AllowWaitlistSignupRequest>(create);
  static AllowWaitlistSignupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class ListWaitlistSignupsResponse extends $pb.GeneratedMessage {
  factory ListWaitlistSignupsResponse({
    $core.Iterable<WaitlistSignup>? signups,
  }) {
    final $result = create();
    if (signups != null) {
      $result.signups.addAll(signups);
    }
    return $result;
  }
  ListWaitlistSignupsResponse._() : super();
  factory ListWaitlistSignupsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListWaitlistSignupsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListWaitlistSignupsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<WaitlistSignup>(1, _omitFieldNames ? '' : 'signups', $pb.PbFieldType.PM, subBuilder: WaitlistSignup.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListWaitlistSignupsResponse clone() => ListWaitlistSignupsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListWaitlistSignupsResponse copyWith(void Function(ListWaitlistSignupsResponse) updates) => super.copyWith((message) => updates(message as ListWaitlistSignupsResponse)) as ListWaitlistSignupsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListWaitlistSignupsResponse create() => ListWaitlistSignupsResponse._();
  ListWaitlistSignupsResponse createEmptyInstance() => create();
  static $pb.PbList<ListWaitlistSignupsResponse> createRepeated() => $pb.PbList<ListWaitlistSignupsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListWaitlistSignupsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListWaitlistSignupsResponse>(create);
  static ListWaitlistSignupsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<WaitlistSignup> get signups => $_getList(0);
}

class WaitlistSignup extends $pb.GeneratedMessage {
  factory WaitlistSignup({
    $core.String? id,
    $core.String? name,
    $core.String? email,
    $core.bool? betaOptIn,
    $core.bool? canSignup,
    $core.String? mugId,
    $core.String? countryCode,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (name != null) {
      $result.name = name;
    }
    if (email != null) {
      $result.email = email;
    }
    if (betaOptIn != null) {
      $result.betaOptIn = betaOptIn;
    }
    if (canSignup != null) {
      $result.canSignup = canSignup;
    }
    if (mugId != null) {
      $result.mugId = mugId;
    }
    if (countryCode != null) {
      $result.countryCode = countryCode;
    }
    return $result;
  }
  WaitlistSignup._() : super();
  factory WaitlistSignup.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitlistSignup.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitlistSignup', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'email')
    ..aOB(4, _omitFieldNames ? '' : 'betaOptIn')
    ..aOB(5, _omitFieldNames ? '' : 'canSignup')
    ..aOS(6, _omitFieldNames ? '' : 'mugId')
    ..aOS(7, _omitFieldNames ? '' : 'countryCode')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitlistSignup clone() => WaitlistSignup()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitlistSignup copyWith(void Function(WaitlistSignup) updates) => super.copyWith((message) => updates(message as WaitlistSignup)) as WaitlistSignup;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitlistSignup create() => WaitlistSignup._();
  WaitlistSignup createEmptyInstance() => create();
  static $pb.PbList<WaitlistSignup> createRepeated() => $pb.PbList<WaitlistSignup>();
  @$core.pragma('dart2js:noInline')
  static WaitlistSignup getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitlistSignup>(create);
  static WaitlistSignup? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get email => $_getSZ(2);
  @$pb.TagNumber(3)
  set email($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasEmail() => $_has(2);
  @$pb.TagNumber(3)
  void clearEmail() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get betaOptIn => $_getBF(3);
  @$pb.TagNumber(4)
  set betaOptIn($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasBetaOptIn() => $_has(3);
  @$pb.TagNumber(4)
  void clearBetaOptIn() => clearField(4);

  @$pb.TagNumber(5)
  $core.bool get canSignup => $_getBF(4);
  @$pb.TagNumber(5)
  set canSignup($core.bool v) { $_setBool(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCanSignup() => $_has(4);
  @$pb.TagNumber(5)
  void clearCanSignup() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get mugId => $_getSZ(5);
  @$pb.TagNumber(6)
  set mugId($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasMugId() => $_has(5);
  @$pb.TagNumber(6)
  void clearMugId() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get countryCode => $_getSZ(6);
  @$pb.TagNumber(7)
  set countryCode($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasCountryCode() => $_has(6);
  @$pb.TagNumber(7)
  void clearCountryCode() => clearField(7);
}

class FormSubmissionCount extends $pb.GeneratedMessage {
  factory FormSubmissionCount({
    $core.String? formId,
    $core.int? submissionCount,
  }) {
    final $result = create();
    if (formId != null) {
      $result.formId = formId;
    }
    if (submissionCount != null) {
      $result.submissionCount = submissionCount;
    }
    return $result;
  }
  FormSubmissionCount._() : super();
  factory FormSubmissionCount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory FormSubmissionCount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'FormSubmissionCount', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'formId')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'submissionCount', $pb.PbFieldType.O3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  FormSubmissionCount clone() => FormSubmissionCount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  FormSubmissionCount copyWith(void Function(FormSubmissionCount) updates) => super.copyWith((message) => updates(message as FormSubmissionCount)) as FormSubmissionCount;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FormSubmissionCount create() => FormSubmissionCount._();
  FormSubmissionCount createEmptyInstance() => create();
  static $pb.PbList<FormSubmissionCount> createRepeated() => $pb.PbList<FormSubmissionCount>();
  @$core.pragma('dart2js:noInline')
  static FormSubmissionCount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FormSubmissionCount>(create);
  static FormSubmissionCount? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get formId => $_getSZ(0);
  @$pb.TagNumber(1)
  set formId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFormId() => $_has(0);
  @$pb.TagNumber(1)
  void clearFormId() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get submissionCount => $_getIZ(1);
  @$pb.TagNumber(2)
  set submissionCount($core.int v) { $_setSignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasSubmissionCount() => $_has(1);
  @$pb.TagNumber(2)
  void clearSubmissionCount() => clearField(2);
}

class ListFormSubmissionCountsResponse extends $pb.GeneratedMessage {
  factory ListFormSubmissionCountsResponse({
    $core.Iterable<FormSubmissionCount>? formSubmissionCounts,
    $core.String? nextPageToken,
  }) {
    final $result = create();
    if (formSubmissionCounts != null) {
      $result.formSubmissionCounts.addAll(formSubmissionCounts);
    }
    if (nextPageToken != null) {
      $result.nextPageToken = nextPageToken;
    }
    return $result;
  }
  ListFormSubmissionCountsResponse._() : super();
  factory ListFormSubmissionCountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListFormSubmissionCountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListFormSubmissionCountsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<FormSubmissionCount>(1, _omitFieldNames ? '' : 'formSubmissionCounts', $pb.PbFieldType.PM, subBuilder: FormSubmissionCount.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListFormSubmissionCountsResponse clone() => ListFormSubmissionCountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListFormSubmissionCountsResponse copyWith(void Function(ListFormSubmissionCountsResponse) updates) => super.copyWith((message) => updates(message as ListFormSubmissionCountsResponse)) as ListFormSubmissionCountsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListFormSubmissionCountsResponse create() => ListFormSubmissionCountsResponse._();
  ListFormSubmissionCountsResponse createEmptyInstance() => create();
  static $pb.PbList<ListFormSubmissionCountsResponse> createRepeated() => $pb.PbList<ListFormSubmissionCountsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListFormSubmissionCountsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListFormSubmissionCountsResponse>(create);
  static ListFormSubmissionCountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<FormSubmissionCount> get formSubmissionCounts => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => clearField(2);
}

class ExportFormSubmissionsRequest extends $pb.GeneratedMessage {
  factory ExportFormSubmissionsRequest({
    $core.String? formId,
  }) {
    final $result = create();
    if (formId != null) {
      $result.formId = formId;
    }
    return $result;
  }
  ExportFormSubmissionsRequest._() : super();
  factory ExportFormSubmissionsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ExportFormSubmissionsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ExportFormSubmissionsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'formId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ExportFormSubmissionsRequest clone() => ExportFormSubmissionsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ExportFormSubmissionsRequest copyWith(void Function(ExportFormSubmissionsRequest) updates) => super.copyWith((message) => updates(message as ExportFormSubmissionsRequest)) as ExportFormSubmissionsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ExportFormSubmissionsRequest create() => ExportFormSubmissionsRequest._();
  ExportFormSubmissionsRequest createEmptyInstance() => create();
  static $pb.PbList<ExportFormSubmissionsRequest> createRepeated() => $pb.PbList<ExportFormSubmissionsRequest>();
  @$core.pragma('dart2js:noInline')
  static ExportFormSubmissionsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ExportFormSubmissionsRequest>(create);
  static ExportFormSubmissionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get formId => $_getSZ(0);
  @$pb.TagNumber(1)
  set formId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFormId() => $_has(0);
  @$pb.TagNumber(1)
  void clearFormId() => clearField(1);
}

class ExportFormSubmissionsResponse extends $pb.GeneratedMessage {
  factory ExportFormSubmissionsResponse({
    $core.List<$core.int>? chunk,
  }) {
    final $result = create();
    if (chunk != null) {
      $result.chunk = chunk;
    }
    return $result;
  }
  ExportFormSubmissionsResponse._() : super();
  factory ExportFormSubmissionsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ExportFormSubmissionsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ExportFormSubmissionsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, _omitFieldNames ? '' : 'chunk', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ExportFormSubmissionsResponse clone() => ExportFormSubmissionsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ExportFormSubmissionsResponse copyWith(void Function(ExportFormSubmissionsResponse) updates) => super.copyWith((message) => updates(message as ExportFormSubmissionsResponse)) as ExportFormSubmissionsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ExportFormSubmissionsResponse create() => ExportFormSubmissionsResponse._();
  ExportFormSubmissionsResponse createEmptyInstance() => create();
  static $pb.PbList<ExportFormSubmissionsResponse> createRepeated() => $pb.PbList<ExportFormSubmissionsResponse>();
  @$core.pragma('dart2js:noInline')
  static ExportFormSubmissionsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ExportFormSubmissionsResponse>(create);
  static ExportFormSubmissionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get chunk => $_getN(0);
  @$pb.TagNumber(1)
  set chunk($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasChunk() => $_has(0);
  @$pb.TagNumber(1)
  void clearChunk() => clearField(1);
}

class FormSubmission extends $pb.GeneratedMessage {
  factory FormSubmission({
    $core.String? id,
    $core.String? formId,
    $6.Timestamp? timestamp,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (formId != null) {
      $result.formId = formId;
    }
    if (timestamp != null) {
      $result.timestamp = timestamp;
    }
    return $result;
  }
  FormSubmission._() : super();
  factory FormSubmission.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory FormSubmission.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'FormSubmission', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'formId')
    ..aOM<$6.Timestamp>(3, _omitFieldNames ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  FormSubmission clone() => FormSubmission()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  FormSubmission copyWith(void Function(FormSubmission) updates) => super.copyWith((message) => updates(message as FormSubmission)) as FormSubmission;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FormSubmission create() => FormSubmission._();
  FormSubmission createEmptyInstance() => create();
  static $pb.PbList<FormSubmission> createRepeated() => $pb.PbList<FormSubmission>();
  @$core.pragma('dart2js:noInline')
  static FormSubmission getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FormSubmission>(create);
  static FormSubmission? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get formId => $_getSZ(1);
  @$pb.TagNumber(2)
  set formId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasFormId() => $_has(1);
  @$pb.TagNumber(2)
  void clearFormId() => clearField(2);

  @$pb.TagNumber(3)
  $6.Timestamp get timestamp => $_getN(2);
  @$pb.TagNumber(3)
  set timestamp($6.Timestamp v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasTimestamp() => $_has(2);
  @$pb.TagNumber(3)
  void clearTimestamp() => clearField(3);
  @$pb.TagNumber(3)
  $6.Timestamp ensureTimestamp() => $_ensure(2);
}

class ListFormSubmissionsRequest extends $pb.GeneratedMessage {
  factory ListFormSubmissionsRequest({
    $core.String? formId,
  }) {
    final $result = create();
    if (formId != null) {
      $result.formId = formId;
    }
    return $result;
  }
  ListFormSubmissionsRequest._() : super();
  factory ListFormSubmissionsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListFormSubmissionsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListFormSubmissionsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'formId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListFormSubmissionsRequest clone() => ListFormSubmissionsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListFormSubmissionsRequest copyWith(void Function(ListFormSubmissionsRequest) updates) => super.copyWith((message) => updates(message as ListFormSubmissionsRequest)) as ListFormSubmissionsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListFormSubmissionsRequest create() => ListFormSubmissionsRequest._();
  ListFormSubmissionsRequest createEmptyInstance() => create();
  static $pb.PbList<ListFormSubmissionsRequest> createRepeated() => $pb.PbList<ListFormSubmissionsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListFormSubmissionsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListFormSubmissionsRequest>(create);
  static ListFormSubmissionsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get formId => $_getSZ(0);
  @$pb.TagNumber(1)
  set formId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFormId() => $_has(0);
  @$pb.TagNumber(1)
  void clearFormId() => clearField(1);
}

class ListFormSubmissionsResponse extends $pb.GeneratedMessage {
  factory ListFormSubmissionsResponse({
    $core.Iterable<FormSubmission>? formSubmissions,
  }) {
    final $result = create();
    if (formSubmissions != null) {
      $result.formSubmissions.addAll(formSubmissions);
    }
    return $result;
  }
  ListFormSubmissionsResponse._() : super();
  factory ListFormSubmissionsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListFormSubmissionsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListFormSubmissionsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..pc<FormSubmission>(1, _omitFieldNames ? '' : 'formSubmissions', $pb.PbFieldType.PM, subBuilder: FormSubmission.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListFormSubmissionsResponse clone() => ListFormSubmissionsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListFormSubmissionsResponse copyWith(void Function(ListFormSubmissionsResponse) updates) => super.copyWith((message) => updates(message as ListFormSubmissionsResponse)) as ListFormSubmissionsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListFormSubmissionsResponse create() => ListFormSubmissionsResponse._();
  ListFormSubmissionsResponse createEmptyInstance() => create();
  static $pb.PbList<ListFormSubmissionsResponse> createRepeated() => $pb.PbList<ListFormSubmissionsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListFormSubmissionsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListFormSubmissionsResponse>(create);
  static ListFormSubmissionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<FormSubmission> get formSubmissions => $_getList(0);
}

class GetFormSubmissionDetailsRequest extends $pb.GeneratedMessage {
  factory GetFormSubmissionDetailsRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetFormSubmissionDetailsRequest._() : super();
  factory GetFormSubmissionDetailsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetFormSubmissionDetailsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetFormSubmissionDetailsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetFormSubmissionDetailsRequest clone() => GetFormSubmissionDetailsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetFormSubmissionDetailsRequest copyWith(void Function(GetFormSubmissionDetailsRequest) updates) => super.copyWith((message) => updates(message as GetFormSubmissionDetailsRequest)) as GetFormSubmissionDetailsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetFormSubmissionDetailsRequest create() => GetFormSubmissionDetailsRequest._();
  GetFormSubmissionDetailsRequest createEmptyInstance() => create();
  static $pb.PbList<GetFormSubmissionDetailsRequest> createRepeated() => $pb.PbList<GetFormSubmissionDetailsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetFormSubmissionDetailsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetFormSubmissionDetailsRequest>(create);
  static GetFormSubmissionDetailsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class FormSubmissionDetails extends $pb.GeneratedMessage {
  factory FormSubmissionDetails({
    $core.String? id,
    $core.String? walletId,
    $core.String? formId,
    $core.String? data,
    $6.Timestamp? timestamp,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (walletId != null) {
      $result.walletId = walletId;
    }
    if (formId != null) {
      $result.formId = formId;
    }
    if (data != null) {
      $result.data = data;
    }
    if (timestamp != null) {
      $result.timestamp = timestamp;
    }
    return $result;
  }
  FormSubmissionDetails._() : super();
  factory FormSubmissionDetails.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory FormSubmissionDetails.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'FormSubmissionDetails', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'walletId')
    ..aOS(3, _omitFieldNames ? '' : 'formId')
    ..aOS(4, _omitFieldNames ? '' : 'data')
    ..aOM<$6.Timestamp>(5, _omitFieldNames ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  FormSubmissionDetails clone() => FormSubmissionDetails()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  FormSubmissionDetails copyWith(void Function(FormSubmissionDetails) updates) => super.copyWith((message) => updates(message as FormSubmissionDetails)) as FormSubmissionDetails;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FormSubmissionDetails create() => FormSubmissionDetails._();
  FormSubmissionDetails createEmptyInstance() => create();
  static $pb.PbList<FormSubmissionDetails> createRepeated() => $pb.PbList<FormSubmissionDetails>();
  @$core.pragma('dart2js:noInline')
  static FormSubmissionDetails getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FormSubmissionDetails>(create);
  static FormSubmissionDetails? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get walletId => $_getSZ(1);
  @$pb.TagNumber(2)
  set walletId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasWalletId() => $_has(1);
  @$pb.TagNumber(2)
  void clearWalletId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get formId => $_getSZ(2);
  @$pb.TagNumber(3)
  set formId($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasFormId() => $_has(2);
  @$pb.TagNumber(3)
  void clearFormId() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get data => $_getSZ(3);
  @$pb.TagNumber(4)
  set data($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasData() => $_has(3);
  @$pb.TagNumber(4)
  void clearData() => clearField(4);

  @$pb.TagNumber(5)
  $6.Timestamp get timestamp => $_getN(4);
  @$pb.TagNumber(5)
  set timestamp($6.Timestamp v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasTimestamp() => $_has(4);
  @$pb.TagNumber(5)
  void clearTimestamp() => clearField(5);
  @$pb.TagNumber(5)
  $6.Timestamp ensureTimestamp() => $_ensure(4);
}

class SetWalletXagoBalanceEnabledRequest extends $pb.GeneratedMessage {
  factory SetWalletXagoBalanceEnabledRequest({
    $core.String? walletId,
  }) {
    final $result = create();
    if (walletId != null) {
      $result.walletId = walletId;
    }
    return $result;
  }
  SetWalletXagoBalanceEnabledRequest._() : super();
  factory SetWalletXagoBalanceEnabledRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetWalletXagoBalanceEnabledRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetWalletXagoBalanceEnabledRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetWalletXagoBalanceEnabledRequest clone() => SetWalletXagoBalanceEnabledRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetWalletXagoBalanceEnabledRequest copyWith(void Function(SetWalletXagoBalanceEnabledRequest) updates) => super.copyWith((message) => updates(message as SetWalletXagoBalanceEnabledRequest)) as SetWalletXagoBalanceEnabledRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetWalletXagoBalanceEnabledRequest create() => SetWalletXagoBalanceEnabledRequest._();
  SetWalletXagoBalanceEnabledRequest createEmptyInstance() => create();
  static $pb.PbList<SetWalletXagoBalanceEnabledRequest> createRepeated() => $pb.PbList<SetWalletXagoBalanceEnabledRequest>();
  @$core.pragma('dart2js:noInline')
  static SetWalletXagoBalanceEnabledRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetWalletXagoBalanceEnabledRequest>(create);
  static SetWalletXagoBalanceEnabledRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletId => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletId() => clearField(1);
}

class GetWalletXagoBalanceRequest extends $pb.GeneratedMessage {
  factory GetWalletXagoBalanceRequest({
    $core.String? walletId,
  }) {
    final $result = create();
    if (walletId != null) {
      $result.walletId = walletId;
    }
    return $result;
  }
  GetWalletXagoBalanceRequest._() : super();
  factory GetWalletXagoBalanceRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetWalletXagoBalanceRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetWalletXagoBalanceRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetWalletXagoBalanceRequest clone() => GetWalletXagoBalanceRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetWalletXagoBalanceRequest copyWith(void Function(GetWalletXagoBalanceRequest) updates) => super.copyWith((message) => updates(message as GetWalletXagoBalanceRequest)) as GetWalletXagoBalanceRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetWalletXagoBalanceRequest create() => GetWalletXagoBalanceRequest._();
  GetWalletXagoBalanceRequest createEmptyInstance() => create();
  static $pb.PbList<GetWalletXagoBalanceRequest> createRepeated() => $pb.PbList<GetWalletXagoBalanceRequest>();
  @$core.pragma('dart2js:noInline')
  static GetWalletXagoBalanceRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetWalletXagoBalanceRequest>(create);
  static GetWalletXagoBalanceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletId => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletId() => clearField(1);
}

class GetWalletXagoBalanceResponse extends $pb.GeneratedMessage {
  factory GetWalletXagoBalanceResponse({
    Amount? balance,
    Amount? available,
  }) {
    final $result = create();
    if (balance != null) {
      $result.balance = balance;
    }
    if (available != null) {
      $result.available = available;
    }
    return $result;
  }
  GetWalletXagoBalanceResponse._() : super();
  factory GetWalletXagoBalanceResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetWalletXagoBalanceResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetWalletXagoBalanceResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..aOM<Amount>(1, _omitFieldNames ? '' : 'balance', subBuilder: Amount.create)
    ..aOM<Amount>(2, _omitFieldNames ? '' : 'available', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetWalletXagoBalanceResponse clone() => GetWalletXagoBalanceResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetWalletXagoBalanceResponse copyWith(void Function(GetWalletXagoBalanceResponse) updates) => super.copyWith((message) => updates(message as GetWalletXagoBalanceResponse)) as GetWalletXagoBalanceResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetWalletXagoBalanceResponse create() => GetWalletXagoBalanceResponse._();
  GetWalletXagoBalanceResponse createEmptyInstance() => create();
  static $pb.PbList<GetWalletXagoBalanceResponse> createRepeated() => $pb.PbList<GetWalletXagoBalanceResponse>();
  @$core.pragma('dart2js:noInline')
  static GetWalletXagoBalanceResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetWalletXagoBalanceResponse>(create);
  static GetWalletXagoBalanceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Amount get balance => $_getN(0);
  @$pb.TagNumber(1)
  set balance(Amount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBalance() => $_has(0);
  @$pb.TagNumber(1)
  void clearBalance() => clearField(1);
  @$pb.TagNumber(1)
  Amount ensureBalance() => $_ensure(0);

  @$pb.TagNumber(2)
  Amount get available => $_getN(1);
  @$pb.TagNumber(2)
  set available(Amount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasAvailable() => $_has(1);
  @$pb.TagNumber(2)
  void clearAvailable() => clearField(2);
  @$pb.TagNumber(2)
  Amount ensureAvailable() => $_ensure(1);
}

class Amount extends $pb.GeneratedMessage {
  factory Amount({
    $fixnum.Int64? amount,
    $core.String? asset,
    $core.int? assetScale,
    $core.String? country,
  }) {
    final $result = create();
    if (amount != null) {
      $result.amount = amount;
    }
    if (asset != null) {
      $result.asset = asset;
    }
    if (assetScale != null) {
      $result.assetScale = assetScale;
    }
    if (country != null) {
      $result.country = country;
    }
    return $result;
  }
  Amount._() : super();
  factory Amount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Amount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Amount', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.admin.v1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, _omitFieldNames ? '' : 'amount', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, _omitFieldNames ? '' : 'asset')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'assetScale', $pb.PbFieldType.O3, protoName: 'assetScale')
    ..aOS(4, _omitFieldNames ? '' : 'country')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Amount clone() => Amount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Amount copyWith(void Function(Amount) updates) => super.copyWith((message) => updates(message as Amount)) as Amount;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Amount create() => Amount._();
  Amount createEmptyInstance() => create();
  static $pb.PbList<Amount> createRepeated() => $pb.PbList<Amount>();
  @$core.pragma('dart2js:noInline')
  static Amount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Amount>(create);
  static Amount? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get amount => $_getI64(0);
  @$pb.TagNumber(1)
  set amount($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAmount() => $_has(0);
  @$pb.TagNumber(1)
  void clearAmount() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get asset => $_getSZ(1);
  @$pb.TagNumber(2)
  set asset($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAsset() => $_has(1);
  @$pb.TagNumber(2)
  void clearAsset() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get assetScale => $_getIZ(2);
  @$pb.TagNumber(3)
  set assetScale($core.int v) { $_setSignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAssetScale() => $_has(2);
  @$pb.TagNumber(3)
  void clearAssetScale() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get country => $_getSZ(3);
  @$pb.TagNumber(4)
  set country($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCountry() => $_has(3);
  @$pb.TagNumber(4)
  void clearCountry() => clearField(4);
}


const _omitFieldNames = $core.bool.fromEnvironment('protobuf.omit_field_names');
const _omitMessageNames = $core.bool.fromEnvironment('protobuf.omit_message_names');
