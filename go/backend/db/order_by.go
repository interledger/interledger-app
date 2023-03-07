package db

import (
	"errors"
	"fmt"
	"strings"
)

var (
	OrderByColumnViolationError    = errors.New("db: cant order by that column")
	OrderByDirectionViolationError = errors.New("db: cant order by that direction")
)

type OrderBy struct {
	column    string
	direction string
	table     string
}

func NewOrderBy(orderBy string, allowedCols []string, table string) (OrderBy, error) {

	s := strings.Split(orderBy, " ")

	column := ""
	direction := "asc" //default

	if len(s) == 1 {
		column = s[0]
	}
	if len(s) == 2 {
		column = s[0]
		direction = s[1]
	}

	// Check col allowed
	if !contains(allowedCols, column) {
		return OrderBy{}, OrderByColumnViolationError
	}

	// check dir allowed
	if !isDirectionAllowed(direction) {
		return OrderBy{}, OrderByDirectionViolationError
	}

	return OrderBy{
		column:    column,
		direction: direction,
		table:     table,
	}, nil
}

func (ob *OrderBy) SQLWhere(subId string) string {
	q := "WHERE (COLUMN > (select COLUMN from TABLE where id = SUB_ID) OR (COLUMN = (select COLUMN from TABLE where id = SUB_ID) AND id > SUB_ID))"
	q = strings.ReplaceAll(q, "COLUMN", ob.column)
	q = strings.ReplaceAll(q, "TABLE", ob.table)
	q = strings.ReplaceAll(q, "SUB_ID", subId)
	return q
}

func (ob *OrderBy) SQLOrderBy() string {
	return fmt.Sprintf("ORDER BY %s %s,id asc", ob.column, ob.direction)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func isDirectionAllowed(str string) bool {
	return str == "asc" || str == "desc"
}
