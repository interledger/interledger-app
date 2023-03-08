///
//  Generated code. Do not modify.
//  source: backend/v1/backend.proto
//
// @dart = 2.12
// ignore_for_file: annotate_overrides,camel_case_types,constant_identifier_names,directives_ordering,library_prefixes,non_constant_identifier_names,prefer_final_fields,return_of_invalid_type,unnecessary_const,unnecessary_import,unnecessary_this,unused_import,unused_shown_name

import 'dart:async' as $async;

import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'backend.pb.dart' as $2;
export 'backend.pb.dart';

class OpenPaymentServiceClient extends $grpc.Client {
  static final _$createPaymentPointer =
      $grpc.ClientMethod<$2.CreatePaymentPointerRequest, $2.Empty>(
          '/backend.v1.OpenPaymentService/CreatePaymentPointer',
          ($2.CreatePaymentPointerRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getPaymentPointer =
      $grpc.ClientMethod<$2.GetPaymentPointerRequest, $2.PaymentPointer>(
          '/backend.v1.OpenPaymentService/GetPaymentPointer',
          ($2.GetPaymentPointerRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.PaymentPointer.fromBuffer(value));
  static final _$paymentPointerExists = $grpc.ClientMethod<
          $2.PaymentPointerExistsRequest, $2.PaymentPointerExistsResponse>(
      '/backend.v1.OpenPaymentService/PaymentPointerExists',
      ($2.PaymentPointerExistsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $2.PaymentPointerExistsResponse.fromBuffer(value));
  static final _$listWalletPaymentPointers =
      $grpc.ClientMethod<$2.Empty, $2.ListWalletPaymentPointersResponse>(
          '/backend.v1.OpenPaymentService/ListWalletPaymentPointers',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.ListWalletPaymentPointersResponse.fromBuffer(value));
  static final _$createQuote =
      $grpc.ClientMethod<$2.CreateQuoteRequest, $2.Quote>(
          '/backend.v1.OpenPaymentService/CreateQuote',
          ($2.CreateQuoteRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Quote.fromBuffer(value));
  static final _$lookupQuote =
      $grpc.ClientMethod<$2.LookupQuoteRequest, $2.Quote>(
          '/backend.v1.OpenPaymentService/LookupQuote',
          ($2.LookupQuoteRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Quote.fromBuffer(value));
  static final _$createIncomingPayment =
      $grpc.ClientMethod<$2.CreateIncomingPaymentRequest, $2.IncomingPayment>(
          '/backend.v1.OpenPaymentService/CreateIncomingPayment',
          ($2.CreateIncomingPaymentRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.IncomingPayment.fromBuffer(value));
  static final _$lookupIncomingPayment =
      $grpc.ClientMethod<$2.LookupIncomingPaymentRequest, $2.IncomingPayment>(
          '/backend.v1.OpenPaymentService/LookupIncomingPayment',
          ($2.LookupIncomingPaymentRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.IncomingPayment.fromBuffer(value));
  static final _$preCheckOutgoingPayment = $grpc.ClientMethod<
          $2.PreCheckOutgoingPaymentRequest,
          $2.PreCheckOutgoingPaymentResponse>(
      '/backend.v1.OpenPaymentService/PreCheckOutgoingPayment',
      ($2.PreCheckOutgoingPaymentRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $2.PreCheckOutgoingPaymentResponse.fromBuffer(value));
  static final _$createOutgoingPayment =
      $grpc.ClientMethod<$2.CreateOutgoingPaymentRequest, $2.OutgoingPayment>(
          '/backend.v1.OpenPaymentService/CreateOutgoingPayment',
          ($2.CreateOutgoingPaymentRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.OutgoingPayment.fromBuffer(value));
  static final _$lookupOutgoingPayment =
      $grpc.ClientMethod<$2.LookupOutgoingPaymentRequest, $2.OutgoingPayment>(
          '/backend.v1.OpenPaymentService/LookupOutgoingPayment',
          ($2.LookupOutgoingPaymentRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.OutgoingPayment.fromBuffer(value));
  static final _$canSendToPaymentPointer = $grpc.ClientMethod<
          $2.CanSendToPaymentPointerRequest,
          $2.CanSendToPaymentPointerResponse>(
      '/backend.v1.OpenPaymentService/CanSendToPaymentPointer',
      ($2.CanSendToPaymentPointerRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $2.CanSendToPaymentPointerResponse.fromBuffer(value));

  OpenPaymentServiceClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$2.Empty> createPaymentPointer(
      $2.CreatePaymentPointerRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createPaymentPointer, request, options: options);
  }

  $grpc.ResponseFuture<$2.PaymentPointer> getPaymentPointer(
      $2.GetPaymentPointerRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getPaymentPointer, request, options: options);
  }

  $grpc.ResponseFuture<$2.PaymentPointerExistsResponse> paymentPointerExists(
      $2.PaymentPointerExistsRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$paymentPointerExists, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListWalletPaymentPointersResponse>
      listWalletPaymentPointers($2.Empty request,
          {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listWalletPaymentPointers, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Quote> createQuote($2.CreateQuoteRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createQuote, request, options: options);
  }

  $grpc.ResponseFuture<$2.Quote> lookupQuote($2.LookupQuoteRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$lookupQuote, request, options: options);
  }

  $grpc.ResponseFuture<$2.IncomingPayment> createIncomingPayment(
      $2.CreateIncomingPaymentRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createIncomingPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.IncomingPayment> lookupIncomingPayment(
      $2.LookupIncomingPaymentRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$lookupIncomingPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.PreCheckOutgoingPaymentResponse>
      preCheckOutgoingPayment($2.PreCheckOutgoingPaymentRequest request,
          {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$preCheckOutgoingPayment, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.OutgoingPayment> createOutgoingPayment(
      $2.CreateOutgoingPaymentRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createOutgoingPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.OutgoingPayment> lookupOutgoingPayment(
      $2.LookupOutgoingPaymentRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$lookupOutgoingPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.CanSendToPaymentPointerResponse>
      canSendToPaymentPointer($2.CanSendToPaymentPointerRequest request,
          {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$canSendToPaymentPointer, request,
        options: options);
  }
}

abstract class OpenPaymentServiceBase extends $grpc.Service {
  $core.String get $name => 'backend.v1.OpenPaymentService';

  OpenPaymentServiceBase() {
    $addMethod($grpc.ServiceMethod<$2.CreatePaymentPointerRequest, $2.Empty>(
        'CreatePaymentPointer',
        createPaymentPointer_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreatePaymentPointerRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.GetPaymentPointerRequest, $2.PaymentPointer>(
            'GetPaymentPointer',
            getPaymentPointer_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.GetPaymentPointerRequest.fromBuffer(value),
            ($2.PaymentPointer value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.PaymentPointerExistsRequest,
            $2.PaymentPointerExistsResponse>(
        'PaymentPointerExists',
        paymentPointerExists_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.PaymentPointerExistsRequest.fromBuffer(value),
        ($2.PaymentPointerExistsResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.Empty, $2.ListWalletPaymentPointersResponse>(
            'ListWalletPaymentPointers',
            listWalletPaymentPointers_Pre,
            false,
            false,
            ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
            ($2.ListWalletPaymentPointersResponse value) =>
                value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateQuoteRequest, $2.Quote>(
        'CreateQuote',
        createQuote_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateQuoteRequest.fromBuffer(value),
        ($2.Quote value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.LookupQuoteRequest, $2.Quote>(
        'LookupQuote',
        lookupQuote_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.LookupQuoteRequest.fromBuffer(value),
        ($2.Quote value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateIncomingPaymentRequest,
            $2.IncomingPayment>(
        'CreateIncomingPayment',
        createIncomingPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateIncomingPaymentRequest.fromBuffer(value),
        ($2.IncomingPayment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.LookupIncomingPaymentRequest,
            $2.IncomingPayment>(
        'LookupIncomingPayment',
        lookupIncomingPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.LookupIncomingPaymentRequest.fromBuffer(value),
        ($2.IncomingPayment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.PreCheckOutgoingPaymentRequest,
            $2.PreCheckOutgoingPaymentResponse>(
        'PreCheckOutgoingPayment',
        preCheckOutgoingPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.PreCheckOutgoingPaymentRequest.fromBuffer(value),
        ($2.PreCheckOutgoingPaymentResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateOutgoingPaymentRequest,
            $2.OutgoingPayment>(
        'CreateOutgoingPayment',
        createOutgoingPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateOutgoingPaymentRequest.fromBuffer(value),
        ($2.OutgoingPayment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.LookupOutgoingPaymentRequest,
            $2.OutgoingPayment>(
        'LookupOutgoingPayment',
        lookupOutgoingPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.LookupOutgoingPaymentRequest.fromBuffer(value),
        ($2.OutgoingPayment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CanSendToPaymentPointerRequest,
            $2.CanSendToPaymentPointerResponse>(
        'CanSendToPaymentPointer',
        canSendToPaymentPointer_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CanSendToPaymentPointerRequest.fromBuffer(value),
        ($2.CanSendToPaymentPointerResponse value) => value.writeToBuffer()));
  }

  $async.Future<$2.Empty> createPaymentPointer_Pre($grpc.ServiceCall call,
      $async.Future<$2.CreatePaymentPointerRequest> request) async {
    return createPaymentPointer(call, await request);
  }

  $async.Future<$2.PaymentPointer> getPaymentPointer_Pre($grpc.ServiceCall call,
      $async.Future<$2.GetPaymentPointerRequest> request) async {
    return getPaymentPointer(call, await request);
  }

  $async.Future<$2.PaymentPointerExistsResponse> paymentPointerExists_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.PaymentPointerExistsRequest> request) async {
    return paymentPointerExists(call, await request);
  }

  $async.Future<$2.ListWalletPaymentPointersResponse>
      listWalletPaymentPointers_Pre(
          $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return listWalletPaymentPointers(call, await request);
  }

  $async.Future<$2.Quote> createQuote_Pre($grpc.ServiceCall call,
      $async.Future<$2.CreateQuoteRequest> request) async {
    return createQuote(call, await request);
  }

  $async.Future<$2.Quote> lookupQuote_Pre($grpc.ServiceCall call,
      $async.Future<$2.LookupQuoteRequest> request) async {
    return lookupQuote(call, await request);
  }

  $async.Future<$2.IncomingPayment> createIncomingPayment_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.CreateIncomingPaymentRequest> request) async {
    return createIncomingPayment(call, await request);
  }

  $async.Future<$2.IncomingPayment> lookupIncomingPayment_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.LookupIncomingPaymentRequest> request) async {
    return lookupIncomingPayment(call, await request);
  }

  $async.Future<$2.PreCheckOutgoingPaymentResponse> preCheckOutgoingPayment_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.PreCheckOutgoingPaymentRequest> request) async {
    return preCheckOutgoingPayment(call, await request);
  }

  $async.Future<$2.OutgoingPayment> createOutgoingPayment_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.CreateOutgoingPaymentRequest> request) async {
    return createOutgoingPayment(call, await request);
  }

  $async.Future<$2.OutgoingPayment> lookupOutgoingPayment_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.LookupOutgoingPaymentRequest> request) async {
    return lookupOutgoingPayment(call, await request);
  }

  $async.Future<$2.CanSendToPaymentPointerResponse> canSendToPaymentPointer_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.CanSendToPaymentPointerRequest> request) async {
    return canSendToPaymentPointer(call, await request);
  }

  $async.Future<$2.Empty> createPaymentPointer(
      $grpc.ServiceCall call, $2.CreatePaymentPointerRequest request);
  $async.Future<$2.PaymentPointer> getPaymentPointer(
      $grpc.ServiceCall call, $2.GetPaymentPointerRequest request);
  $async.Future<$2.PaymentPointerExistsResponse> paymentPointerExists(
      $grpc.ServiceCall call, $2.PaymentPointerExistsRequest request);
  $async.Future<$2.ListWalletPaymentPointersResponse> listWalletPaymentPointers(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Quote> createQuote(
      $grpc.ServiceCall call, $2.CreateQuoteRequest request);
  $async.Future<$2.Quote> lookupQuote(
      $grpc.ServiceCall call, $2.LookupQuoteRequest request);
  $async.Future<$2.IncomingPayment> createIncomingPayment(
      $grpc.ServiceCall call, $2.CreateIncomingPaymentRequest request);
  $async.Future<$2.IncomingPayment> lookupIncomingPayment(
      $grpc.ServiceCall call, $2.LookupIncomingPaymentRequest request);
  $async.Future<$2.PreCheckOutgoingPaymentResponse> preCheckOutgoingPayment(
      $grpc.ServiceCall call, $2.PreCheckOutgoingPaymentRequest request);
  $async.Future<$2.OutgoingPayment> createOutgoingPayment(
      $grpc.ServiceCall call, $2.CreateOutgoingPaymentRequest request);
  $async.Future<$2.OutgoingPayment> lookupOutgoingPayment(
      $grpc.ServiceCall call, $2.LookupOutgoingPaymentRequest request);
  $async.Future<$2.CanSendToPaymentPointerResponse> canSendToPaymentPointer(
      $grpc.ServiceCall call, $2.CanSendToPaymentPointerRequest request);
}

class BackendServiceClient extends $grpc.Client {
  static final _$updateIndividualKYC =
      $grpc.ClientMethod<$2.UpdateIndividualKYCRequest, $2.Empty>(
          '/backend.v1.BackendService/UpdateIndividualKYC',
          ($2.UpdateIndividualKYCRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getIndividualKYC =
      $grpc.ClientMethod<$2.Empty, $2.IndividualKYCResponse>(
          '/backend.v1.BackendService/GetIndividualKYC',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.IndividualKYCResponse.fromBuffer(value));
  static final _$isUSPSAddress =
      $grpc.ClientMethod<$2.Address, $2.IsUSPSAddressResponse>(
          '/backend.v1.BackendService/IsUSPSAddress',
          ($2.Address value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.IsUSPSAddressResponse.fromBuffer(value));
  static final _$setSignupUserData = $grpc.ClientMethod<
          $2.SetSignupUserDataRequest, $2.SetSignupUserDataResponse>(
      '/backend.v1.BackendService/SetSignupUserData',
      ($2.SetSignupUserDataRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $2.SetSignupUserDataResponse.fromBuffer(value));
  static final _$setSignupMobileNumber =
      $grpc.ClientMethod<$2.SetSignupMobileNumberRequest, $2.Empty>(
          '/backend.v1.BackendService/SetSignupMobileNumber',
          ($2.SetSignupMobileNumberRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getSignup = $grpc.ClientMethod<$2.GetSignupRequest, $2.Signup>(
      '/backend.v1.BackendService/GetSignup',
      ($2.GetSignupRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Signup.fromBuffer(value));
  static final _$completeSignup =
      $grpc.ClientMethod<$2.CompleteSignupRequest, $2.Empty>(
          '/backend.v1.BackendService/CompleteSignup',
          ($2.CompleteSignupRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$createUserDefaultWallet =
      $grpc.ClientMethod<$2.CreateUserDefaultWalletRequest, $2.Empty>(
          '/backend.v1.BackendService/CreateUserDefaultWallet',
          ($2.CreateUserDefaultWalletRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$setWalletName =
      $grpc.ClientMethod<$2.SetWalletNameRequest, $2.Empty>(
          '/backend.v1.BackendService/SetWalletName',
          ($2.SetWalletNameRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$sendPhoneVerification =
      $grpc.ClientMethod<$2.SendPhoneVerificationRequest, $2.Empty>(
          '/backend.v1.BackendService/SendPhoneVerification',
          ($2.SendPhoneVerificationRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$sendOTP = $grpc.ClientMethod<$2.Empty, $2.Empty>(
      '/backend.v1.BackendService/SendOTP',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getAgreement =
      $grpc.ClientMethod<$2.GetAgreementRequest, $2.Agreement>(
          '/backend.v1.BackendService/GetAgreement',
          ($2.GetAgreementRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Agreement.fromBuffer(value));
  static final _$signAgreements =
      $grpc.ClientMethod<$2.SignAgreementsRequest, $2.SignAgreementsResponse>(
          '/backend.v1.BackendService/SignAgreements',
          ($2.SignAgreementsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.SignAgreementsResponse.fromBuffer(value));
  static final _$getLinkedAccounts =
      $grpc.ClientMethod<$2.Empty, $2.GetLinkedAccountsResponse>(
          '/backend.v1.BackendService/GetLinkedAccounts',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.GetLinkedAccountsResponse.fromBuffer(value));
  static final _$getLinkedAccount =
      $grpc.ClientMethod<$2.GetLinkedAccountRequest, $2.LinkedAccount>(
          '/backend.v1.BackendService/GetLinkedAccount',
          ($2.GetLinkedAccountRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$setNicknameLinkedAccount =
      $grpc.ClientMethod<$2.SetNicknameLinkedAccountRequest, $2.LinkedAccount>(
          '/backend.v1.BackendService/SetNicknameLinkedAccount',
          ($2.SetNicknameLinkedAccountRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$deleteLinkedAccount =
      $grpc.ClientMethod<$2.DeleteLinkedAccountRequest, $2.Empty>(
          '/backend.v1.BackendService/DeleteLinkedAccount',
          ($2.DeleteLinkedAccountRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$createSupportTicket =
      $grpc.ClientMethod<$2.CreateSupportTicketRequest, $2.Empty>(
          '/backend.v1.BackendService/CreateSupportTicket',
          ($2.CreateSupportTicketRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getCountries =
      $grpc.ClientMethod<$2.Empty, $2.GetCountriesResponse>(
          '/backend.v1.BackendService/GetCountries',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.GetCountriesResponse.fromBuffer(value));
  static final _$getCurrentWallet =
      $grpc.ClientMethod<$2.Empty, $2.GetCurrentWalletResponse>(
          '/backend.v1.BackendService/GetCurrentWallet',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.GetCurrentWalletResponse.fromBuffer(value));
  static final _$getMachnetWidgetToken =
      $grpc.ClientMethod<$2.Empty, $2.MachnetWidgetToken>(
          '/backend.v1.BackendService/GetMachnetWidgetToken',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.MachnetWidgetToken.fromBuffer(value));
  static final _$listBanks = $grpc.ClientMethod<$2.Empty, $2.ListBanksResponse>(
      '/backend.v1.BackendService/ListBanks',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListBanksResponse.fromBuffer(value));
  static final _$createReceiveBankAccount =
      $grpc.ClientMethod<$2.CreateReceiveBankAccountRequest, $2.LinkedAccount>(
          '/backend.v1.BackendService/CreateReceiveBankAccount',
          ($2.CreateReceiveBankAccountRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$startMachnetKYC = $grpc.ClientMethod<$2.Empty, $2.Empty>(
      '/backend.v1.BackendService/StartMachnetKYC',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$hasSendUser =
      $grpc.ClientMethod<$2.Empty, $2.HasSendUserResponse>(
          '/backend.v1.BackendService/HasSendUser',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.HasSendUserResponse.fromBuffer(value));
  static final _$kYCStatus = $grpc.ClientMethod<$2.Empty, $2.KYCStatusResponse>(
      '/backend.v1.BackendService/KYCStatus',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.KYCStatusResponse.fromBuffer(value));
  static final _$createWallet =
      $grpc.ClientMethod<$2.CreateWalletRequest, $2.LinkedAccount>(
          '/backend.v1.BackendService/CreateWallet',
          ($2.CreateWalletRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$getWalletBalance =
      $grpc.ClientMethod<$2.Empty, $2.WalletBalance>(
          '/backend.v1.BackendService/GetWalletBalance',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.WalletBalance.fromBuffer(value));
  static final _$checkMachnetWithdrawalLimit = $grpc.ClientMethod<
          $2.CheckMachnetTXLimitRequest, $2.CheckMachnetTXLimitResponse>(
      '/backend.v1.BackendService/CheckMachnetWithdrawalLimit',
      ($2.CheckMachnetTXLimitRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $2.CheckMachnetTXLimitResponse.fromBuffer(value));
  static final _$startWithdrawFromMachnetWallet =
      $grpc.ClientMethod<$2.WithdrawFromMachnetWalletRequest, $2.Empty>(
          '/backend.v1.BackendService/StartWithdrawFromMachnetWallet',
          ($2.WithdrawFromMachnetWalletRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$checkMachnetTopupLimit = $grpc.ClientMethod<
          $2.CheckMachnetTXLimitRequest, $2.CheckMachnetTXLimitResponse>(
      '/backend.v1.BackendService/CheckMachnetTopupLimit',
      ($2.CheckMachnetTXLimitRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $2.CheckMachnetTXLimitResponse.fromBuffer(value));
  static final _$startMachnetWalletTopup =
      $grpc.ClientMethod<$2.StartMachnetWalletTopupRequest, $2.Empty>(
          '/backend.v1.BackendService/StartMachnetWalletTopup',
          ($2.StartMachnetWalletTopupRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$joinWaitlist =
      $grpc.ClientMethod<$2.JoinWaitlistRequest, $2.JoinWaitlistResponse>(
          '/backend.v1.BackendService/JoinWaitlist',
          ($2.JoinWaitlistRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.JoinWaitlistResponse.fromBuffer(value));
  static final _$canSignup =
      $grpc.ClientMethod<$2.CanSignupRequest, $2.CanSignupResponse>(
          '/backend.v1.BackendService/CanSignup',
          ($2.CanSignupRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.CanSignupResponse.fromBuffer(value));
  static final _$setSignupComplete =
      $grpc.ClientMethod<$2.SetSignupCompleteRequest, $2.Empty>(
          '/backend.v1.BackendService/SetSignupComplete',
          ($2.SetSignupCompleteRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$isMugAvailable =
      $grpc.ClientMethod<$2.IsMugAvailableRequest, $2.IsMugAvailableResponse>(
          '/backend.v1.BackendService/IsMugAvailable',
          ($2.IsMugAvailableRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.IsMugAvailableResponse.fromBuffer(value));
  static final _$listTransactions =
      $grpc.ClientMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
          '/backend.v1.BackendService/ListTransactions',
          ($2.PaginationRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.ListTransactionsResponse.fromBuffer(value));
  static final _$listTransactionsCompleted =
      $grpc.ClientMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
          '/backend.v1.BackendService/ListTransactionsCompleted',
          ($2.PaginationRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.ListTransactionsResponse.fromBuffer(value));
  static final _$listTransactionsWithPending =
      $grpc.ClientMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
          '/backend.v1.BackendService/ListTransactionsWithPending',
          ($2.PaginationRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.ListTransactionsResponse.fromBuffer(value));
  static final _$lookupTransaction =
      $grpc.ClientMethod<$2.LookupTransactionRequest, $2.Transaction>(
          '/backend.v1.BackendService/LookupTransaction',
          ($2.LookupTransactionRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Transaction.fromBuffer(value));
  static final _$getUserLimits =
      $grpc.ClientMethod<$2.Empty, $2.GetUserLimitsResponse>(
          '/backend.v1.BackendService/GetUserLimits',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.GetUserLimitsResponse.fromBuffer(value));
  static final _$listLimits =
      $grpc.ClientMethod<$2.Empty, $2.ListLimitsResponse>(
          '/backend.v1.BackendService/ListLimits',
          ($2.Empty value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.ListLimitsResponse.fromBuffer(value));
  static final _$updateClientLimits =
      $grpc.ClientMethod<$2.UpdateClientLimitsRequest, $2.Empty>(
          '/backend.v1.BackendService/UpdateClientLimits',
          ($2.UpdateClientLimitsRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getStatementPDF =
      $grpc.ClientMethod<$2.GetStatementPDFRequest, $2.StatementPDF>(
          '/backend.v1.BackendService/GetStatementPDF',
          ($2.GetStatementPDFRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.StatementPDF.fromBuffer(value));
  static final _$createClientPublicKey = $grpc.ClientMethod<$2.JWK, $2.Empty>(
      '/backend.v1.BackendService/CreateClientPublicKey',
      ($2.JWK value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getPublicWalletDetails = $grpc.ClientMethod<
          $2.GetPublicWalletDetailsRequest, $2.GetPublicWalletDetailsResponse>(
      '/backend.v1.BackendService/GetPublicWalletDetails',
      ($2.GetPublicWalletDetailsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) =>
          $2.GetPublicWalletDetailsResponse.fromBuffer(value));
  static final _$createContact =
      $grpc.ClientMethod<$2.CreateContactRequest, $2.Contact>(
          '/backend.v1.BackendService/CreateContact',
          ($2.CreateContactRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) => $2.Contact.fromBuffer(value));
  static final _$listContacts =
      $grpc.ClientMethod<$2.PaginationRequest, $2.ListContactsResponse>(
          '/backend.v1.BackendService/ListContacts',
          ($2.PaginationRequest value) => value.writeToBuffer(),
          ($core.List<$core.int> value) =>
              $2.ListContactsResponse.fromBuffer(value));

  BackendServiceClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options, interceptors: interceptors);

  $grpc.ResponseFuture<$2.Empty> updateIndividualKYC(
      $2.UpdateIndividualKYCRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$updateIndividualKYC, request, options: options);
  }

  $grpc.ResponseFuture<$2.IndividualKYCResponse> getIndividualKYC(
      $2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getIndividualKYC, request, options: options);
  }

  $grpc.ResponseFuture<$2.IsUSPSAddressResponse> isUSPSAddress(
      $2.Address request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$isUSPSAddress, request, options: options);
  }

  $grpc.ResponseFuture<$2.SetSignupUserDataResponse> setSignupUserData(
      $2.SetSignupUserDataRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setSignupUserData, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> setSignupMobileNumber(
      $2.SetSignupMobileNumberRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setSignupMobileNumber, request, options: options);
  }

  $grpc.ResponseFuture<$2.Signup> getSignup($2.GetSignupRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getSignup, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> completeSignup(
      $2.CompleteSignupRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$completeSignup, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> createUserDefaultWallet(
      $2.CreateUserDefaultWalletRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createUserDefaultWallet, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Empty> setWalletName($2.SetWalletNameRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setWalletName, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> sendPhoneVerification(
      $2.SendPhoneVerificationRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$sendPhoneVerification, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> sendOTP($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$sendOTP, request, options: options);
  }

  $grpc.ResponseFuture<$2.Agreement> getAgreement(
      $2.GetAgreementRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getAgreement, request, options: options);
  }

  $grpc.ResponseFuture<$2.SignAgreementsResponse> signAgreements(
      $2.SignAgreementsRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$signAgreements, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetLinkedAccountsResponse> getLinkedAccounts(
      $2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLinkedAccounts, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> getLinkedAccount(
      $2.GetLinkedAccountRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> setNicknameLinkedAccount(
      $2.SetNicknameLinkedAccountRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setNicknameLinkedAccount, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Empty> deleteLinkedAccount(
      $2.DeleteLinkedAccountRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$deleteLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> createSupportTicket(
      $2.CreateSupportTicketRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createSupportTicket, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetCountriesResponse> getCountries($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getCountries, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetCurrentWalletResponse> getCurrentWallet(
      $2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getCurrentWallet, request, options: options);
  }

  $grpc.ResponseFuture<$2.MachnetWidgetToken> getMachnetWidgetToken(
      $2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getMachnetWidgetToken, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListBanksResponse> listBanks($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listBanks, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> createReceiveBankAccount(
      $2.CreateReceiveBankAccountRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createReceiveBankAccount, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Empty> startMachnetKYC($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$startMachnetKYC, request, options: options);
  }

  $grpc.ResponseFuture<$2.HasSendUserResponse> hasSendUser($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$hasSendUser, request, options: options);
  }

  $grpc.ResponseFuture<$2.KYCStatusResponse> kYCStatus($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$kYCStatus, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> createWallet(
      $2.CreateWalletRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createWallet, request, options: options);
  }

  $grpc.ResponseFuture<$2.WalletBalance> getWalletBalance($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWalletBalance, request, options: options);
  }

  $grpc.ResponseFuture<$2.CheckMachnetTXLimitResponse>
      checkMachnetWithdrawalLimit($2.CheckMachnetTXLimitRequest request,
          {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$checkMachnetWithdrawalLimit, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Empty> startWithdrawFromMachnetWallet(
      $2.WithdrawFromMachnetWalletRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$startWithdrawFromMachnetWallet, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.CheckMachnetTXLimitResponse> checkMachnetTopupLimit(
      $2.CheckMachnetTXLimitRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$checkMachnetTopupLimit, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Empty> startMachnetWalletTopup(
      $2.StartMachnetWalletTopupRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$startMachnetWalletTopup, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.JoinWaitlistResponse> joinWaitlist(
      $2.JoinWaitlistRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$joinWaitlist, request, options: options);
  }

  $grpc.ResponseFuture<$2.CanSignupResponse> canSignup(
      $2.CanSignupRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$canSignup, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> setSignupComplete(
      $2.SetSignupCompleteRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setSignupComplete, request, options: options);
  }

  $grpc.ResponseFuture<$2.IsMugAvailableResponse> isMugAvailable(
      $2.IsMugAvailableRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$isMugAvailable, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListTransactionsResponse> listTransactions(
      $2.PaginationRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listTransactions, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListTransactionsResponse> listTransactionsCompleted(
      $2.PaginationRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listTransactionsCompleted, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.ListTransactionsResponse> listTransactionsWithPending(
      $2.PaginationRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listTransactionsWithPending, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Transaction> lookupTransaction(
      $2.LookupTransactionRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$lookupTransaction, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetUserLimitsResponse> getUserLimits($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getUserLimits, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListLimitsResponse> listLimits($2.Empty request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listLimits, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> updateClientLimits(
      $2.UpdateClientLimitsRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$updateClientLimits, request, options: options);
  }

  $grpc.ResponseFuture<$2.StatementPDF> getStatementPDF(
      $2.GetStatementPDFRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getStatementPDF, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> createClientPublicKey($2.JWK request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createClientPublicKey, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetPublicWalletDetailsResponse>
      getPublicWalletDetails($2.GetPublicWalletDetailsRequest request,
          {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getPublicWalletDetails, request,
        options: options);
  }

  $grpc.ResponseFuture<$2.Contact> createContact(
      $2.CreateContactRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createContact, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListContactsResponse> listContacts(
      $2.PaginationRequest request,
      {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listContacts, request, options: options);
  }
}

abstract class BackendServiceBase extends $grpc.Service {
  $core.String get $name => 'backend.v1.BackendService';

  BackendServiceBase() {
    $addMethod($grpc.ServiceMethod<$2.UpdateIndividualKYCRequest, $2.Empty>(
        'UpdateIndividualKYC',
        updateIndividualKYC_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.UpdateIndividualKYCRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.IndividualKYCResponse>(
        'GetIndividualKYC',
        getIndividualKYC_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.IndividualKYCResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Address, $2.IsUSPSAddressResponse>(
        'IsUSPSAddress',
        isUSPSAddress_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Address.fromBuffer(value),
        ($2.IsUSPSAddressResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetSignupUserDataRequest,
            $2.SetSignupUserDataResponse>(
        'SetSignupUserData',
        setSignupUserData_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.SetSignupUserDataRequest.fromBuffer(value),
        ($2.SetSignupUserDataResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetSignupMobileNumberRequest, $2.Empty>(
        'SetSignupMobileNumber',
        setSignupMobileNumber_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.SetSignupMobileNumberRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetSignupRequest, $2.Signup>(
        'GetSignup',
        getSignup_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetSignupRequest.fromBuffer(value),
        ($2.Signup value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CompleteSignupRequest, $2.Empty>(
        'CompleteSignup',
        completeSignup_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CompleteSignupRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateUserDefaultWalletRequest, $2.Empty>(
        'CreateUserDefaultWallet',
        createUserDefaultWallet_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateUserDefaultWalletRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetWalletNameRequest, $2.Empty>(
        'SetWalletName',
        setWalletName_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.SetWalletNameRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SendPhoneVerificationRequest, $2.Empty>(
        'SendPhoneVerification',
        sendPhoneVerification_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.SendPhoneVerificationRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.Empty>(
        'SendOTP',
        sendOTP_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetAgreementRequest, $2.Agreement>(
        'GetAgreement',
        getAgreement_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.GetAgreementRequest.fromBuffer(value),
        ($2.Agreement value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SignAgreementsRequest,
            $2.SignAgreementsResponse>(
        'SignAgreements',
        signAgreements_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.SignAgreementsRequest.fromBuffer(value),
        ($2.SignAgreementsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.GetLinkedAccountsResponse>(
        'GetLinkedAccounts',
        getLinkedAccounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.GetLinkedAccountsResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.GetLinkedAccountRequest, $2.LinkedAccount>(
            'GetLinkedAccount',
            getLinkedAccount_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.GetLinkedAccountRequest.fromBuffer(value),
            ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetNicknameLinkedAccountRequest,
            $2.LinkedAccount>(
        'SetNicknameLinkedAccount',
        setNicknameLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.SetNicknameLinkedAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.DeleteLinkedAccountRequest, $2.Empty>(
        'DeleteLinkedAccount',
        deleteLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.DeleteLinkedAccountRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateSupportTicketRequest, $2.Empty>(
        'CreateSupportTicket',
        createSupportTicket_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateSupportTicketRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.GetCountriesResponse>(
        'GetCountries',
        getCountries_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.GetCountriesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.GetCurrentWalletResponse>(
        'GetCurrentWallet',
        getCurrentWallet_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.GetCurrentWalletResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.MachnetWidgetToken>(
        'GetMachnetWidgetToken',
        getMachnetWidgetToken_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.MachnetWidgetToken value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.ListBanksResponse>(
        'ListBanks',
        listBanks_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.ListBanksResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateReceiveBankAccountRequest,
            $2.LinkedAccount>(
        'CreateReceiveBankAccount',
        createReceiveBankAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateReceiveBankAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.Empty>(
        'StartMachnetKYC',
        startMachnetKYC_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.HasSendUserResponse>(
        'HasSendUser',
        hasSendUser_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.HasSendUserResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.KYCStatusResponse>(
        'KYCStatus',
        kYCStatus_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.KYCStatusResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateWalletRequest, $2.LinkedAccount>(
        'CreateWallet',
        createWallet_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateWalletRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.WalletBalance>(
        'GetWalletBalance',
        getWalletBalance_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.WalletBalance value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CheckMachnetTXLimitRequest,
            $2.CheckMachnetTXLimitResponse>(
        'CheckMachnetWithdrawalLimit',
        checkMachnetWithdrawalLimit_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CheckMachnetTXLimitRequest.fromBuffer(value),
        ($2.CheckMachnetTXLimitResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.WithdrawFromMachnetWalletRequest, $2.Empty>(
            'StartWithdrawFromMachnetWallet',
            startWithdrawFromMachnetWallet_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.WithdrawFromMachnetWalletRequest.fromBuffer(value),
            ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CheckMachnetTXLimitRequest,
            $2.CheckMachnetTXLimitResponse>(
        'CheckMachnetTopupLimit',
        checkMachnetTopupLimit_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CheckMachnetTXLimitRequest.fromBuffer(value),
        ($2.CheckMachnetTXLimitResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.StartMachnetWalletTopupRequest, $2.Empty>(
        'StartMachnetWalletTopup',
        startMachnetWalletTopup_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.StartMachnetWalletTopupRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.JoinWaitlistRequest, $2.JoinWaitlistResponse>(
            'JoinWaitlist',
            joinWaitlist_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.JoinWaitlistRequest.fromBuffer(value),
            ($2.JoinWaitlistResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CanSignupRequest, $2.CanSignupResponse>(
        'CanSignup',
        canSignup_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CanSignupRequest.fromBuffer(value),
        ($2.CanSignupResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetSignupCompleteRequest, $2.Empty>(
        'SetSignupComplete',
        setSignupComplete_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.SetSignupCompleteRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.IsMugAvailableRequest,
            $2.IsMugAvailableResponse>(
        'IsMugAvailable',
        isMugAvailable_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.IsMugAvailableRequest.fromBuffer(value),
        ($2.IsMugAvailableResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
            'ListTransactions',
            listTransactions_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.PaginationRequest.fromBuffer(value),
            ($2.ListTransactionsResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
            'ListTransactionsCompleted',
            listTransactionsCompleted_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.PaginationRequest.fromBuffer(value),
            ($2.ListTransactionsResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
            'ListTransactionsWithPending',
            listTransactionsWithPending_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.PaginationRequest.fromBuffer(value),
            ($2.ListTransactionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.LookupTransactionRequest, $2.Transaction>(
        'LookupTransaction',
        lookupTransaction_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.LookupTransactionRequest.fromBuffer(value),
        ($2.Transaction value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.GetUserLimitsResponse>(
        'GetUserLimits',
        getUserLimits_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.GetUserLimitsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.ListLimitsResponse>(
        'ListLimits',
        listLimits_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.ListLimitsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.UpdateClientLimitsRequest, $2.Empty>(
        'UpdateClientLimits',
        updateClientLimits_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.UpdateClientLimitsRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetStatementPDFRequest, $2.StatementPDF>(
        'GetStatementPDF',
        getStatementPDF_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.GetStatementPDFRequest.fromBuffer(value),
        ($2.StatementPDF value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.JWK, $2.Empty>(
        'CreateClientPublicKey',
        createClientPublicKey_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.JWK.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetPublicWalletDetailsRequest,
            $2.GetPublicWalletDetailsResponse>(
        'GetPublicWalletDetails',
        getPublicWalletDetails_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.GetPublicWalletDetailsRequest.fromBuffer(value),
        ($2.GetPublicWalletDetailsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateContactRequest, $2.Contact>(
        'CreateContact',
        createContact_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $2.CreateContactRequest.fromBuffer(value),
        ($2.Contact value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$2.PaginationRequest, $2.ListContactsResponse>(
            'ListContacts',
            listContacts_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $2.PaginationRequest.fromBuffer(value),
            ($2.ListContactsResponse value) => value.writeToBuffer()));
  }

  $async.Future<$2.Empty> updateIndividualKYC_Pre($grpc.ServiceCall call,
      $async.Future<$2.UpdateIndividualKYCRequest> request) async {
    return updateIndividualKYC(call, await request);
  }

  $async.Future<$2.IndividualKYCResponse> getIndividualKYC_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getIndividualKYC(call, await request);
  }

  $async.Future<$2.IsUSPSAddressResponse> isUSPSAddress_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Address> request) async {
    return isUSPSAddress(call, await request);
  }

  $async.Future<$2.SetSignupUserDataResponse> setSignupUserData_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.SetSignupUserDataRequest> request) async {
    return setSignupUserData(call, await request);
  }

  $async.Future<$2.Empty> setSignupMobileNumber_Pre($grpc.ServiceCall call,
      $async.Future<$2.SetSignupMobileNumberRequest> request) async {
    return setSignupMobileNumber(call, await request);
  }

  $async.Future<$2.Signup> getSignup_Pre($grpc.ServiceCall call,
      $async.Future<$2.GetSignupRequest> request) async {
    return getSignup(call, await request);
  }

  $async.Future<$2.Empty> completeSignup_Pre($grpc.ServiceCall call,
      $async.Future<$2.CompleteSignupRequest> request) async {
    return completeSignup(call, await request);
  }

  $async.Future<$2.Empty> createUserDefaultWallet_Pre($grpc.ServiceCall call,
      $async.Future<$2.CreateUserDefaultWalletRequest> request) async {
    return createUserDefaultWallet(call, await request);
  }

  $async.Future<$2.Empty> setWalletName_Pre($grpc.ServiceCall call,
      $async.Future<$2.SetWalletNameRequest> request) async {
    return setWalletName(call, await request);
  }

  $async.Future<$2.Empty> sendPhoneVerification_Pre($grpc.ServiceCall call,
      $async.Future<$2.SendPhoneVerificationRequest> request) async {
    return sendPhoneVerification(call, await request);
  }

  $async.Future<$2.Empty> sendOTP_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return sendOTP(call, await request);
  }

  $async.Future<$2.Agreement> getAgreement_Pre($grpc.ServiceCall call,
      $async.Future<$2.GetAgreementRequest> request) async {
    return getAgreement(call, await request);
  }

  $async.Future<$2.SignAgreementsResponse> signAgreements_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.SignAgreementsRequest> request) async {
    return signAgreements(call, await request);
  }

  $async.Future<$2.GetLinkedAccountsResponse> getLinkedAccounts_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getLinkedAccounts(call, await request);
  }

  $async.Future<$2.LinkedAccount> getLinkedAccount_Pre($grpc.ServiceCall call,
      $async.Future<$2.GetLinkedAccountRequest> request) async {
    return getLinkedAccount(call, await request);
  }

  $async.Future<$2.LinkedAccount> setNicknameLinkedAccount_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.SetNicknameLinkedAccountRequest> request) async {
    return setNicknameLinkedAccount(call, await request);
  }

  $async.Future<$2.Empty> deleteLinkedAccount_Pre($grpc.ServiceCall call,
      $async.Future<$2.DeleteLinkedAccountRequest> request) async {
    return deleteLinkedAccount(call, await request);
  }

  $async.Future<$2.Empty> createSupportTicket_Pre($grpc.ServiceCall call,
      $async.Future<$2.CreateSupportTicketRequest> request) async {
    return createSupportTicket(call, await request);
  }

  $async.Future<$2.GetCountriesResponse> getCountries_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getCountries(call, await request);
  }

  $async.Future<$2.GetCurrentWalletResponse> getCurrentWallet_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getCurrentWallet(call, await request);
  }

  $async.Future<$2.MachnetWidgetToken> getMachnetWidgetToken_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getMachnetWidgetToken(call, await request);
  }

  $async.Future<$2.ListBanksResponse> listBanks_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return listBanks(call, await request);
  }

  $async.Future<$2.LinkedAccount> createReceiveBankAccount_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.CreateReceiveBankAccountRequest> request) async {
    return createReceiveBankAccount(call, await request);
  }

  $async.Future<$2.Empty> startMachnetKYC_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return startMachnetKYC(call, await request);
  }

  $async.Future<$2.HasSendUserResponse> hasSendUser_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return hasSendUser(call, await request);
  }

  $async.Future<$2.KYCStatusResponse> kYCStatus_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return kYCStatus(call, await request);
  }

  $async.Future<$2.LinkedAccount> createWallet_Pre($grpc.ServiceCall call,
      $async.Future<$2.CreateWalletRequest> request) async {
    return createWallet(call, await request);
  }

  $async.Future<$2.WalletBalance> getWalletBalance_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getWalletBalance(call, await request);
  }

  $async.Future<$2.CheckMachnetTXLimitResponse> checkMachnetWithdrawalLimit_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.CheckMachnetTXLimitRequest> request) async {
    return checkMachnetWithdrawalLimit(call, await request);
  }

  $async.Future<$2.Empty> startWithdrawFromMachnetWallet_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.WithdrawFromMachnetWalletRequest> request) async {
    return startWithdrawFromMachnetWallet(call, await request);
  }

  $async.Future<$2.CheckMachnetTXLimitResponse> checkMachnetTopupLimit_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.CheckMachnetTXLimitRequest> request) async {
    return checkMachnetTopupLimit(call, await request);
  }

  $async.Future<$2.Empty> startMachnetWalletTopup_Pre($grpc.ServiceCall call,
      $async.Future<$2.StartMachnetWalletTopupRequest> request) async {
    return startMachnetWalletTopup(call, await request);
  }

  $async.Future<$2.JoinWaitlistResponse> joinWaitlist_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.JoinWaitlistRequest> request) async {
    return joinWaitlist(call, await request);
  }

  $async.Future<$2.CanSignupResponse> canSignup_Pre($grpc.ServiceCall call,
      $async.Future<$2.CanSignupRequest> request) async {
    return canSignup(call, await request);
  }

  $async.Future<$2.Empty> setSignupComplete_Pre($grpc.ServiceCall call,
      $async.Future<$2.SetSignupCompleteRequest> request) async {
    return setSignupComplete(call, await request);
  }

  $async.Future<$2.IsMugAvailableResponse> isMugAvailable_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.IsMugAvailableRequest> request) async {
    return isMugAvailable(call, await request);
  }

  $async.Future<$2.ListTransactionsResponse> listTransactions_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.PaginationRequest> request) async {
    return listTransactions(call, await request);
  }

  $async.Future<$2.ListTransactionsResponse> listTransactionsCompleted_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.PaginationRequest> request) async {
    return listTransactionsCompleted(call, await request);
  }

  $async.Future<$2.ListTransactionsResponse> listTransactionsWithPending_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.PaginationRequest> request) async {
    return listTransactionsWithPending(call, await request);
  }

  $async.Future<$2.Transaction> lookupTransaction_Pre($grpc.ServiceCall call,
      $async.Future<$2.LookupTransactionRequest> request) async {
    return lookupTransaction(call, await request);
  }

  $async.Future<$2.GetUserLimitsResponse> getUserLimits_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getUserLimits(call, await request);
  }

  $async.Future<$2.ListLimitsResponse> listLimits_Pre(
      $grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return listLimits(call, await request);
  }

  $async.Future<$2.Empty> updateClientLimits_Pre($grpc.ServiceCall call,
      $async.Future<$2.UpdateClientLimitsRequest> request) async {
    return updateClientLimits(call, await request);
  }

  $async.Future<$2.StatementPDF> getStatementPDF_Pre($grpc.ServiceCall call,
      $async.Future<$2.GetStatementPDFRequest> request) async {
    return getStatementPDF(call, await request);
  }

  $async.Future<$2.Empty> createClientPublicKey_Pre(
      $grpc.ServiceCall call, $async.Future<$2.JWK> request) async {
    return createClientPublicKey(call, await request);
  }

  $async.Future<$2.GetPublicWalletDetailsResponse> getPublicWalletDetails_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.GetPublicWalletDetailsRequest> request) async {
    return getPublicWalletDetails(call, await request);
  }

  $async.Future<$2.Contact> createContact_Pre($grpc.ServiceCall call,
      $async.Future<$2.CreateContactRequest> request) async {
    return createContact(call, await request);
  }

  $async.Future<$2.ListContactsResponse> listContacts_Pre(
      $grpc.ServiceCall call,
      $async.Future<$2.PaginationRequest> request) async {
    return listContacts(call, await request);
  }

  $async.Future<$2.Empty> updateIndividualKYC(
      $grpc.ServiceCall call, $2.UpdateIndividualKYCRequest request);
  $async.Future<$2.IndividualKYCResponse> getIndividualKYC(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.IsUSPSAddressResponse> isUSPSAddress(
      $grpc.ServiceCall call, $2.Address request);
  $async.Future<$2.SetSignupUserDataResponse> setSignupUserData(
      $grpc.ServiceCall call, $2.SetSignupUserDataRequest request);
  $async.Future<$2.Empty> setSignupMobileNumber(
      $grpc.ServiceCall call, $2.SetSignupMobileNumberRequest request);
  $async.Future<$2.Signup> getSignup(
      $grpc.ServiceCall call, $2.GetSignupRequest request);
  $async.Future<$2.Empty> completeSignup(
      $grpc.ServiceCall call, $2.CompleteSignupRequest request);
  $async.Future<$2.Empty> createUserDefaultWallet(
      $grpc.ServiceCall call, $2.CreateUserDefaultWalletRequest request);
  $async.Future<$2.Empty> setWalletName(
      $grpc.ServiceCall call, $2.SetWalletNameRequest request);
  $async.Future<$2.Empty> sendPhoneVerification(
      $grpc.ServiceCall call, $2.SendPhoneVerificationRequest request);
  $async.Future<$2.Empty> sendOTP($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Agreement> getAgreement(
      $grpc.ServiceCall call, $2.GetAgreementRequest request);
  $async.Future<$2.SignAgreementsResponse> signAgreements(
      $grpc.ServiceCall call, $2.SignAgreementsRequest request);
  $async.Future<$2.GetLinkedAccountsResponse> getLinkedAccounts(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.LinkedAccount> getLinkedAccount(
      $grpc.ServiceCall call, $2.GetLinkedAccountRequest request);
  $async.Future<$2.LinkedAccount> setNicknameLinkedAccount(
      $grpc.ServiceCall call, $2.SetNicknameLinkedAccountRequest request);
  $async.Future<$2.Empty> deleteLinkedAccount(
      $grpc.ServiceCall call, $2.DeleteLinkedAccountRequest request);
  $async.Future<$2.Empty> createSupportTicket(
      $grpc.ServiceCall call, $2.CreateSupportTicketRequest request);
  $async.Future<$2.GetCountriesResponse> getCountries(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.GetCurrentWalletResponse> getCurrentWallet(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.MachnetWidgetToken> getMachnetWidgetToken(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.ListBanksResponse> listBanks(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.LinkedAccount> createReceiveBankAccount(
      $grpc.ServiceCall call, $2.CreateReceiveBankAccountRequest request);
  $async.Future<$2.Empty> startMachnetKYC(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.HasSendUserResponse> hasSendUser(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.KYCStatusResponse> kYCStatus(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.LinkedAccount> createWallet(
      $grpc.ServiceCall call, $2.CreateWalletRequest request);
  $async.Future<$2.WalletBalance> getWalletBalance(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.CheckMachnetTXLimitResponse> checkMachnetWithdrawalLimit(
      $grpc.ServiceCall call, $2.CheckMachnetTXLimitRequest request);
  $async.Future<$2.Empty> startWithdrawFromMachnetWallet(
      $grpc.ServiceCall call, $2.WithdrawFromMachnetWalletRequest request);
  $async.Future<$2.CheckMachnetTXLimitResponse> checkMachnetTopupLimit(
      $grpc.ServiceCall call, $2.CheckMachnetTXLimitRequest request);
  $async.Future<$2.Empty> startMachnetWalletTopup(
      $grpc.ServiceCall call, $2.StartMachnetWalletTopupRequest request);
  $async.Future<$2.JoinWaitlistResponse> joinWaitlist(
      $grpc.ServiceCall call, $2.JoinWaitlistRequest request);
  $async.Future<$2.CanSignupResponse> canSignup(
      $grpc.ServiceCall call, $2.CanSignupRequest request);
  $async.Future<$2.Empty> setSignupComplete(
      $grpc.ServiceCall call, $2.SetSignupCompleteRequest request);
  $async.Future<$2.IsMugAvailableResponse> isMugAvailable(
      $grpc.ServiceCall call, $2.IsMugAvailableRequest request);
  $async.Future<$2.ListTransactionsResponse> listTransactions(
      $grpc.ServiceCall call, $2.PaginationRequest request);
  $async.Future<$2.ListTransactionsResponse> listTransactionsCompleted(
      $grpc.ServiceCall call, $2.PaginationRequest request);
  $async.Future<$2.ListTransactionsResponse> listTransactionsWithPending(
      $grpc.ServiceCall call, $2.PaginationRequest request);
  $async.Future<$2.Transaction> lookupTransaction(
      $grpc.ServiceCall call, $2.LookupTransactionRequest request);
  $async.Future<$2.GetUserLimitsResponse> getUserLimits(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.ListLimitsResponse> listLimits(
      $grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Empty> updateClientLimits(
      $grpc.ServiceCall call, $2.UpdateClientLimitsRequest request);
  $async.Future<$2.StatementPDF> getStatementPDF(
      $grpc.ServiceCall call, $2.GetStatementPDFRequest request);
  $async.Future<$2.Empty> createClientPublicKey(
      $grpc.ServiceCall call, $2.JWK request);
  $async.Future<$2.GetPublicWalletDetailsResponse> getPublicWalletDetails(
      $grpc.ServiceCall call, $2.GetPublicWalletDetailsRequest request);
  $async.Future<$2.Contact> createContact(
      $grpc.ServiceCall call, $2.CreateContactRequest request);
  $async.Future<$2.ListContactsResponse> listContacts(
      $grpc.ServiceCall call, $2.PaginationRequest request);
}
