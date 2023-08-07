package ops

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const oneTemplateID = "d-d1d84d89553a43f89d6c60e2497b24c3"

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

	err = b.External().SendTemplate(ctx, "Application denied", sendTo, oneTemplateID, map[string]interface{}{
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

	err = b.External().SendTemplate(ctx, "Your wallet has been created", sendTo, oneTemplateID, map[string]interface{}{
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

	err = b.External().SendTemplate(ctx, "Your wallet is under review", sendTo, oneTemplateID, map[string]interface{}{
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
	var capabilities []string
	if la.Provider == tabapay.ProviderName {
		table = append(table, map[string]interface{}{
			"label": "Card ending",
			"text":  strings.ReplaceAll(la.Mask, "*", ""),
		})
		capabilities = append(capabilities, "This card")
	} else if la.Provider == mx.ProviderName {
		table = append(table, map[string]interface{}{
			"label": "Account ending",
			"text":  strings.ReplaceAll(la.Mask, "*", ""),
		})
		capabilities = append(capabilities, "This account")
	}

	if la.CanReceive && !la.CanSend {
		capabilities = append(capabilities, "is enabled to receive payments, but unable to send payments.")
		table = append(table, map[string]interface{}{
			"label": "Capabilities",
			"text":  strings.Join(capabilities, " "),
		})
	} else if la.CanSend && !la.CanReceive {
		capabilities = append(capabilities, "is enabled to send payments, but unable to receive payments.")
		table = append(table, map[string]interface{}{
			"label": "Capabilities",
			"text":  strings.Join(capabilities, " "),
		})
	} else if la.CanSend && la.CanReceive {
		capabilities = append(capabilities, "is enabled to send and receive payments.")
		table = append(table, map[string]interface{}{
			"label": "Capabilities",
			"text":  strings.Join(capabilities, " "),
		})
	}

	err = b.External().SendTemplate(ctx, "A new account has been connected", sendTo, oneTemplateID, map[string]interface{}{
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

	err = b.External().SendTemplate(ctx, "We need some documents from you", sendTo, oneTemplateID, map[string]interface{}{
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

func SendPaymentSentEmail(ctx context.Context, b Backends, walletID, trxID string, op openpayments.OutgoingPayment) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send payment sent email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", trxID))
		return
	}

	txURL, err := url.JoinPath(env.GetUrl(), "transactions", trxID)
	if err != nil {
		log.Error("Failed to send payment sent email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", trxID))
		return
	}
	err = b.External().SendTemplate(ctx, "Payment sent", sendTo, oneTemplateID, map[string]interface{}{
		"subject": "Payment sent",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "Your recent payment was successful"},
			{
				"table": []map[string]interface{}{
					{"label": "Total amount", "text": op.SentAmount.Format(), "large": true},
					{"label": "To", "text": op.ToPaymentPointer},
					{"label": "Date", "text": op.UpdatedAt.Format("02 Jan 2006")},
				},
			},
		},
		"cta": map[string]interface{}{
			"text": "View transaction",
			"url":  txURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send payment sent email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", trxID))
	}
}

func SendPaymentReceivedEmail(ctx context.Context, b Backends, walletID, trxID string, ip openpayments.IncomingPayment) {
	sendTo, greeting, err := getEmailsAndGreeting(ctx, b, walletID)
	if err != nil {
		log.Error("Failed to send payment received email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", trxID))
		return
	}

	txURL, err := url.JoinPath(env.GetUrl(), "transactions", trxID)
	if err != nil {
		log.Error("Failed to send payment received email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", trxID))
		return
	}
	err = b.External().SendTemplate(ctx, "Payment received", sendTo, oneTemplateID, map[string]interface{}{
		"subject": "Payment received",
		"data": []map[string]interface{}{
			{"paragraph": greeting},
			{"heading": "You have received a payment"},
			{
				"table": []map[string]interface{}{
					{"label": "Total amount", "text": ip.ReceivedAmount.Format(), "large": true},
					{"label": "From", "text": ip.FromPaymentPointer},
					{"label": "Date", "text": ip.UpdatedAt.Format("02 Jan 2006")},
				},
			},
		},
		"cta": map[string]interface{}{
			"text": "View transaction",
			"url":  txURL,
		},
	}, nil)
	if err != nil {
		log.Error("Failed to send payment received email.", zap.Error(err), zap.String("walletID", walletID), zap.String("trxID", trxID))
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

	err = b.External().SendTemplate(ctx, "Payment unsuccessful", sendTo, oneTemplateID, map[string]interface{}{
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
