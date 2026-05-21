package ops

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func getEmailsAndGreeting(ctx context.Context, b Backends, walletID string) ([]sendgrid.Email, string, error) {
	users, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, "", err
	}
	if len(users) < 1 {
		err = fmt.Errorf("%w wallet has (%d) users associated", email.ErrInternal, len(users))
		return nil, "", err
	}

	id, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		id = &kyc.IndividualDetails{}
	}

	var emails []sendgrid.Email
	for _, u := range users {
		emails = append(emails, sendgrid.Email{
			Name:    id.FirstName + " " + id.LastName,
			Address: u.Email,
		})
	}

	greeting := strings.TrimSpace(fmt.Sprintf("Hello %s", id.FirstName)) + ","
	return emails, greeting, nil
}

func SendApplicationDeniedEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send application denied email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Application denied", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Application denied",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your wallet verification was denied"},
			{"paragraph": "We are unable to verify your identity at this time. Please contact support using the details below."},
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send application denied email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendApplicationApprovedEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send application approved email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	var w *wallets.Wallet
	w, err = b.Wallets().Get(ctx, walletID)
	if err != nil {
		w = &wallets.Wallet{}
	}

	err = b.External().SendTemplate(ctx, "Your wallet has been created", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Your wallet has been created",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your wallet has been activated,"},
			{"code": w.AddressString()},
			{"paragraph": "You can now use your wallet to send and receive payments."},
		},
		"cta": map[string]interface{}{
			"text": "Connect an account",
			"url":  fmt.Sprintf("%s/connect/card", env.GetUrl()),
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send application approved email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendApplicationPendingEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send application pending email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Your wallet is under review", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Your wallet is under review",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Pending review"},
			{"paragraph": "Your wallet is currently under review. We will notify you once it's complete or if any further information is needed."},
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send application pending email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendConnectedAccountEmail(ctx context.Context, b Backends, la linkedaccounts.LinkedAccount) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, la.WalletID)
	if err != nil {
		log.Error("Failed to send connected account email.", zap.Error(err), zap.String("walletID", la.WalletID), zap.String("linkedAccountID", la.ID))
		return
	}

	var table []map[string]interface{}
	if la.Provider == pti.ProviderName && la.Type == pti.TypeCard {
		table = append(table, map[string]interface{}{
			"label": "Card ending",
			"text":  strings.ReplaceAll(la.Mask, "*", ""),
		})
	} else if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBank {
		table = append(table, map[string]interface{}{
			"label": "Institution",
			"text":  la.ReceiveNetwork,
		}, map[string]interface{}{
			"label": "Account number",
			"text":  la.Mask,
		})
	}

	err = b.External().SendTemplate(ctx, "A new account has been connected", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "A new account has been connected",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "You have successfully connected a new account"},
			{"table": table},
		},
		"cta": map[string]interface{}{
			"text": "View your account",
			"url":  fmt.Sprintf("%s/accounts/%s", env.GetUrl(), la.ID),
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send connected account email.", zap.Error(err), zap.String("walletID", la.WalletID), zap.String("linkedAccountID", la.ID))
	}
}

func SendConnectedAccountDocumentsNeededEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send connected account documents needed email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "We need some documents from you", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "We need some documents from you",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Documents needed"},
			{"paragraph": "While connecting your account, the provided billing address appears to be incorrect. Please send a copy of your billing address to the support email address below."},
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send connected account documents needed email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendPaymentSentEmailV2(ctx context.Context, b Backends, walletID string, payment *payments.Payment) {
	if payment == nil {
		return
	}

	receiverWallet, err := b.Wallets().Get(ctx, payment.Receiver.WalletID)
	if err != nil {
		log.Error("Failed to send payment sent email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.SendTransactionID))
		return
	}

	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send payment sent email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.SendTransactionID))
		return
	}

	txURL, err := url.JoinPath(env.GetUrl(), "transactions", payment.SendTransactionID)
	if err != nil {
		log.Error("Failed to send payment sent email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.SendTransactionID))
		return
	}
	err = b.External().SendTemplate(ctx, "Payment sent", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Payment sent",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your recent payment was successful"},
			{
				"table": []map[string]interface{}{
					{"label": "Total amount", "text": payment.SenderAmount.Format(), "large": true},
					{"label": "To", "text": receiverWallet.Name},
					{"label": "Date", "text": payment.UpdatedAt.Format("02 Jan 2006")},
				},
			},
		},
		"cta": map[string]interface{}{
			"text": "View transaction",
			"url":  txURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send payment sent email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.SendTransactionID))
	}
}

func SendPaymentReceivedEmailV2(ctx context.Context, b Backends, walletID string, payment *payments.Payment) {
	if payment == nil {
		return
	}

	senderWallet, err := b.Wallets().Get(ctx, payment.Sender.WalletID)
	if err != nil {
		log.Error("Failed to send payment received email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.ReceiveTransactionID))
		return
	}

	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send payment received email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.ReceiveTransactionID))
		return
	}

	txURL, err := url.JoinPath(env.GetUrl(), "transactions", payment.ReceiveTransactionID)
	if err != nil {
		log.Error("Failed to send payment received email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.ReceiveTransactionID))
		return
	}
	err = b.External().SendTemplate(ctx, "Payment received", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Payment received",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "You have received a payment"},
			{
				"table": []map[string]interface{}{
					{"label": "Total amount", "text": payment.ReceiverAmount.Format(), "large": true},
					{"label": "From", "text": senderWallet.Name},
					{"label": "Date", "text": payment.UpdatedAt.Format("02 Jan 2006")},
				},
			},
		},
		"cta": map[string]interface{}{
			"text": "View transaction",
			"url":  txURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send payment received email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", payment.ReceiveTransactionID))
	}
}

func SendPaymentFailedEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send payment failed email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	actionUrl, err := url.JoinPath(env.GetUrl(), "pay")
	if err != nil {
		log.Error("Failed to send payment failed email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Payment unsuccessful", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Payment unsuccessful",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your recent payment was unsuccessful"},
			{"paragraph": "Please try again or contact support using the details below."},
		},
		"cta": map[string]interface{}{
			"text": "Try again",
			"url":  actionUrl,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send payment failed email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendDepositReceivedEmail(ctx context.Context, b Backends, walletID string, amt currency.Amount, sourceAccount, date string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send deposit received email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	tableinfo := []map[string]interface{}{{"label": "Amount", "text": amt.Format(), "large": true}}
	if sourceAccount != "" {
		tableinfo = append(tableinfo, map[string]interface{}{"label": "From", "text": sourceAccount, "large": true})
	}
	if date != "" {
		tableinfo = append(tableinfo, map[string]interface{}{"label": "Date", "text": date, "large": true})
	}

	err = b.External().SendTemplate(ctx, "Deposit received", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Deposit received",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your deposit has been received"},
			{"table": tableinfo},
		},
		"cta": map[string]interface{}{
			"text": "View new Balance",
			"url":  env.GetUrl(),
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send deposit received email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendWithdrawalEmail(ctx context.Context, b Backends, walletID string, amt currency.Amount, destinationAccount, date string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send withdrawal email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	tableinfo := []map[string]interface{}{{"label": "Amount", "text": amt.Format(), "large": true}}
	if destinationAccount != "" {
		tableinfo = append(tableinfo, map[string]interface{}{"label": "To", "text": destinationAccount, "large": true})
	}
	if date != "" {
		tableinfo = append(tableinfo, map[string]interface{}{"label": "Date", "text": date, "large": true})
	}

	err = b.External().SendTemplate(ctx, "Withdrawal complete", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Withdrawal complete",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your withdrawal has been processed"},
			{"table": tableinfo},
		},
		"cta": map[string]interface{}{
			"text": "View new Balance",
			"url":  env.GetUrl(),
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send withdrawal processed email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendWithdrawalFailedEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send withdrawal failed email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Withdrawal unsuccessful", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Withdrawal unsuccessful",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your recent withdrawal was unsuccessful"},
			{"paragraph": "Please try again or contact support using the details below."},
		},
		"cta": map[string]interface{}{
			"text": "Try again",
			"url":  env.GetUrl(),
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send withdrawal failed email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendDepositFailedEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send deposit failed email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Deposit unsuccessful", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Deposit unsuccessful",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your recent deposit was unsuccessful"},
			{"paragraph": "Please try again or contact support using the details below."},
		},
		"cta": map[string]interface{}{
			"text": "Try again",
			"url":  env.GetUrl(),
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send deposit failed email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendLimitsExceededEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send deposit received email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	support, err := url.JoinPath(env.GetUrl(), "support")
	if err != nil {
		log.Error("Failed to send limits exceeded email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Limits Exceeded", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Limits Exceeded",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "The limits on your account have been exceeded"},
			{"paragraph": "A recent transaction has exceeded the limits on your account. Please contact support to provide additional documentation."},
		},
		"cta": map[string]interface{}{
			"text": "Contact support",
			"url":  support,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send deposit received email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendCardCreatedEmail(ctx context.Context, b Backends, walletID, cardID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send card created email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	cardURL, err := url.JoinPath(env.GetUrl(), "cards", cardID)
	if err != nil {
		log.Error("Failed to send card created email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}
	err = b.External().SendTemplate(ctx, "💳 Your new card is ready for use!", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "💳 Your new card is ready for use!",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"paragraph": "We are pleased to let you know that your card is now ready."},
			{"paragraph": "You can view and manage it securely from your account."},
		},
		"cta": map[string]interface{}{
			"text": "View card",
			"url":  cardURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send card created email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendPending3DSConfirmation(ctx context.Context, b Backends, walletID, confirmationID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send pending 3ds ocnfirmation email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	confirmationURL, err := url.JoinPath(env.GetUrl(), "confirmations", confirmationID)
	if err != nil {
		log.Error("Failed to send card created email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}
	err = b.External().SendTemplate(ctx, "⚠️ Action Required: Pending 3D Secure Transaction", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "⚠️ Action Required: Pending 3D Secure Transaction",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"paragraph": "We noticed that a recent transaction requires 3D Secure (3DS) authentication to complete."},
			{"paragraph": "Please verify the transaction as soon as possible to ensure it’s processed successfully."},
			{"paragraph": "If you didn’t initiate this transaction, you can safely ignore it or contact our support team."},
		},
		"cta": map[string]interface{}{
			"text": "View confirmation",
			"url":  confirmationURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send card created email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendKYCDocumentsRequiredEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to retrieve email and greeting when sending KYC documents required email..", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	kycURL, err := url.JoinPath(env.GetUrl(), "personal-details")
	if err != nil {
		log.Error("Failed to send KYC documents required email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Action Required – Please Resubmit Your Verification Documents", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Action Required – Please Resubmit Your Verification Documents",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"paragraph": "Your KYC submission needs attention."},
			{"paragraph": "To continue using your wallet, please log in to your account and upload the requested documents."},
		},
		"cta": map[string]interface{}{
			"text": "Upload Documents",
			"url":  kycURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send KYC documents required email.", zap.Error(err), zap.String("walletID", walletID))
	}
}

func SendAccountDeletionRequestedEmail(ctx context.Context, b Backends, userID string) error {
	support := strings.TrimSpace(b.SupportEmail())
	if support == "" {
		return email.ErrSupportInboxNotConfigured
	}

	u, err := b.Users().GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %w", email.ErrInternal, err)
	}

	userWallets, err := b.Wallets().List(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %w", email.ErrInternal, err)
	}

	paragraphs := []string{
		"A user requested app account deletion.",
		"Environment: " + env.GetEnv(),
		"User ID: " + u.ID,
		"Email: " + strings.TrimSpace(u.Email),
		fmt.Sprintf("Wallet count: %d", len(userWallets)),
	}
	for _, w := range userWallets {
		paragraphs = append(paragraphs, "Wallet ID: "+w.ID)
	}

	sendTo := []sendgrid.Email{{Name: "Support", Address: support}}
	subject := "[" + env.GetEnv() + "] Account deletion requested — user " + userID
	supportData := make([]map[string]interface{}, 0, len(paragraphs))
	for _, p := range paragraphs {
		supportData = append(supportData, map[string]interface{}{"paragraph": p})
	}
	if err := b.External().SendTemplate(ctx, subject, sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": subject,
		"data":    supportData,
	}, nil); err != nil {
		return fmt.Errorf("%w: %w", email.ErrInternal, err)
	}

	// User-facing acknowledgement is courtesy only — log failures, don't fail the RPC.
	if strings.TrimSpace(u.Email) == "" {
		log.Warn("skipping account deletion confirmation email; user has no address on file",
			zap.String("userID", userID))
		return nil
	}
	name := strings.TrimSpace(u.FirstName + " " + u.LastName)
	confirmTo := []sendgrid.Email{{Name: name, Address: u.Email}}
	confirmSubject := "We've received your account deletion request"
	userData := []map[string]interface{}{
		{"paragraph": "We've received your request to delete your account."},
		{"paragraph": "If you still have funds in your account, please withdraw them within the next 2–3 days."},
		{"paragraph": "Our support team will contact you if anything else is needed."},
	}
	if err := b.External().SendTemplate(ctx, confirmSubject, confirmTo, b.OneTemplateID(), map[string]interface{}{
		"subject": confirmSubject,
		"data":    userData,
	}, nil); err != nil {
		log.Error("Failed to send account deletion confirmation email to user.",
			zap.Error(err),
			zap.String("userID", userID))
	}
	return nil
}

func SendAuthenticatorResetEmail(ctx context.Context, b Backends, walletID string) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to retrieve email and greeting when sending authenticator reset email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	securityURL, err := url.JoinPath(env.GetUrl(), "settings")
	if err != nil {
		log.Error("Failed to build settings URL for authenticator reset email.", zap.Error(err), zap.String("walletID", walletID))
		return
	}

	err = b.External().SendTemplate(ctx, "Your authenticator app was reset", sendTo, b.OneTemplateID(), map[string]interface{}{
		"subject": "Your authenticator app was reset",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Authenticator reset completed"},
			{"paragraph": "Your authenticator app setup has been reset by support. For your security, you will be asked to scan a new QR code and re-enroll your authenticator app on your next login."},
			{"paragraph": "If you did not request this action, please contact support immediately."},
		},
		"cta": map[string]interface{}{
			"text": "Review security settings",
			"url":  securityURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send authenticator reset email.", zap.Error(err), zap.String("walletID", walletID))
	}
}
