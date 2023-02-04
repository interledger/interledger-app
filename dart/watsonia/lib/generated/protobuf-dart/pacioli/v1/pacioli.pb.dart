//
//  Generated code. Do not modify.
//  source: pacioli/v1/pacioli.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

class Empty extends $pb.GeneratedMessage {
  factory Empty() => create();
  Empty._() : super();
  factory Empty.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Empty.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Empty', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
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

class Ledger extends $pb.GeneratedMessage {
  factory Ledger({
    $core.int? id,
    $core.String? name,
    $core.String? asset,
    $core.int? scale,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (name != null) {
      $result.name = name;
    }
    if (asset != null) {
      $result.asset = asset;
    }
    if (scale != null) {
      $result.scale = scale;
    }
    return $result;
  }
  Ledger._() : super();
  factory Ledger.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Ledger.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Ledger', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'id', $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'asset')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'scale', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Ledger clone() => Ledger()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Ledger copyWith(void Function(Ledger) updates) => super.copyWith((message) => updates(message as Ledger)) as Ledger;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Ledger create() => Ledger._();
  Ledger createEmptyInstance() => create();
  static $pb.PbList<Ledger> createRepeated() => $pb.PbList<Ledger>();
  @$core.pragma('dart2js:noInline')
  static Ledger getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Ledger>(create);
  static Ledger? _defaultInstance;

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
  $core.String get asset => $_getSZ(2);
  @$pb.TagNumber(3)
  set asset($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAsset() => $_has(2);
  @$pb.TagNumber(3)
  void clearAsset() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get scale => $_getIZ(3);
  @$pb.TagNumber(4)
  set scale($core.int v) { $_setUnsignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasScale() => $_has(3);
  @$pb.TagNumber(4)
  void clearScale() => clearField(4);
}

class ConfigureLedgersRequest extends $pb.GeneratedMessage {
  factory ConfigureLedgersRequest({
    $core.Iterable<Ledger>? args,
  }) {
    final $result = create();
    if (args != null) {
      $result.args.addAll(args);
    }
    return $result;
  }
  ConfigureLedgersRequest._() : super();
  factory ConfigureLedgersRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfigureLedgersRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConfigureLedgersRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<Ledger>(1, _omitFieldNames ? '' : 'args', $pb.PbFieldType.PM, subBuilder: Ledger.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfigureLedgersRequest clone() => ConfigureLedgersRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfigureLedgersRequest copyWith(void Function(ConfigureLedgersRequest) updates) => super.copyWith((message) => updates(message as ConfigureLedgersRequest)) as ConfigureLedgersRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConfigureLedgersRequest create() => ConfigureLedgersRequest._();
  ConfigureLedgersRequest createEmptyInstance() => create();
  static $pb.PbList<ConfigureLedgersRequest> createRepeated() => $pb.PbList<ConfigureLedgersRequest>();
  @$core.pragma('dart2js:noInline')
  static ConfigureLedgersRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfigureLedgersRequest>(create);
  static ConfigureLedgersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Ledger> get args => $_getList(0);
}

class ConfigureLedgersResponse extends $pb.GeneratedMessage {
  factory ConfigureLedgersResponse({
    $core.Iterable<EventError>? errors,
  }) {
    final $result = create();
    if (errors != null) {
      $result.errors.addAll(errors);
    }
    return $result;
  }
  ConfigureLedgersResponse._() : super();
  factory ConfigureLedgersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfigureLedgersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConfigureLedgersResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<EventError>(1, _omitFieldNames ? '' : 'errors', $pb.PbFieldType.PM, subBuilder: EventError.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfigureLedgersResponse clone() => ConfigureLedgersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfigureLedgersResponse copyWith(void Function(ConfigureLedgersResponse) updates) => super.copyWith((message) => updates(message as ConfigureLedgersResponse)) as ConfigureLedgersResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConfigureLedgersResponse create() => ConfigureLedgersResponse._();
  ConfigureLedgersResponse createEmptyInstance() => create();
  static $pb.PbList<ConfigureLedgersResponse> createRepeated() => $pb.PbList<ConfigureLedgersResponse>();
  @$core.pragma('dart2js:noInline')
  static ConfigureLedgersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfigureLedgersResponse>(create);
  static ConfigureLedgersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<EventError> get errors => $_getList(0);
}

class GetLedgersRequest extends $pb.GeneratedMessage {
  factory GetLedgersRequest({
    $core.Iterable<$core.int>? ids,
  }) {
    final $result = create();
    if (ids != null) {
      $result.ids.addAll(ids);
    }
    return $result;
  }
  GetLedgersRequest._() : super();
  factory GetLedgersRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLedgersRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLedgersRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..p<$core.int>(1, _omitFieldNames ? '' : 'ids', $pb.PbFieldType.KU3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLedgersRequest clone() => GetLedgersRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLedgersRequest copyWith(void Function(GetLedgersRequest) updates) => super.copyWith((message) => updates(message as GetLedgersRequest)) as GetLedgersRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetLedgersRequest create() => GetLedgersRequest._();
  GetLedgersRequest createEmptyInstance() => create();
  static $pb.PbList<GetLedgersRequest> createRepeated() => $pb.PbList<GetLedgersRequest>();
  @$core.pragma('dart2js:noInline')
  static GetLedgersRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetLedgersRequest>(create);
  static GetLedgersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get ids => $_getList(0);
}

class GetLedgersResponse extends $pb.GeneratedMessage {
  factory GetLedgersResponse({
    $core.Iterable<Ledger>? ledgers,
  }) {
    final $result = create();
    if (ledgers != null) {
      $result.ledgers.addAll(ledgers);
    }
    return $result;
  }
  GetLedgersResponse._() : super();
  factory GetLedgersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetLedgersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetLedgersResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<Ledger>(1, _omitFieldNames ? '' : 'ledgers', $pb.PbFieldType.PM, subBuilder: Ledger.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetLedgersResponse clone() => GetLedgersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetLedgersResponse copyWith(void Function(GetLedgersResponse) updates) => super.copyWith((message) => updates(message as GetLedgersResponse)) as GetLedgersResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetLedgersResponse create() => GetLedgersResponse._();
  GetLedgersResponse createEmptyInstance() => create();
  static $pb.PbList<GetLedgersResponse> createRepeated() => $pb.PbList<GetLedgersResponse>();
  @$core.pragma('dart2js:noInline')
  static GetLedgersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetLedgersResponse>(create);
  static GetLedgersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Ledger> get ledgers => $_getList(0);
}

class Account extends $pb.GeneratedMessage {
  factory Account({
    $core.String? id,
    $core.int? ledgerId,
    $core.int? code,
    $fixnum.Int64? debitsReserved,
    $fixnum.Int64? debitsAccepted,
    $fixnum.Int64? creditsReserved,
    $fixnum.Int64? creditsAccepted,
    $core.bool? debitsMustNotExceedCredits,
    $core.bool? creditsMustNotExceedDebits,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (ledgerId != null) {
      $result.ledgerId = ledgerId;
    }
    if (code != null) {
      $result.code = code;
    }
    if (debitsReserved != null) {
      $result.debitsReserved = debitsReserved;
    }
    if (debitsAccepted != null) {
      $result.debitsAccepted = debitsAccepted;
    }
    if (creditsReserved != null) {
      $result.creditsReserved = creditsReserved;
    }
    if (creditsAccepted != null) {
      $result.creditsAccepted = creditsAccepted;
    }
    if (debitsMustNotExceedCredits != null) {
      $result.debitsMustNotExceedCredits = debitsMustNotExceedCredits;
    }
    if (creditsMustNotExceedDebits != null) {
      $result.creditsMustNotExceedDebits = creditsMustNotExceedDebits;
    }
    return $result;
  }
  Account._() : super();
  factory Account.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Account.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Account', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'ledgerId', $pb.PbFieldType.OU3, protoName: 'ledgerId')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'code', $pb.PbFieldType.OU3)
    ..a<$fixnum.Int64>(4, _omitFieldNames ? '' : 'debitsReserved', $pb.PbFieldType.OU6, protoName: 'debitsReserved', defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(5, _omitFieldNames ? '' : 'debitsAccepted', $pb.PbFieldType.OU6, protoName: 'debitsAccepted', defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(6, _omitFieldNames ? '' : 'creditsReserved', $pb.PbFieldType.OU6, protoName: 'creditsReserved', defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(7, _omitFieldNames ? '' : 'creditsAccepted', $pb.PbFieldType.OU6, protoName: 'creditsAccepted', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOB(8, _omitFieldNames ? '' : 'debitsMustNotExceedCredits', protoName: 'debitsMustNotExceedCredits')
    ..aOB(9, _omitFieldNames ? '' : 'creditsMustNotExceedDebits', protoName: 'creditsMustNotExceedDebits')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Account clone() => Account()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Account copyWith(void Function(Account) updates) => super.copyWith((message) => updates(message as Account)) as Account;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Account create() => Account._();
  Account createEmptyInstance() => create();
  static $pb.PbList<Account> createRepeated() => $pb.PbList<Account>();
  @$core.pragma('dart2js:noInline')
  static Account getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Account>(create);
  static Account? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get ledgerId => $_getIZ(1);
  @$pb.TagNumber(2)
  set ledgerId($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearLedgerId() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get code => $_getIZ(2);
  @$pb.TagNumber(3)
  set code($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearCode() => clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get debitsReserved => $_getI64(3);
  @$pb.TagNumber(4)
  set debitsReserved($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasDebitsReserved() => $_has(3);
  @$pb.TagNumber(4)
  void clearDebitsReserved() => clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get debitsAccepted => $_getI64(4);
  @$pb.TagNumber(5)
  set debitsAccepted($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasDebitsAccepted() => $_has(4);
  @$pb.TagNumber(5)
  void clearDebitsAccepted() => clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get creditsReserved => $_getI64(5);
  @$pb.TagNumber(6)
  set creditsReserved($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasCreditsReserved() => $_has(5);
  @$pb.TagNumber(6)
  void clearCreditsReserved() => clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get creditsAccepted => $_getI64(6);
  @$pb.TagNumber(7)
  set creditsAccepted($fixnum.Int64 v) { $_setInt64(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasCreditsAccepted() => $_has(6);
  @$pb.TagNumber(7)
  void clearCreditsAccepted() => clearField(7);

  @$pb.TagNumber(8)
  $core.bool get debitsMustNotExceedCredits => $_getBF(7);
  @$pb.TagNumber(8)
  set debitsMustNotExceedCredits($core.bool v) { $_setBool(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasDebitsMustNotExceedCredits() => $_has(7);
  @$pb.TagNumber(8)
  void clearDebitsMustNotExceedCredits() => clearField(8);

  @$pb.TagNumber(9)
  $core.bool get creditsMustNotExceedDebits => $_getBF(8);
  @$pb.TagNumber(9)
  set creditsMustNotExceedDebits($core.bool v) { $_setBool(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasCreditsMustNotExceedDebits() => $_has(8);
  @$pb.TagNumber(9)
  void clearCreditsMustNotExceedDebits() => clearField(9);
}

class ConfigureAccountsArgs extends $pb.GeneratedMessage {
  factory ConfigureAccountsArgs({
    $core.String? id,
    $core.int? ledgerId,
    $core.int? code,
    $core.bool? debitsMustNotExceedCredits,
    $core.bool? creditsMustNotExceedDebits,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (ledgerId != null) {
      $result.ledgerId = ledgerId;
    }
    if (code != null) {
      $result.code = code;
    }
    if (debitsMustNotExceedCredits != null) {
      $result.debitsMustNotExceedCredits = debitsMustNotExceedCredits;
    }
    if (creditsMustNotExceedDebits != null) {
      $result.creditsMustNotExceedDebits = creditsMustNotExceedDebits;
    }
    return $result;
  }
  ConfigureAccountsArgs._() : super();
  factory ConfigureAccountsArgs.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfigureAccountsArgs.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConfigureAccountsArgs', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'ledgerId', $pb.PbFieldType.OU3, protoName: 'ledgerId')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'code', $pb.PbFieldType.OU3)
    ..aOB(4, _omitFieldNames ? '' : 'debitsMustNotExceedCredits', protoName: 'debitsMustNotExceedCredits')
    ..aOB(5, _omitFieldNames ? '' : 'creditsMustNotExceedDebits', protoName: 'creditsMustNotExceedDebits')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfigureAccountsArgs clone() => ConfigureAccountsArgs()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfigureAccountsArgs copyWith(void Function(ConfigureAccountsArgs) updates) => super.copyWith((message) => updates(message as ConfigureAccountsArgs)) as ConfigureAccountsArgs;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConfigureAccountsArgs create() => ConfigureAccountsArgs._();
  ConfigureAccountsArgs createEmptyInstance() => create();
  static $pb.PbList<ConfigureAccountsArgs> createRepeated() => $pb.PbList<ConfigureAccountsArgs>();
  @$core.pragma('dart2js:noInline')
  static ConfigureAccountsArgs getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfigureAccountsArgs>(create);
  static ConfigureAccountsArgs? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get ledgerId => $_getIZ(1);
  @$pb.TagNumber(2)
  set ledgerId($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasLedgerId() => $_has(1);
  @$pb.TagNumber(2)
  void clearLedgerId() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get code => $_getIZ(2);
  @$pb.TagNumber(3)
  set code($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCode() => $_has(2);
  @$pb.TagNumber(3)
  void clearCode() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get debitsMustNotExceedCredits => $_getBF(3);
  @$pb.TagNumber(4)
  set debitsMustNotExceedCredits($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasDebitsMustNotExceedCredits() => $_has(3);
  @$pb.TagNumber(4)
  void clearDebitsMustNotExceedCredits() => clearField(4);

  @$pb.TagNumber(5)
  $core.bool get creditsMustNotExceedDebits => $_getBF(4);
  @$pb.TagNumber(5)
  set creditsMustNotExceedDebits($core.bool v) { $_setBool(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCreditsMustNotExceedDebits() => $_has(4);
  @$pb.TagNumber(5)
  void clearCreditsMustNotExceedDebits() => clearField(5);
}

class ConfigureAccountsRequest extends $pb.GeneratedMessage {
  factory ConfigureAccountsRequest({
    $core.Iterable<ConfigureAccountsArgs>? args,
  }) {
    final $result = create();
    if (args != null) {
      $result.args.addAll(args);
    }
    return $result;
  }
  ConfigureAccountsRequest._() : super();
  factory ConfigureAccountsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfigureAccountsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConfigureAccountsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<ConfigureAccountsArgs>(1, _omitFieldNames ? '' : 'args', $pb.PbFieldType.PM, subBuilder: ConfigureAccountsArgs.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfigureAccountsRequest clone() => ConfigureAccountsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfigureAccountsRequest copyWith(void Function(ConfigureAccountsRequest) updates) => super.copyWith((message) => updates(message as ConfigureAccountsRequest)) as ConfigureAccountsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConfigureAccountsRequest create() => ConfigureAccountsRequest._();
  ConfigureAccountsRequest createEmptyInstance() => create();
  static $pb.PbList<ConfigureAccountsRequest> createRepeated() => $pb.PbList<ConfigureAccountsRequest>();
  @$core.pragma('dart2js:noInline')
  static ConfigureAccountsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfigureAccountsRequest>(create);
  static ConfigureAccountsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<ConfigureAccountsArgs> get args => $_getList(0);
}

class ConfigureAccountsResponse extends $pb.GeneratedMessage {
  factory ConfigureAccountsResponse({
    $core.Iterable<EventError>? errors,
  }) {
    final $result = create();
    if (errors != null) {
      $result.errors.addAll(errors);
    }
    return $result;
  }
  ConfigureAccountsResponse._() : super();
  factory ConfigureAccountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ConfigureAccountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ConfigureAccountsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<EventError>(1, _omitFieldNames ? '' : 'errors', $pb.PbFieldType.PM, subBuilder: EventError.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ConfigureAccountsResponse clone() => ConfigureAccountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ConfigureAccountsResponse copyWith(void Function(ConfigureAccountsResponse) updates) => super.copyWith((message) => updates(message as ConfigureAccountsResponse)) as ConfigureAccountsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ConfigureAccountsResponse create() => ConfigureAccountsResponse._();
  ConfigureAccountsResponse createEmptyInstance() => create();
  static $pb.PbList<ConfigureAccountsResponse> createRepeated() => $pb.PbList<ConfigureAccountsResponse>();
  @$core.pragma('dart2js:noInline')
  static ConfigureAccountsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfigureAccountsResponse>(create);
  static ConfigureAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<EventError> get errors => $_getList(0);
}

class GetAccountsRequest extends $pb.GeneratedMessage {
  factory GetAccountsRequest({
    $core.Iterable<$core.String>? ids,
  }) {
    final $result = create();
    if (ids != null) {
      $result.ids.addAll(ids);
    }
    return $result;
  }
  GetAccountsRequest._() : super();
  factory GetAccountsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetAccountsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetAccountsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'ids')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetAccountsRequest clone() => GetAccountsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetAccountsRequest copyWith(void Function(GetAccountsRequest) updates) => super.copyWith((message) => updates(message as GetAccountsRequest)) as GetAccountsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetAccountsRequest create() => GetAccountsRequest._();
  GetAccountsRequest createEmptyInstance() => create();
  static $pb.PbList<GetAccountsRequest> createRepeated() => $pb.PbList<GetAccountsRequest>();
  @$core.pragma('dart2js:noInline')
  static GetAccountsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetAccountsRequest>(create);
  static GetAccountsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get ids => $_getList(0);
}

class GetAccountsResponse extends $pb.GeneratedMessage {
  factory GetAccountsResponse({
    $core.Iterable<Account>? accounts,
  }) {
    final $result = create();
    if (accounts != null) {
      $result.accounts.addAll(accounts);
    }
    return $result;
  }
  GetAccountsResponse._() : super();
  factory GetAccountsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetAccountsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetAccountsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<Account>(1, _omitFieldNames ? '' : 'accounts', $pb.PbFieldType.PM, subBuilder: Account.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetAccountsResponse clone() => GetAccountsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetAccountsResponse copyWith(void Function(GetAccountsResponse) updates) => super.copyWith((message) => updates(message as GetAccountsResponse)) as GetAccountsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetAccountsResponse create() => GetAccountsResponse._();
  GetAccountsResponse createEmptyInstance() => create();
  static $pb.PbList<GetAccountsResponse> createRepeated() => $pb.PbList<GetAccountsResponse>();
  @$core.pragma('dart2js:noInline')
  static GetAccountsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetAccountsResponse>(create);
  static GetAccountsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Account> get accounts => $_getList(0);
}

class Transfer extends $pb.GeneratedMessage {
  factory Transfer({
    $core.String? id,
    $core.String? debitAccountId,
    $core.String? creditAccountId,
    $fixnum.Int64? amount,
    $core.int? code,
    $fixnum.Int64? timeout,
    $core.int? ledger,
    $core.String? pendingId,
    $core.bool? pending,
  }) {
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (debitAccountId != null) {
      $result.debitAccountId = debitAccountId;
    }
    if (creditAccountId != null) {
      $result.creditAccountId = creditAccountId;
    }
    if (amount != null) {
      $result.amount = amount;
    }
    if (code != null) {
      $result.code = code;
    }
    if (timeout != null) {
      $result.timeout = timeout;
    }
    if (ledger != null) {
      $result.ledger = ledger;
    }
    if (pendingId != null) {
      $result.pendingId = pendingId;
    }
    if (pending != null) {
      $result.pending = pending;
    }
    return $result;
  }
  Transfer._() : super();
  factory Transfer.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Transfer.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Transfer', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'debitAccountId', protoName: 'debitAccountId')
    ..aOS(3, _omitFieldNames ? '' : 'creditAccountId', protoName: 'creditAccountId')
    ..a<$fixnum.Int64>(4, _omitFieldNames ? '' : 'amount', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$core.int>(5, _omitFieldNames ? '' : 'code', $pb.PbFieldType.OU3)
    ..a<$fixnum.Int64>(7, _omitFieldNames ? '' : 'timeout', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$core.int>(8, _omitFieldNames ? '' : 'ledger', $pb.PbFieldType.OU3)
    ..aOS(9, _omitFieldNames ? '' : 'pendingId', protoName: 'pendingId')
    ..aOB(10, _omitFieldNames ? '' : 'pending')
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
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get debitAccountId => $_getSZ(1);
  @$pb.TagNumber(2)
  set debitAccountId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDebitAccountId() => $_has(1);
  @$pb.TagNumber(2)
  void clearDebitAccountId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get creditAccountId => $_getSZ(2);
  @$pb.TagNumber(3)
  set creditAccountId($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCreditAccountId() => $_has(2);
  @$pb.TagNumber(3)
  void clearCreditAccountId() => clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get amount => $_getI64(3);
  @$pb.TagNumber(4)
  set amount($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasAmount() => $_has(3);
  @$pb.TagNumber(4)
  void clearAmount() => clearField(4);

  @$pb.TagNumber(5)
  $core.int get code => $_getIZ(4);
  @$pb.TagNumber(5)
  set code($core.int v) { $_setUnsignedInt32(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasCode() => $_has(4);
  @$pb.TagNumber(5)
  void clearCode() => clearField(5);

  @$pb.TagNumber(7)
  $fixnum.Int64 get timeout => $_getI64(5);
  @$pb.TagNumber(7)
  set timeout($fixnum.Int64 v) { $_setInt64(5, v); }
  @$pb.TagNumber(7)
  $core.bool hasTimeout() => $_has(5);
  @$pb.TagNumber(7)
  void clearTimeout() => clearField(7);

  @$pb.TagNumber(8)
  $core.int get ledger => $_getIZ(6);
  @$pb.TagNumber(8)
  set ledger($core.int v) { $_setUnsignedInt32(6, v); }
  @$pb.TagNumber(8)
  $core.bool hasLedger() => $_has(6);
  @$pb.TagNumber(8)
  void clearLedger() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get pendingId => $_getSZ(7);
  @$pb.TagNumber(9)
  set pendingId($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(9)
  $core.bool hasPendingId() => $_has(7);
  @$pb.TagNumber(9)
  void clearPendingId() => clearField(9);

  @$pb.TagNumber(10)
  $core.bool get pending => $_getBF(8);
  @$pb.TagNumber(10)
  set pending($core.bool v) { $_setBool(8, v); }
  @$pb.TagNumber(10)
  $core.bool hasPending() => $_has(8);
  @$pb.TagNumber(10)
  void clearPending() => clearField(10);
}

class CreateTransfersRequest extends $pb.GeneratedMessage {
  factory CreateTransfersRequest({
    $core.Iterable<Transfer>? transfers,
  }) {
    final $result = create();
    if (transfers != null) {
      $result.transfers.addAll(transfers);
    }
    return $result;
  }
  CreateTransfersRequest._() : super();
  factory CreateTransfersRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateTransfersRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateTransfersRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<Transfer>(1, _omitFieldNames ? '' : 'transfers', $pb.PbFieldType.PM, subBuilder: Transfer.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateTransfersRequest clone() => CreateTransfersRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateTransfersRequest copyWith(void Function(CreateTransfersRequest) updates) => super.copyWith((message) => updates(message as CreateTransfersRequest)) as CreateTransfersRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateTransfersRequest create() => CreateTransfersRequest._();
  CreateTransfersRequest createEmptyInstance() => create();
  static $pb.PbList<CreateTransfersRequest> createRepeated() => $pb.PbList<CreateTransfersRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateTransfersRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateTransfersRequest>(create);
  static CreateTransfersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Transfer> get transfers => $_getList(0);
}

class EventError extends $pb.GeneratedMessage {
  factory EventError({
    $core.int? index,
    $core.int? code,
  }) {
    final $result = create();
    if (index != null) {
      $result.index = index;
    }
    if (code != null) {
      $result.code = code;
    }
    return $result;
  }
  EventError._() : super();
  factory EventError.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory EventError.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'EventError', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'index', $pb.PbFieldType.OU3)
    ..a<$core.int>(2, _omitFieldNames ? '' : 'code', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  EventError clone() => EventError()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  EventError copyWith(void Function(EventError) updates) => super.copyWith((message) => updates(message as EventError)) as EventError;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EventError create() => EventError._();
  EventError createEmptyInstance() => create();
  static $pb.PbList<EventError> createRepeated() => $pb.PbList<EventError>();
  @$core.pragma('dart2js:noInline')
  static EventError getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EventError>(create);
  static EventError? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get index => $_getIZ(0);
  @$pb.TagNumber(1)
  set index($core.int v) { $_setUnsignedInt32(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasIndex() => $_has(0);
  @$pb.TagNumber(1)
  void clearIndex() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get code => $_getIZ(1);
  @$pb.TagNumber(2)
  set code($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => clearField(2);
}

class CreateTransfersResponse extends $pb.GeneratedMessage {
  factory CreateTransfersResponse({
    $core.Iterable<EventError>? errors,
  }) {
    final $result = create();
    if (errors != null) {
      $result.errors.addAll(errors);
    }
    return $result;
  }
  CreateTransfersResponse._() : super();
  factory CreateTransfersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateTransfersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateTransfersResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<EventError>(1, _omitFieldNames ? '' : 'errors', $pb.PbFieldType.PM, subBuilder: EventError.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateTransfersResponse clone() => CreateTransfersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateTransfersResponse copyWith(void Function(CreateTransfersResponse) updates) => super.copyWith((message) => updates(message as CreateTransfersResponse)) as CreateTransfersResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateTransfersResponse create() => CreateTransfersResponse._();
  CreateTransfersResponse createEmptyInstance() => create();
  static $pb.PbList<CreateTransfersResponse> createRepeated() => $pb.PbList<CreateTransfersResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateTransfersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateTransfersResponse>(create);
  static CreateTransfersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<EventError> get errors => $_getList(0);
}

class GetTransfersRequest extends $pb.GeneratedMessage {
  factory GetTransfersRequest({
    $core.Iterable<$core.String>? ids,
  }) {
    final $result = create();
    if (ids != null) {
      $result.ids.addAll(ids);
    }
    return $result;
  }
  GetTransfersRequest._() : super();
  factory GetTransfersRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetTransfersRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetTransfersRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'ids')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetTransfersRequest clone() => GetTransfersRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetTransfersRequest copyWith(void Function(GetTransfersRequest) updates) => super.copyWith((message) => updates(message as GetTransfersRequest)) as GetTransfersRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTransfersRequest create() => GetTransfersRequest._();
  GetTransfersRequest createEmptyInstance() => create();
  static $pb.PbList<GetTransfersRequest> createRepeated() => $pb.PbList<GetTransfersRequest>();
  @$core.pragma('dart2js:noInline')
  static GetTransfersRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTransfersRequest>(create);
  static GetTransfersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get ids => $_getList(0);
}

class GetTransfersResponse extends $pb.GeneratedMessage {
  factory GetTransfersResponse({
    $core.Iterable<Transfer>? transfers,
  }) {
    final $result = create();
    if (transfers != null) {
      $result.transfers.addAll(transfers);
    }
    return $result;
  }
  GetTransfersResponse._() : super();
  factory GetTransfersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GetTransfersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GetTransfersResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<Transfer>(1, _omitFieldNames ? '' : 'transfers', $pb.PbFieldType.PM, subBuilder: Transfer.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GetTransfersResponse clone() => GetTransfersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GetTransfersResponse copyWith(void Function(GetTransfersResponse) updates) => super.copyWith((message) => updates(message as GetTransfersResponse)) as GetTransfersResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTransfersResponse create() => GetTransfersResponse._();
  GetTransfersResponse createEmptyInstance() => create();
  static $pb.PbList<GetTransfersResponse> createRepeated() => $pb.PbList<GetTransfersResponse>();
  @$core.pragma('dart2js:noInline')
  static GetTransfersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetTransfersResponse>(create);
  static GetTransfersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Transfer> get transfers => $_getList(0);
}

class PostTransfersRequest extends $pb.GeneratedMessage {
  factory PostTransfersRequest({
    $core.Iterable<$core.String>? transferIds,
  }) {
    final $result = create();
    if (transferIds != null) {
      $result.transferIds.addAll(transferIds);
    }
    return $result;
  }
  PostTransfersRequest._() : super();
  factory PostTransfersRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PostTransfersRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PostTransfersRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'transferIds', protoName: 'transferIds')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PostTransfersRequest clone() => PostTransfersRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PostTransfersRequest copyWith(void Function(PostTransfersRequest) updates) => super.copyWith((message) => updates(message as PostTransfersRequest)) as PostTransfersRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PostTransfersRequest create() => PostTransfersRequest._();
  PostTransfersRequest createEmptyInstance() => create();
  static $pb.PbList<PostTransfersRequest> createRepeated() => $pb.PbList<PostTransfersRequest>();
  @$core.pragma('dart2js:noInline')
  static PostTransfersRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PostTransfersRequest>(create);
  static PostTransfersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get transferIds => $_getList(0);
}

class PostTransfersResponse extends $pb.GeneratedMessage {
  factory PostTransfersResponse({
    $core.Iterable<EventError>? errors,
  }) {
    final $result = create();
    if (errors != null) {
      $result.errors.addAll(errors);
    }
    return $result;
  }
  PostTransfersResponse._() : super();
  factory PostTransfersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PostTransfersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PostTransfersResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<EventError>(1, _omitFieldNames ? '' : 'errors', $pb.PbFieldType.PM, subBuilder: EventError.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PostTransfersResponse clone() => PostTransfersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PostTransfersResponse copyWith(void Function(PostTransfersResponse) updates) => super.copyWith((message) => updates(message as PostTransfersResponse)) as PostTransfersResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PostTransfersResponse create() => PostTransfersResponse._();
  PostTransfersResponse createEmptyInstance() => create();
  static $pb.PbList<PostTransfersResponse> createRepeated() => $pb.PbList<PostTransfersResponse>();
  @$core.pragma('dart2js:noInline')
  static PostTransfersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PostTransfersResponse>(create);
  static PostTransfersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<EventError> get errors => $_getList(0);
}

class VoidTransfersRequest extends $pb.GeneratedMessage {
  factory VoidTransfersRequest({
    $core.Iterable<$core.String>? transferIds,
  }) {
    final $result = create();
    if (transferIds != null) {
      $result.transferIds.addAll(transferIds);
    }
    return $result;
  }
  VoidTransfersRequest._() : super();
  factory VoidTransfersRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VoidTransfersRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'VoidTransfersRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pPS(1, _omitFieldNames ? '' : 'transferIds', protoName: 'transferIds')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VoidTransfersRequest clone() => VoidTransfersRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VoidTransfersRequest copyWith(void Function(VoidTransfersRequest) updates) => super.copyWith((message) => updates(message as VoidTransfersRequest)) as VoidTransfersRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VoidTransfersRequest create() => VoidTransfersRequest._();
  VoidTransfersRequest createEmptyInstance() => create();
  static $pb.PbList<VoidTransfersRequest> createRepeated() => $pb.PbList<VoidTransfersRequest>();
  @$core.pragma('dart2js:noInline')
  static VoidTransfersRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VoidTransfersRequest>(create);
  static VoidTransfersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get transferIds => $_getList(0);
}

class VoidTransfersResponse extends $pb.GeneratedMessage {
  factory VoidTransfersResponse({
    $core.Iterable<EventError>? errors,
  }) {
    final $result = create();
    if (errors != null) {
      $result.errors.addAll(errors);
    }
    return $result;
  }
  VoidTransfersResponse._() : super();
  factory VoidTransfersResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory VoidTransfersResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'VoidTransfersResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pacioli.v1'), createEmptyInstance: create)
    ..pc<EventError>(1, _omitFieldNames ? '' : 'errors', $pb.PbFieldType.PM, subBuilder: EventError.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  VoidTransfersResponse clone() => VoidTransfersResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  VoidTransfersResponse copyWith(void Function(VoidTransfersResponse) updates) => super.copyWith((message) => updates(message as VoidTransfersResponse)) as VoidTransfersResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VoidTransfersResponse create() => VoidTransfersResponse._();
  VoidTransfersResponse createEmptyInstance() => create();
  static $pb.PbList<VoidTransfersResponse> createRepeated() => $pb.PbList<VoidTransfersResponse>();
  @$core.pragma('dart2js:noInline')
  static VoidTransfersResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VoidTransfersResponse>(create);
  static VoidTransfersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<EventError> get errors => $_getList(0);
}


const _omitFieldNames = $core.bool.fromEnvironment('protobuf.omit_field_names');
const _omitMessageNames = $core.bool.fromEnvironment('protobuf.omit_message_names');
