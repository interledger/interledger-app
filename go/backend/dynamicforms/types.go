package dynamicforms

import (
	"database/sql"
	"time"
)

type (
	Submission struct {
		ID        string         `db:"id"`
		FormID    string         `db:"form_id"`
		Data      string         `db:"data"`
		WalletID  sql.NullString `db:"wallet_id"`
		CreatedAt time.Time      `db:"created_at"`
	}
	SubmissionCount struct {
		FormID string `db:"form_id"`
		Count  int32  `db:"count"`
	}
	SubmitArgs struct {
		FormID   string
		Data     interface{}
		WalletID string
	}
)
