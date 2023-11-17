package ops

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
)

/*
This creates the payments to the sender and receiver for the referral programme.
NB! This assumes it is being called at the end of the payments workflow.
Eligibility criteria:
1) Send amount >= $ 20.00
2) Sender has a limit of 3 referrals in rolling 24 hour window
3) Receiver must have just received their first payment
*/
func (a *Activity) CreateReferrals(ctx context.Context, originalPaymentID string) ([]string, error) {
	var referralPayments []string
	minSendAmount := currency.FromUInt64(20_00, currency.USD)

	p, err := Lookup(ctx, a.b, originalPaymentID)
	if err != nil {
		return nil, err
	}

	if p.SenderAmount.Value < minSendAmount.Value {
		return nil, nil
	}

	if p.Type != payments.TypePeer2Peer {
		return nil, nil
	}

	receiverLa, err := a.b.LinkedAccounts().GetDefaultReceive(ctx, p.Receiver.WalletID, p.ReceiverAmount.Currency)
	if err != nil {
		return nil, err
	}

	receiveTxs, err := a.b.Transactions().CountReceiveTransactions(ctx, p.Receiver.WalletID)
	if err != nil {
		return nil, err
	}
	if receiveTxs > 1 { // not eligible for referral. Called at end of payment workflow so there'll be 1 trx.
		return nil, err
	}

	receiveIds, err := a.b.Identities().List(ctx, p.Receiver.WalletID)
	if err != nil {
		return nil, err
	}
	var boost bool
	for _, i := range receiveIds {
		if i.Platform == identities.PlatformTwitter || i.Platform == identities.PlatformDiscord ||
			i.Platform == identities.PlatformSlack || i.Platform == identities.PlatformDomain {
			boost = true
			break
		}
	}

	referralAmount := currency.FromUInt64(10_00, currency.USD)
	if boost {
		referralAmount.Value = 20_00
	}

	// need to use lower amounts to not exceed tabapay sanbox limits
	if env.IsDev() {
		referralAmount.Value = 1
	}
	if env.IsDev() && boost {
		referralAmount.Value = 2
	}

	sendersLa, err := a.b.LinkedAccounts().GetDefaultReceive(ctx, p.Sender.WalletID, p.SenderAmount.Currency)
	if err != nil {
		return nil, err
	}

	referralsWallet, err := a.b.Wallets().Get(ctx, wallets.ReferralsWalletID)
	if err != nil {
		return nil, err
	}
	referralLa, err := a.b.LinkedAccounts().GetDefaultSend(ctx, wallets.ReferralsWalletID, p.SenderAmount.Currency)
	if err != nil {
		return nil, err
	}

	senderWallet, err := a.b.Wallets().Get(ctx, p.Sender.WalletID)
	if err != nil {
		return nil, err
	}

	senderReferralCount, err := a.b.Transactions().CountReferralsInPastDay(ctx, senderWallet.AddressString())
	if err != nil {
		return nil, err
	}
	if senderReferralCount < 3 {
		senderReferral, err := Create(ctx, a.b, payments.CreateArgs{
			Receiver:        p.Sender,
			ReceiverAccount: sendersLa.ID,
			ReceiverAmount:  referralAmount,
			Sender: payments.Identity{
				Type:       payments.IdentityTypeWalletID,
				Identifier: referralsWallet.ID,
			},
			SenderAmount:  referralAmount,
			SenderAccount: referralLa.ID,
			Type:          payments.TypeReferral,
			IPAddress:     "10.0.0.10",
			Note:          "Here's a little something to say thank you for using Fynbos!", // TODO: update copy
		})
		if err != nil {
			return nil, err
		}

		referralPayments = append(referralPayments, senderReferral.ID)
	}

	receiverReferral, err := Create(ctx, a.b, payments.CreateArgs{
		Receiver:        p.Receiver,
		ReceiverAccount: receiverLa.ID,
		ReceiverAmount:  referralAmount,
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: referralsWallet.ID,
		},
		SenderAmount:  referralAmount,
		SenderAccount: referralLa.ID,
		Type:          payments.TypeReferral,
		IPAddress:     "10.0.0.10",
		Note:          "Here's a little something to say thank you for using Fynbos!", // TODO: update copy
	})
	if err != nil {
		return nil, err
	}
	referralPayments = append(referralPayments, receiverReferral.ID)

	return referralPayments, nil
}
