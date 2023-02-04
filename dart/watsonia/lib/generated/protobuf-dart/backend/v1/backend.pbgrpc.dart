//
//  Generated code. Do not modify.
//  source: backend/v1/backend.proto
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

import 'backend.pb.dart' as $2;

export 'backend.pb.dart';

@$pb.GrpcServiceName('backend.v1.BackendService')
class BackendServiceClient extends $grpc.Client {
  static final _$updateIndividualKYC = $grpc.ClientMethod<$2.UpdateIndividualKYCRequest, $2.Empty>(
      '/backend.v1.BackendService/UpdateIndividualKYC',
      ($2.UpdateIndividualKYCRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getIndividualKYC = $grpc.ClientMethod<$2.Empty, $2.IndividualKYCResponse>(
      '/backend.v1.BackendService/GetIndividualKYC',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.IndividualKYCResponse.fromBuffer(value));
  static final _$isUSPSAddress = $grpc.ClientMethod<$2.Address, $2.IsUSPSAddressResponse>(
      '/backend.v1.BackendService/IsUSPSAddress',
      ($2.Address value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.IsUSPSAddressResponse.fromBuffer(value));
  static final _$setSignupUserData = $grpc.ClientMethod<$2.SetSignupUserDataRequest, $2.SetSignupUserDataResponse>(
      '/backend.v1.BackendService/SetSignupUserData',
      ($2.SetSignupUserDataRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.SetSignupUserDataResponse.fromBuffer(value));
  static final _$setSignupMobileNumber = $grpc.ClientMethod<$2.SetSignupMobileNumberRequest, $2.Empty>(
      '/backend.v1.BackendService/SetSignupMobileNumber',
      ($2.SetSignupMobileNumberRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getSignup = $grpc.ClientMethod<$2.GetSignupRequest, $2.Signup>(
      '/backend.v1.BackendService/GetSignup',
      ($2.GetSignupRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Signup.fromBuffer(value));
  static final _$completeSignup = $grpc.ClientMethod<$2.CompleteSignupRequest, $2.Empty>(
      '/backend.v1.BackendService/CompleteSignup',
      ($2.CompleteSignupRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$createUserDefaultWallet = $grpc.ClientMethod<$2.CreateUserDefaultWalletRequest, $2.Empty>(
      '/backend.v1.BackendService/CreateUserDefaultWallet',
      ($2.CreateUserDefaultWalletRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$createWalletAddress = $grpc.ClientMethod<$2.CreateWalletAddressRequest, $2.Empty>(
      '/backend.v1.BackendService/CreateWalletAddress',
      ($2.CreateWalletAddressRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$walletAddressExists = $grpc.ClientMethod<$2.WalletAddressExistsRequest, $2.WalletAddressExistsResponse>(
      '/backend.v1.BackendService/WalletAddressExists',
      ($2.WalletAddressExistsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.WalletAddressExistsResponse.fromBuffer(value));
  static final _$setWalletName = $grpc.ClientMethod<$2.SetWalletNameRequest, $2.Empty>(
      '/backend.v1.BackendService/SetWalletName',
      ($2.SetWalletNameRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getWalletInfo = $grpc.ClientMethod<$2.Empty, $2.WalletInfo>(
      '/backend.v1.BackendService/GetWalletInfo',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.WalletInfo.fromBuffer(value));
  static final _$getPublicWalletInfo = $grpc.ClientMethod<$2.GetPublicWalletInfoRequest, $2.PublicWalletInfo>(
      '/backend.v1.BackendService/GetPublicWalletInfo',
      ($2.GetPublicWalletInfoRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.PublicWalletInfo.fromBuffer(value));
  static final _$sendPhoneVerification = $grpc.ClientMethod<$2.SendPhoneVerificationRequest, $2.Empty>(
      '/backend.v1.BackendService/SendPhoneVerification',
      ($2.SendPhoneVerificationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$checkPhoneVerification = $grpc.ClientMethod<$2.CheckPhoneVerificationRequest, $2.Empty>(
      '/backend.v1.BackendService/CheckPhoneVerification',
      ($2.CheckPhoneVerificationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$sendOTP = $grpc.ClientMethod<$2.Empty, $2.Empty>(
      '/backend.v1.BackendService/SendOTP',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getAgreement = $grpc.ClientMethod<$2.GetAgreementRequest, $2.Agreement>(
      '/backend.v1.BackendService/GetAgreement',
      ($2.GetAgreementRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Agreement.fromBuffer(value));
  static final _$signAgreements = $grpc.ClientMethod<$2.SignAgreementsRequest, $2.SignAgreementsResponse>(
      '/backend.v1.BackendService/SignAgreements',
      ($2.SignAgreementsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.SignAgreementsResponse.fromBuffer(value));
  static final _$getLinkedAccounts = $grpc.ClientMethod<$2.Empty, $2.GetLinkedAccountsResponse>(
      '/backend.v1.BackendService/GetLinkedAccounts',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetLinkedAccountsResponse.fromBuffer(value));
  static final _$getLinkedAccount = $grpc.ClientMethod<$2.GetLinkedAccountRequest, $2.LinkedAccount>(
      '/backend.v1.BackendService/GetLinkedAccount',
      ($2.GetLinkedAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$setDefaultReceiveLinkedAccount = $grpc.ClientMethod<$2.SetDefaultReceiveLinkedAccountRequest, $2.LinkedAccount>(
      '/backend.v1.BackendService/SetDefaultReceiveLinkedAccount',
      ($2.SetDefaultReceiveLinkedAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$setDefaultSendLinkedAccount = $grpc.ClientMethod<$2.SetDefaultSendLinkedAccountRequest, $2.LinkedAccount>(
      '/backend.v1.BackendService/SetDefaultSendLinkedAccount',
      ($2.SetDefaultSendLinkedAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$setNicknameLinkedAccount = $grpc.ClientMethod<$2.SetNicknameLinkedAccountRequest, $2.LinkedAccount>(
      '/backend.v1.BackendService/SetNicknameLinkedAccount',
      ($2.SetNicknameLinkedAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$deleteLinkedAccount = $grpc.ClientMethod<$2.DeleteLinkedAccountRequest, $2.Empty>(
      '/backend.v1.BackendService/DeleteLinkedAccount',
      ($2.DeleteLinkedAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$createSupportTicket = $grpc.ClientMethod<$2.CreateSupportTicketRequest, $2.Empty>(
      '/backend.v1.BackendService/CreateSupportTicket',
      ($2.CreateSupportTicketRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getCountries = $grpc.ClientMethod<$2.Empty, $2.GetCountriesResponse>(
      '/backend.v1.BackendService/GetCountries',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetCountriesResponse.fromBuffer(value));
  static final _$getCurrentWallet = $grpc.ClientMethod<$2.Empty, $2.GetCurrentWalletResponse>(
      '/backend.v1.BackendService/GetCurrentWallet',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetCurrentWalletResponse.fromBuffer(value));
  static final _$joinWaitlist = $grpc.ClientMethod<$2.JoinWaitlistRequest, $2.JoinWaitlistResponse>(
      '/backend.v1.BackendService/JoinWaitlist',
      ($2.JoinWaitlistRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.JoinWaitlistResponse.fromBuffer(value));
  static final _$canSignup = $grpc.ClientMethod<$2.CanSignupRequest, $2.CanSignupResponse>(
      '/backend.v1.BackendService/CanSignup',
      ($2.CanSignupRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.CanSignupResponse.fromBuffer(value));
  static final _$setSignupComplete = $grpc.ClientMethod<$2.SetSignupCompleteRequest, $2.Empty>(
      '/backend.v1.BackendService/SetSignupComplete',
      ($2.SetSignupCompleteRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$isMugAvailable = $grpc.ClientMethod<$2.IsMugAvailableRequest, $2.IsMugAvailableResponse>(
      '/backend.v1.BackendService/IsMugAvailable',
      ($2.IsMugAvailableRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.IsMugAvailableResponse.fromBuffer(value));
  static final _$listTransactions = $grpc.ClientMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
      '/backend.v1.BackendService/ListTransactions',
      ($2.PaginationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListTransactionsResponse.fromBuffer(value));
  static final _$listTransactionsCompleted = $grpc.ClientMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
      '/backend.v1.BackendService/ListTransactionsCompleted',
      ($2.PaginationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListTransactionsResponse.fromBuffer(value));
  static final _$listTransactionsWithPending = $grpc.ClientMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
      '/backend.v1.BackendService/ListTransactionsWithPending',
      ($2.PaginationRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListTransactionsResponse.fromBuffer(value));
  static final _$lookupTransaction = $grpc.ClientMethod<$2.LookupTransactionRequest, $2.Transaction>(
      '/backend.v1.BackendService/LookupTransaction',
      ($2.LookupTransactionRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Transaction.fromBuffer(value));
  static final _$listLimits = $grpc.ClientMethod<$2.Empty, $2.ListLimitsResponse>(
      '/backend.v1.BackendService/ListLimits',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListLimitsResponse.fromBuffer(value));
  static final _$updateClientLimits = $grpc.ClientMethod<$2.UpdateClientLimitsRequest, $2.Empty>(
      '/backend.v1.BackendService/UpdateClientLimits',
      ($2.UpdateClientLimitsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$createConnection = $grpc.ClientMethod<$2.CreateConnectionRequest, $2.Empty>(
      '/backend.v1.BackendService/CreateConnection',
      ($2.CreateConnectionRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$listConnections = $grpc.ClientMethod<$2.Empty, $2.ListConnectionsResponse>(
      '/backend.v1.BackendService/ListConnections',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListConnectionsResponse.fromBuffer(value));
  static final _$getConnection = $grpc.ClientMethod<$2.GetConnectionRequest, $2.Connection>(
      '/backend.v1.BackendService/GetConnection',
      ($2.GetConnectionRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Connection.fromBuffer(value));
  static final _$getConnectionLimits = $grpc.ClientMethod<$2.GetConnectionLimitsRequest, $2.ConnectionLimits>(
      '/backend.v1.BackendService/GetConnectionLimits',
      ($2.GetConnectionLimitsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ConnectionLimits.fromBuffer(value));
  static final _$updateConnectionLimits = $grpc.ClientMethod<$2.UpdateConnectionLimitsRequest, $2.Empty>(
      '/backend.v1.BackendService/UpdateConnectionLimits',
      ($2.UpdateConnectionLimitsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$deleteConnection = $grpc.ClientMethod<$2.DeleteConnectionRequest, $2.Empty>(
      '/backend.v1.BackendService/DeleteConnection',
      ($2.DeleteConnectionRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getPublicWalletDetails = $grpc.ClientMethod<$2.GetPublicWalletDetailsRequest, $2.GetPublicWalletDetailsResponse>(
      '/backend.v1.BackendService/GetPublicWalletDetails',
      ($2.GetPublicWalletDetailsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetPublicWalletDetailsResponse.fromBuffer(value));
  static final _$createContact = $grpc.ClientMethod<$2.CreateContactRequest, $2.Contact>(
      '/backend.v1.BackendService/CreateContact',
      ($2.CreateContactRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Contact.fromBuffer(value));
  static final _$listContacts = $grpc.ClientMethod<$2.ListContactsRequest, $2.ListContactsResponse>(
      '/backend.v1.BackendService/ListContacts',
      ($2.ListContactsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListContactsResponse.fromBuffer(value));
  static final _$listIdentities = $grpc.ClientMethod<$2.Empty, $2.ListIdentitiesResponse>(
      '/backend.v1.BackendService/ListIdentities',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListIdentitiesResponse.fromBuffer(value));
  static final _$listPublicIdentities = $grpc.ClientMethod<$2.ListPublicIdentitiesRequest, $2.ListIdentitiesResponse>(
      '/backend.v1.BackendService/ListPublicIdentities',
      ($2.ListPublicIdentitiesRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.ListIdentitiesResponse.fromBuffer(value));
  static final _$deleteIdentity = $grpc.ClientMethod<$2.DeleteIdentityRequest, $2.Empty>(
      '/backend.v1.BackendService/DeleteIdentity',
      ($2.DeleteIdentityRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$setIdentityPublic = $grpc.ClientMethod<$2.SetIdentityPublicRequest, $2.Identity>(
      '/backend.v1.BackendService/SetIdentityPublic',
      ($2.SetIdentityPublicRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Identity.fromBuffer(value));
  static final _$getIdentity = $grpc.ClientMethod<$2.GetIdentityRequest, $2.GetIdentityResponse>(
      '/backend.v1.BackendService/GetIdentity',
      ($2.GetIdentityRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetIdentityResponse.fromBuffer(value));
  static final _$getIdentityBySignatureHash = $grpc.ClientMethod<$2.GetIdentityBySignatureHashRequest, $2.GetIdentityResponse>(
      '/backend.v1.BackendService/GetIdentityBySignatureHash',
      ($2.GetIdentityBySignatureHashRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetIdentityResponse.fromBuffer(value));
  static final _$verifyIdentity = $grpc.ClientMethod<$2.VerifyIdentityRequest, $2.Empty>(
      '/backend.v1.BackendService/VerifyIdentity',
      ($2.VerifyIdentityRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$kYCStatus = $grpc.ClientMethod<$2.Empty, $2.KYCStatusResponse>(
      '/backend.v1.BackendService/KYCStatus',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.KYCStatusResponse.fromBuffer(value));
  static final _$setKYCStatusPending = $grpc.ClientMethod<$2.Empty, $2.Empty>(
      '/backend.v1.BackendService/SetKYCStatusPending',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$startKYC = $grpc.ClientMethod<$2.Empty, $2.Empty>(
      '/backend.v1.BackendService/StartKYC',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$getPersonaInquiry = $grpc.ClientMethod<$2.KYCPersonaInquiryRequest, $2.KYCPersonaInquiryResponse>(
      '/backend.v1.BackendService/GetPersonaInquiry',
      ($2.KYCPersonaInquiryRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.KYCPersonaInquiryResponse.fromBuffer(value));
  static final _$getMXWidget = $grpc.ClientMethod<$2.Empty, $2.MXWidgetResponse>(
      '/backend.v1.BackendService/GetMXWidget',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.MXWidgetResponse.fromBuffer(value));
  static final _$createMXBankAccounts = $grpc.ClientMethod<$2.CreateMXBankAccountsRequest, $2.CreateMXBankAccountsResponse>(
      '/backend.v1.BackendService/CreateMXBankAccounts',
      ($2.CreateMXBankAccountsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.CreateMXBankAccountsResponse.fromBuffer(value));
  static final _$onboardGMTUser = $grpc.ClientMethod<$2.Empty, $2.Empty>(
      '/backend.v1.BackendService/OnboardGMTUser',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$init3DS = $grpc.ClientMethod<$2.Init3DSRequest, $2.Init3DSResponse>(
      '/backend.v1.BackendService/Init3DS',
      ($2.Init3DSRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Init3DSResponse.fromBuffer(value));
  static final _$lookup3DS = $grpc.ClientMethod<$2.Lookup3DSRequest, $2.Lookup3DSResponse>(
      '/backend.v1.BackendService/Lookup3DS',
      ($2.Lookup3DSRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Lookup3DSResponse.fromBuffer(value));
  static final _$authenticate3DS = $grpc.ClientMethod<$2.Authenticate3DSRequest, $2.Authenticate3DSResponse>(
      '/backend.v1.BackendService/Authenticate3DS',
      ($2.Authenticate3DSRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Authenticate3DSResponse.fromBuffer(value));
  static final _$createCard = $grpc.ClientMethod<$2.CreateCardRequest, $2.LinkedAccount>(
      '/backend.v1.BackendService/CreateCard',
      ($2.CreateCardRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$getCardDetails = $grpc.ClientMethod<$2.GetCardDetailsRequest, $2.CardDetails>(
      '/backend.v1.BackendService/GetCardDetails',
      ($2.GetCardDetailsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.CardDetails.fromBuffer(value));
  static final _$listFeatures = $grpc.ClientMethod<$2.Empty, $2.Features>(
      '/backend.v1.BackendService/ListFeatures',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Features.fromBuffer(value));
  static final _$createTwitterAuthURL = $grpc.ClientMethod<$2.Empty, $2.CreateTwitterAuthURLResponse>(
      '/backend.v1.BackendService/CreateTwitterAuthURL',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.CreateTwitterAuthURLResponse.fromBuffer(value));
  static final _$twitterCallback = $grpc.ClientMethod<$2.TwitterCallbackRequest, $2.TwitterCallbackResponse>(
      '/backend.v1.BackendService/TwitterCallback',
      ($2.TwitterCallbackRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.TwitterCallbackResponse.fromBuffer(value));
  static final _$createDomainIdentity = $grpc.ClientMethod<$2.CreateDomainIdentityRequest, $2.CreateDomainIdentityResponse>(
      '/backend.v1.BackendService/CreateDomainIdentity',
      ($2.CreateDomainIdentityRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.CreateDomainIdentityResponse.fromBuffer(value));
  static final _$getPaymentAddress = $grpc.ClientMethod<$2.GetPaymentAddressRequest, $2.GetPaymentAddressResponse>(
      '/backend.v1.BackendService/GetPaymentAddress',
      ($2.GetPaymentAddressRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetPaymentAddressResponse.fromBuffer(value));
  static final _$createPayment = $grpc.ClientMethod<$2.CreatePaymentRequest, $2.Payment>(
      '/backend.v1.BackendService/CreatePayment',
      ($2.CreatePaymentRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Payment.fromBuffer(value));
  static final _$updatePayment = $grpc.ClientMethod<$2.UpdatePaymentRequest, $2.Payment>(
      '/backend.v1.BackendService/UpdatePayment',
      ($2.UpdatePaymentRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Payment.fromBuffer(value));
  static final _$getPayment = $grpc.ClientMethod<$2.GetPaymentRequest, $2.Payment>(
      '/backend.v1.BackendService/GetPayment',
      ($2.GetPaymentRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Payment.fromBuffer(value));
  static final _$confirmPayment = $grpc.ClientMethod<$2.ConfirmPaymentRequest, $2.Payment>(
      '/backend.v1.BackendService/ConfirmPayment',
      ($2.ConfirmPaymentRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Payment.fromBuffer(value));
  static final _$getLinkedAccountsForPayment = $grpc.ClientMethod<$2.GetLinkedAccountsForPaymentRequest, $2.GetLinkedAccountsForPaymentResponse>(
      '/backend.v1.BackendService/GetLinkedAccountsForPayment',
      ($2.GetLinkedAccountsForPaymentRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetLinkedAccountsForPaymentResponse.fromBuffer(value));
  static final _$searchWallets = $grpc.ClientMethod<$2.SearchWalletsRequest, $2.SearchWalletsResponse>(
      '/backend.v1.BackendService/SearchWallets',
      ($2.SearchWalletsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.SearchWalletsResponse.fromBuffer(value));
  static final _$discordCallback = $grpc.ClientMethod<$2.DiscordCallbackRequest, $2.DiscordCallbackResponse>(
      '/backend.v1.BackendService/DiscordCallback',
      ($2.DiscordCallbackRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.DiscordCallbackResponse.fromBuffer(value));
  static final _$createDiscordAuthURL = $grpc.ClientMethod<$2.Empty, $2.CreateDiscordAuthURLResponse>(
      '/backend.v1.BackendService/CreateDiscordAuthURL',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.CreateDiscordAuthURLResponse.fromBuffer(value));
  static final _$submitForm = $grpc.ClientMethod<$2.SubmitFormRequest, $2.Empty>(
      '/backend.v1.BackendService/SubmitForm',
      ($2.SubmitFormRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Empty.fromBuffer(value));
  static final _$createSlackAuthURL = $grpc.ClientMethod<$2.Empty, $2.CreateSlackAuthURLResponse>(
      '/backend.v1.BackendService/CreateSlackAuthURL',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.CreateSlackAuthURLResponse.fromBuffer(value));
  static final _$slackCallback = $grpc.ClientMethod<$2.SlackCallbackRequest, $2.SlackCallbackResponse>(
      '/backend.v1.BackendService/SlackCallback',
      ($2.SlackCallbackRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.SlackCallbackResponse.fromBuffer(value));
  static final _$addXagoBankAccount = $grpc.ClientMethod<$2.AddXagoBankAccountRequest, $2.LinkedAccount>(
      '/backend.v1.BackendService/AddXagoBankAccount',
      ($2.AddXagoBankAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$addXagoBalanceAccount = $grpc.ClientMethod<$2.AddXagoBalanceAccountRequest, $2.LinkedAccount>(
      '/backend.v1.BackendService/AddXagoBalanceAccount',
      ($2.AddXagoBalanceAccountRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.LinkedAccount.fromBuffer(value));
  static final _$withdrawXagoBalance = $grpc.ClientMethod<$2.WithdrawXagoBalanceRequest, $2.Payment>(
      '/backend.v1.BackendService/WithdrawXagoBalance',
      ($2.WithdrawXagoBalanceRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.Payment.fromBuffer(value));
  static final _$getXagoBalances = $grpc.ClientMethod<$2.Empty, $2.GetXagoBalanceResponse>(
      '/backend.v1.BackendService/GetXagoBalances',
      ($2.Empty value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetXagoBalanceResponse.fromBuffer(value));
  static final _$getXagoDepositDetails = $grpc.ClientMethod<$2.GetXagoDepositDetailsRequest, $2.GetXagoDepositDetailsResponse>(
      '/backend.v1.BackendService/GetXagoDepositDetails',
      ($2.GetXagoDepositDetailsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $2.GetXagoDepositDetailsResponse.fromBuffer(value));

  BackendServiceClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options,
        interceptors: interceptors);

  $grpc.ResponseFuture<$2.Empty> updateIndividualKYC($2.UpdateIndividualKYCRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$updateIndividualKYC, request, options: options);
  }

  $grpc.ResponseFuture<$2.IndividualKYCResponse> getIndividualKYC($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getIndividualKYC, request, options: options);
  }

  $grpc.ResponseFuture<$2.IsUSPSAddressResponse> isUSPSAddress($2.Address request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$isUSPSAddress, request, options: options);
  }

  $grpc.ResponseFuture<$2.SetSignupUserDataResponse> setSignupUserData($2.SetSignupUserDataRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setSignupUserData, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> setSignupMobileNumber($2.SetSignupMobileNumberRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setSignupMobileNumber, request, options: options);
  }

  $grpc.ResponseFuture<$2.Signup> getSignup($2.GetSignupRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getSignup, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> completeSignup($2.CompleteSignupRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$completeSignup, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> createUserDefaultWallet($2.CreateUserDefaultWalletRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createUserDefaultWallet, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> createWalletAddress($2.CreateWalletAddressRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createWalletAddress, request, options: options);
  }

  $grpc.ResponseFuture<$2.WalletAddressExistsResponse> walletAddressExists($2.WalletAddressExistsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$walletAddressExists, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> setWalletName($2.SetWalletNameRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setWalletName, request, options: options);
  }

  $grpc.ResponseFuture<$2.WalletInfo> getWalletInfo($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWalletInfo, request, options: options);
  }

  $grpc.ResponseFuture<$2.PublicWalletInfo> getPublicWalletInfo($2.GetPublicWalletInfoRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getPublicWalletInfo, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> sendPhoneVerification($2.SendPhoneVerificationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$sendPhoneVerification, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> checkPhoneVerification($2.CheckPhoneVerificationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$checkPhoneVerification, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> sendOTP($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$sendOTP, request, options: options);
  }

  $grpc.ResponseFuture<$2.Agreement> getAgreement($2.GetAgreementRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getAgreement, request, options: options);
  }

  $grpc.ResponseFuture<$2.SignAgreementsResponse> signAgreements($2.SignAgreementsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$signAgreements, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetLinkedAccountsResponse> getLinkedAccounts($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLinkedAccounts, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> getLinkedAccount($2.GetLinkedAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> setDefaultReceiveLinkedAccount($2.SetDefaultReceiveLinkedAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setDefaultReceiveLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> setDefaultSendLinkedAccount($2.SetDefaultSendLinkedAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setDefaultSendLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> setNicknameLinkedAccount($2.SetNicknameLinkedAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setNicknameLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> deleteLinkedAccount($2.DeleteLinkedAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$deleteLinkedAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> createSupportTicket($2.CreateSupportTicketRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createSupportTicket, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetCountriesResponse> getCountries($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getCountries, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetCurrentWalletResponse> getCurrentWallet($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getCurrentWallet, request, options: options);
  }

  $grpc.ResponseFuture<$2.JoinWaitlistResponse> joinWaitlist($2.JoinWaitlistRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$joinWaitlist, request, options: options);
  }

  $grpc.ResponseFuture<$2.CanSignupResponse> canSignup($2.CanSignupRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$canSignup, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> setSignupComplete($2.SetSignupCompleteRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setSignupComplete, request, options: options);
  }

  $grpc.ResponseFuture<$2.IsMugAvailableResponse> isMugAvailable($2.IsMugAvailableRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$isMugAvailable, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListTransactionsResponse> listTransactions($2.PaginationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listTransactions, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListTransactionsResponse> listTransactionsCompleted($2.PaginationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listTransactionsCompleted, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListTransactionsResponse> listTransactionsWithPending($2.PaginationRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listTransactionsWithPending, request, options: options);
  }

  $grpc.ResponseFuture<$2.Transaction> lookupTransaction($2.LookupTransactionRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$lookupTransaction, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListLimitsResponse> listLimits($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listLimits, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> updateClientLimits($2.UpdateClientLimitsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$updateClientLimits, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> createConnection($2.CreateConnectionRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createConnection, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListConnectionsResponse> listConnections($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listConnections, request, options: options);
  }

  $grpc.ResponseFuture<$2.Connection> getConnection($2.GetConnectionRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getConnection, request, options: options);
  }

  $grpc.ResponseFuture<$2.ConnectionLimits> getConnectionLimits($2.GetConnectionLimitsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getConnectionLimits, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> updateConnectionLimits($2.UpdateConnectionLimitsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$updateConnectionLimits, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> deleteConnection($2.DeleteConnectionRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$deleteConnection, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetPublicWalletDetailsResponse> getPublicWalletDetails($2.GetPublicWalletDetailsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getPublicWalletDetails, request, options: options);
  }

  $grpc.ResponseFuture<$2.Contact> createContact($2.CreateContactRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createContact, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListContactsResponse> listContacts($2.ListContactsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listContacts, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListIdentitiesResponse> listIdentities($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listIdentities, request, options: options);
  }

  $grpc.ResponseFuture<$2.ListIdentitiesResponse> listPublicIdentities($2.ListPublicIdentitiesRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listPublicIdentities, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> deleteIdentity($2.DeleteIdentityRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$deleteIdentity, request, options: options);
  }

  $grpc.ResponseFuture<$2.Identity> setIdentityPublic($2.SetIdentityPublicRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setIdentityPublic, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetIdentityResponse> getIdentity($2.GetIdentityRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getIdentity, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetIdentityResponse> getIdentityBySignatureHash($2.GetIdentityBySignatureHashRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getIdentityBySignatureHash, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> verifyIdentity($2.VerifyIdentityRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$verifyIdentity, request, options: options);
  }

  $grpc.ResponseFuture<$2.KYCStatusResponse> kYCStatus($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$kYCStatus, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> setKYCStatusPending($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$setKYCStatusPending, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> startKYC($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$startKYC, request, options: options);
  }

  $grpc.ResponseFuture<$2.KYCPersonaInquiryResponse> getPersonaInquiry($2.KYCPersonaInquiryRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getPersonaInquiry, request, options: options);
  }

  $grpc.ResponseFuture<$2.MXWidgetResponse> getMXWidget($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getMXWidget, request, options: options);
  }

  $grpc.ResponseFuture<$2.CreateMXBankAccountsResponse> createMXBankAccounts($2.CreateMXBankAccountsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createMXBankAccounts, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> onboardGMTUser($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$onboardGMTUser, request, options: options);
  }

  $grpc.ResponseFuture<$2.Init3DSResponse> init3DS($2.Init3DSRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$init3DS, request, options: options);
  }

  $grpc.ResponseFuture<$2.Lookup3DSResponse> lookup3DS($2.Lookup3DSRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$lookup3DS, request, options: options);
  }

  $grpc.ResponseFuture<$2.Authenticate3DSResponse> authenticate3DS($2.Authenticate3DSRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$authenticate3DS, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> createCard($2.CreateCardRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createCard, request, options: options);
  }

  $grpc.ResponseFuture<$2.CardDetails> getCardDetails($2.GetCardDetailsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getCardDetails, request, options: options);
  }

  $grpc.ResponseFuture<$2.Features> listFeatures($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$listFeatures, request, options: options);
  }

  $grpc.ResponseFuture<$2.CreateTwitterAuthURLResponse> createTwitterAuthURL($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createTwitterAuthURL, request, options: options);
  }

  $grpc.ResponseFuture<$2.TwitterCallbackResponse> twitterCallback($2.TwitterCallbackRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$twitterCallback, request, options: options);
  }

  $grpc.ResponseFuture<$2.CreateDomainIdentityResponse> createDomainIdentity($2.CreateDomainIdentityRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createDomainIdentity, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetPaymentAddressResponse> getPaymentAddress($2.GetPaymentAddressRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getPaymentAddress, request, options: options);
  }

  $grpc.ResponseFuture<$2.Payment> createPayment($2.CreatePaymentRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.Payment> updatePayment($2.UpdatePaymentRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$updatePayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.Payment> getPayment($2.GetPaymentRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.Payment> confirmPayment($2.ConfirmPaymentRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$confirmPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetLinkedAccountsForPaymentResponse> getLinkedAccountsForPayment($2.GetLinkedAccountsForPaymentRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getLinkedAccountsForPayment, request, options: options);
  }

  $grpc.ResponseFuture<$2.SearchWalletsResponse> searchWallets($2.SearchWalletsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$searchWallets, request, options: options);
  }

  $grpc.ResponseFuture<$2.DiscordCallbackResponse> discordCallback($2.DiscordCallbackRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$discordCallback, request, options: options);
  }

  $grpc.ResponseFuture<$2.CreateDiscordAuthURLResponse> createDiscordAuthURL($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createDiscordAuthURL, request, options: options);
  }

  $grpc.ResponseFuture<$2.Empty> submitForm($2.SubmitFormRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$submitForm, request, options: options);
  }

  $grpc.ResponseFuture<$2.CreateSlackAuthURLResponse> createSlackAuthURL($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createSlackAuthURL, request, options: options);
  }

  $grpc.ResponseFuture<$2.SlackCallbackResponse> slackCallback($2.SlackCallbackRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$slackCallback, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> addXagoBankAccount($2.AddXagoBankAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$addXagoBankAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.LinkedAccount> addXagoBalanceAccount($2.AddXagoBalanceAccountRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$addXagoBalanceAccount, request, options: options);
  }

  $grpc.ResponseFuture<$2.Payment> withdrawXagoBalance($2.WithdrawXagoBalanceRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$withdrawXagoBalance, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetXagoBalanceResponse> getXagoBalances($2.Empty request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getXagoBalances, request, options: options);
  }

  $grpc.ResponseFuture<$2.GetXagoDepositDetailsResponse> getXagoDepositDetails($2.GetXagoDepositDetailsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getXagoDepositDetails, request, options: options);
  }
}

@$pb.GrpcServiceName('backend.v1.BackendService')
abstract class BackendServiceBase extends $grpc.Service {
  $core.String get $name => 'backend.v1.BackendService';

  BackendServiceBase() {
    $addMethod($grpc.ServiceMethod<$2.UpdateIndividualKYCRequest, $2.Empty>(
        'UpdateIndividualKYC',
        updateIndividualKYC_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.UpdateIndividualKYCRequest.fromBuffer(value),
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
    $addMethod($grpc.ServiceMethod<$2.SetSignupUserDataRequest, $2.SetSignupUserDataResponse>(
        'SetSignupUserData',
        setSignupUserData_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SetSignupUserDataRequest.fromBuffer(value),
        ($2.SetSignupUserDataResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetSignupMobileNumberRequest, $2.Empty>(
        'SetSignupMobileNumber',
        setSignupMobileNumber_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SetSignupMobileNumberRequest.fromBuffer(value),
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
        ($core.List<$core.int> value) => $2.CompleteSignupRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateUserDefaultWalletRequest, $2.Empty>(
        'CreateUserDefaultWallet',
        createUserDefaultWallet_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateUserDefaultWalletRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateWalletAddressRequest, $2.Empty>(
        'CreateWalletAddress',
        createWalletAddress_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateWalletAddressRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.WalletAddressExistsRequest, $2.WalletAddressExistsResponse>(
        'WalletAddressExists',
        walletAddressExists_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.WalletAddressExistsRequest.fromBuffer(value),
        ($2.WalletAddressExistsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetWalletNameRequest, $2.Empty>(
        'SetWalletName',
        setWalletName_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SetWalletNameRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.WalletInfo>(
        'GetWalletInfo',
        getWalletInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.WalletInfo value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetPublicWalletInfoRequest, $2.PublicWalletInfo>(
        'GetPublicWalletInfo',
        getPublicWalletInfo_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetPublicWalletInfoRequest.fromBuffer(value),
        ($2.PublicWalletInfo value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SendPhoneVerificationRequest, $2.Empty>(
        'SendPhoneVerification',
        sendPhoneVerification_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SendPhoneVerificationRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CheckPhoneVerificationRequest, $2.Empty>(
        'CheckPhoneVerification',
        checkPhoneVerification_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CheckPhoneVerificationRequest.fromBuffer(value),
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
        ($core.List<$core.int> value) => $2.GetAgreementRequest.fromBuffer(value),
        ($2.Agreement value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SignAgreementsRequest, $2.SignAgreementsResponse>(
        'SignAgreements',
        signAgreements_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SignAgreementsRequest.fromBuffer(value),
        ($2.SignAgreementsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.GetLinkedAccountsResponse>(
        'GetLinkedAccounts',
        getLinkedAccounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.GetLinkedAccountsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetLinkedAccountRequest, $2.LinkedAccount>(
        'GetLinkedAccount',
        getLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetLinkedAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetDefaultReceiveLinkedAccountRequest, $2.LinkedAccount>(
        'SetDefaultReceiveLinkedAccount',
        setDefaultReceiveLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SetDefaultReceiveLinkedAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetDefaultSendLinkedAccountRequest, $2.LinkedAccount>(
        'SetDefaultSendLinkedAccount',
        setDefaultSendLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SetDefaultSendLinkedAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetNicknameLinkedAccountRequest, $2.LinkedAccount>(
        'SetNicknameLinkedAccount',
        setNicknameLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SetNicknameLinkedAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.DeleteLinkedAccountRequest, $2.Empty>(
        'DeleteLinkedAccount',
        deleteLinkedAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.DeleteLinkedAccountRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateSupportTicketRequest, $2.Empty>(
        'CreateSupportTicket',
        createSupportTicket_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateSupportTicketRequest.fromBuffer(value),
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
    $addMethod($grpc.ServiceMethod<$2.JoinWaitlistRequest, $2.JoinWaitlistResponse>(
        'JoinWaitlist',
        joinWaitlist_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.JoinWaitlistRequest.fromBuffer(value),
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
        ($core.List<$core.int> value) => $2.SetSignupCompleteRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.IsMugAvailableRequest, $2.IsMugAvailableResponse>(
        'IsMugAvailable',
        isMugAvailable_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.IsMugAvailableRequest.fromBuffer(value),
        ($2.IsMugAvailableResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
        'ListTransactions',
        listTransactions_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.PaginationRequest.fromBuffer(value),
        ($2.ListTransactionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
        'ListTransactionsCompleted',
        listTransactionsCompleted_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.PaginationRequest.fromBuffer(value),
        ($2.ListTransactionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.PaginationRequest, $2.ListTransactionsResponse>(
        'ListTransactionsWithPending',
        listTransactionsWithPending_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.PaginationRequest.fromBuffer(value),
        ($2.ListTransactionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.LookupTransactionRequest, $2.Transaction>(
        'LookupTransaction',
        lookupTransaction_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.LookupTransactionRequest.fromBuffer(value),
        ($2.Transaction value) => value.writeToBuffer()));
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
        ($core.List<$core.int> value) => $2.UpdateClientLimitsRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateConnectionRequest, $2.Empty>(
        'CreateConnection',
        createConnection_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateConnectionRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.ListConnectionsResponse>(
        'ListConnections',
        listConnections_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.ListConnectionsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetConnectionRequest, $2.Connection>(
        'GetConnection',
        getConnection_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetConnectionRequest.fromBuffer(value),
        ($2.Connection value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetConnectionLimitsRequest, $2.ConnectionLimits>(
        'GetConnectionLimits',
        getConnectionLimits_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetConnectionLimitsRequest.fromBuffer(value),
        ($2.ConnectionLimits value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.UpdateConnectionLimitsRequest, $2.Empty>(
        'UpdateConnectionLimits',
        updateConnectionLimits_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.UpdateConnectionLimitsRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.DeleteConnectionRequest, $2.Empty>(
        'DeleteConnection',
        deleteConnection_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.DeleteConnectionRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetPublicWalletDetailsRequest, $2.GetPublicWalletDetailsResponse>(
        'GetPublicWalletDetails',
        getPublicWalletDetails_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetPublicWalletDetailsRequest.fromBuffer(value),
        ($2.GetPublicWalletDetailsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateContactRequest, $2.Contact>(
        'CreateContact',
        createContact_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateContactRequest.fromBuffer(value),
        ($2.Contact value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.ListContactsRequest, $2.ListContactsResponse>(
        'ListContacts',
        listContacts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.ListContactsRequest.fromBuffer(value),
        ($2.ListContactsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.ListIdentitiesResponse>(
        'ListIdentities',
        listIdentities_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.ListIdentitiesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.ListPublicIdentitiesRequest, $2.ListIdentitiesResponse>(
        'ListPublicIdentities',
        listPublicIdentities_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.ListPublicIdentitiesRequest.fromBuffer(value),
        ($2.ListIdentitiesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.DeleteIdentityRequest, $2.Empty>(
        'DeleteIdentity',
        deleteIdentity_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.DeleteIdentityRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SetIdentityPublicRequest, $2.Identity>(
        'SetIdentityPublic',
        setIdentityPublic_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SetIdentityPublicRequest.fromBuffer(value),
        ($2.Identity value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetIdentityRequest, $2.GetIdentityResponse>(
        'GetIdentity',
        getIdentity_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetIdentityRequest.fromBuffer(value),
        ($2.GetIdentityResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetIdentityBySignatureHashRequest, $2.GetIdentityResponse>(
        'GetIdentityBySignatureHash',
        getIdentityBySignatureHash_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetIdentityBySignatureHashRequest.fromBuffer(value),
        ($2.GetIdentityResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.VerifyIdentityRequest, $2.Empty>(
        'VerifyIdentity',
        verifyIdentity_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.VerifyIdentityRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.KYCStatusResponse>(
        'KYCStatus',
        kYCStatus_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.KYCStatusResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.Empty>(
        'SetKYCStatusPending',
        setKYCStatusPending_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.Empty>(
        'StartKYC',
        startKYC_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.KYCPersonaInquiryRequest, $2.KYCPersonaInquiryResponse>(
        'GetPersonaInquiry',
        getPersonaInquiry_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.KYCPersonaInquiryRequest.fromBuffer(value),
        ($2.KYCPersonaInquiryResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.MXWidgetResponse>(
        'GetMXWidget',
        getMXWidget_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.MXWidgetResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateMXBankAccountsRequest, $2.CreateMXBankAccountsResponse>(
        'CreateMXBankAccounts',
        createMXBankAccounts_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateMXBankAccountsRequest.fromBuffer(value),
        ($2.CreateMXBankAccountsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.Empty>(
        'OnboardGMTUser',
        onboardGMTUser_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Init3DSRequest, $2.Init3DSResponse>(
        'Init3DS',
        init3DS_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Init3DSRequest.fromBuffer(value),
        ($2.Init3DSResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Lookup3DSRequest, $2.Lookup3DSResponse>(
        'Lookup3DS',
        lookup3DS_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Lookup3DSRequest.fromBuffer(value),
        ($2.Lookup3DSResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Authenticate3DSRequest, $2.Authenticate3DSResponse>(
        'Authenticate3DS',
        authenticate3DS_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Authenticate3DSRequest.fromBuffer(value),
        ($2.Authenticate3DSResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateCardRequest, $2.LinkedAccount>(
        'CreateCard',
        createCard_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateCardRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetCardDetailsRequest, $2.CardDetails>(
        'GetCardDetails',
        getCardDetails_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetCardDetailsRequest.fromBuffer(value),
        ($2.CardDetails value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.Features>(
        'ListFeatures',
        listFeatures_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.Features value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.CreateTwitterAuthURLResponse>(
        'CreateTwitterAuthURL',
        createTwitterAuthURL_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.CreateTwitterAuthURLResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.TwitterCallbackRequest, $2.TwitterCallbackResponse>(
        'TwitterCallback',
        twitterCallback_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.TwitterCallbackRequest.fromBuffer(value),
        ($2.TwitterCallbackResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreateDomainIdentityRequest, $2.CreateDomainIdentityResponse>(
        'CreateDomainIdentity',
        createDomainIdentity_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreateDomainIdentityRequest.fromBuffer(value),
        ($2.CreateDomainIdentityResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetPaymentAddressRequest, $2.GetPaymentAddressResponse>(
        'GetPaymentAddress',
        getPaymentAddress_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetPaymentAddressRequest.fromBuffer(value),
        ($2.GetPaymentAddressResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.CreatePaymentRequest, $2.Payment>(
        'CreatePayment',
        createPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.CreatePaymentRequest.fromBuffer(value),
        ($2.Payment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.UpdatePaymentRequest, $2.Payment>(
        'UpdatePayment',
        updatePayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.UpdatePaymentRequest.fromBuffer(value),
        ($2.Payment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetPaymentRequest, $2.Payment>(
        'GetPayment',
        getPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetPaymentRequest.fromBuffer(value),
        ($2.Payment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.ConfirmPaymentRequest, $2.Payment>(
        'ConfirmPayment',
        confirmPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.ConfirmPaymentRequest.fromBuffer(value),
        ($2.Payment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetLinkedAccountsForPaymentRequest, $2.GetLinkedAccountsForPaymentResponse>(
        'GetLinkedAccountsForPayment',
        getLinkedAccountsForPayment_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetLinkedAccountsForPaymentRequest.fromBuffer(value),
        ($2.GetLinkedAccountsForPaymentResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SearchWalletsRequest, $2.SearchWalletsResponse>(
        'SearchWallets',
        searchWallets_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SearchWalletsRequest.fromBuffer(value),
        ($2.SearchWalletsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.DiscordCallbackRequest, $2.DiscordCallbackResponse>(
        'DiscordCallback',
        discordCallback_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.DiscordCallbackRequest.fromBuffer(value),
        ($2.DiscordCallbackResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.CreateDiscordAuthURLResponse>(
        'CreateDiscordAuthURL',
        createDiscordAuthURL_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.CreateDiscordAuthURLResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SubmitFormRequest, $2.Empty>(
        'SubmitForm',
        submitForm_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SubmitFormRequest.fromBuffer(value),
        ($2.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.CreateSlackAuthURLResponse>(
        'CreateSlackAuthURL',
        createSlackAuthURL_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.CreateSlackAuthURLResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.SlackCallbackRequest, $2.SlackCallbackResponse>(
        'SlackCallback',
        slackCallback_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.SlackCallbackRequest.fromBuffer(value),
        ($2.SlackCallbackResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.AddXagoBankAccountRequest, $2.LinkedAccount>(
        'AddXagoBankAccount',
        addXagoBankAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.AddXagoBankAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.AddXagoBalanceAccountRequest, $2.LinkedAccount>(
        'AddXagoBalanceAccount',
        addXagoBalanceAccount_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.AddXagoBalanceAccountRequest.fromBuffer(value),
        ($2.LinkedAccount value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.WithdrawXagoBalanceRequest, $2.Payment>(
        'WithdrawXagoBalance',
        withdrawXagoBalance_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.WithdrawXagoBalanceRequest.fromBuffer(value),
        ($2.Payment value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.Empty, $2.GetXagoBalanceResponse>(
        'GetXagoBalances',
        getXagoBalances_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.Empty.fromBuffer(value),
        ($2.GetXagoBalanceResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$2.GetXagoDepositDetailsRequest, $2.GetXagoDepositDetailsResponse>(
        'GetXagoDepositDetails',
        getXagoDepositDetails_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $2.GetXagoDepositDetailsRequest.fromBuffer(value),
        ($2.GetXagoDepositDetailsResponse value) => value.writeToBuffer()));
  }

  $async.Future<$2.Empty> updateIndividualKYC_Pre($grpc.ServiceCall call, $async.Future<$2.UpdateIndividualKYCRequest> request) async {
    return updateIndividualKYC(call, await request);
  }

  $async.Future<$2.IndividualKYCResponse> getIndividualKYC_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getIndividualKYC(call, await request);
  }

  $async.Future<$2.IsUSPSAddressResponse> isUSPSAddress_Pre($grpc.ServiceCall call, $async.Future<$2.Address> request) async {
    return isUSPSAddress(call, await request);
  }

  $async.Future<$2.SetSignupUserDataResponse> setSignupUserData_Pre($grpc.ServiceCall call, $async.Future<$2.SetSignupUserDataRequest> request) async {
    return setSignupUserData(call, await request);
  }

  $async.Future<$2.Empty> setSignupMobileNumber_Pre($grpc.ServiceCall call, $async.Future<$2.SetSignupMobileNumberRequest> request) async {
    return setSignupMobileNumber(call, await request);
  }

  $async.Future<$2.Signup> getSignup_Pre($grpc.ServiceCall call, $async.Future<$2.GetSignupRequest> request) async {
    return getSignup(call, await request);
  }

  $async.Future<$2.Empty> completeSignup_Pre($grpc.ServiceCall call, $async.Future<$2.CompleteSignupRequest> request) async {
    return completeSignup(call, await request);
  }

  $async.Future<$2.Empty> createUserDefaultWallet_Pre($grpc.ServiceCall call, $async.Future<$2.CreateUserDefaultWalletRequest> request) async {
    return createUserDefaultWallet(call, await request);
  }

  $async.Future<$2.Empty> createWalletAddress_Pre($grpc.ServiceCall call, $async.Future<$2.CreateWalletAddressRequest> request) async {
    return createWalletAddress(call, await request);
  }

  $async.Future<$2.WalletAddressExistsResponse> walletAddressExists_Pre($grpc.ServiceCall call, $async.Future<$2.WalletAddressExistsRequest> request) async {
    return walletAddressExists(call, await request);
  }

  $async.Future<$2.Empty> setWalletName_Pre($grpc.ServiceCall call, $async.Future<$2.SetWalletNameRequest> request) async {
    return setWalletName(call, await request);
  }

  $async.Future<$2.WalletInfo> getWalletInfo_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getWalletInfo(call, await request);
  }

  $async.Future<$2.PublicWalletInfo> getPublicWalletInfo_Pre($grpc.ServiceCall call, $async.Future<$2.GetPublicWalletInfoRequest> request) async {
    return getPublicWalletInfo(call, await request);
  }

  $async.Future<$2.Empty> sendPhoneVerification_Pre($grpc.ServiceCall call, $async.Future<$2.SendPhoneVerificationRequest> request) async {
    return sendPhoneVerification(call, await request);
  }

  $async.Future<$2.Empty> checkPhoneVerification_Pre($grpc.ServiceCall call, $async.Future<$2.CheckPhoneVerificationRequest> request) async {
    return checkPhoneVerification(call, await request);
  }

  $async.Future<$2.Empty> sendOTP_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return sendOTP(call, await request);
  }

  $async.Future<$2.Agreement> getAgreement_Pre($grpc.ServiceCall call, $async.Future<$2.GetAgreementRequest> request) async {
    return getAgreement(call, await request);
  }

  $async.Future<$2.SignAgreementsResponse> signAgreements_Pre($grpc.ServiceCall call, $async.Future<$2.SignAgreementsRequest> request) async {
    return signAgreements(call, await request);
  }

  $async.Future<$2.GetLinkedAccountsResponse> getLinkedAccounts_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getLinkedAccounts(call, await request);
  }

  $async.Future<$2.LinkedAccount> getLinkedAccount_Pre($grpc.ServiceCall call, $async.Future<$2.GetLinkedAccountRequest> request) async {
    return getLinkedAccount(call, await request);
  }

  $async.Future<$2.LinkedAccount> setDefaultReceiveLinkedAccount_Pre($grpc.ServiceCall call, $async.Future<$2.SetDefaultReceiveLinkedAccountRequest> request) async {
    return setDefaultReceiveLinkedAccount(call, await request);
  }

  $async.Future<$2.LinkedAccount> setDefaultSendLinkedAccount_Pre($grpc.ServiceCall call, $async.Future<$2.SetDefaultSendLinkedAccountRequest> request) async {
    return setDefaultSendLinkedAccount(call, await request);
  }

  $async.Future<$2.LinkedAccount> setNicknameLinkedAccount_Pre($grpc.ServiceCall call, $async.Future<$2.SetNicknameLinkedAccountRequest> request) async {
    return setNicknameLinkedAccount(call, await request);
  }

  $async.Future<$2.Empty> deleteLinkedAccount_Pre($grpc.ServiceCall call, $async.Future<$2.DeleteLinkedAccountRequest> request) async {
    return deleteLinkedAccount(call, await request);
  }

  $async.Future<$2.Empty> createSupportTicket_Pre($grpc.ServiceCall call, $async.Future<$2.CreateSupportTicketRequest> request) async {
    return createSupportTicket(call, await request);
  }

  $async.Future<$2.GetCountriesResponse> getCountries_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getCountries(call, await request);
  }

  $async.Future<$2.GetCurrentWalletResponse> getCurrentWallet_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getCurrentWallet(call, await request);
  }

  $async.Future<$2.JoinWaitlistResponse> joinWaitlist_Pre($grpc.ServiceCall call, $async.Future<$2.JoinWaitlistRequest> request) async {
    return joinWaitlist(call, await request);
  }

  $async.Future<$2.CanSignupResponse> canSignup_Pre($grpc.ServiceCall call, $async.Future<$2.CanSignupRequest> request) async {
    return canSignup(call, await request);
  }

  $async.Future<$2.Empty> setSignupComplete_Pre($grpc.ServiceCall call, $async.Future<$2.SetSignupCompleteRequest> request) async {
    return setSignupComplete(call, await request);
  }

  $async.Future<$2.IsMugAvailableResponse> isMugAvailable_Pre($grpc.ServiceCall call, $async.Future<$2.IsMugAvailableRequest> request) async {
    return isMugAvailable(call, await request);
  }

  $async.Future<$2.ListTransactionsResponse> listTransactions_Pre($grpc.ServiceCall call, $async.Future<$2.PaginationRequest> request) async {
    return listTransactions(call, await request);
  }

  $async.Future<$2.ListTransactionsResponse> listTransactionsCompleted_Pre($grpc.ServiceCall call, $async.Future<$2.PaginationRequest> request) async {
    return listTransactionsCompleted(call, await request);
  }

  $async.Future<$2.ListTransactionsResponse> listTransactionsWithPending_Pre($grpc.ServiceCall call, $async.Future<$2.PaginationRequest> request) async {
    return listTransactionsWithPending(call, await request);
  }

  $async.Future<$2.Transaction> lookupTransaction_Pre($grpc.ServiceCall call, $async.Future<$2.LookupTransactionRequest> request) async {
    return lookupTransaction(call, await request);
  }

  $async.Future<$2.ListLimitsResponse> listLimits_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return listLimits(call, await request);
  }

  $async.Future<$2.Empty> updateClientLimits_Pre($grpc.ServiceCall call, $async.Future<$2.UpdateClientLimitsRequest> request) async {
    return updateClientLimits(call, await request);
  }

  $async.Future<$2.Empty> createConnection_Pre($grpc.ServiceCall call, $async.Future<$2.CreateConnectionRequest> request) async {
    return createConnection(call, await request);
  }

  $async.Future<$2.ListConnectionsResponse> listConnections_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return listConnections(call, await request);
  }

  $async.Future<$2.Connection> getConnection_Pre($grpc.ServiceCall call, $async.Future<$2.GetConnectionRequest> request) async {
    return getConnection(call, await request);
  }

  $async.Future<$2.ConnectionLimits> getConnectionLimits_Pre($grpc.ServiceCall call, $async.Future<$2.GetConnectionLimitsRequest> request) async {
    return getConnectionLimits(call, await request);
  }

  $async.Future<$2.Empty> updateConnectionLimits_Pre($grpc.ServiceCall call, $async.Future<$2.UpdateConnectionLimitsRequest> request) async {
    return updateConnectionLimits(call, await request);
  }

  $async.Future<$2.Empty> deleteConnection_Pre($grpc.ServiceCall call, $async.Future<$2.DeleteConnectionRequest> request) async {
    return deleteConnection(call, await request);
  }

  $async.Future<$2.GetPublicWalletDetailsResponse> getPublicWalletDetails_Pre($grpc.ServiceCall call, $async.Future<$2.GetPublicWalletDetailsRequest> request) async {
    return getPublicWalletDetails(call, await request);
  }

  $async.Future<$2.Contact> createContact_Pre($grpc.ServiceCall call, $async.Future<$2.CreateContactRequest> request) async {
    return createContact(call, await request);
  }

  $async.Future<$2.ListContactsResponse> listContacts_Pre($grpc.ServiceCall call, $async.Future<$2.ListContactsRequest> request) async {
    return listContacts(call, await request);
  }

  $async.Future<$2.ListIdentitiesResponse> listIdentities_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return listIdentities(call, await request);
  }

  $async.Future<$2.ListIdentitiesResponse> listPublicIdentities_Pre($grpc.ServiceCall call, $async.Future<$2.ListPublicIdentitiesRequest> request) async {
    return listPublicIdentities(call, await request);
  }

  $async.Future<$2.Empty> deleteIdentity_Pre($grpc.ServiceCall call, $async.Future<$2.DeleteIdentityRequest> request) async {
    return deleteIdentity(call, await request);
  }

  $async.Future<$2.Identity> setIdentityPublic_Pre($grpc.ServiceCall call, $async.Future<$2.SetIdentityPublicRequest> request) async {
    return setIdentityPublic(call, await request);
  }

  $async.Future<$2.GetIdentityResponse> getIdentity_Pre($grpc.ServiceCall call, $async.Future<$2.GetIdentityRequest> request) async {
    return getIdentity(call, await request);
  }

  $async.Future<$2.GetIdentityResponse> getIdentityBySignatureHash_Pre($grpc.ServiceCall call, $async.Future<$2.GetIdentityBySignatureHashRequest> request) async {
    return getIdentityBySignatureHash(call, await request);
  }

  $async.Future<$2.Empty> verifyIdentity_Pre($grpc.ServiceCall call, $async.Future<$2.VerifyIdentityRequest> request) async {
    return verifyIdentity(call, await request);
  }

  $async.Future<$2.KYCStatusResponse> kYCStatus_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return kYCStatus(call, await request);
  }

  $async.Future<$2.Empty> setKYCStatusPending_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return setKYCStatusPending(call, await request);
  }

  $async.Future<$2.Empty> startKYC_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return startKYC(call, await request);
  }

  $async.Future<$2.KYCPersonaInquiryResponse> getPersonaInquiry_Pre($grpc.ServiceCall call, $async.Future<$2.KYCPersonaInquiryRequest> request) async {
    return getPersonaInquiry(call, await request);
  }

  $async.Future<$2.MXWidgetResponse> getMXWidget_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getMXWidget(call, await request);
  }

  $async.Future<$2.CreateMXBankAccountsResponse> createMXBankAccounts_Pre($grpc.ServiceCall call, $async.Future<$2.CreateMXBankAccountsRequest> request) async {
    return createMXBankAccounts(call, await request);
  }

  $async.Future<$2.Empty> onboardGMTUser_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return onboardGMTUser(call, await request);
  }

  $async.Future<$2.Init3DSResponse> init3DS_Pre($grpc.ServiceCall call, $async.Future<$2.Init3DSRequest> request) async {
    return init3DS(call, await request);
  }

  $async.Future<$2.Lookup3DSResponse> lookup3DS_Pre($grpc.ServiceCall call, $async.Future<$2.Lookup3DSRequest> request) async {
    return lookup3DS(call, await request);
  }

  $async.Future<$2.Authenticate3DSResponse> authenticate3DS_Pre($grpc.ServiceCall call, $async.Future<$2.Authenticate3DSRequest> request) async {
    return authenticate3DS(call, await request);
  }

  $async.Future<$2.LinkedAccount> createCard_Pre($grpc.ServiceCall call, $async.Future<$2.CreateCardRequest> request) async {
    return createCard(call, await request);
  }

  $async.Future<$2.CardDetails> getCardDetails_Pre($grpc.ServiceCall call, $async.Future<$2.GetCardDetailsRequest> request) async {
    return getCardDetails(call, await request);
  }

  $async.Future<$2.Features> listFeatures_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return listFeatures(call, await request);
  }

  $async.Future<$2.CreateTwitterAuthURLResponse> createTwitterAuthURL_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return createTwitterAuthURL(call, await request);
  }

  $async.Future<$2.TwitterCallbackResponse> twitterCallback_Pre($grpc.ServiceCall call, $async.Future<$2.TwitterCallbackRequest> request) async {
    return twitterCallback(call, await request);
  }

  $async.Future<$2.CreateDomainIdentityResponse> createDomainIdentity_Pre($grpc.ServiceCall call, $async.Future<$2.CreateDomainIdentityRequest> request) async {
    return createDomainIdentity(call, await request);
  }

  $async.Future<$2.GetPaymentAddressResponse> getPaymentAddress_Pre($grpc.ServiceCall call, $async.Future<$2.GetPaymentAddressRequest> request) async {
    return getPaymentAddress(call, await request);
  }

  $async.Future<$2.Payment> createPayment_Pre($grpc.ServiceCall call, $async.Future<$2.CreatePaymentRequest> request) async {
    return createPayment(call, await request);
  }

  $async.Future<$2.Payment> updatePayment_Pre($grpc.ServiceCall call, $async.Future<$2.UpdatePaymentRequest> request) async {
    return updatePayment(call, await request);
  }

  $async.Future<$2.Payment> getPayment_Pre($grpc.ServiceCall call, $async.Future<$2.GetPaymentRequest> request) async {
    return getPayment(call, await request);
  }

  $async.Future<$2.Payment> confirmPayment_Pre($grpc.ServiceCall call, $async.Future<$2.ConfirmPaymentRequest> request) async {
    return confirmPayment(call, await request);
  }

  $async.Future<$2.GetLinkedAccountsForPaymentResponse> getLinkedAccountsForPayment_Pre($grpc.ServiceCall call, $async.Future<$2.GetLinkedAccountsForPaymentRequest> request) async {
    return getLinkedAccountsForPayment(call, await request);
  }

  $async.Future<$2.SearchWalletsResponse> searchWallets_Pre($grpc.ServiceCall call, $async.Future<$2.SearchWalletsRequest> request) async {
    return searchWallets(call, await request);
  }

  $async.Future<$2.DiscordCallbackResponse> discordCallback_Pre($grpc.ServiceCall call, $async.Future<$2.DiscordCallbackRequest> request) async {
    return discordCallback(call, await request);
  }

  $async.Future<$2.CreateDiscordAuthURLResponse> createDiscordAuthURL_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return createDiscordAuthURL(call, await request);
  }

  $async.Future<$2.Empty> submitForm_Pre($grpc.ServiceCall call, $async.Future<$2.SubmitFormRequest> request) async {
    return submitForm(call, await request);
  }

  $async.Future<$2.CreateSlackAuthURLResponse> createSlackAuthURL_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return createSlackAuthURL(call, await request);
  }

  $async.Future<$2.SlackCallbackResponse> slackCallback_Pre($grpc.ServiceCall call, $async.Future<$2.SlackCallbackRequest> request) async {
    return slackCallback(call, await request);
  }

  $async.Future<$2.LinkedAccount> addXagoBankAccount_Pre($grpc.ServiceCall call, $async.Future<$2.AddXagoBankAccountRequest> request) async {
    return addXagoBankAccount(call, await request);
  }

  $async.Future<$2.LinkedAccount> addXagoBalanceAccount_Pre($grpc.ServiceCall call, $async.Future<$2.AddXagoBalanceAccountRequest> request) async {
    return addXagoBalanceAccount(call, await request);
  }

  $async.Future<$2.Payment> withdrawXagoBalance_Pre($grpc.ServiceCall call, $async.Future<$2.WithdrawXagoBalanceRequest> request) async {
    return withdrawXagoBalance(call, await request);
  }

  $async.Future<$2.GetXagoBalanceResponse> getXagoBalances_Pre($grpc.ServiceCall call, $async.Future<$2.Empty> request) async {
    return getXagoBalances(call, await request);
  }

  $async.Future<$2.GetXagoDepositDetailsResponse> getXagoDepositDetails_Pre($grpc.ServiceCall call, $async.Future<$2.GetXagoDepositDetailsRequest> request) async {
    return getXagoDepositDetails(call, await request);
  }

  $async.Future<$2.Empty> updateIndividualKYC($grpc.ServiceCall call, $2.UpdateIndividualKYCRequest request);
  $async.Future<$2.IndividualKYCResponse> getIndividualKYC($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.IsUSPSAddressResponse> isUSPSAddress($grpc.ServiceCall call, $2.Address request);
  $async.Future<$2.SetSignupUserDataResponse> setSignupUserData($grpc.ServiceCall call, $2.SetSignupUserDataRequest request);
  $async.Future<$2.Empty> setSignupMobileNumber($grpc.ServiceCall call, $2.SetSignupMobileNumberRequest request);
  $async.Future<$2.Signup> getSignup($grpc.ServiceCall call, $2.GetSignupRequest request);
  $async.Future<$2.Empty> completeSignup($grpc.ServiceCall call, $2.CompleteSignupRequest request);
  $async.Future<$2.Empty> createUserDefaultWallet($grpc.ServiceCall call, $2.CreateUserDefaultWalletRequest request);
  $async.Future<$2.Empty> createWalletAddress($grpc.ServiceCall call, $2.CreateWalletAddressRequest request);
  $async.Future<$2.WalletAddressExistsResponse> walletAddressExists($grpc.ServiceCall call, $2.WalletAddressExistsRequest request);
  $async.Future<$2.Empty> setWalletName($grpc.ServiceCall call, $2.SetWalletNameRequest request);
  $async.Future<$2.WalletInfo> getWalletInfo($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.PublicWalletInfo> getPublicWalletInfo($grpc.ServiceCall call, $2.GetPublicWalletInfoRequest request);
  $async.Future<$2.Empty> sendPhoneVerification($grpc.ServiceCall call, $2.SendPhoneVerificationRequest request);
  $async.Future<$2.Empty> checkPhoneVerification($grpc.ServiceCall call, $2.CheckPhoneVerificationRequest request);
  $async.Future<$2.Empty> sendOTP($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Agreement> getAgreement($grpc.ServiceCall call, $2.GetAgreementRequest request);
  $async.Future<$2.SignAgreementsResponse> signAgreements($grpc.ServiceCall call, $2.SignAgreementsRequest request);
  $async.Future<$2.GetLinkedAccountsResponse> getLinkedAccounts($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.LinkedAccount> getLinkedAccount($grpc.ServiceCall call, $2.GetLinkedAccountRequest request);
  $async.Future<$2.LinkedAccount> setDefaultReceiveLinkedAccount($grpc.ServiceCall call, $2.SetDefaultReceiveLinkedAccountRequest request);
  $async.Future<$2.LinkedAccount> setDefaultSendLinkedAccount($grpc.ServiceCall call, $2.SetDefaultSendLinkedAccountRequest request);
  $async.Future<$2.LinkedAccount> setNicknameLinkedAccount($grpc.ServiceCall call, $2.SetNicknameLinkedAccountRequest request);
  $async.Future<$2.Empty> deleteLinkedAccount($grpc.ServiceCall call, $2.DeleteLinkedAccountRequest request);
  $async.Future<$2.Empty> createSupportTicket($grpc.ServiceCall call, $2.CreateSupportTicketRequest request);
  $async.Future<$2.GetCountriesResponse> getCountries($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.GetCurrentWalletResponse> getCurrentWallet($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.JoinWaitlistResponse> joinWaitlist($grpc.ServiceCall call, $2.JoinWaitlistRequest request);
  $async.Future<$2.CanSignupResponse> canSignup($grpc.ServiceCall call, $2.CanSignupRequest request);
  $async.Future<$2.Empty> setSignupComplete($grpc.ServiceCall call, $2.SetSignupCompleteRequest request);
  $async.Future<$2.IsMugAvailableResponse> isMugAvailable($grpc.ServiceCall call, $2.IsMugAvailableRequest request);
  $async.Future<$2.ListTransactionsResponse> listTransactions($grpc.ServiceCall call, $2.PaginationRequest request);
  $async.Future<$2.ListTransactionsResponse> listTransactionsCompleted($grpc.ServiceCall call, $2.PaginationRequest request);
  $async.Future<$2.ListTransactionsResponse> listTransactionsWithPending($grpc.ServiceCall call, $2.PaginationRequest request);
  $async.Future<$2.Transaction> lookupTransaction($grpc.ServiceCall call, $2.LookupTransactionRequest request);
  $async.Future<$2.ListLimitsResponse> listLimits($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Empty> updateClientLimits($grpc.ServiceCall call, $2.UpdateClientLimitsRequest request);
  $async.Future<$2.Empty> createConnection($grpc.ServiceCall call, $2.CreateConnectionRequest request);
  $async.Future<$2.ListConnectionsResponse> listConnections($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Connection> getConnection($grpc.ServiceCall call, $2.GetConnectionRequest request);
  $async.Future<$2.ConnectionLimits> getConnectionLimits($grpc.ServiceCall call, $2.GetConnectionLimitsRequest request);
  $async.Future<$2.Empty> updateConnectionLimits($grpc.ServiceCall call, $2.UpdateConnectionLimitsRequest request);
  $async.Future<$2.Empty> deleteConnection($grpc.ServiceCall call, $2.DeleteConnectionRequest request);
  $async.Future<$2.GetPublicWalletDetailsResponse> getPublicWalletDetails($grpc.ServiceCall call, $2.GetPublicWalletDetailsRequest request);
  $async.Future<$2.Contact> createContact($grpc.ServiceCall call, $2.CreateContactRequest request);
  $async.Future<$2.ListContactsResponse> listContacts($grpc.ServiceCall call, $2.ListContactsRequest request);
  $async.Future<$2.ListIdentitiesResponse> listIdentities($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.ListIdentitiesResponse> listPublicIdentities($grpc.ServiceCall call, $2.ListPublicIdentitiesRequest request);
  $async.Future<$2.Empty> deleteIdentity($grpc.ServiceCall call, $2.DeleteIdentityRequest request);
  $async.Future<$2.Identity> setIdentityPublic($grpc.ServiceCall call, $2.SetIdentityPublicRequest request);
  $async.Future<$2.GetIdentityResponse> getIdentity($grpc.ServiceCall call, $2.GetIdentityRequest request);
  $async.Future<$2.GetIdentityResponse> getIdentityBySignatureHash($grpc.ServiceCall call, $2.GetIdentityBySignatureHashRequest request);
  $async.Future<$2.Empty> verifyIdentity($grpc.ServiceCall call, $2.VerifyIdentityRequest request);
  $async.Future<$2.KYCStatusResponse> kYCStatus($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Empty> setKYCStatusPending($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Empty> startKYC($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.KYCPersonaInquiryResponse> getPersonaInquiry($grpc.ServiceCall call, $2.KYCPersonaInquiryRequest request);
  $async.Future<$2.MXWidgetResponse> getMXWidget($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.CreateMXBankAccountsResponse> createMXBankAccounts($grpc.ServiceCall call, $2.CreateMXBankAccountsRequest request);
  $async.Future<$2.Empty> onboardGMTUser($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Init3DSResponse> init3DS($grpc.ServiceCall call, $2.Init3DSRequest request);
  $async.Future<$2.Lookup3DSResponse> lookup3DS($grpc.ServiceCall call, $2.Lookup3DSRequest request);
  $async.Future<$2.Authenticate3DSResponse> authenticate3DS($grpc.ServiceCall call, $2.Authenticate3DSRequest request);
  $async.Future<$2.LinkedAccount> createCard($grpc.ServiceCall call, $2.CreateCardRequest request);
  $async.Future<$2.CardDetails> getCardDetails($grpc.ServiceCall call, $2.GetCardDetailsRequest request);
  $async.Future<$2.Features> listFeatures($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.CreateTwitterAuthURLResponse> createTwitterAuthURL($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.TwitterCallbackResponse> twitterCallback($grpc.ServiceCall call, $2.TwitterCallbackRequest request);
  $async.Future<$2.CreateDomainIdentityResponse> createDomainIdentity($grpc.ServiceCall call, $2.CreateDomainIdentityRequest request);
  $async.Future<$2.GetPaymentAddressResponse> getPaymentAddress($grpc.ServiceCall call, $2.GetPaymentAddressRequest request);
  $async.Future<$2.Payment> createPayment($grpc.ServiceCall call, $2.CreatePaymentRequest request);
  $async.Future<$2.Payment> updatePayment($grpc.ServiceCall call, $2.UpdatePaymentRequest request);
  $async.Future<$2.Payment> getPayment($grpc.ServiceCall call, $2.GetPaymentRequest request);
  $async.Future<$2.Payment> confirmPayment($grpc.ServiceCall call, $2.ConfirmPaymentRequest request);
  $async.Future<$2.GetLinkedAccountsForPaymentResponse> getLinkedAccountsForPayment($grpc.ServiceCall call, $2.GetLinkedAccountsForPaymentRequest request);
  $async.Future<$2.SearchWalletsResponse> searchWallets($grpc.ServiceCall call, $2.SearchWalletsRequest request);
  $async.Future<$2.DiscordCallbackResponse> discordCallback($grpc.ServiceCall call, $2.DiscordCallbackRequest request);
  $async.Future<$2.CreateDiscordAuthURLResponse> createDiscordAuthURL($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.Empty> submitForm($grpc.ServiceCall call, $2.SubmitFormRequest request);
  $async.Future<$2.CreateSlackAuthURLResponse> createSlackAuthURL($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.SlackCallbackResponse> slackCallback($grpc.ServiceCall call, $2.SlackCallbackRequest request);
  $async.Future<$2.LinkedAccount> addXagoBankAccount($grpc.ServiceCall call, $2.AddXagoBankAccountRequest request);
  $async.Future<$2.LinkedAccount> addXagoBalanceAccount($grpc.ServiceCall call, $2.AddXagoBalanceAccountRequest request);
  $async.Future<$2.Payment> withdrawXagoBalance($grpc.ServiceCall call, $2.WithdrawXagoBalanceRequest request);
  $async.Future<$2.GetXagoBalanceResponse> getXagoBalances($grpc.ServiceCall call, $2.Empty request);
  $async.Future<$2.GetXagoDepositDetailsResponse> getXagoDepositDetails($grpc.ServiceCall call, $2.GetXagoDepositDetailsRequest request);
}
