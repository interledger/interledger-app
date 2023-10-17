package db

import "database/sql"

func NullStrFromStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid:  s != ""}
}

func NullFloat64FromFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid:   f > 0}
}
