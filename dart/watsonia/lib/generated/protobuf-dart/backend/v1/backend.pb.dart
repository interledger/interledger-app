//
//  Generated code. Do not modify.
//  source: backend/v1/backend.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../google/protobuf/timestamp.pb.dart' as $6;

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

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PaginationRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
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

class Empty extends $pb.GeneratedMessage {
  factory Empty() => create();
  Empty._() : super();
  factory Empty.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Empty.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Empty', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
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

class GetLinkedAccountsForPaymentRequest extends $pb.GeneratedMessage {
  factory GetLinkedAccountsForPaymentRequest({
    $core.String? paymentId,
  }) {
    final $result = create();
    if (paymentId != null) {
      $result.paymentId = paymentId;
    }
    return $result;
  }
  GetLinkedAccountsForPaymentRequest._() : super();
  factory GetLinkedAccountsForPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLinkedAccountsForPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLinkedAccountsForPaymentRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'paymentId', protoName: 'paymentId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsForPaymentRequest clone() => GetLinkedAccountsForPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsForPaymentRequest copyWith(void Function(GetLinkedAccountsForPaymentRequest) updates) => super.copyWith((message) => updates(message as GetLinkedAccountsForPaymentRequest)) as GetLinkedAccountsForPaymentRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountsForPaymentRequest create() => GetLinkedAccountsForPaymentRequest._();
  GetLinkedAccountsForPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<GetLinkedAccountsForPaymentRequest> createRepeated() => $pb.PbList<GetLinkedAccountsForPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountsForPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetLinkedAccountsForPaymentRequest>(create);
  static GetLinkedAccountsForPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get paymentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set paymentId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPaymentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPaymentId() => clearField(1);
}

class GetLinkedAccountsForPaymentResponse extends $pb.GeneratedMessage {
  factory GetLinkedAccountsForPaymentResponse({
    $core.Iterable<LinkedAccountForPayment>? linkedAccounts,
  }) {
    final $result = create();
    if (linkedAccounts != null) {
      $result.linkedAccounts.addAll(linkedAccounts);
    }
    return $result;
  }
  GetLinkedAccountsForPaymentResponse._() : super();
  factory GetLinkedAccountsForPaymentResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLinkedAccountsForPaymentResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLinkedAccountsForPaymentResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<LinkedAccountForPayment>(1, _omitFieldNames ? '' : 'linkedAccounts', $pb.PbFieldType.PM, protoName: 'linkedAccounts', subBuilder: LinkedAccountForPayment.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsForPaymentResponse clone() => GetLinkedAccountsForPaymentResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsForPaymentResponse copyWith(void Function(GetLinkedAccountsForPaymentResponse) updates) => super.copyWith((message) => updates(message as GetLinkedAccountsForPaymentResponse)) as GetLinkedAccountsForPaymentResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountsForPaymentResponse create() => GetLinkedAccountsForPaymentResponse._();
  GetLinkedAccountsForPaymentResponse createEmptyInstance() => create();
  static $pb.PbList<GetLinkedAccountsForPaymentResponse> createRepeated() => $pb.PbList<GetLinkedAccountsForPaymentResponse>();
  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountsForPaymentResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetLinkedAccountsForPaymentResponse>(create);
  static GetLinkedAccountsForPaymentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<LinkedAccountForPayment> get linkedAccounts => $_getList(0);
}

class LinkedAccountForPayment extends $pb.GeneratedMessage {
  factory LinkedAccountForPayment({
    LinkedAccount? details,
    $core.bool? enabled,
  }) {
    final $result = create();
    if (details != null) {
      $result.details = details;
    }
    if (enabled != null) {
      $result.enabled = enabled;
    }
    return $result;
  }
  LinkedAccountForPayment._() : super();
  factory LinkedAccountForPayment.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LinkedAccountForPayment.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LinkedAccountForPayment', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<LinkedAccount>(1, _omitFieldNames ? '' : 'details', subBuilder: LinkedAccount.create)
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LinkedAccountForPayment clone() => LinkedAccountForPayment()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LinkedAccountForPayment copyWith(void Function(LinkedAccountForPayment) updates) => super.copyWith((message) => updates(message as LinkedAccountForPayment)) as LinkedAccountForPayment;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LinkedAccountForPayment create() => LinkedAccountForPayment._();
  LinkedAccountForPayment createEmptyInstance() => create();
  static $pb.PbList<LinkedAccountForPayment> createRepeated() => $pb.PbList<LinkedAccountForPayment>();
  @$core.pragma('dart2js:noInline')
  static LinkedAccountForPayment getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LinkedAccountForPayment>(create);
  static LinkedAccountForPayment? _defaultInstance;

  @$pb.TagNumber(1)
  LinkedAccount get details => $_getN(0);
  @$pb.TagNumber(1)
  set details(LinkedAccount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasDetails() => $_has(0);
  @$pb.TagNumber(1)
  void clearDetails() => clearField(1);
  @$pb.TagNumber(1)
  LinkedAccount ensureDetails() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => clearField(2);
}

class GetXagoDepositDetailsRequest extends $pb.GeneratedMessage {
  factory GetXagoDepositDetailsRequest({
    $core.String? linkedAccount,
  }) {
    final $result = create();
    if (linkedAccount != null) {
      $result.linkedAccount = linkedAccount;
    }
    return $result;
  }
  GetXagoDepositDetailsRequest._() : super();
  factory GetXagoDepositDetailsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetXagoDepositDetailsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetXagoDepositDetailsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'linkedAccount', protoName: 'linkedAccount')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetXagoDepositDetailsRequest clone() => GetXagoDepositDetailsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetXagoDepositDetailsRequest copyWith(void Function(GetXagoDepositDetailsRequest) updates) => super.copyWith((message) => updates(message as GetXagoDepositDetailsRequest)) as GetXagoDepositDetailsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetXagoDepositDetailsRequest create() => GetXagoDepositDetailsRequest._();
  GetXagoDepositDetailsRequest createEmptyInstance() => create();
  static $pb.PbList<GetXagoDepositDetailsRequest> createRepeated() => $pb.PbList<GetXagoDepositDetailsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetXagoDepositDetailsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetXagoDepositDetailsRequest>(create);
  static GetXagoDepositDetailsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get linkedAccount => $_getSZ(0);
  @$pb.TagNumber(1)
  set linkedAccount($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasLinkedAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearLinkedAccount() => clearField(1);
}

class GetXagoDepositDetailsResponse extends $pb.GeneratedMessage {
  factory GetXagoDepositDetailsResponse({
    $core.Iterable<XagoDepositDetails>? details,
  }) {
    final $result = create();
    if (details != null) {
      $result.details.addAll(details);
    }
    return $result;
  }
  GetXagoDepositDetailsResponse._() : super();
  factory GetXagoDepositDetailsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetXagoDepositDetailsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetXagoDepositDetailsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<XagoDepositDetails>(1, _omitFieldNames ? '' : 'details', $pb.PbFieldType.PM, subBuilder: XagoDepositDetails.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetXagoDepositDetailsResponse clone() => GetXagoDepositDetailsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetXagoDepositDetailsResponse copyWith(void Function(GetXagoDepositDetailsResponse) updates) => super.copyWith((message) => updates(message as GetXagoDepositDetailsResponse)) as GetXagoDepositDetailsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetXagoDepositDetailsResponse create() => GetXagoDepositDetailsResponse._();
  GetXagoDepositDetailsResponse createEmptyInstance() => create();
  static $pb.PbList<GetXagoDepositDetailsResponse> createRepeated() => $pb.PbList<GetXagoDepositDetailsResponse>();
  @$core.pragma('dart2js:noInline')
  static GetXagoDepositDetailsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetXagoDepositDetailsResponse>(create);
  static GetXagoDepositDetailsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<XagoDepositDetails> get details => $_getList(0);
}

class XagoDepositDetails extends $pb.GeneratedMessage {
  factory XagoDepositDetails({
    $core.String? currency,
    $core.String? accountNumber,
    $core.String? branchCode,
    $core.String? bankName,
    $core.String? depositReference,
  }) {
    final $result = create();
    if (currency != null) {
      $result.currency = currency;
    }
    if (accountNumber != null) {
      $result.accountNumber = accountNumber;
    }
    if (branchCode != null) {
      $result.branchCode = branchCode;
    }
    if (bankName != null) {
      $result.bankName = bankName;
    }
    if (depositReference != null) {
      $result.depositReference = depositReference;
    }
    return $result;
  }
  XagoDepositDetails._() : super();
  factory XagoDepositDetails.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory XagoDepositDetails.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'XagoDepositDetails', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'currency')
    ..aOS(2, _omitFieldNames ? '' : 'accountNumber', protoName: 'accountNumber')
    ..aOS(3, _omitFieldNames ? '' : 'branchCode', protoName: 'branchCode')
    ..aOS(4, _omitFieldNames ? '' : 'bankName', protoName: 'bankName')
    ..aOS(5, _omitFieldNames ? '' : 'depositReference', protoName: 'depositReference')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  XagoDepositDetails clone() => XagoDepositDetails()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  XagoDepositDetails copyWith(void Function(XagoDepositDetails) updates) => super.copyWith((message) => updates(message as XagoDepositDetails)) as XagoDepositDetails;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static XagoDepositDetails create() => XagoDepositDetails._();
  XagoDepositDetails createEmptyInstance() => create();
  static $pb.PbList<XagoDepositDetails> createRepeated() => $pb.PbList<XagoDepositDetails>();
  @$core.pragma('dart2js:noInline')
  static XagoDepositDetails getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<XagoDepositDetails>(create);
  static XagoDepositDetails? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get currency => $_getSZ(0);
  @$pb.TagNumber(1)
  set currency($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCurrency() => $_has(0);
  @$pb.TagNumber(1)
  void clearCurrency() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get accountNumber => $_getSZ(1);
  @$pb.TagNumber(2)
  set accountNumber($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAccountNumber() => $_has(1);
  @$pb.TagNumber(2)
  void clearAccountNumber() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get branchCode => $_getSZ(2);
  @$pb.TagNumber(3)
  set branchCode($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBranchCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearBranchCode() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get bankName => $_getSZ(3);
  @$pb.TagNumber(4)
  set bankName($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasBankName() => $_has(3);
  @$pb.TagNumber(4)
  void clearBankName() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get depositReference => $_getSZ(4);
  @$pb.TagNumber(5)
  set depositReference($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDepositReference() => $_has(4);
  @$pb.TagNumber(5)
  void clearDepositReference() => clearField(5);
}

class GetXagoBalanceRequest extends $pb.GeneratedMessage {
  factory GetXagoBalanceRequest({
    $core.String? linkedAccount,
  }) {
    final $result = create();
    if (linkedAccount != null) {
      $result.linkedAccount = linkedAccount;
    }
    return $result;
  }
  GetXagoBalanceRequest._() : super();
  factory GetXagoBalanceRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetXagoBalanceRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetXagoBalanceRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'linkedAccount', protoName: 'linkedAccount')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetXagoBalanceRequest clone() => GetXagoBalanceRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetXagoBalanceRequest copyWith(void Function(GetXagoBalanceRequest) updates) => super.copyWith((message) => updates(message as GetXagoBalanceRequest)) as GetXagoBalanceRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetXagoBalanceRequest create() => GetXagoBalanceRequest._();
  GetXagoBalanceRequest createEmptyInstance() => create();
  static $pb.PbList<GetXagoBalanceRequest> createRepeated() => $pb.PbList<GetXagoBalanceRequest>();
  @$core.pragma('dart2js:noInline')
  static GetXagoBalanceRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetXagoBalanceRequest>(create);
  static GetXagoBalanceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get linkedAccount => $_getSZ(0);
  @$pb.TagNumber(1)
  set linkedAccount($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasLinkedAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearLinkedAccount() => clearField(1);
}

class GetXagoBalanceResponse extends $pb.GeneratedMessage {
  factory GetXagoBalanceResponse({
    $core.Iterable<XagoBalance>? balances,
  }) {
    final $result = create();
    if (balances != null) {
      $result.balances.addAll(balances);
    }
    return $result;
  }
  GetXagoBalanceResponse._() : super();
  factory GetXagoBalanceResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetXagoBalanceResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetXagoBalanceResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<XagoBalance>(2, _omitFieldNames ? '' : 'balances', $pb.PbFieldType.PM, subBuilder: XagoBalance.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetXagoBalanceResponse clone() => GetXagoBalanceResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetXagoBalanceResponse copyWith(void Function(GetXagoBalanceResponse) updates) => super.copyWith((message) => updates(message as GetXagoBalanceResponse)) as GetXagoBalanceResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetXagoBalanceResponse create() => GetXagoBalanceResponse._();
  GetXagoBalanceResponse createEmptyInstance() => create();
  static $pb.PbList<GetXagoBalanceResponse> createRepeated() => $pb.PbList<GetXagoBalanceResponse>();
  @$core.pragma('dart2js:noInline')
  static GetXagoBalanceResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetXagoBalanceResponse>(create);
  static GetXagoBalanceResponse? _defaultInstance;

  @$pb.TagNumber(2)
  $core.List<XagoBalance> get balances => $_getList(0);
}

class XagoBalance extends $pb.GeneratedMessage {
  factory XagoBalance({
    Amount? balance,
    Amount? available,
    $core.String? currency,
    $core.String? linkedAccount,
    $core.String? formattedBalance,
    $core.String? formattedAvailableBalance,
  }) {
    final $result = create();
    if (balance != null) {
      $result.balance = balance;
    }
    if (available != null) {
      $result.available = available;
    }
    if (currency != null) {
      $result.currency = currency;
    }
    if (linkedAccount != null) {
      $result.linkedAccount = linkedAccount;
    }
    if (formattedBalance != null) {
      $result.formattedBalance = formattedBalance;
    }
    if (formattedAvailableBalance != null) {
      $result.formattedAvailableBalance = formattedAvailableBalance;
    }
    return $result;
  }
  XagoBalance._() : super();
  factory XagoBalance.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory XagoBalance.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'XagoBalance', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<Amount>(1, _omitFieldNames ? '' : 'balance', subBuilder: Amount.create)
    ..aOM<Amount>(2, _omitFieldNames ? '' : 'available', subBuilder: Amount.create)
    ..aOS(3, _omitFieldNames ? '' : 'currency')
    ..aOS(4, _omitFieldNames ? '' : 'linkedAccount', protoName: 'linkedAccount')
    ..aOS(5, _omitFieldNames ? '' : 'formattedBalance', protoName: 'formattedBalance')
    ..aOS(6, _omitFieldNames ? '' : 'formattedAvailableBalance', protoName: 'formattedAvailableBalance')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  XagoBalance clone() => XagoBalance()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  XagoBalance copyWith(void Function(XagoBalance) updates) => super.copyWith((message) => updates(message as XagoBalance)) as XagoBalance;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static XagoBalance create() => XagoBalance._();
  XagoBalance createEmptyInstance() => create();
  static $pb.PbList<XagoBalance> createRepeated() => $pb.PbList<XagoBalance>();
  @$core.pragma('dart2js:noInline')
  static XagoBalance getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<XagoBalance>(create);
  static XagoBalance? _defaultInstance;

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

  @$pb.TagNumber(3)
  $core.String get currency => $_getSZ(2);
  @$pb.TagNumber(3)
  set currency($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCurrency() => $_has(2);
  @$pb.TagNumber(3)
  void clearCurrency() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get linkedAccount => $_getSZ(3);
  @$pb.TagNumber(4)
  set linkedAccount($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasLinkedAccount() => $_has(3);
  @$pb.TagNumber(4)
  void clearLinkedAccount() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get formattedBalance => $_getSZ(4);
  @$pb.TagNumber(5)
  set formattedBalance($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasFormattedBalance() => $_has(4);
  @$pb.TagNumber(5)
  void clearFormattedBalance() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get formattedAvailableBalance => $_getSZ(5);
  @$pb.TagNumber(6)
  set formattedAvailableBalance($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasFormattedAvailableBalance() => $_has(5);
  @$pb.TagNumber(6)
  void clearFormattedAvailableBalance() => clearField(6);
}

class WithdrawXagoBalanceRequest extends $pb.GeneratedMessage {
  factory WithdrawXagoBalanceRequest({
    $core.String? fromLinkedAccount,
    $core.String? toLinkedAccount,
    Amount? amount,
    $core.String? note,
  }) {
    final $result = create();
    if (fromLinkedAccount != null) {
      $result.fromLinkedAccount = fromLinkedAccount;
    }
    if (toLinkedAccount != null) {
      $result.toLinkedAccount = toLinkedAccount;
    }
    if (amount != null) {
      $result.amount = amount;
    }
    if (note != null) {
      $result.note = note;
    }
    return $result;
  }
  WithdrawXagoBalanceRequest._() : super();
  factory WithdrawXagoBalanceRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WithdrawXagoBalanceRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WithdrawXagoBalanceRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'fromLinkedAccount', protoName: 'fromLinkedAccount')
    ..aOS(2, _omitFieldNames ? '' : 'toLinkedAccount', protoName: 'toLinkedAccount')
    ..aOM<Amount>(3, _omitFieldNames ? '' : 'amount', subBuilder: Amount.create)
    ..aOS(4, _omitFieldNames ? '' : 'note')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WithdrawXagoBalanceRequest clone() => WithdrawXagoBalanceRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WithdrawXagoBalanceRequest copyWith(void Function(WithdrawXagoBalanceRequest) updates) => super.copyWith((message) => updates(message as WithdrawXagoBalanceRequest)) as WithdrawXagoBalanceRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WithdrawXagoBalanceRequest create() => WithdrawXagoBalanceRequest._();
  WithdrawXagoBalanceRequest createEmptyInstance() => create();
  static $pb.PbList<WithdrawXagoBalanceRequest> createRepeated() => $pb.PbList<WithdrawXagoBalanceRequest>();
  @$core.pragma('dart2js:noInline')
  static WithdrawXagoBalanceRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WithdrawXagoBalanceRequest>(create);
  static WithdrawXagoBalanceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get fromLinkedAccount => $_getSZ(0);
  @$pb.TagNumber(1)
  set fromLinkedAccount($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFromLinkedAccount() => $_has(0);
  @$pb.TagNumber(1)
  void clearFromLinkedAccount() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get toLinkedAccount => $_getSZ(1);
  @$pb.TagNumber(2)
  set toLinkedAccount($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasToLinkedAccount() => $_has(1);
  @$pb.TagNumber(2)
  void clearToLinkedAccount() => clearField(2);

  @$pb.TagNumber(3)
  Amount get amount => $_getN(2);
  @$pb.TagNumber(3)
  set amount(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearAmount() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureAmount() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get note => $_getSZ(3);
  @$pb.TagNumber(4)
  set note($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasNote() => $_has(3);
  @$pb.TagNumber(4)
  void clearNote() => clearField(4);
}

class AddXagoBalanceAccountRequest extends $pb.GeneratedMessage {
  factory AddXagoBalanceAccountRequest({
    $core.String? currencyCode,
    $core.String? nickname,
    $core.String? title,
  }) {
    final $result = create();
    if (currencyCode != null) {
      $result.currencyCode = currencyCode;
    }
    if (nickname != null) {
      $result.nickname = nickname;
    }
    if (title != null) {
      $result.title = title;
    }
    return $result;
  }
  AddXagoBalanceAccountRequest._() : super();
  factory AddXagoBalanceAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AddXagoBalanceAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AddXagoBalanceAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'currencyCode', protoName: 'currencyCode')
    ..aOS(2, _omitFieldNames ? '' : 'nickname')
    ..aOS(3, _omitFieldNames ? '' : 'title')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AddXagoBalanceAccountRequest clone() => AddXagoBalanceAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AddXagoBalanceAccountRequest copyWith(void Function(AddXagoBalanceAccountRequest) updates) => super.copyWith((message) => updates(message as AddXagoBalanceAccountRequest)) as AddXagoBalanceAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AddXagoBalanceAccountRequest create() => AddXagoBalanceAccountRequest._();
  AddXagoBalanceAccountRequest createEmptyInstance() => create();
  static $pb.PbList<AddXagoBalanceAccountRequest> createRepeated() => $pb.PbList<AddXagoBalanceAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static AddXagoBalanceAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AddXagoBalanceAccountRequest>(create);
  static AddXagoBalanceAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get currencyCode => $_getSZ(0);
  @$pb.TagNumber(1)
  set currencyCode($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCurrencyCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCurrencyCode() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get nickname => $_getSZ(1);
  @$pb.TagNumber(2)
  set nickname($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNickname() => $_has(1);
  @$pb.TagNumber(2)
  void clearNickname() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get title => $_getSZ(2);
  @$pb.TagNumber(3)
  set title($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTitle() => $_has(2);
  @$pb.TagNumber(3)
  void clearTitle() => clearField(3);
}

class AddXagoBankAccountRequest extends $pb.GeneratedMessage {
  factory AddXagoBankAccountRequest({
    $core.String? accountNumber,
    $core.String? branchCode,
    $core.String? bankName,
    $core.String? iban,
    $core.String? bic,
  }) {
    final $result = create();
    if (accountNumber != null) {
      $result.accountNumber = accountNumber;
    }
    if (branchCode != null) {
      $result.branchCode = branchCode;
    }
    if (bankName != null) {
      $result.bankName = bankName;
    }
    if (iban != null) {
      $result.iban = iban;
    }
    if (bic != null) {
      $result.bic = bic;
    }
    return $result;
  }
  AddXagoBankAccountRequest._() : super();
  factory AddXagoBankAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AddXagoBankAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AddXagoBankAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'accountNumber', protoName: 'accountNumber')
    ..aOS(2, _omitFieldNames ? '' : 'branchCode', protoName: 'branchCode')
    ..aOS(3, _omitFieldNames ? '' : 'bankName', protoName: 'bankName')
    ..aOS(4, _omitFieldNames ? '' : 'iban')
    ..aOS(5, _omitFieldNames ? '' : 'bic')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AddXagoBankAccountRequest clone() => AddXagoBankAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AddXagoBankAccountRequest copyWith(void Function(AddXagoBankAccountRequest) updates) => super.copyWith((message) => updates(message as AddXagoBankAccountRequest)) as AddXagoBankAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AddXagoBankAccountRequest create() => AddXagoBankAccountRequest._();
  AddXagoBankAccountRequest createEmptyInstance() => create();
  static $pb.PbList<AddXagoBankAccountRequest> createRepeated() => $pb.PbList<AddXagoBankAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static AddXagoBankAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AddXagoBankAccountRequest>(create);
  static AddXagoBankAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get accountNumber => $_getSZ(0);
  @$pb.TagNumber(1)
  set accountNumber($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAccountNumber() => $_has(0);
  @$pb.TagNumber(1)
  void clearAccountNumber() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get branchCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set branchCode($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBranchCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearBranchCode() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get bankName => $_getSZ(2);
  @$pb.TagNumber(3)
  set bankName($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBankName() => $_has(2);
  @$pb.TagNumber(3)
  void clearBankName() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get iban => $_getSZ(3);
  @$pb.TagNumber(4)
  set iban($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasIban() => $_has(3);
  @$pb.TagNumber(4)
  void clearIban() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get bic => $_getSZ(4);
  @$pb.TagNumber(5)
  set bic($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasBic() => $_has(4);
  @$pb.TagNumber(5)
  void clearBic() => clearField(5);
}

class SetDefaultSendLinkedAccountRequest extends $pb.GeneratedMessage {
  factory SetDefaultSendLinkedAccountRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  SetDefaultSendLinkedAccountRequest._() : super();
  factory SetDefaultSendLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetDefaultSendLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetDefaultSendLinkedAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetDefaultSendLinkedAccountRequest clone() => SetDefaultSendLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetDefaultSendLinkedAccountRequest copyWith(void Function(SetDefaultSendLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as SetDefaultSendLinkedAccountRequest)) as SetDefaultSendLinkedAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetDefaultSendLinkedAccountRequest create() => SetDefaultSendLinkedAccountRequest._();
  SetDefaultSendLinkedAccountRequest createEmptyInstance() => create();
  static $pb.PbList<SetDefaultSendLinkedAccountRequest> createRepeated() => $pb.PbList<SetDefaultSendLinkedAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static SetDefaultSendLinkedAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetDefaultSendLinkedAccountRequest>(create);
  static SetDefaultSendLinkedAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class SetDefaultReceiveLinkedAccountRequest extends $pb.GeneratedMessage {
  factory SetDefaultReceiveLinkedAccountRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  SetDefaultReceiveLinkedAccountRequest._() : super();
  factory SetDefaultReceiveLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetDefaultReceiveLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetDefaultReceiveLinkedAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetDefaultReceiveLinkedAccountRequest clone() => SetDefaultReceiveLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetDefaultReceiveLinkedAccountRequest copyWith(void Function(SetDefaultReceiveLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as SetDefaultReceiveLinkedAccountRequest)) as SetDefaultReceiveLinkedAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetDefaultReceiveLinkedAccountRequest create() => SetDefaultReceiveLinkedAccountRequest._();
  SetDefaultReceiveLinkedAccountRequest createEmptyInstance() => create();
  static $pb.PbList<SetDefaultReceiveLinkedAccountRequest> createRepeated() => $pb.PbList<SetDefaultReceiveLinkedAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static SetDefaultReceiveLinkedAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetDefaultReceiveLinkedAccountRequest>(create);
  static SetDefaultReceiveLinkedAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class SlackCallbackRequest extends $pb.GeneratedMessage {
  factory SlackCallbackRequest({
    $core.String? state,
    $core.String? code,
  }) {
    final $result = create();
    if (state != null) {
      $result.state = state;
    }
    if (code != null) {
      $result.code = code;
    }
    return $result;
  }
  SlackCallbackRequest._() : super();
  factory SlackCallbackRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SlackCallbackRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SlackCallbackRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'state')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SlackCallbackRequest clone() => SlackCallbackRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SlackCallbackRequest copyWith(void Function(SlackCallbackRequest) updates) => super.copyWith((message) => updates(message as SlackCallbackRequest)) as SlackCallbackRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SlackCallbackRequest create() => SlackCallbackRequest._();
  SlackCallbackRequest createEmptyInstance() => create();
  static $pb.PbList<SlackCallbackRequest> createRepeated() => $pb.PbList<SlackCallbackRequest>();
  @$core.pragma('dart2js:noInline')
  static SlackCallbackRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SlackCallbackRequest>(create);
  static SlackCallbackRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get state => $_getSZ(0);
  @$pb.TagNumber(1)
  set state($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => clearField(2);
}

class SlackCallbackResponse extends $pb.GeneratedMessage {
  factory SlackCallbackResponse({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  SlackCallbackResponse._() : super();
  factory SlackCallbackResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SlackCallbackResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SlackCallbackResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SlackCallbackResponse clone() => SlackCallbackResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SlackCallbackResponse copyWith(void Function(SlackCallbackResponse) updates) => super.copyWith((message) => updates(message as SlackCallbackResponse)) as SlackCallbackResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SlackCallbackResponse create() => SlackCallbackResponse._();
  SlackCallbackResponse createEmptyInstance() => create();
  static $pb.PbList<SlackCallbackResponse> createRepeated() => $pb.PbList<SlackCallbackResponse>();
  @$core.pragma('dart2js:noInline')
  static SlackCallbackResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SlackCallbackResponse>(create);
  static SlackCallbackResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class CreateSlackAuthURLResponse extends $pb.GeneratedMessage {
  factory CreateSlackAuthURLResponse({
    $core.String? url,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    return $result;
  }
  CreateSlackAuthURLResponse._() : super();
  factory CreateSlackAuthURLResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateSlackAuthURLResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateSlackAuthURLResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateSlackAuthURLResponse clone() => CreateSlackAuthURLResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateSlackAuthURLResponse copyWith(void Function(CreateSlackAuthURLResponse) updates) => super.copyWith((message) => updates(message as CreateSlackAuthURLResponse)) as CreateSlackAuthURLResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateSlackAuthURLResponse create() => CreateSlackAuthURLResponse._();
  CreateSlackAuthURLResponse createEmptyInstance() => create();
  static $pb.PbList<CreateSlackAuthURLResponse> createRepeated() => $pb.PbList<CreateSlackAuthURLResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateSlackAuthURLResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateSlackAuthURLResponse>(create);
  static CreateSlackAuthURLResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
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

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Amount', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
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

class Transaction extends $pb.GeneratedMessage {
  factory Transaction({
    $core.String? id,
    $core.String? type,
    Amount? amount,
    $core.String? source,
    $core.String? destination,
    $6.Timestamp? timestamp,
    $core.String? state,
    $core.String? foreignId,
    $core.String? receiverAccountId,
    $core.String? senderAccountId,
    $core.String? title,
    $core.String? formattedAmount,
    $core.String? formattedTime,
    $core.String? formattedDate,
    $core.String? subtotal,
    $core.String? fees,
    $core.String? accountTitle,
    $core.String? reference,
    $core.String? destinationIdentity,
    $core.String? destinationIdentityType,
    $core.int? refundState,
    $core.String? paymentProtectionAmount,
    $core.bool? hasPaymentProtection,
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
    if (state != null) {
      $result.state = state;
    }
    if (foreignId != null) {
      $result.foreignId = foreignId;
    }
    if (receiverAccountId != null) {
      $result.receiverAccountId = receiverAccountId;
    }
    if (senderAccountId != null) {
      $result.senderAccountId = senderAccountId;
    }
    if (title != null) {
      $result.title = title;
    }
    if (formattedAmount != null) {
      $result.formattedAmount = formattedAmount;
    }
    if (formattedTime != null) {
      $result.formattedTime = formattedTime;
    }
    if (formattedDate != null) {
      $result.formattedDate = formattedDate;
    }
    if (subtotal != null) {
      $result.subtotal = subtotal;
    }
    if (fees != null) {
      $result.fees = fees;
    }
    if (accountTitle != null) {
      $result.accountTitle = accountTitle;
    }
    if (reference != null) {
      $result.reference = reference;
    }
    if (destinationIdentity != null) {
      $result.destinationIdentity = destinationIdentity;
    }
    if (destinationIdentityType != null) {
      $result.destinationIdentityType = destinationIdentityType;
    }
    if (refundState != null) {
      $result.refundState = refundState;
    }
    if (paymentProtectionAmount != null) {
      $result.paymentProtectionAmount = paymentProtectionAmount;
    }
    if (hasPaymentProtection != null) {
      $result.hasPaymentProtection = hasPaymentProtection;
    }
    return $result;
  }
  Transaction._() : super();
  factory Transaction.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transaction.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Transaction', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOM<Amount>(3, _omitFieldNames ? '' : 'amount', subBuilder: Amount.create)
    ..aOS(4, _omitFieldNames ? '' : 'source')
    ..aOS(5, _omitFieldNames ? '' : 'destination')
    ..aOM<$6.Timestamp>(6, _omitFieldNames ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..aOS(7, _omitFieldNames ? '' : 'state')
    ..aOS(9, _omitFieldNames ? '' : 'foreignId', protoName: 'foreignId')
    ..aOS(10, _omitFieldNames ? '' : 'receiverAccountId', protoName: 'receiverAccountId')
    ..aOS(11, _omitFieldNames ? '' : 'senderAccountId', protoName: 'senderAccountId')
    ..aOS(12, _omitFieldNames ? '' : 'title')
    ..aOS(13, _omitFieldNames ? '' : 'formattedAmount', protoName: 'formattedAmount')
    ..aOS(14, _omitFieldNames ? '' : 'formattedTime', protoName: 'formattedTime')
    ..aOS(15, _omitFieldNames ? '' : 'formattedDate', protoName: 'formattedDate')
    ..aOS(16, _omitFieldNames ? '' : 'subtotal')
    ..aOS(17, _omitFieldNames ? '' : 'fees')
    ..aOS(18, _omitFieldNames ? '' : 'accountTitle', protoName: 'accountTitle')
    ..aOS(19, _omitFieldNames ? '' : 'reference')
    ..aOS(20, _omitFieldNames ? '' : 'destinationIdentity', protoName: 'destinationIdentity')
    ..aOS(21, _omitFieldNames ? '' : 'destinationIdentityType', protoName: 'destinationIdentityType')
    ..a<$core.int>(22, _omitFieldNames ? '' : 'refundState', $pb.PbFieldType.O3, protoName: 'refundState')
    ..aOS(23, _omitFieldNames ? '' : 'paymentProtectionAmount', protoName: 'paymentProtectionAmount')
    ..aOB(24, _omitFieldNames ? '' : 'hasPaymentProtection', protoName: 'hasPaymentProtection')
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
  Amount get amount => $_getN(2);
  @$pb.TagNumber(3)
  set amount(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearAmount() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureAmount() => $_ensure(2);

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
  $core.String get state => $_getSZ(6);
  @$pb.TagNumber(7)
  set state($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasState() => $_has(6);
  @$pb.TagNumber(7)
  void clearState() => clearField(7);

  @$pb.TagNumber(9)
  $core.String get foreignId => $_getSZ(7);
  @$pb.TagNumber(9)
  set foreignId($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(9)
  $core.bool hasForeignId() => $_has(7);
  @$pb.TagNumber(9)
  void clearForeignId() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get receiverAccountId => $_getSZ(8);
  @$pb.TagNumber(10)
  set receiverAccountId($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(10)
  $core.bool hasReceiverAccountId() => $_has(8);
  @$pb.TagNumber(10)
  void clearReceiverAccountId() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get senderAccountId => $_getSZ(9);
  @$pb.TagNumber(11)
  set senderAccountId($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(11)
  $core.bool hasSenderAccountId() => $_has(9);
  @$pb.TagNumber(11)
  void clearSenderAccountId() => clearField(11);

  /// Display
  @$pb.TagNumber(12)
  $core.String get title => $_getSZ(10);
  @$pb.TagNumber(12)
  set title($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(12)
  $core.bool hasTitle() => $_has(10);
  @$pb.TagNumber(12)
  void clearTitle() => clearField(12);

  @$pb.TagNumber(13)
  $core.String get formattedAmount => $_getSZ(11);
  @$pb.TagNumber(13)
  set formattedAmount($core.String v) { $_setString(11, v); }
  @$pb.TagNumber(13)
  $core.bool hasFormattedAmount() => $_has(11);
  @$pb.TagNumber(13)
  void clearFormattedAmount() => clearField(13);

  @$pb.TagNumber(14)
  $core.String get formattedTime => $_getSZ(12);
  @$pb.TagNumber(14)
  set formattedTime($core.String v) { $_setString(12, v); }
  @$pb.TagNumber(14)
  $core.bool hasFormattedTime() => $_has(12);
  @$pb.TagNumber(14)
  void clearFormattedTime() => clearField(14);

  @$pb.TagNumber(15)
  $core.String get formattedDate => $_getSZ(13);
  @$pb.TagNumber(15)
  set formattedDate($core.String v) { $_setString(13, v); }
  @$pb.TagNumber(15)
  $core.bool hasFormattedDate() => $_has(13);
  @$pb.TagNumber(15)
  void clearFormattedDate() => clearField(15);

  @$pb.TagNumber(16)
  $core.String get subtotal => $_getSZ(14);
  @$pb.TagNumber(16)
  set subtotal($core.String v) { $_setString(14, v); }
  @$pb.TagNumber(16)
  $core.bool hasSubtotal() => $_has(14);
  @$pb.TagNumber(16)
  void clearSubtotal() => clearField(16);

  @$pb.TagNumber(17)
  $core.String get fees => $_getSZ(15);
  @$pb.TagNumber(17)
  set fees($core.String v) { $_setString(15, v); }
  @$pb.TagNumber(17)
  $core.bool hasFees() => $_has(15);
  @$pb.TagNumber(17)
  void clearFees() => clearField(17);

  @$pb.TagNumber(18)
  $core.String get accountTitle => $_getSZ(16);
  @$pb.TagNumber(18)
  set accountTitle($core.String v) { $_setString(16, v); }
  @$pb.TagNumber(18)
  $core.bool hasAccountTitle() => $_has(16);
  @$pb.TagNumber(18)
  void clearAccountTitle() => clearField(18);

  @$pb.TagNumber(19)
  $core.String get reference => $_getSZ(17);
  @$pb.TagNumber(19)
  set reference($core.String v) { $_setString(17, v); }
  @$pb.TagNumber(19)
  $core.bool hasReference() => $_has(17);
  @$pb.TagNumber(19)
  void clearReference() => clearField(19);

  @$pb.TagNumber(20)
  $core.String get destinationIdentity => $_getSZ(18);
  @$pb.TagNumber(20)
  set destinationIdentity($core.String v) { $_setString(18, v); }
  @$pb.TagNumber(20)
  $core.bool hasDestinationIdentity() => $_has(18);
  @$pb.TagNumber(20)
  void clearDestinationIdentity() => clearField(20);

  @$pb.TagNumber(21)
  $core.String get destinationIdentityType => $_getSZ(19);
  @$pb.TagNumber(21)
  set destinationIdentityType($core.String v) { $_setString(19, v); }
  @$pb.TagNumber(21)
  $core.bool hasDestinationIdentityType() => $_has(19);
  @$pb.TagNumber(21)
  void clearDestinationIdentityType() => clearField(21);

  @$pb.TagNumber(22)
  $core.int get refundState => $_getIZ(20);
  @$pb.TagNumber(22)
  set refundState($core.int v) { $_setSignedInt32(20, v); }
  @$pb.TagNumber(22)
  $core.bool hasRefundState() => $_has(20);
  @$pb.TagNumber(22)
  void clearRefundState() => clearField(22);

  @$pb.TagNumber(23)
  $core.String get paymentProtectionAmount => $_getSZ(21);
  @$pb.TagNumber(23)
  set paymentProtectionAmount($core.String v) { $_setString(21, v); }
  @$pb.TagNumber(23)
  $core.bool hasPaymentProtectionAmount() => $_has(21);
  @$pb.TagNumber(23)
  void clearPaymentProtectionAmount() => clearField(23);

  @$pb.TagNumber(24)
  $core.bool get hasPaymentProtection => $_getBF(22);
  @$pb.TagNumber(24)
  set hasPaymentProtection($core.bool v) { $_setBool(22, v); }
  @$pb.TagNumber(24)
  $core.bool hasHasPaymentProtection() => $_has(22);
  @$pb.TagNumber(24)
  void clearHasPaymentProtection() => clearField(24);
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

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListTransactionsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
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

class ConfirmPaymentRequest extends $pb.GeneratedMessage {
  factory ConfirmPaymentRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  ConfirmPaymentRequest._() : super();
  factory ConfirmPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfirmPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConfirmPaymentRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfirmPaymentRequest clone() => ConfirmPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfirmPaymentRequest copyWith(void Function(ConfirmPaymentRequest) updates) => super.copyWith((message) => updates(message as ConfirmPaymentRequest)) as ConfirmPaymentRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConfirmPaymentRequest create() => ConfirmPaymentRequest._();
  ConfirmPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<ConfirmPaymentRequest> createRepeated() => $pb.PbList<ConfirmPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static ConfirmPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfirmPaymentRequest>(create);
  static ConfirmPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class GetPaymentRequest extends $pb.GeneratedMessage {
  factory GetPaymentRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetPaymentRequest._() : super();
  factory GetPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetPaymentRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPaymentRequest clone() => GetPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPaymentRequest copyWith(void Function(GetPaymentRequest) updates) => super.copyWith((message) => updates(message as GetPaymentRequest)) as GetPaymentRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPaymentRequest create() => GetPaymentRequest._();
  GetPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<GetPaymentRequest> createRepeated() => $pb.PbList<GetPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static GetPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetPaymentRequest>(create);
  static GetPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class UpdatePaymentRequest extends $pb.GeneratedMessage {
  factory UpdatePaymentRequest({
    $core.String? id,
    Amount? senderAmount,
    Amount? receiverAmount,
    $core.String? note,
    $core.String? senderAccount,
    $core.String? receiverAccount,
    $core.String? receiverIdentity,
    $core.int? receiverIdentityType,
    $core.String? threeDSID,
    $core.String? otp,
    $core.String? ipAddress,
    $core.bool? addPaymentProtection,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (senderAmount != null) {
      $result.senderAmount = senderAmount;
    }
    if (receiverAmount != null) {
      $result.receiverAmount = receiverAmount;
    }
    if (note != null) {
      $result.note = note;
    }
    if (senderAccount != null) {
      $result.senderAccount = senderAccount;
    }
    if (receiverAccount != null) {
      $result.receiverAccount = receiverAccount;
    }
    if (receiverIdentity != null) {
      $result.receiverIdentity = receiverIdentity;
    }
    if (receiverIdentityType != null) {
      $result.receiverIdentityType = receiverIdentityType;
    }
    if (threeDSID != null) {
      $result.threeDSID = threeDSID;
    }
    if (otp != null) {
      $result.otp = otp;
    }
    if (ipAddress != null) {
      $result.ipAddress = ipAddress;
    }
    if (addPaymentProtection != null) {
      $result.addPaymentProtection = addPaymentProtection;
    }
    return $result;
  }
  UpdatePaymentRequest._() : super();
  factory UpdatePaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdatePaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdatePaymentRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<Amount>(2, _omitFieldNames ? '' : 'senderAmount', protoName: 'senderAmount', subBuilder: Amount.create)
    ..aOM<Amount>(3, _omitFieldNames ? '' : 'receiverAmount', protoName: 'receiverAmount', subBuilder: Amount.create)
    ..aOS(4, _omitFieldNames ? '' : 'note')
    ..aOS(5, _omitFieldNames ? '' : 'senderAccount', protoName: 'senderAccount')
    ..aOS(6, _omitFieldNames ? '' : 'receiverAccount', protoName: 'receiverAccount')
    ..aOS(7, _omitFieldNames ? '' : 'receiverIdentity', protoName: 'receiverIdentity')
    ..a<$core.int>(8, _omitFieldNames ? '' : 'receiverIdentityType', $pb.PbFieldType.O3, protoName: 'receiverIdentityType')
    ..aOS(9, _omitFieldNames ? '' : 'threeDSID', protoName: 'threeDSID')
    ..aOS(10, _omitFieldNames ? '' : 'otp')
    ..aOS(11, _omitFieldNames ? '' : 'ipAddress', protoName: 'ipAddress')
    ..aOB(12, _omitFieldNames ? '' : 'addPaymentProtection', protoName: 'addPaymentProtection')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdatePaymentRequest clone() => UpdatePaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdatePaymentRequest copyWith(void Function(UpdatePaymentRequest) updates) => super.copyWith((message) => updates(message as UpdatePaymentRequest)) as UpdatePaymentRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdatePaymentRequest create() => UpdatePaymentRequest._();
  UpdatePaymentRequest createEmptyInstance() => create();
  static $pb.PbList<UpdatePaymentRequest> createRepeated() => $pb.PbList<UpdatePaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdatePaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdatePaymentRequest>(create);
  static UpdatePaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  Amount get senderAmount => $_getN(1);
  @$pb.TagNumber(2)
  set senderAmount(Amount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasSenderAmount() => $_has(1);
  @$pb.TagNumber(2)
  void clearSenderAmount() => clearField(2);
  @$pb.TagNumber(2)
  Amount ensureSenderAmount() => $_ensure(1);

  @$pb.TagNumber(3)
  Amount get receiverAmount => $_getN(2);
  @$pb.TagNumber(3)
  set receiverAmount(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasReceiverAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearReceiverAmount() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureReceiverAmount() => $_ensure(2);

  @$pb.TagNumber(4)
  $core.String get note => $_getSZ(3);
  @$pb.TagNumber(4)
  set note($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasNote() => $_has(3);
  @$pb.TagNumber(4)
  void clearNote() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get senderAccount => $_getSZ(4);
  @$pb.TagNumber(5)
  set senderAccount($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSenderAccount() => $_has(4);
  @$pb.TagNumber(5)
  void clearSenderAccount() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get receiverAccount => $_getSZ(5);
  @$pb.TagNumber(6)
  set receiverAccount($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasReceiverAccount() => $_has(5);
  @$pb.TagNumber(6)
  void clearReceiverAccount() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get receiverIdentity => $_getSZ(6);
  @$pb.TagNumber(7)
  set receiverIdentity($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasReceiverIdentity() => $_has(6);
  @$pb.TagNumber(7)
  void clearReceiverIdentity() => clearField(7);

  @$pb.TagNumber(8)
  $core.int get receiverIdentityType => $_getIZ(7);
  @$pb.TagNumber(8)
  set receiverIdentityType($core.int v) { $_setSignedInt32(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasReceiverIdentityType() => $_has(7);
  @$pb.TagNumber(8)
  void clearReceiverIdentityType() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get threeDSID => $_getSZ(8);
  @$pb.TagNumber(9)
  set threeDSID($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasThreeDSID() => $_has(8);
  @$pb.TagNumber(9)
  void clearThreeDSID() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get otp => $_getSZ(9);
  @$pb.TagNumber(10)
  set otp($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasOtp() => $_has(9);
  @$pb.TagNumber(10)
  void clearOtp() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get ipAddress => $_getSZ(10);
  @$pb.TagNumber(11)
  set ipAddress($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasIpAddress() => $_has(10);
  @$pb.TagNumber(11)
  void clearIpAddress() => clearField(11);

  @$pb.TagNumber(12)
  $core.bool get addPaymentProtection => $_getBF(11);
  @$pb.TagNumber(12)
  set addPaymentProtection($core.bool v) { $_setBool(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasAddPaymentProtection() => $_has(11);
  @$pb.TagNumber(12)
  void clearAddPaymentProtection() => clearField(12);
}

class Payment extends $pb.GeneratedMessage {
  factory Payment({
    $core.String? id,
    $core.String? publicID,
    $core.int? state,
    $core.String? receiverWalletUrl,
    $core.String? receiverIdentity,
    $core.int? receiverIdentityType,
    Amount? senderAmount,
    $core.String? senderAccount,
    $core.String? note,
    $core.Iterable<$core.int>? requiredActions,
    $core.bool? hasPaymentProtection,
    $core.String? paymentProtectionAmount,
    $core.String? fxRate,
    Amount? receiverAmount,
    $core.String? totalSendAmount,
    $core.String? receiverLinkedAccountCountryCode,
    $core.String? formattedFees,
    $core.String? receiverAccount,
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
    if (hasPaymentProtection != null) {
      $result.hasPaymentProtection = hasPaymentProtection;
    }
    if (paymentProtectionAmount != null) {
      $result.paymentProtectionAmount = paymentProtectionAmount;
    }
    if (fxRate != null) {
      $result.fxRate = fxRate;
    }
    if (receiverAmount != null) {
      $result.receiverAmount = receiverAmount;
    }
    if (totalSendAmount != null) {
      $result.totalSendAmount = totalSendAmount;
    }
    if (receiverLinkedAccountCountryCode != null) {
      $result.receiverLinkedAccountCountryCode = receiverLinkedAccountCountryCode;
    }
    if (formattedFees != null) {
      $result.formattedFees = formattedFees;
    }
    if (receiverAccount != null) {
      $result.receiverAccount = receiverAccount;
    }
    return $result;
  }
  Payment._() : super();
  factory Payment.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Payment.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Payment', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'publicID', protoName: 'publicID')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'state', $pb.PbFieldType.O3)
    ..aOS(4, _omitFieldNames ? '' : 'receiverWalletUrl', protoName: 'receiverWalletUrl')
    ..aOS(5, _omitFieldNames ? '' : 'receiverIdentity', protoName: 'receiverIdentity')
    ..a<$core.int>(6, _omitFieldNames ? '' : 'receiverIdentityType', $pb.PbFieldType.O3, protoName: 'receiverIdentityType')
    ..aOM<Amount>(7, _omitFieldNames ? '' : 'senderAmount', protoName: 'senderAmount', subBuilder: Amount.create)
    ..aOS(8, _omitFieldNames ? '' : 'senderAccount', protoName: 'senderAccount')
    ..aOS(9, _omitFieldNames ? '' : 'note')
    ..p<$core.int>(10, _omitFieldNames ? '' : 'requiredActions', $pb.PbFieldType.K3, protoName: 'requiredActions')
    ..aOB(11, _omitFieldNames ? '' : 'hasPaymentProtection', protoName: 'hasPaymentProtection')
    ..aOS(12, _omitFieldNames ? '' : 'paymentProtectionAmount', protoName: 'paymentProtectionAmount')
    ..aOS(13, _omitFieldNames ? '' : 'fxRate', protoName: 'fxRate')
    ..aOM<Amount>(14, _omitFieldNames ? '' : 'receiverAmount', protoName: 'receiverAmount', subBuilder: Amount.create)
    ..aOS(15, _omitFieldNames ? '' : 'totalSendAmount', protoName: 'totalSendAmount')
    ..aOS(16, _omitFieldNames ? '' : 'receiverLinkedAccountCountryCode', protoName: 'receiverLinkedAccountCountryCode')
    ..aOS(17, _omitFieldNames ? '' : 'formattedFees', protoName: 'formattedFees')
    ..aOS(18, _omitFieldNames ? '' : 'receiverAccount', protoName: 'receiverAccount')
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
  $core.int get state => $_getIZ(2);
  @$pb.TagNumber(3)
  set state($core.int v) { $_setSignedInt32(2, v); }
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
  $core.int get receiverIdentityType => $_getIZ(5);
  @$pb.TagNumber(6)
  set receiverIdentityType($core.int v) { $_setSignedInt32(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasReceiverIdentityType() => $_has(5);
  @$pb.TagNumber(6)
  void clearReceiverIdentityType() => clearField(6);

  @$pb.TagNumber(7)
  Amount get senderAmount => $_getN(6);
  @$pb.TagNumber(7)
  set senderAmount(Amount v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasSenderAmount() => $_has(6);
  @$pb.TagNumber(7)
  void clearSenderAmount() => clearField(7);
  @$pb.TagNumber(7)
  Amount ensureSenderAmount() => $_ensure(6);

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
  $core.List<$core.int> get requiredActions => $_getList(9);

  @$pb.TagNumber(11)
  $core.bool get hasPaymentProtection => $_getBF(10);
  @$pb.TagNumber(11)
  set hasPaymentProtection($core.bool v) { $_setBool(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasHasPaymentProtection() => $_has(10);
  @$pb.TagNumber(11)
  void clearHasPaymentProtection() => clearField(11);

  @$pb.TagNumber(12)
  $core.String get paymentProtectionAmount => $_getSZ(11);
  @$pb.TagNumber(12)
  set paymentProtectionAmount($core.String v) { $_setString(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasPaymentProtectionAmount() => $_has(11);
  @$pb.TagNumber(12)
  void clearPaymentProtectionAmount() => clearField(12);

  @$pb.TagNumber(13)
  $core.String get fxRate => $_getSZ(12);
  @$pb.TagNumber(13)
  set fxRate($core.String v) { $_setString(12, v); }
  @$pb.TagNumber(13)
  $core.bool hasFxRate() => $_has(12);
  @$pb.TagNumber(13)
  void clearFxRate() => clearField(13);

  @$pb.TagNumber(14)
  Amount get receiverAmount => $_getN(13);
  @$pb.TagNumber(14)
  set receiverAmount(Amount v) { setField(14, v); }
  @$pb.TagNumber(14)
  $core.bool hasReceiverAmount() => $_has(13);
  @$pb.TagNumber(14)
  void clearReceiverAmount() => clearField(14);
  @$pb.TagNumber(14)
  Amount ensureReceiverAmount() => $_ensure(13);

  @$pb.TagNumber(15)
  $core.String get totalSendAmount => $_getSZ(14);
  @$pb.TagNumber(15)
  set totalSendAmount($core.String v) { $_setString(14, v); }
  @$pb.TagNumber(15)
  $core.bool hasTotalSendAmount() => $_has(14);
  @$pb.TagNumber(15)
  void clearTotalSendAmount() => clearField(15);

  @$pb.TagNumber(16)
  $core.String get receiverLinkedAccountCountryCode => $_getSZ(15);
  @$pb.TagNumber(16)
  set receiverLinkedAccountCountryCode($core.String v) { $_setString(15, v); }
  @$pb.TagNumber(16)
  $core.bool hasReceiverLinkedAccountCountryCode() => $_has(15);
  @$pb.TagNumber(16)
  void clearReceiverLinkedAccountCountryCode() => clearField(16);

  @$pb.TagNumber(17)
  $core.String get formattedFees => $_getSZ(16);
  @$pb.TagNumber(17)
  set formattedFees($core.String v) { $_setString(16, v); }
  @$pb.TagNumber(17)
  $core.bool hasFormattedFees() => $_has(16);
  @$pb.TagNumber(17)
  void clearFormattedFees() => clearField(17);

  @$pb.TagNumber(18)
  $core.String get receiverAccount => $_getSZ(17);
  @$pb.TagNumber(18)
  set receiverAccount($core.String v) { $_setString(17, v); }
  @$pb.TagNumber(18)
  $core.bool hasReceiverAccount() => $_has(17);
  @$pb.TagNumber(18)
  void clearReceiverAccount() => clearField(18);
}

class CreatePaymentRequest extends $pb.GeneratedMessage {
  factory CreatePaymentRequest({
    Amount? senderAmount,
    Amount? receiverAmount,
    $core.String? receiverIdentity,
    $core.int? receiverIdentityType,
    $core.String? senderAccount,
    $core.String? receiverAccount,
    $core.String? note,
    $core.String? ipAddress,
    $core.bool? addPaymentProtection,
  }) {
    final $result = create();
    if (senderAmount != null) {
      $result.senderAmount = senderAmount;
    }
    if (receiverAmount != null) {
      $result.receiverAmount = receiverAmount;
    }
    if (receiverIdentity != null) {
      $result.receiverIdentity = receiverIdentity;
    }
    if (receiverIdentityType != null) {
      $result.receiverIdentityType = receiverIdentityType;
    }
    if (senderAccount != null) {
      $result.senderAccount = senderAccount;
    }
    if (receiverAccount != null) {
      $result.receiverAccount = receiverAccount;
    }
    if (note != null) {
      $result.note = note;
    }
    if (ipAddress != null) {
      $result.ipAddress = ipAddress;
    }
    if (addPaymentProtection != null) {
      $result.addPaymentProtection = addPaymentProtection;
    }
    return $result;
  }
  CreatePaymentRequest._() : super();
  factory CreatePaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreatePaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreatePaymentRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<Amount>(1, _omitFieldNames ? '' : 'senderAmount', protoName: 'senderAmount', subBuilder: Amount.create)
    ..aOM<Amount>(2, _omitFieldNames ? '' : 'receiverAmount', protoName: 'receiverAmount', subBuilder: Amount.create)
    ..aOS(3, _omitFieldNames ? '' : 'receiverIdentity', protoName: 'receiverIdentity')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'receiverIdentityType', $pb.PbFieldType.O3, protoName: 'receiverIdentityType')
    ..aOS(5, _omitFieldNames ? '' : 'senderAccount', protoName: 'senderAccount')
    ..aOS(6, _omitFieldNames ? '' : 'receiverAccount', protoName: 'receiverAccount')
    ..aOS(7, _omitFieldNames ? '' : 'note')
    ..aOS(8, _omitFieldNames ? '' : 'ipAddress', protoName: 'ipAddress')
    ..aOB(9, _omitFieldNames ? '' : 'addPaymentProtection', protoName: 'addPaymentProtection')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreatePaymentRequest clone() => CreatePaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreatePaymentRequest copyWith(void Function(CreatePaymentRequest) updates) => super.copyWith((message) => updates(message as CreatePaymentRequest)) as CreatePaymentRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreatePaymentRequest create() => CreatePaymentRequest._();
  CreatePaymentRequest createEmptyInstance() => create();
  static $pb.PbList<CreatePaymentRequest> createRepeated() => $pb.PbList<CreatePaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static CreatePaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreatePaymentRequest>(create);
  static CreatePaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  Amount get senderAmount => $_getN(0);
  @$pb.TagNumber(1)
  set senderAmount(Amount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasSenderAmount() => $_has(0);
  @$pb.TagNumber(1)
  void clearSenderAmount() => clearField(1);
  @$pb.TagNumber(1)
  Amount ensureSenderAmount() => $_ensure(0);

  @$pb.TagNumber(2)
  Amount get receiverAmount => $_getN(1);
  @$pb.TagNumber(2)
  set receiverAmount(Amount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasReceiverAmount() => $_has(1);
  @$pb.TagNumber(2)
  void clearReceiverAmount() => clearField(2);
  @$pb.TagNumber(2)
  Amount ensureReceiverAmount() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get receiverIdentity => $_getSZ(2);
  @$pb.TagNumber(3)
  set receiverIdentity($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasReceiverIdentity() => $_has(2);
  @$pb.TagNumber(3)
  void clearReceiverIdentity() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get receiverIdentityType => $_getIZ(3);
  @$pb.TagNumber(4)
  set receiverIdentityType($core.int v) { $_setSignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasReceiverIdentityType() => $_has(3);
  @$pb.TagNumber(4)
  void clearReceiverIdentityType() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get senderAccount => $_getSZ(4);
  @$pb.TagNumber(5)
  set senderAccount($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasSenderAccount() => $_has(4);
  @$pb.TagNumber(5)
  void clearSenderAccount() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get receiverAccount => $_getSZ(5);
  @$pb.TagNumber(6)
  set receiverAccount($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasReceiverAccount() => $_has(5);
  @$pb.TagNumber(6)
  void clearReceiverAccount() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get note => $_getSZ(6);
  @$pb.TagNumber(7)
  set note($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasNote() => $_has(6);
  @$pb.TagNumber(7)
  void clearNote() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get ipAddress => $_getSZ(7);
  @$pb.TagNumber(8)
  set ipAddress($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasIpAddress() => $_has(7);
  @$pb.TagNumber(8)
  void clearIpAddress() => clearField(8);

  @$pb.TagNumber(9)
  $core.bool get addPaymentProtection => $_getBF(8);
  @$pb.TagNumber(9)
  set addPaymentProtection($core.bool v) { $_setBool(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasAddPaymentProtection() => $_has(8);
  @$pb.TagNumber(9)
  void clearAddPaymentProtection() => clearField(9);
}

class GetCardDetailsRequest extends $pb.GeneratedMessage {
  factory GetCardDetailsRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetCardDetailsRequest._() : super();
  factory GetCardDetailsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetCardDetailsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetCardDetailsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetCardDetailsRequest clone() => GetCardDetailsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetCardDetailsRequest copyWith(void Function(GetCardDetailsRequest) updates) => super.copyWith((message) => updates(message as GetCardDetailsRequest)) as GetCardDetailsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetCardDetailsRequest create() => GetCardDetailsRequest._();
  GetCardDetailsRequest createEmptyInstance() => create();
  static $pb.PbList<GetCardDetailsRequest> createRepeated() => $pb.PbList<GetCardDetailsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetCardDetailsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetCardDetailsRequest>(create);
  static GetCardDetailsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class CardDetails extends $pb.GeneratedMessage {
  factory CardDetails({
    $core.String? id,
    $core.String? network,
    $core.String? bin,
    $core.String? last4,
    $core.String? type,
    $core.String? expiration,
    $core.String? nickname,
    $core.String? state,
    $core.bool? canSend,
    $core.bool? canReceive,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (network != null) {
      $result.network = network;
    }
    if (bin != null) {
      $result.bin = bin;
    }
    if (last4 != null) {
      $result.last4 = last4;
    }
    if (type != null) {
      $result.type = type;
    }
    if (expiration != null) {
      $result.expiration = expiration;
    }
    if (nickname != null) {
      $result.nickname = nickname;
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
    return $result;
  }
  CardDetails._() : super();
  factory CardDetails.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CardDetails.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CardDetails', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'network')
    ..aOS(3, _omitFieldNames ? '' : 'bin')
    ..aOS(4, _omitFieldNames ? '' : 'last4')
    ..aOS(5, _omitFieldNames ? '' : 'type')
    ..aOS(6, _omitFieldNames ? '' : 'expiration')
    ..aOS(7, _omitFieldNames ? '' : 'nickname')
    ..aOS(8, _omitFieldNames ? '' : 'state')
    ..aOB(9, _omitFieldNames ? '' : 'canSend', protoName: 'canSend')
    ..aOB(10, _omitFieldNames ? '' : 'canReceive', protoName: 'canReceive')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CardDetails clone() => CardDetails()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CardDetails copyWith(void Function(CardDetails) updates) => super.copyWith((message) => updates(message as CardDetails)) as CardDetails;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CardDetails create() => CardDetails._();
  CardDetails createEmptyInstance() => create();
  static $pb.PbList<CardDetails> createRepeated() => $pb.PbList<CardDetails>();
  @$core.pragma('dart2js:noInline')
  static CardDetails getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CardDetails>(create);
  static CardDetails? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get network => $_getSZ(1);
  @$pb.TagNumber(2)
  set network($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNetwork() => $_has(1);
  @$pb.TagNumber(2)
  void clearNetwork() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get bin => $_getSZ(2);
  @$pb.TagNumber(3)
  set bin($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBin() => $_has(2);
  @$pb.TagNumber(3)
  void clearBin() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get last4 => $_getSZ(3);
  @$pb.TagNumber(4)
  set last4($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasLast4() => $_has(3);
  @$pb.TagNumber(4)
  void clearLast4() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get type => $_getSZ(4);
  @$pb.TagNumber(5)
  set type($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasType() => $_has(4);
  @$pb.TagNumber(5)
  void clearType() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get expiration => $_getSZ(5);
  @$pb.TagNumber(6)
  set expiration($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasExpiration() => $_has(5);
  @$pb.TagNumber(6)
  void clearExpiration() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get nickname => $_getSZ(6);
  @$pb.TagNumber(7)
  set nickname($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasNickname() => $_has(6);
  @$pb.TagNumber(7)
  void clearNickname() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get state => $_getSZ(7);
  @$pb.TagNumber(8)
  set state($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasState() => $_has(7);
  @$pb.TagNumber(8)
  void clearState() => clearField(8);

  @$pb.TagNumber(9)
  $core.bool get canSend => $_getBF(8);
  @$pb.TagNumber(9)
  set canSend($core.bool v) { $_setBool(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasCanSend() => $_has(8);
  @$pb.TagNumber(9)
  void clearCanSend() => clearField(9);

  @$pb.TagNumber(10)
  $core.bool get canReceive => $_getBF(9);
  @$pb.TagNumber(10)
  set canReceive($core.bool v) { $_setBool(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasCanReceive() => $_has(9);
  @$pb.TagNumber(10)
  void clearCanReceive() => clearField(10);
}

class SearchWalletsRequest extends $pb.GeneratedMessage {
  factory SearchWalletsRequest({
    $core.String? term,
  }) {
    final $result = create();
    if (term != null) {
      $result.term = term;
    }
    return $result;
  }
  SearchWalletsRequest._() : super();
  factory SearchWalletsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SearchWalletsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SearchWalletsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'term')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SearchWalletsRequest clone() => SearchWalletsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SearchWalletsRequest copyWith(void Function(SearchWalletsRequest) updates) => super.copyWith((message) => updates(message as SearchWalletsRequest)) as SearchWalletsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SearchWalletsRequest create() => SearchWalletsRequest._();
  SearchWalletsRequest createEmptyInstance() => create();
  static $pb.PbList<SearchWalletsRequest> createRepeated() => $pb.PbList<SearchWalletsRequest>();
  @$core.pragma('dart2js:noInline')
  static SearchWalletsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SearchWalletsRequest>(create);
  static SearchWalletsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get term => $_getSZ(0);
  @$pb.TagNumber(1)
  set term($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTerm() => $_has(0);
  @$pb.TagNumber(1)
  void clearTerm() => clearField(1);
}

class SearchWalletsResponse extends $pb.GeneratedMessage {
  factory SearchWalletsResponse({
    $core.Iterable<SearchResult>? results,
  }) {
    final $result = create();
    if (results != null) {
      $result.results.addAll(results);
    }
    return $result;
  }
  SearchWalletsResponse._() : super();
  factory SearchWalletsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SearchWalletsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SearchWalletsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<SearchResult>(1, _omitFieldNames ? '' : 'results', $pb.PbFieldType.PM, subBuilder: SearchResult.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SearchWalletsResponse clone() => SearchWalletsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SearchWalletsResponse copyWith(void Function(SearchWalletsResponse) updates) => super.copyWith((message) => updates(message as SearchWalletsResponse)) as SearchWalletsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SearchWalletsResponse create() => SearchWalletsResponse._();
  SearchWalletsResponse createEmptyInstance() => create();
  static $pb.PbList<SearchWalletsResponse> createRepeated() => $pb.PbList<SearchWalletsResponse>();
  @$core.pragma('dart2js:noInline')
  static SearchWalletsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SearchWalletsResponse>(create);
  static SearchWalletsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<SearchResult> get results => $_getList(0);
}

class SearchResult extends $pb.GeneratedMessage {
  factory SearchResult({
    $core.String? walletID,
    $core.String? identifier,
    $core.String? identifierType,
  @$core.Deprecated('This field is deprecated.')
    $core.bool? canSend,
    $core.String? walletUrl,
    $core.Iterable<SearchResult>? subResults,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (identifier != null) {
      $result.identifier = identifier;
    }
    if (identifierType != null) {
      $result.identifierType = identifierType;
    }
    if (canSend != null) {
      // ignore: deprecated_member_use_from_same_package
      $result.canSend = canSend;
    }
    if (walletUrl != null) {
      $result.walletUrl = walletUrl;
    }
    if (subResults != null) {
      $result.subResults.addAll(subResults);
    }
    return $result;
  }
  SearchResult._() : super();
  factory SearchResult.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SearchResult.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SearchResult', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(2, _omitFieldNames ? '' : 'identifier')
    ..aOS(3, _omitFieldNames ? '' : 'identifierType', protoName: 'identifierType')
    ..aOB(4, _omitFieldNames ? '' : 'canSend', protoName: 'canSend')
    ..aOS(5, _omitFieldNames ? '' : 'walletUrl', protoName: 'walletUrl')
    ..pc<SearchResult>(6, _omitFieldNames ? '' : 'subResults', $pb.PbFieldType.PM, protoName: 'subResults', subBuilder: SearchResult.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SearchResult clone() => SearchResult()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SearchResult copyWith(void Function(SearchResult) updates) => super.copyWith((message) => updates(message as SearchResult)) as SearchResult;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SearchResult create() => SearchResult._();
  SearchResult createEmptyInstance() => create();
  static $pb.PbList<SearchResult> createRepeated() => $pb.PbList<SearchResult>();
  @$core.pragma('dart2js:noInline')
  static SearchResult getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SearchResult>(create);
  static SearchResult? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get identifier => $_getSZ(1);
  @$pb.TagNumber(2)
  set identifier($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasIdentifier() => $_has(1);
  @$pb.TagNumber(2)
  void clearIdentifier() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get identifierType => $_getSZ(2);
  @$pb.TagNumber(3)
  set identifierType($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasIdentifierType() => $_has(2);
  @$pb.TagNumber(3)
  void clearIdentifierType() => clearField(3);

  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  $core.bool get canSend => $_getBF(3);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  set canSend($core.bool v) { $_setBool(3, v); }
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  $core.bool hasCanSend() => $_has(3);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  void clearCanSend() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get walletUrl => $_getSZ(4);
  @$pb.TagNumber(5)
  set walletUrl($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasWalletUrl() => $_has(4);
  @$pb.TagNumber(5)
  void clearWalletUrl() => clearField(5);

  @$pb.TagNumber(6)
  $core.List<SearchResult> get subResults => $_getList(5);
}

class GetPublicWalletInfoRequest extends $pb.GeneratedMessage {
  factory GetPublicWalletInfoRequest({
    $core.String? walletAddress,
  }) {
    final $result = create();
    if (walletAddress != null) {
      $result.walletAddress = walletAddress;
    }
    return $result;
  }
  GetPublicWalletInfoRequest._() : super();
  factory GetPublicWalletInfoRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPublicWalletInfoRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetPublicWalletInfoRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletAddress', protoName: 'walletAddress')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPublicWalletInfoRequest clone() => GetPublicWalletInfoRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPublicWalletInfoRequest copyWith(void Function(GetPublicWalletInfoRequest) updates) => super.copyWith((message) => updates(message as GetPublicWalletInfoRequest)) as GetPublicWalletInfoRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPublicWalletInfoRequest create() => GetPublicWalletInfoRequest._();
  GetPublicWalletInfoRequest createEmptyInstance() => create();
  static $pb.PbList<GetPublicWalletInfoRequest> createRepeated() => $pb.PbList<GetPublicWalletInfoRequest>();
  @$core.pragma('dart2js:noInline')
  static GetPublicWalletInfoRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetPublicWalletInfoRequest>(create);
  static GetPublicWalletInfoRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletAddress() => clearField(1);
}

class PublicWalletInfo extends $pb.GeneratedMessage {
  factory PublicWalletInfo({
    $core.String? walletID,
    $core.String? address,
    $core.String? shortAddress,
    $core.String? publicName,
    $core.Iterable<Identity>? identities,
    $core.bool? canReceive,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (address != null) {
      $result.address = address;
    }
    if (shortAddress != null) {
      $result.shortAddress = shortAddress;
    }
    if (publicName != null) {
      $result.publicName = publicName;
    }
    if (identities != null) {
      $result.identities.addAll(identities);
    }
    if (canReceive != null) {
      $result.canReceive = canReceive;
    }
    return $result;
  }
  PublicWalletInfo._() : super();
  factory PublicWalletInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PublicWalletInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PublicWalletInfo', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(2, _omitFieldNames ? '' : 'address')
    ..aOS(3, _omitFieldNames ? '' : 'shortAddress', protoName: 'shortAddress')
    ..aOS(4, _omitFieldNames ? '' : 'publicName', protoName: 'publicName')
    ..pc<Identity>(5, _omitFieldNames ? '' : 'identities', $pb.PbFieldType.PM, subBuilder: Identity.create)
    ..aOB(6, _omitFieldNames ? '' : 'canReceive', protoName: 'canReceive')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PublicWalletInfo clone() => PublicWalletInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PublicWalletInfo copyWith(void Function(PublicWalletInfo) updates) => super.copyWith((message) => updates(message as PublicWalletInfo)) as PublicWalletInfo;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PublicWalletInfo create() => PublicWalletInfo._();
  PublicWalletInfo createEmptyInstance() => create();
  static $pb.PbList<PublicWalletInfo> createRepeated() => $pb.PbList<PublicWalletInfo>();
  @$core.pragma('dart2js:noInline')
  static PublicWalletInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PublicWalletInfo>(create);
  static PublicWalletInfo? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get address => $_getSZ(1);
  @$pb.TagNumber(2)
  set address($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearAddress() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get shortAddress => $_getSZ(2);
  @$pb.TagNumber(3)
  set shortAddress($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasShortAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearShortAddress() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get publicName => $_getSZ(3);
  @$pb.TagNumber(4)
  set publicName($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasPublicName() => $_has(3);
  @$pb.TagNumber(4)
  void clearPublicName() => clearField(4);

  @$pb.TagNumber(5)
  $core.List<Identity> get identities => $_getList(4);

  @$pb.TagNumber(6)
  $core.bool get canReceive => $_getBF(5);
  @$pb.TagNumber(6)
  set canReceive($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasCanReceive() => $_has(5);
  @$pb.TagNumber(6)
  void clearCanReceive() => clearField(6);
}

class WalletInfo extends $pb.GeneratedMessage {
  factory WalletInfo({
    $core.String? walletID,
    $core.String? url,
    $core.String? formattedURL,
    $core.bool? hasCard,
    $core.bool? hasBank,
    $core.bool? hasIdentities,
    $core.bool? hasTransacted,
    $core.bool? hasWalletAddress,
    $core.bool? hasBalances,
    $core.bool? exceededLimits,
  }) {
    final $result = create();
    if (walletID != null) {
      $result.walletID = walletID;
    }
    if (url != null) {
      $result.url = url;
    }
    if (formattedURL != null) {
      $result.formattedURL = formattedURL;
    }
    if (hasCard != null) {
      $result.hasCard = hasCard;
    }
    if (hasBank != null) {
      $result.hasBank = hasBank;
    }
    if (hasIdentities != null) {
      $result.hasIdentities = hasIdentities;
    }
    if (hasTransacted != null) {
      $result.hasTransacted = hasTransacted;
    }
    if (hasWalletAddress != null) {
      $result.hasWalletAddress = hasWalletAddress;
    }
    if (hasBalances != null) {
      $result.hasBalances = hasBalances;
    }
    if (exceededLimits != null) {
      $result.exceededLimits = exceededLimits;
    }
    return $result;
  }
  WalletInfo._() : super();
  factory WalletInfo.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WalletInfo.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WalletInfo', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletID', protoName: 'walletID')
    ..aOS(2, _omitFieldNames ? '' : 'url')
    ..aOS(3, _omitFieldNames ? '' : 'formattedURL', protoName: 'formattedURL')
    ..aOB(4, _omitFieldNames ? '' : 'hasCard', protoName: 'hasCard')
    ..aOB(5, _omitFieldNames ? '' : 'hasBank', protoName: 'hasBank')
    ..aOB(6, _omitFieldNames ? '' : 'hasIdentities', protoName: 'hasIdentities')
    ..aOB(7, _omitFieldNames ? '' : 'hasTransacted', protoName: 'hasTransacted')
    ..aOB(8, _omitFieldNames ? '' : 'hasWalletAddress', protoName: 'hasWalletAddress')
    ..aOB(9, _omitFieldNames ? '' : 'hasBalances', protoName: 'hasBalances')
    ..aOB(10, _omitFieldNames ? '' : 'exceededLimits', protoName: 'exceededLimits')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WalletInfo clone() => WalletInfo()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WalletInfo copyWith(void Function(WalletInfo) updates) => super.copyWith((message) => updates(message as WalletInfo)) as WalletInfo;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WalletInfo create() => WalletInfo._();
  WalletInfo createEmptyInstance() => create();
  static $pb.PbList<WalletInfo> createRepeated() => $pb.PbList<WalletInfo>();
  @$core.pragma('dart2js:noInline')
  static WalletInfo getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WalletInfo>(create);
  static WalletInfo? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletID => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletID() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get url => $_getSZ(1);
  @$pb.TagNumber(2)
  set url($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearUrl() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get formattedURL => $_getSZ(2);
  @$pb.TagNumber(3)
  set formattedURL($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasFormattedURL() => $_has(2);
  @$pb.TagNumber(3)
  void clearFormattedURL() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get hasCard => $_getBF(3);
  @$pb.TagNumber(4)
  set hasCard($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasHasCard() => $_has(3);
  @$pb.TagNumber(4)
  void clearHasCard() => clearField(4);

  @$pb.TagNumber(5)
  $core.bool get hasBank => $_getBF(4);
  @$pb.TagNumber(5)
  set hasBank($core.bool v) { $_setBool(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasHasBank() => $_has(4);
  @$pb.TagNumber(5)
  void clearHasBank() => clearField(5);

  @$pb.TagNumber(6)
  $core.bool get hasIdentities => $_getBF(5);
  @$pb.TagNumber(6)
  set hasIdentities($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasHasIdentities() => $_has(5);
  @$pb.TagNumber(6)
  void clearHasIdentities() => clearField(6);

  @$pb.TagNumber(7)
  $core.bool get hasTransacted => $_getBF(6);
  @$pb.TagNumber(7)
  set hasTransacted($core.bool v) { $_setBool(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasHasTransacted() => $_has(6);
  @$pb.TagNumber(7)
  void clearHasTransacted() => clearField(7);

  @$pb.TagNumber(8)
  $core.bool get hasWalletAddress => $_getBF(7);
  @$pb.TagNumber(8)
  set hasWalletAddress($core.bool v) { $_setBool(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasHasWalletAddress() => $_has(7);
  @$pb.TagNumber(8)
  void clearHasWalletAddress() => clearField(8);

  @$pb.TagNumber(9)
  $core.bool get hasBalances => $_getBF(8);
  @$pb.TagNumber(9)
  set hasBalances($core.bool v) { $_setBool(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasHasBalances() => $_has(8);
  @$pb.TagNumber(9)
  void clearHasBalances() => clearField(9);

  @$pb.TagNumber(10)
  $core.bool get exceededLimits => $_getBF(9);
  @$pb.TagNumber(10)
  set exceededLimits($core.bool v) { $_setBool(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasExceededLimits() => $_has(9);
  @$pb.TagNumber(10)
  void clearExceededLimits() => clearField(10);
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
    $core.bool? addCardsEnabled,
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
    if (addCardsEnabled != null) {
      $result.addCardsEnabled = addCardsEnabled;
    }
    return $result;
  }
  Features._() : super();
  factory Features.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Features.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Features', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'sendEnabled', protoName: 'sendEnabled')
    ..aOB(2, _omitFieldNames ? '' : 'receiveEnabled', protoName: 'receiveEnabled')
    ..aOB(3, _omitFieldNames ? '' : 'linkedAccountsEnabled', protoName: 'linkedAccountsEnabled')
    ..aOB(4, _omitFieldNames ? '' : 'cardsEnabled', protoName: 'cardsEnabled')
    ..aOB(5, _omitFieldNames ? '' : 'banksEnabled', protoName: 'banksEnabled')
    ..aOB(6, _omitFieldNames ? '' : 'identitiesEnabled', protoName: 'identitiesEnabled')
    ..aOB(7, _omitFieldNames ? '' : 'twitterEnabled', protoName: 'twitterEnabled')
    ..aOB(8, _omitFieldNames ? '' : 'addCardsEnabled', protoName: 'addCardsEnabled')
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
  $core.bool get addCardsEnabled => $_getBF(7);
  @$pb.TagNumber(8)
  set addCardsEnabled($core.bool v) { $_setBool(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasAddCardsEnabled() => $_has(7);
  @$pb.TagNumber(8)
  void clearAddCardsEnabled() => clearField(8);
}

class CreateCardRequest extends $pb.GeneratedMessage {
  factory CreateCardRequest({
    $core.String? tokenID,
  }) {
    final $result = create();
    if (tokenID != null) {
      $result.tokenID = tokenID;
    }
    return $result;
  }
  CreateCardRequest._() : super();
  factory CreateCardRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateCardRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateCardRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'tokenID', protoName: 'tokenID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateCardRequest clone() => CreateCardRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateCardRequest copyWith(void Function(CreateCardRequest) updates) => super.copyWith((message) => updates(message as CreateCardRequest)) as CreateCardRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateCardRequest create() => CreateCardRequest._();
  CreateCardRequest createEmptyInstance() => create();
  static $pb.PbList<CreateCardRequest> createRepeated() => $pb.PbList<CreateCardRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateCardRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateCardRequest>(create);
  static CreateCardRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get tokenID => $_getSZ(0);
  @$pb.TagNumber(1)
  set tokenID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTokenID() => $_has(0);
  @$pb.TagNumber(1)
  void clearTokenID() => clearField(1);
}

class InitQuote3DSRequest extends $pb.GeneratedMessage {
  factory InitQuote3DSRequest({
    $core.String? quoteID,
  }) {
    final $result = create();
    if (quoteID != null) {
      $result.quoteID = quoteID;
    }
    return $result;
  }
  InitQuote3DSRequest._() : super();
  factory InitQuote3DSRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory InitQuote3DSRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'InitQuote3DSRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'quoteID', protoName: 'quoteID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  InitQuote3DSRequest clone() => InitQuote3DSRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  InitQuote3DSRequest copyWith(void Function(InitQuote3DSRequest) updates) => super.copyWith((message) => updates(message as InitQuote3DSRequest)) as InitQuote3DSRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static InitQuote3DSRequest create() => InitQuote3DSRequest._();
  InitQuote3DSRequest createEmptyInstance() => create();
  static $pb.PbList<InitQuote3DSRequest> createRepeated() => $pb.PbList<InitQuote3DSRequest>();
  @$core.pragma('dart2js:noInline')
  static InitQuote3DSRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<InitQuote3DSRequest>(create);
  static InitQuote3DSRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get quoteID => $_getSZ(0);
  @$pb.TagNumber(1)
  set quoteID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasQuoteID() => $_has(0);
  @$pb.TagNumber(1)
  void clearQuoteID() => clearField(1);
}

class Init3DSRequest extends $pb.GeneratedMessage {
  factory Init3DSRequest({
    $core.String? paymentID,
  }) {
    final $result = create();
    if (paymentID != null) {
      $result.paymentID = paymentID;
    }
    return $result;
  }
  Init3DSRequest._() : super();
  factory Init3DSRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Init3DSRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Init3DSRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'paymentID', protoName: 'paymentID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Init3DSRequest clone() => Init3DSRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Init3DSRequest copyWith(void Function(Init3DSRequest) updates) => super.copyWith((message) => updates(message as Init3DSRequest)) as Init3DSRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Init3DSRequest create() => Init3DSRequest._();
  Init3DSRequest createEmptyInstance() => create();
  static $pb.PbList<Init3DSRequest> createRepeated() => $pb.PbList<Init3DSRequest>();
  @$core.pragma('dart2js:noInline')
  static Init3DSRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Init3DSRequest>(create);
  static Init3DSRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get paymentID => $_getSZ(0);
  @$pb.TagNumber(1)
  set paymentID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPaymentID() => $_has(0);
  @$pb.TagNumber(1)
  void clearPaymentID() => clearField(1);
}

class Init3DSResponse extends $pb.GeneratedMessage {
  factory Init3DSResponse({
    $core.String? id,
    $core.String? jwt,
    $core.String? deviceCollectionURL,
    $core.String? songbirdURL,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (jwt != null) {
      $result.jwt = jwt;
    }
    if (deviceCollectionURL != null) {
      $result.deviceCollectionURL = deviceCollectionURL;
    }
    if (songbirdURL != null) {
      $result.songbirdURL = songbirdURL;
    }
    return $result;
  }
  Init3DSResponse._() : super();
  factory Init3DSResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Init3DSResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Init3DSResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'jwt')
    ..aOS(3, _omitFieldNames ? '' : 'deviceCollectionURL', protoName: 'deviceCollectionURL')
    ..aOS(4, _omitFieldNames ? '' : 'songbirdURL', protoName: 'songbirdURL')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Init3DSResponse clone() => Init3DSResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Init3DSResponse copyWith(void Function(Init3DSResponse) updates) => super.copyWith((message) => updates(message as Init3DSResponse)) as Init3DSResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Init3DSResponse create() => Init3DSResponse._();
  Init3DSResponse createEmptyInstance() => create();
  static $pb.PbList<Init3DSResponse> createRepeated() => $pb.PbList<Init3DSResponse>();
  @$core.pragma('dart2js:noInline')
  static Init3DSResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Init3DSResponse>(create);
  static Init3DSResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get jwt => $_getSZ(1);
  @$pb.TagNumber(2)
  set jwt($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasJwt() => $_has(1);
  @$pb.TagNumber(2)
  void clearJwt() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get deviceCollectionURL => $_getSZ(2);
  @$pb.TagNumber(3)
  set deviceCollectionURL($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasDeviceCollectionURL() => $_has(2);
  @$pb.TagNumber(3)
  void clearDeviceCollectionURL() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get songbirdURL => $_getSZ(3);
  @$pb.TagNumber(4)
  set songbirdURL($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasSongbirdURL() => $_has(3);
  @$pb.TagNumber(4)
  void clearSongbirdURL() => clearField(4);
}

class Lookup3DSResponse extends $pb.GeneratedMessage {
  factory Lookup3DSResponse({
    $core.String? processorTransactionID,
    $core.String? challengeURL,
    $core.String? payload,
  }) {
    final $result = create();
    if (processorTransactionID != null) {
      $result.processorTransactionID = processorTransactionID;
    }
    if (challengeURL != null) {
      $result.challengeURL = challengeURL;
    }
    if (payload != null) {
      $result.payload = payload;
    }
    return $result;
  }
  Lookup3DSResponse._() : super();
  factory Lookup3DSResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Lookup3DSResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Lookup3DSResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'processorTransactionID', protoName: 'processorTransactionID')
    ..aOS(2, _omitFieldNames ? '' : 'challengeURL', protoName: 'challengeURL')
    ..aOS(3, _omitFieldNames ? '' : 'payload')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Lookup3DSResponse clone() => Lookup3DSResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Lookup3DSResponse copyWith(void Function(Lookup3DSResponse) updates) => super.copyWith((message) => updates(message as Lookup3DSResponse)) as Lookup3DSResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Lookup3DSResponse create() => Lookup3DSResponse._();
  Lookup3DSResponse createEmptyInstance() => create();
  static $pb.PbList<Lookup3DSResponse> createRepeated() => $pb.PbList<Lookup3DSResponse>();
  @$core.pragma('dart2js:noInline')
  static Lookup3DSResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Lookup3DSResponse>(create);
  static Lookup3DSResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get processorTransactionID => $_getSZ(0);
  @$pb.TagNumber(1)
  set processorTransactionID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProcessorTransactionID() => $_has(0);
  @$pb.TagNumber(1)
  void clearProcessorTransactionID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get challengeURL => $_getSZ(1);
  @$pb.TagNumber(2)
  set challengeURL($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChallengeURL() => $_has(1);
  @$pb.TagNumber(2)
  void clearChallengeURL() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get payload => $_getSZ(2);
  @$pb.TagNumber(3)
  set payload($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPayload() => $_has(2);
  @$pb.TagNumber(3)
  void clearPayload() => clearField(3);
}

class Lookup3DSRequest extends $pb.GeneratedMessage {
  factory Lookup3DSRequest({
    $core.String? threeDSID,
    $core.bool? javascriptEnabled,
    $core.String? userAgent,
    $core.String? header,
    $core.bool? javaEnabled,
    $core.String? language,
    $core.String? colorDepth,
    $core.String? screenHeight,
    $core.String? screenWidth,
    $core.String? timezone,
    $core.String? ipAddress,
  }) {
    final $result = create();
    if (threeDSID != null) {
      $result.threeDSID = threeDSID;
    }
    if (javascriptEnabled != null) {
      $result.javascriptEnabled = javascriptEnabled;
    }
    if (userAgent != null) {
      $result.userAgent = userAgent;
    }
    if (header != null) {
      $result.header = header;
    }
    if (javaEnabled != null) {
      $result.javaEnabled = javaEnabled;
    }
    if (language != null) {
      $result.language = language;
    }
    if (colorDepth != null) {
      $result.colorDepth = colorDepth;
    }
    if (screenHeight != null) {
      $result.screenHeight = screenHeight;
    }
    if (screenWidth != null) {
      $result.screenWidth = screenWidth;
    }
    if (timezone != null) {
      $result.timezone = timezone;
    }
    if (ipAddress != null) {
      $result.ipAddress = ipAddress;
    }
    return $result;
  }
  Lookup3DSRequest._() : super();
  factory Lookup3DSRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Lookup3DSRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Lookup3DSRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'threeDSID', protoName: 'threeDSID')
    ..aOB(2, _omitFieldNames ? '' : 'javascriptEnabled', protoName: 'javascriptEnabled')
    ..aOS(3, _omitFieldNames ? '' : 'userAgent', protoName: 'userAgent')
    ..aOS(4, _omitFieldNames ? '' : 'header')
    ..aOB(5, _omitFieldNames ? '' : 'javaEnabled', protoName: 'javaEnabled')
    ..aOS(6, _omitFieldNames ? '' : 'language')
    ..aOS(7, _omitFieldNames ? '' : 'colorDepth', protoName: 'colorDepth')
    ..aOS(8, _omitFieldNames ? '' : 'screenHeight', protoName: 'screenHeight')
    ..aOS(9, _omitFieldNames ? '' : 'screenWidth', protoName: 'screenWidth')
    ..aOS(10, _omitFieldNames ? '' : 'timezone')
    ..aOS(12, _omitFieldNames ? '' : 'ipAddress', protoName: 'ipAddress')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Lookup3DSRequest clone() => Lookup3DSRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Lookup3DSRequest copyWith(void Function(Lookup3DSRequest) updates) => super.copyWith((message) => updates(message as Lookup3DSRequest)) as Lookup3DSRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Lookup3DSRequest create() => Lookup3DSRequest._();
  Lookup3DSRequest createEmptyInstance() => create();
  static $pb.PbList<Lookup3DSRequest> createRepeated() => $pb.PbList<Lookup3DSRequest>();
  @$core.pragma('dart2js:noInline')
  static Lookup3DSRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Lookup3DSRequest>(create);
  static Lookup3DSRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get threeDSID => $_getSZ(0);
  @$pb.TagNumber(1)
  set threeDSID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasThreeDSID() => $_has(0);
  @$pb.TagNumber(1)
  void clearThreeDSID() => clearField(1);

  @$pb.TagNumber(2)
  $core.bool get javascriptEnabled => $_getBF(1);
  @$pb.TagNumber(2)
  set javascriptEnabled($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasJavascriptEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearJavascriptEnabled() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get userAgent => $_getSZ(2);
  @$pb.TagNumber(3)
  set userAgent($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasUserAgent() => $_has(2);
  @$pb.TagNumber(3)
  void clearUserAgent() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get header => $_getSZ(3);
  @$pb.TagNumber(4)
  set header($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasHeader() => $_has(3);
  @$pb.TagNumber(4)
  void clearHeader() => clearField(4);

  @$pb.TagNumber(5)
  $core.bool get javaEnabled => $_getBF(4);
  @$pb.TagNumber(5)
  set javaEnabled($core.bool v) { $_setBool(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasJavaEnabled() => $_has(4);
  @$pb.TagNumber(5)
  void clearJavaEnabled() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get language => $_getSZ(5);
  @$pb.TagNumber(6)
  set language($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasLanguage() => $_has(5);
  @$pb.TagNumber(6)
  void clearLanguage() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get colorDepth => $_getSZ(6);
  @$pb.TagNumber(7)
  set colorDepth($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasColorDepth() => $_has(6);
  @$pb.TagNumber(7)
  void clearColorDepth() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get screenHeight => $_getSZ(7);
  @$pb.TagNumber(8)
  set screenHeight($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasScreenHeight() => $_has(7);
  @$pb.TagNumber(8)
  void clearScreenHeight() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get screenWidth => $_getSZ(8);
  @$pb.TagNumber(9)
  set screenWidth($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasScreenWidth() => $_has(8);
  @$pb.TagNumber(9)
  void clearScreenWidth() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get timezone => $_getSZ(9);
  @$pb.TagNumber(10)
  set timezone($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasTimezone() => $_has(9);
  @$pb.TagNumber(10)
  void clearTimezone() => clearField(10);

  @$pb.TagNumber(12)
  $core.String get ipAddress => $_getSZ(10);
  @$pb.TagNumber(12)
  set ipAddress($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(12)
  $core.bool hasIpAddress() => $_has(10);
  @$pb.TagNumber(12)
  void clearIpAddress() => clearField(12);
}

class Authenticate3DSRequest extends $pb.GeneratedMessage {
  factory Authenticate3DSRequest({
    $core.String? threeDSID,
    $core.String? jwt,
  }) {
    final $result = create();
    if (threeDSID != null) {
      $result.threeDSID = threeDSID;
    }
    if (jwt != null) {
      $result.jwt = jwt;
    }
    return $result;
  }
  Authenticate3DSRequest._() : super();
  factory Authenticate3DSRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Authenticate3DSRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Authenticate3DSRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'threeDSID', protoName: 'threeDSID')
    ..aOS(2, _omitFieldNames ? '' : 'jwt')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Authenticate3DSRequest clone() => Authenticate3DSRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Authenticate3DSRequest copyWith(void Function(Authenticate3DSRequest) updates) => super.copyWith((message) => updates(message as Authenticate3DSRequest)) as Authenticate3DSRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Authenticate3DSRequest create() => Authenticate3DSRequest._();
  Authenticate3DSRequest createEmptyInstance() => create();
  static $pb.PbList<Authenticate3DSRequest> createRepeated() => $pb.PbList<Authenticate3DSRequest>();
  @$core.pragma('dart2js:noInline')
  static Authenticate3DSRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Authenticate3DSRequest>(create);
  static Authenticate3DSRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get threeDSID => $_getSZ(0);
  @$pb.TagNumber(1)
  set threeDSID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasThreeDSID() => $_has(0);
  @$pb.TagNumber(1)
  void clearThreeDSID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get jwt => $_getSZ(1);
  @$pb.TagNumber(2)
  set jwt($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasJwt() => $_has(1);
  @$pb.TagNumber(2)
  void clearJwt() => clearField(2);
}

class Authenticate3DSResponse extends $pb.GeneratedMessage {
  factory Authenticate3DSResponse({
    $core.String? status,
  }) {
    final $result = create();
    if (status != null) {
      $result.status = status;
    }
    return $result;
  }
  Authenticate3DSResponse._() : super();
  factory Authenticate3DSResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Authenticate3DSResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Authenticate3DSResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Authenticate3DSResponse clone() => Authenticate3DSResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Authenticate3DSResponse copyWith(void Function(Authenticate3DSResponse) updates) => super.copyWith((message) => updates(message as Authenticate3DSResponse)) as Authenticate3DSResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Authenticate3DSResponse create() => Authenticate3DSResponse._();
  Authenticate3DSResponse createEmptyInstance() => create();
  static $pb.PbList<Authenticate3DSResponse> createRepeated() => $pb.PbList<Authenticate3DSResponse>();
  @$core.pragma('dart2js:noInline')
  static Authenticate3DSResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Authenticate3DSResponse>(create);
  static Authenticate3DSResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get status => $_getSZ(0);
  @$pb.TagNumber(1)
  set status($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => clearField(1);
}

class CreateMXBankAccountsRequest extends $pb.GeneratedMessage {
  factory CreateMXBankAccountsRequest({
    $core.String? sessionGuid,
    $core.String? userGuid,
    $core.String? memberGuid,
  }) {
    final $result = create();
    if (sessionGuid != null) {
      $result.sessionGuid = sessionGuid;
    }
    if (userGuid != null) {
      $result.userGuid = userGuid;
    }
    if (memberGuid != null) {
      $result.memberGuid = memberGuid;
    }
    return $result;
  }
  CreateMXBankAccountsRequest._() : super();
  factory CreateMXBankAccountsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateMXBankAccountsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateMXBankAccountsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'sessionGuid', protoName: 'sessionGuid')
    ..aOS(2, _omitFieldNames ? '' : 'userGuid', protoName: 'userGuid')
    ..aOS(3, _omitFieldNames ? '' : 'memberGuid', protoName: 'memberGuid')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateMXBankAccountsRequest clone() => CreateMXBankAccountsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateMXBankAccountsRequest copyWith(void Function(CreateMXBankAccountsRequest) updates) => super.copyWith((message) => updates(message as CreateMXBankAccountsRequest)) as CreateMXBankAccountsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMXBankAccountsRequest create() => CreateMXBankAccountsRequest._();
  CreateMXBankAccountsRequest createEmptyInstance() => create();
  static $pb.PbList<CreateMXBankAccountsRequest> createRepeated() => $pb.PbList<CreateMXBankAccountsRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateMXBankAccountsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateMXBankAccountsRequest>(create);
  static CreateMXBankAccountsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionGuid => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionGuid($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSessionGuid() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionGuid() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get userGuid => $_getSZ(1);
  @$pb.TagNumber(2)
  set userGuid($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasUserGuid() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserGuid() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get memberGuid => $_getSZ(2);
  @$pb.TagNumber(3)
  set memberGuid($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMemberGuid() => $_has(2);
  @$pb.TagNumber(3)
  void clearMemberGuid() => clearField(3);
}

class CreateMXBankAccountsResponse extends $pb.GeneratedMessage {
  factory CreateMXBankAccountsResponse({
    $core.Iterable<LinkedAccount>? linkedAccounts,
  }) {
    final $result = create();
    if (linkedAccounts != null) {
      $result.linkedAccounts.addAll(linkedAccounts);
    }
    return $result;
  }
  CreateMXBankAccountsResponse._() : super();
  factory CreateMXBankAccountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateMXBankAccountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateMXBankAccountsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<LinkedAccount>(1, _omitFieldNames ? '' : 'linkedAccounts', $pb.PbFieldType.PM, protoName: 'linkedAccounts', subBuilder: LinkedAccount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateMXBankAccountsResponse clone() => CreateMXBankAccountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateMXBankAccountsResponse copyWith(void Function(CreateMXBankAccountsResponse) updates) => super.copyWith((message) => updates(message as CreateMXBankAccountsResponse)) as CreateMXBankAccountsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMXBankAccountsResponse create() => CreateMXBankAccountsResponse._();
  CreateMXBankAccountsResponse createEmptyInstance() => create();
  static $pb.PbList<CreateMXBankAccountsResponse> createRepeated() => $pb.PbList<CreateMXBankAccountsResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateMXBankAccountsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateMXBankAccountsResponse>(create);
  static CreateMXBankAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<LinkedAccount> get linkedAccounts => $_getList(0);
}

class MXWidgetResponse extends $pb.GeneratedMessage {
  factory MXWidgetResponse({
    $core.String? url,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    return $result;
  }
  MXWidgetResponse._() : super();
  factory MXWidgetResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MXWidgetResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'MXWidgetResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MXWidgetResponse clone() => MXWidgetResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MXWidgetResponse copyWith(void Function(MXWidgetResponse) updates) => super.copyWith((message) => updates(message as MXWidgetResponse)) as MXWidgetResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MXWidgetResponse create() => MXWidgetResponse._();
  MXWidgetResponse createEmptyInstance() => create();
  static $pb.PbList<MXWidgetResponse> createRepeated() => $pb.PbList<MXWidgetResponse>();
  @$core.pragma('dart2js:noInline')
  static MXWidgetResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MXWidgetResponse>(create);
  static MXWidgetResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class ConnectionLimits extends $pb.GeneratedMessage {
  factory ConnectionLimits({
    Amount? daily,
    Amount? monthly,
    Amount? overall,
  }) {
    final $result = create();
    if (daily != null) {
      $result.daily = daily;
    }
    if (monthly != null) {
      $result.monthly = monthly;
    }
    if (overall != null) {
      $result.overall = overall;
    }
    return $result;
  }
  ConnectionLimits._() : super();
  factory ConnectionLimits.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConnectionLimits.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConnectionLimits', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<Amount>(1, _omitFieldNames ? '' : 'daily', subBuilder: Amount.create)
    ..aOM<Amount>(2, _omitFieldNames ? '' : 'monthly', subBuilder: Amount.create)
    ..aOM<Amount>(3, _omitFieldNames ? '' : 'overall', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConnectionLimits clone() => ConnectionLimits()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConnectionLimits copyWith(void Function(ConnectionLimits) updates) => super.copyWith((message) => updates(message as ConnectionLimits)) as ConnectionLimits;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConnectionLimits create() => ConnectionLimits._();
  ConnectionLimits createEmptyInstance() => create();
  static $pb.PbList<ConnectionLimits> createRepeated() => $pb.PbList<ConnectionLimits>();
  @$core.pragma('dart2js:noInline')
  static ConnectionLimits getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConnectionLimits>(create);
  static ConnectionLimits? _defaultInstance;

  @$pb.TagNumber(1)
  Amount get daily => $_getN(0);
  @$pb.TagNumber(1)
  set daily(Amount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasDaily() => $_has(0);
  @$pb.TagNumber(1)
  void clearDaily() => clearField(1);
  @$pb.TagNumber(1)
  Amount ensureDaily() => $_ensure(0);

  @$pb.TagNumber(2)
  Amount get monthly => $_getN(1);
  @$pb.TagNumber(2)
  set monthly(Amount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasMonthly() => $_has(1);
  @$pb.TagNumber(2)
  void clearMonthly() => clearField(2);
  @$pb.TagNumber(2)
  Amount ensureMonthly() => $_ensure(1);

  @$pb.TagNumber(3)
  Amount get overall => $_getN(2);
  @$pb.TagNumber(3)
  set overall(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasOverall() => $_has(2);
  @$pb.TagNumber(3)
  void clearOverall() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureOverall() => $_ensure(2);
}

class Connection extends $pb.GeneratedMessage {
  factory Connection({
    $core.String? id,
    $core.String? applicationName,
    $core.String? publicKeyFingerprint,
    $core.String? createdAt,
    $core.String? lastUsedAt,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (applicationName != null) {
      $result.applicationName = applicationName;
    }
    if (publicKeyFingerprint != null) {
      $result.publicKeyFingerprint = publicKeyFingerprint;
    }
    if (createdAt != null) {
      $result.createdAt = createdAt;
    }
    if (lastUsedAt != null) {
      $result.lastUsedAt = lastUsedAt;
    }
    return $result;
  }
  Connection._() : super();
  factory Connection.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Connection.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Connection', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'applicationName', protoName: 'applicationName')
    ..aOS(3, _omitFieldNames ? '' : 'publicKeyFingerprint', protoName: 'publicKeyFingerprint')
    ..aOS(4, _omitFieldNames ? '' : 'createdAt', protoName: 'createdAt')
    ..aOS(5, _omitFieldNames ? '' : 'lastUsedAt', protoName: 'lastUsedAt')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Connection clone() => Connection()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Connection copyWith(void Function(Connection) updates) => super.copyWith((message) => updates(message as Connection)) as Connection;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Connection create() => Connection._();
  Connection createEmptyInstance() => create();
  static $pb.PbList<Connection> createRepeated() => $pb.PbList<Connection>();
  @$core.pragma('dart2js:noInline')
  static Connection getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Connection>(create);
  static Connection? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get applicationName => $_getSZ(1);
  @$pb.TagNumber(2)
  set applicationName($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasApplicationName() => $_has(1);
  @$pb.TagNumber(2)
  void clearApplicationName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get publicKeyFingerprint => $_getSZ(2);
  @$pb.TagNumber(3)
  set publicKeyFingerprint($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPublicKeyFingerprint() => $_has(2);
  @$pb.TagNumber(3)
  void clearPublicKeyFingerprint() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get createdAt => $_getSZ(3);
  @$pb.TagNumber(4)
  set createdAt($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCreatedAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearCreatedAt() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get lastUsedAt => $_getSZ(4);
  @$pb.TagNumber(5)
  set lastUsedAt($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasLastUsedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearLastUsedAt() => clearField(5);
}

class CreateConnectionRequest extends $pb.GeneratedMessage {
  factory CreateConnectionRequest({
    $core.String? applicationName,
    $core.String? publicKey,
    Amount? dailyLimit,
    Amount? monthlyLimit,
    Amount? overallLimit,
  }) {
    final $result = create();
    if (applicationName != null) {
      $result.applicationName = applicationName;
    }
    if (publicKey != null) {
      $result.publicKey = publicKey;
    }
    if (dailyLimit != null) {
      $result.dailyLimit = dailyLimit;
    }
    if (monthlyLimit != null) {
      $result.monthlyLimit = monthlyLimit;
    }
    if (overallLimit != null) {
      $result.overallLimit = overallLimit;
    }
    return $result;
  }
  CreateConnectionRequest._() : super();
  factory CreateConnectionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateConnectionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateConnectionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'applicationName', protoName: 'applicationName')
    ..aOS(2, _omitFieldNames ? '' : 'publicKey', protoName: 'publicKey')
    ..aOM<Amount>(3, _omitFieldNames ? '' : 'dailyLimit', protoName: 'dailyLimit', subBuilder: Amount.create)
    ..aOM<Amount>(4, _omitFieldNames ? '' : 'monthlyLimit', protoName: 'monthlyLimit', subBuilder: Amount.create)
    ..aOM<Amount>(5, _omitFieldNames ? '' : 'overallLimit', protoName: 'overallLimit', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateConnectionRequest clone() => CreateConnectionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateConnectionRequest copyWith(void Function(CreateConnectionRequest) updates) => super.copyWith((message) => updates(message as CreateConnectionRequest)) as CreateConnectionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateConnectionRequest create() => CreateConnectionRequest._();
  CreateConnectionRequest createEmptyInstance() => create();
  static $pb.PbList<CreateConnectionRequest> createRepeated() => $pb.PbList<CreateConnectionRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateConnectionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateConnectionRequest>(create);
  static CreateConnectionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get applicationName => $_getSZ(0);
  @$pb.TagNumber(1)
  set applicationName($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasApplicationName() => $_has(0);
  @$pb.TagNumber(1)
  void clearApplicationName() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get publicKey => $_getSZ(1);
  @$pb.TagNumber(2)
  set publicKey($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPublicKey() => $_has(1);
  @$pb.TagNumber(2)
  void clearPublicKey() => clearField(2);

  @$pb.TagNumber(3)
  Amount get dailyLimit => $_getN(2);
  @$pb.TagNumber(3)
  set dailyLimit(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasDailyLimit() => $_has(2);
  @$pb.TagNumber(3)
  void clearDailyLimit() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureDailyLimit() => $_ensure(2);

  @$pb.TagNumber(4)
  Amount get monthlyLimit => $_getN(3);
  @$pb.TagNumber(4)
  set monthlyLimit(Amount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasMonthlyLimit() => $_has(3);
  @$pb.TagNumber(4)
  void clearMonthlyLimit() => clearField(4);
  @$pb.TagNumber(4)
  Amount ensureMonthlyLimit() => $_ensure(3);

  @$pb.TagNumber(5)
  Amount get overallLimit => $_getN(4);
  @$pb.TagNumber(5)
  set overallLimit(Amount v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasOverallLimit() => $_has(4);
  @$pb.TagNumber(5)
  void clearOverallLimit() => clearField(5);
  @$pb.TagNumber(5)
  Amount ensureOverallLimit() => $_ensure(4);
}

class GetConnectionRequest extends $pb.GeneratedMessage {
  factory GetConnectionRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetConnectionRequest._() : super();
  factory GetConnectionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetConnectionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetConnectionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetConnectionRequest clone() => GetConnectionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetConnectionRequest copyWith(void Function(GetConnectionRequest) updates) => super.copyWith((message) => updates(message as GetConnectionRequest)) as GetConnectionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetConnectionRequest create() => GetConnectionRequest._();
  GetConnectionRequest createEmptyInstance() => create();
  static $pb.PbList<GetConnectionRequest> createRepeated() => $pb.PbList<GetConnectionRequest>();
  @$core.pragma('dart2js:noInline')
  static GetConnectionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetConnectionRequest>(create);
  static GetConnectionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class GetConnectionLimitsRequest extends $pb.GeneratedMessage {
  factory GetConnectionLimitsRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetConnectionLimitsRequest._() : super();
  factory GetConnectionLimitsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetConnectionLimitsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetConnectionLimitsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetConnectionLimitsRequest clone() => GetConnectionLimitsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetConnectionLimitsRequest copyWith(void Function(GetConnectionLimitsRequest) updates) => super.copyWith((message) => updates(message as GetConnectionLimitsRequest)) as GetConnectionLimitsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetConnectionLimitsRequest create() => GetConnectionLimitsRequest._();
  GetConnectionLimitsRequest createEmptyInstance() => create();
  static $pb.PbList<GetConnectionLimitsRequest> createRepeated() => $pb.PbList<GetConnectionLimitsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetConnectionLimitsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetConnectionLimitsRequest>(create);
  static GetConnectionLimitsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class DeleteConnectionRequest extends $pb.GeneratedMessage {
  factory DeleteConnectionRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  DeleteConnectionRequest._() : super();
  factory DeleteConnectionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DeleteConnectionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'DeleteConnectionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DeleteConnectionRequest clone() => DeleteConnectionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DeleteConnectionRequest copyWith(void Function(DeleteConnectionRequest) updates) => super.copyWith((message) => updates(message as DeleteConnectionRequest)) as DeleteConnectionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteConnectionRequest create() => DeleteConnectionRequest._();
  DeleteConnectionRequest createEmptyInstance() => create();
  static $pb.PbList<DeleteConnectionRequest> createRepeated() => $pb.PbList<DeleteConnectionRequest>();
  @$core.pragma('dart2js:noInline')
  static DeleteConnectionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DeleteConnectionRequest>(create);
  static DeleteConnectionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class ListConnectionsResponse extends $pb.GeneratedMessage {
  factory ListConnectionsResponse({
    $core.Iterable<Connection>? keys,
  }) {
    final $result = create();
    if (keys != null) {
      $result.keys.addAll(keys);
    }
    return $result;
  }
  ListConnectionsResponse._() : super();
  factory ListConnectionsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListConnectionsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListConnectionsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Connection>(1, _omitFieldNames ? '' : 'keys', $pb.PbFieldType.PM, subBuilder: Connection.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListConnectionsResponse clone() => ListConnectionsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListConnectionsResponse copyWith(void Function(ListConnectionsResponse) updates) => super.copyWith((message) => updates(message as ListConnectionsResponse)) as ListConnectionsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListConnectionsResponse create() => ListConnectionsResponse._();
  ListConnectionsResponse createEmptyInstance() => create();
  static $pb.PbList<ListConnectionsResponse> createRepeated() => $pb.PbList<ListConnectionsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListConnectionsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListConnectionsResponse>(create);
  static ListConnectionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Connection> get keys => $_getList(0);
}

class UpdateConnectionLimitsRequest extends $pb.GeneratedMessage {
  factory UpdateConnectionLimitsRequest({
    $core.String? id,
    Amount? daily,
    Amount? monthly,
    Amount? overall,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (daily != null) {
      $result.daily = daily;
    }
    if (monthly != null) {
      $result.monthly = monthly;
    }
    if (overall != null) {
      $result.overall = overall;
    }
    return $result;
  }
  UpdateConnectionLimitsRequest._() : super();
  factory UpdateConnectionLimitsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateConnectionLimitsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateConnectionLimitsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<Amount>(2, _omitFieldNames ? '' : 'daily', subBuilder: Amount.create)
    ..aOM<Amount>(3, _omitFieldNames ? '' : 'monthly', subBuilder: Amount.create)
    ..aOM<Amount>(4, _omitFieldNames ? '' : 'overall', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateConnectionLimitsRequest clone() => UpdateConnectionLimitsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateConnectionLimitsRequest copyWith(void Function(UpdateConnectionLimitsRequest) updates) => super.copyWith((message) => updates(message as UpdateConnectionLimitsRequest)) as UpdateConnectionLimitsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateConnectionLimitsRequest create() => UpdateConnectionLimitsRequest._();
  UpdateConnectionLimitsRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateConnectionLimitsRequest> createRepeated() => $pb.PbList<UpdateConnectionLimitsRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateConnectionLimitsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateConnectionLimitsRequest>(create);
  static UpdateConnectionLimitsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  Amount get daily => $_getN(1);
  @$pb.TagNumber(2)
  set daily(Amount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasDaily() => $_has(1);
  @$pb.TagNumber(2)
  void clearDaily() => clearField(2);
  @$pb.TagNumber(2)
  Amount ensureDaily() => $_ensure(1);

  @$pb.TagNumber(3)
  Amount get monthly => $_getN(2);
  @$pb.TagNumber(3)
  set monthly(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasMonthly() => $_has(2);
  @$pb.TagNumber(3)
  void clearMonthly() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureMonthly() => $_ensure(2);

  @$pb.TagNumber(4)
  Amount get overall => $_getN(3);
  @$pb.TagNumber(4)
  set overall(Amount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasOverall() => $_has(3);
  @$pb.TagNumber(4)
  void clearOverall() => clearField(4);
  @$pb.TagNumber(4)
  Amount ensureOverall() => $_ensure(3);
}

class Transfer extends $pb.GeneratedMessage {
  factory Transfer({
    $core.String? type,
    $core.String? state,
    $6.Timestamp? timestamp,
    Amount? amount,
    $core.String? foreignId,
    $core.String? linkedAccountId,
  }) {
    final $result = create();
    if (type != null) {
      $result.type = type;
    }
    if (state != null) {
      $result.state = state;
    }
    if (timestamp != null) {
      $result.timestamp = timestamp;
    }
    if (amount != null) {
      $result.amount = amount;
    }
    if (foreignId != null) {
      $result.foreignId = foreignId;
    }
    if (linkedAccountId != null) {
      $result.linkedAccountId = linkedAccountId;
    }
    return $result;
  }
  Transfer._() : super();
  factory Transfer.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transfer.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Transfer', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'type')
    ..aOS(2, _omitFieldNames ? '' : 'state')
    ..aOM<$6.Timestamp>(3, _omitFieldNames ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..aOM<Amount>(4, _omitFieldNames ? '' : 'amount', subBuilder: Amount.create)
    ..aOS(5, _omitFieldNames ? '' : 'foreignId', protoName: 'foreignId')
    ..aOS(6, _omitFieldNames ? '' : 'linkedAccountId', protoName: 'linkedAccountId')
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
  $core.String get type => $_getSZ(0);
  @$pb.TagNumber(1)
  set type($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get state => $_getSZ(1);
  @$pb.TagNumber(2)
  set state($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasState() => $_has(1);
  @$pb.TagNumber(2)
  void clearState() => clearField(2);

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

  @$pb.TagNumber(4)
  Amount get amount => $_getN(3);
  @$pb.TagNumber(4)
  set amount(Amount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmount() => clearField(4);
  @$pb.TagNumber(4)
  Amount ensureAmount() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get foreignId => $_getSZ(4);
  @$pb.TagNumber(5)
  set foreignId($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasForeignId() => $_has(4);
  @$pb.TagNumber(5)
  void clearForeignId() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get linkedAccountId => $_getSZ(5);
  @$pb.TagNumber(6)
  set linkedAccountId($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasLinkedAccountId() => $_has(5);
  @$pb.TagNumber(6)
  void clearLinkedAccountId() => clearField(6);
}

class ListStatementsResponse extends $pb.GeneratedMessage {
  factory ListStatementsResponse({
    $core.Iterable<$core.String>? periods,
    $core.String? nextPageToken,
  }) {
    final $result = create();
    if (periods != null) {
      $result.periods.addAll(periods);
    }
    if (nextPageToken != null) {
      $result.nextPageToken = nextPageToken;
    }
    return $result;
  }
  ListStatementsResponse._() : super();
  factory ListStatementsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListStatementsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListStatementsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'periods')
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken', protoName: 'nextPageToken')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListStatementsResponse clone() => ListStatementsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListStatementsResponse copyWith(void Function(ListStatementsResponse) updates) => super.copyWith((message) => updates(message as ListStatementsResponse)) as ListStatementsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListStatementsResponse create() => ListStatementsResponse._();
  ListStatementsResponse createEmptyInstance() => create();
  static $pb.PbList<ListStatementsResponse> createRepeated() => $pb.PbList<ListStatementsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListStatementsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListStatementsResponse>(create);
  static ListStatementsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get periods => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => clearField(2);
}

class CreateSupportTicketRequest extends $pb.GeneratedMessage {
  factory CreateSupportTicketRequest({
    $core.String? description,
    $core.String? firstName,
    $core.String? lastName,
    $core.String? email,
  }) {
    final $result = create();
    if (description != null) {
      $result.description = description;
    }
    if (firstName != null) {
      $result.firstName = firstName;
    }
    if (lastName != null) {
      $result.lastName = lastName;
    }
    if (email != null) {
      $result.email = email;
    }
    return $result;
  }
  CreateSupportTicketRequest._() : super();
  factory CreateSupportTicketRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateSupportTicketRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateSupportTicketRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'description')
    ..aOS(2, _omitFieldNames ? '' : 'firstName', protoName: 'firstName')
    ..aOS(3, _omitFieldNames ? '' : 'lastName', protoName: 'lastName')
    ..aOS(4, _omitFieldNames ? '' : 'email')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateSupportTicketRequest clone() => CreateSupportTicketRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateSupportTicketRequest copyWith(void Function(CreateSupportTicketRequest) updates) => super.copyWith((message) => updates(message as CreateSupportTicketRequest)) as CreateSupportTicketRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateSupportTicketRequest create() => CreateSupportTicketRequest._();
  CreateSupportTicketRequest createEmptyInstance() => create();
  static $pb.PbList<CreateSupportTicketRequest> createRepeated() => $pb.PbList<CreateSupportTicketRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateSupportTicketRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateSupportTicketRequest>(create);
  static CreateSupportTicketRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get description => $_getSZ(0);
  @$pb.TagNumber(1)
  set description($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDescription() => $_has(0);
  @$pb.TagNumber(1)
  void clearDescription() => clearField(1);

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
  $core.String get email => $_getSZ(3);
  @$pb.TagNumber(4)
  set email($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasEmail() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmail() => clearField(4);
}

class IndividualKYCResponse extends $pb.GeneratedMessage {
  factory IndividualKYCResponse({
    $core.String? firstName,
    $core.String? lastName,
    $core.String? countryCode,
    $core.int? gender,
    $6.Timestamp? dateOfBirth,
    Address? address,
    $core.String? placeOfBirth,
    $core.String? nationality,
  }) {
    final $result = create();
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
    if (placeOfBirth != null) {
      $result.placeOfBirth = placeOfBirth;
    }
    if (nationality != null) {
      $result.nationality = nationality;
    }
    return $result;
  }
  IndividualKYCResponse._() : super();
  factory IndividualKYCResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IndividualKYCResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'IndividualKYCResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'firstName', protoName: 'firstName')
    ..aOS(2, _omitFieldNames ? '' : 'lastName', protoName: 'lastName')
    ..aOS(3, _omitFieldNames ? '' : 'countryCode', protoName: 'countryCode')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'gender', $pb.PbFieldType.O3)
    ..aOM<$6.Timestamp>(5, _omitFieldNames ? '' : 'dateOfBirth', protoName: 'dateOfBirth', subBuilder: $6.Timestamp.create)
    ..aOM<Address>(6, _omitFieldNames ? '' : 'address', subBuilder: Address.create)
    ..aOS(7, _omitFieldNames ? '' : 'placeOfBirth', protoName: 'placeOfBirth')
    ..aOS(8, _omitFieldNames ? '' : 'nationality')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IndividualKYCResponse clone() => IndividualKYCResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IndividualKYCResponse copyWith(void Function(IndividualKYCResponse) updates) => super.copyWith((message) => updates(message as IndividualKYCResponse)) as IndividualKYCResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static IndividualKYCResponse create() => IndividualKYCResponse._();
  IndividualKYCResponse createEmptyInstance() => create();
  static $pb.PbList<IndividualKYCResponse> createRepeated() => $pb.PbList<IndividualKYCResponse>();
  @$core.pragma('dart2js:noInline')
  static IndividualKYCResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IndividualKYCResponse>(create);
  static IndividualKYCResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get firstName => $_getSZ(0);
  @$pb.TagNumber(1)
  set firstName($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFirstName() => $_has(0);
  @$pb.TagNumber(1)
  void clearFirstName() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get lastName => $_getSZ(1);
  @$pb.TagNumber(2)
  set lastName($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLastName() => $_has(1);
  @$pb.TagNumber(2)
  void clearLastName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get countryCode => $_getSZ(2);
  @$pb.TagNumber(3)
  set countryCode($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCountryCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearCountryCode() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get gender => $_getIZ(3);
  @$pb.TagNumber(4)
  set gender($core.int v) { $_setSignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasGender() => $_has(3);
  @$pb.TagNumber(4)
  void clearGender() => clearField(4);

  @$pb.TagNumber(5)
  $6.Timestamp get dateOfBirth => $_getN(4);
  @$pb.TagNumber(5)
  set dateOfBirth($6.Timestamp v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasDateOfBirth() => $_has(4);
  @$pb.TagNumber(5)
  void clearDateOfBirth() => clearField(5);
  @$pb.TagNumber(5)
  $6.Timestamp ensureDateOfBirth() => $_ensure(4);

  @$pb.TagNumber(6)
  Address get address => $_getN(5);
  @$pb.TagNumber(6)
  set address(Address v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasAddress() => $_has(5);
  @$pb.TagNumber(6)
  void clearAddress() => clearField(6);
  @$pb.TagNumber(6)
  Address ensureAddress() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.String get placeOfBirth => $_getSZ(6);
  @$pb.TagNumber(7)
  set placeOfBirth($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasPlaceOfBirth() => $_has(6);
  @$pb.TagNumber(7)
  void clearPlaceOfBirth() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get nationality => $_getSZ(7);
  @$pb.TagNumber(8)
  set nationality($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasNationality() => $_has(7);
  @$pb.TagNumber(8)
  void clearNationality() => clearField(8);
}

class UpdateIndividualKYCRequest extends $pb.GeneratedMessage {
  factory UpdateIndividualKYCRequest({
    $core.String? firstName,
    $core.String? lastName,
    $core.String? countryCode,
    $core.int? gender,
    $6.Timestamp? dateOfBirth,
    Address? address,
    $core.String? ipAddress,
    $core.String? placeOfBirth,
    $core.String? nationality,
  }) {
    final $result = create();
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
    if (ipAddress != null) {
      $result.ipAddress = ipAddress;
    }
    if (placeOfBirth != null) {
      $result.placeOfBirth = placeOfBirth;
    }
    if (nationality != null) {
      $result.nationality = nationality;
    }
    return $result;
  }
  UpdateIndividualKYCRequest._() : super();
  factory UpdateIndividualKYCRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateIndividualKYCRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateIndividualKYCRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'firstName', protoName: 'firstName')
    ..aOS(2, _omitFieldNames ? '' : 'lastName', protoName: 'lastName')
    ..aOS(3, _omitFieldNames ? '' : 'countryCode', protoName: 'countryCode')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'gender', $pb.PbFieldType.O3)
    ..aOM<$6.Timestamp>(5, _omitFieldNames ? '' : 'dateOfBirth', protoName: 'dateOfBirth', subBuilder: $6.Timestamp.create)
    ..aOM<Address>(6, _omitFieldNames ? '' : 'address', subBuilder: Address.create)
    ..aOS(7, _omitFieldNames ? '' : 'ipAddress', protoName: 'ipAddress')
    ..aOS(8, _omitFieldNames ? '' : 'placeOfBirth', protoName: 'placeOfBirth')
    ..aOS(9, _omitFieldNames ? '' : 'nationality')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateIndividualKYCRequest clone() => UpdateIndividualKYCRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateIndividualKYCRequest copyWith(void Function(UpdateIndividualKYCRequest) updates) => super.copyWith((message) => updates(message as UpdateIndividualKYCRequest)) as UpdateIndividualKYCRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateIndividualKYCRequest create() => UpdateIndividualKYCRequest._();
  UpdateIndividualKYCRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateIndividualKYCRequest> createRepeated() => $pb.PbList<UpdateIndividualKYCRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateIndividualKYCRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateIndividualKYCRequest>(create);
  static UpdateIndividualKYCRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get firstName => $_getSZ(0);
  @$pb.TagNumber(1)
  set firstName($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFirstName() => $_has(0);
  @$pb.TagNumber(1)
  void clearFirstName() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get lastName => $_getSZ(1);
  @$pb.TagNumber(2)
  set lastName($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLastName() => $_has(1);
  @$pb.TagNumber(2)
  void clearLastName() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get countryCode => $_getSZ(2);
  @$pb.TagNumber(3)
  set countryCode($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCountryCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearCountryCode() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get gender => $_getIZ(3);
  @$pb.TagNumber(4)
  set gender($core.int v) { $_setSignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasGender() => $_has(3);
  @$pb.TagNumber(4)
  void clearGender() => clearField(4);

  @$pb.TagNumber(5)
  $6.Timestamp get dateOfBirth => $_getN(4);
  @$pb.TagNumber(5)
  set dateOfBirth($6.Timestamp v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasDateOfBirth() => $_has(4);
  @$pb.TagNumber(5)
  void clearDateOfBirth() => clearField(5);
  @$pb.TagNumber(5)
  $6.Timestamp ensureDateOfBirth() => $_ensure(4);

  @$pb.TagNumber(6)
  Address get address => $_getN(5);
  @$pb.TagNumber(6)
  set address(Address v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasAddress() => $_has(5);
  @$pb.TagNumber(6)
  void clearAddress() => clearField(6);
  @$pb.TagNumber(6)
  Address ensureAddress() => $_ensure(5);

  @$pb.TagNumber(7)
  $core.String get ipAddress => $_getSZ(6);
  @$pb.TagNumber(7)
  set ipAddress($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasIpAddress() => $_has(6);
  @$pb.TagNumber(7)
  void clearIpAddress() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get placeOfBirth => $_getSZ(7);
  @$pb.TagNumber(8)
  set placeOfBirth($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasPlaceOfBirth() => $_has(7);
  @$pb.TagNumber(8)
  void clearPlaceOfBirth() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get nationality => $_getSZ(8);
  @$pb.TagNumber(9)
  set nationality($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasNationality() => $_has(8);
  @$pb.TagNumber(9)
  void clearNationality() => clearField(9);
}

class Address extends $pb.GeneratedMessage {
  factory Address({
    $core.String? line1,
    $core.String? line2,
    $core.String? building,
    $core.String? apartment,
    $core.String? city,
    $core.String? state,
    $core.String? zipCode,
    $core.String? countryCode,
    $core.String? placeID,
    $core.String? formattedAddress,
  }) {
    final $result = create();
    if (line1 != null) {
      $result.line1 = line1;
    }
    if (line2 != null) {
      $result.line2 = line2;
    }
    if (building != null) {
      $result.building = building;
    }
    if (apartment != null) {
      $result.apartment = apartment;
    }
    if (city != null) {
      $result.city = city;
    }
    if (state != null) {
      $result.state = state;
    }
    if (zipCode != null) {
      $result.zipCode = zipCode;
    }
    if (countryCode != null) {
      $result.countryCode = countryCode;
    }
    if (placeID != null) {
      $result.placeID = placeID;
    }
    if (formattedAddress != null) {
      $result.formattedAddress = formattedAddress;
    }
    return $result;
  }
  Address._() : super();
  factory Address.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Address.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Address', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'line1')
    ..aOS(2, _omitFieldNames ? '' : 'line2')
    ..aOS(3, _omitFieldNames ? '' : 'building')
    ..aOS(4, _omitFieldNames ? '' : 'apartment')
    ..aOS(5, _omitFieldNames ? '' : 'city')
    ..aOS(6, _omitFieldNames ? '' : 'state')
    ..aOS(7, _omitFieldNames ? '' : 'zipCode', protoName: 'zipCode')
    ..aOS(8, _omitFieldNames ? '' : 'countryCode', protoName: 'countryCode')
    ..aOS(9, _omitFieldNames ? '' : 'placeID', protoName: 'placeID')
    ..aOS(10, _omitFieldNames ? '' : 'formattedAddress', protoName: 'formattedAddress')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Address clone() => Address()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Address copyWith(void Function(Address) updates) => super.copyWith((message) => updates(message as Address)) as Address;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Address create() => Address._();
  Address createEmptyInstance() => create();
  static $pb.PbList<Address> createRepeated() => $pb.PbList<Address>();
  @$core.pragma('dart2js:noInline')
  static Address getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Address>(create);
  static Address? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get line1 => $_getSZ(0);
  @$pb.TagNumber(1)
  set line1($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasLine1() => $_has(0);
  @$pb.TagNumber(1)
  void clearLine1() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get line2 => $_getSZ(1);
  @$pb.TagNumber(2)
  set line2($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLine2() => $_has(1);
  @$pb.TagNumber(2)
  void clearLine2() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get building => $_getSZ(2);
  @$pb.TagNumber(3)
  set building($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBuilding() => $_has(2);
  @$pb.TagNumber(3)
  void clearBuilding() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get apartment => $_getSZ(3);
  @$pb.TagNumber(4)
  set apartment($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasApartment() => $_has(3);
  @$pb.TagNumber(4)
  void clearApartment() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get city => $_getSZ(4);
  @$pb.TagNumber(5)
  set city($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCity() => $_has(4);
  @$pb.TagNumber(5)
  void clearCity() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get state => $_getSZ(5);
  @$pb.TagNumber(6)
  set state($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasState() => $_has(5);
  @$pb.TagNumber(6)
  void clearState() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get zipCode => $_getSZ(6);
  @$pb.TagNumber(7)
  set zipCode($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasZipCode() => $_has(6);
  @$pb.TagNumber(7)
  void clearZipCode() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get countryCode => $_getSZ(7);
  @$pb.TagNumber(8)
  set countryCode($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasCountryCode() => $_has(7);
  @$pb.TagNumber(8)
  void clearCountryCode() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get placeID => $_getSZ(8);
  @$pb.TagNumber(9)
  set placeID($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasPlaceID() => $_has(8);
  @$pb.TagNumber(9)
  void clearPlaceID() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get formattedAddress => $_getSZ(9);
  @$pb.TagNumber(10)
  set formattedAddress($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasFormattedAddress() => $_has(9);
  @$pb.TagNumber(10)
  void clearFormattedAddress() => clearField(10);
}

class IsUSPSAddressResponse extends $pb.GeneratedMessage {
  factory IsUSPSAddressResponse({
    $core.bool? valid,
  }) {
    final $result = create();
    if (valid != null) {
      $result.valid = valid;
    }
    return $result;
  }
  IsUSPSAddressResponse._() : super();
  factory IsUSPSAddressResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IsUSPSAddressResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'IsUSPSAddressResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'valid')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IsUSPSAddressResponse clone() => IsUSPSAddressResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IsUSPSAddressResponse copyWith(void Function(IsUSPSAddressResponse) updates) => super.copyWith((message) => updates(message as IsUSPSAddressResponse)) as IsUSPSAddressResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static IsUSPSAddressResponse create() => IsUSPSAddressResponse._();
  IsUSPSAddressResponse createEmptyInstance() => create();
  static $pb.PbList<IsUSPSAddressResponse> createRepeated() => $pb.PbList<IsUSPSAddressResponse>();
  @$core.pragma('dart2js:noInline')
  static IsUSPSAddressResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IsUSPSAddressResponse>(create);
  static IsUSPSAddressResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get valid => $_getBF(0);
  @$pb.TagNumber(1)
  set valid($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValid() => $_has(0);
  @$pb.TagNumber(1)
  void clearValid() => clearField(1);
}

class GetBankAccountWidgetRequest extends $pb.GeneratedMessage {
  factory GetBankAccountWidgetRequest() => create();
  GetBankAccountWidgetRequest._() : super();
  factory GetBankAccountWidgetRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetBankAccountWidgetRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetBankAccountWidgetRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetRequest clone() => GetBankAccountWidgetRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetRequest copyWith(void Function(GetBankAccountWidgetRequest) updates) => super.copyWith((message) => updates(message as GetBankAccountWidgetRequest)) as GetBankAccountWidgetRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetBankAccountWidgetRequest create() => GetBankAccountWidgetRequest._();
  GetBankAccountWidgetRequest createEmptyInstance() => create();
  static $pb.PbList<GetBankAccountWidgetRequest> createRepeated() => $pb.PbList<GetBankAccountWidgetRequest>();
  @$core.pragma('dart2js:noInline')
  static GetBankAccountWidgetRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetBankAccountWidgetRequest>(create);
  static GetBankAccountWidgetRequest? _defaultInstance;
}

class GetBankAccountWidgetResponse extends $pb.GeneratedMessage {
  factory GetBankAccountWidgetResponse({
    $core.String? url,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    return $result;
  }
  GetBankAccountWidgetResponse._() : super();
  factory GetBankAccountWidgetResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetBankAccountWidgetResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetBankAccountWidgetResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetResponse clone() => GetBankAccountWidgetResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetResponse copyWith(void Function(GetBankAccountWidgetResponse) updates) => super.copyWith((message) => updates(message as GetBankAccountWidgetResponse)) as GetBankAccountWidgetResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetBankAccountWidgetResponse create() => GetBankAccountWidgetResponse._();
  GetBankAccountWidgetResponse createEmptyInstance() => create();
  static $pb.PbList<GetBankAccountWidgetResponse> createRepeated() => $pb.PbList<GetBankAccountWidgetResponse>();
  @$core.pragma('dart2js:noInline')
  static GetBankAccountWidgetResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetBankAccountWidgetResponse>(create);
  static GetBankAccountWidgetResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class AddBankAccountRequest extends $pb.GeneratedMessage {
  factory AddBankAccountRequest({
    $core.String? userGuid,
    $core.String? memberGuid,
  }) {
    final $result = create();
    if (userGuid != null) {
      $result.userGuid = userGuid;
    }
    if (memberGuid != null) {
      $result.memberGuid = memberGuid;
    }
    return $result;
  }
  AddBankAccountRequest._() : super();
  factory AddBankAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AddBankAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AddBankAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userGuid', protoName: 'userGuid')
    ..aOS(2, _omitFieldNames ? '' : 'memberGuid', protoName: 'memberGuid')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AddBankAccountRequest clone() => AddBankAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AddBankAccountRequest copyWith(void Function(AddBankAccountRequest) updates) => super.copyWith((message) => updates(message as AddBankAccountRequest)) as AddBankAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AddBankAccountRequest create() => AddBankAccountRequest._();
  AddBankAccountRequest createEmptyInstance() => create();
  static $pb.PbList<AddBankAccountRequest> createRepeated() => $pb.PbList<AddBankAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static AddBankAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AddBankAccountRequest>(create);
  static AddBankAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userGuid => $_getSZ(0);
  @$pb.TagNumber(1)
  set userGuid($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUserGuid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserGuid() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get memberGuid => $_getSZ(1);
  @$pb.TagNumber(2)
  set memberGuid($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMemberGuid() => $_has(1);
  @$pb.TagNumber(2)
  void clearMemberGuid() => clearField(2);
}

class AddBankAccountResponse extends $pb.GeneratedMessage {
  factory AddBankAccountResponse({
    $core.String? fundingsourceId,
  }) {
    final $result = create();
    if (fundingsourceId != null) {
      $result.fundingsourceId = fundingsourceId;
    }
    return $result;
  }
  AddBankAccountResponse._() : super();
  factory AddBankAccountResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AddBankAccountResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AddBankAccountResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'fundingsourceId', protoName: 'fundingsourceId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AddBankAccountResponse clone() => AddBankAccountResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AddBankAccountResponse copyWith(void Function(AddBankAccountResponse) updates) => super.copyWith((message) => updates(message as AddBankAccountResponse)) as AddBankAccountResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AddBankAccountResponse create() => AddBankAccountResponse._();
  AddBankAccountResponse createEmptyInstance() => create();
  static $pb.PbList<AddBankAccountResponse> createRepeated() => $pb.PbList<AddBankAccountResponse>();
  @$core.pragma('dart2js:noInline')
  static AddBankAccountResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AddBankAccountResponse>(create);
  static AddBankAccountResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get fundingsourceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set fundingsourceId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFundingsourceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearFundingsourceId() => clearField(1);
}

class LinkedAccount extends $pb.GeneratedMessage {
  factory LinkedAccount({
    $core.String? id,
    $core.String? type,
    $core.String? name,
    $core.String? mask,
    $core.String? nickname,
    $core.bool? canSend,
    $core.bool? canReceive,
    $core.String? title,
    $core.String? sendCurrencyCode,
    $core.String? sendCurrencyCountryCode,
    $core.String? receiveCurrencyCode,
    $core.String? receiveCurrencyCountryCode,
    $core.bool? defaultSend,
    $core.bool? defaultReceive,
    $core.String? state,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (type != null) {
      $result.type = type;
    }
    if (name != null) {
      $result.name = name;
    }
    if (mask != null) {
      $result.mask = mask;
    }
    if (nickname != null) {
      $result.nickname = nickname;
    }
    if (canSend != null) {
      $result.canSend = canSend;
    }
    if (canReceive != null) {
      $result.canReceive = canReceive;
    }
    if (title != null) {
      $result.title = title;
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
    if (defaultSend != null) {
      $result.defaultSend = defaultSend;
    }
    if (defaultReceive != null) {
      $result.defaultReceive = defaultReceive;
    }
    if (state != null) {
      $result.state = state;
    }
    return $result;
  }
  LinkedAccount._() : super();
  factory LinkedAccount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LinkedAccount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LinkedAccount', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'mask')
    ..aOS(5, _omitFieldNames ? '' : 'nickname')
    ..aOB(6, _omitFieldNames ? '' : 'canSend', protoName: 'canSend')
    ..aOB(7, _omitFieldNames ? '' : 'canReceive', protoName: 'canReceive')
    ..aOS(8, _omitFieldNames ? '' : 'title')
    ..aOS(9, _omitFieldNames ? '' : 'sendCurrencyCode', protoName: 'sendCurrencyCode')
    ..aOS(10, _omitFieldNames ? '' : 'sendCurrencyCountryCode', protoName: 'sendCurrencyCountryCode')
    ..aOS(11, _omitFieldNames ? '' : 'receiveCurrencyCode', protoName: 'receiveCurrencyCode')
    ..aOS(12, _omitFieldNames ? '' : 'receiveCurrencyCountryCode', protoName: 'receiveCurrencyCountryCode')
    ..aOB(13, _omitFieldNames ? '' : 'defaultSend', protoName: 'defaultSend')
    ..aOB(14, _omitFieldNames ? '' : 'defaultReceive', protoName: 'defaultReceive')
    ..aOS(15, _omitFieldNames ? '' : 'state')
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
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get mask => $_getSZ(3);
  @$pb.TagNumber(4)
  set mask($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasMask() => $_has(3);
  @$pb.TagNumber(4)
  void clearMask() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get nickname => $_getSZ(4);
  @$pb.TagNumber(5)
  set nickname($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasNickname() => $_has(4);
  @$pb.TagNumber(5)
  void clearNickname() => clearField(5);

  /// Display
  @$pb.TagNumber(6)
  $core.bool get canSend => $_getBF(5);
  @$pb.TagNumber(6)
  set canSend($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasCanSend() => $_has(5);
  @$pb.TagNumber(6)
  void clearCanSend() => clearField(6);

  @$pb.TagNumber(7)
  $core.bool get canReceive => $_getBF(6);
  @$pb.TagNumber(7)
  set canReceive($core.bool v) { $_setBool(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasCanReceive() => $_has(6);
  @$pb.TagNumber(7)
  void clearCanReceive() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get title => $_getSZ(7);
  @$pb.TagNumber(8)
  set title($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasTitle() => $_has(7);
  @$pb.TagNumber(8)
  void clearTitle() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get sendCurrencyCode => $_getSZ(8);
  @$pb.TagNumber(9)
  set sendCurrencyCode($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasSendCurrencyCode() => $_has(8);
  @$pb.TagNumber(9)
  void clearSendCurrencyCode() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get sendCurrencyCountryCode => $_getSZ(9);
  @$pb.TagNumber(10)
  set sendCurrencyCountryCode($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasSendCurrencyCountryCode() => $_has(9);
  @$pb.TagNumber(10)
  void clearSendCurrencyCountryCode() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get receiveCurrencyCode => $_getSZ(10);
  @$pb.TagNumber(11)
  set receiveCurrencyCode($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasReceiveCurrencyCode() => $_has(10);
  @$pb.TagNumber(11)
  void clearReceiveCurrencyCode() => clearField(11);

  @$pb.TagNumber(12)
  $core.String get receiveCurrencyCountryCode => $_getSZ(11);
  @$pb.TagNumber(12)
  set receiveCurrencyCountryCode($core.String v) { $_setString(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasReceiveCurrencyCountryCode() => $_has(11);
  @$pb.TagNumber(12)
  void clearReceiveCurrencyCountryCode() => clearField(12);

  @$pb.TagNumber(13)
  $core.bool get defaultSend => $_getBF(12);
  @$pb.TagNumber(13)
  set defaultSend($core.bool v) { $_setBool(12, v); }
  @$pb.TagNumber(13)
  $core.bool hasDefaultSend() => $_has(12);
  @$pb.TagNumber(13)
  void clearDefaultSend() => clearField(13);

  @$pb.TagNumber(14)
  $core.bool get defaultReceive => $_getBF(13);
  @$pb.TagNumber(14)
  set defaultReceive($core.bool v) { $_setBool(13, v); }
  @$pb.TagNumber(14)
  $core.bool hasDefaultReceive() => $_has(13);
  @$pb.TagNumber(14)
  void clearDefaultReceive() => clearField(14);

  @$pb.TagNumber(15)
  $core.String get state => $_getSZ(14);
  @$pb.TagNumber(15)
  set state($core.String v) { $_setString(14, v); }
  @$pb.TagNumber(15)
  $core.bool hasState() => $_has(14);
  @$pb.TagNumber(15)
  void clearState() => clearField(15);
}

class GetSignupRequest extends $pb.GeneratedMessage {
  factory GetSignupRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetSignupRequest._() : super();
  factory GetSignupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetSignupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetSignupRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetSignupRequest clone() => GetSignupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetSignupRequest copyWith(void Function(GetSignupRequest) updates) => super.copyWith((message) => updates(message as GetSignupRequest)) as GetSignupRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetSignupRequest create() => GetSignupRequest._();
  GetSignupRequest createEmptyInstance() => create();
  static $pb.PbList<GetSignupRequest> createRepeated() => $pb.PbList<GetSignupRequest>();
  @$core.pragma('dart2js:noInline')
  static GetSignupRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetSignupRequest>(create);
  static GetSignupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class SetSignupUserDataRequest extends $pb.GeneratedMessage {
  factory SetSignupUserDataRequest({
    $core.String? id,
    $core.String? firstName,
    $core.String? lastName,
    $core.String? email,
    $core.String? countryCode,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (firstName != null) {
      $result.firstName = firstName;
    }
    if (lastName != null) {
      $result.lastName = lastName;
    }
    if (email != null) {
      $result.email = email;
    }
    if (countryCode != null) {
      $result.countryCode = countryCode;
    }
    return $result;
  }
  SetSignupUserDataRequest._() : super();
  factory SetSignupUserDataRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupUserDataRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetSignupUserDataRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'firstName', protoName: 'firstName')
    ..aOS(3, _omitFieldNames ? '' : 'lastName', protoName: 'lastName')
    ..aOS(4, _omitFieldNames ? '' : 'email')
    ..aOS(5, _omitFieldNames ? '' : 'countryCode', protoName: 'countryCode')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupUserDataRequest clone() => SetSignupUserDataRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupUserDataRequest copyWith(void Function(SetSignupUserDataRequest) updates) => super.copyWith((message) => updates(message as SetSignupUserDataRequest)) as SetSignupUserDataRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetSignupUserDataRequest create() => SetSignupUserDataRequest._();
  SetSignupUserDataRequest createEmptyInstance() => create();
  static $pb.PbList<SetSignupUserDataRequest> createRepeated() => $pb.PbList<SetSignupUserDataRequest>();
  @$core.pragma('dart2js:noInline')
  static SetSignupUserDataRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetSignupUserDataRequest>(create);
  static SetSignupUserDataRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

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
  $core.String get email => $_getSZ(3);
  @$pb.TagNumber(4)
  set email($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasEmail() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmail() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get countryCode => $_getSZ(4);
  @$pb.TagNumber(5)
  set countryCode($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCountryCode() => $_has(4);
  @$pb.TagNumber(5)
  void clearCountryCode() => clearField(5);
}

class SetSignupUserDataResponse extends $pb.GeneratedMessage {
  factory SetSignupUserDataResponse({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  SetSignupUserDataResponse._() : super();
  factory SetSignupUserDataResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupUserDataResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetSignupUserDataResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupUserDataResponse clone() => SetSignupUserDataResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupUserDataResponse copyWith(void Function(SetSignupUserDataResponse) updates) => super.copyWith((message) => updates(message as SetSignupUserDataResponse)) as SetSignupUserDataResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetSignupUserDataResponse create() => SetSignupUserDataResponse._();
  SetSignupUserDataResponse createEmptyInstance() => create();
  static $pb.PbList<SetSignupUserDataResponse> createRepeated() => $pb.PbList<SetSignupUserDataResponse>();
  @$core.pragma('dart2js:noInline')
  static SetSignupUserDataResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetSignupUserDataResponse>(create);
  static SetSignupUserDataResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class SetSignupMobileNumberRequest extends $pb.GeneratedMessage {
  factory SetSignupMobileNumberRequest({
    $core.String? id,
    $core.String? mobile,
    $core.String? otp,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (mobile != null) {
      $result.mobile = mobile;
    }
    if (otp != null) {
      $result.otp = otp;
    }
    return $result;
  }
  SetSignupMobileNumberRequest._() : super();
  factory SetSignupMobileNumberRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupMobileNumberRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetSignupMobileNumberRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'mobile')
    ..aOS(3, _omitFieldNames ? '' : 'otp')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupMobileNumberRequest clone() => SetSignupMobileNumberRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupMobileNumberRequest copyWith(void Function(SetSignupMobileNumberRequest) updates) => super.copyWith((message) => updates(message as SetSignupMobileNumberRequest)) as SetSignupMobileNumberRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetSignupMobileNumberRequest create() => SetSignupMobileNumberRequest._();
  SetSignupMobileNumberRequest createEmptyInstance() => create();
  static $pb.PbList<SetSignupMobileNumberRequest> createRepeated() => $pb.PbList<SetSignupMobileNumberRequest>();
  @$core.pragma('dart2js:noInline')
  static SetSignupMobileNumberRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetSignupMobileNumberRequest>(create);
  static SetSignupMobileNumberRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get mobile => $_getSZ(1);
  @$pb.TagNumber(2)
  set mobile($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMobile() => $_has(1);
  @$pb.TagNumber(2)
  void clearMobile() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get otp => $_getSZ(2);
  @$pb.TagNumber(3)
  set otp($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasOtp() => $_has(2);
  @$pb.TagNumber(3)
  void clearOtp() => clearField(3);
}

class Signup extends $pb.GeneratedMessage {
  factory Signup({
    $core.String? id,
    $core.String? firstName,
    $core.String? lastName,
    $core.String? email,
    $core.String? countryCode,
    $core.String? mobileNumber,
    $core.String? userId,
    $core.bool? completed,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (firstName != null) {
      $result.firstName = firstName;
    }
    if (lastName != null) {
      $result.lastName = lastName;
    }
    if (email != null) {
      $result.email = email;
    }
    if (countryCode != null) {
      $result.countryCode = countryCode;
    }
    if (mobileNumber != null) {
      $result.mobileNumber = mobileNumber;
    }
    if (userId != null) {
      $result.userId = userId;
    }
    if (completed != null) {
      $result.completed = completed;
    }
    return $result;
  }
  Signup._() : super();
  factory Signup.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Signup.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Signup', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'firstName', protoName: 'firstName')
    ..aOS(3, _omitFieldNames ? '' : 'lastName', protoName: 'lastName')
    ..aOS(4, _omitFieldNames ? '' : 'email')
    ..aOS(5, _omitFieldNames ? '' : 'countryCode', protoName: 'countryCode')
    ..aOS(6, _omitFieldNames ? '' : 'mobileNumber', protoName: 'mobileNumber')
    ..aOS(7, _omitFieldNames ? '' : 'userId', protoName: 'userId')
    ..aOB(8, _omitFieldNames ? '' : 'completed')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Signup clone() => Signup()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Signup copyWith(void Function(Signup) updates) => super.copyWith((message) => updates(message as Signup)) as Signup;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Signup create() => Signup._();
  Signup createEmptyInstance() => create();
  static $pb.PbList<Signup> createRepeated() => $pb.PbList<Signup>();
  @$core.pragma('dart2js:noInline')
  static Signup getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Signup>(create);
  static Signup? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

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
  $core.String get email => $_getSZ(3);
  @$pb.TagNumber(4)
  set email($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasEmail() => $_has(3);
  @$pb.TagNumber(4)
  void clearEmail() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get countryCode => $_getSZ(4);
  @$pb.TagNumber(5)
  set countryCode($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCountryCode() => $_has(4);
  @$pb.TagNumber(5)
  void clearCountryCode() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get mobileNumber => $_getSZ(5);
  @$pb.TagNumber(6)
  set mobileNumber($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasMobileNumber() => $_has(5);
  @$pb.TagNumber(6)
  void clearMobileNumber() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get userId => $_getSZ(6);
  @$pb.TagNumber(7)
  set userId($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasUserId() => $_has(6);
  @$pb.TagNumber(7)
  void clearUserId() => clearField(7);

  @$pb.TagNumber(8)
  $core.bool get completed => $_getBF(7);
  @$pb.TagNumber(8)
  set completed($core.bool v) { $_setBool(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasCompleted() => $_has(7);
  @$pb.TagNumber(8)
  void clearCompleted() => clearField(8);
}

class CompleteSignupRequest extends $pb.GeneratedMessage {
  factory CompleteSignupRequest({
    $core.String? id,
    $core.String? userId,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (userId != null) {
      $result.userId = userId;
    }
    return $result;
  }
  CompleteSignupRequest._() : super();
  factory CompleteSignupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CompleteSignupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CompleteSignupRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'userId', protoName: 'userId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CompleteSignupRequest clone() => CompleteSignupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CompleteSignupRequest copyWith(void Function(CompleteSignupRequest) updates) => super.copyWith((message) => updates(message as CompleteSignupRequest)) as CompleteSignupRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CompleteSignupRequest create() => CompleteSignupRequest._();
  CompleteSignupRequest createEmptyInstance() => create();
  static $pb.PbList<CompleteSignupRequest> createRepeated() => $pb.PbList<CompleteSignupRequest>();
  @$core.pragma('dart2js:noInline')
  static CompleteSignupRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CompleteSignupRequest>(create);
  static CompleteSignupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => clearField(2);
}

class CreateUserDefaultWalletRequest extends $pb.GeneratedMessage {
  factory CreateUserDefaultWalletRequest({
    $core.String? userID,
  }) {
    final $result = create();
    if (userID != null) {
      $result.userID = userID;
    }
    return $result;
  }
  CreateUserDefaultWalletRequest._() : super();
  factory CreateUserDefaultWalletRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateUserDefaultWalletRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateUserDefaultWalletRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userID', protoName: 'userID')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateUserDefaultWalletRequest clone() => CreateUserDefaultWalletRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateUserDefaultWalletRequest copyWith(void Function(CreateUserDefaultWalletRequest) updates) => super.copyWith((message) => updates(message as CreateUserDefaultWalletRequest)) as CreateUserDefaultWalletRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateUserDefaultWalletRequest create() => CreateUserDefaultWalletRequest._();
  CreateUserDefaultWalletRequest createEmptyInstance() => create();
  static $pb.PbList<CreateUserDefaultWalletRequest> createRepeated() => $pb.PbList<CreateUserDefaultWalletRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateUserDefaultWalletRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateUserDefaultWalletRequest>(create);
  static CreateUserDefaultWalletRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userID => $_getSZ(0);
  @$pb.TagNumber(1)
  set userID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUserID() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserID() => clearField(1);
}

class SendPhoneVerificationRequest extends $pb.GeneratedMessage {
  factory SendPhoneVerificationRequest({
    $core.String? to,
  }) {
    final $result = create();
    if (to != null) {
      $result.to = to;
    }
    return $result;
  }
  SendPhoneVerificationRequest._() : super();
  factory SendPhoneVerificationRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SendPhoneVerificationRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SendPhoneVerificationRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'to')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SendPhoneVerificationRequest clone() => SendPhoneVerificationRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SendPhoneVerificationRequest copyWith(void Function(SendPhoneVerificationRequest) updates) => super.copyWith((message) => updates(message as SendPhoneVerificationRequest)) as SendPhoneVerificationRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SendPhoneVerificationRequest create() => SendPhoneVerificationRequest._();
  SendPhoneVerificationRequest createEmptyInstance() => create();
  static $pb.PbList<SendPhoneVerificationRequest> createRepeated() => $pb.PbList<SendPhoneVerificationRequest>();
  @$core.pragma('dart2js:noInline')
  static SendPhoneVerificationRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SendPhoneVerificationRequest>(create);
  static SendPhoneVerificationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get to => $_getSZ(0);
  @$pb.TagNumber(1)
  set to($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTo() => $_has(0);
  @$pb.TagNumber(1)
  void clearTo() => clearField(1);
}

class CheckPhoneVerificationRequest extends $pb.GeneratedMessage {
  factory CheckPhoneVerificationRequest({
    $core.String? to,
    $core.String? otp,
  }) {
    final $result = create();
    if (to != null) {
      $result.to = to;
    }
    if (otp != null) {
      $result.otp = otp;
    }
    return $result;
  }
  CheckPhoneVerificationRequest._() : super();
  factory CheckPhoneVerificationRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CheckPhoneVerificationRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CheckPhoneVerificationRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'to')
    ..aOS(2, _omitFieldNames ? '' : 'otp')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CheckPhoneVerificationRequest clone() => CheckPhoneVerificationRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CheckPhoneVerificationRequest copyWith(void Function(CheckPhoneVerificationRequest) updates) => super.copyWith((message) => updates(message as CheckPhoneVerificationRequest)) as CheckPhoneVerificationRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CheckPhoneVerificationRequest create() => CheckPhoneVerificationRequest._();
  CheckPhoneVerificationRequest createEmptyInstance() => create();
  static $pb.PbList<CheckPhoneVerificationRequest> createRepeated() => $pb.PbList<CheckPhoneVerificationRequest>();
  @$core.pragma('dart2js:noInline')
  static CheckPhoneVerificationRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CheckPhoneVerificationRequest>(create);
  static CheckPhoneVerificationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get to => $_getSZ(0);
  @$pb.TagNumber(1)
  set to($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTo() => $_has(0);
  @$pb.TagNumber(1)
  void clearTo() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get otp => $_getSZ(1);
  @$pb.TagNumber(2)
  set otp($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasOtp() => $_has(1);
  @$pb.TagNumber(2)
  void clearOtp() => clearField(2);
}

class GetAgreementRequest extends $pb.GeneratedMessage {
  factory GetAgreementRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetAgreementRequest._() : super();
  factory GetAgreementRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetAgreementRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetAgreementRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetAgreementRequest clone() => GetAgreementRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetAgreementRequest copyWith(void Function(GetAgreementRequest) updates) => super.copyWith((message) => updates(message as GetAgreementRequest)) as GetAgreementRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetAgreementRequest create() => GetAgreementRequest._();
  GetAgreementRequest createEmptyInstance() => create();
  static $pb.PbList<GetAgreementRequest> createRepeated() => $pb.PbList<GetAgreementRequest>();
  @$core.pragma('dart2js:noInline')
  static GetAgreementRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetAgreementRequest>(create);
  static GetAgreementRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class Agreement extends $pb.GeneratedMessage {
  factory Agreement({
    $core.String? content,
  }) {
    final $result = create();
    if (content != null) {
      $result.content = content;
    }
    return $result;
  }
  Agreement._() : super();
  factory Agreement.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Agreement.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Agreement', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'content')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Agreement clone() => Agreement()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Agreement copyWith(void Function(Agreement) updates) => super.copyWith((message) => updates(message as Agreement)) as Agreement;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Agreement create() => Agreement._();
  Agreement createEmptyInstance() => create();
  static $pb.PbList<Agreement> createRepeated() => $pb.PbList<Agreement>();
  @$core.pragma('dart2js:noInline')
  static Agreement getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Agreement>(create);
  static Agreement? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get content => $_getSZ(0);
  @$pb.TagNumber(1)
  set content($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasContent() => $_has(0);
  @$pb.TagNumber(1)
  void clearContent() => clearField(1);
}

class SignAgreementsRequest extends $pb.GeneratedMessage {
  factory SignAgreementsRequest({
    $core.Iterable<$core.String>? agreementIds,
    $core.String? userId,
    $core.String? ipAddress,
  }) {
    final $result = create();
    if (agreementIds != null) {
      $result.agreementIds.addAll(agreementIds);
    }
    if (userId != null) {
      $result.userId = userId;
    }
    if (ipAddress != null) {
      $result.ipAddress = ipAddress;
    }
    return $result;
  }
  SignAgreementsRequest._() : super();
  factory SignAgreementsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SignAgreementsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SignAgreementsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'agreementIds', protoName: 'agreementIds')
    ..aOS(2, _omitFieldNames ? '' : 'userId', protoName: 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'ipAddress', protoName: 'ipAddress')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SignAgreementsRequest clone() => SignAgreementsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SignAgreementsRequest copyWith(void Function(SignAgreementsRequest) updates) => super.copyWith((message) => updates(message as SignAgreementsRequest)) as SignAgreementsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SignAgreementsRequest create() => SignAgreementsRequest._();
  SignAgreementsRequest createEmptyInstance() => create();
  static $pb.PbList<SignAgreementsRequest> createRepeated() => $pb.PbList<SignAgreementsRequest>();
  @$core.pragma('dart2js:noInline')
  static SignAgreementsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SignAgreementsRequest>(create);
  static SignAgreementsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get agreementIds => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get ipAddress => $_getSZ(2);
  @$pb.TagNumber(3)
  set ipAddress($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasIpAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearIpAddress() => clearField(3);
}

class SignAgreementsResponse extends $pb.GeneratedMessage {
  factory SignAgreementsResponse({
    $core.bool? signed,
  }) {
    final $result = create();
    if (signed != null) {
      $result.signed = signed;
    }
    return $result;
  }
  SignAgreementsResponse._() : super();
  factory SignAgreementsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SignAgreementsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SignAgreementsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'signed')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SignAgreementsResponse clone() => SignAgreementsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SignAgreementsResponse copyWith(void Function(SignAgreementsResponse) updates) => super.copyWith((message) => updates(message as SignAgreementsResponse)) as SignAgreementsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SignAgreementsResponse create() => SignAgreementsResponse._();
  SignAgreementsResponse createEmptyInstance() => create();
  static $pb.PbList<SignAgreementsResponse> createRepeated() => $pb.PbList<SignAgreementsResponse>();
  @$core.pragma('dart2js:noInline')
  static SignAgreementsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SignAgreementsResponse>(create);
  static SignAgreementsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get signed => $_getBF(0);
  @$pb.TagNumber(1)
  set signed($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSigned() => $_has(0);
  @$pb.TagNumber(1)
  void clearSigned() => clearField(1);
}

class JoinWaitlistRequest extends $pb.GeneratedMessage {
  factory JoinWaitlistRequest({
    $core.String? email,
    $core.String? countryCode,
    $core.String? fullName,
    $core.bool? betaOptIn,
    $core.String? mugId,
  }) {
    final $result = create();
    if (email != null) {
      $result.email = email;
    }
    if (countryCode != null) {
      $result.countryCode = countryCode;
    }
    if (fullName != null) {
      $result.fullName = fullName;
    }
    if (betaOptIn != null) {
      $result.betaOptIn = betaOptIn;
    }
    if (mugId != null) {
      $result.mugId = mugId;
    }
    return $result;
  }
  JoinWaitlistRequest._() : super();
  factory JoinWaitlistRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory JoinWaitlistRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'JoinWaitlistRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'email')
    ..aOS(2, _omitFieldNames ? '' : 'countryCode')
    ..aOS(3, _omitFieldNames ? '' : 'fullName')
    ..aOB(4, _omitFieldNames ? '' : 'betaOptIn')
    ..aOS(5, _omitFieldNames ? '' : 'mugId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  JoinWaitlistRequest clone() => JoinWaitlistRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  JoinWaitlistRequest copyWith(void Function(JoinWaitlistRequest) updates) => super.copyWith((message) => updates(message as JoinWaitlistRequest)) as JoinWaitlistRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static JoinWaitlistRequest create() => JoinWaitlistRequest._();
  JoinWaitlistRequest createEmptyInstance() => create();
  static $pb.PbList<JoinWaitlistRequest> createRepeated() => $pb.PbList<JoinWaitlistRequest>();
  @$core.pragma('dart2js:noInline')
  static JoinWaitlistRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<JoinWaitlistRequest>(create);
  static JoinWaitlistRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get email => $_getSZ(0);
  @$pb.TagNumber(1)
  set email($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasEmail() => $_has(0);
  @$pb.TagNumber(1)
  void clearEmail() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get countryCode => $_getSZ(1);
  @$pb.TagNumber(2)
  set countryCode($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCountryCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCountryCode() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get fullName => $_getSZ(2);
  @$pb.TagNumber(3)
  set fullName($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasFullName() => $_has(2);
  @$pb.TagNumber(3)
  void clearFullName() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get betaOptIn => $_getBF(3);
  @$pb.TagNumber(4)
  set betaOptIn($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasBetaOptIn() => $_has(3);
  @$pb.TagNumber(4)
  void clearBetaOptIn() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get mugId => $_getSZ(4);
  @$pb.TagNumber(5)
  set mugId($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasMugId() => $_has(4);
  @$pb.TagNumber(5)
  void clearMugId() => clearField(5);
}

class JoinWaitlistResponse extends $pb.GeneratedMessage {
  factory JoinWaitlistResponse() => create();
  JoinWaitlistResponse._() : super();
  factory JoinWaitlistResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory JoinWaitlistResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'JoinWaitlistResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  JoinWaitlistResponse clone() => JoinWaitlistResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  JoinWaitlistResponse copyWith(void Function(JoinWaitlistResponse) updates) => super.copyWith((message) => updates(message as JoinWaitlistResponse)) as JoinWaitlistResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static JoinWaitlistResponse create() => JoinWaitlistResponse._();
  JoinWaitlistResponse createEmptyInstance() => create();
  static $pb.PbList<JoinWaitlistResponse> createRepeated() => $pb.PbList<JoinWaitlistResponse>();
  @$core.pragma('dart2js:noInline')
  static JoinWaitlistResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<JoinWaitlistResponse>(create);
  static JoinWaitlistResponse? _defaultInstance;
}

class IsMugAvailableRequest extends $pb.GeneratedMessage {
  factory IsMugAvailableRequest({
    $core.String? mugId,
  }) {
    final $result = create();
    if (mugId != null) {
      $result.mugId = mugId;
    }
    return $result;
  }
  IsMugAvailableRequest._() : super();
  factory IsMugAvailableRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IsMugAvailableRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'IsMugAvailableRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'mugId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IsMugAvailableRequest clone() => IsMugAvailableRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IsMugAvailableRequest copyWith(void Function(IsMugAvailableRequest) updates) => super.copyWith((message) => updates(message as IsMugAvailableRequest)) as IsMugAvailableRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static IsMugAvailableRequest create() => IsMugAvailableRequest._();
  IsMugAvailableRequest createEmptyInstance() => create();
  static $pb.PbList<IsMugAvailableRequest> createRepeated() => $pb.PbList<IsMugAvailableRequest>();
  @$core.pragma('dart2js:noInline')
  static IsMugAvailableRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IsMugAvailableRequest>(create);
  static IsMugAvailableRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get mugId => $_getSZ(0);
  @$pb.TagNumber(1)
  set mugId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMugId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMugId() => clearField(1);
}

class IsMugAvailableResponse extends $pb.GeneratedMessage {
  factory IsMugAvailableResponse({
    $core.bool? available,
  }) {
    final $result = create();
    if (available != null) {
      $result.available = available;
    }
    return $result;
  }
  IsMugAvailableResponse._() : super();
  factory IsMugAvailableResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IsMugAvailableResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'IsMugAvailableResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'available')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IsMugAvailableResponse clone() => IsMugAvailableResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IsMugAvailableResponse copyWith(void Function(IsMugAvailableResponse) updates) => super.copyWith((message) => updates(message as IsMugAvailableResponse)) as IsMugAvailableResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static IsMugAvailableResponse create() => IsMugAvailableResponse._();
  IsMugAvailableResponse createEmptyInstance() => create();
  static $pb.PbList<IsMugAvailableResponse> createRepeated() => $pb.PbList<IsMugAvailableResponse>();
  @$core.pragma('dart2js:noInline')
  static IsMugAvailableResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IsMugAvailableResponse>(create);
  static IsMugAvailableResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get available => $_getBF(0);
  @$pb.TagNumber(1)
  set available($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAvailable() => $_has(0);
  @$pb.TagNumber(1)
  void clearAvailable() => clearField(1);
}

class GetLinkedAccountsResponse extends $pb.GeneratedMessage {
  factory GetLinkedAccountsResponse({
    $core.Iterable<LinkedAccount>? linkedAccounts,
  }) {
    final $result = create();
    if (linkedAccounts != null) {
      $result.linkedAccounts.addAll(linkedAccounts);
    }
    return $result;
  }
  GetLinkedAccountsResponse._() : super();
  factory GetLinkedAccountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLinkedAccountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLinkedAccountsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<LinkedAccount>(1, _omitFieldNames ? '' : 'linkedAccounts', $pb.PbFieldType.PM, protoName: 'linkedAccounts', subBuilder: LinkedAccount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsResponse clone() => GetLinkedAccountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsResponse copyWith(void Function(GetLinkedAccountsResponse) updates) => super.copyWith((message) => updates(message as GetLinkedAccountsResponse)) as GetLinkedAccountsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountsResponse create() => GetLinkedAccountsResponse._();
  GetLinkedAccountsResponse createEmptyInstance() => create();
  static $pb.PbList<GetLinkedAccountsResponse> createRepeated() => $pb.PbList<GetLinkedAccountsResponse>();
  @$core.pragma('dart2js:noInline')
  static GetLinkedAccountsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetLinkedAccountsResponse>(create);
  static GetLinkedAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<LinkedAccount> get linkedAccounts => $_getList(0);
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

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLinkedAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
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

class SetNicknameLinkedAccountRequest extends $pb.GeneratedMessage {
  factory SetNicknameLinkedAccountRequest({
    $core.String? id,
    $core.String? nickname,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (nickname != null) {
      $result.nickname = nickname;
    }
    return $result;
  }
  SetNicknameLinkedAccountRequest._() : super();
  factory SetNicknameLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetNicknameLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetNicknameLinkedAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'nickname')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetNicknameLinkedAccountRequest clone() => SetNicknameLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetNicknameLinkedAccountRequest copyWith(void Function(SetNicknameLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as SetNicknameLinkedAccountRequest)) as SetNicknameLinkedAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetNicknameLinkedAccountRequest create() => SetNicknameLinkedAccountRequest._();
  SetNicknameLinkedAccountRequest createEmptyInstance() => create();
  static $pb.PbList<SetNicknameLinkedAccountRequest> createRepeated() => $pb.PbList<SetNicknameLinkedAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static SetNicknameLinkedAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetNicknameLinkedAccountRequest>(create);
  static SetNicknameLinkedAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get nickname => $_getSZ(1);
  @$pb.TagNumber(2)
  set nickname($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNickname() => $_has(1);
  @$pb.TagNumber(2)
  void clearNickname() => clearField(2);
}

class DeleteLinkedAccountRequest extends $pb.GeneratedMessage {
  factory DeleteLinkedAccountRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  DeleteLinkedAccountRequest._() : super();
  factory DeleteLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DeleteLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'DeleteLinkedAccountRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DeleteLinkedAccountRequest clone() => DeleteLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DeleteLinkedAccountRequest copyWith(void Function(DeleteLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as DeleteLinkedAccountRequest)) as DeleteLinkedAccountRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteLinkedAccountRequest create() => DeleteLinkedAccountRequest._();
  DeleteLinkedAccountRequest createEmptyInstance() => create();
  static $pb.PbList<DeleteLinkedAccountRequest> createRepeated() => $pb.PbList<DeleteLinkedAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static DeleteLinkedAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DeleteLinkedAccountRequest>(create);
  static DeleteLinkedAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class Country extends $pb.GeneratedMessage {
  factory Country({
    $core.String? id,
    $core.String? name,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (name != null) {
      $result.name = name;
    }
    return $result;
  }
  Country._() : super();
  factory Country.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Country.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Country', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
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
}

class GetCountriesResponse extends $pb.GeneratedMessage {
  factory GetCountriesResponse({
    $core.Iterable<Country>? countries,
  }) {
    final $result = create();
    if (countries != null) {
      $result.countries.addAll(countries);
    }
    return $result;
  }
  GetCountriesResponse._() : super();
  factory GetCountriesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetCountriesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetCountriesResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Country>(1, _omitFieldNames ? '' : 'countries', $pb.PbFieldType.PM, subBuilder: Country.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetCountriesResponse clone() => GetCountriesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetCountriesResponse copyWith(void Function(GetCountriesResponse) updates) => super.copyWith((message) => updates(message as GetCountriesResponse)) as GetCountriesResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetCountriesResponse create() => GetCountriesResponse._();
  GetCountriesResponse createEmptyInstance() => create();
  static $pb.PbList<GetCountriesResponse> createRepeated() => $pb.PbList<GetCountriesResponse>();
  @$core.pragma('dart2js:noInline')
  static GetCountriesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetCountriesResponse>(create);
  static GetCountriesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Country> get countries => $_getList(0);
}

class CanSignupRequest extends $pb.GeneratedMessage {
  factory CanSignupRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  CanSignupRequest._() : super();
  factory CanSignupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CanSignupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CanSignupRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CanSignupRequest clone() => CanSignupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CanSignupRequest copyWith(void Function(CanSignupRequest) updates) => super.copyWith((message) => updates(message as CanSignupRequest)) as CanSignupRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CanSignupRequest create() => CanSignupRequest._();
  CanSignupRequest createEmptyInstance() => create();
  static $pb.PbList<CanSignupRequest> createRepeated() => $pb.PbList<CanSignupRequest>();
  @$core.pragma('dart2js:noInline')
  static CanSignupRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CanSignupRequest>(create);
  static CanSignupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class CanSignupResponse extends $pb.GeneratedMessage {
  factory CanSignupResponse({
    $core.bool? canSignup,
  }) {
    final $result = create();
    if (canSignup != null) {
      $result.canSignup = canSignup;
    }
    return $result;
  }
  CanSignupResponse._() : super();
  factory CanSignupResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CanSignupResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CanSignupResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'canSignup', protoName: 'canSignup')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CanSignupResponse clone() => CanSignupResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CanSignupResponse copyWith(void Function(CanSignupResponse) updates) => super.copyWith((message) => updates(message as CanSignupResponse)) as CanSignupResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CanSignupResponse create() => CanSignupResponse._();
  CanSignupResponse createEmptyInstance() => create();
  static $pb.PbList<CanSignupResponse> createRepeated() => $pb.PbList<CanSignupResponse>();
  @$core.pragma('dart2js:noInline')
  static CanSignupResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CanSignupResponse>(create);
  static CanSignupResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get canSignup => $_getBF(0);
  @$pb.TagNumber(1)
  set canSignup($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCanSignup() => $_has(0);
  @$pb.TagNumber(1)
  void clearCanSignup() => clearField(1);
}

class SetSignupCompleteRequest extends $pb.GeneratedMessage {
  factory SetSignupCompleteRequest({
    $core.String? id,
    $core.String? userId,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (userId != null) {
      $result.userId = userId;
    }
    return $result;
  }
  SetSignupCompleteRequest._() : super();
  factory SetSignupCompleteRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupCompleteRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetSignupCompleteRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'userId', protoName: 'userId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupCompleteRequest clone() => SetSignupCompleteRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupCompleteRequest copyWith(void Function(SetSignupCompleteRequest) updates) => super.copyWith((message) => updates(message as SetSignupCompleteRequest)) as SetSignupCompleteRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetSignupCompleteRequest create() => SetSignupCompleteRequest._();
  SetSignupCompleteRequest createEmptyInstance() => create();
  static $pb.PbList<SetSignupCompleteRequest> createRepeated() => $pb.PbList<SetSignupCompleteRequest>();
  @$core.pragma('dart2js:noInline')
  static SetSignupCompleteRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetSignupCompleteRequest>(create);
  static SetSignupCompleteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => clearField(2);
}

class LookupTransactionRequest extends $pb.GeneratedMessage {
  factory LookupTransactionRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  LookupTransactionRequest._() : super();
  factory LookupTransactionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LookupTransactionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LookupTransactionRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LookupTransactionRequest clone() => LookupTransactionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LookupTransactionRequest copyWith(void Function(LookupTransactionRequest) updates) => super.copyWith((message) => updates(message as LookupTransactionRequest)) as LookupTransactionRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LookupTransactionRequest create() => LookupTransactionRequest._();
  LookupTransactionRequest createEmptyInstance() => create();
  static $pb.PbList<LookupTransactionRequest> createRepeated() => $pb.PbList<LookupTransactionRequest>();
  @$core.pragma('dart2js:noInline')
  static LookupTransactionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LookupTransactionRequest>(create);
  static LookupTransactionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class GetCurrentWalletResponse extends $pb.GeneratedMessage {
  factory GetCurrentWalletResponse({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetCurrentWalletResponse._() : super();
  factory GetCurrentWalletResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetCurrentWalletResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetCurrentWalletResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetCurrentWalletResponse clone() => GetCurrentWalletResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetCurrentWalletResponse copyWith(void Function(GetCurrentWalletResponse) updates) => super.copyWith((message) => updates(message as GetCurrentWalletResponse)) as GetCurrentWalletResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetCurrentWalletResponse create() => GetCurrentWalletResponse._();
  GetCurrentWalletResponse createEmptyInstance() => create();
  static $pb.PbList<GetCurrentWalletResponse> createRepeated() => $pb.PbList<GetCurrentWalletResponse>();
  @$core.pragma('dart2js:noInline')
  static GetCurrentWalletResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetCurrentWalletResponse>(create);
  static GetCurrentWalletResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class Limit extends $pb.GeneratedMessage {
  factory Limit({
    LimitAmount? annual,
    LimitAmount? daily,
    LimitAmount? monthly,
    LimitAmount? walletHold,
  }) {
    final $result = create();
    if (annual != null) {
      $result.annual = annual;
    }
    if (daily != null) {
      $result.daily = daily;
    }
    if (monthly != null) {
      $result.monthly = monthly;
    }
    if (walletHold != null) {
      $result.walletHold = walletHold;
    }
    return $result;
  }
  Limit._() : super();
  factory Limit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Limit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Limit', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<LimitAmount>(1, _omitFieldNames ? '' : 'Annual', protoName: 'Annual', subBuilder: LimitAmount.create)
    ..aOM<LimitAmount>(2, _omitFieldNames ? '' : 'Daily', protoName: 'Daily', subBuilder: LimitAmount.create)
    ..aOM<LimitAmount>(3, _omitFieldNames ? '' : 'Monthly', protoName: 'Monthly', subBuilder: LimitAmount.create)
    ..aOM<LimitAmount>(4, _omitFieldNames ? '' : 'WalletHold', protoName: 'WalletHold', subBuilder: LimitAmount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Limit clone() => Limit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Limit copyWith(void Function(Limit) updates) => super.copyWith((message) => updates(message as Limit)) as Limit;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Limit create() => Limit._();
  Limit createEmptyInstance() => create();
  static $pb.PbList<Limit> createRepeated() => $pb.PbList<Limit>();
  @$core.pragma('dart2js:noInline')
  static Limit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Limit>(create);
  static Limit? _defaultInstance;

  @$pb.TagNumber(1)
  LimitAmount get annual => $_getN(0);
  @$pb.TagNumber(1)
  set annual(LimitAmount v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasAnnual() => $_has(0);
  @$pb.TagNumber(1)
  void clearAnnual() => clearField(1);
  @$pb.TagNumber(1)
  LimitAmount ensureAnnual() => $_ensure(0);

  @$pb.TagNumber(2)
  LimitAmount get daily => $_getN(1);
  @$pb.TagNumber(2)
  set daily(LimitAmount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasDaily() => $_has(1);
  @$pb.TagNumber(2)
  void clearDaily() => clearField(2);
  @$pb.TagNumber(2)
  LimitAmount ensureDaily() => $_ensure(1);

  @$pb.TagNumber(3)
  LimitAmount get monthly => $_getN(2);
  @$pb.TagNumber(3)
  set monthly(LimitAmount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasMonthly() => $_has(2);
  @$pb.TagNumber(3)
  void clearMonthly() => clearField(3);
  @$pb.TagNumber(3)
  LimitAmount ensureMonthly() => $_ensure(2);

  @$pb.TagNumber(4)
  LimitAmount get walletHold => $_getN(3);
  @$pb.TagNumber(4)
  set walletHold(LimitAmount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasWalletHold() => $_has(3);
  @$pb.TagNumber(4)
  void clearWalletHold() => clearField(4);
  @$pb.TagNumber(4)
  LimitAmount ensureWalletHold() => $_ensure(3);
}

class LimitAmount extends $pb.GeneratedMessage {
  factory LimitAmount({
    $core.String? remaining,
    $core.String? total,
    $core.int? percentage,
  }) {
    final $result = create();
    if (remaining != null) {
      $result.remaining = remaining;
    }
    if (total != null) {
      $result.total = total;
    }
    if (percentage != null) {
      $result.percentage = percentage;
    }
    return $result;
  }
  LimitAmount._() : super();
  factory LimitAmount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LimitAmount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LimitAmount', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'remaining')
    ..aOS(2, _omitFieldNames ? '' : 'total')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'percentage', $pb.PbFieldType.O3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LimitAmount clone() => LimitAmount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LimitAmount copyWith(void Function(LimitAmount) updates) => super.copyWith((message) => updates(message as LimitAmount)) as LimitAmount;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LimitAmount create() => LimitAmount._();
  LimitAmount createEmptyInstance() => create();
  static $pb.PbList<LimitAmount> createRepeated() => $pb.PbList<LimitAmount>();
  @$core.pragma('dart2js:noInline')
  static LimitAmount getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LimitAmount>(create);
  static LimitAmount? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get remaining => $_getSZ(0);
  @$pb.TagNumber(1)
  set remaining($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasRemaining() => $_has(0);
  @$pb.TagNumber(1)
  void clearRemaining() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get total => $_getSZ(1);
  @$pb.TagNumber(2)
  set total($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasTotal() => $_has(1);
  @$pb.TagNumber(2)
  void clearTotal() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get percentage => $_getIZ(2);
  @$pb.TagNumber(3)
  set percentage($core.int v) { $_setSignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPercentage() => $_has(2);
  @$pb.TagNumber(3)
  void clearPercentage() => clearField(3);
}

class WalletAddressExistsRequest extends $pb.GeneratedMessage {
  factory WalletAddressExistsRequest({
    $core.String? url,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    return $result;
  }
  WalletAddressExistsRequest._() : super();
  factory WalletAddressExistsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WalletAddressExistsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WalletAddressExistsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WalletAddressExistsRequest clone() => WalletAddressExistsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WalletAddressExistsRequest copyWith(void Function(WalletAddressExistsRequest) updates) => super.copyWith((message) => updates(message as WalletAddressExistsRequest)) as WalletAddressExistsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WalletAddressExistsRequest create() => WalletAddressExistsRequest._();
  WalletAddressExistsRequest createEmptyInstance() => create();
  static $pb.PbList<WalletAddressExistsRequest> createRepeated() => $pb.PbList<WalletAddressExistsRequest>();
  @$core.pragma('dart2js:noInline')
  static WalletAddressExistsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WalletAddressExistsRequest>(create);
  static WalletAddressExistsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class WalletAddressExistsResponse extends $pb.GeneratedMessage {
  factory WalletAddressExistsResponse({
    $core.bool? exists,
  }) {
    final $result = create();
    if (exists != null) {
      $result.exists = exists;
    }
    return $result;
  }
  WalletAddressExistsResponse._() : super();
  factory WalletAddressExistsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WalletAddressExistsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WalletAddressExistsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'exists')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WalletAddressExistsResponse clone() => WalletAddressExistsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WalletAddressExistsResponse copyWith(void Function(WalletAddressExistsResponse) updates) => super.copyWith((message) => updates(message as WalletAddressExistsResponse)) as WalletAddressExistsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WalletAddressExistsResponse create() => WalletAddressExistsResponse._();
  WalletAddressExistsResponse createEmptyInstance() => create();
  static $pb.PbList<WalletAddressExistsResponse> createRepeated() => $pb.PbList<WalletAddressExistsResponse>();
  @$core.pragma('dart2js:noInline')
  static WalletAddressExistsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WalletAddressExistsResponse>(create);
  static WalletAddressExistsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get exists => $_getBF(0);
  @$pb.TagNumber(1)
  set exists($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasExists() => $_has(0);
  @$pb.TagNumber(1)
  void clearExists() => clearField(1);
}

class CreateWalletAddressRequest extends $pb.GeneratedMessage {
  factory CreateWalletAddressRequest({
    $core.String? url,
    $core.String? asset,
    $core.int? assetScale,
    $core.String? alias,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    if (asset != null) {
      $result.asset = asset;
    }
    if (assetScale != null) {
      $result.assetScale = assetScale;
    }
    if (alias != null) {
      $result.alias = alias;
    }
    return $result;
  }
  CreateWalletAddressRequest._() : super();
  factory CreateWalletAddressRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateWalletAddressRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateWalletAddressRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..aOS(2, _omitFieldNames ? '' : 'asset')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'assetScale', $pb.PbFieldType.O3, protoName: 'assetScale')
    ..aOS(4, _omitFieldNames ? '' : 'alias')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateWalletAddressRequest clone() => CreateWalletAddressRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateWalletAddressRequest copyWith(void Function(CreateWalletAddressRequest) updates) => super.copyWith((message) => updates(message as CreateWalletAddressRequest)) as CreateWalletAddressRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateWalletAddressRequest create() => CreateWalletAddressRequest._();
  CreateWalletAddressRequest createEmptyInstance() => create();
  static $pb.PbList<CreateWalletAddressRequest> createRepeated() => $pb.PbList<CreateWalletAddressRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateWalletAddressRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateWalletAddressRequest>(create);
  static CreateWalletAddressRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);

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
  $core.String get alias => $_getSZ(3);
  @$pb.TagNumber(4)
  set alias($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasAlias() => $_has(3);
  @$pb.TagNumber(4)
  void clearAlias() => clearField(4);
}

class SetWalletNameRequest extends $pb.GeneratedMessage {
  factory SetWalletNameRequest({
    $core.String? name,
  }) {
    final $result = create();
    if (name != null) {
      $result.name = name;
    }
    return $result;
  }
  SetWalletNameRequest._() : super();
  factory SetWalletNameRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetWalletNameRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetWalletNameRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetWalletNameRequest clone() => SetWalletNameRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetWalletNameRequest copyWith(void Function(SetWalletNameRequest) updates) => super.copyWith((message) => updates(message as SetWalletNameRequest)) as SetWalletNameRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetWalletNameRequest create() => SetWalletNameRequest._();
  SetWalletNameRequest createEmptyInstance() => create();
  static $pb.PbList<SetWalletNameRequest> createRepeated() => $pb.PbList<SetWalletNameRequest>();
  @$core.pragma('dart2js:noInline')
  static SetWalletNameRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetWalletNameRequest>(create);
  static SetWalletNameRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => clearField(1);
}

class GetPublicWalletDetailsRequest extends $pb.GeneratedMessage {
  factory GetPublicWalletDetailsRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetPublicWalletDetailsRequest._() : super();
  factory GetPublicWalletDetailsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPublicWalletDetailsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetPublicWalletDetailsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsRequest clone() => GetPublicWalletDetailsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsRequest copyWith(void Function(GetPublicWalletDetailsRequest) updates) => super.copyWith((message) => updates(message as GetPublicWalletDetailsRequest)) as GetPublicWalletDetailsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPublicWalletDetailsRequest create() => GetPublicWalletDetailsRequest._();
  GetPublicWalletDetailsRequest createEmptyInstance() => create();
  static $pb.PbList<GetPublicWalletDetailsRequest> createRepeated() => $pb.PbList<GetPublicWalletDetailsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetPublicWalletDetailsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetPublicWalletDetailsRequest>(create);
  static GetPublicWalletDetailsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class GetPublicWalletDetailsResponse extends $pb.GeneratedMessage {
  factory GetPublicWalletDetailsResponse({
    $core.String? id,
    $core.String? publicName,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (publicName != null) {
      $result.publicName = publicName;
    }
    return $result;
  }
  GetPublicWalletDetailsResponse._() : super();
  factory GetPublicWalletDetailsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPublicWalletDetailsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetPublicWalletDetailsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'publicName', protoName: 'publicName')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsResponse clone() => GetPublicWalletDetailsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsResponse copyWith(void Function(GetPublicWalletDetailsResponse) updates) => super.copyWith((message) => updates(message as GetPublicWalletDetailsResponse)) as GetPublicWalletDetailsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPublicWalletDetailsResponse create() => GetPublicWalletDetailsResponse._();
  GetPublicWalletDetailsResponse createEmptyInstance() => create();
  static $pb.PbList<GetPublicWalletDetailsResponse> createRepeated() => $pb.PbList<GetPublicWalletDetailsResponse>();
  @$core.pragma('dart2js:noInline')
  static GetPublicWalletDetailsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetPublicWalletDetailsResponse>(create);
  static GetPublicWalletDetailsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get publicName => $_getSZ(1);
  @$pb.TagNumber(2)
  set publicName($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPublicName() => $_has(1);
  @$pb.TagNumber(2)
  void clearPublicName() => clearField(2);
}

class ListLimitsResponse extends $pb.GeneratedMessage {
  factory ListLimitsResponse({
    $core.Iterable<ConfiguredLimit>? limits,
  }) {
    final $result = create();
    if (limits != null) {
      $result.limits.addAll(limits);
    }
    return $result;
  }
  ListLimitsResponse._() : super();
  factory ListLimitsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListLimitsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListLimitsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<ConfiguredLimit>(1, _omitFieldNames ? '' : 'limits', $pb.PbFieldType.PM, subBuilder: ConfiguredLimit.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListLimitsResponse clone() => ListLimitsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListLimitsResponse copyWith(void Function(ListLimitsResponse) updates) => super.copyWith((message) => updates(message as ListLimitsResponse)) as ListLimitsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListLimitsResponse create() => ListLimitsResponse._();
  ListLimitsResponse createEmptyInstance() => create();
  static $pb.PbList<ListLimitsResponse> createRepeated() => $pb.PbList<ListLimitsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListLimitsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListLimitsResponse>(create);
  static ListLimitsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<ConfiguredLimit> get limits => $_getList(0);
}

class ConfiguredLimit extends $pb.GeneratedMessage {
  factory ConfiguredLimit({
    $core.String? foreignId,
    $core.String? foreignDisplay,
    $core.String? foreignType,
    Amount? daily,
    Amount? monthly,
    Amount? overall,
  }) {
    final $result = create();
    if (foreignId != null) {
      $result.foreignId = foreignId;
    }
    if (foreignDisplay != null) {
      $result.foreignDisplay = foreignDisplay;
    }
    if (foreignType != null) {
      $result.foreignType = foreignType;
    }
    if (daily != null) {
      $result.daily = daily;
    }
    if (monthly != null) {
      $result.monthly = monthly;
    }
    if (overall != null) {
      $result.overall = overall;
    }
    return $result;
  }
  ConfiguredLimit._() : super();
  factory ConfiguredLimit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfiguredLimit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConfiguredLimit', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'foreignId', protoName: 'foreignId')
    ..aOS(2, _omitFieldNames ? '' : 'foreignDisplay', protoName: 'foreignDisplay')
    ..aOS(3, _omitFieldNames ? '' : 'foreignType', protoName: 'foreignType')
    ..aOM<Amount>(4, _omitFieldNames ? '' : 'daily', subBuilder: Amount.create)
    ..aOM<Amount>(5, _omitFieldNames ? '' : 'monthly', subBuilder: Amount.create)
    ..aOM<Amount>(6, _omitFieldNames ? '' : 'overall', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfiguredLimit clone() => ConfiguredLimit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfiguredLimit copyWith(void Function(ConfiguredLimit) updates) => super.copyWith((message) => updates(message as ConfiguredLimit)) as ConfiguredLimit;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConfiguredLimit create() => ConfiguredLimit._();
  ConfiguredLimit createEmptyInstance() => create();
  static $pb.PbList<ConfiguredLimit> createRepeated() => $pb.PbList<ConfiguredLimit>();
  @$core.pragma('dart2js:noInline')
  static ConfiguredLimit getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfiguredLimit>(create);
  static ConfiguredLimit? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get foreignId => $_getSZ(0);
  @$pb.TagNumber(1)
  set foreignId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasForeignId() => $_has(0);
  @$pb.TagNumber(1)
  void clearForeignId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get foreignDisplay => $_getSZ(1);
  @$pb.TagNumber(2)
  set foreignDisplay($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasForeignDisplay() => $_has(1);
  @$pb.TagNumber(2)
  void clearForeignDisplay() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get foreignType => $_getSZ(2);
  @$pb.TagNumber(3)
  set foreignType($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasForeignType() => $_has(2);
  @$pb.TagNumber(3)
  void clearForeignType() => clearField(3);

  @$pb.TagNumber(4)
  Amount get daily => $_getN(3);
  @$pb.TagNumber(4)
  set daily(Amount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasDaily() => $_has(3);
  @$pb.TagNumber(4)
  void clearDaily() => clearField(4);
  @$pb.TagNumber(4)
  Amount ensureDaily() => $_ensure(3);

  @$pb.TagNumber(5)
  Amount get monthly => $_getN(4);
  @$pb.TagNumber(5)
  set monthly(Amount v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasMonthly() => $_has(4);
  @$pb.TagNumber(5)
  void clearMonthly() => clearField(5);
  @$pb.TagNumber(5)
  Amount ensureMonthly() => $_ensure(4);

  @$pb.TagNumber(6)
  Amount get overall => $_getN(5);
  @$pb.TagNumber(6)
  set overall(Amount v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasOverall() => $_has(5);
  @$pb.TagNumber(6)
  void clearOverall() => clearField(6);
  @$pb.TagNumber(6)
  Amount ensureOverall() => $_ensure(5);
}

class UpdateClientLimitsRequest extends $pb.GeneratedMessage {
  factory UpdateClientLimitsRequest({
    $core.String? clientUrl,
    Amount? daily,
    Amount? monthly,
    Amount? overall,
  }) {
    final $result = create();
    if (clientUrl != null) {
      $result.clientUrl = clientUrl;
    }
    if (daily != null) {
      $result.daily = daily;
    }
    if (monthly != null) {
      $result.monthly = monthly;
    }
    if (overall != null) {
      $result.overall = overall;
    }
    return $result;
  }
  UpdateClientLimitsRequest._() : super();
  factory UpdateClientLimitsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateClientLimitsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UpdateClientLimitsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientUrl', protoName: 'clientUrl')
    ..aOM<Amount>(2, _omitFieldNames ? '' : 'daily', subBuilder: Amount.create)
    ..aOM<Amount>(3, _omitFieldNames ? '' : 'monthly', subBuilder: Amount.create)
    ..aOM<Amount>(4, _omitFieldNames ? '' : 'overall', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateClientLimitsRequest clone() => UpdateClientLimitsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateClientLimitsRequest copyWith(void Function(UpdateClientLimitsRequest) updates) => super.copyWith((message) => updates(message as UpdateClientLimitsRequest)) as UpdateClientLimitsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UpdateClientLimitsRequest create() => UpdateClientLimitsRequest._();
  UpdateClientLimitsRequest createEmptyInstance() => create();
  static $pb.PbList<UpdateClientLimitsRequest> createRepeated() => $pb.PbList<UpdateClientLimitsRequest>();
  @$core.pragma('dart2js:noInline')
  static UpdateClientLimitsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpdateClientLimitsRequest>(create);
  static UpdateClientLimitsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientUrl($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientUrl() => clearField(1);

  @$pb.TagNumber(2)
  Amount get daily => $_getN(1);
  @$pb.TagNumber(2)
  set daily(Amount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasDaily() => $_has(1);
  @$pb.TagNumber(2)
  void clearDaily() => clearField(2);
  @$pb.TagNumber(2)
  Amount ensureDaily() => $_ensure(1);

  @$pb.TagNumber(3)
  Amount get monthly => $_getN(2);
  @$pb.TagNumber(3)
  set monthly(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasMonthly() => $_has(2);
  @$pb.TagNumber(3)
  void clearMonthly() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureMonthly() => $_ensure(2);

  @$pb.TagNumber(4)
  Amount get overall => $_getN(3);
  @$pb.TagNumber(4)
  set overall(Amount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasOverall() => $_has(3);
  @$pb.TagNumber(4)
  void clearOverall() => clearField(4);
  @$pb.TagNumber(4)
  Amount ensureOverall() => $_ensure(3);
}

class Contact extends $pb.GeneratedMessage {
  factory Contact({
    $core.String? id,
    $core.String? paymentPointer,
    $core.String? name,
    $core.String? walletId,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (paymentPointer != null) {
      $result.paymentPointer = paymentPointer;
    }
    if (name != null) {
      $result.name = name;
    }
    if (walletId != null) {
      $result.walletId = walletId;
    }
    return $result;
  }
  Contact._() : super();
  factory Contact.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Contact.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Contact', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'paymentPointer')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'walletId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Contact clone() => Contact()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Contact copyWith(void Function(Contact) updates) => super.copyWith((message) => updates(message as Contact)) as Contact;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Contact create() => Contact._();
  Contact createEmptyInstance() => create();
  static $pb.PbList<Contact> createRepeated() => $pb.PbList<Contact>();
  @$core.pragma('dart2js:noInline')
  static Contact getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Contact>(create);
  static Contact? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get paymentPointer => $_getSZ(1);
  @$pb.TagNumber(2)
  set paymentPointer($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPaymentPointer() => $_has(1);
  @$pb.TagNumber(2)
  void clearPaymentPointer() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get walletId => $_getSZ(3);
  @$pb.TagNumber(4)
  set walletId($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasWalletId() => $_has(3);
  @$pb.TagNumber(4)
  void clearWalletId() => clearField(4);
}

class ListContactsRequest extends $pb.GeneratedMessage {
  factory ListContactsRequest({
    $core.int? pageSize,
    $core.String? pageToken,
    $core.String? orderBy,
  }) {
    final $result = create();
    if (pageSize != null) {
      $result.pageSize = pageSize;
    }
    if (pageToken != null) {
      $result.pageToken = pageToken;
    }
    if (orderBy != null) {
      $result.orderBy = orderBy;
    }
    return $result;
  }
  ListContactsRequest._() : super();
  factory ListContactsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListContactsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListContactsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3)
    ..aOS(2, _omitFieldNames ? '' : 'pageToken')
    ..aOS(3, _omitFieldNames ? '' : 'orderBy')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListContactsRequest clone() => ListContactsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListContactsRequest copyWith(void Function(ListContactsRequest) updates) => super.copyWith((message) => updates(message as ListContactsRequest)) as ListContactsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListContactsRequest create() => ListContactsRequest._();
  ListContactsRequest createEmptyInstance() => create();
  static $pb.PbList<ListContactsRequest> createRepeated() => $pb.PbList<ListContactsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListContactsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListContactsRequest>(create);
  static ListContactsRequest? _defaultInstance;

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

  @$pb.TagNumber(3)
  $core.String get orderBy => $_getSZ(2);
  @$pb.TagNumber(3)
  set orderBy($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasOrderBy() => $_has(2);
  @$pb.TagNumber(3)
  void clearOrderBy() => clearField(3);
}

class ListContactsResponse extends $pb.GeneratedMessage {
  factory ListContactsResponse({
    $core.Iterable<Contact>? contacts,
    $core.String? nextPageToken,
  }) {
    final $result = create();
    if (contacts != null) {
      $result.contacts.addAll(contacts);
    }
    if (nextPageToken != null) {
      $result.nextPageToken = nextPageToken;
    }
    return $result;
  }
  ListContactsResponse._() : super();
  factory ListContactsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListContactsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListContactsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Contact>(1, _omitFieldNames ? '' : 'contacts', $pb.PbFieldType.PM, subBuilder: Contact.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListContactsResponse clone() => ListContactsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListContactsResponse copyWith(void Function(ListContactsResponse) updates) => super.copyWith((message) => updates(message as ListContactsResponse)) as ListContactsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListContactsResponse create() => ListContactsResponse._();
  ListContactsResponse createEmptyInstance() => create();
  static $pb.PbList<ListContactsResponse> createRepeated() => $pb.PbList<ListContactsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListContactsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListContactsResponse>(create);
  static ListContactsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Contact> get contacts => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => clearField(2);
}

class CreateContactRequest extends $pb.GeneratedMessage {
  factory CreateContactRequest({
    $core.String? paymentPointer,
  }) {
    final $result = create();
    if (paymentPointer != null) {
      $result.paymentPointer = paymentPointer;
    }
    return $result;
  }
  CreateContactRequest._() : super();
  factory CreateContactRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateContactRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateContactRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'paymentPointer')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateContactRequest clone() => CreateContactRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateContactRequest copyWith(void Function(CreateContactRequest) updates) => super.copyWith((message) => updates(message as CreateContactRequest)) as CreateContactRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateContactRequest create() => CreateContactRequest._();
  CreateContactRequest createEmptyInstance() => create();
  static $pb.PbList<CreateContactRequest> createRepeated() => $pb.PbList<CreateContactRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateContactRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateContactRequest>(create);
  static CreateContactRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get paymentPointer => $_getSZ(0);
  @$pb.TagNumber(1)
  set paymentPointer($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPaymentPointer() => $_has(0);
  @$pb.TagNumber(1)
  void clearPaymentPointer() => clearField(1);
}

class ListIdentitiesResponse extends $pb.GeneratedMessage {
  factory ListIdentitiesResponse({
    $core.Iterable<Identity>? identities,
  }) {
    final $result = create();
    if (identities != null) {
      $result.identities.addAll(identities);
    }
    return $result;
  }
  ListIdentitiesResponse._() : super();
  factory ListIdentitiesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListIdentitiesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListIdentitiesResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Identity>(1, _omitFieldNames ? '' : 'identities', $pb.PbFieldType.PM, subBuilder: Identity.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListIdentitiesResponse clone() => ListIdentitiesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListIdentitiesResponse copyWith(void Function(ListIdentitiesResponse) updates) => super.copyWith((message) => updates(message as ListIdentitiesResponse)) as ListIdentitiesResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListIdentitiesResponse create() => ListIdentitiesResponse._();
  ListIdentitiesResponse createEmptyInstance() => create();
  static $pb.PbList<ListIdentitiesResponse> createRepeated() => $pb.PbList<ListIdentitiesResponse>();
  @$core.pragma('dart2js:noInline')
  static ListIdentitiesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListIdentitiesResponse>(create);
  static ListIdentitiesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Identity> get identities => $_getList(0);
}

class Identity extends $pb.GeneratedMessage {
  factory Identity({
    $core.String? id,
    $core.String? wallet,
    $core.String? platform,
    $core.String? identifier,
    $core.String? state,
    $core.String? keyId,
    $core.String? signature,
    $core.String? signatureHash,
    $core.String? proof,
    $core.String? ctime,
    $core.bool? public,
    $6.Timestamp? verifiedAt,
    $core.String? walletId,
    $core.String? txtRecord,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (wallet != null) {
      $result.wallet = wallet;
    }
    if (platform != null) {
      $result.platform = platform;
    }
    if (identifier != null) {
      $result.identifier = identifier;
    }
    if (state != null) {
      $result.state = state;
    }
    if (keyId != null) {
      $result.keyId = keyId;
    }
    if (signature != null) {
      $result.signature = signature;
    }
    if (signatureHash != null) {
      $result.signatureHash = signatureHash;
    }
    if (proof != null) {
      $result.proof = proof;
    }
    if (ctime != null) {
      $result.ctime = ctime;
    }
    if (public != null) {
      $result.public = public;
    }
    if (verifiedAt != null) {
      $result.verifiedAt = verifiedAt;
    }
    if (walletId != null) {
      $result.walletId = walletId;
    }
    if (txtRecord != null) {
      $result.txtRecord = txtRecord;
    }
    return $result;
  }
  Identity._() : super();
  factory Identity.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Identity.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Identity', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'wallet')
    ..aOS(3, _omitFieldNames ? '' : 'platform')
    ..aOS(4, _omitFieldNames ? '' : 'identifier')
    ..aOS(5, _omitFieldNames ? '' : 'state')
    ..aOS(6, _omitFieldNames ? '' : 'keyId')
    ..aOS(7, _omitFieldNames ? '' : 'signature')
    ..aOS(8, _omitFieldNames ? '' : 'signatureHash')
    ..aOS(9, _omitFieldNames ? '' : 'proof')
    ..aOS(10, _omitFieldNames ? '' : 'ctime')
    ..aOB(11, _omitFieldNames ? '' : 'public')
    ..aOM<$6.Timestamp>(12, _omitFieldNames ? '' : 'verifiedAt', subBuilder: $6.Timestamp.create)
    ..aOS(13, _omitFieldNames ? '' : 'walletId')
    ..aOS(14, _omitFieldNames ? '' : 'txtRecord')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Identity clone() => Identity()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Identity copyWith(void Function(Identity) updates) => super.copyWith((message) => updates(message as Identity)) as Identity;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Identity create() => Identity._();
  Identity createEmptyInstance() => create();
  static $pb.PbList<Identity> createRepeated() => $pb.PbList<Identity>();
  @$core.pragma('dart2js:noInline')
  static Identity getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Identity>(create);
  static Identity? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get wallet => $_getSZ(1);
  @$pb.TagNumber(2)
  set wallet($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasWallet() => $_has(1);
  @$pb.TagNumber(2)
  void clearWallet() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get platform => $_getSZ(2);
  @$pb.TagNumber(3)
  set platform($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPlatform() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlatform() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get identifier => $_getSZ(3);
  @$pb.TagNumber(4)
  set identifier($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasIdentifier() => $_has(3);
  @$pb.TagNumber(4)
  void clearIdentifier() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get state => $_getSZ(4);
  @$pb.TagNumber(5)
  set state($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasState() => $_has(4);
  @$pb.TagNumber(5)
  void clearState() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get keyId => $_getSZ(5);
  @$pb.TagNumber(6)
  set keyId($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasKeyId() => $_has(5);
  @$pb.TagNumber(6)
  void clearKeyId() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get signature => $_getSZ(6);
  @$pb.TagNumber(7)
  set signature($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasSignature() => $_has(6);
  @$pb.TagNumber(7)
  void clearSignature() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get signatureHash => $_getSZ(7);
  @$pb.TagNumber(8)
  set signatureHash($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasSignatureHash() => $_has(7);
  @$pb.TagNumber(8)
  void clearSignatureHash() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get proof => $_getSZ(8);
  @$pb.TagNumber(9)
  set proof($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasProof() => $_has(8);
  @$pb.TagNumber(9)
  void clearProof() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get ctime => $_getSZ(9);
  @$pb.TagNumber(10)
  set ctime($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasCtime() => $_has(9);
  @$pb.TagNumber(10)
  void clearCtime() => clearField(10);

  @$pb.TagNumber(11)
  $core.bool get public => $_getBF(10);
  @$pb.TagNumber(11)
  set public($core.bool v) { $_setBool(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasPublic() => $_has(10);
  @$pb.TagNumber(11)
  void clearPublic() => clearField(11);

  @$pb.TagNumber(12)
  $6.Timestamp get verifiedAt => $_getN(11);
  @$pb.TagNumber(12)
  set verifiedAt($6.Timestamp v) { setField(12, v); }
  @$pb.TagNumber(12)
  $core.bool hasVerifiedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearVerifiedAt() => clearField(12);
  @$pb.TagNumber(12)
  $6.Timestamp ensureVerifiedAt() => $_ensure(11);

  @$pb.TagNumber(13)
  $core.String get walletId => $_getSZ(12);
  @$pb.TagNumber(13)
  set walletId($core.String v) { $_setString(12, v); }
  @$pb.TagNumber(13)
  $core.bool hasWalletId() => $_has(12);
  @$pb.TagNumber(13)
  void clearWalletId() => clearField(13);

  @$pb.TagNumber(14)
  $core.String get txtRecord => $_getSZ(13);
  @$pb.TagNumber(14)
  set txtRecord($core.String v) { $_setString(13, v); }
  @$pb.TagNumber(14)
  $core.bool hasTxtRecord() => $_has(13);
  @$pb.TagNumber(14)
  void clearTxtRecord() => clearField(14);
}

class IdentityVerificationInstructions extends $pb.GeneratedMessage {
  factory IdentityVerificationInstructions({
    $core.String? identityId,
    $core.String? code,
    $core.String? instructions,
  }) {
    final $result = create();
    if (identityId != null) {
      $result.identityId = identityId;
    }
    if (code != null) {
      $result.code = code;
    }
    if (instructions != null) {
      $result.instructions = instructions;
    }
    return $result;
  }
  IdentityVerificationInstructions._() : super();
  factory IdentityVerificationInstructions.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IdentityVerificationInstructions.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'IdentityVerificationInstructions', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'identityId')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'instructions')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IdentityVerificationInstructions clone() => IdentityVerificationInstructions()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IdentityVerificationInstructions copyWith(void Function(IdentityVerificationInstructions) updates) => super.copyWith((message) => updates(message as IdentityVerificationInstructions)) as IdentityVerificationInstructions;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static IdentityVerificationInstructions create() => IdentityVerificationInstructions._();
  IdentityVerificationInstructions createEmptyInstance() => create();
  static $pb.PbList<IdentityVerificationInstructions> createRepeated() => $pb.PbList<IdentityVerificationInstructions>();
  @$core.pragma('dart2js:noInline')
  static IdentityVerificationInstructions getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IdentityVerificationInstructions>(create);
  static IdentityVerificationInstructions? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get identityId => $_getSZ(0);
  @$pb.TagNumber(1)
  set identityId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasIdentityId() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentityId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get instructions => $_getSZ(2);
  @$pb.TagNumber(3)
  set instructions($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasInstructions() => $_has(2);
  @$pb.TagNumber(3)
  void clearInstructions() => clearField(3);
}

class DeleteIdentityRequest extends $pb.GeneratedMessage {
  factory DeleteIdentityRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  DeleteIdentityRequest._() : super();
  factory DeleteIdentityRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DeleteIdentityRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'DeleteIdentityRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DeleteIdentityRequest clone() => DeleteIdentityRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DeleteIdentityRequest copyWith(void Function(DeleteIdentityRequest) updates) => super.copyWith((message) => updates(message as DeleteIdentityRequest)) as DeleteIdentityRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteIdentityRequest create() => DeleteIdentityRequest._();
  DeleteIdentityRequest createEmptyInstance() => create();
  static $pb.PbList<DeleteIdentityRequest> createRepeated() => $pb.PbList<DeleteIdentityRequest>();
  @$core.pragma('dart2js:noInline')
  static DeleteIdentityRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DeleteIdentityRequest>(create);
  static DeleteIdentityRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class SetIdentityPublicRequest extends $pb.GeneratedMessage {
  factory SetIdentityPublicRequest({
    $core.String? id,
    $core.bool? public,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (public != null) {
      $result.public = public;
    }
    return $result;
  }
  SetIdentityPublicRequest._() : super();
  factory SetIdentityPublicRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetIdentityPublicRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SetIdentityPublicRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOB(2, _omitFieldNames ? '' : 'public')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetIdentityPublicRequest clone() => SetIdentityPublicRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetIdentityPublicRequest copyWith(void Function(SetIdentityPublicRequest) updates) => super.copyWith((message) => updates(message as SetIdentityPublicRequest)) as SetIdentityPublicRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SetIdentityPublicRequest create() => SetIdentityPublicRequest._();
  SetIdentityPublicRequest createEmptyInstance() => create();
  static $pb.PbList<SetIdentityPublicRequest> createRepeated() => $pb.PbList<SetIdentityPublicRequest>();
  @$core.pragma('dart2js:noInline')
  static SetIdentityPublicRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SetIdentityPublicRequest>(create);
  static SetIdentityPublicRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.bool get public => $_getBF(1);
  @$pb.TagNumber(2)
  set public($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPublic() => $_has(1);
  @$pb.TagNumber(2)
  void clearPublic() => clearField(2);
}

class ListPublicIdentitiesRequest extends $pb.GeneratedMessage {
  factory ListPublicIdentitiesRequest({
    $core.String? walletId,
  }) {
    final $result = create();
    if (walletId != null) {
      $result.walletId = walletId;
    }
    return $result;
  }
  ListPublicIdentitiesRequest._() : super();
  factory ListPublicIdentitiesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListPublicIdentitiesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ListPublicIdentitiesRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListPublicIdentitiesRequest clone() => ListPublicIdentitiesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListPublicIdentitiesRequest copyWith(void Function(ListPublicIdentitiesRequest) updates) => super.copyWith((message) => updates(message as ListPublicIdentitiesRequest)) as ListPublicIdentitiesRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListPublicIdentitiesRequest create() => ListPublicIdentitiesRequest._();
  ListPublicIdentitiesRequest createEmptyInstance() => create();
  static $pb.PbList<ListPublicIdentitiesRequest> createRepeated() => $pb.PbList<ListPublicIdentitiesRequest>();
  @$core.pragma('dart2js:noInline')
  static ListPublicIdentitiesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListPublicIdentitiesRequest>(create);
  static ListPublicIdentitiesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletId => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletId() => clearField(1);
}

class KYCStatusResponse extends $pb.GeneratedMessage {
  factory KYCStatusResponse({
    $core.int? kycStatus,
  }) {
    final $result = create();
    if (kycStatus != null) {
      $result.kycStatus = kycStatus;
    }
    return $result;
  }
  KYCStatusResponse._() : super();
  factory KYCStatusResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory KYCStatusResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'KYCStatusResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'kycStatus', $pb.PbFieldType.O3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  KYCStatusResponse clone() => KYCStatusResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  KYCStatusResponse copyWith(void Function(KYCStatusResponse) updates) => super.copyWith((message) => updates(message as KYCStatusResponse)) as KYCStatusResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static KYCStatusResponse create() => KYCStatusResponse._();
  KYCStatusResponse createEmptyInstance() => create();
  static $pb.PbList<KYCStatusResponse> createRepeated() => $pb.PbList<KYCStatusResponse>();
  @$core.pragma('dart2js:noInline')
  static KYCStatusResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<KYCStatusResponse>(create);
  static KYCStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get kycStatus => $_getIZ(0);
  @$pb.TagNumber(1)
  set kycStatus($core.int v) { $_setSignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKycStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearKycStatus() => clearField(1);
}

class KYCPersonaInquiryRequest extends $pb.GeneratedMessage {
  factory KYCPersonaInquiryRequest({
    $core.String? idempotencyKey,
  }) {
    final $result = create();
    if (idempotencyKey != null) {
      $result.idempotencyKey = idempotencyKey;
    }
    return $result;
  }
  KYCPersonaInquiryRequest._() : super();
  factory KYCPersonaInquiryRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory KYCPersonaInquiryRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'KYCPersonaInquiryRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'idempotencyKey')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  KYCPersonaInquiryRequest clone() => KYCPersonaInquiryRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  KYCPersonaInquiryRequest copyWith(void Function(KYCPersonaInquiryRequest) updates) => super.copyWith((message) => updates(message as KYCPersonaInquiryRequest)) as KYCPersonaInquiryRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static KYCPersonaInquiryRequest create() => KYCPersonaInquiryRequest._();
  KYCPersonaInquiryRequest createEmptyInstance() => create();
  static $pb.PbList<KYCPersonaInquiryRequest> createRepeated() => $pb.PbList<KYCPersonaInquiryRequest>();
  @$core.pragma('dart2js:noInline')
  static KYCPersonaInquiryRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<KYCPersonaInquiryRequest>(create);
  static KYCPersonaInquiryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get idempotencyKey => $_getSZ(0);
  @$pb.TagNumber(1)
  set idempotencyKey($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasIdempotencyKey() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdempotencyKey() => clearField(1);
}

class KYCPersonaInquiryResponse extends $pb.GeneratedMessage {
  factory KYCPersonaInquiryResponse({
    $core.String? id,
  @$core.Deprecated('This field is deprecated.')
    $core.String? sessionToken,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (sessionToken != null) {
      // ignore: deprecated_member_use_from_same_package
      $result.sessionToken = sessionToken;
    }
    return $result;
  }
  KYCPersonaInquiryResponse._() : super();
  factory KYCPersonaInquiryResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory KYCPersonaInquiryResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'KYCPersonaInquiryResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'sessionToken')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  KYCPersonaInquiryResponse clone() => KYCPersonaInquiryResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  KYCPersonaInquiryResponse copyWith(void Function(KYCPersonaInquiryResponse) updates) => super.copyWith((message) => updates(message as KYCPersonaInquiryResponse)) as KYCPersonaInquiryResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static KYCPersonaInquiryResponse create() => KYCPersonaInquiryResponse._();
  KYCPersonaInquiryResponse createEmptyInstance() => create();
  static $pb.PbList<KYCPersonaInquiryResponse> createRepeated() => $pb.PbList<KYCPersonaInquiryResponse>();
  @$core.pragma('dart2js:noInline')
  static KYCPersonaInquiryResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<KYCPersonaInquiryResponse>(create);
  static KYCPersonaInquiryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  $core.String get sessionToken => $_getSZ(1);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  set sessionToken($core.String v) { $_setString(1, v); }
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  $core.bool hasSessionToken() => $_has(1);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  void clearSessionToken() => clearField(2);
}

class CreateTwitterAuthURLResponse extends $pb.GeneratedMessage {
  factory CreateTwitterAuthURLResponse({
    $core.String? url,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    return $result;
  }
  CreateTwitterAuthURLResponse._() : super();
  factory CreateTwitterAuthURLResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateTwitterAuthURLResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateTwitterAuthURLResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateTwitterAuthURLResponse clone() => CreateTwitterAuthURLResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateTwitterAuthURLResponse copyWith(void Function(CreateTwitterAuthURLResponse) updates) => super.copyWith((message) => updates(message as CreateTwitterAuthURLResponse)) as CreateTwitterAuthURLResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateTwitterAuthURLResponse create() => CreateTwitterAuthURLResponse._();
  CreateTwitterAuthURLResponse createEmptyInstance() => create();
  static $pb.PbList<CreateTwitterAuthURLResponse> createRepeated() => $pb.PbList<CreateTwitterAuthURLResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateTwitterAuthURLResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateTwitterAuthURLResponse>(create);
  static CreateTwitterAuthURLResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class TwitterCallbackRequest extends $pb.GeneratedMessage {
  factory TwitterCallbackRequest({
    $core.String? state,
    $core.String? code,
  }) {
    final $result = create();
    if (state != null) {
      $result.state = state;
    }
    if (code != null) {
      $result.code = code;
    }
    return $result;
  }
  TwitterCallbackRequest._() : super();
  factory TwitterCallbackRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TwitterCallbackRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'TwitterCallbackRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'state')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TwitterCallbackRequest clone() => TwitterCallbackRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TwitterCallbackRequest copyWith(void Function(TwitterCallbackRequest) updates) => super.copyWith((message) => updates(message as TwitterCallbackRequest)) as TwitterCallbackRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TwitterCallbackRequest create() => TwitterCallbackRequest._();
  TwitterCallbackRequest createEmptyInstance() => create();
  static $pb.PbList<TwitterCallbackRequest> createRepeated() => $pb.PbList<TwitterCallbackRequest>();
  @$core.pragma('dart2js:noInline')
  static TwitterCallbackRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TwitterCallbackRequest>(create);
  static TwitterCallbackRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get state => $_getSZ(0);
  @$pb.TagNumber(1)
  set state($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => clearField(2);
}

class TwitterCallbackResponse extends $pb.GeneratedMessage {
  factory TwitterCallbackResponse({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  TwitterCallbackResponse._() : super();
  factory TwitterCallbackResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory TwitterCallbackResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'TwitterCallbackResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  TwitterCallbackResponse clone() => TwitterCallbackResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  TwitterCallbackResponse copyWith(void Function(TwitterCallbackResponse) updates) => super.copyWith((message) => updates(message as TwitterCallbackResponse)) as TwitterCallbackResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TwitterCallbackResponse create() => TwitterCallbackResponse._();
  TwitterCallbackResponse createEmptyInstance() => create();
  static $pb.PbList<TwitterCallbackResponse> createRepeated() => $pb.PbList<TwitterCallbackResponse>();
  @$core.pragma('dart2js:noInline')
  static TwitterCallbackResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TwitterCallbackResponse>(create);
  static TwitterCallbackResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class DiscordCallbackRequest extends $pb.GeneratedMessage {
  factory DiscordCallbackRequest({
    $core.String? state,
    $core.String? code,
  }) {
    final $result = create();
    if (state != null) {
      $result.state = state;
    }
    if (code != null) {
      $result.code = code;
    }
    return $result;
  }
  DiscordCallbackRequest._() : super();
  factory DiscordCallbackRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DiscordCallbackRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'DiscordCallbackRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'state')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DiscordCallbackRequest clone() => DiscordCallbackRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DiscordCallbackRequest copyWith(void Function(DiscordCallbackRequest) updates) => super.copyWith((message) => updates(message as DiscordCallbackRequest)) as DiscordCallbackRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DiscordCallbackRequest create() => DiscordCallbackRequest._();
  DiscordCallbackRequest createEmptyInstance() => create();
  static $pb.PbList<DiscordCallbackRequest> createRepeated() => $pb.PbList<DiscordCallbackRequest>();
  @$core.pragma('dart2js:noInline')
  static DiscordCallbackRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DiscordCallbackRequest>(create);
  static DiscordCallbackRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get state => $_getSZ(0);
  @$pb.TagNumber(1)
  set state($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => clearField(2);
}

class DiscordCallbackResponse extends $pb.GeneratedMessage {
  factory DiscordCallbackResponse({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  DiscordCallbackResponse._() : super();
  factory DiscordCallbackResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DiscordCallbackResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'DiscordCallbackResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DiscordCallbackResponse clone() => DiscordCallbackResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DiscordCallbackResponse copyWith(void Function(DiscordCallbackResponse) updates) => super.copyWith((message) => updates(message as DiscordCallbackResponse)) as DiscordCallbackResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DiscordCallbackResponse create() => DiscordCallbackResponse._();
  DiscordCallbackResponse createEmptyInstance() => create();
  static $pb.PbList<DiscordCallbackResponse> createRepeated() => $pb.PbList<DiscordCallbackResponse>();
  @$core.pragma('dart2js:noInline')
  static DiscordCallbackResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DiscordCallbackResponse>(create);
  static DiscordCallbackResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class CreateDiscordAuthURLResponse extends $pb.GeneratedMessage {
  factory CreateDiscordAuthURLResponse({
    $core.String? url,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    return $result;
  }
  CreateDiscordAuthURLResponse._() : super();
  factory CreateDiscordAuthURLResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateDiscordAuthURLResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateDiscordAuthURLResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateDiscordAuthURLResponse clone() => CreateDiscordAuthURLResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateDiscordAuthURLResponse copyWith(void Function(CreateDiscordAuthURLResponse) updates) => super.copyWith((message) => updates(message as CreateDiscordAuthURLResponse)) as CreateDiscordAuthURLResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateDiscordAuthURLResponse create() => CreateDiscordAuthURLResponse._();
  CreateDiscordAuthURLResponse createEmptyInstance() => create();
  static $pb.PbList<CreateDiscordAuthURLResponse> createRepeated() => $pb.PbList<CreateDiscordAuthURLResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateDiscordAuthURLResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateDiscordAuthURLResponse>(create);
  static CreateDiscordAuthURLResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class GetIdentityRequest extends $pb.GeneratedMessage {
  factory GetIdentityRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  GetIdentityRequest._() : super();
  factory GetIdentityRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetIdentityRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetIdentityRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetIdentityRequest clone() => GetIdentityRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetIdentityRequest copyWith(void Function(GetIdentityRequest) updates) => super.copyWith((message) => updates(message as GetIdentityRequest)) as GetIdentityRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetIdentityRequest create() => GetIdentityRequest._();
  GetIdentityRequest createEmptyInstance() => create();
  static $pb.PbList<GetIdentityRequest> createRepeated() => $pb.PbList<GetIdentityRequest>();
  @$core.pragma('dart2js:noInline')
  static GetIdentityRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetIdentityRequest>(create);
  static GetIdentityRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class GetIdentityResponse extends $pb.GeneratedMessage {
  factory GetIdentityResponse({
    Identity? identity,
  }) {
    final $result = create();
    if (identity != null) {
      $result.identity = identity;
    }
    return $result;
  }
  GetIdentityResponse._() : super();
  factory GetIdentityResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetIdentityResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetIdentityResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<Identity>(1, _omitFieldNames ? '' : 'identity', subBuilder: Identity.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetIdentityResponse clone() => GetIdentityResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetIdentityResponse copyWith(void Function(GetIdentityResponse) updates) => super.copyWith((message) => updates(message as GetIdentityResponse)) as GetIdentityResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetIdentityResponse create() => GetIdentityResponse._();
  GetIdentityResponse createEmptyInstance() => create();
  static $pb.PbList<GetIdentityResponse> createRepeated() => $pb.PbList<GetIdentityResponse>();
  @$core.pragma('dart2js:noInline')
  static GetIdentityResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetIdentityResponse>(create);
  static GetIdentityResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Identity get identity => $_getN(0);
  @$pb.TagNumber(1)
  set identity(Identity v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasIdentity() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentity() => clearField(1);
  @$pb.TagNumber(1)
  Identity ensureIdentity() => $_ensure(0);
}

class GetIdentityBySignatureHashRequest extends $pb.GeneratedMessage {
  factory GetIdentityBySignatureHashRequest({
    $core.String? signatureHash,
  }) {
    final $result = create();
    if (signatureHash != null) {
      $result.signatureHash = signatureHash;
    }
    return $result;
  }
  GetIdentityBySignatureHashRequest._() : super();
  factory GetIdentityBySignatureHashRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetIdentityBySignatureHashRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetIdentityBySignatureHashRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'signatureHash')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetIdentityBySignatureHashRequest clone() => GetIdentityBySignatureHashRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetIdentityBySignatureHashRequest copyWith(void Function(GetIdentityBySignatureHashRequest) updates) => super.copyWith((message) => updates(message as GetIdentityBySignatureHashRequest)) as GetIdentityBySignatureHashRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetIdentityBySignatureHashRequest create() => GetIdentityBySignatureHashRequest._();
  GetIdentityBySignatureHashRequest createEmptyInstance() => create();
  static $pb.PbList<GetIdentityBySignatureHashRequest> createRepeated() => $pb.PbList<GetIdentityBySignatureHashRequest>();
  @$core.pragma('dart2js:noInline')
  static GetIdentityBySignatureHashRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetIdentityBySignatureHashRequest>(create);
  static GetIdentityBySignatureHashRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get signatureHash => $_getSZ(0);
  @$pb.TagNumber(1)
  set signatureHash($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSignatureHash() => $_has(0);
  @$pb.TagNumber(1)
  void clearSignatureHash() => clearField(1);
}

class GetPaymentAddressRequest extends $pb.GeneratedMessage {
  factory GetPaymentAddressRequest({
    $core.String? address,
  }) {
    final $result = create();
    if (address != null) {
      $result.address = address;
    }
    return $result;
  }
  GetPaymentAddressRequest._() : super();
  factory GetPaymentAddressRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPaymentAddressRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetPaymentAddressRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'address')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPaymentAddressRequest clone() => GetPaymentAddressRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPaymentAddressRequest copyWith(void Function(GetPaymentAddressRequest) updates) => super.copyWith((message) => updates(message as GetPaymentAddressRequest)) as GetPaymentAddressRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPaymentAddressRequest create() => GetPaymentAddressRequest._();
  GetPaymentAddressRequest createEmptyInstance() => create();
  static $pb.PbList<GetPaymentAddressRequest> createRepeated() => $pb.PbList<GetPaymentAddressRequest>();
  @$core.pragma('dart2js:noInline')
  static GetPaymentAddressRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetPaymentAddressRequest>(create);
  static GetPaymentAddressRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get address => $_getSZ(0);
  @$pb.TagNumber(1)
  set address($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => clearField(1);
}

class GetPaymentAddressResponse extends $pb.GeneratedMessage {
  factory GetPaymentAddressResponse({
    $core.String? walletUrl,
    $core.String? type,
    $core.String? handle,
    $core.bool? canSendToAddress,
  }) {
    final $result = create();
    if (walletUrl != null) {
      $result.walletUrl = walletUrl;
    }
    if (type != null) {
      $result.type = type;
    }
    if (handle != null) {
      $result.handle = handle;
    }
    if (canSendToAddress != null) {
      $result.canSendToAddress = canSendToAddress;
    }
    return $result;
  }
  GetPaymentAddressResponse._() : super();
  factory GetPaymentAddressResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPaymentAddressResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetPaymentAddressResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'walletUrl')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOS(3, _omitFieldNames ? '' : 'handle')
    ..aOB(4, _omitFieldNames ? '' : 'canSendToAddress', protoName: 'canSendToAddress')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPaymentAddressResponse clone() => GetPaymentAddressResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPaymentAddressResponse copyWith(void Function(GetPaymentAddressResponse) updates) => super.copyWith((message) => updates(message as GetPaymentAddressResponse)) as GetPaymentAddressResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPaymentAddressResponse create() => GetPaymentAddressResponse._();
  GetPaymentAddressResponse createEmptyInstance() => create();
  static $pb.PbList<GetPaymentAddressResponse> createRepeated() => $pb.PbList<GetPaymentAddressResponse>();
  @$core.pragma('dart2js:noInline')
  static GetPaymentAddressResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetPaymentAddressResponse>(create);
  static GetPaymentAddressResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get walletUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set walletUrl($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWalletUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearWalletUrl() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get handle => $_getSZ(2);
  @$pb.TagNumber(3)
  set handle($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasHandle() => $_has(2);
  @$pb.TagNumber(3)
  void clearHandle() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get canSendToAddress => $_getBF(3);
  @$pb.TagNumber(4)
  set canSendToAddress($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCanSendToAddress() => $_has(3);
  @$pb.TagNumber(4)
  void clearCanSendToAddress() => clearField(4);
}

class CreateDomainIdentityRequest extends $pb.GeneratedMessage {
  factory CreateDomainIdentityRequest({
    $core.String? url,
  }) {
    final $result = create();
    if (url != null) {
      $result.url = url;
    }
    return $result;
  }
  CreateDomainIdentityRequest._() : super();
  factory CreateDomainIdentityRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateDomainIdentityRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateDomainIdentityRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'url')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateDomainIdentityRequest clone() => CreateDomainIdentityRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateDomainIdentityRequest copyWith(void Function(CreateDomainIdentityRequest) updates) => super.copyWith((message) => updates(message as CreateDomainIdentityRequest)) as CreateDomainIdentityRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateDomainIdentityRequest create() => CreateDomainIdentityRequest._();
  CreateDomainIdentityRequest createEmptyInstance() => create();
  static $pb.PbList<CreateDomainIdentityRequest> createRepeated() => $pb.PbList<CreateDomainIdentityRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateDomainIdentityRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateDomainIdentityRequest>(create);
  static CreateDomainIdentityRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class CreateDomainIdentityResponse extends $pb.GeneratedMessage {
  factory CreateDomainIdentityResponse({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  CreateDomainIdentityResponse._() : super();
  factory CreateDomainIdentityResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateDomainIdentityResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateDomainIdentityResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateDomainIdentityResponse clone() => CreateDomainIdentityResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateDomainIdentityResponse copyWith(void Function(CreateDomainIdentityResponse) updates) => super.copyWith((message) => updates(message as CreateDomainIdentityResponse)) as CreateDomainIdentityResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateDomainIdentityResponse create() => CreateDomainIdentityResponse._();
  CreateDomainIdentityResponse createEmptyInstance() => create();
  static $pb.PbList<CreateDomainIdentityResponse> createRepeated() => $pb.PbList<CreateDomainIdentityResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateDomainIdentityResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateDomainIdentityResponse>(create);
  static CreateDomainIdentityResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class VerifyIdentityRequest extends $pb.GeneratedMessage {
  factory VerifyIdentityRequest({
    $core.String? id,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    return $result;
  }
  VerifyIdentityRequest._() : super();
  factory VerifyIdentityRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VerifyIdentityRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'VerifyIdentityRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VerifyIdentityRequest clone() => VerifyIdentityRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VerifyIdentityRequest copyWith(void Function(VerifyIdentityRequest) updates) => super.copyWith((message) => updates(message as VerifyIdentityRequest)) as VerifyIdentityRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VerifyIdentityRequest create() => VerifyIdentityRequest._();
  VerifyIdentityRequest createEmptyInstance() => create();
  static $pb.PbList<VerifyIdentityRequest> createRepeated() => $pb.PbList<VerifyIdentityRequest>();
  @$core.pragma('dart2js:noInline')
  static VerifyIdentityRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VerifyIdentityRequest>(create);
  static VerifyIdentityRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class SubmitFormRequest extends $pb.GeneratedMessage {
  factory SubmitFormRequest({
    $core.String? formId,
    $core.String? data,
  }) {
    final $result = create();
    if (formId != null) {
      $result.formId = formId;
    }
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  SubmitFormRequest._() : super();
  factory SubmitFormRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SubmitFormRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SubmitFormRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'formId')
    ..aOS(2, _omitFieldNames ? '' : 'data')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SubmitFormRequest clone() => SubmitFormRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SubmitFormRequest copyWith(void Function(SubmitFormRequest) updates) => super.copyWith((message) => updates(message as SubmitFormRequest)) as SubmitFormRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SubmitFormRequest create() => SubmitFormRequest._();
  SubmitFormRequest createEmptyInstance() => create();
  static $pb.PbList<SubmitFormRequest> createRepeated() => $pb.PbList<SubmitFormRequest>();
  @$core.pragma('dart2js:noInline')
  static SubmitFormRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SubmitFormRequest>(create);
  static SubmitFormRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get formId => $_getSZ(0);
  @$pb.TagNumber(1)
  set formId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFormId() => $_has(0);
  @$pb.TagNumber(1)
  void clearFormId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get data => $_getSZ(1);
  @$pb.TagNumber(2)
  set data($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasData() => $_has(1);
  @$pb.TagNumber(2)
  void clearData() => clearField(2);
}


const _omitFieldNames = $core.bool.fromEnvironment('protobuf.omit_field_names');
const _omitMessageNames = $core.bool.fromEnvironment('protobuf.omit_message_names');
