package client

import (
	"context"
	"errors"
	"io"
	"strings"

	"gitlab.com/fynbos/backend/providers/gmt/external"

	"github.com/clbanning/mxj"
)

func Redact(ctx context.Context, req []byte) ([]byte, error) {
	var redactedReq []byte
	reqMap, err := mxj.NewMapXmlSeq(req)
	if errors.Is(err, io.EOF) {
		redactedReq = req
	} else if err != nil {
		return nil, err
	}

	redact(reqMap, "")

	if reqMap != nil {
		redactedReq, err = reqMap.XmlSeq()
		if err != nil {
			return nil, err
		}
	}

	return redactedReq, nil
}

func redact(fields map[string]interface{}, parent string) {
	for k, v := range fields {
		switch v := v.(type) {
		case map[string]interface{}:
			redact(v, k)
		case string:
			for _, i := range external.RedactFields {
				if strings.HasSuffix(parent, i) {
					fields[k] = "*****"
				}
			}
		}
	}
}
