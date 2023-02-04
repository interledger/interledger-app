//
//  Generated code. Do not modify.
//  source: backend/admin/v1/backend.proto
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

import '../../../google/protobuf/empty.pb.dart' as $0;
import 'backend.pb.dart' as $1;

export 'backend.pb.dart';

@$pb.GrpcServiceName('backend.admin.v1.Backend')
class BackendClient extends $grpc.Client {
  static final _$listWaitlistSignups = $grpc.ClientMethod<$0.Empty, $1.ListWaitlistSignupsResponse>(
      '/backend.admin.v1.Backend/ListWaitlistSignups',
      ($0.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListWaitlistSignupsResponse.fromBuffer(value));
  static final _$allowWaitlistSignup = $grpc.ClientMethod<$1.AllowWaitlistSignupRequest, $1.Empty>(
      '/backend.admin.v1.Backend/AllowWaitlistSignup',
      ($1.AllowWaitlistSignupRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.Empty.fromBuffer(value));
  static final _$listWallets = $grpc.ClientMethod<$1.PaginationRequest, $1.ListWalletsResponse>(
      '/backend.admin.v1.Backend/ListWallets',
      ($1.PaginationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListWalletsResponse.fromBuffer(value));
  static final _$getWalletDetails = $grpc.ClientMethod<$1.GetWalletDetailsRequest, $1.WalletDetails>(
      '/backend.admin.v1.Backend/GetWalletDetails',
      ($1.GetWalletDetailsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.WalletDetails.fromBuffer(value));
  static final _$listTransactions = $grpc.ClientMethod<$1.ListTransactionsRequest, $1.ListTransactionsResponse>(
      '/backend.admin.v1.Backend/ListTransactions',
      ($1.ListTransactionsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListTransactionsResponse.fromBuffer(value));
  static final _$getTransactionDetails = $grpc.ClientMethod<$1.GetTransactionDetailsRequest, $1.GetTransactionDetailsResponse>(
      '/backend.admin.v1.Backend/GetTransactionDetails',
      ($1.GetTransactionDetailsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.GetTransactionDetailsResponse.fromBuffer(value));
  static final _$listLinkedAccounts = $grpc.ClientMethod<$1.ListLinkedAccountsRequest, $1.ListLinkedAccountsResponse>(
      '/backend.admin.v1.Backend/ListLinkedAccounts',
      ($1.ListLinkedAccountsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListLinkedAccountsResponse.fromBuffer(value));
  static final _$listAudit = $grpc.ClientMethod<$1.ListAuditRequest, $1.ListAuditResponse>(
      '/backend.admin.v1.Backend/ListAudit',
      ($1.ListAuditRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListAuditResponse.fromBuffer(value));
  static final _$getWalletFeatures = $grpc.ClientMethod<$1.GetWalletFeaturesRequest, $1.Features>(
      '/backend.admin.v1.Backend/GetWalletFeatures',
      ($1.GetWalletFeaturesRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.Features.fromBuffer(value));
  static final _$setWalletFeatures = $grpc.ClientMethod<$1.Features, $1.Features>(
      '/backend.admin.v1.Backend/SetWalletFeatures',
      ($1.Features value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.Features.fromBuffer(value));
  static final _$listIncompleteLinkedAccountReviews = $grpc.ClientMethod<$1.PaginationRequest, $1.LinkedAccountReviews>(
      '/backend.admin.v1.Backend/ListIncompleteLinkedAccountReviews',
      ($1.PaginationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.LinkedAccountReviews.fromBuffer(value));
  static final _$getLinkedAccountReview = $grpc.ClientMethod<$1.GetLinkedAccountReviewRequest, $1.LinkedAccountReview>(
      '/backend.admin.v1.Backend/GetLinkedAccountReview',
      ($1.GetLinkedAccountReviewRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.LinkedAccountReview.fromBuffer(value));
  static final _$completeLinkedAccountReview = $grpc.ClientMethod<$1.CompleteLinkedAccountReviewRequest, $1.LinkedAccountReview>(
      '/backend.admin.v1.Backend/CompleteLinkedAccountReview',
      ($1.CompleteLinkedAccountReviewRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.LinkedAccountReview.fromBuffer(value));
  static final _$getLinkedAccount = $grpc.ClientMethod<$1.GetLinkedAccountRequest, $1.LinkedAccount>(
      '/backend.admin.v1.Backend/GetLinkedAccount',
      ($1.GetLinkedAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.LinkedAccount.fromBuffer(value));
  static final _$listFormSubmissionCounts = $grpc.ClientMethod<$1.PaginationRequest, $1.ListFormSubmissionCountsResponse>(
      '/backend.admin.v1.Backend/ListFormSubmissionCounts',
      ($1.PaginationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListFormSubmissionCountsResponse.fromBuffer(value));
  static final _$exportFormSubmissions = $grpc.ClientMethod<$1.ExportFormSubmissionsRequest, $1.ExportFormSubmissionsResponse>(
      '/backend.admin.v1.Backend/ExportFormSubmissions',
      ($1.ExportFormSubmissionsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ExportFormSubmissionsResponse.fromBuffer(value));
  static final _$listFormSubmissions = $grpc.ClientMethod<$1.ListFormSubmissionsRequest, $1.ListFormSubmissionsResponse>(
      '/backend.admin.v1.Backend/ListFormSubmissions',
      ($1.ListFormSubmissionsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListFormSubmissionsResponse.fromBuffer(value));
  static final _$getFormSubmissionDetails = $grpc.ClientMethod<$1.GetFormSubmissionDetailsRequest, $1.FormSubmissionDetails>(
      '/backend.admin.v1.Backend/GetFormSubmissionDetails',
      ($1.GetFormSubmissionDetailsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.FormSubmissionDetails.fromBuffer(value));
  static final _$listExternalApiCalls = $grpc.ClientMethod<$1.ListExternalApiCallsRequest, $1.ListExternalApiCallsResponse>(
      '/backend.admin.v1.Backend/ListExternalApiCalls',
      ($1.ListExternalApiCallsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListExternalApiCallsResponse.fromBuffer(value));
  static final _$listPaymentsAwaitingSignal = $grpc.ClientMethod<$0.Empty, $1.ListPaymentsAwaitingSignalResponse>(
      '/backend.admin.v1.Backend/ListPaymentsAwaitingSignal',
      ($0.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListPaymentsAwaitingSignalResponse.fromBuffer(value));
  static final _$setWalletXagoBalanceEnabled = $grpc.ClientMethod<$1.SetWalletXagoBalanceEnabledRequest, $1.Empty>(
      '/backend.admin.v1.Backend/SetWalletXagoBalanceEnabled',
      ($1.SetWalletXagoBalanceEnabledRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.Empty.fromBuffer(value));
  static final _$getWalletXagoBalance = $grpc.ClientMethod<$1.GetWalletXagoBalanceRequest, $1.GetWalletXagoBalanceResponse>(
      '/backend.admin.v1.Backend/GetWalletXagoBalance',
      ($1.GetWalletXagoBalanceRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.GetWalletXagoBalanceResponse.fromBuffer(value));
  static final _$setWalletCountry = $grpc.ClientMethod<$1.SetWalletCountryRequest, $1.Empty>(
      '/backend.admin.v1.Backend/SetWalletCountry',
      ($1.SetWalletCountryRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.Empty.fromBuffer(value));
  static final _$listCountries = $grpc.ClientMethod<$1.Empty, $1.ListCountriesResponse>(
      '/backend.admin.v1.Backend/ListCountries',
      ($1.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $1.ListCountriesResponse.fromBuffer(value));

  BackendClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options,
        interceptors: interceptors);

  $grpc.ResponseFuture<$1.ListWaitlistSignupsResponse> listWaitlistSignups($0.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listWaitlistSignups, request, options: options);
  }

  $grpc.ResponseFuture<$1.Empty> allowWaitlistSignup($1.AllowWaitlistSignupRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$allowWaitlistSignup, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListWalletsResponse> listWallets($1.PaginationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listWallets, request, options: options);
  }

  $grpc.ResponseFuture<$1.WalletDetails> getWalletDetails($1.GetWalletDetailsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWalletDetails, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListTransactionsResponse> listTransactions($1.ListTransactionsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listTransactions, request, options: options);
  }

  $grpc.ResponseFuture<$1.GetTransactionDetailsResponse> getTransactionDetails($1.GetTransactionDetailsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getTransactionDetails, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListLinkedAccountsResponse> listLinkedAccounts($1.ListLinkedAccountsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listLinkedAccounts, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListAuditResponse> listAudit($1.ListAuditRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listAudit, request, options: options);
  }

  $grpc.ResponseFuture<$1.Features> getWalletFeatures($1.GetWalletFeaturesRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWalletFeatures, request, options: options);
  }

  $grpc.ResponseFuture<$1.Features> setWalletFeatures($1.Features request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setWalletFeatures, request, options: options);
  }

  $grpc.ResponseFuture<$1.LinkedAccountReviews> listIncompleteLinkedAccountReviews($1.PaginationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listIncompleteLinkedAccountReviews, request, options: options);
  }

  $grpc.ResponseFuture<$1.LinkedAccountReview> getLinkedAccountReview($1.GetLinkedAccountReviewRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLinkedAccountReview, request, options: options);
  }

  $grpc.ResponseFuture<$1.LinkedAccountReview> completeLinkedAccountReview($1.CompleteLinkedAccountReviewRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$completeLinkedAccountReview, request, options: options);
  }

  $grpc.ResponseFuture<$1.LinkedAccount> getLinkedAccount($1.GetLinkedAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListFormSubmissionCountsResponse> listFormSubmissionCounts($1.PaginationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listFormSubmissionCounts, request, options: options);
  }

  $grpc.ResponseStream<$1.ExportFormSubmissionsResponse> exportFormSubmissions($1.ExportFormSubmissionsRequest request, {$grpc.CallOptions? options}) {
    return $createStreamingCall(_$exportFormSubmissions, $async.Stream.fromIterable([request]), options: options);
  }

  $grpc.ResponseFuture<$1.ListFormSubmissionsResponse> listFormSubmissions($1.ListFormSubmissionsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listFormSubmissions, request, options: options);
  }

  $grpc.ResponseFuture<$1.FormSubmissionDetails> getFormSubmissionDetails($1.GetFormSubmissionDetailsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getFormSubmissionDetails, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListExternalApiCallsResponse> listExternalApiCalls($1.ListExternalApiCallsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listExternalApiCalls, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListPaymentsAwaitingSignalResponse> listPaymentsAwaitingSignal($0.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listPaymentsAwaitingSignal, request, options: options);
  }

  $grpc.ResponseFuture<$1.Empty> setWalletXagoBalanceEnabled($1.SetWalletXagoBalanceEnabledRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setWalletXagoBalanceEnabled, request, options: options);
  }

  $grpc.ResponseFuture<$1.GetWalletXagoBalanceResponse> getWalletXagoBalance($1.GetWalletXagoBalanceRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWalletXagoBalance, request, options: options);
  }

  $grpc.ResponseFuture<$1.Empty> setWalletCountry($1.SetWalletCountryRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setWalletCountry, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListCountriesResponse> listCountries($1.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listCountries, request, options: options);
  }
}

@$pb.GrpcServiceName('backend.admin.v1.Backend')
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
        ($core.List<$core.int> value) => $1.AllowWaitlistSignupRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.PaginationRequest, $1.ListWalletsResponse>(
        'ListWallets',
        listWallets_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.PaginationRequest.fromBuffer(value),
        ($1.ListWalletsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetWalletDetailsRequest, $1.WalletDetails>(
        'GetWalletDetails',
        getWalletDetails_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.GetWalletDetailsRequest.fromBuffer(value),
        ($1.WalletDetails value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ListTransactionsRequest, $1.ListTransactionsResponse>(
        'ListTransactions',
        listTransactions_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ListTransactionsRequest.fromBuffer(value),
        ($1.ListTransactionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetTransactionDetailsRequest, $1.GetTransactionDetailsResponse>(
        'GetTransactionDetails',
        getTransactionDetails_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.GetTransactionDetailsRequest.fromBuffer(value),
        ($1.GetTransactionDetailsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ListLinkedAccountsRequest, $1.ListLinkedAccountsResponse>(
        'ListLinkedAccounts',
        listLinkedAccounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ListLinkedAccountsRequest.fromBuffer(value),
        ($1.ListLinkedAccountsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ListAuditRequest, $1.ListAuditResponse>(
        'ListAudit',
        listAudit_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ListAuditRequest.fromBuffer(value),
        ($1.ListAuditResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetWalletFeaturesRequest, $1.Features>(
        'GetWalletFeatures',
        getWalletFeatures_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.GetWalletFeaturesRequest.fromBuffer(value),
        ($1.Features value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Features, $1.Features>(
        'SetWalletFeatures',
        setWalletFeatures_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Features.fromBuffer(value),
        ($1.Features value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.PaginationRequest, $1.LinkedAccountReviews>(
        'ListIncompleteLinkedAccountReviews',
        listIncompleteLinkedAccountReviews_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.PaginationRequest.fromBuffer(value),
        ($1.LinkedAccountReviews value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetLinkedAccountReviewRequest, $1.LinkedAccountReview>(
        'GetLinkedAccountReview',
        getLinkedAccountReview_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.GetLinkedAccountReviewRequest.fromBuffer(value),
        ($1.LinkedAccountReview value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.CompleteLinkedAccountReviewRequest, $1.LinkedAccountReview>(
        'CompleteLinkedAccountReview',
        completeLinkedAccountReview_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.CompleteLinkedAccountReviewRequest.fromBuffer(value),
        ($1.LinkedAccountReview value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetLinkedAccountRequest, $1.LinkedAccount>(
        'GetLinkedAccount',
        getLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.GetLinkedAccountRequest.fromBuffer(value),
        ($1.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.PaginationRequest, $1.ListFormSubmissionCountsResponse>(
        'ListFormSubmissionCounts',
        listFormSubmissionCounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.PaginationRequest.fromBuffer(value),
        ($1.ListFormSubmissionCountsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ExportFormSubmissionsRequest, $1.ExportFormSubmissionsResponse>(
        'ExportFormSubmissions',
        exportFormSubmissions_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $1.ExportFormSubmissionsRequest.fromBuffer(value),
        ($1.ExportFormSubmissionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ListFormSubmissionsRequest, $1.ListFormSubmissionsResponse>(
        'ListFormSubmissions',
        listFormSubmissions_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ListFormSubmissionsRequest.fromBuffer(value),
        ($1.ListFormSubmissionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetFormSubmissionDetailsRequest, $1.FormSubmissionDetails>(
        'GetFormSubmissionDetails',
        getFormSubmissionDetails_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.GetFormSubmissionDetailsRequest.fromBuffer(value),
        ($1.FormSubmissionDetails value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ListExternalApiCallsRequest, $1.ListExternalApiCallsResponse>(
        'ListExternalApiCalls',
        listExternalApiCalls_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ListExternalApiCallsRequest.fromBuffer(value),
        ($1.ListExternalApiCallsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.ListPaymentsAwaitingSignalResponse>(
        'ListPaymentsAwaitingSignal',
        listPaymentsAwaitingSignal_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.ListPaymentsAwaitingSignalResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.SetWalletXagoBalanceEnabledRequest, $1.Empty>(
        'SetWalletXagoBalanceEnabled',
        setWalletXagoBalanceEnabled_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.SetWalletXagoBalanceEnabledRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.GetWalletXagoBalanceRequest, $1.GetWalletXagoBalanceResponse>(
        'GetWalletXagoBalance',
        getWalletXagoBalance_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.GetWalletXagoBalanceRequest.fromBuffer(value),
        ($1.GetWalletXagoBalanceResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.SetWalletCountryRequest, $1.Empty>(
        'SetWalletCountry',
        setWalletCountry_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.SetWalletCountryRequest.fromBuffer(value),
        ($1.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.Empty, $1.ListCountriesResponse>(
        'ListCountries',
        listCountries_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.Empty.fromBuffer(value),
        ($1.ListCountriesResponse value) => value.writeToBuffer()));
  }

  $async.Future<$1.ListWaitlistSignupsResponse> listWaitlistSignups_Pre($grpc.ServiceCall call, $async.Future<$0.Empty> request) async {
    return listWaitlistSignups(call, await request);
  }

  $async.Future<$1.Empty> allowWaitlistSignup_Pre($grpc.ServiceCall call, $async.Future<$1.AllowWaitlistSignupRequest> request) async {
    return allowWaitlistSignup(call, await request);
  }

  $async.Future<$1.ListWalletsResponse> listWallets_Pre($grpc.ServiceCall call, $async.Future<$1.PaginationRequest> request) async {
    return listWallets(call, await request);
  }

  $async.Future<$1.WalletDetails> getWalletDetails_Pre($grpc.ServiceCall call, $async.Future<$1.GetWalletDetailsRequest> request) async {
    return getWalletDetails(call, await request);
  }

  $async.Future<$1.ListTransactionsResponse> listTransactions_Pre($grpc.ServiceCall call, $async.Future<$1.ListTransactionsRequest> request) async {
    return listTransactions(call, await request);
  }

  $async.Future<$1.GetTransactionDetailsResponse> getTransactionDetails_Pre($grpc.ServiceCall call, $async.Future<$1.GetTransactionDetailsRequest> request) async {
    return getTransactionDetails(call, await request);
  }

  $async.Future<$1.ListLinkedAccountsResponse> listLinkedAccounts_Pre($grpc.ServiceCall call, $async.Future<$1.ListLinkedAccountsRequest> request) async {
    return listLinkedAccounts(call, await request);
  }

  $async.Future<$1.ListAuditResponse> listAudit_Pre($grpc.ServiceCall call, $async.Future<$1.ListAuditRequest> request) async {
    return listAudit(call, await request);
  }

  $async.Future<$1.Features> getWalletFeatures_Pre($grpc.ServiceCall call, $async.Future<$1.GetWalletFeaturesRequest> request) async {
    return getWalletFeatures(call, await request);
  }

  $async.Future<$1.Features> setWalletFeatures_Pre($grpc.ServiceCall call, $async.Future<$1.Features> request) async {
    return setWalletFeatures(call, await request);
  }

  $async.Future<$1.LinkedAccountReviews> listIncompleteLinkedAccountReviews_Pre($grpc.ServiceCall call, $async.Future<$1.PaginationRequest> request) async {
    return listIncompleteLinkedAccountReviews(call, await request);
  }

  $async.Future<$1.LinkedAccountReview> getLinkedAccountReview_Pre($grpc.ServiceCall call, $async.Future<$1.GetLinkedAccountReviewRequest> request) async {
    return getLinkedAccountReview(call, await request);
  }

  $async.Future<$1.LinkedAccountReview> completeLinkedAccountReview_Pre($grpc.ServiceCall call, $async.Future<$1.CompleteLinkedAccountReviewRequest> request) async {
    return completeLinkedAccountReview(call, await request);
  }

  $async.Future<$1.LinkedAccount> getLinkedAccount_Pre($grpc.ServiceCall call, $async.Future<$1.GetLinkedAccountRequest> request) async {
    return getLinkedAccount(call, await request);
  }

  $async.Future<$1.ListFormSubmissionCountsResponse> listFormSubmissionCounts_Pre($grpc.ServiceCall call, $async.Future<$1.PaginationRequest> request) async {
    return listFormSubmissionCounts(call, await request);
  }

  $async.Stream<$1.ExportFormSubmissionsResponse> exportFormSubmissions_Pre($grpc.ServiceCall call, $async.Future<$1.ExportFormSubmissionsRequest> request) async* {
    yield* exportFormSubmissions(call, await request);
  }

  $async.Future<$1.ListFormSubmissionsResponse> listFormSubmissions_Pre($grpc.ServiceCall call, $async.Future<$1.ListFormSubmissionsRequest> request) async {
    return listFormSubmissions(call, await request);
  }

  $async.Future<$1.FormSubmissionDetails> getFormSubmissionDetails_Pre($grpc.ServiceCall call, $async.Future<$1.GetFormSubmissionDetailsRequest> request) async {
    return getFormSubmissionDetails(call, await request);
  }

  $async.Future<$1.ListExternalApiCallsResponse> listExternalApiCalls_Pre($grpc.ServiceCall call, $async.Future<$1.ListExternalApiCallsRequest> request) async {
    return listExternalApiCalls(call, await request);
  }

  $async.Future<$1.ListPaymentsAwaitingSignalResponse> listPaymentsAwaitingSignal_Pre($grpc.ServiceCall call, $async.Future<$0.Empty> request) async {
    return listPaymentsAwaitingSignal(call, await request);
  }

  $async.Future<$1.Empty> setWalletXagoBalanceEnabled_Pre($grpc.ServiceCall call, $async.Future<$1.SetWalletXagoBalanceEnabledRequest> request) async {
    return setWalletXagoBalanceEnabled(call, await request);
  }

  $async.Future<$1.GetWalletXagoBalanceResponse> getWalletXagoBalance_Pre($grpc.ServiceCall call, $async.Future<$1.GetWalletXagoBalanceRequest> request) async {
    return getWalletXagoBalance(call, await request);
  }

  $async.Future<$1.Empty> setWalletCountry_Pre($grpc.ServiceCall call, $async.Future<$1.SetWalletCountryRequest> request) async {
    return setWalletCountry(call, await request);
  }

  $async.Future<$1.ListCountriesResponse> listCountries_Pre($grpc.ServiceCall call, $async.Future<$1.Empty> request) async {
    return listCountries(call, await request);
  }

  $async.Future<$1.ListWaitlistSignupsResponse> listWaitlistSignups($grpc.ServiceCall call, $0.Empty request);
  $async.Future<$1.Empty> allowWaitlistSignup($grpc.ServiceCall call, $1.AllowWaitlistSignupRequest request);
  $async.Future<$1.ListWalletsResponse> listWallets($grpc.ServiceCall call, $1.PaginationRequest request);
  $async.Future<$1.WalletDetails> getWalletDetails($grpc.ServiceCall call, $1.GetWalletDetailsRequest request);
  $async.Future<$1.ListTransactionsResponse> listTransactions($grpc.ServiceCall call, $1.ListTransactionsRequest request);
  $async.Future<$1.GetTransactionDetailsResponse> getTransactionDetails($grpc.ServiceCall call, $1.GetTransactionDetailsRequest request);
  $async.Future<$1.ListLinkedAccountsResponse> listLinkedAccounts($grpc.ServiceCall call, $1.ListLinkedAccountsRequest request);
  $async.Future<$1.ListAuditResponse> listAudit($grpc.ServiceCall call, $1.ListAuditRequest request);
  $async.Future<$1.Features> getWalletFeatures($grpc.ServiceCall call, $1.GetWalletFeaturesRequest request);
  $async.Future<$1.Features> setWalletFeatures($grpc.ServiceCall call, $1.Features request);
  $async.Future<$1.LinkedAccountReviews> listIncompleteLinkedAccountReviews($grpc.ServiceCall call, $1.PaginationRequest request);
  $async.Future<$1.LinkedAccountReview> getLinkedAccountReview($grpc.ServiceCall call, $1.GetLinkedAccountReviewRequest request);
  $async.Future<$1.LinkedAccountReview> completeLinkedAccountReview($grpc.ServiceCall call, $1.CompleteLinkedAccountReviewRequest request);
  $async.Future<$1.LinkedAccount> getLinkedAccount($grpc.ServiceCall call, $1.GetLinkedAccountRequest request);
  $async.Future<$1.ListFormSubmissionCountsResponse> listFormSubmissionCounts($grpc.ServiceCall call, $1.PaginationRequest request);
  $async.Stream<$1.ExportFormSubmissionsResponse> exportFormSubmissions($grpc.ServiceCall call, $1.ExportFormSubmissionsRequest request);
  $async.Future<$1.ListFormSubmissionsResponse> listFormSubmissions($grpc.ServiceCall call, $1.ListFormSubmissionsRequest request);
  $async.Future<$1.FormSubmissionDetails> getFormSubmissionDetails($grpc.ServiceCall call, $1.GetFormSubmissionDetailsRequest request);
  $async.Future<$1.ListExternalApiCallsResponse> listExternalApiCalls($grpc.ServiceCall call, $1.ListExternalApiCallsRequest request);
  $async.Future<$1.ListPaymentsAwaitingSignalResponse> listPaymentsAwaitingSignal($grpc.ServiceCall call, $0.Empty request);
  $async.Future<$1.Empty> setWalletXagoBalanceEnabled($grpc.ServiceCall call, $1.SetWalletXagoBalanceEnabledRequest request);
  $async.Future<$1.GetWalletXagoBalanceResponse> getWalletXagoBalance($grpc.ServiceCall call, $1.GetWalletXagoBalanceRequest request);
  $async.Future<$1.Empty> setWalletCountry($grpc.ServiceCall call, $1.SetWalletCountryRequest request);
  $async.Future<$1.ListCountriesResponse> listCountries($grpc.ServiceCall call, $1.Empty request);
}
