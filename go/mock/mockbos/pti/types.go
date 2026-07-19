package pti

import "github.com/interledger/interledger-app/go/backend/providers/pti/external"

type (
	CreateWithdrawalArgs struct {
		Initiator         external.User                        `json:"initiator,omitempty"`
		SourceMethod      external.SourceMethod                `json:"sourceMethod,omitempty"`
		DestinationMethod external.WithdrawalDestinationMethod `json:"destinationMethod,omitempty"`
		Amount            float64                              `json:"amount,omitempty"`
		USDAmount         float64                              `json:"usdValue,omitempty"`
		Type              string                               `json:"type,omitempty"`
	}

	CreateDepositArgs struct {
		Initiator         external.User              `json:"initiator,omitempty"`
		SourceMethod      external.SourceMethod      `json:"sourceMethod,omitempty"`
		DestinationMethod external.DestinationMethod `json:"destinationMethod,omitempty"`
		Amount            float64                    `json:"amount,omitempty"`
		Type              string                     `json:"type,omitempty"`
		Date              string                     `json:"date,omitempty"`
	}
)
