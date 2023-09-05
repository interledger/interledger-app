package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.com/fynbos/backend/dynamicforms"
)

func CreateForm(ctx context.Context, b Backends, args *dynamicforms.CreateFormArgs) (*dynamicforms.Form, error) {
	var form dynamicforms.Form

	jsonData, err := json.Marshal(args.FormData)
	if err != nil {
		return nil, fmt.Errorf("%w failed to marshal form data: %w", dynamicforms.ErrInternal, err)
	}

	if args.WalletID == "" {
		err = b.DB().GetContext(ctx, &form, "INSERT INTO dynamic_forms(form_id, data) VALUES($1, $2) RETURNING id, form_id, data", args.FormID, jsonData)
		if err != nil {
			return nil, fmt.Errorf("%w failed to create form: %w", dynamicforms.ErrInternal, err)
		}
		return &form, nil
	}

	err = b.DB().GetContext(ctx, &form, "INSERT INTO dynamic_forms(form_id, data, wallet_id) VALUES($1, $2, $3) RETURNING id, form_id, data, wallet_id", args.FormID, jsonData, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w failed to create form: %w", dynamicforms.ErrInternal, err)
	}

	return &form, nil
}
