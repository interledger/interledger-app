///
//  Generated code. Do not modify.
//  source: backend/v1/backend.proto
//
// @dart = 2.12
// ignore_for_file: annotate_overrides,camel_case_types,constant_identifier_names,directives_ordering,library_prefixes,non_constant_identifier_names,prefer_final_fields,return_of_invalid_type,unnecessary_const,unnecessary_import,unnecessary_this,unused_import,unused_shown_name

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../google/protobuf/timestamp.pb.dart' as $6;

class PaginationRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PaginationRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pageSize', $pb.PbFieldType.O3, protoName: 'pageSize')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pageToken', protoName: 'pageToken')
    ..hasRequiredFields = false
  ;

  PaginationRequest._() : super();
  factory PaginationRequest({
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final _result = create();
    if (pageSize != null) {
      _result.pageSize = pageSize;
    }
    if (pageToken != null) {
      _result.pageToken = pageToken;
    }
    return _result;
  }
  factory PaginationRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PaginationRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PaginationRequest clone() => PaginationRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PaginationRequest copyWith(void Function(PaginationRequest) updates) => super.copyWith((message) => updates(message as PaginationRequest)) as PaginationRequest; // ignore: deprecated_member_use
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

class CanSendToPaymentPointerRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CanSendToPaymentPointerRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paymentPointer', protoName: 'paymentPointer')
    ..hasRequiredFields = false
  ;

  CanSendToPaymentPointerRequest._() : super();
  factory CanSendToPaymentPointerRequest({
    $core.String? paymentPointer,
  }) {
    final _result = create();
    if (paymentPointer != null) {
      _result.paymentPointer = paymentPointer;
    }
    return _result;
  }
  factory CanSendToPaymentPointerRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CanSendToPaymentPointerRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CanSendToPaymentPointerRequest clone() => CanSendToPaymentPointerRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CanSendToPaymentPointerRequest copyWith(void Function(CanSendToPaymentPointerRequest) updates) => super.copyWith((message) => updates(message as CanSendToPaymentPointerRequest)) as CanSendToPaymentPointerRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CanSendToPaymentPointerRequest create() => CanSendToPaymentPointerRequest._();
  CanSendToPaymentPointerRequest createEmptyInstance() => create();
  static $pb.PbList<CanSendToPaymentPointerRequest> createRepeated() => $pb.PbList<CanSendToPaymentPointerRequest>();
  @$core.pragma('dart2js:noInline')
  static CanSendToPaymentPointerRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CanSendToPaymentPointerRequest>(create);
  static CanSendToPaymentPointerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get paymentPointer => $_getSZ(0);
  @$pb.TagNumber(1)
  set paymentPointer($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPaymentPointer() => $_has(0);
  @$pb.TagNumber(1)
  void clearPaymentPointer() => clearField(1);
}

class CanSendToPaymentPointerResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CanSendToPaymentPointerResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'canSend', protoName: 'canSend')
    ..hasRequiredFields = false
  ;

  CanSendToPaymentPointerResponse._() : super();
  factory CanSendToPaymentPointerResponse({
    $core.bool? canSend,
  }) {
    final _result = create();
    if (canSend != null) {
      _result.canSend = canSend;
    }
    return _result;
  }
  factory CanSendToPaymentPointerResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CanSendToPaymentPointerResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CanSendToPaymentPointerResponse clone() => CanSendToPaymentPointerResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CanSendToPaymentPointerResponse copyWith(void Function(CanSendToPaymentPointerResponse) updates) => super.copyWith((message) => updates(message as CanSendToPaymentPointerResponse)) as CanSendToPaymentPointerResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CanSendToPaymentPointerResponse create() => CanSendToPaymentPointerResponse._();
  CanSendToPaymentPointerResponse createEmptyInstance() => create();
  static $pb.PbList<CanSendToPaymentPointerResponse> createRepeated() => $pb.PbList<CanSendToPaymentPointerResponse>();
  @$core.pragma('dart2js:noInline')
  static CanSendToPaymentPointerResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CanSendToPaymentPointerResponse>(create);
  static CanSendToPaymentPointerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get canSend => $_getBF(0);
  @$pb.TagNumber(1)
  set canSend($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCanSend() => $_has(0);
  @$pb.TagNumber(1)
  void clearCanSend() => clearField(1);
}

class Transaction extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Transaction', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'type')
    ..aOM<Amount>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', subBuilder: Amount.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'source')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'destination')
    ..aOM<$6.Timestamp>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'state')
    ..pc<Transfer>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'transfers', $pb.PbFieldType.PM, subBuilder: Transfer.create)
    ..aOS(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'foreignId', protoName: 'foreignId')
    ..hasRequiredFields = false
  ;

  Transaction._() : super();
  factory Transaction({
    $core.String? id,
    $core.String? type,
    Amount? amount,
    $core.String? source,
    $core.String? destination,
    $6.Timestamp? timestamp,
    $core.String? state,
    $core.Iterable<Transfer>? transfers,
    $core.String? foreignId,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (type != null) {
      _result.type = type;
    }
    if (amount != null) {
      _result.amount = amount;
    }
    if (source != null) {
      _result.source = source;
    }
    if (destination != null) {
      _result.destination = destination;
    }
    if (timestamp != null) {
      _result.timestamp = timestamp;
    }
    if (state != null) {
      _result.state = state;
    }
    if (transfers != null) {
      _result.transfers.addAll(transfers);
    }
    if (foreignId != null) {
      _result.foreignId = foreignId;
    }
    return _result;
  }
  factory Transaction.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transaction.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Transaction clone() => Transaction()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Transaction copyWith(void Function(Transaction) updates) => super.copyWith((message) => updates(message as Transaction)) as Transaction; // ignore: deprecated_member_use
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

  @$pb.TagNumber(8)
  $core.List<Transfer> get transfers => $_getList(7);

  @$pb.TagNumber(9)
  $core.String get foreignId => $_getSZ(8);
  @$pb.TagNumber(9)
  set foreignId($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasForeignId() => $_has(8);
  @$pb.TagNumber(9)
  void clearForeignId() => clearField(9);
}

class ListTransactionsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ListTransactionsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Transaction>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'transactions', $pb.PbFieldType.PM, subBuilder: Transaction.create)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nextPageToken', protoName: 'nextPageToken')
    ..hasRequiredFields = false
  ;

  ListTransactionsResponse._() : super();
  factory ListTransactionsResponse({
    $core.Iterable<Transaction>? transactions,
    $core.String? nextPageToken,
  }) {
    final _result = create();
    if (transactions != null) {
      _result.transactions.addAll(transactions);
    }
    if (nextPageToken != null) {
      _result.nextPageToken = nextPageToken;
    }
    return _result;
  }
  factory ListTransactionsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListTransactionsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListTransactionsResponse clone() => ListTransactionsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListTransactionsResponse copyWith(void Function(ListTransactionsResponse) updates) => super.copyWith((message) => updates(message as ListTransactionsResponse)) as ListTransactionsResponse; // ignore: deprecated_member_use
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

class Amount extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Amount', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'asset')
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'assetScale', $pb.PbFieldType.O3, protoName: 'assetScale')
    ..hasRequiredFields = false
  ;

  Amount._() : super();
  factory Amount({
    $fixnum.Int64? amount,
    $core.String? asset,
    $core.int? assetScale,
  }) {
    final _result = create();
    if (amount != null) {
      _result.amount = amount;
    }
    if (asset != null) {
      _result.asset = asset;
    }
    if (assetScale != null) {
      _result.assetScale = assetScale;
    }
    return _result;
  }
  factory Amount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Amount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Amount clone() => Amount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Amount copyWith(void Function(Amount) updates) => super.copyWith((message) => updates(message as Amount)) as Amount; // ignore: deprecated_member_use
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
}

class LookupOutgoingPaymentRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LookupOutgoingPaymentRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  LookupOutgoingPaymentRequest._() : super();
  factory LookupOutgoingPaymentRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory LookupOutgoingPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LookupOutgoingPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LookupOutgoingPaymentRequest clone() => LookupOutgoingPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LookupOutgoingPaymentRequest copyWith(void Function(LookupOutgoingPaymentRequest) updates) => super.copyWith((message) => updates(message as LookupOutgoingPaymentRequest)) as LookupOutgoingPaymentRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LookupOutgoingPaymentRequest create() => LookupOutgoingPaymentRequest._();
  LookupOutgoingPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<LookupOutgoingPaymentRequest> createRepeated() => $pb.PbList<LookupOutgoingPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static LookupOutgoingPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LookupOutgoingPaymentRequest>(create);
  static LookupOutgoingPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class OutgoingPayment extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'OutgoingPayment', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paymentPointer', protoName: 'paymentPointer')
    ..aOB(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'failed')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'receiver')
    ..aOM<Amount>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sendAmount', protoName: 'sendAmount', subBuilder: Amount.create)
    ..aOM<Amount>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'receiveAmount', protoName: 'receiveAmount', subBuilder: Amount.create)
    ..aOM<Amount>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sentAmount', protoName: 'sentAmount', subBuilder: Amount.create)
    ..aOS(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description')
    ..aOM<$6.Timestamp>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'createdAt', protoName: 'createdAt', subBuilder: $6.Timestamp.create)
    ..aOM<$6.Timestamp>(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'updatedAt', protoName: 'updatedAt', subBuilder: $6.Timestamp.create)
    ..aOS(11, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'toPaymentPointer', protoName: 'toPaymentPointer')
    ..hasRequiredFields = false
  ;

  OutgoingPayment._() : super();
  factory OutgoingPayment({
    $core.String? id,
    $core.String? paymentPointer,
    $core.bool? failed,
    $core.String? receiver,
    Amount? sendAmount,
    Amount? receiveAmount,
    Amount? sentAmount,
    $core.String? description,
    $6.Timestamp? createdAt,
    $6.Timestamp? updatedAt,
    $core.String? toPaymentPointer,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (paymentPointer != null) {
      _result.paymentPointer = paymentPointer;
    }
    if (failed != null) {
      _result.failed = failed;
    }
    if (receiver != null) {
      _result.receiver = receiver;
    }
    if (sendAmount != null) {
      _result.sendAmount = sendAmount;
    }
    if (receiveAmount != null) {
      _result.receiveAmount = receiveAmount;
    }
    if (sentAmount != null) {
      _result.sentAmount = sentAmount;
    }
    if (description != null) {
      _result.description = description;
    }
    if (createdAt != null) {
      _result.createdAt = createdAt;
    }
    if (updatedAt != null) {
      _result.updatedAt = updatedAt;
    }
    if (toPaymentPointer != null) {
      _result.toPaymentPointer = toPaymentPointer;
    }
    return _result;
  }
  factory OutgoingPayment.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory OutgoingPayment.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  OutgoingPayment clone() => OutgoingPayment()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  OutgoingPayment copyWith(void Function(OutgoingPayment) updates) => super.copyWith((message) => updates(message as OutgoingPayment)) as OutgoingPayment; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static OutgoingPayment create() => OutgoingPayment._();
  OutgoingPayment createEmptyInstance() => create();
  static $pb.PbList<OutgoingPayment> createRepeated() => $pb.PbList<OutgoingPayment>();
  @$core.pragma('dart2js:noInline')
  static OutgoingPayment getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<OutgoingPayment>(create);
  static OutgoingPayment? _defaultInstance;

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
  $core.bool get failed => $_getBF(2);
  @$pb.TagNumber(3)
  set failed($core.bool v) { $_setBool(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasFailed() => $_has(2);
  @$pb.TagNumber(3)
  void clearFailed() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get receiver => $_getSZ(3);
  @$pb.TagNumber(4)
  set receiver($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasReceiver() => $_has(3);
  @$pb.TagNumber(4)
  void clearReceiver() => clearField(4);

  @$pb.TagNumber(5)
  Amount get sendAmount => $_getN(4);
  @$pb.TagNumber(5)
  set sendAmount(Amount v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasSendAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearSendAmount() => clearField(5);
  @$pb.TagNumber(5)
  Amount ensureSendAmount() => $_ensure(4);

  @$pb.TagNumber(6)
  Amount get receiveAmount => $_getN(5);
  @$pb.TagNumber(6)
  set receiveAmount(Amount v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasReceiveAmount() => $_has(5);
  @$pb.TagNumber(6)
  void clearReceiveAmount() => clearField(6);
  @$pb.TagNumber(6)
  Amount ensureReceiveAmount() => $_ensure(5);

  @$pb.TagNumber(7)
  Amount get sentAmount => $_getN(6);
  @$pb.TagNumber(7)
  set sentAmount(Amount v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasSentAmount() => $_has(6);
  @$pb.TagNumber(7)
  void clearSentAmount() => clearField(7);
  @$pb.TagNumber(7)
  Amount ensureSentAmount() => $_ensure(6);

  @$pb.TagNumber(8)
  $core.String get description => $_getSZ(7);
  @$pb.TagNumber(8)
  set description($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasDescription() => $_has(7);
  @$pb.TagNumber(8)
  void clearDescription() => clearField(8);

  @$pb.TagNumber(9)
  $6.Timestamp get createdAt => $_getN(8);
  @$pb.TagNumber(9)
  set createdAt($6.Timestamp v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasCreatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearCreatedAt() => clearField(9);
  @$pb.TagNumber(9)
  $6.Timestamp ensureCreatedAt() => $_ensure(8);

  @$pb.TagNumber(10)
  $6.Timestamp get updatedAt => $_getN(9);
  @$pb.TagNumber(10)
  set updatedAt($6.Timestamp v) { setField(10, v); }
  @$pb.TagNumber(10)
  $core.bool hasUpdatedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearUpdatedAt() => clearField(10);
  @$pb.TagNumber(10)
  $6.Timestamp ensureUpdatedAt() => $_ensure(9);

  @$pb.TagNumber(11)
  $core.String get toPaymentPointer => $_getSZ(10);
  @$pb.TagNumber(11)
  set toPaymentPointer($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasToPaymentPointer() => $_has(10);
  @$pb.TagNumber(11)
  void clearToPaymentPointer() => clearField(11);
}

class PreCheckOutgoingPaymentRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PreCheckOutgoingPaymentRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'quoteID', protoName: 'quoteID')
    ..hasRequiredFields = false
  ;

  PreCheckOutgoingPaymentRequest._() : super();
  factory PreCheckOutgoingPaymentRequest({
    $core.String? quoteID,
  }) {
    final _result = create();
    if (quoteID != null) {
      _result.quoteID = quoteID;
    }
    return _result;
  }
  factory PreCheckOutgoingPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PreCheckOutgoingPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PreCheckOutgoingPaymentRequest clone() => PreCheckOutgoingPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PreCheckOutgoingPaymentRequest copyWith(void Function(PreCheckOutgoingPaymentRequest) updates) => super.copyWith((message) => updates(message as PreCheckOutgoingPaymentRequest)) as PreCheckOutgoingPaymentRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PreCheckOutgoingPaymentRequest create() => PreCheckOutgoingPaymentRequest._();
  PreCheckOutgoingPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<PreCheckOutgoingPaymentRequest> createRepeated() => $pb.PbList<PreCheckOutgoingPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static PreCheckOutgoingPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PreCheckOutgoingPaymentRequest>(create);
  static PreCheckOutgoingPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get quoteID => $_getSZ(0);
  @$pb.TagNumber(1)
  set quoteID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasQuoteID() => $_has(0);
  @$pb.TagNumber(1)
  void clearQuoteID() => clearField(1);
}

class PreCheckOutgoingPaymentResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PreCheckOutgoingPaymentResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exceedsLimits', protoName: 'exceedsLimits')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'limitType', protoName: 'limitType')
    ..aOB(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'insufficientBalance', protoName: 'insufficientBalance')
    ..hasRequiredFields = false
  ;

  PreCheckOutgoingPaymentResponse._() : super();
  factory PreCheckOutgoingPaymentResponse({
    $core.bool? exceedsLimits,
    $core.String? limitType,
    $core.bool? insufficientBalance,
  }) {
    final _result = create();
    if (exceedsLimits != null) {
      _result.exceedsLimits = exceedsLimits;
    }
    if (limitType != null) {
      _result.limitType = limitType;
    }
    if (insufficientBalance != null) {
      _result.insufficientBalance = insufficientBalance;
    }
    return _result;
  }
  factory PreCheckOutgoingPaymentResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PreCheckOutgoingPaymentResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PreCheckOutgoingPaymentResponse clone() => PreCheckOutgoingPaymentResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PreCheckOutgoingPaymentResponse copyWith(void Function(PreCheckOutgoingPaymentResponse) updates) => super.copyWith((message) => updates(message as PreCheckOutgoingPaymentResponse)) as PreCheckOutgoingPaymentResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PreCheckOutgoingPaymentResponse create() => PreCheckOutgoingPaymentResponse._();
  PreCheckOutgoingPaymentResponse createEmptyInstance() => create();
  static $pb.PbList<PreCheckOutgoingPaymentResponse> createRepeated() => $pb.PbList<PreCheckOutgoingPaymentResponse>();
  @$core.pragma('dart2js:noInline')
  static PreCheckOutgoingPaymentResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PreCheckOutgoingPaymentResponse>(create);
  static PreCheckOutgoingPaymentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get exceedsLimits => $_getBF(0);
  @$pb.TagNumber(1)
  set exceedsLimits($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasExceedsLimits() => $_has(0);
  @$pb.TagNumber(1)
  void clearExceedsLimits() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get limitType => $_getSZ(1);
  @$pb.TagNumber(2)
  set limitType($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLimitType() => $_has(1);
  @$pb.TagNumber(2)
  void clearLimitType() => clearField(2);

  @$pb.TagNumber(3)
  $core.bool get insufficientBalance => $_getBF(2);
  @$pb.TagNumber(3)
  set insufficientBalance($core.bool v) { $_setBool(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasInsufficientBalance() => $_has(2);
  @$pb.TagNumber(3)
  void clearInsufficientBalance() => clearField(3);
}

class CreateOutgoingPaymentRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateOutgoingPaymentRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'quoteID', protoName: 'quoteID')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'externalRef', protoName: 'externalRef')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ipAddress', protoName: 'ipAddress')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'idempotencyKey', protoName: 'idempotencyKey')
    ..hasRequiredFields = false
  ;

  CreateOutgoingPaymentRequest._() : super();
  factory CreateOutgoingPaymentRequest({
    $core.String? quoteID,
    $core.String? description,
    $core.String? externalRef,
    $core.String? ipAddress,
    $core.String? idempotencyKey,
  }) {
    final _result = create();
    if (quoteID != null) {
      _result.quoteID = quoteID;
    }
    if (description != null) {
      _result.description = description;
    }
    if (externalRef != null) {
      _result.externalRef = externalRef;
    }
    if (ipAddress != null) {
      _result.ipAddress = ipAddress;
    }
    if (idempotencyKey != null) {
      _result.idempotencyKey = idempotencyKey;
    }
    return _result;
  }
  factory CreateOutgoingPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateOutgoingPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateOutgoingPaymentRequest clone() => CreateOutgoingPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateOutgoingPaymentRequest copyWith(void Function(CreateOutgoingPaymentRequest) updates) => super.copyWith((message) => updates(message as CreateOutgoingPaymentRequest)) as CreateOutgoingPaymentRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CreateOutgoingPaymentRequest create() => CreateOutgoingPaymentRequest._();
  CreateOutgoingPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<CreateOutgoingPaymentRequest> createRepeated() => $pb.PbList<CreateOutgoingPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateOutgoingPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateOutgoingPaymentRequest>(create);
  static CreateOutgoingPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get quoteID => $_getSZ(0);
  @$pb.TagNumber(1)
  set quoteID($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasQuoteID() => $_has(0);
  @$pb.TagNumber(1)
  void clearQuoteID() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get externalRef => $_getSZ(2);
  @$pb.TagNumber(3)
  set externalRef($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasExternalRef() => $_has(2);
  @$pb.TagNumber(3)
  void clearExternalRef() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get ipAddress => $_getSZ(3);
  @$pb.TagNumber(4)
  set ipAddress($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasIpAddress() => $_has(3);
  @$pb.TagNumber(4)
  void clearIpAddress() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get idempotencyKey => $_getSZ(4);
  @$pb.TagNumber(5)
  set idempotencyKey($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasIdempotencyKey() => $_has(4);
  @$pb.TagNumber(5)
  void clearIdempotencyKey() => clearField(5);
}

class LookupIncomingPaymentRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LookupIncomingPaymentRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  LookupIncomingPaymentRequest._() : super();
  factory LookupIncomingPaymentRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory LookupIncomingPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LookupIncomingPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LookupIncomingPaymentRequest clone() => LookupIncomingPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LookupIncomingPaymentRequest copyWith(void Function(LookupIncomingPaymentRequest) updates) => super.copyWith((message) => updates(message as LookupIncomingPaymentRequest)) as LookupIncomingPaymentRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LookupIncomingPaymentRequest create() => LookupIncomingPaymentRequest._();
  LookupIncomingPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<LookupIncomingPaymentRequest> createRepeated() => $pb.PbList<LookupIncomingPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static LookupIncomingPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LookupIncomingPaymentRequest>(create);
  static LookupIncomingPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class CreateIncomingPaymentRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateIncomingPaymentRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paymentPointer', protoName: 'paymentPointer')
    ..aOM<Amount>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', subBuilder: Amount.create)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'reference')
    ..aOM<$6.Timestamp>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'expiresAt', protoName: 'expiresAt', subBuilder: $6.Timestamp.create)
    ..hasRequiredFields = false
  ;

  CreateIncomingPaymentRequest._() : super();
  factory CreateIncomingPaymentRequest({
    $core.String? paymentPointer,
    Amount? amount,
    $core.String? reference,
    $6.Timestamp? expiresAt,
  }) {
    final _result = create();
    if (paymentPointer != null) {
      _result.paymentPointer = paymentPointer;
    }
    if (amount != null) {
      _result.amount = amount;
    }
    if (reference != null) {
      _result.reference = reference;
    }
    if (expiresAt != null) {
      _result.expiresAt = expiresAt;
    }
    return _result;
  }
  factory CreateIncomingPaymentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateIncomingPaymentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateIncomingPaymentRequest clone() => CreateIncomingPaymentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateIncomingPaymentRequest copyWith(void Function(CreateIncomingPaymentRequest) updates) => super.copyWith((message) => updates(message as CreateIncomingPaymentRequest)) as CreateIncomingPaymentRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CreateIncomingPaymentRequest create() => CreateIncomingPaymentRequest._();
  CreateIncomingPaymentRequest createEmptyInstance() => create();
  static $pb.PbList<CreateIncomingPaymentRequest> createRepeated() => $pb.PbList<CreateIncomingPaymentRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateIncomingPaymentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateIncomingPaymentRequest>(create);
  static CreateIncomingPaymentRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get paymentPointer => $_getSZ(0);
  @$pb.TagNumber(1)
  set paymentPointer($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPaymentPointer() => $_has(0);
  @$pb.TagNumber(1)
  void clearPaymentPointer() => clearField(1);

  @$pb.TagNumber(2)
  Amount get amount => $_getN(1);
  @$pb.TagNumber(2)
  set amount(Amount v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasAmount() => $_has(1);
  @$pb.TagNumber(2)
  void clearAmount() => clearField(2);
  @$pb.TagNumber(2)
  Amount ensureAmount() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.String get reference => $_getSZ(2);
  @$pb.TagNumber(3)
  set reference($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasReference() => $_has(2);
  @$pb.TagNumber(3)
  void clearReference() => clearField(3);

  @$pb.TagNumber(4)
  $6.Timestamp get expiresAt => $_getN(3);
  @$pb.TagNumber(4)
  set expiresAt($6.Timestamp v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasExpiresAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearExpiresAt() => clearField(4);
  @$pb.TagNumber(4)
  $6.Timestamp ensureExpiresAt() => $_ensure(3);
}

class IncomingPayment extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'IncomingPayment', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paymentPointer', protoName: 'paymentPointer')
    ..aOM<Amount>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'incomingAmount', protoName: 'incomingAmount', subBuilder: Amount.create)
    ..aOM<Amount>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'receivedAmount', protoName: 'receivedAmount', subBuilder: Amount.create)
    ..aOB(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'completed')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'externalRef', protoName: 'externalRef')
    ..aOM<$6.Timestamp>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'expiresAt', protoName: 'expiresAt', subBuilder: $6.Timestamp.create)
    ..aOM<$6.Timestamp>(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'createdAt', protoName: 'createdAt', subBuilder: $6.Timestamp.create)
    ..aOM<$6.Timestamp>(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'updatedAt', protoName: 'updatedAt', subBuilder: $6.Timestamp.create)
    ..aOS(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'fromPaymentPointer', protoName: 'fromPaymentPointer')
    ..hasRequiredFields = false
  ;

  IncomingPayment._() : super();
  factory IncomingPayment({
    $core.String? id,
    $core.String? paymentPointer,
    Amount? incomingAmount,
    Amount? receivedAmount,
    $core.bool? completed,
    $core.String? externalRef,
    $6.Timestamp? expiresAt,
    $6.Timestamp? createdAt,
    $6.Timestamp? updatedAt,
    $core.String? fromPaymentPointer,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (paymentPointer != null) {
      _result.paymentPointer = paymentPointer;
    }
    if (incomingAmount != null) {
      _result.incomingAmount = incomingAmount;
    }
    if (receivedAmount != null) {
      _result.receivedAmount = receivedAmount;
    }
    if (completed != null) {
      _result.completed = completed;
    }
    if (externalRef != null) {
      _result.externalRef = externalRef;
    }
    if (expiresAt != null) {
      _result.expiresAt = expiresAt;
    }
    if (createdAt != null) {
      _result.createdAt = createdAt;
    }
    if (updatedAt != null) {
      _result.updatedAt = updatedAt;
    }
    if (fromPaymentPointer != null) {
      _result.fromPaymentPointer = fromPaymentPointer;
    }
    return _result;
  }
  factory IncomingPayment.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IncomingPayment.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IncomingPayment clone() => IncomingPayment()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IncomingPayment copyWith(void Function(IncomingPayment) updates) => super.copyWith((message) => updates(message as IncomingPayment)) as IncomingPayment; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static IncomingPayment create() => IncomingPayment._();
  IncomingPayment createEmptyInstance() => create();
  static $pb.PbList<IncomingPayment> createRepeated() => $pb.PbList<IncomingPayment>();
  @$core.pragma('dart2js:noInline')
  static IncomingPayment getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<IncomingPayment>(create);
  static IncomingPayment? _defaultInstance;

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
  Amount get incomingAmount => $_getN(2);
  @$pb.TagNumber(3)
  set incomingAmount(Amount v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasIncomingAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearIncomingAmount() => clearField(3);
  @$pb.TagNumber(3)
  Amount ensureIncomingAmount() => $_ensure(2);

  @$pb.TagNumber(4)
  Amount get receivedAmount => $_getN(3);
  @$pb.TagNumber(4)
  set receivedAmount(Amount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasReceivedAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearReceivedAmount() => clearField(4);
  @$pb.TagNumber(4)
  Amount ensureReceivedAmount() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.bool get completed => $_getBF(4);
  @$pb.TagNumber(5)
  set completed($core.bool v) { $_setBool(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCompleted() => $_has(4);
  @$pb.TagNumber(5)
  void clearCompleted() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get externalRef => $_getSZ(5);
  @$pb.TagNumber(6)
  set externalRef($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasExternalRef() => $_has(5);
  @$pb.TagNumber(6)
  void clearExternalRef() => clearField(6);

  @$pb.TagNumber(7)
  $6.Timestamp get expiresAt => $_getN(6);
  @$pb.TagNumber(7)
  set expiresAt($6.Timestamp v) { setField(7, v); }
  @$pb.TagNumber(7)
  $core.bool hasExpiresAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearExpiresAt() => clearField(7);
  @$pb.TagNumber(7)
  $6.Timestamp ensureExpiresAt() => $_ensure(6);

  @$pb.TagNumber(8)
  $6.Timestamp get createdAt => $_getN(7);
  @$pb.TagNumber(8)
  set createdAt($6.Timestamp v) { setField(8, v); }
  @$pb.TagNumber(8)
  $core.bool hasCreatedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearCreatedAt() => clearField(8);
  @$pb.TagNumber(8)
  $6.Timestamp ensureCreatedAt() => $_ensure(7);

  @$pb.TagNumber(9)
  $6.Timestamp get updatedAt => $_getN(8);
  @$pb.TagNumber(9)
  set updatedAt($6.Timestamp v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasUpdatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearUpdatedAt() => clearField(9);
  @$pb.TagNumber(9)
  $6.Timestamp ensureUpdatedAt() => $_ensure(8);

  @$pb.TagNumber(10)
  $core.String get fromPaymentPointer => $_getSZ(9);
  @$pb.TagNumber(10)
  set fromPaymentPointer($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasFromPaymentPointer() => $_has(9);
  @$pb.TagNumber(10)
  void clearFromPaymentPointer() => clearField(10);
}

class LookupQuoteRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LookupQuoteRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  LookupQuoteRequest._() : super();
  factory LookupQuoteRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory LookupQuoteRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LookupQuoteRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LookupQuoteRequest clone() => LookupQuoteRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LookupQuoteRequest copyWith(void Function(LookupQuoteRequest) updates) => super.copyWith((message) => updates(message as LookupQuoteRequest)) as LookupQuoteRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static LookupQuoteRequest create() => LookupQuoteRequest._();
  LookupQuoteRequest createEmptyInstance() => create();
  static $pb.PbList<LookupQuoteRequest> createRepeated() => $pb.PbList<LookupQuoteRequest>();
  @$core.pragma('dart2js:noInline')
  static LookupQuoteRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LookupQuoteRequest>(create);
  static LookupQuoteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class CreateQuoteRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateQuoteRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sendPaymentPointer', protoName: 'sendPaymentPointer')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'receivePaymentPointer', protoName: 'receivePaymentPointer')
    ..aOM<Amount>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', subBuilder: Amount.create)
    ..aOM<$6.Timestamp>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'expiresAt', protoName: 'expiresAt', subBuilder: $6.Timestamp.create)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sendLinkedAccount', protoName: 'sendLinkedAccount')
    ..hasRequiredFields = false
  ;

  CreateQuoteRequest._() : super();
  factory CreateQuoteRequest({
    $core.String? sendPaymentPointer,
    $core.String? receivePaymentPointer,
    Amount? amount,
    $6.Timestamp? expiresAt,
    $core.String? description,
    $core.String? sendLinkedAccount,
  }) {
    final _result = create();
    if (sendPaymentPointer != null) {
      _result.sendPaymentPointer = sendPaymentPointer;
    }
    if (receivePaymentPointer != null) {
      _result.receivePaymentPointer = receivePaymentPointer;
    }
    if (amount != null) {
      _result.amount = amount;
    }
    if (expiresAt != null) {
      _result.expiresAt = expiresAt;
    }
    if (description != null) {
      _result.description = description;
    }
    if (sendLinkedAccount != null) {
      _result.sendLinkedAccount = sendLinkedAccount;
    }
    return _result;
  }
  factory CreateQuoteRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateQuoteRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateQuoteRequest clone() => CreateQuoteRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateQuoteRequest copyWith(void Function(CreateQuoteRequest) updates) => super.copyWith((message) => updates(message as CreateQuoteRequest)) as CreateQuoteRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CreateQuoteRequest create() => CreateQuoteRequest._();
  CreateQuoteRequest createEmptyInstance() => create();
  static $pb.PbList<CreateQuoteRequest> createRepeated() => $pb.PbList<CreateQuoteRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateQuoteRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateQuoteRequest>(create);
  static CreateQuoteRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sendPaymentPointer => $_getSZ(0);
  @$pb.TagNumber(1)
  set sendPaymentPointer($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSendPaymentPointer() => $_has(0);
  @$pb.TagNumber(1)
  void clearSendPaymentPointer() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get receivePaymentPointer => $_getSZ(1);
  @$pb.TagNumber(2)
  set receivePaymentPointer($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasReceivePaymentPointer() => $_has(1);
  @$pb.TagNumber(2)
  void clearReceivePaymentPointer() => clearField(2);

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
  $6.Timestamp get expiresAt => $_getN(3);
  @$pb.TagNumber(4)
  set expiresAt($6.Timestamp v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasExpiresAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearExpiresAt() => clearField(4);
  @$pb.TagNumber(4)
  $6.Timestamp ensureExpiresAt() => $_ensure(3);

  @$pb.TagNumber(5)
  $core.String get description => $_getSZ(4);
  @$pb.TagNumber(5)
  set description($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDescription() => $_has(4);
  @$pb.TagNumber(5)
  void clearDescription() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get sendLinkedAccount => $_getSZ(5);
  @$pb.TagNumber(6)
  set sendLinkedAccount($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasSendLinkedAccount() => $_has(5);
  @$pb.TagNumber(6)
  void clearSendLinkedAccount() => clearField(6);
}

class Quote extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Quote', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paymentPointer', protoName: 'paymentPointer')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'receiver')
    ..aOM<Amount>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sendAmount', protoName: 'sendAmount', subBuilder: Amount.create)
    ..aOM<Amount>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'receiveAmount', protoName: 'receiveAmount', subBuilder: Amount.create)
    ..aOM<$6.Timestamp>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'expiresAt', protoName: 'expiresAt', subBuilder: $6.Timestamp.create)
    ..aOM<$6.Timestamp>(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'createdAt', protoName: 'createdAt', subBuilder: $6.Timestamp.create)
    ..hasRequiredFields = false
  ;

  Quote._() : super();
  factory Quote({
    $core.String? id,
    $core.String? paymentPointer,
    $core.String? receiver,
    Amount? sendAmount,
    Amount? receiveAmount,
    $6.Timestamp? expiresAt,
    $6.Timestamp? createdAt,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (paymentPointer != null) {
      _result.paymentPointer = paymentPointer;
    }
    if (receiver != null) {
      _result.receiver = receiver;
    }
    if (sendAmount != null) {
      _result.sendAmount = sendAmount;
    }
    if (receiveAmount != null) {
      _result.receiveAmount = receiveAmount;
    }
    if (expiresAt != null) {
      _result.expiresAt = expiresAt;
    }
    if (createdAt != null) {
      _result.createdAt = createdAt;
    }
    return _result;
  }
  factory Quote.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Quote.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Quote clone() => Quote()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Quote copyWith(void Function(Quote) updates) => super.copyWith((message) => updates(message as Quote)) as Quote; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Quote create() => Quote._();
  Quote createEmptyInstance() => create();
  static $pb.PbList<Quote> createRepeated() => $pb.PbList<Quote>();
  @$core.pragma('dart2js:noInline')
  static Quote getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Quote>(create);
  static Quote? _defaultInstance;

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
  $core.String get receiver => $_getSZ(2);
  @$pb.TagNumber(3)
  set receiver($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasReceiver() => $_has(2);
  @$pb.TagNumber(3)
  void clearReceiver() => clearField(3);

  @$pb.TagNumber(4)
  Amount get sendAmount => $_getN(3);
  @$pb.TagNumber(4)
  set sendAmount(Amount v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasSendAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearSendAmount() => clearField(4);
  @$pb.TagNumber(4)
  Amount ensureSendAmount() => $_ensure(3);

  @$pb.TagNumber(5)
  Amount get receiveAmount => $_getN(4);
  @$pb.TagNumber(5)
  set receiveAmount(Amount v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasReceiveAmount() => $_has(4);
  @$pb.TagNumber(5)
  void clearReceiveAmount() => clearField(5);
  @$pb.TagNumber(5)
  Amount ensureReceiveAmount() => $_ensure(4);

  @$pb.TagNumber(6)
  $6.Timestamp get expiresAt => $_getN(5);
  @$pb.TagNumber(6)
  set expiresAt($6.Timestamp v) { setField(6, v); }
  @$pb.TagNumber(6)
  $core.bool hasExpiresAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearExpiresAt() => clearField(6);
  @$pb.TagNumber(6)
  $6.Timestamp ensureExpiresAt() => $_ensure(5);

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
}

class PaymentPointerExistsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PaymentPointerExistsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'url')
    ..hasRequiredFields = false
  ;

  PaymentPointerExistsRequest._() : super();
  factory PaymentPointerExistsRequest({
    $core.String? url,
  }) {
    final _result = create();
    if (url != null) {
      _result.url = url;
    }
    return _result;
  }
  factory PaymentPointerExistsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PaymentPointerExistsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PaymentPointerExistsRequest clone() => PaymentPointerExistsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PaymentPointerExistsRequest copyWith(void Function(PaymentPointerExistsRequest) updates) => super.copyWith((message) => updates(message as PaymentPointerExistsRequest)) as PaymentPointerExistsRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PaymentPointerExistsRequest create() => PaymentPointerExistsRequest._();
  PaymentPointerExistsRequest createEmptyInstance() => create();
  static $pb.PbList<PaymentPointerExistsRequest> createRepeated() => $pb.PbList<PaymentPointerExistsRequest>();
  @$core.pragma('dart2js:noInline')
  static PaymentPointerExistsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PaymentPointerExistsRequest>(create);
  static PaymentPointerExistsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class PaymentPointerExistsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PaymentPointerExistsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exists')
    ..hasRequiredFields = false
  ;

  PaymentPointerExistsResponse._() : super();
  factory PaymentPointerExistsResponse({
    $core.bool? exists,
  }) {
    final _result = create();
    if (exists != null) {
      _result.exists = exists;
    }
    return _result;
  }
  factory PaymentPointerExistsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PaymentPointerExistsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PaymentPointerExistsResponse clone() => PaymentPointerExistsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PaymentPointerExistsResponse copyWith(void Function(PaymentPointerExistsResponse) updates) => super.copyWith((message) => updates(message as PaymentPointerExistsResponse)) as PaymentPointerExistsResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PaymentPointerExistsResponse create() => PaymentPointerExistsResponse._();
  PaymentPointerExistsResponse createEmptyInstance() => create();
  static $pb.PbList<PaymentPointerExistsResponse> createRepeated() => $pb.PbList<PaymentPointerExistsResponse>();
  @$core.pragma('dart2js:noInline')
  static PaymentPointerExistsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PaymentPointerExistsResponse>(create);
  static PaymentPointerExistsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get exists => $_getBF(0);
  @$pb.TagNumber(1)
  set exists($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasExists() => $_has(0);
  @$pb.TagNumber(1)
  void clearExists() => clearField(1);
}

class GetPaymentPointerRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetPaymentPointerRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'url')
    ..hasRequiredFields = false
  ;

  GetPaymentPointerRequest._() : super();
  factory GetPaymentPointerRequest({
    $core.String? url,
  }) {
    final _result = create();
    if (url != null) {
      _result.url = url;
    }
    return _result;
  }
  factory GetPaymentPointerRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPaymentPointerRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPaymentPointerRequest clone() => GetPaymentPointerRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPaymentPointerRequest copyWith(void Function(GetPaymentPointerRequest) updates) => super.copyWith((message) => updates(message as GetPaymentPointerRequest)) as GetPaymentPointerRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GetPaymentPointerRequest create() => GetPaymentPointerRequest._();
  GetPaymentPointerRequest createEmptyInstance() => create();
  static $pb.PbList<GetPaymentPointerRequest> createRepeated() => $pb.PbList<GetPaymentPointerRequest>();
  @$core.pragma('dart2js:noInline')
  static GetPaymentPointerRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetPaymentPointerRequest>(create);
  static GetPaymentPointerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get url => $_getSZ(0);
  @$pb.TagNumber(1)
  set url($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearUrl() => clearField(1);
}

class ListWalletPaymentPointersResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ListWalletPaymentPointersResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<PaymentPointer>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pointers', $pb.PbFieldType.PM, subBuilder: PaymentPointer.create)
    ..hasRequiredFields = false
  ;

  ListWalletPaymentPointersResponse._() : super();
  factory ListWalletPaymentPointersResponse({
    $core.Iterable<PaymentPointer>? pointers,
  }) {
    final _result = create();
    if (pointers != null) {
      _result.pointers.addAll(pointers);
    }
    return _result;
  }
  factory ListWalletPaymentPointersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListWalletPaymentPointersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListWalletPaymentPointersResponse clone() => ListWalletPaymentPointersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListWalletPaymentPointersResponse copyWith(void Function(ListWalletPaymentPointersResponse) updates) => super.copyWith((message) => updates(message as ListWalletPaymentPointersResponse)) as ListWalletPaymentPointersResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ListWalletPaymentPointersResponse create() => ListWalletPaymentPointersResponse._();
  ListWalletPaymentPointersResponse createEmptyInstance() => create();
  static $pb.PbList<ListWalletPaymentPointersResponse> createRepeated() => $pb.PbList<ListWalletPaymentPointersResponse>();
  @$core.pragma('dart2js:noInline')
  static ListWalletPaymentPointersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListWalletPaymentPointersResponse>(create);
  static ListWalletPaymentPointersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<PaymentPointer> get pointers => $_getList(0);
}

class CreatePaymentPointerRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreatePaymentPointerRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'url')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'asset')
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'assetScale', $pb.PbFieldType.O3, protoName: 'assetScale')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'alias')
    ..hasRequiredFields = false
  ;

  CreatePaymentPointerRequest._() : super();
  factory CreatePaymentPointerRequest({
    $core.String? url,
    $core.String? asset,
    $core.int? assetScale,
    $core.String? alias,
  }) {
    final _result = create();
    if (url != null) {
      _result.url = url;
    }
    if (asset != null) {
      _result.asset = asset;
    }
    if (assetScale != null) {
      _result.assetScale = assetScale;
    }
    if (alias != null) {
      _result.alias = alias;
    }
    return _result;
  }
  factory CreatePaymentPointerRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreatePaymentPointerRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreatePaymentPointerRequest clone() => CreatePaymentPointerRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreatePaymentPointerRequest copyWith(void Function(CreatePaymentPointerRequest) updates) => super.copyWith((message) => updates(message as CreatePaymentPointerRequest)) as CreatePaymentPointerRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CreatePaymentPointerRequest create() => CreatePaymentPointerRequest._();
  CreatePaymentPointerRequest createEmptyInstance() => create();
  static $pb.PbList<CreatePaymentPointerRequest> createRepeated() => $pb.PbList<CreatePaymentPointerRequest>();
  @$core.pragma('dart2js:noInline')
  static CreatePaymentPointerRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreatePaymentPointerRequest>(create);
  static CreatePaymentPointerRequest? _defaultInstance;

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

class PaymentPointer extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'PaymentPointer', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'url')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'asset')
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'assetScale', $pb.PbFieldType.O3, protoName: 'assetScale')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'alias')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'walletID', protoName: 'walletID')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'formatted')
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'legalName', protoName: 'legalName')
    ..hasRequiredFields = false
  ;

  PaymentPointer._() : super();
  factory PaymentPointer({
    $core.String? url,
    $core.String? asset,
    $core.int? assetScale,
    $core.String? alias,
    $core.String? walletID,
    $core.String? formatted,
    $core.String? legalName,
  }) {
    final _result = create();
    if (url != null) {
      _result.url = url;
    }
    if (asset != null) {
      _result.asset = asset;
    }
    if (assetScale != null) {
      _result.assetScale = assetScale;
    }
    if (alias != null) {
      _result.alias = alias;
    }
    if (walletID != null) {
      _result.walletID = walletID;
    }
    if (formatted != null) {
      _result.formatted = formatted;
    }
    if (legalName != null) {
      _result.legalName = legalName;
    }
    return _result;
  }
  factory PaymentPointer.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PaymentPointer.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PaymentPointer clone() => PaymentPointer()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PaymentPointer copyWith(void Function(PaymentPointer) updates) => super.copyWith((message) => updates(message as PaymentPointer)) as PaymentPointer; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static PaymentPointer create() => PaymentPointer._();
  PaymentPointer createEmptyInstance() => create();
  static $pb.PbList<PaymentPointer> createRepeated() => $pb.PbList<PaymentPointer>();
  @$core.pragma('dart2js:noInline')
  static PaymentPointer getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PaymentPointer>(create);
  static PaymentPointer? _defaultInstance;

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

  @$pb.TagNumber(5)
  $core.String get walletID => $_getSZ(4);
  @$pb.TagNumber(5)
  set walletID($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasWalletID() => $_has(4);
  @$pb.TagNumber(5)
  void clearWalletID() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get formatted => $_getSZ(5);
  @$pb.TagNumber(6)
  set formatted($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasFormatted() => $_has(5);
  @$pb.TagNumber(6)
  void clearFormatted() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get legalName => $_getSZ(6);
  @$pb.TagNumber(7)
  set legalName($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasLegalName() => $_has(6);
  @$pb.TagNumber(7)
  void clearLegalName() => clearField(7);
}

class Empty extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Empty', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  Empty._() : super();
  factory Empty() => create();
  factory Empty.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Empty.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Empty clone() => Empty()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Empty copyWith(void Function(Empty) updates) => super.copyWith((message) => updates(message as Empty)) as Empty; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Empty create() => Empty._();
  Empty createEmptyInstance() => create();
  static $pb.PbList<Empty> createRepeated() => $pb.PbList<Empty>();
  @$core.pragma('dart2js:noInline')
  static Empty getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Empty>(create);
  static Empty? _defaultInstance;
}

class JWK extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'JWK', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'kty')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'kid')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'alg')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'x')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'crv')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'use')
    ..hasRequiredFields = false
  ;

  JWK._() : super();
  factory JWK({
    $core.String? kty,
    $core.String? kid,
    $core.String? alg,
    $core.String? x,
    $core.String? crv,
    $core.String? use,
  }) {
    final _result = create();
    if (kty != null) {
      _result.kty = kty;
    }
    if (kid != null) {
      _result.kid = kid;
    }
    if (alg != null) {
      _result.alg = alg;
    }
    if (x != null) {
      _result.x = x;
    }
    if (crv != null) {
      _result.crv = crv;
    }
    if (use != null) {
      _result.use = use;
    }
    return _result;
  }
  factory JWK.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory JWK.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  JWK clone() => JWK()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  JWK copyWith(void Function(JWK) updates) => super.copyWith((message) => updates(message as JWK)) as JWK; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static JWK create() => JWK._();
  JWK createEmptyInstance() => create();
  static $pb.PbList<JWK> createRepeated() => $pb.PbList<JWK>();
  @$core.pragma('dart2js:noInline')
  static JWK getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<JWK>(create);
  static JWK? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get kty => $_getSZ(0);
  @$pb.TagNumber(1)
  set kty($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasKty() => $_has(0);
  @$pb.TagNumber(1)
  void clearKty() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get kid => $_getSZ(1);
  @$pb.TagNumber(2)
  set kid($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasKid() => $_has(1);
  @$pb.TagNumber(2)
  void clearKid() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get alg => $_getSZ(2);
  @$pb.TagNumber(3)
  set alg($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAlg() => $_has(2);
  @$pb.TagNumber(3)
  void clearAlg() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get x => $_getSZ(3);
  @$pb.TagNumber(4)
  set x($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasX() => $_has(3);
  @$pb.TagNumber(4)
  void clearX() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get crv => $_getSZ(4);
  @$pb.TagNumber(5)
  set crv($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCrv() => $_has(4);
  @$pb.TagNumber(5)
  void clearCrv() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get use => $_getSZ(5);
  @$pb.TagNumber(6)
  set use($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasUse() => $_has(5);
  @$pb.TagNumber(6)
  void clearUse() => clearField(6);
}

class Transfer extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Transfer', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'type')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'state')
    ..aOM<$6.Timestamp>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'timestamp', subBuilder: $6.Timestamp.create)
    ..aOM<Amount>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', subBuilder: Amount.create)
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'foreignId', protoName: 'foreignId')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'linkedAccountId', protoName: 'linkedAccountId')
    ..hasRequiredFields = false
  ;

  Transfer._() : super();
  factory Transfer({
    $core.String? type,
    $core.String? state,
    $6.Timestamp? timestamp,
    Amount? amount,
    $core.String? foreignId,
    $core.String? linkedAccountId,
  }) {
    final _result = create();
    if (type != null) {
      _result.type = type;
    }
    if (state != null) {
      _result.state = state;
    }
    if (timestamp != null) {
      _result.timestamp = timestamp;
    }
    if (amount != null) {
      _result.amount = amount;
    }
    if (foreignId != null) {
      _result.foreignId = foreignId;
    }
    if (linkedAccountId != null) {
      _result.linkedAccountId = linkedAccountId;
    }
    return _result;
  }
  factory Transfer.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transfer.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Transfer clone() => Transfer()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Transfer copyWith(void Function(Transfer) updates) => super.copyWith((message) => updates(message as Transfer)) as Transfer; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ListStatementsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pPS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'periods')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nextPageToken', protoName: 'nextPageToken')
    ..hasRequiredFields = false
  ;

  ListStatementsResponse._() : super();
  factory ListStatementsResponse({
    $core.Iterable<$core.String>? periods,
    $core.String? nextPageToken,
  }) {
    final _result = create();
    if (periods != null) {
      _result.periods.addAll(periods);
    }
    if (nextPageToken != null) {
      _result.nextPageToken = nextPageToken;
    }
    return _result;
  }
  factory ListStatementsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListStatementsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListStatementsResponse clone() => ListStatementsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListStatementsResponse copyWith(void Function(ListStatementsResponse) updates) => super.copyWith((message) => updates(message as ListStatementsResponse)) as ListStatementsResponse; // ignore: deprecated_member_use
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

class GetStatementPDFRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetStatementPDFRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'period')
    ..hasRequiredFields = false
  ;

  GetStatementPDFRequest._() : super();
  factory GetStatementPDFRequest({
    $core.String? period,
  }) {
    final _result = create();
    if (period != null) {
      _result.period = period;
    }
    return _result;
  }
  factory GetStatementPDFRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetStatementPDFRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetStatementPDFRequest clone() => GetStatementPDFRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetStatementPDFRequest copyWith(void Function(GetStatementPDFRequest) updates) => super.copyWith((message) => updates(message as GetStatementPDFRequest)) as GetStatementPDFRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GetStatementPDFRequest create() => GetStatementPDFRequest._();
  GetStatementPDFRequest createEmptyInstance() => create();
  static $pb.PbList<GetStatementPDFRequest> createRepeated() => $pb.PbList<GetStatementPDFRequest>();
  @$core.pragma('dart2js:noInline')
  static GetStatementPDFRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetStatementPDFRequest>(create);
  static GetStatementPDFRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get period => $_getSZ(0);
  @$pb.TagNumber(1)
  set period($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPeriod() => $_has(0);
  @$pb.TagNumber(1)
  void clearPeriod() => clearField(1);
}

class StatementPDF extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'StatementPDF', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'chunks', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  StatementPDF._() : super();
  factory StatementPDF({
    $core.List<$core.int>? chunks,
  }) {
    final _result = create();
    if (chunks != null) {
      _result.chunks = chunks;
    }
    return _result;
  }
  factory StatementPDF.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory StatementPDF.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  StatementPDF clone() => StatementPDF()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  StatementPDF copyWith(void Function(StatementPDF) updates) => super.copyWith((message) => updates(message as StatementPDF)) as StatementPDF; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static StatementPDF create() => StatementPDF._();
  StatementPDF createEmptyInstance() => create();
  static $pb.PbList<StatementPDF> createRepeated() => $pb.PbList<StatementPDF>();
  @$core.pragma('dart2js:noInline')
  static StatementPDF getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StatementPDF>(create);
  static StatementPDF? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get chunks => $_getN(0);
  @$pb.TagNumber(1)
  set chunks($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasChunks() => $_has(0);
  @$pb.TagNumber(1)
  void clearChunks() => clearField(1);
}

class CreateSupportTicketRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateSupportTicketRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'description')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'firstName', protoName: 'firstName')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastName', protoName: 'lastName')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'email')
    ..hasRequiredFields = false
  ;

  CreateSupportTicketRequest._() : super();
  factory CreateSupportTicketRequest({
    $core.String? description,
    $core.String? firstName,
    $core.String? lastName,
    $core.String? email,
  }) {
    final _result = create();
    if (description != null) {
      _result.description = description;
    }
    if (firstName != null) {
      _result.firstName = firstName;
    }
    if (lastName != null) {
      _result.lastName = lastName;
    }
    if (email != null) {
      _result.email = email;
    }
    return _result;
  }
  factory CreateSupportTicketRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateSupportTicketRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateSupportTicketRequest clone() => CreateSupportTicketRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateSupportTicketRequest copyWith(void Function(CreateSupportTicketRequest) updates) => super.copyWith((message) => updates(message as CreateSupportTicketRequest)) as CreateSupportTicketRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'IndividualKYCResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'firstName', protoName: 'firstName')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastName', protoName: 'lastName')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'countryCode', protoName: 'countryCode')
    ..a<$core.int>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gender', $pb.PbFieldType.O3)
    ..aOM<$6.Timestamp>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'dateOfBirth', protoName: 'dateOfBirth', subBuilder: $6.Timestamp.create)
    ..aOM<Address>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address', subBuilder: Address.create)
    ..hasRequiredFields = false
  ;

  IndividualKYCResponse._() : super();
  factory IndividualKYCResponse({
    $core.String? firstName,
    $core.String? lastName,
    $core.String? countryCode,
    $core.int? gender,
    $6.Timestamp? dateOfBirth,
    Address? address,
  }) {
    final _result = create();
    if (firstName != null) {
      _result.firstName = firstName;
    }
    if (lastName != null) {
      _result.lastName = lastName;
    }
    if (countryCode != null) {
      _result.countryCode = countryCode;
    }
    if (gender != null) {
      _result.gender = gender;
    }
    if (dateOfBirth != null) {
      _result.dateOfBirth = dateOfBirth;
    }
    if (address != null) {
      _result.address = address;
    }
    return _result;
  }
  factory IndividualKYCResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IndividualKYCResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IndividualKYCResponse clone() => IndividualKYCResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IndividualKYCResponse copyWith(void Function(IndividualKYCResponse) updates) => super.copyWith((message) => updates(message as IndividualKYCResponse)) as IndividualKYCResponse; // ignore: deprecated_member_use
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
}

class UpdateIndividualKYCRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'UpdateIndividualKYCRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'firstName', protoName: 'firstName')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastName', protoName: 'lastName')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'countryCode', protoName: 'countryCode')
    ..a<$core.int>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gender', $pb.PbFieldType.O3)
    ..aOM<$6.Timestamp>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'dateOfBirth', protoName: 'dateOfBirth', subBuilder: $6.Timestamp.create)
    ..aOM<Address>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'address', subBuilder: Address.create)
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ipAddress', protoName: 'ipAddress')
    ..hasRequiredFields = false
  ;

  UpdateIndividualKYCRequest._() : super();
  factory UpdateIndividualKYCRequest({
    $core.String? firstName,
    $core.String? lastName,
    $core.String? countryCode,
    $core.int? gender,
    $6.Timestamp? dateOfBirth,
    Address? address,
    $core.String? ipAddress,
  }) {
    final _result = create();
    if (firstName != null) {
      _result.firstName = firstName;
    }
    if (lastName != null) {
      _result.lastName = lastName;
    }
    if (countryCode != null) {
      _result.countryCode = countryCode;
    }
    if (gender != null) {
      _result.gender = gender;
    }
    if (dateOfBirth != null) {
      _result.dateOfBirth = dateOfBirth;
    }
    if (address != null) {
      _result.address = address;
    }
    if (ipAddress != null) {
      _result.ipAddress = ipAddress;
    }
    return _result;
  }
  factory UpdateIndividualKYCRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateIndividualKYCRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateIndividualKYCRequest clone() => UpdateIndividualKYCRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateIndividualKYCRequest copyWith(void Function(UpdateIndividualKYCRequest) updates) => super.copyWith((message) => updates(message as UpdateIndividualKYCRequest)) as UpdateIndividualKYCRequest; // ignore: deprecated_member_use
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
}

class Address extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Address', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'line1')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'line2')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'building')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'apartment')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'city')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'state')
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'zipCode', protoName: 'zipCode')
    ..aOS(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'countryCode', protoName: 'countryCode')
    ..aOS(9, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'placeID', protoName: 'placeID')
    ..aOS(10, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'formattedAddress', protoName: 'formattedAddress')
    ..hasRequiredFields = false
  ;

  Address._() : super();
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
    final _result = create();
    if (line1 != null) {
      _result.line1 = line1;
    }
    if (line2 != null) {
      _result.line2 = line2;
    }
    if (building != null) {
      _result.building = building;
    }
    if (apartment != null) {
      _result.apartment = apartment;
    }
    if (city != null) {
      _result.city = city;
    }
    if (state != null) {
      _result.state = state;
    }
    if (zipCode != null) {
      _result.zipCode = zipCode;
    }
    if (countryCode != null) {
      _result.countryCode = countryCode;
    }
    if (placeID != null) {
      _result.placeID = placeID;
    }
    if (formattedAddress != null) {
      _result.formattedAddress = formattedAddress;
    }
    return _result;
  }
  factory Address.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Address.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Address clone() => Address()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Address copyWith(void Function(Address) updates) => super.copyWith((message) => updates(message as Address)) as Address; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'IsUSPSAddressResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'valid')
    ..hasRequiredFields = false
  ;

  IsUSPSAddressResponse._() : super();
  factory IsUSPSAddressResponse({
    $core.bool? valid,
  }) {
    final _result = create();
    if (valid != null) {
      _result.valid = valid;
    }
    return _result;
  }
  factory IsUSPSAddressResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IsUSPSAddressResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IsUSPSAddressResponse clone() => IsUSPSAddressResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IsUSPSAddressResponse copyWith(void Function(IsUSPSAddressResponse) updates) => super.copyWith((message) => updates(message as IsUSPSAddressResponse)) as IsUSPSAddressResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetBankAccountWidgetRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  GetBankAccountWidgetRequest._() : super();
  factory GetBankAccountWidgetRequest() => create();
  factory GetBankAccountWidgetRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetBankAccountWidgetRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetRequest clone() => GetBankAccountWidgetRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetRequest copyWith(void Function(GetBankAccountWidgetRequest) updates) => super.copyWith((message) => updates(message as GetBankAccountWidgetRequest)) as GetBankAccountWidgetRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetBankAccountWidgetResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'url')
    ..hasRequiredFields = false
  ;

  GetBankAccountWidgetResponse._() : super();
  factory GetBankAccountWidgetResponse({
    $core.String? url,
  }) {
    final _result = create();
    if (url != null) {
      _result.url = url;
    }
    return _result;
  }
  factory GetBankAccountWidgetResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetBankAccountWidgetResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetResponse clone() => GetBankAccountWidgetResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetBankAccountWidgetResponse copyWith(void Function(GetBankAccountWidgetResponse) updates) => super.copyWith((message) => updates(message as GetBankAccountWidgetResponse)) as GetBankAccountWidgetResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AddBankAccountRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'userGuid', protoName: 'userGuid')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'memberGuid', protoName: 'memberGuid')
    ..hasRequiredFields = false
  ;

  AddBankAccountRequest._() : super();
  factory AddBankAccountRequest({
    $core.String? userGuid,
    $core.String? memberGuid,
  }) {
    final _result = create();
    if (userGuid != null) {
      _result.userGuid = userGuid;
    }
    if (memberGuid != null) {
      _result.memberGuid = memberGuid;
    }
    return _result;
  }
  factory AddBankAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AddBankAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AddBankAccountRequest clone() => AddBankAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AddBankAccountRequest copyWith(void Function(AddBankAccountRequest) updates) => super.copyWith((message) => updates(message as AddBankAccountRequest)) as AddBankAccountRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'AddBankAccountResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'fundingsourceId', protoName: 'fundingsourceId')
    ..hasRequiredFields = false
  ;

  AddBankAccountResponse._() : super();
  factory AddBankAccountResponse({
    $core.String? fundingsourceId,
  }) {
    final _result = create();
    if (fundingsourceId != null) {
      _result.fundingsourceId = fundingsourceId;
    }
    return _result;
  }
  factory AddBankAccountResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AddBankAccountResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AddBankAccountResponse clone() => AddBankAccountResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AddBankAccountResponse copyWith(void Function(AddBankAccountResponse) updates) => super.copyWith((message) => updates(message as AddBankAccountResponse)) as AddBankAccountResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LinkedAccount', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'type')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'name')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'mask')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nickname')
    ..hasRequiredFields = false
  ;

  LinkedAccount._() : super();
  factory LinkedAccount({
    $core.String? id,
    $core.String? type,
    $core.String? name,
    $core.String? mask,
    $core.String? nickname,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (type != null) {
      _result.type = type;
    }
    if (name != null) {
      _result.name = name;
    }
    if (mask != null) {
      _result.mask = mask;
    }
    if (nickname != null) {
      _result.nickname = nickname;
    }
    return _result;
  }
  factory LinkedAccount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LinkedAccount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LinkedAccount clone() => LinkedAccount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LinkedAccount copyWith(void Function(LinkedAccount) updates) => super.copyWith((message) => updates(message as LinkedAccount)) as LinkedAccount; // ignore: deprecated_member_use
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
}

class GetSignupRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetSignupRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  GetSignupRequest._() : super();
  factory GetSignupRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory GetSignupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetSignupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetSignupRequest clone() => GetSignupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetSignupRequest copyWith(void Function(GetSignupRequest) updates) => super.copyWith((message) => updates(message as GetSignupRequest)) as GetSignupRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SetSignupUserDataRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'firstName', protoName: 'firstName')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastName', protoName: 'lastName')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'email')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'countryCode', protoName: 'countryCode')
    ..hasRequiredFields = false
  ;

  SetSignupUserDataRequest._() : super();
  factory SetSignupUserDataRequest({
    $core.String? id,
    $core.String? firstName,
    $core.String? lastName,
    $core.String? email,
    $core.String? countryCode,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (firstName != null) {
      _result.firstName = firstName;
    }
    if (lastName != null) {
      _result.lastName = lastName;
    }
    if (email != null) {
      _result.email = email;
    }
    if (countryCode != null) {
      _result.countryCode = countryCode;
    }
    return _result;
  }
  factory SetSignupUserDataRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupUserDataRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupUserDataRequest clone() => SetSignupUserDataRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupUserDataRequest copyWith(void Function(SetSignupUserDataRequest) updates) => super.copyWith((message) => updates(message as SetSignupUserDataRequest)) as SetSignupUserDataRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SetSignupUserDataResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  SetSignupUserDataResponse._() : super();
  factory SetSignupUserDataResponse({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory SetSignupUserDataResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupUserDataResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupUserDataResponse clone() => SetSignupUserDataResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupUserDataResponse copyWith(void Function(SetSignupUserDataResponse) updates) => super.copyWith((message) => updates(message as SetSignupUserDataResponse)) as SetSignupUserDataResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SetSignupMobileNumberRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'mobile')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'otp')
    ..hasRequiredFields = false
  ;

  SetSignupMobileNumberRequest._() : super();
  factory SetSignupMobileNumberRequest({
    $core.String? id,
    $core.String? mobile,
    $core.String? otp,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (mobile != null) {
      _result.mobile = mobile;
    }
    if (otp != null) {
      _result.otp = otp;
    }
    return _result;
  }
  factory SetSignupMobileNumberRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupMobileNumberRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupMobileNumberRequest clone() => SetSignupMobileNumberRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupMobileNumberRequest copyWith(void Function(SetSignupMobileNumberRequest) updates) => super.copyWith((message) => updates(message as SetSignupMobileNumberRequest)) as SetSignupMobileNumberRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Signup', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'firstName', protoName: 'firstName')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lastName', protoName: 'lastName')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'email')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'countryCode', protoName: 'countryCode')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'mobileNumber', protoName: 'mobileNumber')
    ..aOS(7, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'userId', protoName: 'userId')
    ..aOB(8, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'completed')
    ..hasRequiredFields = false
  ;

  Signup._() : super();
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
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (firstName != null) {
      _result.firstName = firstName;
    }
    if (lastName != null) {
      _result.lastName = lastName;
    }
    if (email != null) {
      _result.email = email;
    }
    if (countryCode != null) {
      _result.countryCode = countryCode;
    }
    if (mobileNumber != null) {
      _result.mobileNumber = mobileNumber;
    }
    if (userId != null) {
      _result.userId = userId;
    }
    if (completed != null) {
      _result.completed = completed;
    }
    return _result;
  }
  factory Signup.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Signup.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Signup clone() => Signup()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Signup copyWith(void Function(Signup) updates) => super.copyWith((message) => updates(message as Signup)) as Signup; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CompleteSignupRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'userId', protoName: 'userId')
    ..hasRequiredFields = false
  ;

  CompleteSignupRequest._() : super();
  factory CompleteSignupRequest({
    $core.String? id,
    $core.String? userId,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (userId != null) {
      _result.userId = userId;
    }
    return _result;
  }
  factory CompleteSignupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CompleteSignupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CompleteSignupRequest clone() => CompleteSignupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CompleteSignupRequest copyWith(void Function(CompleteSignupRequest) updates) => super.copyWith((message) => updates(message as CompleteSignupRequest)) as CompleteSignupRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateUserDefaultWalletRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'userID', protoName: 'userID')
    ..hasRequiredFields = false
  ;

  CreateUserDefaultWalletRequest._() : super();
  factory CreateUserDefaultWalletRequest({
    $core.String? userID,
  }) {
    final _result = create();
    if (userID != null) {
      _result.userID = userID;
    }
    return _result;
  }
  factory CreateUserDefaultWalletRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateUserDefaultWalletRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateUserDefaultWalletRequest clone() => CreateUserDefaultWalletRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateUserDefaultWalletRequest copyWith(void Function(CreateUserDefaultWalletRequest) updates) => super.copyWith((message) => updates(message as CreateUserDefaultWalletRequest)) as CreateUserDefaultWalletRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SendPhoneVerificationRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'to')
    ..hasRequiredFields = false
  ;

  SendPhoneVerificationRequest._() : super();
  factory SendPhoneVerificationRequest({
    $core.String? to,
  }) {
    final _result = create();
    if (to != null) {
      _result.to = to;
    }
    return _result;
  }
  factory SendPhoneVerificationRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SendPhoneVerificationRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SendPhoneVerificationRequest clone() => SendPhoneVerificationRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SendPhoneVerificationRequest copyWith(void Function(SendPhoneVerificationRequest) updates) => super.copyWith((message) => updates(message as SendPhoneVerificationRequest)) as SendPhoneVerificationRequest; // ignore: deprecated_member_use
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

class GetAgreementRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetAgreementRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  GetAgreementRequest._() : super();
  factory GetAgreementRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory GetAgreementRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetAgreementRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetAgreementRequest clone() => GetAgreementRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetAgreementRequest copyWith(void Function(GetAgreementRequest) updates) => super.copyWith((message) => updates(message as GetAgreementRequest)) as GetAgreementRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Agreement', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'content')
    ..hasRequiredFields = false
  ;

  Agreement._() : super();
  factory Agreement({
    $core.String? content,
  }) {
    final _result = create();
    if (content != null) {
      _result.content = content;
    }
    return _result;
  }
  factory Agreement.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Agreement.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Agreement clone() => Agreement()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Agreement copyWith(void Function(Agreement) updates) => super.copyWith((message) => updates(message as Agreement)) as Agreement; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SignAgreementsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pPS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'agreementIds', protoName: 'agreementIds')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'userId', protoName: 'userId')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ipAddress', protoName: 'ipAddress')
    ..hasRequiredFields = false
  ;

  SignAgreementsRequest._() : super();
  factory SignAgreementsRequest({
    $core.Iterable<$core.String>? agreementIds,
    $core.String? userId,
    $core.String? ipAddress,
  }) {
    final _result = create();
    if (agreementIds != null) {
      _result.agreementIds.addAll(agreementIds);
    }
    if (userId != null) {
      _result.userId = userId;
    }
    if (ipAddress != null) {
      _result.ipAddress = ipAddress;
    }
    return _result;
  }
  factory SignAgreementsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SignAgreementsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SignAgreementsRequest clone() => SignAgreementsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SignAgreementsRequest copyWith(void Function(SignAgreementsRequest) updates) => super.copyWith((message) => updates(message as SignAgreementsRequest)) as SignAgreementsRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SignAgreementsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'signed')
    ..hasRequiredFields = false
  ;

  SignAgreementsResponse._() : super();
  factory SignAgreementsResponse({
    $core.bool? signed,
  }) {
    final _result = create();
    if (signed != null) {
      _result.signed = signed;
    }
    return _result;
  }
  factory SignAgreementsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SignAgreementsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SignAgreementsResponse clone() => SignAgreementsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SignAgreementsResponse copyWith(void Function(SignAgreementsResponse) updates) => super.copyWith((message) => updates(message as SignAgreementsResponse)) as SignAgreementsResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'JoinWaitlistRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'email')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'countryCode')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'fullName')
    ..aOB(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'betaOptIn')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'mugId')
    ..hasRequiredFields = false
  ;

  JoinWaitlistRequest._() : super();
  factory JoinWaitlistRequest({
    $core.String? email,
    $core.String? countryCode,
    $core.String? fullName,
    $core.bool? betaOptIn,
    $core.String? mugId,
  }) {
    final _result = create();
    if (email != null) {
      _result.email = email;
    }
    if (countryCode != null) {
      _result.countryCode = countryCode;
    }
    if (fullName != null) {
      _result.fullName = fullName;
    }
    if (betaOptIn != null) {
      _result.betaOptIn = betaOptIn;
    }
    if (mugId != null) {
      _result.mugId = mugId;
    }
    return _result;
  }
  factory JoinWaitlistRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory JoinWaitlistRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  JoinWaitlistRequest clone() => JoinWaitlistRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  JoinWaitlistRequest copyWith(void Function(JoinWaitlistRequest) updates) => super.copyWith((message) => updates(message as JoinWaitlistRequest)) as JoinWaitlistRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'JoinWaitlistResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  JoinWaitlistResponse._() : super();
  factory JoinWaitlistResponse() => create();
  factory JoinWaitlistResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory JoinWaitlistResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  JoinWaitlistResponse clone() => JoinWaitlistResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  JoinWaitlistResponse copyWith(void Function(JoinWaitlistResponse) updates) => super.copyWith((message) => updates(message as JoinWaitlistResponse)) as JoinWaitlistResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'IsMugAvailableRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'mugId')
    ..hasRequiredFields = false
  ;

  IsMugAvailableRequest._() : super();
  factory IsMugAvailableRequest({
    $core.String? mugId,
  }) {
    final _result = create();
    if (mugId != null) {
      _result.mugId = mugId;
    }
    return _result;
  }
  factory IsMugAvailableRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IsMugAvailableRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IsMugAvailableRequest clone() => IsMugAvailableRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IsMugAvailableRequest copyWith(void Function(IsMugAvailableRequest) updates) => super.copyWith((message) => updates(message as IsMugAvailableRequest)) as IsMugAvailableRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'IsMugAvailableResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'available')
    ..hasRequiredFields = false
  ;

  IsMugAvailableResponse._() : super();
  factory IsMugAvailableResponse({
    $core.bool? available,
  }) {
    final _result = create();
    if (available != null) {
      _result.available = available;
    }
    return _result;
  }
  factory IsMugAvailableResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory IsMugAvailableResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  IsMugAvailableResponse clone() => IsMugAvailableResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  IsMugAvailableResponse copyWith(void Function(IsMugAvailableResponse) updates) => super.copyWith((message) => updates(message as IsMugAvailableResponse)) as IsMugAvailableResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetLinkedAccountsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<LinkedAccount>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'linkedAccounts', $pb.PbFieldType.PM, protoName: 'linkedAccounts', subBuilder: LinkedAccount.create)
    ..hasRequiredFields = false
  ;

  GetLinkedAccountsResponse._() : super();
  factory GetLinkedAccountsResponse({
    $core.Iterable<LinkedAccount>? linkedAccounts,
  }) {
    final _result = create();
    if (linkedAccounts != null) {
      _result.linkedAccounts.addAll(linkedAccounts);
    }
    return _result;
  }
  factory GetLinkedAccountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLinkedAccountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsResponse clone() => GetLinkedAccountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLinkedAccountsResponse copyWith(void Function(GetLinkedAccountsResponse) updates) => super.copyWith((message) => updates(message as GetLinkedAccountsResponse)) as GetLinkedAccountsResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetLinkedAccountRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  GetLinkedAccountRequest._() : super();
  factory GetLinkedAccountRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory GetLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLinkedAccountRequest clone() => GetLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLinkedAccountRequest copyWith(void Function(GetLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as GetLinkedAccountRequest)) as GetLinkedAccountRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SetNicknameLinkedAccountRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nickname')
    ..hasRequiredFields = false
  ;

  SetNicknameLinkedAccountRequest._() : super();
  factory SetNicknameLinkedAccountRequest({
    $core.String? id,
    $core.String? nickname,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (nickname != null) {
      _result.nickname = nickname;
    }
    return _result;
  }
  factory SetNicknameLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetNicknameLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetNicknameLinkedAccountRequest clone() => SetNicknameLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetNicknameLinkedAccountRequest copyWith(void Function(SetNicknameLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as SetNicknameLinkedAccountRequest)) as SetNicknameLinkedAccountRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'DeleteLinkedAccountRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  DeleteLinkedAccountRequest._() : super();
  factory DeleteLinkedAccountRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory DeleteLinkedAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory DeleteLinkedAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  DeleteLinkedAccountRequest clone() => DeleteLinkedAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  DeleteLinkedAccountRequest copyWith(void Function(DeleteLinkedAccountRequest) updates) => super.copyWith((message) => updates(message as DeleteLinkedAccountRequest)) as DeleteLinkedAccountRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Country', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'name')
    ..hasRequiredFields = false
  ;

  Country._() : super();
  factory Country({
    $core.String? id,
    $core.String? name,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (name != null) {
      _result.name = name;
    }
    return _result;
  }
  factory Country.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Country.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Country clone() => Country()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Country copyWith(void Function(Country) updates) => super.copyWith((message) => updates(message as Country)) as Country; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetCountriesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Country>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'countries', $pb.PbFieldType.PM, subBuilder: Country.create)
    ..hasRequiredFields = false
  ;

  GetCountriesResponse._() : super();
  factory GetCountriesResponse({
    $core.Iterable<Country>? countries,
  }) {
    final _result = create();
    if (countries != null) {
      _result.countries.addAll(countries);
    }
    return _result;
  }
  factory GetCountriesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetCountriesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetCountriesResponse clone() => GetCountriesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetCountriesResponse copyWith(void Function(GetCountriesResponse) updates) => super.copyWith((message) => updates(message as GetCountriesResponse)) as GetCountriesResponse; // ignore: deprecated_member_use
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

class MachnetWidgetToken extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'MachnetWidgetToken', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'value')
    ..aInt64(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'expiresInMinutes', protoName: 'expiresInMinutes')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'userId', protoName: 'userId')
    ..hasRequiredFields = false
  ;

  MachnetWidgetToken._() : super();
  factory MachnetWidgetToken({
    $core.String? value,
    $fixnum.Int64? expiresInMinutes,
    $core.String? userId,
  }) {
    final _result = create();
    if (value != null) {
      _result.value = value;
    }
    if (expiresInMinutes != null) {
      _result.expiresInMinutes = expiresInMinutes;
    }
    if (userId != null) {
      _result.userId = userId;
    }
    return _result;
  }
  factory MachnetWidgetToken.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory MachnetWidgetToken.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  MachnetWidgetToken clone() => MachnetWidgetToken()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  MachnetWidgetToken copyWith(void Function(MachnetWidgetToken) updates) => super.copyWith((message) => updates(message as MachnetWidgetToken)) as MachnetWidgetToken; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static MachnetWidgetToken create() => MachnetWidgetToken._();
  MachnetWidgetToken createEmptyInstance() => create();
  static $pb.PbList<MachnetWidgetToken> createRepeated() => $pb.PbList<MachnetWidgetToken>();
  @$core.pragma('dart2js:noInline')
  static MachnetWidgetToken getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MachnetWidgetToken>(create);
  static MachnetWidgetToken? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get value => $_getSZ(0);
  @$pb.TagNumber(1)
  set value($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValue() => $_has(0);
  @$pb.TagNumber(1)
  void clearValue() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get expiresInMinutes => $_getI64(1);
  @$pb.TagNumber(2)
  set expiresInMinutes($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasExpiresInMinutes() => $_has(1);
  @$pb.TagNumber(2)
  void clearExpiresInMinutes() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get userId => $_getSZ(2);
  @$pb.TagNumber(3)
  set userId($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasUserId() => $_has(2);
  @$pb.TagNumber(3)
  void clearUserId() => clearField(3);
}

class Branch extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Branch', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id', $pb.PbFieldType.OU3)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'name')
    ..hasRequiredFields = false
  ;

  Branch._() : super();
  factory Branch({
    $core.int? id,
    $core.String? name,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (name != null) {
      _result.name = name;
    }
    return _result;
  }
  factory Branch.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Branch.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Branch clone() => Branch()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Branch copyWith(void Function(Branch) updates) => super.copyWith((message) => updates(message as Branch)) as Branch; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Branch create() => Branch._();
  Branch createEmptyInstance() => create();
  static $pb.PbList<Branch> createRepeated() => $pb.PbList<Branch>();
  @$core.pragma('dart2js:noInline')
  static Branch getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Branch>(create);
  static Branch? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int v) { $_setUnsignedInt32(0, v); }
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

class Bank extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Bank', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id', $pb.PbFieldType.OU3)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'name')
    ..pc<Branch>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'branches', $pb.PbFieldType.PM, subBuilder: Branch.create)
    ..hasRequiredFields = false
  ;

  Bank._() : super();
  factory Bank({
    $core.int? id,
    $core.String? name,
    $core.Iterable<Branch>? branches,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (name != null) {
      _result.name = name;
    }
    if (branches != null) {
      _result.branches.addAll(branches);
    }
    return _result;
  }
  factory Bank.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Bank.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Bank clone() => Bank()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Bank copyWith(void Function(Bank) updates) => super.copyWith((message) => updates(message as Bank)) as Bank; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static Bank create() => Bank._();
  Bank createEmptyInstance() => create();
  static $pb.PbList<Bank> createRepeated() => $pb.PbList<Bank>();
  @$core.pragma('dart2js:noInline')
  static Bank getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Bank>(create);
  static Bank? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int v) { $_setUnsignedInt32(0, v); }
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
  $core.List<Branch> get branches => $_getList(2);
}

class ListBanksResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ListBanksResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Bank>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'banks', $pb.PbFieldType.PM, subBuilder: Bank.create)
    ..hasRequiredFields = false
  ;

  ListBanksResponse._() : super();
  factory ListBanksResponse({
    $core.Iterable<Bank>? banks,
  }) {
    final _result = create();
    if (banks != null) {
      _result.banks.addAll(banks);
    }
    return _result;
  }
  factory ListBanksResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListBanksResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListBanksResponse clone() => ListBanksResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListBanksResponse copyWith(void Function(ListBanksResponse) updates) => super.copyWith((message) => updates(message as ListBanksResponse)) as ListBanksResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ListBanksResponse create() => ListBanksResponse._();
  ListBanksResponse createEmptyInstance() => create();
  static $pb.PbList<ListBanksResponse> createRepeated() => $pb.PbList<ListBanksResponse>();
  @$core.pragma('dart2js:noInline')
  static ListBanksResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListBanksResponse>(create);
  static ListBanksResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Bank> get banks => $_getList(0);
}

class CreateReceiveBankAccountRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateReceiveBankAccountRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'name')
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'bankId', $pb.PbFieldType.OU3)
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'branchId', $pb.PbFieldType.OU3)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'accountType')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'accountNumber')
    ..aOS(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'otp')
    ..hasRequiredFields = false
  ;

  CreateReceiveBankAccountRequest._() : super();
  factory CreateReceiveBankAccountRequest({
    $core.String? name,
    $core.int? bankId,
    $core.int? branchId,
    $core.String? accountType,
    $core.String? accountNumber,
    $core.String? otp,
  }) {
    final _result = create();
    if (name != null) {
      _result.name = name;
    }
    if (bankId != null) {
      _result.bankId = bankId;
    }
    if (branchId != null) {
      _result.branchId = branchId;
    }
    if (accountType != null) {
      _result.accountType = accountType;
    }
    if (accountNumber != null) {
      _result.accountNumber = accountNumber;
    }
    if (otp != null) {
      _result.otp = otp;
    }
    return _result;
  }
  factory CreateReceiveBankAccountRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateReceiveBankAccountRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateReceiveBankAccountRequest clone() => CreateReceiveBankAccountRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateReceiveBankAccountRequest copyWith(void Function(CreateReceiveBankAccountRequest) updates) => super.copyWith((message) => updates(message as CreateReceiveBankAccountRequest)) as CreateReceiveBankAccountRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CreateReceiveBankAccountRequest create() => CreateReceiveBankAccountRequest._();
  CreateReceiveBankAccountRequest createEmptyInstance() => create();
  static $pb.PbList<CreateReceiveBankAccountRequest> createRepeated() => $pb.PbList<CreateReceiveBankAccountRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateReceiveBankAccountRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateReceiveBankAccountRequest>(create);
  static CreateReceiveBankAccountRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get bankId => $_getIZ(1);
  @$pb.TagNumber(2)
  set bankId($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBankId() => $_has(1);
  @$pb.TagNumber(2)
  void clearBankId() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get branchId => $_getIZ(2);
  @$pb.TagNumber(3)
  set branchId($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBranchId() => $_has(2);
  @$pb.TagNumber(3)
  void clearBranchId() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get accountType => $_getSZ(3);
  @$pb.TagNumber(4)
  set accountType($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasAccountType() => $_has(3);
  @$pb.TagNumber(4)
  void clearAccountType() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get accountNumber => $_getSZ(4);
  @$pb.TagNumber(5)
  set accountNumber($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasAccountNumber() => $_has(4);
  @$pb.TagNumber(5)
  void clearAccountNumber() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get otp => $_getSZ(5);
  @$pb.TagNumber(6)
  set otp($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasOtp() => $_has(5);
  @$pb.TagNumber(6)
  void clearOtp() => clearField(6);
}

class CanSignupRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CanSignupRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  CanSignupRequest._() : super();
  factory CanSignupRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory CanSignupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CanSignupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CanSignupRequest clone() => CanSignupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CanSignupRequest copyWith(void Function(CanSignupRequest) updates) => super.copyWith((message) => updates(message as CanSignupRequest)) as CanSignupRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CanSignupResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'canSignup', protoName: 'canSignup')
    ..hasRequiredFields = false
  ;

  CanSignupResponse._() : super();
  factory CanSignupResponse({
    $core.bool? canSignup,
  }) {
    final _result = create();
    if (canSignup != null) {
      _result.canSignup = canSignup;
    }
    return _result;
  }
  factory CanSignupResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CanSignupResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CanSignupResponse clone() => CanSignupResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CanSignupResponse copyWith(void Function(CanSignupResponse) updates) => super.copyWith((message) => updates(message as CanSignupResponse)) as CanSignupResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SetSignupCompleteRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'userId', protoName: 'userId')
    ..hasRequiredFields = false
  ;

  SetSignupCompleteRequest._() : super();
  factory SetSignupCompleteRequest({
    $core.String? id,
    $core.String? userId,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (userId != null) {
      _result.userId = userId;
    }
    return _result;
  }
  factory SetSignupCompleteRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetSignupCompleteRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetSignupCompleteRequest clone() => SetSignupCompleteRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetSignupCompleteRequest copyWith(void Function(SetSignupCompleteRequest) updates) => super.copyWith((message) => updates(message as SetSignupCompleteRequest)) as SetSignupCompleteRequest; // ignore: deprecated_member_use
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

class HasSendUserResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'HasSendUserResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hasSendUser', protoName: 'hasSendUser')
    ..hasRequiredFields = false
  ;

  HasSendUserResponse._() : super();
  factory HasSendUserResponse({
    $core.bool? hasSendUser,
  }) {
    final _result = create();
    if (hasSendUser != null) {
      _result.hasSendUser = hasSendUser;
    }
    return _result;
  }
  factory HasSendUserResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory HasSendUserResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  HasSendUserResponse clone() => HasSendUserResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  HasSendUserResponse copyWith(void Function(HasSendUserResponse) updates) => super.copyWith((message) => updates(message as HasSendUserResponse)) as HasSendUserResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static HasSendUserResponse create() => HasSendUserResponse._();
  HasSendUserResponse createEmptyInstance() => create();
  static $pb.PbList<HasSendUserResponse> createRepeated() => $pb.PbList<HasSendUserResponse>();
  @$core.pragma('dart2js:noInline')
  static HasSendUserResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HasSendUserResponse>(create);
  static HasSendUserResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get hasSendUser => $_getBF(0);
  @$pb.TagNumber(1)
  set hasSendUser($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHasSendUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearHasSendUser() => clearField(1);
}

class KYCStatusResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'KYCStatusResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'hasSendUser', protoName: 'hasSendUser')
    ..a<$core.int>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'kycStatus', $pb.PbFieldType.O3, protoName: 'kycStatus')
    ..pPS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'failedFields', protoName: 'failedFields')
    ..hasRequiredFields = false
  ;

  KYCStatusResponse._() : super();
  factory KYCStatusResponse({
    $core.bool? hasSendUser,
    $core.int? kycStatus,
    $core.Iterable<$core.String>? failedFields,
  }) {
    final _result = create();
    if (hasSendUser != null) {
      _result.hasSendUser = hasSendUser;
    }
    if (kycStatus != null) {
      _result.kycStatus = kycStatus;
    }
    if (failedFields != null) {
      _result.failedFields.addAll(failedFields);
    }
    return _result;
  }
  factory KYCStatusResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory KYCStatusResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  KYCStatusResponse clone() => KYCStatusResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  KYCStatusResponse copyWith(void Function(KYCStatusResponse) updates) => super.copyWith((message) => updates(message as KYCStatusResponse)) as KYCStatusResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static KYCStatusResponse create() => KYCStatusResponse._();
  KYCStatusResponse createEmptyInstance() => create();
  static $pb.PbList<KYCStatusResponse> createRepeated() => $pb.PbList<KYCStatusResponse>();
  @$core.pragma('dart2js:noInline')
  static KYCStatusResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<KYCStatusResponse>(create);
  static KYCStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get hasSendUser => $_getBF(0);
  @$pb.TagNumber(1)
  set hasSendUser($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHasSendUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearHasSendUser() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get kycStatus => $_getIZ(1);
  @$pb.TagNumber(2)
  set kycStatus($core.int v) { $_setSignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasKycStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearKycStatus() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.String> get failedFields => $_getList(2);
}

class CreateWalletRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateWalletRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nickname')
    ..hasRequiredFields = false
  ;

  CreateWalletRequest._() : super();
  factory CreateWalletRequest({
    $core.String? nickname,
  }) {
    final _result = create();
    if (nickname != null) {
      _result.nickname = nickname;
    }
    return _result;
  }
  factory CreateWalletRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateWalletRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateWalletRequest clone() => CreateWalletRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateWalletRequest copyWith(void Function(CreateWalletRequest) updates) => super.copyWith((message) => updates(message as CreateWalletRequest)) as CreateWalletRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CreateWalletRequest create() => CreateWalletRequest._();
  CreateWalletRequest createEmptyInstance() => create();
  static $pb.PbList<CreateWalletRequest> createRepeated() => $pb.PbList<CreateWalletRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateWalletRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateWalletRequest>(create);
  static CreateWalletRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get nickname => $_getSZ(0);
  @$pb.TagNumber(1)
  set nickname($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasNickname() => $_has(0);
  @$pb.TagNumber(1)
  void clearNickname() => clearField(1);
}

class WalletBalance extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'WalletBalance', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'balance', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'available', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  WalletBalance._() : super();
  factory WalletBalance({
    $fixnum.Int64? balance,
    $fixnum.Int64? available,
  }) {
    final _result = create();
    if (balance != null) {
      _result.balance = balance;
    }
    if (available != null) {
      _result.available = available;
    }
    return _result;
  }
  factory WalletBalance.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WalletBalance.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WalletBalance clone() => WalletBalance()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WalletBalance copyWith(void Function(WalletBalance) updates) => super.copyWith((message) => updates(message as WalletBalance)) as WalletBalance; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static WalletBalance create() => WalletBalance._();
  WalletBalance createEmptyInstance() => create();
  static $pb.PbList<WalletBalance> createRepeated() => $pb.PbList<WalletBalance>();
  @$core.pragma('dart2js:noInline')
  static WalletBalance getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WalletBalance>(create);
  static WalletBalance? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get balance => $_getI64(0);
  @$pb.TagNumber(1)
  set balance($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasBalance() => $_has(0);
  @$pb.TagNumber(1)
  void clearBalance() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get available => $_getI64(1);
  @$pb.TagNumber(2)
  set available($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAvailable() => $_has(1);
  @$pb.TagNumber(2)
  void clearAvailable() => clearField(2);
}

class WithdrawFromMachnetWalletRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'WithdrawFromMachnetWalletRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'toLinkedAccountId', protoName: 'toLinkedAccountId')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ipAddress', protoName: 'ipAddress')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'idempotencyKey', protoName: 'idempotencyKey')
    ..hasRequiredFields = false
  ;

  WithdrawFromMachnetWalletRequest._() : super();
  factory WithdrawFromMachnetWalletRequest({
    $core.String? toLinkedAccountId,
    $fixnum.Int64? amount,
    $core.String? ipAddress,
    $core.String? idempotencyKey,
  }) {
    final _result = create();
    if (toLinkedAccountId != null) {
      _result.toLinkedAccountId = toLinkedAccountId;
    }
    if (amount != null) {
      _result.amount = amount;
    }
    if (ipAddress != null) {
      _result.ipAddress = ipAddress;
    }
    if (idempotencyKey != null) {
      _result.idempotencyKey = idempotencyKey;
    }
    return _result;
  }
  factory WithdrawFromMachnetWalletRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WithdrawFromMachnetWalletRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WithdrawFromMachnetWalletRequest clone() => WithdrawFromMachnetWalletRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WithdrawFromMachnetWalletRequest copyWith(void Function(WithdrawFromMachnetWalletRequest) updates) => super.copyWith((message) => updates(message as WithdrawFromMachnetWalletRequest)) as WithdrawFromMachnetWalletRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static WithdrawFromMachnetWalletRequest create() => WithdrawFromMachnetWalletRequest._();
  WithdrawFromMachnetWalletRequest createEmptyInstance() => create();
  static $pb.PbList<WithdrawFromMachnetWalletRequest> createRepeated() => $pb.PbList<WithdrawFromMachnetWalletRequest>();
  @$core.pragma('dart2js:noInline')
  static WithdrawFromMachnetWalletRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WithdrawFromMachnetWalletRequest>(create);
  static WithdrawFromMachnetWalletRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get toLinkedAccountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set toLinkedAccountId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasToLinkedAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearToLinkedAccountId() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get amount => $_getI64(1);
  @$pb.TagNumber(2)
  set amount($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAmount() => $_has(1);
  @$pb.TagNumber(2)
  void clearAmount() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get ipAddress => $_getSZ(2);
  @$pb.TagNumber(3)
  set ipAddress($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasIpAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearIpAddress() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get idempotencyKey => $_getSZ(3);
  @$pb.TagNumber(4)
  set idempotencyKey($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasIdempotencyKey() => $_has(3);
  @$pb.TagNumber(4)
  void clearIdempotencyKey() => clearField(4);
}

class CheckMachnetTXLimitRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CheckMachnetTXLimitRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'currency')
    ..hasRequiredFields = false
  ;

  CheckMachnetTXLimitRequest._() : super();
  factory CheckMachnetTXLimitRequest({
    $fixnum.Int64? amount,
    $core.String? currency,
  }) {
    final _result = create();
    if (amount != null) {
      _result.amount = amount;
    }
    if (currency != null) {
      _result.currency = currency;
    }
    return _result;
  }
  factory CheckMachnetTXLimitRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CheckMachnetTXLimitRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CheckMachnetTXLimitRequest clone() => CheckMachnetTXLimitRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CheckMachnetTXLimitRequest copyWith(void Function(CheckMachnetTXLimitRequest) updates) => super.copyWith((message) => updates(message as CheckMachnetTXLimitRequest)) as CheckMachnetTXLimitRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CheckMachnetTXLimitRequest create() => CheckMachnetTXLimitRequest._();
  CheckMachnetTXLimitRequest createEmptyInstance() => create();
  static $pb.PbList<CheckMachnetTXLimitRequest> createRepeated() => $pb.PbList<CheckMachnetTXLimitRequest>();
  @$core.pragma('dart2js:noInline')
  static CheckMachnetTXLimitRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CheckMachnetTXLimitRequest>(create);
  static CheckMachnetTXLimitRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get amount => $_getI64(0);
  @$pb.TagNumber(1)
  set amount($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAmount() => $_has(0);
  @$pb.TagNumber(1)
  void clearAmount() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get currency => $_getSZ(1);
  @$pb.TagNumber(2)
  set currency($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCurrency() => $_has(1);
  @$pb.TagNumber(2)
  void clearCurrency() => clearField(2);
}

class CheckMachnetTXLimitResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CheckMachnetTXLimitResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOB(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'exceedsLimits', protoName: 'exceedsLimits')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'limitType', protoName: 'limitType')
    ..hasRequiredFields = false
  ;

  CheckMachnetTXLimitResponse._() : super();
  factory CheckMachnetTXLimitResponse({
    $core.bool? exceedsLimits,
    $core.String? limitType,
  }) {
    final _result = create();
    if (exceedsLimits != null) {
      _result.exceedsLimits = exceedsLimits;
    }
    if (limitType != null) {
      _result.limitType = limitType;
    }
    return _result;
  }
  factory CheckMachnetTXLimitResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CheckMachnetTXLimitResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CheckMachnetTXLimitResponse clone() => CheckMachnetTXLimitResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CheckMachnetTXLimitResponse copyWith(void Function(CheckMachnetTXLimitResponse) updates) => super.copyWith((message) => updates(message as CheckMachnetTXLimitResponse)) as CheckMachnetTXLimitResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static CheckMachnetTXLimitResponse create() => CheckMachnetTXLimitResponse._();
  CheckMachnetTXLimitResponse createEmptyInstance() => create();
  static $pb.PbList<CheckMachnetTXLimitResponse> createRepeated() => $pb.PbList<CheckMachnetTXLimitResponse>();
  @$core.pragma('dart2js:noInline')
  static CheckMachnetTXLimitResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CheckMachnetTXLimitResponse>(create);
  static CheckMachnetTXLimitResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get exceedsLimits => $_getBF(0);
  @$pb.TagNumber(1)
  set exceedsLimits($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasExceedsLimits() => $_has(0);
  @$pb.TagNumber(1)
  void clearExceedsLimits() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get limitType => $_getSZ(1);
  @$pb.TagNumber(2)
  set limitType($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLimitType() => $_has(1);
  @$pb.TagNumber(2)
  void clearLimitType() => clearField(2);
}

class StartMachnetWalletTopupRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'StartMachnetWalletTopupRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'fromLinkedAccountId', protoName: 'fromLinkedAccountId')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'amount', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'ipAddress', protoName: 'ipAddress')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'currency')
    ..aOS(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'idempotencyKey', protoName: 'idempotencyKey')
    ..hasRequiredFields = false
  ;

  StartMachnetWalletTopupRequest._() : super();
  factory StartMachnetWalletTopupRequest({
    $core.String? fromLinkedAccountId,
    $fixnum.Int64? amount,
    $core.String? ipAddress,
    $core.String? currency,
    $core.String? idempotencyKey,
  }) {
    final _result = create();
    if (fromLinkedAccountId != null) {
      _result.fromLinkedAccountId = fromLinkedAccountId;
    }
    if (amount != null) {
      _result.amount = amount;
    }
    if (ipAddress != null) {
      _result.ipAddress = ipAddress;
    }
    if (currency != null) {
      _result.currency = currency;
    }
    if (idempotencyKey != null) {
      _result.idempotencyKey = idempotencyKey;
    }
    return _result;
  }
  factory StartMachnetWalletTopupRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory StartMachnetWalletTopupRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  StartMachnetWalletTopupRequest clone() => StartMachnetWalletTopupRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  StartMachnetWalletTopupRequest copyWith(void Function(StartMachnetWalletTopupRequest) updates) => super.copyWith((message) => updates(message as StartMachnetWalletTopupRequest)) as StartMachnetWalletTopupRequest; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static StartMachnetWalletTopupRequest create() => StartMachnetWalletTopupRequest._();
  StartMachnetWalletTopupRequest createEmptyInstance() => create();
  static $pb.PbList<StartMachnetWalletTopupRequest> createRepeated() => $pb.PbList<StartMachnetWalletTopupRequest>();
  @$core.pragma('dart2js:noInline')
  static StartMachnetWalletTopupRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StartMachnetWalletTopupRequest>(create);
  static StartMachnetWalletTopupRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get fromLinkedAccountId => $_getSZ(0);
  @$pb.TagNumber(1)
  set fromLinkedAccountId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasFromLinkedAccountId() => $_has(0);
  @$pb.TagNumber(1)
  void clearFromLinkedAccountId() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get amount => $_getI64(1);
  @$pb.TagNumber(2)
  set amount($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAmount() => $_has(1);
  @$pb.TagNumber(2)
  void clearAmount() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get ipAddress => $_getSZ(2);
  @$pb.TagNumber(3)
  set ipAddress($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasIpAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearIpAddress() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get currency => $_getSZ(3);
  @$pb.TagNumber(4)
  set currency($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCurrency() => $_has(3);
  @$pb.TagNumber(4)
  void clearCurrency() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get idempotencyKey => $_getSZ(4);
  @$pb.TagNumber(5)
  set idempotencyKey($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasIdempotencyKey() => $_has(4);
  @$pb.TagNumber(5)
  void clearIdempotencyKey() => clearField(5);
}

class LookupTransactionRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LookupTransactionRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  LookupTransactionRequest._() : super();
  factory LookupTransactionRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory LookupTransactionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LookupTransactionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LookupTransactionRequest clone() => LookupTransactionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LookupTransactionRequest copyWith(void Function(LookupTransactionRequest) updates) => super.copyWith((message) => updates(message as LookupTransactionRequest)) as LookupTransactionRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetCurrentWalletResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  GetCurrentWalletResponse._() : super();
  factory GetCurrentWalletResponse({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory GetCurrentWalletResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetCurrentWalletResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetCurrentWalletResponse clone() => GetCurrentWalletResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetCurrentWalletResponse copyWith(void Function(GetCurrentWalletResponse) updates) => super.copyWith((message) => updates(message as GetCurrentWalletResponse)) as GetCurrentWalletResponse; // ignore: deprecated_member_use
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

class GetUserLimitsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetUserLimitsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<Limit>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'FundWallet', protoName: 'FundWallet', subBuilder: Limit.create)
    ..aOM<Limit>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'Withdrawal', protoName: 'Withdrawal', subBuilder: Limit.create)
    ..aOM<Limit>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'Transfer', protoName: 'Transfer', subBuilder: Limit.create)
    ..hasRequiredFields = false
  ;

  GetUserLimitsResponse._() : super();
  factory GetUserLimitsResponse({
    Limit? fundWallet,
    Limit? withdrawal,
    Limit? transfer,
  }) {
    final _result = create();
    if (fundWallet != null) {
      _result.fundWallet = fundWallet;
    }
    if (withdrawal != null) {
      _result.withdrawal = withdrawal;
    }
    if (transfer != null) {
      _result.transfer = transfer;
    }
    return _result;
  }
  factory GetUserLimitsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetUserLimitsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetUserLimitsResponse clone() => GetUserLimitsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetUserLimitsResponse copyWith(void Function(GetUserLimitsResponse) updates) => super.copyWith((message) => updates(message as GetUserLimitsResponse)) as GetUserLimitsResponse; // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GetUserLimitsResponse create() => GetUserLimitsResponse._();
  GetUserLimitsResponse createEmptyInstance() => create();
  static $pb.PbList<GetUserLimitsResponse> createRepeated() => $pb.PbList<GetUserLimitsResponse>();
  @$core.pragma('dart2js:noInline')
  static GetUserLimitsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetUserLimitsResponse>(create);
  static GetUserLimitsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Limit get fundWallet => $_getN(0);
  @$pb.TagNumber(1)
  set fundWallet(Limit v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasFundWallet() => $_has(0);
  @$pb.TagNumber(1)
  void clearFundWallet() => clearField(1);
  @$pb.TagNumber(1)
  Limit ensureFundWallet() => $_ensure(0);

  @$pb.TagNumber(2)
  Limit get withdrawal => $_getN(1);
  @$pb.TagNumber(2)
  set withdrawal(Limit v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasWithdrawal() => $_has(1);
  @$pb.TagNumber(2)
  void clearWithdrawal() => clearField(2);
  @$pb.TagNumber(2)
  Limit ensureWithdrawal() => $_ensure(1);

  @$pb.TagNumber(3)
  Limit get transfer => $_getN(2);
  @$pb.TagNumber(3)
  set transfer(Limit v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasTransfer() => $_has(2);
  @$pb.TagNumber(3)
  void clearTransfer() => clearField(3);
  @$pb.TagNumber(3)
  Limit ensureTransfer() => $_ensure(2);
}

class Limit extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Limit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOM<LimitAmount>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'Annual', protoName: 'Annual', subBuilder: LimitAmount.create)
    ..aOM<LimitAmount>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'Daily', protoName: 'Daily', subBuilder: LimitAmount.create)
    ..aOM<LimitAmount>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'Monthly', protoName: 'Monthly', subBuilder: LimitAmount.create)
    ..aOM<LimitAmount>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'WalletHold', protoName: 'WalletHold', subBuilder: LimitAmount.create)
    ..hasRequiredFields = false
  ;

  Limit._() : super();
  factory Limit({
    LimitAmount? annual,
    LimitAmount? daily,
    LimitAmount? monthly,
    LimitAmount? walletHold,
  }) {
    final _result = create();
    if (annual != null) {
      _result.annual = annual;
    }
    if (daily != null) {
      _result.daily = daily;
    }
    if (monthly != null) {
      _result.monthly = monthly;
    }
    if (walletHold != null) {
      _result.walletHold = walletHold;
    }
    return _result;
  }
  factory Limit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Limit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Limit clone() => Limit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Limit copyWith(void Function(Limit) updates) => super.copyWith((message) => updates(message as Limit)) as Limit; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'LimitAmount', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'remaining')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'total')
    ..a<$core.int>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'percentage', $pb.PbFieldType.O3)
    ..hasRequiredFields = false
  ;

  LimitAmount._() : super();
  factory LimitAmount({
    $core.String? remaining,
    $core.String? total,
    $core.int? percentage,
  }) {
    final _result = create();
    if (remaining != null) {
      _result.remaining = remaining;
    }
    if (total != null) {
      _result.total = total;
    }
    if (percentage != null) {
      _result.percentage = percentage;
    }
    return _result;
  }
  factory LimitAmount.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LimitAmount.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LimitAmount clone() => LimitAmount()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LimitAmount copyWith(void Function(LimitAmount) updates) => super.copyWith((message) => updates(message as LimitAmount)) as LimitAmount; // ignore: deprecated_member_use
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

class SetWalletNameRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'SetWalletNameRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'name')
    ..hasRequiredFields = false
  ;

  SetWalletNameRequest._() : super();
  factory SetWalletNameRequest({
    $core.String? name,
  }) {
    final _result = create();
    if (name != null) {
      _result.name = name;
    }
    return _result;
  }
  factory SetWalletNameRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SetWalletNameRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SetWalletNameRequest clone() => SetWalletNameRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SetWalletNameRequest copyWith(void Function(SetWalletNameRequest) updates) => super.copyWith((message) => updates(message as SetWalletNameRequest)) as SetWalletNameRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetPublicWalletDetailsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..hasRequiredFields = false
  ;

  GetPublicWalletDetailsRequest._() : super();
  factory GetPublicWalletDetailsRequest({
    $core.String? id,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    return _result;
  }
  factory GetPublicWalletDetailsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPublicWalletDetailsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsRequest clone() => GetPublicWalletDetailsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsRequest copyWith(void Function(GetPublicWalletDetailsRequest) updates) => super.copyWith((message) => updates(message as GetPublicWalletDetailsRequest)) as GetPublicWalletDetailsRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GetPublicWalletDetailsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'publicName', protoName: 'publicName')
    ..hasRequiredFields = false
  ;

  GetPublicWalletDetailsResponse._() : super();
  factory GetPublicWalletDetailsResponse({
    $core.String? id,
    $core.String? publicName,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (publicName != null) {
      _result.publicName = publicName;
    }
    return _result;
  }
  factory GetPublicWalletDetailsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetPublicWalletDetailsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsResponse clone() => GetPublicWalletDetailsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetPublicWalletDetailsResponse copyWith(void Function(GetPublicWalletDetailsResponse) updates) => super.copyWith((message) => updates(message as GetPublicWalletDetailsResponse)) as GetPublicWalletDetailsResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ListLimitsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<ConfiguredLimit>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'limits', $pb.PbFieldType.PM, subBuilder: ConfiguredLimit.create)
    ..hasRequiredFields = false
  ;

  ListLimitsResponse._() : super();
  factory ListLimitsResponse({
    $core.Iterable<ConfiguredLimit>? limits,
  }) {
    final _result = create();
    if (limits != null) {
      _result.limits.addAll(limits);
    }
    return _result;
  }
  factory ListLimitsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListLimitsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListLimitsResponse clone() => ListLimitsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListLimitsResponse copyWith(void Function(ListLimitsResponse) updates) => super.copyWith((message) => updates(message as ListLimitsResponse)) as ListLimitsResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ConfiguredLimit', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'foreignId', protoName: 'foreignId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'foreignDisplay', protoName: 'foreignDisplay')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'foreignType', protoName: 'foreignType')
    ..aOM<Amount>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'daily', subBuilder: Amount.create)
    ..aOM<Amount>(5, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'monthly', subBuilder: Amount.create)
    ..aOM<Amount>(6, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'overall', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  ConfiguredLimit._() : super();
  factory ConfiguredLimit({
    $core.String? foreignId,
    $core.String? foreignDisplay,
    $core.String? foreignType,
    Amount? daily,
    Amount? monthly,
    Amount? overall,
  }) {
    final _result = create();
    if (foreignId != null) {
      _result.foreignId = foreignId;
    }
    if (foreignDisplay != null) {
      _result.foreignDisplay = foreignDisplay;
    }
    if (foreignType != null) {
      _result.foreignType = foreignType;
    }
    if (daily != null) {
      _result.daily = daily;
    }
    if (monthly != null) {
      _result.monthly = monthly;
    }
    if (overall != null) {
      _result.overall = overall;
    }
    return _result;
  }
  factory ConfiguredLimit.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfiguredLimit.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfiguredLimit clone() => ConfiguredLimit()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfiguredLimit copyWith(void Function(ConfiguredLimit) updates) => super.copyWith((message) => updates(message as ConfiguredLimit)) as ConfiguredLimit; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'UpdateClientLimitsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientUrl', protoName: 'clientUrl')
    ..aOM<Amount>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'daily', subBuilder: Amount.create)
    ..aOM<Amount>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'monthly', subBuilder: Amount.create)
    ..aOM<Amount>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'overall', subBuilder: Amount.create)
    ..hasRequiredFields = false
  ;

  UpdateClientLimitsRequest._() : super();
  factory UpdateClientLimitsRequest({
    $core.String? clientUrl,
    Amount? daily,
    Amount? monthly,
    Amount? overall,
  }) {
    final _result = create();
    if (clientUrl != null) {
      _result.clientUrl = clientUrl;
    }
    if (daily != null) {
      _result.daily = daily;
    }
    if (monthly != null) {
      _result.monthly = monthly;
    }
    if (overall != null) {
      _result.overall = overall;
    }
    return _result;
  }
  factory UpdateClientLimitsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpdateClientLimitsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpdateClientLimitsRequest clone() => UpdateClientLimitsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpdateClientLimitsRequest copyWith(void Function(UpdateClientLimitsRequest) updates) => super.copyWith((message) => updates(message as UpdateClientLimitsRequest)) as UpdateClientLimitsRequest; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'Contact', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paymentPointer')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'name')
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'walletId')
    ..hasRequiredFields = false
  ;

  Contact._() : super();
  factory Contact({
    $core.String? id,
    $core.String? paymentPointer,
    $core.String? name,
    $core.String? walletId,
  }) {
    final _result = create();
    if (id != null) {
      _result.id = id;
    }
    if (paymentPointer != null) {
      _result.paymentPointer = paymentPointer;
    }
    if (name != null) {
      _result.name = name;
    }
    if (walletId != null) {
      _result.walletId = walletId;
    }
    return _result;
  }
  factory Contact.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Contact.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Contact clone() => Contact()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Contact copyWith(void Function(Contact) updates) => super.copyWith((message) => updates(message as Contact)) as Contact; // ignore: deprecated_member_use
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

class ListContactsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ListContactsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..pc<Contact>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'contacts', $pb.PbFieldType.PM, subBuilder: Contact.create)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nextPageToken')
    ..hasRequiredFields = false
  ;

  ListContactsResponse._() : super();
  factory ListContactsResponse({
    $core.Iterable<Contact>? contacts,
    $core.String? nextPageToken,
  }) {
    final _result = create();
    if (contacts != null) {
      _result.contacts.addAll(contacts);
    }
    if (nextPageToken != null) {
      _result.nextPageToken = nextPageToken;
    }
    return _result;
  }
  factory ListContactsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ListContactsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ListContactsResponse clone() => ListContactsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ListContactsResponse copyWith(void Function(ListContactsResponse) updates) => super.copyWith((message) => updates(message as ListContactsResponse)) as ListContactsResponse; // ignore: deprecated_member_use
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
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'CreateContactRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'backend.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paymentPointer')
    ..hasRequiredFields = false
  ;

  CreateContactRequest._() : super();
  factory CreateContactRequest({
    $core.String? paymentPointer,
  }) {
    final _result = create();
    if (paymentPointer != null) {
      _result.paymentPointer = paymentPointer;
    }
    return _result;
  }
  factory CreateContactRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateContactRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateContactRequest clone() => CreateContactRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateContactRequest copyWith(void Function(CreateContactRequest) updates) => super.copyWith((message) => updates(message as CreateContactRequest)) as CreateContactRequest; // ignore: deprecated_member_use
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

