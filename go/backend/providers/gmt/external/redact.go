package external

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/clbanning/mxj"
)

func Redact(ctx context.Context, req []byte) ([]byte, error) {
	var redactedReq []byte
	reqMap, err := mxj.NewMapXml(req)
	if errors.Is(err, io.EOF) {
		redactedReq = req
	} else if err != nil {
		return nil, err
	}

	redact(reqMap)

	if reqMap != nil {
		redactedReq, err = reqMap.Xml()
		if err != nil {
			return nil, err
		}
	}

	return redactedReq, nil
}

func redact(fields map[string]interface{}) {
	for k, v := range fields {
		switch v := v.(type) {
		case map[string]interface{}:
			redact(v)
		case string:
			for _, i := range RedactFields {
				if strings.EqualFold(i, k) {
					fields[k] = "*****"
				}
			}
		}
	}
}
