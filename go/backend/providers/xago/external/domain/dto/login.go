package dto

import (
	"bytes"
	"encoding/json"
	"io"
)

type LoginField struct {
	FieldName  string `json:"fieldName"`
	FieldValue string `json:"fieldValue"`
}

type LoginRequest struct {
	PolicyID string       `json:"policyId"`
	Fields   []LoginField `json:"fields"`
}

func NewLoginRequest(policyID, publicKey, secret string) LoginRequest {
	return LoginRequest{
		PolicyID: policyID,
		Fields: []LoginField{
			{FieldName: "apiPublicKey", FieldValue: publicKey},
			{FieldName: "apiSecretKey", FieldValue: secret},
		},
	}
}

func (r LoginRequest) Body() (io.Reader, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

type LoginResponse struct {
	Token string `json:"tokenValue"`
}
