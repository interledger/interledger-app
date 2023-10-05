package workflows

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const tabapayRatesBucket = "fynbos-tabapay"

func (a *Activity) GetLatestRatesFile(ctx context.Context) (string, error) {
	var latest types.Object
	pl := a.b.AWS().S3ListObjects(tabapayRatesBucket, "")

	for pl.HasMorePages() {
		page, err := pl.NextPage(ctx)
		if err != nil {
			return "", err
		}

		// Load go for the latest file only
		for _, obj := range page.Contents {
			if latest.LastModified == nil {
				latest = obj
				continue
			}
			if latest.LastModified.Before(*obj.LastModified) {
				latest = obj
			}
		}
	}

	// No rates files in the bucket
	if latest.Key == nil {
		return "", nil
	}

	return *latest.Key, nil
}

func (a *Activity) LoadFXRatesFromS3(ctx context.Context, filename string) error {

	// Read the contents of the latest file
	data, err := a.b.AWS().S3GetObjectData(ctx, tabapayRatesBucket, filename)
	if err != nil {
		return err
	}
	defer data.Close()

	csvReader := csv.NewReader(data)
	var i int
	for {
		i++
		line, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if i == 1 {
			continue
		}

		cc := currency.FromISO4217(strings.TrimSpace(line[1]))
		if !cc.Valid() {
			continue
		}
		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_fx_rates (currency_code, network, buy_rate, buy_rate_inverted, sell_rate, sell_rate_inverted, file_name) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING;",
			cc, strings.TrimSpace(line[4]), convertRate(line[2]), convertRate(line[5]), convertRate(line[3]), convertRate(line[6]), filename)
		if err != nil {
			log.Error("failed to load line from tabapay fx rates", zap.Error(err), zap.Int("line", i))
		}
	}

	return nil
}

func convertRate(rt string) float64 {
	rt = strings.TrimSpace(rt)
	rate, _ := strconv.ParseFloat(rt[len(rt)-6:], 64)
	pow, _ := strconv.ParseFloat(rt[:len(rt)-6], 64)

	return math.Pow(10, pow) * rate
}
