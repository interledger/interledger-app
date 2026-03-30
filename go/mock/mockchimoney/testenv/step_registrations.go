//go:build e2e
// +build e2e

package main

import "github.com/cucumber/godog"

type scenarioStep struct {
	pattern string
	handler any
}

func (tc *TestContext) registerScenarioSteps(ctx *godog.ScenarioContext) {
	registerStepGroup(ctx, tc.serviceConfigSteps()...)
	registerStepGroup(ctx, tc.walletSetupSteps()...)
	registerStepGroup(ctx, tc.walletOperationSteps()...)
	registerStepGroup(ctx, tc.paymentSteps()...)
	registerStepGroup(ctx, tc.payoutSteps()...)
	registerStepGroup(ctx, tc.infoSteps()...)
	registerStepGroup(ctx, tc.kycSteps()...)
	registerStepGroup(ctx, tc.webhookTriggerSteps()...)
	registerStepGroup(ctx, tc.responseAssertionSteps()...)
	registerStepGroup(ctx, tc.payoutAssertionSteps()...)
	registerStepGroup(ctx, tc.infoAssertionSteps()...)
	registerStepGroup(ctx, tc.errorAssertionSteps()...)
	registerStepGroup(ctx, tc.webhookAssertionSteps()...)
}

func registerStepGroup(ctx *godog.ScenarioContext, steps ...scenarioStep) {
	for _, step := range steps {
		ctx.Step(step.pattern, step.handler)
	}
}

func (tc *TestContext) serviceConfigSteps() []scenarioStep {
	return []scenarioStep{
		{`^MockChimoney is running$`, tc.mockChimoneyIsRunning},
		{`^MockChimoney is running with authentication enforced$`, tc.mockChimoneyIsRunningWithAuthenticationEnforced},
		{`^MockChimoney is running with authentication disabled$`, tc.mockChimoneyIsRunningWithAuthenticationDisabled},
		{`^authentication is enforced$`, tc.authenticationIsEnforced},
		{`^I authenticate with a valid API key$`, tc.authenticateWithValidAPIKey},
		{`^the configured API key is "([^"]*)"$`, tc.setConfiguredAPIKey},
		{`^a webhook receiver is listening$`, tc.webhookReceiverIsListening},
		{`^the configured webhook secret is "([^"]*)"$`, tc.setConfiguredWebhookSecret},
		{`^the webhook secret is "([^"]*)"$`, tc.setWebhookSecret},
		{`^MockChimoney is configured with INTERAC_FEE_FLAT of "([^"]*)"$`, tc.setConfiguredInteracFeeFlat},
		{`^MockChimoney is configured with CAD_TO_USD_RATE of "([^"]*)"$`, tc.setConfiguredCADToUSDRate},
		{`^I send GET /health$`, tc.sendGetHealth},
		{`^I send GET /health without an X-API-KEY header$`, tc.sendGetHealthWithoutAPIKey},
		{`^I GET /health without an X-API-KEY header$`, tc.sendGetHealthWithoutAPIKey},
	}
}

func (tc *TestContext) walletSetupSteps() []scenarioStep {
	return []scenarioStep{
		{`^a sub-account exists with ID "([^"]*)"$`, tc.subAccountExistsWithID},
		{`^a wallet exists with name "([^"]*)"$`, tc.walletExistsWithName},
		{`^a wallet exists for "([^"]*)" with a known ID$`, tc.walletExistsForWithKnownID},
		{`^two wallets exist$`, tc.twoWalletsExist},
	}
}

func (tc *TestContext) walletOperationSteps() []scenarioStep {
	return []scenarioStep{
		{`^I POST /v0.2.4/multicurrency-wallets/create with body:$`, tc.postCreateWallet},
		{`^I POST /v0.2.4/multicurrency-wallets/create with header "X-API-KEY: ([^"]*)" and body:$`, tc.postCreateWalletWithHeader},
		{`^I POST /v0.2.4/multicurrency-wallets/create without an X-API-KEY header and body:$`, tc.postCreateWalletWithoutAPIKey},
		{`^I create two wallets both named "([^"]*)"$`, tc.createTwoWalletsBothNamed},
		{`^I GET /v0.2.4/multicurrency-wallets/get\?id=<wallet ID>$`, tc.getWalletByStoredID},
		{`^I GET /v0.2.4/multicurrency-wallets/get\?id=does-not-exist$`, tc.getWalletThatDoesNotExist},
		{`^I GET /v0.2.4/multicurrency-wallets/get without a query parameter$`, tc.getWalletWithoutQueryParameter},
		{`^I POST /v0.2.4/multicurrency-wallets/transfer with body:$`, tc.postTransferWithBody},
		{`^I POST /v0.2.4/multicurrency-wallets/transfer with:$`, tc.postTransferWithFields},
		{`^I POST /v0.2.4/multicurrency-wallets/transfer without amountToSend$`, tc.postTransferWithoutAmountToSend},
		{`^I POST /v0.2.4/multicurrency-wallets/transfer without originCurrency$`, tc.postTransferWithoutOriginCurrency},
		{`^I POST /v0.2.4/multicurrency-wallets/transfer without destinationCurrency$`, tc.postTransferWithoutDestinationCurrency},
		{`^I POST /v0.2.4/multicurrency-wallets/transfer with sendViaInterledger true$`, tc.postTransferWithSendViaInterledgerTrue},
	}
}

func (tc *TestContext) paymentSteps() []scenarioStep {
	return []scenarioStep{
		{`^I POST /v0.2.4/payment/initiate with body:$`, tc.postPaymentInitiate},
		{`^I POST /v0.2.4/payment/initiate without an X-API-KEY header and body:$`, tc.postPaymentInitiateWithoutAPIKey},
		{`^I have initiated a deposit for chi-sub-001 and recorded the issueID$`, tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID},
		{`^I have initiated a deposit and have the paymentLink$`, tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID},
		{`^I have initiated a deposit with:$`, tc.haveInitiatedDepositWith},
		{`^I GET the paymentLink URL$`, tc.getPaymentLinkURL},
		{`^I GET /pay/non-existent-issue-id$`, tc.getMissingPayPage},
		{`^I POST to the pay page confirm endpoint for the issueID$`, tc.postPayPageConfirmForIssueID},
		{`^I POST /v0.2.4/payment/verify with body:$`, tc.postPaymentVerifyWithBody},
		{`^I POST /v0.2.4/payment/verify with:$`, tc.postPaymentVerifyWithFields},
		{`^I POST /v0.2.4/payment/verify with an empty body$`, tc.postPaymentVerifyWithEmptyBody},
		{`^I have completed payment via the pay page$`, tc.haveCompletedPaymentViaThePayPage},
		{`^I have completed payment via the pay page for issueID "([^"]*)"$`, tc.haveCompletedPaymentViaThePayPageForIssueID},
		{`^I verify the payment$`, tc.verifyThePayment},
		{`^I have initiated a deposit and completed it via the pay page$`, tc.iHaveInitiatedADepositAndCompletedItViaThePayPage},
		{`^I wait for webhook delivery$`, tc.waitForWebhookDelivery},
		{`^I wait for all webhooks to be delivered$`, tc.waitForAllWebhooksToBeDelivered},
	}
}

func (tc *TestContext) payoutSteps() []scenarioStep {
	return []scenarioStep{
		{`^I POST /v0.2.4/payouts/interac with body:$`, tc.postPayoutInterac},
		{`^I POST /v0.2.4/payouts/interac without an X-API-KEY header and body:$`, tc.postPayoutInteracWithoutAPIKey},
		{`^I POST /v0.2.4/payouts/interac without debitCurrency and body:$`, tc.postPayoutInteracWithoutDebitCurrency},
		{`^I POST /v0.2.4/payouts/interac with two interac entries:$`, tc.postPayoutInteracWithTwoEntries},
		{`^I have initiated an Interac withdrawal$`, tc.iHaveInitiatedAnInteracWithdrawal},
		{`^I have initiated an Interac withdrawal and recorded the chiRef$`, tc.iHaveInitiatedAnInteracWithdrawal},
		{`^I have initiated a withdrawal and the payout.interac.completed webhook has been delivered$`, tc.haveInitiatedWithdrawalAndWebhookDelivered},
		{`^I have initiated a withdrawal and waited for webhook delivery$`, tc.haveInitiatedWithdrawalAndWaitedForWebhookDelivery},
		{`^I POST /v0.2.4/payouts/status with body:$`, tc.postPayoutStatusWithBody},
		{`^I POST /v0.2.4/payouts/status with an empty body$`, tc.postPayoutStatusWithEmptyBody},
		{`^I POST /v0.2.4/payouts/status with the chiRef$`, tc.postPayoutStatusWithTheChiRef},
		{`^I initiate an Interac withdrawal$`, tc.iHaveInitiatedAnInteracWithdrawal},
	}
}

func (tc *TestContext) infoSteps() []scenarioStep {
	return []scenarioStep{
		{`^I POST /v0.2.4/info/fee-estimate with body:$`, tc.postFeeEstimate},
		{`^I POST /v0.2.4/info/fee-estimate twice with the same body$`, tc.postFeeEstimateTwiceWithTheSameBody},
		{`^I GET /v0.2.4/info/convert/local-amount-to-usd with query params:$`, tc.getConvertLocalAmountToUSDWithQueryParams},
		{`^I convert 200 CAD to USD$`, tc.convert200CADToUSD},
		{`^I convert 200 CAD to USD again$`, tc.convert200CADToUSDAgain},
	}
}

func (tc *TestContext) kycSteps() []scenarioStep {
	return []scenarioStep{
		{`^the redirect URL is "([^"]*)"$`, tc.setRedirectURL},
		{`^I POST to the KYC approval endpoint for (kyc-sub-[0-9]+)$`, tc.postToKYCApprovalEndpoint},
		{`^I POST to the KYC approval endpoint for kyc-sub-009 again$`, tc.postToKYCApprovalEndpointAgain},
		{`^I POST to the KYC decline endpoint for (kyc-sub-[0-9]+)$`, tc.postToKYCDeclineEndpoint},
		{`^I GET /verify/kyc/kyc-sub-001\?redirect=https://app.test/callbacks/chimoney%3Fkyc$`, tc.getKYCPageForKYCSub001},
		{`^I GET /verify/kyc/does-not-exist\?redirect=https://app.test/callbacks/chimoney$`, tc.getMissingKYCPage},
		{`^I GET /verify/kyc/kyc-sub-002 without a redirect parameter$`, tc.getKYCPageWithoutRedirect},
		{`^I approve KYC for (kyc-sub-[0-9]+)$`, tc.approveKYCFor},
		{`^I decline KYC for (kyc-sub-[0-9]+)$`, tc.declineKYCFor},
		{`^a sub-account "kyc-sub-009" has already been approved$`, tc.subAccountKYCSub009HasAlreadyBeenApproved},
	}
}

func (tc *TestContext) webhookTriggerSteps() []scenarioStep {
	return []scenarioStep{
		{`^I have triggered a deposit and waited for webhook delivery$`, tc.haveTriggeredDepositAndWaitedForWebhookDelivery},
		{`^I have triggered a deposit and captured a webhook delivery$`, tc.haveTriggeredDepositAndCapturedWebhookDelivery},
		{`^I have triggered KYC approval and waited for webhook delivery$`, tc.haveTriggeredKYCApprovalAndWaitedForWebhookDelivery},
		{`^I trigger a deposit and capture a webhook$`, tc.triggerDepositAndCaptureAWebhook},
		{`^I verify the signature using the expected secret "([^"]*)"$`, tc.verifySignatureUsingTheExpectedSecret},
		{`^I verify the signature using the wrong secret "([^"]*)"$`, tc.verifySignatureUsingTheWrongSecret},
		{`^I manually compute HMAC-SHA256 over "\{svix-id\}\.\{svix-timestamp\}\.\{raw-body\}" with the configured key$`, tc.manuallyComputeHMACSHA256WithTheConfiguredKey},
	}
}

func (tc *TestContext) responseAssertionSteps() []scenarioStep {
	return []scenarioStep{
		{`^the response status is ([0-9]+)$`, tc.theResponseStatusIs},
		{`^the response JSON "([^"]*)" is "([^"]*)"$`, tc.theResponseJSONIs},
		{`^the response body is JSON with "([^"]*)" equal to "([^"]*)"$`, tc.theResponseBodyIsJSONWithEqualTo},
		{`^the response data contains a[n]? "([^"]*)"$`, tc.theResponseDataContains},
		{`^the response data "([^"]*)" is "([^"]*)"$`, tc.theResponseDataIs},
		{`^the response data "([^"]*)" is true$`, tc.theResponseDataIsTrue},
		{`^the response data nested field "([^"]*)" is "([^"]*)"$`, tc.theResponseDataNestedFieldIs},
		{`^the response data "issueID" matches the pattern "\{subAccountID\}_\{uuid\}"$`, tc.theResponseDataIssueIDMatchesThePattern},
		{`^the response data "chiRef" is a non-empty string$`, tc.theResponseDataChiRefIsANonEmptyString},
		{`^the payment data "status" is "([^"]*)"$`, tc.thePaymentDataStatusIs},
		{`^the payment data nested field "([^"]*)" is "([^"]*)"$`, tc.thePaymentDataNestedFieldIs},
		{`^the payment data nested field "([^"]*)" is a positive number$`, tc.thePaymentDataNestedFieldIsAPositiveNumber},
		{`^the Content-Type is "([^"]*)"$`, tc.theContentTypeIs},
		{`^the body contains a form for confirming payment$`, tc.theBodyContainsAFormForConfirmingPayment},
		{`^the body contains a form for completing KYC$`, tc.theBodyContainsAFormForCompletingKYC},
		{`^the body contains an "Approve KYC" action$`, tc.theBodyContainsAnApproveKYCAction},
		{`^the body contains a "Decline KYC" action$`, tc.theBodyContainsADeclineKYCAction},
		{`^the redirect URL includes query parameter "issueID" matching the issueID$`, tc.theRedirectURLIncludesQueryParameterIssueIDMatchingTheIssueID},
		{`^the redirect URL includes query parameter "status" equal to "success"$`, tc.theRedirectURLIncludesQueryParameterStatusEqualToSuccess},
		{`^the response redirects to a URL starting with "([^"]*)"$`, tc.theResponseRedirectsToAURLStartingWith},
		{`^the redirect URL includes a failure indicator query parameter$`, tc.theRedirectURLIncludesAFailureIndicatorQueryParameter},
	}
}

func (tc *TestContext) payoutAssertionSteps() []scenarioStep {
	return []scenarioStep{
		{`^the response data array contains ([0-9]+) payout[s]?$`, tc.theResponseDataArrayContainsPayouts},
		{`^each payout has an "issueID" matching the pattern "\{subAccountID\}_\{uuid\}"$`, tc.eachPayoutHasAnIssueIDMatchingThePattern},
		{`^each payout has a "chiref" field$`, tc.eachPayoutHasAChiRefField},
		{`^each payout has a distinct "issueID"$`, tc.eachPayoutHasADistinctIssueID},
		{`^the payout data "status" is "([^"]*)"$`, tc.thePayoutDataStatusIs},
		{`^the payout data "type" is "([^"]*)"$`, tc.thePayoutDataTypeIs},
		{`^the payout data "amount" equals the withdrawal amount$`, tc.thePayoutDataAmountEqualsTheWithdrawalAmount},
	}
}

func (tc *TestContext) infoAssertionSteps() []scenarioStep {
	return []scenarioStep{
		{`^the response data "totalFee" is a positive number$`, tc.theResponseDataTotalFeeIsAPositiveNumber},
		{`^the response data "netAmount" equals amount minus totalFee$`, tc.theResponseDataNetAmountEqualsAmountMinusTotalFee},
		{`^the response data "direction" is "([^"]*)"$`, tc.theResponseDataDirectionIs},
		{`^both responses have identical "totalFee" values$`, tc.bothResponsesHaveIdenticalTotalFeeValues},
		{`^the response data "totalFee" is ([0-9.]+)$`, tc.theResponseDataTotalFeeIs},
		{`^the response data "netAmount" is "amount" minus "totalFee"$`, tc.theResponseDataNetAmountIsAmountMinusTotalFee},
		{`^the response data "originCurrency" is "([^"]*)"$`, tc.theResponseDataOriginCurrencyIs},
		{`^the response data "amountInOriginCurrency" is "([^"]*)"$`, tc.theResponseDataAmountInOriginCurrencyIs},
		{`^the response data "amountInUSD" is a positive number$`, tc.theResponseDataAmountInUSDIsAPositiveNumber},
		{`^the response data contains "validUntil"$`, tc.theResponseDataContainsValidUntil},
		{`^the response data "amountInUSD" is ([0-9.]+)$`, tc.theResponseDataAmountInUSDIs},
		{`^the response data "amountInUSD" is 0$`, tc.theResponseDataAmountInUSDIsZero},
		{`^both responses return the same "amountInUSD"$`, tc.bothResponsesReturnTheSameAmountInUSD},
	}
}

func (tc *TestContext) errorAssertionSteps() []scenarioStep {
	return []scenarioStep{
		{`^the error message mentions "([^"]*)"$`, tc.theErrorMessageMentions},
		{`^the error message indicates currency must be USD when rail is not specified$`, tc.theErrorMessageIndicatesCurrencyMustBeUSDWhenRailIsNotSpecified},
		{`^the error message indicates KYC is already completed$`, tc.theErrorMessageIndicatesKYCIsAlreadyCompleted},
		{`^each wallet has a different "id" value$`, tc.eachWalletHasADifferentIDValue},
		{`^the sub-account kyc status is "([^"]*)"$`, tc.theSubAccountKYCStatusIs},
	}
}

func (tc *TestContext) webhookAssertionSteps() []scenarioStep {
	return []scenarioStep{
		{`^the webhook receiver received exactly ([0-9]+) requests$`, tc.theWebhookReceiverReceivedExactlyRequests},
		{`^the webhook receiver received a request with body:$`, tc.theWebhookReceiverReceivedARequestWithBody},
		{`^the webhook receiver received a request with:$`, tc.theWebhookReceiverReceivedARequestWithFields},
		{`^the webhook receiver received a request with body fields:$`, tc.theWebhookReceiverReceivedARequestWithBodyFields},
		{`^the webhook body "([^"]*)" matches the deposit issueID$`, tc.theWebhookBodyMatchesTheDepositIssueID},
		{`^the webhook body "([^"]*)" matches the withdrawal issueID$`, tc.theWebhookBodyMatchesTheWithdrawalIssueID},
		{`^the webhook body "issueID" starts with the sub-account ID followed by "_"$`, tc.theWebhookBodyIssueIDStartsWithTheSubAccountID},
		{`^the webhook "issueID" starts with "chi-sub-002_"$`, tc.theWebhookIssueIDStartsWithChiSub002},
		{`^the webhook body "meta.issuer" equals the sub-account ID$`, tc.theWebhookBodyMetaIssuerEqualsTheSubAccountID},
		{`^the webhook body "meta.currency" is "CAD"$`, tc.theWebhookBodyMetaCurrencyIsCAD},
		{`^the webhook body "meta.amount" equals the withdrawal amount$`, tc.theWebhookBodyMetaAmountEqualsTheWithdrawalAmount},
		{`^the webhook request includes valid svix signature headers$`, tc.theWebhookRequestIncludesValidSvixSignatureHeaders},
		{`^the eventTypes received are "([^"]*)" and "([^"]*)"$`, tc.theEventTypesReceivedAre},
		{`^the received webhook includes the header "([^"]*)"$`, tc.theReceivedWebhookIncludesTheHeader},
		{`^the "svix-id" value matches the pattern "msg_<uuid>"$`, tc.theSvixIDValueMatchesThePattern},
		{`^the "svix-timestamp" value is a Unix epoch integer$`, tc.theSvixTimestampValueIsAUnixEpochInteger},
		{`^the "svix-signature" value starts with "v1,"$`, tc.theSvixSignatureValueStartsWithV1},
		{`^the signature is valid$`, tc.theSignatureIsValid},
		{`^the signature is invalid$`, tc.theSignatureIsInvalid},
		{`^the result matches the base64 value in the "v1," prefix of "svix-signature"$`, tc.theResultMatchesTheBase64ValueInTheSignature},
		{`^the webhook body top-level fields include "([^"]*)"$`, tc.theWebhookBodyTopLevelFieldsInclude},
		{`^the webhook body does NOT contain a top-level "data" key wrapping the payload$`, tc.theWebhookBodyDoesNotContainATopLevelDataKey},
		{`^the signature is valid when verified with key "([^"]*)"$`, tc.theSignatureIsValidWhenVerifiedWithKey},
	}
}
