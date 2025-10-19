package jobs

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func PtiSettleDepositAndWithdrawsForWallet(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var transactions []string
	err := workflow.ExecuteActivity(ctx, a.GetPTIPendingTransactions, walletID).Get(ctx, &transactions)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.GetAndPtiSettleTransactions, transactions, walletID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return "pti settle initiated", nil
}

func (a *Activity) GetPTIPendingTransactions(ctx context.Context, walletID string) ([]string, error) {
	var transactions []string
	err := a.b.DB().SelectContext(ctx, &transactions, "Select foreign_id FROM transactions where (type ='deposit' or type = 'withdrawal') AND asset_code = 'USD' AND state = 'Pending' AND wallet_id=$1; ", walletID)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (a *Activity) GetAndPtiSettleTransactions(ctx context.Context, transactions []string, walletID string) error {
	ptiPrivateKey, err := jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
	if err != nil {
		log.Fatalln(err)
	}

	ptiExternal, err := external.NewWithOptions(
		external.WithBaseURL(os.Getenv("PTI_BASE_URL")),
		external.WithOTELLHTTPClient(),
		external.WithClientID(os.Getenv("PTI_CLIENT_ID")),
		external.WithDerivedKeys(ptiPrivateKey),
	)
	if err != nil {
		log.Fatalln(fmt.Errorf("%w Failed to create PTI external client: %s", pti.ErrInternal, err))
	}
	for _, t := range transactions {
		data, err := ptiExternal.GetTransaction(ctx, t)
		if err != nil {
			log.Warn("Cannot get transaction", zap.String("transaction", t))
			continue
		}
		if data.Status == "SETTLED" {
			var dst pti.TransactionStatusPayload
			dst.ResourceType = data.ResourceType
			dst.RequestID = data.RequestID
			dst.ClientID = data.ClientID
			dst.UserID = data.UserID
			dst.Status = data.Status
			dst.Amount = data.Amount
			dst.Currency = data.Currency
			dst.TransactionType = data.TransactionType
			dst.PaymentMethod = data.PaymentMethod
			dst.Total.SubTotal.Amount = int(data.Total.Total.Amount)
			dst.Total.SubTotal.Currency = "USD"
			dst.Total.Fee.Amount = int(data.Total.Fee.Amount)
			dst.Total.Fee.Currency = data.Total.Fee.Currency
			dst.Total.Total.Amount = int(data.Total.Total.Amount)
			dst.Total.Total.Currency = data.Total.Total.Currency

			if t, err := time.Parse(time.RFC3339, data.Date); err == nil {
				dst.Date = pti.FlexibleTime{Time: t}
			} else {
				log.Warn("Can't convert date")
			}

			dst.Fees = fmt.Sprintf("%d %s", data.Total.Fee.Amount, data.Total.Fee.Currency)

			if v, ok := data.AdditionalInfos["transactionGroupId"].(string); ok {
				dst.TransactionGroupID = v
			}

			if data.TransactionType == "DEPOSIT" {
				err = a.b.PTI().HandleSettleDeposit(ctx, dst)
				if err != nil {
					log.Warn("Cannot handle DEPOSIT")
				}
			}
			if data.TransactionType == "WITHDRAWAL" {
				err = a.b.PTI().HandleSettleWithdraw(ctx, dst)
				if err != nil {
					log.Warn("Cannot handle WITHDRAWAL")
				}
			}

		}

	}
	return nil
}
