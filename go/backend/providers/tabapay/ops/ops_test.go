package ops_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/tabapay/ops"
)

func TestGetFXRate(t *testing.T) {
	ctx := context.Background()

	b := ops.NewTestBackends(func(b *ops.TestBackends) {
		b.Db = db.MigrateTestDB(t, ctx)
	})

	_, err := b.DB().ExecContext(ctx, `INSERT INTO tabapay_fx_rates (currency_code, network, buy_rate, buy_rate_inverted, sell_rate, sell_rate_inverted, file_name)
		VALUES ('ZAR', 'Visa', 0.0523412, 19.1054, 0.0518710, 19.2785, 'text.csv'), ('ZAR', 'Mastercard', 0.0523389,19.1062, 0.0518730, 19.2778, 'text.csv')`)
	require.NoError(t, err)

	fxr, err := ops.GetFXRate(ctx, b, "ZAR")
	require.NoError(t, err)

	assert.Equal(t, "ZAR", fxr.Currency.String())
	assert.Equal(t, 191.062, fxr.MatercardRate.FromUSD(10))
	assert.Equal(t, 0.51873, fxr.MatercardRate.ToUSD(10))
	assert.Equal(t, 191.054, fxr.VisaRate.FromUSD(10))
	assert.Equal(t, 0.51871, fxr.VisaRate.ToUSD(10))
}
