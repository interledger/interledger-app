package ops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gitlab.com/fynbos/backend/providers/ilp"
)

func CreateStreamCredentials(
	ctx context.Context, b Backends, args ilp.CreateStreamCredentialsArgs,
) (*ilp.StreamCredentials, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"paymentTag": args.PaymentTag,
		"asset": map[string]interface{}{
			"code":  args.Currency.String(),
			"scale": args.Currency.Scale(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ilp.ErrInternal, err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/credentials", b.StreamServerBaseURL()), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ilp.ErrInternal, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ilp.ErrInternal, err)
	}

	var creds ilp.StreamCredentials
	if err = json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("%w %s", ilp.ErrInternal, err)
	}

	return &creds, nil
}

const insertDbIncomingPacketFields = "peer, payment_tag, amount, asset_scale, asset_code"

func ClearIncomingPackets(ctx context.Context, b Backends, packets []ilp.IncomingPacket) error {
	var values []interface{}
	var placeholders []string
	for i, packet := range packets {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		values = append(values, packet.Peer, packet.PaymentTag, packet.Amount.Value, packet.Amount.Currency.Scale(), packet.Amount.Currency.String())
	}
	sql := fmt.Sprintf("INSERT INTO ilp_packets (%s) VALUES %s;", insertDbIncomingPacketFields, strings.Join(placeholders, ","))
	result, err := b.DB().ExecContext(ctx, sql, values...)
	if err != nil {
		return fmt.Errorf("%w %s", ilp.ErrInternal, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", ilp.ErrInternal, err)
	}
	if rows < int64(len(packets)) {
		return fmt.Errorf("%w failed to insert packets into database.", ilp.ErrInternal)
	}

	return nil
}
