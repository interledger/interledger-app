///
//  Generated code. Do not modify.
//  source: backend/admin/v1/backend.proto
//
// @dart = 2.12
// ignore_for_file: annotate_overrides,camel_case_types,constant_identifier_names,directives_ordering,library_prefixes,non_constant_identifier_names,prefer_final_fields,return_of_invalid_type,unnecessary_const,unnecessary_import,unnecessary_this,unused_import,unused_shown_name

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import '../../../google/protobuf/empty.pb.dart' as $0;
import 'backend.pb.dart' as $1;
export 'backend.pb.dart';

class BackendClient extends $grpc.Client {
  static final _$listWaitlistSignups =
      $grpc.ClientMethod<$0.Empty, $1.ListWaitlistSignupsResponse>(
          '/backend.admin.v1.Backend/ListWaitlistSignups',
          ($0.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.ListWaitlistSignupsResponse.fromBuffer(value));
  static final _$allowWaitlistSignup =
      $grpc.ClientMethod<$1.AllowWaitlistSignupRequest, $1.Empty>(
          '/backend.admin.v1.Backend/AllowWaitlistSignup',
          ($1.AllowWaitlistSignupRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $1.Empty.fromBuffer(value));
  static final _$listWallets =
      $grpc.ClientMethod<$1.PaginationRequest, $1.ListWalletsResponse>(
          '/backend.admin.v1.Backend/ListWallets',
          ($1.PaginationRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $1.ListWalletsResponse.fromBuffer(value));
  static final _$getWalletDetails =
      $grpc.ClientMethod<$1.GetWalletDetailsRequest, $1.WalletDetails>(
          '/backend.admin.v1.Backend/GetWalletDetails',
          ($1.GetWalletDetailsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $1.WalletDetails.fromBuffer(value));
  static final _$emailWalletStatement =
      $grpc.ClientMethod<$1.EmailWalletStatementRequest, $1.Empty>(
          '/backend.admin.v1.Backend/EmailWalletStatement',
          ($1.EmailWalletStatementRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $1.Empty.fromBuffer(value));

  BackendClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$1.ListWaitlistSignupsResponse> listWaitlistSignups(
      $0.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listWaitlistSignups, request, options: options);
  }

  $grpc.ResponseFuture<$1.Empty> allowWaitlistSignup(
      $1.AllowWaitlistSignupRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$allowWaitlistSignup, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListWalletsResponse> listWallets(
      $1.PaginationRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listWallets, request, options: options);
  }

  $grpc.ResponseFuture<$1.WalletDetails> getWalletDetails(
      $1.GetWalletDetailsRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWalletDetails, request, options: options);
  }

  $grpc.ResponseFuture<$1.Empty> emailWalletStatement(
      $1.EmailWalletStatementRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$emailWalletStatement, request, options: options);
  }
}

abstract class BackendServiceBase extends $grpc.Service {
  $core.String get $name => 'backend.admin.v1.Backend';

  BackendServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.ListWaitlistSignupsResponse>(
        'ListWaitlistSignups',
        listWaitlistSignups_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.ListWaitlistSignupsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.AllowWaitlistSignupRequest, $1.Empty>(
        'AllowWaitlistSignup',
        allowWaitlistSignup_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.AllowWaitlistSignupRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.PaginationRequest, $1.ListWalletsResponse>(
            'ListWallets',
            listWallets_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.PaginationRequest.fromBuffer(value),
            ($1.ListWalletsResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.GetWalletDetailsRequest, $1.WalletDetails>(
            'GetWalletDetails',
            getWalletDetails_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.GetWalletDetailsRequest.fromBuffer(value),
            ($1.WalletDetails value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.EmailWalletStatementRequest, $1.Empty>(
        'EmailWalletStatement',
        emailWalletStatement_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.EmailWalletStatementRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
  }

  $async.Future<$1.ListWaitlistSignupsResponse> listWaitlistSignups_Pre(
      $grpc.ServiceCall call, $async.Future<$0.Empty> request) async {
    return listWaitlistSignups(call, await request);
  }

  $async.Future<$1.Empty> allowWaitlistSignup_Pre($grpc.ServiceCall call,
      $async.Future<$1.AllowWaitlistSignupRequest> request) async {
    return allowWaitlistSignup(call, await request);
  }

  $async.Future<$1.ListWalletsResponse> listWallets_Pre($grpc.ServiceCall call,
      $async.Future<$1.PaginationRequest> request) async {
    return listWallets(call, await request);
  }

  $async.Future<$1.WalletDetails> getWalletDetails_Pre($grpc.ServiceCall call,
      $async.Future<$1.GetWalletDetailsRequest> request) async {
    return getWalletDetails(call, await request);
  }

  $async.Future<$1.Empty> emailWalletStatement_Pre($grpc.ServiceCall call,
      $async.Future<$1.EmailWalletStatementRequest> request) async {
    return emailWalletStatement(call, await request);
  }

  $async.Future<$1.ListWaitlistSignupsResponse> listWaitlistSignups(
      $grpc.ServiceCall call, $0.Empty request);
  $async.Future<$1.Empty> allowWaitlistSignup(
      $grpc.ServiceCall call, $1.AllowWaitlistSignupRequest request);
  $async.Future<$1.ListWalletsResponse> listWallets(
      $grpc.ServiceCall call, $1.PaginationRequest request);
  $async.Future<$1.WalletDetails> getWalletDetails(
      $grpc.ServiceCall call, $1.GetWalletDetailsRequest request);
  $async.Future<$1.Empty> emailWalletStatement(
      $grpc.ServiceCall call, $1.EmailWalletStatementRequest request);
}
