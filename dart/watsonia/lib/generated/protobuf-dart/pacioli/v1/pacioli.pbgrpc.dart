//
//  Generated code. Do not modify.
//  source: pacioli/v1/pacioli.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'pacioli.pb.dart' as $3;

export 'pacioli.pb.dart';

@$pb.GrpcServiceName('pacioli.v1.PacioliService')
class PacioliServiceClient extends $grpc.Client {
  static final _$configureLedgers = $grpc.ClientMethod<$3.ConfigureLedgersRequest, $3.ConfigureLedgersResponse>(
      '/pacioli.v1.PacioliService/ConfigureLedgers',
      ($3.ConfigureLedgersRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.ConfigureLedgersResponse.fromBuffer(value));
  static final _$getLedgers = $grpc.ClientMethod<$3.GetLedgersRequest, $3.GetLedgersResponse>(
      '/pacioli.v1.PacioliService/GetLedgers',
      ($3.GetLedgersRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.GetLedgersResponse.fromBuffer(value));
  static final _$configureAccounts = $grpc.ClientMethod<$3.ConfigureAccountsRequest, $3.ConfigureAccountsResponse>(
      '/pacioli.v1.PacioliService/ConfigureAccounts',
      ($3.ConfigureAccountsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.ConfigureAccountsResponse.fromBuffer(value));
  static final _$getAccounts = $grpc.ClientMethod<$3.GetAccountsRequest, $3.GetAccountsResponse>(
      '/pacioli.v1.PacioliService/GetAccounts',
      ($3.GetAccountsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.GetAccountsResponse.fromBuffer(value));
  static final _$createTransfers = $grpc.ClientMethod<$3.CreateTransfersRequest, $3.CreateTransfersResponse>(
      '/pacioli.v1.PacioliService/CreateTransfers',
      ($3.CreateTransfersRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.CreateTransfersResponse.fromBuffer(value));
  static final _$getTransfers = $grpc.ClientMethod<$3.GetTransfersRequest, $3.GetTransfersResponse>(
      '/pacioli.v1.PacioliService/GetTransfers',
      ($3.GetTransfersRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.GetTransfersResponse.fromBuffer(value));
  static final _$postTransfers = $grpc.ClientMethod<$3.PostTransfersRequest, $3.PostTransfersResponse>(
      '/pacioli.v1.PacioliService/PostTransfers',
      ($3.PostTransfersRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.PostTransfersResponse.fromBuffer(value));
  static final _$voidTransfers = $grpc.ClientMethod<$3.VoidTransfersRequest, $3.VoidTransfersResponse>(
      '/pacioli.v1.PacioliService/VoidTransfers',
      ($3.VoidTransfersRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $3.VoidTransfersResponse.fromBuffer(value));

  PacioliServiceClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options,
        interceptors: interceptors);

  $grpc.ResponseFuture<$3.ConfigureLedgersResponse> configureLedgers($3.ConfigureLedgersRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$configureLedgers, request, options: options);
  }

  $grpc.ResponseFuture<$3.GetLedgersResponse> getLedgers($3.GetLedgersRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLedgers, request, options: options);
  }

  $grpc.ResponseFuture<$3.ConfigureAccountsResponse> configureAccounts($3.ConfigureAccountsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$configureAccounts, request, options: options);
  }

  $grpc.ResponseFuture<$3.GetAccountsResponse> getAccounts($3.GetAccountsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getAccounts, request, options: options);
  }

  $grpc.ResponseFuture<$3.CreateTransfersResponse> createTransfers($3.CreateTransfersRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createTransfers, request, options: options);
  }

  $grpc.ResponseFuture<$3.GetTransfersResponse> getTransfers($3.GetTransfersRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getTransfers, request, options: options);
  }

  $grpc.ResponseFuture<$3.PostTransfersResponse> postTransfers($3.PostTransfersRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$postTransfers, request, options: options);
  }

  $grpc.ResponseFuture<$3.VoidTransfersResponse> voidTransfers($3.VoidTransfersRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$voidTransfers, request, options: options);
  }
}

@$pb.GrpcServiceName('pacioli.v1.PacioliService')
abstract class PacioliServiceBase extends $grpc.Service {
  $core.String get $name => 'pacioli.v1.PacioliService';

  PacioliServiceBase() {
    $addMethod($grpc.ServiceMethod<$3.ConfigureLedgersRequest, $3.ConfigureLedgersResponse>(
        'ConfigureLedgers',
        configureLedgers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.ConfigureLedgersRequest.fromBuffer(value),
        ($3.ConfigureLedgersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.GetLedgersRequest, $3.GetLedgersResponse>(
        'GetLedgers',
        getLedgers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.GetLedgersRequest.fromBuffer(value),
        ($3.GetLedgersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.ConfigureAccountsRequest, $3.ConfigureAccountsResponse>(
        'ConfigureAccounts',
        configureAccounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.ConfigureAccountsRequest.fromBuffer(value),
        ($3.ConfigureAccountsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.GetAccountsRequest, $3.GetAccountsResponse>(
        'GetAccounts',
        getAccounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.GetAccountsRequest.fromBuffer(value),
        ($3.GetAccountsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.CreateTransfersRequest, $3.CreateTransfersResponse>(
        'CreateTransfers',
        createTransfers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.CreateTransfersRequest.fromBuffer(value),
        ($3.CreateTransfersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.GetTransfersRequest, $3.GetTransfersResponse>(
        'GetTransfers',
        getTransfers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.GetTransfersRequest.fromBuffer(value),
        ($3.GetTransfersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.PostTransfersRequest, $3.PostTransfersResponse>(
        'PostTransfers',
        postTransfers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.PostTransfersRequest.fromBuffer(value),
        ($3.PostTransfersResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$3.VoidTransfersRequest, $3.VoidTransfersResponse>(
        'VoidTransfers',
        voidTransfers_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $3.VoidTransfersRequest.fromBuffer(value),
        ($3.VoidTransfersResponse value) => value.writeToBuffer()));
  }

  $async.Future<$3.ConfigureLedgersResponse> configureLedgers_Pre($grpc.ServiceCall call, $async.Future<$3.ConfigureLedgersRequest> request) async {
    return configureLedgers(call, await request);
  }

  $async.Future<$3.GetLedgersResponse> getLedgers_Pre($grpc.ServiceCall call, $async.Future<$3.GetLedgersRequest> request) async {
    return getLedgers(call, await request);
  }

  $async.Future<$3.ConfigureAccountsResponse> configureAccounts_Pre($grpc.ServiceCall call, $async.Future<$3.ConfigureAccountsRequest> request) async {
    return configureAccounts(call, await request);
  }

  $async.Future<$3.GetAccountsResponse> getAccounts_Pre($grpc.ServiceCall call, $async.Future<$3.GetAccountsRequest> request) async {
    return getAccounts(call, await request);
  }

  $async.Future<$3.CreateTransfersResponse> createTransfers_Pre($grpc.ServiceCall call, $async.Future<$3.CreateTransfersRequest> request) async {
    return createTransfers(call, await request);
  }

  $async.Future<$3.GetTransfersResponse> getTransfers_Pre($grpc.ServiceCall call, $async.Future<$3.GetTransfersRequest> request) async {
    return getTransfers(call, await request);
  }

  $async.Future<$3.PostTransfersResponse> postTransfers_Pre($grpc.ServiceCall call, $async.Future<$3.PostTransfersRequest> request) async {
    return postTransfers(call, await request);
  }

  $async.Future<$3.VoidTransfersResponse> voidTransfers_Pre($grpc.ServiceCall call, $async.Future<$3.VoidTransfersRequest> request) async {
    return voidTransfers(call, await request);
  }

  $async.Future<$3.ConfigureLedgersResponse> configureLedgers($grpc.ServiceCall call, $3.ConfigureLedgersRequest request);
  $async.Future<$3.GetLedgersResponse> getLedgers($grpc.ServiceCall call, $3.GetLedgersRequest request);
  $async.Future<$3.ConfigureAccountsResponse> configureAccounts($grpc.ServiceCall call, $3.ConfigureAccountsRequest request);
  $async.Future<$3.GetAccountsResponse> getAccounts($grpc.ServiceCall call, $3.GetAccountsRequest request);
  $async.Future<$3.CreateTransfersResponse> createTransfers($grpc.ServiceCall call, $3.CreateTransfersRequest request);
  $async.Future<$3.GetTransfersResponse> getTransfers($grpc.ServiceCall call, $3.GetTransfersRequest request);
  $async.Future<$3.PostTransfersResponse> postTransfers($grpc.ServiceCall call, $3.PostTransfersRequest request);
  $async.Future<$3.VoidTransfersResponse> voidTransfers($grpc.ServiceCall call, $3.VoidTransfersRequest request);
}
