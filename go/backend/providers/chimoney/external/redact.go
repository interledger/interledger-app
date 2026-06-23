package external

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

var redactFields = []string{"email", "phoneNumber"}

func Redact(ctx context.Context, req []byte) ([]byte, error) {
	js := make(map[string]interface{})
	err := json.Unmarshal(req, &js)
	if err != nil {
		log.Error("Redacting Chimoney request failed. JSON unmarshalling failed", zap.Error(err))
		return req, nil
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
