package cmd

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gitlab.com/fynbos/httpmessagesignatures"
)

func NewPayCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pay [payment-pointer]",
		Short: "Pay a payment pointer",
		Example: `
		fynbos pay https://ilp.link/money
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var opts struct {
				Amount   float64
				Currency string
				ToPP     string `survey:"toPP"`
			}
			var err error
			var qs []*survey.Question

			if len(args) < 1 {
				qs = append(qs, &survey.Question{
					Name: "toPP",
					Prompt: &survey.Input{
						Message: "Who do you want to pay?",
					},
				})
			} else {
				opts.ToPP = args[0]
			}

			opts.Amount, err = cmd.Flags().GetFloat64("amount")
			if err != nil || opts.Amount == 0 {
				qs = append(qs, &survey.Question{
					Name: "amount",
					Prompt: &survey.Input{
						Message: "How much?",
					},
				})
			}

			opts.Currency, err = cmd.Flags().GetString("currency")
			if err != nil || opts.Currency == "" {
				qs = append(qs, &survey.Question{
					Name: "amount",
					Prompt: &survey.Select{
						Message: "Which currency?",
						Options: []string{"USD"},
						Default: "USD",
					},
				})
			}

			if len(qs) > 0 {
				err = survey.Ask(qs, &opts)
				if err != nil {
					return err
				}
			}

			return CreateOutgoingPayment(context.Background(), b, CreateOutgoingPaymentArgs{
				FromPP: b.Config().GetString("wallet"),
				Type:   "outgoing-payment",
				ToPP:   opts.ToPP,
				SendAmount: Amount{
					Amount:   opts.Amount,
					Currency: opts.Currency,
				},
			})
		},
	}

	_ = cmd.Flags().Float64P("amount", "a", 0, "amount to send")
	_ = cmd.Flags().StringP("currency", "c", "USD", "currency")

	return cmd
}

func CreateOutgoingPayment(ctx context.Context, b Backends, args CreateOutgoingPaymentArgs) error {
	signer, err := NewEd25519Signer(b.Config())
	if err != nil {
		return err
	}
	grantRequest, err := newGrantRequest(ctx, b, newGrantRequestArgs{
		GrantRequest: GrantRequest{
			AccessToken: []AccessTokenReq{
				{
					Access: []Access{
						{
							Type:       "outgoing-payment",
							Actions:    []string{"write, read"},
							Identifier: b.Config().GetString("wallet"),
						},
					},
				},
			},
			Client: b.Config().GetString("wallet"),
		},
		Signer: signer,
	})
	if err != nil {
		return err
	}

	resp, err := b.HttpClient().Do(grantRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var grant Grant
	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		err := json.NewDecoder(resp.Body).Decode(&grant)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsuccessful grant request. statusCode=%d url=%s", resp.StatusCode, b.Config().GetString("wallet"))
	}

	var authToken, outgoingPaymentURL string
	for _, token := range grant.Tokens {
		for _, access := range token.Access {
			if access.Type == "outgoing-payment" && len(access.Locations) > 0 {
				outgoingPaymentURL = access.Locations[0]
				break
			}
		}

		if outgoingPaymentURL != "" {
			authToken = token.Value
			break
		}
	}
	if authToken == "" || outgoingPaymentURL == "" {
		return fmt.Errorf("Failed to get grant.")
	}

	opRequest, err := newOutgoingPaymentRequest(ctx, b, newOutgoingPaymentRequestArgs{
		CreateOutgoingPaymentArgs: args,
		Signer:                    signer,
		OutgoingPaymentsURL:       outgoingPaymentURL,
		AuthToken:                 authToken,
	})
	if err != nil {
		return err
	}

	opResp, err := b.HttpClient().Do(opRequest)
	if err != nil {
		return err
	}

	var op OutgoingPayment
	switch {
	case opResp.StatusCode >= http.StatusOK && opResp.StatusCode < http.StatusMultipleChoices:
		err := json.NewDecoder(opResp.Body).Decode(&op)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsuccessful outgoing payment. statusCode=%d url=%s", opResp.StatusCode, outgoingPaymentURL)
	}

	fmt.Printf("view outgoing payment: %s\n", op.ID)

	return nil
}

type newGrantRequestArgs struct {
	Signer       crypto.Signer
	GrantRequest GrantRequest
}

func newGrantRequest(ctx context.Context, b Backends, args newGrantRequestArgs) (*http.Request, error) {
	gr := bytes.NewBuffer(nil)
	err := json.NewEncoder(gr).Encode(args.GrantRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", b.Config().GetString("wallet"), gr)
	if err != nil {
		return nil, err
	}
	digest, err := httpmessagesignatures.CreateContentDigest(ctx, gr.Bytes(), []string{"sha-256"})
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Digest", digest)
	err = httpmessagesignatures.SignRequest(
		ctx,
		req,
		args.Signer,
		[]string{"content-digest"},
		httpmessagesignatures.SignatureParams{
			KeyID:   b.Config().GetString("clientKeyID"),
			Created: uint64(time.Now().Unix()),
		},
		[]string{"content-digest"},
	)
	if err != nil {
		return nil, err
	}

	return req, nil
}

type newOutgoingPaymentRequestArgs struct {
	CreateOutgoingPaymentArgs
	AuthToken           string
	OutgoingPaymentsURL string
	Signer              crypto.Signer
}

func newOutgoingPaymentRequest(ctx context.Context, b Backends, args newOutgoingPaymentRequestArgs) (*http.Request, error) {
	op := bytes.NewBuffer(nil)
	err := json.NewEncoder(op).Encode(args.CreateOutgoingPaymentArgs)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", args.OutgoingPaymentsURL, op)
	if err != nil {
		return nil, err
	}

	digest, err := httpmessagesignatures.CreateContentDigest(ctx, op.Bytes(), []string{"sha-256"})
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Digest", digest)
	req.Header.Set("Authorization", args.AuthToken)

	err = httpmessagesignatures.SignRequest(
		ctx,
		req,
		args.Signer,
		[]string{"content-digest", "authorization"},
		httpmessagesignatures.SignatureParams{
			KeyID:   b.Config().GetString("clientKeyID"),
			Created: uint64(time.Now().Unix()),
		},
		[]string{"content-digest", "authorization"},
	)
	if err != nil {
		return nil, err
	}

	return req, nil
}
