package client

import (
	"context"
	"encoding/json"
	"strings"
)

var redactFields = []string{"IBAN", "BIC", "tokenValue", "fieldValue", "accountNumber", "email", "mobileNumber", "identificationNumber"}

func Redact(ctx context.Context, req []byte) ([]byte, error) {
	js := make(map[string]interface{})
	err := json.Unmarshal(req, &js)
	if err != nil {
		return nil, err
	}
	redact(js)

	return json.Marshal(js)
}

func redact(fields map[string]interface{}) {
	for k, v := range fields {
		switch v := v.(type) {
		case map[string]interface{}:
			redact(v)
		case []interface{}:
			for _, entry := range v {
				redact(entry.(map[string]interface{}))
			}
		case string:
			for _, fn := range redactFields {
				if strings.EqualFold(k, fn) {
					fields[k] = "*****"
				}
			}
		}
	}
}
