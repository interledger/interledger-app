package httpmessagesignatures

import (
	"context"
	"crypto"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dunglas/httpsfv"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type zeroHasher struct{}

func (h zeroHasher) HashFunc() crypto.Hash {
	return 0
}

// Signs a request using https://www.ietf.org/archive/id/draft-ietf-httpbis-message-signatures-15.html#name-creating-a-signature
// This function assumes that it'll be adding the only signature. i.e. it'll set Signature: sig1={signature}.
func SignRequest(ctx context.Context, req *http.Request, signer crypto.Signer, components []string, params SignatureParams) error {
	sanitizedComponents := make([]string, len(components))
	for i, component := range components {
		sanitizedComponents[i] = strings.TrimSpace(strings.ToLower(component))
	}

	signatureParams := createSignatureParams(ctx, sanitizedComponents, params)
	base, err := createSignatureBase(ctx, req, signatureParams)
	if err != nil {
		log.Error("Failed to create signature base", zap.Error(err))
		return err
	}

	signature, err := signer.Sign(nil, []byte(base), zeroHasher{})
	if err != nil {
		log.Error("Failed to sign base", zap.Error(err))
		return err
	}

	sigInput := httpsfv.NewDictionary()
	sigInput.Add("sig-1", signatureParams)
	serializedSigInput, err := httpsfv.Marshal(sigInput)
	if err != nil {
		log.Error("Failed to serialize signature-input", zap.Error(err))
		return err
	}

	sigDictionary := httpsfv.NewDictionary()
	sigDictionary.Add("sig-1", httpsfv.NewItem(signature))
	serializedSigDictionary, err := httpsfv.Marshal(sigDictionary)
	if err != nil {
		log.Error("Failed to serialize signature", zap.Error(err))
		return err
	}

	req.Header.Set("Signature-Input", serializedSigInput)
	req.Header.Set("Signature", serializedSigDictionary)

	return nil
}

func createSignatureBase(ctx context.Context, req *http.Request, params httpsfv.InnerList) (string, error) {
	var parts []string

	for _, item := range params.Items {
		component := item.Value.(string)
		if strings.HasPrefix(component, "@") {
			extractedComponent, err := extractDerivedComponent(ctx, component, req)
			if err != nil {
				return "", err
			}
			parts = append(parts, extractedComponent...)
		} else {
			header := req.Header.Get(component)
			if header == "" {
				return "", fmt.Errorf("Header=%s is missing.", header)
			}
			parts = append(parts, fmt.Sprintf(`"%s": %s`, component, header))
		}
	}

	// This must always go last.
	signatureParams, err := httpsfv.Marshal(params)
	if err != nil {
		return "", err
	}
	parts = append(parts, fmt.Sprintf(`"@signature-params": %s`, signatureParams))

	return strings.Join(parts, "\n"), nil
}

// Serializes components and params according to https://www.ietf.org/archive/id/draft-ietf-httpbis-message-signatures-15.html#section-2.3
func createSignatureParams(ctx context.Context, components []string, params SignatureParams) httpsfv.InnerList {
	list := httpsfv.InnerList{Params: httpsfv.NewParams()}
	for _, component := range components {
		list.Items = append(list.Items, httpsfv.NewItem(strings.ToLower(strings.TrimSpace(component))))
	}

	list.Params.Add("created", time.Now().Unix())

	if params.Created != 0 {
		list.Params.Add("created", params.Created)
	}

	if params.Expires != 0 {
		list.Params.Add("expires", params.Expires)
	}

	if params.Alg != "" {
		list.Params.Add("alg", params.Alg)
	}

	if params.Tag != "" {
		list.Params.Add("tag", params.Tag)
	}

	if params.Nonce != "" {
		list.Params.Add("nonce", params.Tag)
	}

	if params.KeyID != "" {
		list.Params.Add("keyid", params.KeyID)
	}

	return list
}

func extractDerivedComponent(ctx context.Context, component string, req *http.Request) ([]string, error) {
	var parts []string
	switch component {
	case "@method":
		parts = append(parts, fmt.Sprintf(`"%s": %s`, component, req.Method))
	case "@target_uri":
		parts = append(parts, fmt.Sprintf(`"%s": %s`, component, req.RequestURI))
	case "@authority":
		parts = append(parts, fmt.Sprintf(`"%s": %s`, component, req.URL.Hostname()))
	case "@scheme":
		parts = append(parts, fmt.Sprintf(`"%s": %s`, component, req.URL.Scheme))
	case "@request-target":
		parts = append(parts, fmt.Sprintf(`"%s": %s?%s`, component, req.URL.Path, req.URL.RawQuery))
	case "@path":
		parts = append(parts, fmt.Sprintf(`"%s": %s`, component, req.URL.Path))
	case "@query":
		parts = append(parts, fmt.Sprintf(`"%s": ?%s`, component, req.URL.RawQuery))
	case "@query-params":
		for key, values := range req.URL.Query() {
			for _, value := range values {
				parts = append(parts, fmt.Sprintf(`"@query-params"; name="%s": %s`, key, value))
			}
		}
	case "@status":
		if req.Response == nil {
			return nil, fmt.Errorf("@status is only valid for http responses.")
		}
		parts = append(parts, fmt.Sprintf(`"%s": %d`, component, req.Response.StatusCode))
	default:
		return nil, fmt.Errorf("Unknown derived component %s", component)
	}

	return parts, nil
}

// Verifies a request signature using https://www.ietf.org/archive/id/draft-ietf-httpbis-message-signatures-15.html#section-3.2
// This function assumes that there's only one signature (sig1).
func VerifySignature(ctx context.Context, req *http.Request, publicKey crypto.PublicKey, verifier Verifier) bool {
	if req.Header.Get("Signature") == "" || req.Header.Get("Signature-Input") == "" {
		return false
	}

	sigInput, err := httpsfv.UnmarshalDictionary([]string{req.Header.Get("Signature-Input")})
	if err != nil {
		log.Error("Failed to unmarshal Signature-Input header.", zap.Error(err))
		return false
	}

	sig1Params, exists := sigInput.Get("sig-1")
	if !exists {
		log.Error("sig-1 does not exist in sig input dictionary.")
		return false
	}

	sigParamList, ok := sig1Params.(httpsfv.InnerList)
	if !ok {
		log.Error("Failed to cast sig input to inner list.")
		return false
	}

	base, err := createSignatureBase(ctx, req, sigParamList)
	if err != nil {
		log.Error("Failed to create signature base.", zap.Error(err))
		return false
	}

	sigDictionary, err := httpsfv.UnmarshalDictionary([]string{req.Header.Get("Signature")})
	if err != nil {
		log.Error("Failed to unmarshal Signature header.", zap.Error(err))
		return false
	}
	sigMember, exists := sigDictionary.Get("sig-1")
	if !exists {
		log.Error("sig-1 does not exist in signature dictionary.")
		return false
	}
	sigItem, ok := sigMember.(httpsfv.Item)
	if !ok {
		log.Error("Failed to cast unmarshal sig-1.")
		return false
	}
	signature, ok := sigItem.Value.([]byte)
	if !ok {
		log.Error("Failed to cast signature to []byte.")
		return false
	}

	return verifier.Verify(publicKey, []byte(base), signature)
}

func ExtractKeyIDForSignature(ctx context.Context, req *http.Request, signatureID string) string {
	if req.Header.Get("Signature-Input") == "" {
		return ""
	}

	sigInput, err := httpsfv.UnmarshalDictionary([]string{req.Header.Get("Signature-Input")})
	if err != nil {
		log.Error("Failed to unmarshal Signature-Input header.", zap.Error(err))
		return ""
	}

	sig1Params, exists := sigInput.Get(signatureID)
	if !exists {
		log.Error("sig-1 does not exist in sig input dictionary.")
		return ""
	}

	sigParamList, ok := sig1Params.(httpsfv.InnerList)
	if !ok {
		log.Error("Failed to cast sig input to inner list.")
		return ""
	}

	id, exists := sigParamList.Params.Get("keyid")
	if !exists {
		log.Error("Failed to get keyid.", zap.String("signature", signatureID))
		return ""
	}

	keyID, ok := id.(string)
	if !ok {
		log.Error("Failed cast keyid to string.")
		return ""
	}

	return keyID
}
