#!/bin/bash
git checkout -- provision/funders.go

sed -i '4i\
\t"net/http"\n\t"bytes"\n\t"encoding/json"\n\t"strconv"' provision/funders.go

sed -i 's/func fundGatehub(ctx context.Context, w \*client.Wallet, spec countrySpec, targetMajor int64)/func fundGatehub(ctx context.Context, w *client.Wallet, spec countrySpec, targetMajor int64, dx DBContext, walletID string)/g' provision/funders.go

sed -i 's/func fund(ctx context.Context, w \*client.Wallet, spec countrySpec, targetMajor int64)/func fund(ctx context.Context, w *client.Wallet, spec countrySpec, targetMajor int64, dx DBContext, walletID string)/g' provision/funders.go

sed -i 's/return fundGatehub(ctx, w, spec, targetMajor)/return fundGatehub(ctx, w, spec, targetMajor, dx, walletID)/g' provision/funders.go

sed -i '106i\
\tvar externalID string\n\t_ = dx.QueryRowContext(ctx, "SELECT gu.external_id FROM gatehub_users gu JOIN user_wallets uw ON uw.wallet_id = gu.wallet_id WHERE uw.user_id = $1 LIMIT 1", w.UserID).Scan(\&externalID)\n\n\tvar providerID string\n\t_ = dx.QueryRowContext(ctx, "SELECT provider_id FROM linked_accounts WHERE wallet_id = $1 AND provider='"'"'gatehub'"'"' AND type='"'"'balance'"'"' AND deleted_at IS NULL", walletID).Scan(\&providerID)\n\n\tbodyMap := map[string]interface{}{\n\t\t"type": 1,\n\t\t"deposit_type": "external",\n\t\t"receiving_address": providerID,\n\t\t"amount": float64(targetMajor),\n\t\t"currency": "EUR",\n\t}\n\tbodyBytes, _ := json.Marshal(bodyMap)\n\turl := "https://mockgatehub.interledger.test/core/v1/transactions"\n\tts := strconv.FormatInt(time.Now().UnixMilli(), 10)\n\tsig := GenerateSignature(ts, "POST", url, string(bodyBytes), "local-test-app-secret")\n\treq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))\n\treq.Header.Set("Content-Type", "application/json")\n\treq.Header.Set("x-gatehub-managed-user-uuid", externalID)\n\treq.Header.Set("x-gatehub-app-id", "local-test-app-id")\n\treq.Header.Set("x-gatehub-timestamp", ts)\n\treq.Header.Set("x-gatehub-signature", sig)\n\thc := http.DefaultClient\n\tresp, err := hc.Do(req)\n\tif err == nil {\n\t\tdefer resp.Body.Close()\n\t}' provision/funders.go

sed -i '155d' provision/funders.go
sed -i '154a\
\twidget, widgetErr := w.GetKYCProviderWidget(ctx, \&pb.GetKYCProviderWidgetRequest{IdempotencyKey: uuid.NewString()})\n\tif widgetErr != nil {\n\t\tfmt.Printf("GetKYCProviderWidget error (tolerated): %s", widgetErr)\n\t} else {' provision/funders.go
sed -i '168i\
\t}' provision/funders.go
